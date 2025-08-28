package message_center

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

// var msgchannl = make(chan *MsgHolderOne, 1000)

// var msgHolderLock = new(sync.RWMutex)
// var msgHolder = make(map[string]*MsgHolder) // new(sync.Map)

func (this *MessageCenter) Init() {
	go func() {
		var mh *MsgHolder
		var ok bool
		var oneItr interface{}
		var msgOne *MsgHolderOne
		// var mhItr interface{}

		// for msgOne = range this.msgchannl {
		for {
			oneItr = this.msgchannl.Get(this.contextRoot)
			if oneItr == nil {
				//程序生命周期结束，退出了
				return
			}
			msgOne = oneItr.(*MsgHolderOne)

			// select {
			// case msgOne = <-this.msgchannl:
			// case <-this.contextRoot.Done():
			// 	return
			// }

			//TODO 查询记录是否存在，换个查询接口
			// _, err := new(sqlite3_db.MessageCache).FindByHash(msgOne.MsgHash) //

			//查询记录是否存在
			exist, err := FindMessageCacheByHashExist(msgOne.MsgHash, this.levelDB)
			if err != nil {
				utils.Log.Error().Msgf("查询消息 错误:%s", err.Error())
				continue
			}
			if exist {
				//消息已经存在
				this.SendNeighborReplyMsg(msgOne.Message, config.MSGID_multicast_return, nil, msgOne.Session)
				continue
			}

			//
			//先查询消息持有者缓存中是否存在
			// mhItr, ok = msgHolder.Load(utils.Bytes2string(msgOne.MsgHash))
			this.msgHolderLock.Lock()
			mh, ok = this.msgHolder[utils.Bytes2string(msgOne.MsgHash)]
			if !ok {
				mh = CreateMsgHolder(msgOne, this)
				// msgHolder.Store(utils.Bytes2string(msgOne.MsgHash), mh)
				this.msgHolder[utils.Bytes2string(msgOne.MsgHash)] = mh
				this.msgHolderLock.Unlock()
				// go GetMulticastMsg(mh)
			} else {
				this.msgHolderLock.Unlock()
				// mh = mhItr.(*MsgHolder)
				//判断这个持有者是否存在
				// utils.Log.Info().Msgf("添加一个消息持有者:%s", msgOne.Addr.B58String())
				mh.AddHolder(msgOne)
			}
			// if nodeStore.FindWhiteList(&msgOne.Addr) {
			// 	utils.Log.Info().Msgf("start sync multicast message from:%s %s", msgOne.Addr.B58String(), hex.EncodeToString(msgOne.MsgHash))
			// }

			//再次查询记录是否存在，防止多线程问题
			//  1. 第一个线程在保存消息之前，第二个线程过来查询，这时消息不存在
			//  2. 第一个线程删除了holder，第二个线程去查没有holder，因此又创建了一个holder去处理消息，因此又获取和处理了一次消息
			exist, err = FindMessageCacheByHashExist(msgOne.MsgHash, this.levelDB)
			if err != nil {
				utils.Log.Error().Msgf("查询消息 错误:%s", err.Error())
				continue
			}
			if exist {
				// 消息已经存在, 不再重复处理
				// 删除holder，防止内存泄漏
				this.msgHolderLock.Lock()
				delete(this.msgHolder, utils.Bytes2string(msgOne.MsgHash))
				this.msgHolderLock.Unlock()

				this.SendNeighborReplyMsg(msgOne.Message, config.MSGID_multicast_return, nil, msgOne.Session)
				continue
			}

			//判断同步消息协程是否在运行
			isRun := false
			//
			mh.Lock.RLock()
			if !mh.IsRun && !mh.IsFinish {
				mh.IsRun = true
				isRun = true
			}
			mh.Lock.RUnlock()
			if !isRun {
				//同步消息协程在运行，也许在等待新的消息提供者
				select {
				case mh.RunSignal <- false:
				default:
				}
				continue
			}
			go mh.GetMulticastMsg(this)
			// msgHolderLock.Unlock()
		}
	}()
}

