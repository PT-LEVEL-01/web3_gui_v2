package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/shirou/gopsutil/v3/mem"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/message_center/flood"
	"web3_gui/libp2parea/adapter/nodeStore"
	pgo_protos "web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

func RegisterMSG() {

	engine.Log.Info("---------- 注册区块链模块消息 ----------")
	// engine.RegisterMsg(config.MSGID_multicast_vote_recv, MulticastVote_recv) //接收投票旷工广播
	Area.Register_multicast(config.MSGID_multicast_vote_recv, MulticastVote_recv)      //接收投票旷工广播
	Area.Register_multicast(config.MSGID_multicast_blockhead, MulticastBlockHead_recv) //接收区块头广播
	Area.Register_neighbor(config.MSGID_heightBlock, FindHeightBlock)                  //查询邻居节点区块高度
	Area.Register_neighbor(config.MSGID_heightBlock_recv, FindHeightBlock_recv)        //查询邻居节点区块高度_返回
	Area.Register_neighbor(config.MSGID_getStartBlockHead, GetStartBlockHead)          //查询起始区块头

	Area.Register_neighbor(config.MSGID_getStartBlockHead_recv, GetStartBlockHead_recv)     //查询起始区块头_返回
	Area.Register_neighbor(config.MSGID_getBlockHeadVO, GetBlockHeadVO)                     //查询整个区块，包含头和交易
	Area.Register_neighbor(config.MSGID_getBlockHeadVO_recv, GetBlockHeadVO_recv)           //查询整个区块，包含头和交易_返回
	Area.Register_multicast(config.MSGID_multicast_transaction, MulticastTransaction_recv)  //接收交易广播
	Area.Register_neighbor(config.MSGID_getUnconfirmedBlock, GetUnconfirmedBlock)           //从邻居节点获取未确认的区块
	Area.Register_neighbor(config.MSGID_getUnconfirmedBlock_recv, GetUnconfirmedBlock_recv) //从邻居节点获取未确认的区块_返回
	Area.Register_neighbor(config.MSGID_multicast_return, MulticastReturn_recv)             //收到广播消息回复_返回
	Area.Register_neighbor(config.MSGID_getblockforwitness, GetBlockForWitness)             //从邻居节点获取指定见证人的区块
	Area.Register_neighbor(config.MSGID_getblockforwitness_recv, GetBlockForWitness_recv)   //从邻居节点获取指定见证人的区块_返回
	Area.Register_neighbor(config.MSGID_getDBKey_one, GetDBKeyOne)                          //查询远程数据库的key对应的value保存到本地数据库
	Area.Register_neighbor(config.MSGID_getDBKey_one_recv, GetDBKeyOne_recv)                //查询远程数据库的key对应的value保存到本地数据库_返回
	Area.Register_multicast(config.MSGID_multicast_find_witness, MulticastWitness)          //广播寻找见证人地址
	Area.Register_p2p(config.MSGID_multicast_find_witness_recv, MulticastWitness_recv)      //广播寻找见证人地址 返回
	Area.Register_neighbor(config.MSGID_getBlockLastCurrent, GetBlockLastCurrent)           //从邻居节点获取已经确认的最高区块
	Area.Register_neighbor(config.MSGID_getBlockLastCurrent_recv, GetBlockLastCurrent_recv) //从邻居节点获取已经确认的最高区块_返回

	Area.Register_neighbor(config.MSGID_multicast_witness_blockhead, MulticastBlockHeadHash)              //接收见证人之间的区块广播
	Area.Register_neighbor(config.MSGID_multicast_witness_blockhead_recv, MulticastBlockHeadHash_recv)    //接收见证人之间的区块广播_返回
	Area.Register_neighbor(config.MSGID_multicast_witness_blockhead_get, GetMulticastBlockHead)           //接收见证人之间的区块广播
	Area.Register_neighbor(config.MSGID_multicast_witness_blockhead_get_recv, GetMulticastBlockHead_recv) //接收见证人之间的区块广播_返回

	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_blockhead, UniformityMulticastBlockHeadHash)            //接收见证人之间的区块广播
	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_blockhead_recv, UniformityMulticastBlockHeadHash_recv)  //接收见证人之间的区块广播_返回
	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_get, UniformityGetMulticastBlockHead)             //接收见证人之间的区块同步
	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_get_recv, UniformityGetMulticastBlockHead_recv)   //接收见证人之间的区块同步_返回
	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_import, UniformityMulticastBlockImport)           //接收见证人之间的区块导入指令
	Area.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_import_recv, UniformityMulticastBlockImport_recv) //接收见证人之间的区块导入指令_返回
	Area.Register_neighbor(config.MSGID_multicast_blockhead_sign, MulticastBlockHeadSign)                                    //接收区块头签名广播
	Area.Register_neighbor(config.MSGID_multicast_blockhead_sign_recv, MulticastBlockHeadSign_recv)                          //接收区块头签名广播_返回

	Area.Register_p2p(config.MSGID_p2p_block_ack, P2pBlockAck)           //接收区块确认广播
	Area.Register_p2p(config.MSGID_p2p_block_ack_recv, P2pBlockAck_recv) //接收区块确认广播_返回

	Area.Register_neighbor(config.MSGID_getBlockHeadVOMore, GetBlockHeadVOMore)           //一次查询多个区块，包含头和交易
	Area.Register_neighbor(config.MSGID_getBlockHeadVOMore_recv, GetBlockHeadVOMore_recv) //一次查询多个区块，包含头和交易_返回

	// flowControllerWaiteSecount()
	// flowControllerWaiteSecountToo()
	Area.Register_p2p(config.MSGID_p2p_node_get_addrnonce, P2pGetAddrNonce)           // 接受 通过地址查nonce
	Area.Register_p2p(config.MSGID_p2p_node_get_addrbalance, P2pGetAddrBal)           // 接受 通过地址查余额
	Area.Register_p2p(config.MSGID_p2p_node_get_addr_lock_balance, P2pGetAddrLockBal) // 接受 通过地址查余额

	Area.Register_neighbor(config.MSGID_GET_FULL_NODE, GetFullNode) // 查询节点保存的全节点地址
	//Area.Register_neighbor(config.MSGID_GET_FULL_NODE_REV, GetFullNode_recv) //

	Area.Register_p2p(config.MSGID_HEIGHTBLOCK, FindHeightBlockToLightNode)     //查询邻居节点区块高度
	Area.Register_p2p(config.MSGID_GETSTARTBLOCKHEAD, GetStartBlockHeadToLight) //查询起始区块头(给轻节点使用，轻节点只能走p2p消息)

	Area.Register_p2p(config.MSGID_multicast_swap_transaction, MulticastSwapTransaction) //接收swap交易广播
}

type FullNodeInfo struct {
	Address      string
	CurrentBlock uint64
	HighestBlock uint64
}

func GetFullNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	add := Area.GetNetId().B58String()
	highestBlock := GetHighestBlock()
	currentBlock := GetLongChain().GetCurrentBlock()
	var b = []FullNodeInfo{{Address: add, HighestBlock: highestBlock, CurrentBlock: currentBlock}}
	marshal, err := json.Marshal(b)
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_GET_FULL_NODE_REV, nil, msg.Session)
		return
	}
	Area.SendNeighborReplyMsg(message, config.MSGID_GET_FULL_NODE_REV, &marshal, msg.Session)
}

type BlockForWitness struct {
	GroupHeight uint64             //见证人组高度
	Addr        crypto.AddressCoin //见证人地址
}

func (this *BlockForWitness) Proto() (*[]byte, error) {
	bwp := go_protos.BlockForWitness{
		GroupHeight: this.GroupHeight,
		Addr:        this.Addr,
	}
	bs, err := bwp.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, nil
	// return bwp.Marshal()
}

func ParseBlockForWitness(bs *[]byte) (*BlockForWitness, error) {
	if bs == nil {
		return nil, nil
	}
	bwp := new(go_protos.BlockForWitness)
	err := proto.Unmarshal(*bs, bwp)
	if err != nil {
		return nil, err
	}
	bw := &BlockForWitness{
		GroupHeight: bwp.GroupHeight,
		Addr:        bwp.Addr,
	}
	return bw, nil
}

