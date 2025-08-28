package message_center

import (
	"bytes"
	"github.com/rs/zerolog"
	"sync"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
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
		for {
			oneItr = this.msgchannl.Get(this.contextRoot)
			if oneItr == nil {
				//程序生命周期结束，退出了
				return
			}
			msgOne = oneItr.(*MsgHolderOne)
			//this.Log.Info().Hex("有新的广播消息", msgOne.MsgHash).Send()
			//查询记录是否存在
			exist, err := FindMessageCacheByHashExist(msgOne.MsgHash, this.levelDB)
			if err != nil {
				this.Log.Error().Msgf("查询消息 错误:%s", err.Error())
				continue
			}
			if exist {
				//this.Log.Info().Hex("消息已经存在", msgOne.MsgHash).Send()
				//消息已经存在
				this.SendNeighborReplyMsg(msgOne.Message, nil, 0)
				continue
			}

			//先查询消息持有者缓存中是否存在
			this.msgHolderLock.Lock()
			mh, ok = this.msgHolder[utils.Bytes2string(msgOne.MsgHash)]
			if !ok {
				mh = CreateMsgHolder(msgOne, &this.Log)
				// msgHolder.Store(utils.Bytes2string(msgOne.MsgHash), mh)
				this.msgHolder[utils.Bytes2string(msgOne.MsgHash)] = mh
				this.msgHolderLock.Unlock()
				// go GetMulticastMsg(mh)
			} else {
				this.msgHolderLock.Unlock()
				// mh = mhItr.(*MsgHolder)
				//判断这个持有者是否存在
				//this.Log.Info().Msgf("添加一个消息持有者:%s", msgOne.Addr.B58String())
				mh.AddHolder(msgOne)
			}
			// if nodeStore.FindWhiteList(&msgOne.Addr) {
			// 	utils.Log.Info().Msgf("start sync multicast message from:%s %s", msgOne.Addr.B58String(), hex.EncodeToString(msgOne.MsgHash))
			// }

			//this.Log.Info().Hex("有新的广播消息", msgOne.MsgHash).Send()

			//再次查询记录是否存在，防止多线程问题
			//  1. 第一个线程在保存消息之前，第二个线程过来查询，这时消息不存在
			//  2. 第一个线程删除了holder，第二个线程去查没有holder，因此又创建了一个holder去处理消息，因此又获取和处理了一次消息
			exist, err = FindMessageCacheByHashExist(msgOne.MsgHash, this.levelDB)
			if err != nil {
				this.Log.Error().Msgf("查询消息 错误:%s", err.Error())
				continue
			}
			if exist {
				// 消息已经存在, 不再重复处理
				// 删除holder，防止内存泄漏
				this.msgHolderLock.Lock()
				delete(this.msgHolder, utils.Bytes2string(msgOne.MsgHash))
				this.msgHolderLock.Unlock()

				this.SendNeighborReplyMsg(msgOne.Message, nil, 0)
				//this.SendNeighborReplyMsg(msgOne.Message, config.MSGID_multicast_return, nil, msgOne.Session)
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
	MsgHash        []byte               //消息hash
	Addr           nodeStore.AddressNet //
	ForwordMessage MessageBase          //原始待转发的广播,content是保存的消息hash，不会变，方便转发
	Message        *MessageBase         //消息经同步后，content保存同步下来的消息内容
	Session        engine.Session       //
}

type MsgHolder struct {
	MsgHash   []byte           //消息hash
	Lock      *sync.RWMutex    //
	Holder    []*MsgHolderOne  //消息持有者
	HolderTag []bool           //消息持有者同步标志，超时或者同步失败，设置为true
	IsRun     bool             //是否正在同步
	IsFinish  bool             //是否已经同步到了消息
	RunSignal chan bool        //继续运行信号
	Log       **zerolog.Logger //日志
}

func (this *MsgHolder) AddHolder(holder *MsgHolderOne) {
	this.Lock.Lock()
	have := false
	for _, one := range this.Holder {
		if bytes.Equal(one.Addr.Data(), holder.Addr.Data()) {
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
		if bytes.Equal(one.Addr.Data(), holder.Addr.Data()) {
			this.HolderTag[i] = true
			break
		}
	}
	this.Lock.RUnlock()
	return
}

func CreateMsgHolder(holder *MsgHolderOne, log **zerolog.Logger) *MsgHolder {
	msgHolder := MsgHolder{
		MsgHash:   holder.MsgHash,
		Lock:      new(sync.RWMutex),
		Holder:    make([]*MsgHolderOne, 0),
		HolderTag: make([]bool, 0),
		Log:       log,
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
	//(*this.Log).Info().Hex("同步广播消息", this.MsgHash).Send()

	//循环从消息持有者同步消息
	var message *MessageBase
	var holder *MsgHolderOne
	var tmpMes []*MessageBase
	var tmpSes []engine.Session

	var err error
	for {
		holder = this.GetHolder()
		if holder == nil || len(holder.Addr.Data()) <= 0 {
			//(*this.Log).Info().Msgf("not holder")
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
				tmpMes = append(tmpMes, tmp.Message)
				tmpSes = append(tmpSes, tmp.Session)
			}
			this.Lock.RUnlock()
		}

		//(*this.Log).Info().Hex("获取广播内容", this.MsgHash).Send()
		//获取广播内容
		bs, ERR := messageCenter.SyncMulticastMsg(holder.Session, &holder.MsgHash)
		if ERR.CheckFail() {
			//(*this.Log).Info().Str("ERR", ERR.String()).Send()
			continue
		}
		//(*this.Log).Info().Hex("获取广播内容", this.MsgHash).Int("内容长度", len(*bs)).Send()
		//先保存这个消息到数据库
		ERR = AddMessageCache(&holder.MsgHash, bs, messageCenter.levelDB)
		if ERR.CheckFail() {
			utils.Log.Info().Str("ERR", ERR.String()).Send()
			continue
		}
		message = holder.Message
		message.Content = *bs
		//(*this.Log).Info().Hex("获取广播内容", this.MsgHash).Send()
		break
	}
	//(*this.Log).Info().Hex("同步广播消息", this.MsgHash).Send()

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
	for i, _ := range tmpMes {
		//(*this.Log).Info().Msgf("回复消息:%s %+v", one.SenderAddr.B58String(), one.GetPacket())
		messageCenter.SendNeighborReplyMsg(tmpMes[i], nil, 0)
	}

	//异步继续广播给其他节点
	//广播给其他超级节点
	utils.Go(func() {
		//superSession := append(messageCenter.SessionManager.GetWhiteListSession(), messageCenter.SessionManager.GetSuperNodeSession()...)
		//proxySession := messageCenter.SessionManager.GetProxyNodeSession()
		//superNodes := append(messageCenter.nodeManager.GetLogicNodes(), messageCenter.nodeManager.GetNodesClient()...)
		//广播给代理对象
		//proxyNodes := messageCenter.nodeManager.GetProxyAll()

		//(*this.Log).Info().Msgf("转发广播消息：%+v", holder.ForwordMessage)
		messageCenter.broadcastsAll(&holder.ForwordMessage, 0)
		return
	}, *this.Log)

	message.MsgID = holder.Message.MsgID

	//(*this.Log).Info().Hex("同步广播消息", this.MsgHash).Send()
	//自己处理
	h, ok := messageCenter.router.GetHandler(message.MsgID)
	if !ok {
		(*this.Log).Info().Msgf("This broadcast message is not registered:%d", message.MsgID)
		return
	}
	h(message)
}

/*
去邻居节点同步广播消息
@param	id			AddressNet			同步对象的地址
@param	hash		[]byte				需要同步的消息hash值
@param	machineId	string				同步对象的设备机器Id，可以穿空串，如果传值，则会根据设备id，获取对应的session进行处理
@return	res			MessageMulticast	同步到的消息信息
@return	err			error				错误信息
*/
func (this *MessageCenter) SyncMulticastMsg(ss engine.Session, msgHash *[]byte) (*[]byte, utils.ERROR) {
	//this.Log.Info().Msgf("同步一条广播消息:%s", hex.EncodeToString(*msgHash))
	mn := NewMessageNeighbor(0, this.nodeManager.NodeSelf.IdInfo.Id, msgHash)
	bs, err := mn.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//优先从广播给我的节点获取广播消息
	resultBs, ERR := ss.SendWait(config.Version_multicast_body, bs, time.Minute)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//this.Log.Info().Msgf("同步一条广播消息:%s 内容长度:%d", hex.EncodeToString(*msgHash), len(*resultBs))
	//解析
	message, err := ParseMessageBase(resultBs) //
	if err != nil {
		(*this.Log).Error().Err(err).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	//this.Log.Info().Msgf("同步一条广播消息:%s 解析后内容长度:%d", hex.EncodeToString(*msgHash), len(message.Content))
	return &message.Content, utils.NewErrorSuccess()
}