type MsgHolderOne struct {
	MsgHash []byte //消息hash
	Addr    nodeStore.AddressNet
	Message *Message
	Session engine.Session
}

type MsgHolder struct {
	MsgHash   []byte          //消息hash
	Lock      *sync.RWMutex   //
	Holder    []*MsgHolderOne //消息持有者
	HolderTag []bool          //消息持有者同步标志，超时或者同步失败，设置为true
	IsRun     bool            //是否正在同步
	IsFinish  bool            //是否已经同步到了消息
	RunSignal chan bool       //继续运行信号
	// mc        *MessageCenter  //
}

func (this *MsgHolder) AddHolder(holder *MsgHolderOne) {
	this.Lock.Lock()
	have := false
	for _, one := range this.Holder {
		if bytes.Equal(one.Addr, holder.Addr) {
			have = true
			break
		}
	}
	if !have {
		this.Holder = append(this.Holder, holder)
		this.HolderTag = append(this.HolderTag, false)
	}
	this.Lock.Unlock()
}

func (this *MsgHolder) GetHolder() (holder *MsgHolderOne) {
	this.Lock.RLock()
	for i, one := range this.HolderTag {
		if one {
			continue
		}
		holder = this.Holder[i]
		break
	}
	this.Lock.RUnlock()
	return
}
func (this *MsgHolder) SetHolder(holder *MsgHolderOne) {
	this.Lock.RLock()
	for i, one := range this.Holder {
		if bytes.Equal(one.Addr, holder.Addr) {
			this.HolderTag[i] = true
			break
		}
	}
	this.Lock.RUnlock()
	return
}

func CreateMsgHolder(holder *MsgHolderOne, mc *MessageCenter) *MsgHolder {
	msgHolder := MsgHolder{
		MsgHash:   holder.MsgHash,
		Lock:      new(sync.RWMutex),
		Holder:    make([]*MsgHolderOne, 0),
		HolderTag: make([]bool, 0),
		// mc:        mc,
	}
	msgHolder.Holder = append(msgHolder.Holder, holder)
	msgHolder.HolderTag = append(msgHolder.HolderTag, false)
	return &msgHolder
}