/*
从邻居节点获取指定见证人的区块
*/
func GetBlockForWitness(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// engine.Log.Info("recv GetBlockForWitness message")
	bs := []byte{}
	if message.Body.Content == nil {
		engine.Log.Warn("GetBlockForWitness message.Body.Content is nil")
		Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	bfw, err := ParseBlockForWitness(message.Body.Content)
	if err != nil {
		engine.Log.Warn("GetBlockForWitness decoder error: %s", err.Error())
		Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	// bfw := new(BlockForWitness)
	// // var jso = jsoniter.ConfigCompatibleWithStandardLibrary
	// // err := json.Unmarshal(*message.Body.Content, bfw)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(bfw)
	// if err != nil {
	// 	engine.Log.Warn("GetBlockForWitness decoder error: %s", err.Error())
	// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
	// 	return
	// }

	witnessGroup := GetLongChain().WitnessChain.WitnessGroup
	for witnessGroup.Height > bfw.GroupHeight && witnessGroup.PreGroup != nil {
		witnessGroup = witnessGroup.PreGroup
	}
	for witnessGroup.Height < bfw.GroupHeight && witnessGroup.NextGroup != nil {
		witnessGroup = witnessGroup.NextGroup
	}
	if witnessGroup.Height != bfw.GroupHeight {
		engine.Log.Warn("GetBlockForWitness not find group height")
		Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	//找到了组高度，遍历这个组中的见证人
	for _, one := range witnessGroup.Witness {
		if !bytes.Equal(*one.Addr, bfw.Addr) {
			continue
		}
		//找到了这个见证人
		//这个见证人还未出块
		if one.Block == nil {
			Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		} else {
			bh, tx, err := one.Block.LoadTxs()
			if err != nil {
				Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
			}
			bhvo := CreateBlockHeadVO(nil, bh, *tx)
			newbs, err := bhvo.Proto() // bhvo.Json()
			if err != nil {
				engine.Log.Warn("GetBlockForWitness bhvo encoding json error: %s", err.Error())
			}
			bs = *newbs
			engine.Log.Info("GetBlockForWitness bhvo encoding Success")
			Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		}
		break
	}
	engine.Log.Warn("GetBlockForWitness not find this witness")
	Area.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)

}

/*
从邻居节点获取指定见证人的区块_返回
*/
func GetBlockForWitness_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// bs := []byte("ok")
	// flood.ResponseWait(config.CLASS_wallet_getblockforwitness, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
收到广播消息回复_返回
*/
func MulticastReturn_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bs := []byte("ok")
	// flood.ResponseWait(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), &bs)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), &bs)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), &bs)

}

/*
接收备用见证人投票广播
*/
func MulticastVote_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收见证人投票广播")
	//	engine.NLog.Debug(engine.LOG_console, "接收投票旷工广播")
	//	log.Println("接收投票旷工广播", msg.Session.GetName())

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// //自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// if !message.CheckSendhash() {
	// 	return
	// }

	//	fmt.Println("--接收见证人投票广播", hex.EncodeToString(bt.Deposit))

	//TODO 先验证选票是否合法

	//TODO 再判断是否要为他投票

	//	bt := ParseBallotTicket(message.Body.Content)
	//	AddBallotTicket(bt)

	// //继续广播给其他节点
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	//		mh := utils.Multihash(*message.Body.Content)
	// 	ids := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 	for _, one := range ids {
	// 		//			log.Println("发送给", one.B58String())
	// 		if ss, ok := engine.GetSession(config.AreaName[:],one.B58String()); ok {
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// 	//广播给代理对象
	// 	pids := nodeStore.GetProxyAll()
	// 	for _, one := range pids {
	// 		if ss, ok := engine.GetSession(config.AreaName[:],one); ok {
	// 			//				ss.Send(MSGID_multicast_online_recv, &msg.Data, false)
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// }

}

/*
接收区块广播
当矿工挖到一个新的区块后，会广播这个区块
*/
func MulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	chain := GetLongChain()
	if chain == nil {
		return
	}
	if !chain.SyncBlockFinish {
		return
	}

	// engine.NLog.Debug(engine.LOG_console, "接收区块头广播")
	//	log.Println("接收区块头广播", msg.Session.GetName())

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("This message is repeated")
	// 	return
	// }

	// fmt.Println("接收区块头广播", string(*message.Body.Content))

	bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	if err != nil {
		// fmt.Println("解析区块广播错误", err)
		engine.Log.Warn("Parse block broadcast error: %s", err.Error())
		return
	}
	// engine.Log.Info("接收到区块广播消息：%d %d %s", bhVO.BH.GroupHeight, bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash))

	if bhVO.BH.Height <= chain.GetCurrentBlock() {
		//engine.Log.Info("height:%d,current:%d", bhVO.BH.Height, chain.GetCurrentBlock()+1)
		engine.Log.Info("Multicast Block height too low **********************send:%s slef:%s height:%d,current:%d", message.Head.Sender.B58String(), config.Area.GetNetId().B58String(), bhVO.BH.Height, chain.GetCurrentBlock()+1)
		return
	}

	// sessionName := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// jsonBs, _ := bhVO.Json()
	// engine.Log.Info("Receiving block head broadcast from %s", sessionName.B58String())

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	bhashkey := config.BuildBlockHead(bhVO.BH.Hash)
	exist, err := db.LevelDB.CheckHashExist(bhashkey)
	if err != nil {
		engine.Log.Warn("this block exist error:%s", err.Error())
		return
	}
	if exist {
		// engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	//bhVO.BH.BuildBlockHash()

	//轻节点模式
	if config.Model == config.Model_light {
		GetLongChain().Temp = bhVO
		//直接保存区块
		_, err := SaveBlockHead(bhVO)
		if err != nil {
			engine.Log.Warn("save block error %s", err.Error())
			return
		}
		return
	}

	go chain.AddBlockOther(bhVO)
	//	go ImportBlock(bhVO)

	// //继续广播给其他节点
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	//		mh := utils.Multihash(*message.Body.Content)
	// 	ids := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 	for _, one := range ids {
	// 		//			log.Println("发送给", one.B58String())
	// 		if ss, ok := engine.GetSession(config.AreaName[:],one.B58String()); ok {
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// 	//广播给代理对象
	// 	pids := nodeStore.GetProxyAll()
	// 	for _, one := range pids {
	// 		if ss, ok := engine.GetSession(config.AreaName[:],one); ok {
	// 			//				ss.Send(MSGID_multicast_online_recv, &msg.Data, false)
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}
	// }

}

/*
接收邻居节点区块高度
*/
func FindHeightBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收邻居节点区块高度")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块高度")
	//	log.Println("接收邻居节点区块高度", msg.Session.GetName())
	chain := GetLongChain()
	if chain == nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_heightBlock_recv, nil, msg.Session)
		engine.Log.Warn("return nil message FindHeightBlock")
		return
	}

	if message.Body.Content != nil && len(*message.Body.Content) > 0 && !bytes.Equal(*message.Body.Content, config.StartBlockHash) {
		Area.SendNeighborReplyMsg(message, config.MSGID_heightBlock_recv, nil, msg.Session)
		engine.Log.Warn("return nil message FindHeightBlock")
		return
	}

	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetStartingBlock())
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetCurrentBlock())
	bs := dataBuf.Bytes()

	Area.SendNeighborReplyMsg(message, config.MSGID_heightBlock_recv, &bs, msg.Session)

}

/*
接收邻居节点区块高度
*/
func FindHeightBlockToLightNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetStartingBlock())
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetCurrentBlock())
	bs := dataBuf.Bytes()

	Area.SendP2pReplyMsg(message, config.MSGID_HEIGHTBLOCK_RECV, &bs)

}

/*
接收邻居节点区块高度_返回
*/
func FindHeightBlock_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(mc.CLASS_findHeightBlock, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if ok {
		flood.GroupWaitRecv.ResponseBytes(fmt.Sprintf("%d", ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	}

}

/*
接收邻居节点起始区块头查询
*/
func GetStartBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("++++接收邻居节点区块头查询")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块头查询")
	//	log.Println("接收邻居节点区块头查询", msg.Session.GetName())

	// var bs *[]byte
	bhash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		return
	}
	// bs = bhash

	chainInfo := ChainInfo{
		StartBlockHash: *bhash,                  //创始区块hash
		HightBlock:     forks.GetHighestBlock(), //最高区块
	}

	bs, err := chainInfo.Proto() // json.Marshal(chainInfo)
	if err != nil {
		return
	}

	Area.SendNeighborReplyMsg(message, config.MSGID_getStartBlockHead_recv, &bs, msg.Session)

}

/*
接收邻居节点起始区块头查询
*/
func GetStartBlockHeadToLight(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bhash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		return
	}
	chainInfo := ChainInfo{
		StartBlockHash: *bhash,                  //创始区块hash
		HightBlock:     forks.GetHighestBlock(), //最高区块
	}

	bs, err := chainInfo.Proto() // json.Marshal(chainInfo)
	if err != nil {
		return
	}

	Area.SendP2pReplyMsg(message, config.MSGID_GETSTARTBLOCKHEAD_RECV, &bs)

}

/*
区块同步信息
*/
type ChainInfo struct {
	StartBlockHash []byte //创始区块hash
	HightBlock     uint64 //最高区块
}

func (this *ChainInfo) Proto() ([]byte, error) {
	cip := go_protos.ChainInfo{
		StartBlockHash: this.StartBlockHash,
		HightBlock:     this.HightBlock,
	}
	return cip.Marshal()
}

func ParseChainInfo(bs *[]byte) (*ChainInfo, error) {
	if bs == nil {
		return nil, nil
	}
	cip := new(go_protos.ChainInfo)
	err := proto.Unmarshal(*bs, cip)
	if err != nil {
		return nil, err
	}

	ci := ChainInfo{
		StartBlockHash: cip.StartBlockHash, //创始区块hash
		HightBlock:     cip.HightBlock,     //最高区块
	}

	return &ci, nil
}

/*
接收邻居节点起始区块头查询_返回本次从邻居节点同步区块
*/
func GetStartBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(mc.CLASS_getBlockHead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if ok {
		flood.GroupWaitRecv.ResponseBytes(fmt.Sprintf("%d", ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	}
}

/*
	接收查询交易
	此接口频繁调用占用带宽较多，需要控制好流量
*/
// var flowController = make(chan bool, 1)
// var flowControllerRelax = make(chan bool, 1)

// //等待1秒钟
// func flowControllerWaiteSecount() {
// 	go func() {
// 		for range time.NewTicker(config.Wallet_sync_block_interval_time).C {
// 			select {
// 			case flowController <- false:
// 			default:
// 			}
// 		}
// 	}()

//		go func() {
//			for range time.NewTicker(config.Wallet_sync_block_interval_time_relax).C {
//				select {
//				case flowControllerRelax <- false:
//				default:
//				}
//			}
//		}()
//	}
func GetBlockHeadVO(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetTransaction, config.Wallet_sync_block_interval_time)
	utils.SetTimeToken(config.TIMETOKEN_GetTransactionRelax, config.Wallet_sync_block_interval_time_relax)
	//未同步完成则不给别人同步，因为在启动节点的时候，大量的请求需要处理，导致自己启动非常慢，缺点会导致节点为同步完成之前，连接它的节点会获取不到创始块而崩溃
	//未同步完成，只给同步初始区块
	// if !GetSyncFinish() && !bytes.Equal(config.StartBlockHash, *message.Body.Content) {
	// 	// engine.Log.Warn("未同步完成则不分配奖励")
	// 	engine.Log.Warn("If the synchronization is not completed, no reward will be allocated")
	// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	// 	return
	// }
	chain := GetLongChain()
	if chain == nil {
		utils.Log.Info().Str("返回了", "").Send()
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}
	//节点限速
	if !chain.SyncBlockFinish {
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if CheckNameStore() { //判断是否是存储节点
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if chain.WitnessChain.FindWitness(Area.Keystore.GetCoinbase().Addr) { //是见证人
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else {
		utils.GetTimeToken(config.TIMETOKEN_GetTransactionRelax, true)
		// <-flowControllerRelax
	}
	//这里查看内存，控制速度
	memInfo, _ := mem.VirtualMemory()
	if memInfo.UsedPercent > 92 {
		time.Sleep(time.Second)
	}
	// <-flowController

	//获取存储超级节点地址
	// nameinfo := name.FindName(config.Name_store)
	// if nameinfo != nil {
	// 	//判断域名是否是自己的
	// 	have := false
	// 	for _, one := range nameinfo.NetIds {
	// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, one) {
	// 			have = true
	// 			break
	// 		}
	// 	}
	// 	//在列表里，则退出，存储节点避免网络占用，不作同步服务器
	// 	if have {
	// 		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	// 		return
	// 	}
	// }

	//	timeout := time.NewTimer(time.Millisecond * 20) //20毫秒
	//	select {
	//	case <-timeout.C:
	//		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	//		return
	//	case <-flowController:
	//		timeout.Stop()
	//	}
	// defer flowControllerWaiteSecount()

	//收到邻居节点查询区块消息
	// engine.Log.Info("Received neighbor node query block message")

	netid := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// addrNet := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// engine.Log.Info("Received neighbor node query block message %s", netid.B58String())

	if message.Body.Content == nil {
		utils.Log.Info().Str("返回了", "").Send()
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}

	bhvo := new(BlockHeadVO)

	//通过区块hash查找区块头
	// bs, err := db.Find(*message.Body.Content)
	bh, err := LoadBlockHeadByHash(message.Body.Content)
	if err != nil {
		utils.Log.Info().Str("返回了", "").Send()
		// engine.Log.Error("querying transaction or block Error: %s", err.Error())
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	} else {
		//查询本区块是否被确认
		// bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(bh.Height))))
		// if err != nil || bytes.Equal(*bhash, bh.Hash) {
		// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		// 	engine.Log.Info("block not confirm")
		// 	return
		// }

		// fmt.Println("查询区块或交易结果2", len(*bs))
		// bh, err := ParseBlockHead(bs)
		// if err != nil {
		// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		// 	// engine.Log.Info("return error message")
		// 	return
		// }
		bhvo.BH = bh
		bhvo.Txs = make([]TxItr, 0, len(bh.Tx))
		for _, one := range bh.Tx {
			// txOne, err := FindTxBase(one, hex.EncodeToString(one))
			// txOne, err := FindTxBase(one)
			txOne, err := LoadTxBase(one)
			// bs, err := db.Find(one)
			// if err != nil {
			// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
			// 	return
			// }
			// txOne, err := ParseTxBase(ParseTxClass(one), bs)
			if err != nil {
				utils.Log.Info().Str("返回了", "").Send()
				Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
				engine.Log.Info("return error message")
				return
			}
			bhvo.Txs = append(bhvo.Txs, txOne)
		}
	}
	if bhvo.BH.Nextblockhash == nil {

		if GetHighestBlock() > bhvo.BH.Height+1 {
			engine.Log.Info("neighbor %s find next block %d hash nil. hight:%d", netid.B58String(), bhvo.BH.Height, GetHighestBlock())
		}
		//TODO Nextblockhash为空，则补充一个，临时解决方案
		tempGroup := GetLongChain().WitnessChain.WitnessGroup
		for tempGroup != nil {

			if tempGroup.Height < bhvo.BH.GroupHeight {
				break
			}
			if tempGroup.Height > bhvo.BH.GroupHeight {
				tempGroup = tempGroup.PreGroup
				continue
			}
			for _, one := range tempGroup.Witness {
				if one.Block == nil {
					continue
				}
				if one.Block.Height == bhvo.BH.Height {
					if one.Block.NextBlock != nil {
						//engine.Log.Info("neighbor %s find next block %d hash nil.", netid.B58String(), bhvo.BH.Height)
						bhvo.BH.Nextblockhash = one.Block.NextBlock.Id
					}
					tempGroup = nil
					break
				}
			}
			break
		}
	}

	bs, err := bhvo.Proto() // bhvo.Json()
	if err != nil {
		utils.Log.Info().Str("返回了", "").Send()
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return json fialt message")
		return
	}
	utils.Log.Info().Str("返回了", "").Send()
	err = Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error: %s", err.Error())
	} else {
		// engine.Log.Info("return success message: %s", netid.B58String())
	}
}

/*
接收查询交易_返回
*/
func GetBlockHeadVO_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if ok {
		flood.GroupWaitRecv.ResponseBytes(fmt.Sprintf("%d", ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	}

}

func GetBlockHeadVOMore(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetTransaction, config.Wallet_sync_block_interval_time)
	utils.SetTimeToken(config.TIMETOKEN_GetTransactionRelax, config.Wallet_sync_block_interval_time_relax)
	//未同步完成则不给别人同步，因为在启动节点的时候，大量的请求需要处理，导致自己启动非常慢，缺点会导致节点为同步完成之前，连接它的节点会获取不到创始块而崩溃
	//未同步完成，只给同步初始区块
	// if !GetSyncFinish() && !bytes.Equal(config.StartBlockHash, *message.Body.Content) {
	// 	// engine.Log.Warn("未同步完成则不分配奖励")
	// 	engine.Log.Warn("If the synchronization is not completed, no reward will be allocated")
	// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	// 	return
	// }
	chain := GetLongChain()
	if chain == nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}
	//节点限速
	if !chain.SyncBlockFinish {
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if CheckNameStore() { //判断是否是存储节点
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if chain.WitnessChain.FindWitness(Area.Keystore.GetCoinbase().Addr) { //是见证人
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else {
		utils.GetTimeToken(config.TIMETOKEN_GetTransactionRelax, true)
		// <-flowControllerRelax
	}
	//这里查看内存，控制速度
	memInfo, _ := mem.VirtualMemory()
	if memInfo.UsedPercent > 92 {
		time.Sleep(time.Second)
	}

	//netid := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// addrNet := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// engine.Log.Info("Received neighbor node query block message %s", netid.B58String())

	if message.Body.Content == nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}

	req := new(go_protos.GetBlockHeadVOMoreRequest)
	err := proto.Unmarshal(*message.Body.Content, req)
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("GetBlockHeadVOMoreRequest Unmarshal err:%s", err.Error())
		return
	}

	if req.Length > config.Deep_Cycle_Sync_Block_More_Length_server {
		req.Length = config.Deep_Cycle_Sync_Block_More_Length_server
	}

	bhvos, err := LoadBlockHeadByHeightMore(req.Height, req.Length, req.BHash)
	//bhvos, err := LoadBlockHeadByHeightMore(req.Height, req.Height+req.Length)

	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("LoadBlockHeadByHeightMore err:%s", err.Error())
		return
	}
	rbhats := new(go_protos.RepeatedBlockHeadAndTxs)
	for _, bh := range bhvos {
		bhvo := &BlockHeadVO{
			BH:  bh,
			Txs: make([]TxItr, 0, len(bh.Tx)),
		}
		checkTx := true
		for _, one := range bh.Tx {
			txOne, err := LoadTxBase(one)
			if err != nil {
				//Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
				engine.Log.Info("getBlockHeadVO_recv not found tx:%s", hex.EncodeToString(one))
				checkTx = false
				break
			}
			bhvo.Txs = append(bhvo.Txs, txOne)
		}
		if !checkTx {
			break
		}

		if bhvo.BH.Nextblockhash == nil {
			//if GetHighestBlock() > bhvo.BH.Height+1 {
			//	engine.Log.Info("neighbor %s find next block %d hash nil. hight:%d", netid.B58String(), bhvo.BH.Height, GetHighestBlock())
			//}
			//TODO Nextblockhash为空，则补充一个，临时解决方案
			tempGroup := GetLongChain().WitnessChain.WitnessGroup
			for tempGroup != nil {

				if tempGroup.Height < bhvo.BH.GroupHeight {
					break
				}
				if tempGroup.Height > bhvo.BH.GroupHeight {
					tempGroup = tempGroup.PreGroup
					continue
				}
				for _, one := range tempGroup.Witness {
					if one.Block == nil {
						continue
					}
					if one.Block.Height == bhvo.BH.Height {
						if one.Block.NextBlock != nil {
							//engine.Log.Info("neighbor %s find next block %d hash nil.", netid.B58String(), bhvo.BH.Height)
							bhvo.BH.Nextblockhash = one.Block.NextBlock.Id
						}
						tempGroup = nil
						break
					}
				}
				break
			}
		}

		bs, err := bhvo.Proto() // bhvo.Json()
		if err != nil {
			Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
			engine.Log.Info("return json fialt message")
			return
		}

		bhatp := new(go_protos.BlockHeadAndTxs)
		err = proto.Unmarshal(*bs, bhatp)
		if err != nil {
			return
		}
		rbhats.Bhat = append(rbhats.Bhat, bhatp)
		//engine.Log.Info("************************************lastHeight %d", bh.Height)
	}
	//engine.Log.Info("一共找到%d个区块************************************Height %d Length %d", len(rbhats.Bhat), req.Height, req.Height+req.Length)
	rbhatsB, err := rbhats.Marshal()
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, nil, msg.Session)
		engine.Log.Info("return json fialt message")
		return
	}

	err = Area.SendNeighborReplyMsg(message, config.MSGID_getBlockHeadVO_recv, &rbhatsB, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error: %s", err.Error())
	} else {
		// engine.Log.Info("return success message: %s", netid.B58String())
	}
}