/*
同步新消息并广播给其他节点
*/
func (this *MsgHolder) GetMulticastMsg(messageCenter *MessageCenter) {
	// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	// _, file, line, _ := runtime.Caller(0)
	// engine.AddRuntime(file, line, goroutineId)
	// defer engine.DelRuntime(file, line, goroutineId)

	//循环从消息持有者同步消息
	var messageProto *go_protobuf.MessageMulticast
	var message *Message
	var holder *MsgHolderOne
	var tmpMes []Message
	var tmpSes []engine.Session

	var err error
	for {
		holder = this.GetHolder()
		if holder == nil || len(holder.Addr) <= 0 {
			// utils.Log.Info().Msgf("not holder")
			//没有可用的消息提供者，设置超时删除这个协程
			timeout := time.NewTimer(time.Second * 10)
			select {
			case <-this.RunSignal:
				timeout.Stop()
				continue
			case <-timeout.C:
				break
			}
			break
		}
		this.SetHolder(holder)
		{
			this.Lock.RLock()
			tmpMes = nil
			tmpSes = nil
			for i := range this.Holder {
				tmp := *this.Holder[i]
				tmpMes = append(tmpMes, *tmp.Message)
				tmpSes = append(tmpSes, tmp.Session)
			}
			this.Lock.RUnlock()
		}

		// 获取需要同步对象的设备机器Id
		var recvMachineId string
		if holder.Session != nil {
			recvMachineId = holder.Session.GetMachineID()
		}
		// utils.Log.Info().Msgf("555555555555")
		messageProto, err = messageCenter.SyncMulticastMsg(holder.Addr, this.MsgHash, recvMachineId)
		if err != nil {
			// utils.Log.Info().Msgf("666666666666")
			utils.Log.Info().Msgf(err.Error())
			continue
		}
		//解析获取到的交易
		message, err = ParserMessageProto(messageProto.Head, messageProto.Body, 0)
		if err != nil {
			utils.Log.Error().Msgf("proto unmarshal error %s", err.Error())
			continue
		}
		err = message.ParserContentProto()
		if err != nil {
			utils.Log.Error().Msgf("proto unmarshal error %s", err.Error())
			continue
		}
		// utils.Log.Info().Msgf("保存这个消息到数据库 hash:%s", hex.EncodeToString(message.Body.Hash))
		//先保存这个消息到数据库
		// err = new(sqlite3_db.MessageCache).Add(message.Body.Hash, messageProto.Head, messageProto.Body)

		//先保存这个消息到数据库
		err = AddMessageCache(&message.Body.Hash, messageProto, messageCenter.levelDB)
		if err != nil {
			utils.Log.Error().Msgf(err.Error())
			continue
		}
		break
	}
	// utils.Log.Info().Msgf("77777777777")

	messageCenter.msgHolderLock.Lock()
	delete(messageCenter.msgHolder, utils.Bytes2string(this.MsgHash))
	messageCenter.msgHolderLock.Unlock()

	this.Lock.Lock()
	defer this.Lock.Unlock()

	this.IsRun = false
	if message != nil && err == nil {
		this.IsFinish = true
	} else {
		return
	}

	//同步到消息，给每个发送者都回复收到
	for i := range tmpMes {
		messageCenter.SendNeighborReplyMsg(&tmpMes[i], config.MSGID_multicast_return, nil, tmpSes[i])
	}

	//继续广播给其他节点
	// utils.Log.Info().Msgf("同步到了广播消息")
	if messageCenter.nodeManager.NodeSelf.GetIsSuper() {
		//广播给其他超级节点
		utils.Go(func() {
			// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			// _, file, line, _ := runtime.Caller(0)
			// engine.AddRuntime(file, line, goroutineId)
			// defer engine.DelRuntime(file, line, goroutineId)
			//先发送给超级节点
			// superNodes := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
			// whiltlistNodes := nodeStore.GetWhiltListNodes()

			superNodes := append(messageCenter.nodeManager.GetLogicNodes(), messageCenter.nodeManager.GetNodesClient()...)
			//广播给代理对象
			proxyNodes := messageCenter.nodeManager.GetProxyAll()
			messageCenter.BroadcastsAll(version_multicast, 0, nil, superNodes, proxyNodes, &this.MsgHash)
			return
		}, nil)
	}
	// utils.Log.Info().Msgf("99999999999999")
	//自己处理
	h := messageCenter.router.GetHandler(message.Body.MessageId)
	if h == nil {
		utils.Log.Info().Msgf("This broadcast message is not registered:", message.Body.MessageId)
		return
	}

	//

	// utils.Log.Info().Msgf("有广播消息，消息编号 %d", message.Body.MessageId)
	//TODO 这里获取不到控制器了,因此传空
	h(nil, engine.Packet{}, message)
}

/*
 * 去邻居节点同步广播消息
 *
 * @param	id			AddressNet			同步对象的地址
 * @param	hash		[]byte				需要同步的消息hash值
 * @param	machineId	string				同步对象的设备机器Id，可以穿空串，如果传值，则会根据设备id，获取对应的session进行处理
 * @return	res			MessageMulticast	同步到的消息信息
 * @return	err			error				错误信息
 */