func GetBlockHeadVOMore_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if ok {
		flood.GroupWaitRecv.ResponseBytes(fmt.Sprintf("%d", ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	}

}

/*
接收交易广播
*/
func MulticastTransaction_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//检查内存
	if CheckOutOfMemory() {
		engine.Log.Error("Memory is too height")
		return
	}

	config.GetRpcRate("MulticastTx", true)

	//engine.Log.Debug("MulticastTransaction_recv %s", string(*message.Body.Content))

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("这个消息重复了")
	// 	return
	// }
	if config.Model == config.Model_light {
		//engine.Log.Error("tx failed ,config.model is light")
		return
	}

	//自己处理
	txbase, err := ParseTxBaseProto(0, message.Body.Content) // ParseTxBase(0, message.Body.Content)
	if err != nil {
		// fmt.Println("解析广播的交易错误", err)
		engine.Log.Error("Broadcast transaction format error %s", err.Error())
		return
	}

	// engine.Log.Info("交易hash:%s", hex.EncodeToString(*txbase.GetHash()))
	//if !TxCheckCond.WaitCheck(txbase) {
	//	engine.Log.Error("tx failed;TxCheckCond.WaitCheck is false")
	//	return
	//}

	//engine.Log.Debug("Broadcast transaction received %s", hex.EncodeToString(*txbase.GetHash()))
	// bs, _ := json.Marshal(txbase.GetVOJSON())
	// engine.Log.Debug("接收到广播交易 %s", string(bs))
	txbase.BuildHash()
	// engine.Log.Debug("接收到广播交易 222 %s", txbase.GetHashStr())
	//判断区块是否同步完成，如果没有同步完成则无法验证交易合法性，则不保存交易

	chain := GetLongChain()
	if chain == nil {
		engine.Log.Error("chain is nil")
		return
	}
	if !chain.SyncBlockFinish {
		engine.Log.Error("chain is not SyncBlockFinish")
		return
	}

	//处理多签交易签名过程
	if mtx, ok := txbase.(DefaultMultTx); ok {
		if err := mtx.HandleMultTx(txbase); err != nil {
			engine.Log.Warn("Multsign Handle: %s %s", hex.EncodeToString(*txbase.GetHash()), err.Error())
			return
		}
	}

	//if txbase.GetGas() < config.Wallet_tx_gas_min {
	//	engine.Log.Error("gas is little")
	//}

	// 支付类检查是否免gas
	if err := CheckTxPayFreeGas(txbase); err != nil {
		engine.Log.Error("gas is little")
		return
	}

	checkTxQueue <- txbase

	// // if !GetSyncFinish() {
	// // 	// engine.Log.Warn("未同步完成则无法验证交易合法性")
	// // 	return
	// // }
	// if len(*txbase.GetVin()) > config.Mining_pay_vin_max {
	// 	//交易太大了
	// 	engine.Log.Warn(config.ERROR_pay_vin_too_much.Error())
	// 	return
	// }
	// //验证交易
	// if err := txbase.CheckLockHeight(GetHighestBlock()); err != nil {
	// 	// engine.Log.Warn("验证交易锁定高度失败")
	// 	engine.Log.Warn("Failed to verify transaction lock height")
	// 	return
	// }
	// // txbase.CheckFrozenHeight(GetHighestBlock())

	// //加载相关交易到缓存
	// keys := make(map[string]uint64, 0) //记录加载了哪些交易到缓存
	// for _, one := range *txbase.GetVin() {
	// 	//已经有了就不用重复查询了
	// 	if _, ok := TxCache.FindTxInCache(one.Txid); !ok {
	// 		continue
	// 	}

	// 	// txItr, err := FindTxBase(one.Txid)
	// 	txItr, err := LoadTxBase(one.Txid)
	// 	if err != nil {
	// 		return
	// 	}
	// 	TxCache.AddTxInCache(one.Txid, txItr)
	// 	// key := utils.Bytes2string(one.Txid) //one.GetTxidStr()
	// 	keys[utils.Bytes2string(one.Txid)] = one.Vout
	// 	// keys = append(keys, key)
	// }

	// // bs, _ := txbase.Json()
	// // engine.Log.Info("交易\n%s", string(*bs))

	// if GetHighestBlock() > config.Mining_block_start_height+config.Mining_block_start_height_jump {
	// 	if err := txbase.Check(); err != nil {
	// 		//交易不合法，则不发送出去
	// 		//验证未通过，删除缓存
	// 		// for k, v := range keys {
	// 		// 	TxCache.RemoveTxInCache(k, v)
	// 		// }
	// 		runtime.GC()
	// 		engine.Log.Warn("Failed to verify transaction signature %s %s", hex.EncodeToString(*txbase.GetHash()), err.Error())
	// 		return
	// 	}
	// 	runtime.GC()
	// }

	// //判断是否有重复交易，并且检查是否有无效区块的交易的标记
	// if db.LevelDB.CheckHashExist(*txbase.GetHash()) && !db.LevelDB.CheckHashExist(config.BuildTxNotImport(*txbase.GetHash())) {
	// 	//验证未通过，删除缓存
	// 	// for k, v := range keys {
	// 	// 	TxCache.RemoveTxInCache(k, v)
	// 	// }
	// 	engine.Log.Warn("Transaction hash collision is the same %s", hex.EncodeToString(*txbase.GetHash()))
	// 	return
	// }

	// forks.GetLongChain().transactionManager.AddTx(txbase)

}

// var flowControllerToo = make(chan bool, 1)

// //等待1秒钟
// func flowControllerWaiteSecountToo() {
// 	go func() {
// 		for range time.NewTicker(time.Second).C {
// 			select {
// 			case flowControllerToo <- false:
// 			default:
// 			}
// 		}
// 	}()
// }

/*
从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetUnconfirmedBlock, config.Wallet_sync_block_interval_time)
	utils.GetTimeToken(config.TIMETOKEN_GetUnconfirmedBlock, true)

	// <-flowControllerToo
	height := utils.BytesToUint64(*message.Body.Content)

	engine.Log.Info("Get the unconfirmed block, the height of this synchronization block %d", height)

	witnessGroup := GetLongChain().WitnessChain.WitnessGroup

	group := witnessGroup
	for {
		group = group.PreGroup
		if group.BlockGroup != nil {
			break
		}
	}
	block := group.BlockGroup.Blocks[0]
	for i := 0; i < config.Mining_group_max*5; i++ {
		if block == nil || block.Height == height {
			break
		}
		if block.Height > height {
			block = block.PreBlock
		}
		if block.Height < height {
			block = block.NextBlock
		}
	}

	var bs []byte
	bhvos := make([]*BlockHeadVO, 0)
	var err error
	//先查询已经确认的块
	for block != nil {
		bh, txs, e := block.LoadTxs()
		if e != nil {
			err = e
			break
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
		bhvos = append(bhvos, bhvo)
		block = block.NextBlock
	}
	if err == nil {
		for {
			if witnessGroup == nil {
				break
			}

			for _, one := range witnessGroup.Witness {
				if one.Block == nil {
					continue
				}

				// fmt.Println("这个block.id怎么会为空", one.Block, witnessGroup.Height)
				bh, txs, e := one.Block.LoadTxs()
				if e != nil {
					err = e
					break
				}
				bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
				bhvos = append(bhvos, bhvo)
				// block = block.NextBlock
			}
			if err != nil {
				break
			}
			witnessGroup = witnessGroup.NextGroup
		}
	}
	if err == nil {
		// bsOne, err := json.Marshal(bhvos)
		// if err == nil {
		// 	bs = bsOne
		// }

		rbsp := go_protos.RepeatedBytes{
			Bss: make([][]byte, 0),
		}
		for _, one := range bhvos {
			bsOne, err := one.Proto()
			if err != nil {
				return
			}
			rbsp.Bss = append(rbsp.Bss, *bsOne)
		}
		bsOne, err := rbsp.Marshal()

		// bsOne, err := json.Marshal(bhvos)
		if err == nil {
			bs = bsOne
		}
	}

	// for _, one := range bhvos {
	// 	engine.Log.Info("find block one:%d", one.BH.Height)
	// }
	err = Area.SendNeighborReplyMsg(message, config.MSGID_getUnconfirmedBlock_recv, &bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error %s", err.Error())
	}
}

/*
从邻居节点获取未确认的区块_返回
*/
func GetUnconfirmedBlock_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func GetDBKeyOne(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetTransaction, config.Wallet_sync_block_interval_time)
	utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
	// defer flowControllerWaiteSecount()
	// <-flowController

	if message.Body.Content == nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getDBKey_one_recv, nil, msg.Session)
		return
	}

	bs, err := db.LevelDB.Find(*message.Body.Content)
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getDBKey_one_recv, nil, msg.Session)
		return
	}

	err = Area.SendNeighborReplyMsg(message, config.MSGID_getDBKey_one_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction message Error: %s", err.Error())
	}
}

/*
接收查询交易_返回
*/
func GetDBKeyOne_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	// flood.ResponseWait(mc.CLASS_getTransaction_one,  utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收见证人地址发现广播
*/
func MulticastWitness(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//engine.Log.Info("接收见证人地址发现广播 %s", message.Head.Sender.B58String())
	chain := GetLongChain()
	if chain == nil {
		return
	}
	//判断自己是否在寻找的目标中
	bss := new(go_protos.RepeatedBytes)

	err := proto.Unmarshal(*message.Body.Content, bss)
	if err != nil {

		engine.Log.Error("proto unmarshal error %s", err.Error())
		return
	}
	have := false
	for _, one := range bss.Bss {
		_, ok := Area.Keystore.GetPukByAddr(crypto.AddressCoin(one))
		if ok {
			have = true
			break
		}
	}
	if !have {
		return
	}
	IsBackup := chain.WitnessChain.FindWitness(Area.Keystore.GetCoinbase().Addr)
	if !IsBackup {
		return
	}
	addr, port := Area.GetAddrAndPort()
	ipv4, err := utils.IPV4String2Long(addr)
	if err != nil {
		return
	}
	wlInfo := go_protos.WhileListInfo{
		TCPHost:  ipv4,
		TCPPort:  uint32(port),
		AddrCoin: Area.Keystore.GetCoinbase().Addr,
		AddrNet:  Area.GetNetId(),
	}
	bs, err := wlInfo.Marshal()
	if err != nil {
		engine.Log.Info("WhileListInfo error:%s", err.Error())
		return
	}
	_, _, _, err = Area.SendP2pMsg(config.MSGID_multicast_find_witness_recv, message.Head.Sender, &bs)
	if err != nil {
		engine.Log.Error("returning query transaction message Error: %s", err.Error())
	}

}