func (this *MessageCenter) SyncMulticastMsg(id nodeStore.AddressNet, hash []byte, machineId string) (*go_protobuf.MessageMulticast, error) {
	// if nodeStore.FindWhiteList(&id) {
	// 	utils.Log.Info().Msgf("sync multicast message from:%s %s", id.B58String(), hex.EncodeToString(hash))
	// }
	head := NewMessageHead(this.nodeManager.NodeSelf, this.nodeManager.GetSuperPeerId(), nil, nil, false, this.nodeManager.GetMachineID(), machineId)
	body := NewMessageBody(0, &hash, 0, nil, 0) //广播采用同步机制后，不需要真实msgid，所以设置为0
	message := NewMessage(head, body)
	message.BuildHash()
	// fmt.Println("给这个session发送消息", recvid.B58String())
	mheadBs := head.Proto()
	mbodyBs, err := body.Proto()
	if err != nil {
		return nil, err
	}

	sessions, ok := this.sessionEngine.GetSessionAll(this.areaName, utils.Bytes2string(id))
	if !ok {
		return nil, config.ERROR_get_node_conn_fail
	}
	sendChan := make(chan error, len(sessions))
	var sendCnt int
	ok = false
	for k := range sessions {
		if machineId != "" && sessions[k].GetMachineID() != machineId {
			continue
		}
		sendCnt++
		ok = true

		go func(k int) {
			// utils.Log.Info().Msgf("[%s] SyncMulticastMsg tid:%s hash:%s msgHash:%s sIndex:%d", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), id.B58String(), hex.EncodeToString(hash), hex.EncodeToString(message.Body.Hash), sessions[k].GetIndex())
			err := sessions[k].Send(version_multicast_sync, &mheadBs, &mbodyBs, 0)
			sendChan <- err
		}(k)
	}
	if !ok {
		return nil, config.ERROR_get_node_conn_fail
	}

	flood.RegisterRequest(utils.Bytes2string(message.Body.Hash))
	if err != nil {
		utils.Log.Error().Msgf("msg send error to:%s %s", id.B58String(), err.Error())
		return nil, err
	}

	for i := 0; i < sendCnt; i++ {
		err = <-sendChan
		if err == nil {
			break
		}
	}

	bs, err := flood.WaitResponse(utils.Bytes2string(message.Body.Hash), config.Mining_block_time*time.Second)
	// bs, err := flood.WaitRequest(config.CLASS_engine_multicast_sync, utils.Bytes2string(message.Body.Hash), config.Mining_block_time) //香港节点网络不稳定
	if err != nil {
		// if nodeStore.FindWhiteList(&id) {
		// 	utils.Log.Info().Msgf("Timeout receiving broadcast reply message:%s %s", id.B58String(), hex.EncodeToString(hash))
		// }
		return nil, err
	}
	if bs == nil {
		// if nodeStore.FindWhiteList(&id) {
		// 	utils.Log.Info().Msgf("Timeout receiving broadcast reply message:%s %s", id.B58String(), hex.EncodeToString(hash))
		// }
		// utils.Log.Warn().Msgf("Timeout receiving broadcast reply message %s %s", id.B58String(), hex.EncodeToString(message.Body.Hash))
		// failNode = append(failNode, broadcasts[j])
		// continue
		return nil, errors.New("Timeout receiving broadcast reply message")
	}

	//验证同步到的消息
	mmp := new(go_protobuf.MessageMulticast)
	err = proto.Unmarshal(*bs, mmp)
	if err != nil {
		utils.Log.Error().Msgf("proto unmarshal error %s", err.Error())
		return nil, err
	}
	// if nodeStore.FindWhiteList(&id) {
	// 	utils.Log.Info().Msgf("sync multicast message success from:%s %s", id.B58String(), hex.EncodeToString(hash))
	// }
	return mmp, nil

	// head := MessageHead{
	// 		RecvId       :mmp
	// RecvSuperId   nodeStore.AddressNet          `json:"r_s_id"` //接收者的超级节点id
	// RecvVnode     virtual_node.AddressNetExtend `json:"r_v_id"` //接收者虚拟节点id
	// Sender        nodeStore.AddressNet          `json:"s_id"`   //发送者id
	// SenderSuperId nodeStore.AddressNet          `json:"s_s_id"` //发送者超级节点id
	// SenderVnode   virtual_node.AddressNetExtend `json:"s_v_id"` //发送者虚拟节点id
	// Accurate      bool                          `json:"a"`      //是否准确发送给一个节点，如果
	// }

	// return
	//将消息放数据库

}