/*
接收见证人地址发现广播_返回
*/
func MulticastWitness_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//engine.Log.Info("接收见证人地址发现广播_返回")
	wlInfo := new(go_protos.WhileListInfo)
	err := proto.Unmarshal(*message.Body.Content, wlInfo)
	if err != nil {
		engine.Log.Error("proto unmarshal error %s", err.Error())
		return
	}
	ipv4Str, err := utils.IPV4Long2String(wlInfo.TCPHost)
	if err != nil {
		engine.Log.Info("IPV4Long2String error:%s", err.Error())
		return
	}
	ok := Area.AddWhiteList(wlInfo.AddrNet)
	if ok {
		addrCoin := crypto.AddressCoin(wlInfo.AddrCoin)
		if findWitnessAddrNet(&addrCoin) != nil {
			return
		}
		addrNet := nodeStore.AddressNet(wlInfo.AddrNet)
		addWitnessAddrNet(&addrCoin, &addrNet)
		return
	}
	_, err = Area.AddConnect(ipv4Str, uint16(wlInfo.TCPPort))
	if err != nil {
		engine.Log.Info("add connect error:%s", err.Error())
		return
	}
	addrCoin := crypto.AddressCoin(wlInfo.AddrCoin)
	addrNet := nodeStore.AddressNet(wlInfo.AddrNet)
	addWitnessAddrNet(&addrCoin, &addrNet)
}

/*
从邻居节点获取已经确认的最高区块
*/
func GetBlockLastCurrent(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if config.Model == config.Model_light {
		currentHeight := GetLongChain().GetCurrentBlock()
		bhash := LoadBlockHashByHeight(currentHeight)
		if bhash == nil {
			Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
			return
		}

		bhashkey := config.BuildBlockHead(*bhash)
		bs, err := db.LevelDB.Find(bhashkey)
		if err != nil {
			Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
			return
		}
		err = Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, bs, msg.Session)
		if err != nil {
			engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
		}
		return
	}

	_, block := GetLongChain().GetLastBlock()

	bh, err := block.Load()
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
		return
	}
	bs, err := bh.Proto()
	if err != nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
		return
	}
	err = Area.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	}
}

/*
从邻居节点获取已经确认的最高区块_返回
*/
func GetBlockLastCurrent_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

var syncHashLock = new(sync.Mutex)
var blockheadHashMap = make(map[string]int)

/*
接收见证人出块的区块hash
*/
func MulticastBlockHeadHash(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//engine.Log.Info("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffhhhhhhhhhhhhhhhhhhhhhhhhhhh")

	//engine.Log.Info("MulticastBlockHeadHash 11111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	chain := forks.GetLongChain()
	if !chain.SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		//engine.Log.Info("MulticastBlockHeadHash 222222222222222222222")
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
		return
	}

	// newmessage, err := message_center.ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	//广播消息头解析失败
	// 	engine.Log.Warn("Parsing of this broadcast header failed")
	// 	return
	// }
	// //解析包体内容
	// if err = newmessage.ParserContentProto(); err != nil {
	// 	engine.Log.Info("Content parsing of this broadcast message failed %s", err.Error())
	// 	return
	// }
	var bhVO *BlockHeadVO
	success := false
	syncHashLock.Lock()
	// _, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	_, err := GetMessageMulticast(*message.Body.Content)

	if err == nil {
		//engine.Log.Info("MulticastBlockHeadHash 333333333333333333333")
		//已经下载好了，就回复对方
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
	} else {
		//engine.Log.Info("MulticastBlockHeadHash 4444444444444444444444")
		//去下载
		addrNet := nodeStore.AddressNet(msg.Session.GetName())

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content, config.Mining_block_time*time.Second)
		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content, config.Mining_sync_timeout)

		// newmsg, _ := message_center.SendNeighborMsg(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content)
		// bs, err := flood.WaitRequest(config.CLASS_witness_get_blockhead, utils.Bytes2string(newmsg.Body.Hash), int64(4))
		if err != nil {
			//engine.Log.Info("MulticastBlockHeadHash 4444444444444444444444 111111111111111111")
			engine.Log.Warn("Receiving broadcast reply message error:%s :%s", addrNet.B58String(), err.Error())
			// failNode = append(failNode, broadcasts[j])
			// continue
		} else {
			//下载成功
			//engine.Log.Info("MulticastBlockHeadHash 4444444444444444444444 222222222222222222222")
			//验证同步到的消息
			mmp := new(pgo_protos.MessageMulticast)
			err = proto.Unmarshal(*bs, mmp)
			if err != nil {
				engine.Log.Error("proto unmarshal error %s", err.Error())
			} else {

				//解析获取到的交易
				bhVOmessage, err := message_center.ParserMessageProto(mmp.Head, mmp.Body, 0)
				if err != nil {
					engine.Log.Error("proto unmarshal error %s", err.Error())
				} else {

					err = bhVOmessage.ParserContentProto()
					if err != nil {
						engine.Log.Error("proto unmarshal error %s", err.Error())
					} else {

						bhVO, err = ParseBlockHeadVOProto(bhVOmessage.Body.Content) //ParseBlockHeadVO(message.Body.Content)
						if err != nil {
							engine.Log.Warn("Parse block broadcast error: %s", err.Error())
						} else {
							//回复消息
							Area.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
							//
							err = SetMessageMulticast(*message.Body.Content, mmp)
							// err = new(sqlite3_db.MessageCache).Add(*message.Body.Content, mmp.Head, mmp.Body)
							if err != nil {
								engine.Log.Error(":%s", err.Error())
							} else {
								//engine.Log.Info("MulticastBlockHeadHash 4444444444444444444444 vvvvvvvvvvvvvvvvvvvvv")
								success = true
							}
						}
					}
				}
			}
		}
	}
	syncHashLock.Unlock()

	if !success {
		//engine.Log.Info("MulticastBlockHeadHash 555555555555555555555555")
		return
	}
	//engine.Log.Info("MulticastBlockHeadHash 666666666666666666666")
	//广播给其他人
	whiltlistNodes := Area.NodeManager.GetWhiltListNodes() //nodeStore.GetWhiltListNodes()
	err = Area.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, message.Body.Content)
	if err != nil {
		// engine.Log.Info("7777777777777777777777")
		//广播超时，网络不好，则不导入这个区块
		return
	}
	// engine.Log.Info("888888888888888888888")
	// bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	// if err != nil {
	// 	// fmt.Println("解析区块广播错误", err)
	// 	engine.Log.Warn("Parse block broadcast error: %s", err.Error())
	// 	return
	// }

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	exist, err := db.LevelDB.CheckHashExist(bhVO.BH.Hash)
	if err != nil {
		engine.Log.Warn("this block exist error:%s", err.Error())
		return
	}
	if exist {
		engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	bhVO.BH.BuildBlockHash()
	go chain.AddBlockOther(bhVO)

	//广播区块
	go MulticastBlock(*bhVO)

	// MulticastBlockAndImport(bhVO)

	// err = message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, bs, msg.Session)
	// if err != nil {
	// 	engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	// }

}

/*
接收见证人出块的区块hash
*/
func MulticastBlockHeadHash_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_wallet_broadcast_return, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收见证人出块的区块hash
*/
func GetMulticastBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// if !forks.GetLongChain().SyncBlockFinish {
	// 	//区块未同步好，直接回复对方已经收到区块
	// 	return
	// }
	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	// messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	messageCache, err := GetMessageMulticast(*message.Body.Content)
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		// engine.Log.Info("GetMulticastBlockHead 2222222222222222222222")
		engine.Log.Error("find message hash error:%s", err.Error())
		return
	}

	mmp := pgo_protos.MessageMulticast{
		Head: messageCache.Head,
		Body: messageCache.Body,
	}

	content, err := mmp.Marshal()
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		engine.Log.Error(":%s", err.Error())
		return
	}
	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息
	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_get_recv, &content, msg.Session)

}

/*
接收见证人出块的区块hash_返回
*/
func GetMulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

var syncUniformityHashLock = new(sync.Mutex)

// var blockheadHashMap = make(map[string]int)

/*
接收见证人出块的区块hash
*/
func UniformityMulticastBlockHeadHash(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// second := utils.GetRandNum(10)
	// time.Sleep(time.Second * time.Duration(second))

	//engine.Log.Info("UniformityMulticastBlockHeadHash 11111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	ok := []byte("ok")
	if !forks.GetLongChain().SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		//engine.Log.Info("UniformityMulticastBlockHeadHash 222222222222222222222")
		Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)
		return
	}
	var bhVO *BlockHeadVO
	success := false
	syncUniformityHashLock.Lock()
	// _, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	_, err := GetMessageMulticast(*message.Body.Content)
	if err == nil {
		//engine.Log.Info("UniformityMulticastBlockHeadHash 333333333333333333333")
		//已经下载好了，就回复对方
		Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, &ok, msg.Session)
	} else {
		//engine.Log.Info("UniformityMulticastBlockHeadHash 4444444444444444444444")
		//去下载
		addrNet := nodeStore.AddressNet(msg.Session.GetName())

		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_uniformity_multicast_witness_block_get, &addrNet, message.Body.Content,
			config.Wallet_sync_block_timeout*time.Second)

		// newmsg, _ := message_center.SendNeighborMsg(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content)
		// bs, err := flood.WaitRequest(config.CLASS_uniformity_witness_get_blockhead, utils.Bytes2string(newmsg.Body.Hash), int64(4))
		if err != nil {
			//engine.Log.Info("UniformityMulticastBlockHeadHash 4444444444444444444444 111111111111111111")
			engine.Log.Warn("Timeout receiving broadcast reply message %s", addrNet.B58String())
			// failNode = append(failNode, broadcasts[j])
			// continue
		} else {

			//下载成功
			//engine.Log.Info("UniformityMulticastBlockHeadHash 4444444444444444444444 222222222222222222222")
			//验证同步到的消息
			mmp := new(pgo_protos.MessageMulticast)
			err = proto.Unmarshal(*bs, mmp)
			if err != nil {
				engine.Log.Error("proto unmarshal error %s", err.Error())
			} else {

				//解析获取到的交易
				bhVOmessage, err := message_center.ParserMessageProto(mmp.Head, mmp.Body, 0)
				if err != nil {
					engine.Log.Error("proto unmarshal error %s", err.Error())
				} else {

					err = bhVOmessage.ParserContentProto()
					if err != nil {
						engine.Log.Error("proto unmarshal error %s", err.Error())
					} else {

						bhVO, err = ParseBlockHeadVOProto(bhVOmessage.Body.Content) //ParseBlockHeadVO(message.Body.Content)
						if err != nil {
							engine.Log.Warn("Parse block broadcast error: %s", err.Error())
						} else {
							//回复消息
							//Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)

							// err = new(sqlite3_db.MessageCache).Add(*message.Body.Content, mmp.Head, mmp.Body)
							err = SetMessageMulticast(*message.Body.Content, mmp)
							if err != nil {
								engine.Log.Error(err.Error(), "")
							} else {
								//engine.Log.Info("UniformityMulticastBlockHeadHash 4444444444444444444444 vvvvvvvvvvvvvvvvvvvvv")
								success = true
							}
						}
					}
				}
			}
		}
	}
	syncUniformityHashLock.Unlock()

	if !success {
		Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)
		//engine.Log.Info("UniformityMulticastBlockHeadHash 555555555555555555555555")
		return
	}
	engine.Log.Info("接收到区块缓存消息：%d %d %s", bhVO.BH.GroupHeight, bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash))
	//engine.Log.Info("UniformityMulticastBlockHeadHash 666666666666666666666")
	//广播给其他人
	// whiltlistNodes := nodeStore.GetWhiltListNodes()
	// err = message_center.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, message.Body.Content)
	// if err != nil {
	// 	// engine.Log.Info("7777777777777777777777")
	// 	//广播超时，网络不好，则不导入这个区块
	// 	return
	// }
	// engine.Log.Info("888888888888888888888")
	// bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	// if err != nil {
	// 	// fmt.Println("解析区块广播错误", err)
	// 	engine.Log.Warn("Parse block broadcast error: %s", err.Error())
	// 	return
	// }

	if bhVO.BH.Height <= GetLongChain().GetCurrentBlock()+1 {
		engine.Log.Info("height:%d,current:%d", bhVO.BH.Height, GetLongChain().GetCurrentBlock()+1)
		engine.Log.Error("Multicast Cache Block height too low")
		return
	}

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	exist, err := db.LevelDB.CheckHashExist(bhVO.BH.Hash)
	if err != nil {
		engine.Log.Warn("this block exist error:%s", err.Error())
		return
	}
	if exist {
		engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true
	bhVO.BH.BuildBlockHash()

	AddBlockToCache(bhVO)

	Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, &ok, msg.Session)

	// CleanCurrSignBlock()

	//engine.Log.Info("直接导入:%d %d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	ImportBlockByCache(&bhVO.BH.Hash)

	//block ack
	//if len(Area.NodeManager.GetWhiltListNodes()) == 0 || ReqBlockAck(&bhVO.BH.Hash, time.Second) {
	//	engine.Log.Info("组高度：%d，验证成功可以导入!", bhVO.BH.GroupHeight)
	//	//清除签名区块缓存
	//	CleanCurrSignBlock()
	//	//清除上一个区块ack缓存
	//	TempBlockAck.DelBlockAck(&bhVO.BH.Previousblockhash)
	//	ImportBlockByCache(&bhVO.BH.Hash)
	//}

	//放入缓存

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	// go forks.AddBlockHead(bhVO)

	//广播区块
	// go MulticastBlock(*bhVO)

	// MulticastBlockAndImport(bhVO)

	// err = message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, bs, msg.Session)
	// if err != nil {
	// 	engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	// }

}

/*
接收见证人出块的区块hash
*/
func UniformityMulticastBlockHeadHash_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_uniformity_witness_multicas_blockhead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收见证人出块的区块hash
*/
func UniformityGetMulticastBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// second := utils.GetRandNum(10)
	// time.Sleep(time.Second * time.Duration(second))
	// if !forks.GetLongChain().SyncBlockFinish {
	// 	//区块未同步好，直接回复对方已经收到区块
	// 	return
	// }
	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	// messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	messageCache, err := GetMessageMulticast(*message.Body.Content)
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		// engine.Log.Info("GetMulticastBlockHead 2222222222222222222222")
		engine.Log.Error("find message hash error:%s", err.Error())
		return
	}

	mmp := pgo_protos.MessageMulticast{
		Head: messageCache.Head,
		Body: messageCache.Body,
	}

	content, err := mmp.Marshal()
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		engine.Log.Error(err.Error(), "")
		return
	}
	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息
	Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_get_recv, &content, msg.Session)

}

/*
接收见证人出块的区块hash_返回
*/
func UniformityGetMulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_uniformity_witness_get_blockhead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收见证人导入区块命令
*/
func UniformityMulticastBlockImport(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_import_recv, nil, msg.Session)
	//Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_import_recv, nil, msg.Session)
	if !forks.GetLongChain().SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_import_recv, nil, msg.Session)
		return
	}
	engine.Log.Info("接收到区块导入消息：%s", hex.EncodeToString(*message.Body.Content))

	ImportBlockByCache(message.Body.Content)

	ok := []byte("ok")
	Area.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_import_recv, &ok, msg.Session)

	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))

	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息

}

/*
接收见证人导入区块命令_返回
*/
func UniformityMulticastBlockImport_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_uniformity_witness_multicas_block_import, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收区块签名广播
当矿工挖到一个新的区块之前，会广播这个区块，等待其他见证者签名
*/
func MulticastBlockHeadSign(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//engine.Log.Info("收到签名广播")
	chain := GetLongChain()
	if chain == nil {
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
		return
	}

	//当前节点是否同步区块完成
	if !chain.SyncBlockFinish {
		//engine.Log.Error("未同步完成")
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
		return
	}

	coinbase := Area.Keystore.GetCoinbase()

	//检查自己是不是见证人
	if !chain.WitnessChain.FindWitness(coinbase.Addr) && !chain.WitnessBackup.haveWitness(&coinbase.Addr) {
		//engine.Log.Error("不是见证人")
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
		return
	}

	//解析区块
	bhVO, err := ParseBlockHeadVOProto(message.Body.Content)
	if err != nil {
		engine.Log.Warn("Parse block broadcast error: %s", err.Error())
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
		return
	}
	// engine.Log.Info("签名组高度：%d，区块高度：%d", bhVO.BH.GroupHeight, bhVO.BH.Height)

	//验证时间段内是否多次签名
	//if !CheckSignExipre.Check(bhVO.BH.Height) {
	//	engine.Log.Error("多次签名高度：%d", bhVO.BH.Height)
	//	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	//	return
	//}

	//验证区块高度
	// currentBlock := chain.GetCurrentBlock()
	// if bhVO.BH.Height <= currentBlock+1 {
	// 	engine.Log.Error("签名区块高度错误: group：%d，bhvo：%d，currentBlock：%d", bhVO.BH.GroupHeight, bhVO.BH.Height, currentBlock)
	// 	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	// 	return
	// }

	// //验证节点自己出的块并签名成功，是则拒绝给其他节点签名
	// if !CurrSignBlock.IsEmpty() {
	// 	if CurrSignBlock.BlockHeight == bhVO.BH.Height ||
	// 		!(bhVO.BH.Height-1 == CurrSignBlock.BlockHeight && bytes.Equal(bhVO.BH.Previousblockhash, CurrSignBlock.BlockHash)) {
	// 		engine.Log.Error("已签名成功自己生成的块: %d %d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	// 		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	// 		return
	// 	}
	// 	//签名
	// 	bhVO.BH.BuildSign(coinbase.Addr)

	// 	//CheckSignExipre.Add(bhVO.BH.Height)

	// 	content := bhVO.BH.Sign
	// 	//回复签名
	// 	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, &content, msg.Session)
	// 	return
	// }

	////验证是否已经导入过其他区块了
	//height := GetHighestBlock()
	////engine.Log.Info("当前组高度：%d，当前区块高度：%d，已导入最高区块高度:%d", bhVO.BH.GroupHeight, bhVO.BH.Height, height)
	//if height >= bhVO.BH.Height {
	//	engine.Log.Error("已经导入过: %d %d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	//	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	//	return
	//}

	//验证当前区块的见证人
	// if currW, _ := chain.WitnessChain.FindWitnessForBlockOnly(bhVO); currW == nil {
	// 	engine.Log.Error("当前区块见证人错误: %d %d %s", bhVO.BH.GroupHeight, bhVO.BH.Height, bhVO.BH.Witness.B58String())
	// 	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	// 	return
	// }

	//查找前置区块是否存在，并且合法
	// preWitness := chain.WitnessChain.FindPreWitnessForBlock(bhVO.BH.Previousblockhash)
	// if preWitness == nil {
	// 	engine.Log.Error("待签名区块前置区块不存在: %d %d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	// 	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	// 	return
	// }

	//验证签名时间
	//witness, _ := GetLongChain().WitnessChain.FindWitnessForBlock(bhVO)
	//if config.TimeNow().UnixNano() > witness.CreateBlockTime+config.Mining_block_time.Nanoseconds() {
	//	engine.Log.Error("待签名区块时间超时")
	//	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	//	return
	//}

	//验证区块高度和前置hash
	// if preWitness.Block.Height+1 != bhVO.BH.Height || !bytes.Equal(preWitness.Block.Id, bhVO.BH.Previousblockhash) {
	// 	engine.Log.Error("验证区块高度和前置hash不一致")
	// 	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
	// 	return
	// }

	//是否允许签名
	permitSign := workModeLockStatic.CheckBlockSignLock(bhVO, chain)
	if !permitSign {
		engine.Log.Error("签名验证失败，不允许签名")
		Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, nil, msg.Session)
		return
	}

	//签名
	bhVO.BH.BuildSign(coinbase.Addr, Area.Keystore)

	//CheckSignExipre.Add(bhVO.BH.Height)

	//content := EncodeBlockSign(bhVO.BH.Hash, bhVO.BH.Sign)
	// content := bhVO.BH.Sign
	bs := go_protos.BlockSign{Sign: bhVO.BH.Sign, Witness: coinbase.Addr, Puk: coinbase.Puk}
	content, _ := bs.Marshal()
	//回复签名
	Area.SendNeighborReplyMsg(message, config.MSGID_multicast_blockhead_sign_recv, &content, msg.Session)
}

/*
接收区块签名广播_返回
当矿工挖到一个新的区块之前，会广播这个区块，等待其他见证者签名
*/
func MulticastBlockHeadSign_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)

	ses, ok := c.GetSession(config.AreaName[:], utils.Bytes2string(*message.Head.Sender))
	if !ok {
		return
	}
	flood.GroupWaitRecv.ResponseBytes(utils.Bytes2string(ses.GetIndex()), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//chain := GetLongChain()
	//if chain == nil {
	//	return
	//}
	//
	////当前节点是否同步区块完成
	//if !chain.SyncBlockFinish {
	//	return
	//}
	//
	//coinbase := Area.Keystore.GetCoinbase()
	//
	////是否是见证人
	//if !chain.WitnessChain.FindWitness(coinbase.Addr) {
	//	return
	//}
	//
	////解码信息
	//hash, sign, err := DecodeBlockSign(*message.Body.Content)
	//if err != nil {
	//	return
	//}
	//
	////获取缓存中的区块
	//bhvo := GetBlockByCache(&hash)
	//
	//if bhvo.BH.IsSign() && !bhvo.BH.SetExtSign(sign) {
	//	return
	//}

	//SignBuildBlock(bhvo)
}

/*
*
接收区块确认广播
*/
func P2pBlockAck(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//验证是否缓存
	if !ExistBlockByCache(message.Body.Content) {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_block_ack_recv, nil)
		return
	}

	ok := []byte("ok")
	Area.SendP2pReplyMsg(message, config.MSGID_p2p_block_ack_recv, &ok)
}

/*
*
接收区块确认广播_返回
*/
func P2pBlockAck_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收地址余额查询
*/
func P2pGetAddrBal(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//校验传入地址
	addr := string(*message.Body.Content)
	addrCoin := crypto.AddressFromB58String(addr)
	ok := crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrbalance_rev, nil)
		return
	}
	chain := GetLongChain()
	if !chain.SyncBlockFinish {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrbalance_rev, nil)
		return
	}
	notspend, _, _ := GetBalanceForAddrSelf(addrCoin)
	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, notspend)
	b := dataBuf.Bytes()
	Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrbalance_rev, &b)
}

/*
接收地址锁定余额
*/
func P2pGetAddrLockBal(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//校验传入地址
	addr := string(*message.Body.Content)
	addrCoin := crypto.AddressFromB58String(addr)
	ok := crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addr_lock_balance_rev, nil)
		return
	}
	lockValue, _ := GetLongChain().GetBalance().FindLockTotalByAddr(&addrCoin)
	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, lockValue)
	b := dataBuf.Bytes()
	Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addr_lock_balance_rev, &b)
}

/*
接收地址nonce查询
*/
func P2pGetAddrNonce(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//校验传入地址
	addr := string(*message.Body.Content)
	addrCoin := crypto.AddressFromB58String(addr)
	ok := crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrnonce_rev, nil)
		return
	}
	chain := GetLongChain()
	if !chain.SyncBlockFinish {
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrnonce_rev, nil)
		return
	}
	nonce, err := GetAddrNonce(&addrCoin)
	if err != nil {
		engine.Log.Info("GetAddrNonce error:%s", err.Error())
		Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrnonce_rev, nil)
		return
	}
	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, nonce.Uint64())
	b := dataBuf.Bytes()
	Area.SendP2pReplyMsg(message, config.MSGID_p2p_node_get_addrnonce_rev, &b)
}

/*
接收SWAP交易广播
*/
func MulticastSwapTransaction(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//检查内存
	if CheckOutOfMemory() {
		engine.Log.Error("Memory is too height")
		return
	}

	config.GetRpcRate("MulticastTx", true)

	if config.Model == config.Model_light {
		return
	}

	txbase, err := ParseSwapTxProto(message.Body.Content) // ParseTxBase(0, message.Body.Content)
	if err != nil {
		engine.Log.Error("Broadcast transaction format error %s", err.Error())
		return
	}
	txbase.BuildHash()

	chain := GetLongChain()
	if chain == nil {
		engine.Log.Error("chain is nil")
		return
	}
	if !chain.SyncBlockFinish {
		engine.Log.Error("chain is not SyncBlockFinish")
		return
	}

	checkSwapTxQueue <- txbase

}
