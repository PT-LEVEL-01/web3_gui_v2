package libp2parea

import (
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

var msgIdRegisterMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

/*
注册一个ID，并判断是否有重复
@return    bool    是否成功
*/
func RegisterMsgId(id uint64) bool {
	_, ok := msgIdRegisterMap.LoadOrStore(id, nil)
	if ok {
		utils.Log.Error().Msgf("重复注册的消息ID:%d", id)
	}
	return !ok
}

/*
注册消息编号，有重复就panic
*/
func RegMsgIdExistPanic(id uint64) uint64 {
	if !RegisterMsgId(id) {
		panic("msg id exist")
	}
	return id
}

/*
 * 节点上下线及关闭回调函数
 *
 * @param	addr		AddressNet	节点地址
 * @param	machineID	string		节点机器id
 */
type NodeEventCallbackHandler func(addr nodeStore.AddressNet, machineID string)

/*
注册邻居节点消息，不转发
*/
func (this *Node) Register_neighbor(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_neighbor(msgid, handler)
}

/*
发送一个新邻居节点消息
@return    bool    消息是否发送成功
*/
func (this *Node) SendNeighborMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) utils.ERROR {
	sss := this.SessionManager.FindSessionAreaSelfByAddr(recvid)
	wg := new(sync.WaitGroup)
	wg.Add(len(sss))
	ERROut := atomic.Pointer[utils.ERROR]{}
	success := new(atomic.Bool)
	success.Store(false)
	for _, one := range sss {
		go func() {
			_, ERR := this.MessageCenter.SendNeighborMsgBySession(one, msgid, content, 0)
			if ERR.CheckSuccess() {
				success.Store(true)
			} else {
				ERROut.Store(&ERR)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if success.Load() {
		return utils.NewErrorSuccess()
	}
	ERR := ERROut.Load()
	return *ERR
}

/*
对某个消息回复
*/
func (this *Node) SendNeighborReplyMsg(message *message_center.MessageBase, content *[]byte) utils.ERROR {
	return this.MessageCenter.SendNeighborReplyMsg(message, content, 0)
}

/*
发送一个新邻居节点消息
发送出去，等返回
*/
func (this *Node) SendNeighborMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	//id := ulid.Make().Bytes()
	//utils.Log.Info().Hex("发送邻居节点消息", id).Send()
	sss := this.SessionManager.FindSessionAreaSelfByAddr(recvid)
	resultChan := make(chan utils.ERROR, len(sss))
	ERROut := atomic.Pointer[utils.ERROR]{}
	resultBs := atomic.Pointer[[]byte]{}
	for _, one := range sss {
		go func() {
			bs, ERR := this.MessageCenter.SendNeighborMsgWaitRequest(one, msgid, content, timeout)
			if ERR.CheckSuccess() {
				resultBs.Store(bs)
			} else {
				ERROut.Store(&ERR)
			}
			resultChan <- ERR
		}()
	}
	//utils.Log.Info().Hex("发送邻居节点消息", id).Send()
	success := false
	for i := 0; i < len(sss); i++ {
		ERR := <-resultChan
		if ERR.CheckSuccess() {
			//其中一个成功，则返回
			success = true
			break
		}
	}
	//utils.Log.Info().Hex("发送邻居节点消息", id).Send()
	if success {
		return resultBs.Load(), utils.NewErrorSuccess()
	}
	ERR := ERROut.Load()
	return nil, *ERR
}

/*
注册广播消息
*/
func (this *Node) Register_multicast(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_multicast(msgid, handler)
}

/*
发送一个新的广播消息
*/
func (this *Node) SendMulticastMsg(msgid uint64, content *[]byte) utils.ERROR {
	return this.MessageCenter.SendMulticastMsg(msgid, content, 0)
}

/*
发送一个新的广播消息
*/
func (this *Node) SendMulticastMsgWaitRequest(msgid uint64, content *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	return this.MessageCenter.SendMulticastMsgWaitRequest(msgid, content, timeout)
}

/*
对广播消息回复
*/
func (this *Node) SendMulticastReplyMsg(message *message_center.MessageBase, content *[]byte) utils.ERROR {
	return this.MessageCenter.SendP2pReplyMsg(message, content)
}

/*
注册点对点通信消息
*/
func (this *Node) Register_p2p(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_p2p(msgid, handler)
}

/*
发送一个新消息
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Node) SendP2pMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*message_center.MessageBase, utils.ERROR) {
	return this.MessageCenter.SendP2pMsg(msgid, recvid, content, timeout)
}

/*
给指定节点发送一个消息
@return    *[]byte      返回的内容
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Node) SendP2pMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	return this.MessageCenter.SendP2pMsgWaitRequest(msgid, recvid, content, timeout)
}

/*
发送一个新消息
是SendP2pMsg方法的定制版本，多了recvSuperId参数。
*/
//func (this *Area) SendP2pMsgEX(msgid uint64, recvid, recvSuperId *nodeStore.AddressNet, content *[]byte, hash *[]byte) (message_bean.MessageItr, utils.ERROR) {
//	return this.MessageCenter.SendP2pMsgEX(msgid, recvid, recvSuperId, content, hash)
//}

/*
对某个消息回复
*/
func (this *Node) SendP2pReplyMsg(message *message_center.MessageBase, content *[]byte) utils.ERROR {
	return this.MessageCenter.SendP2pReplyMsg(message, content)
}

/*
注册点对点通信消息
*/
func (this *Node) Register_p2pHE(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_p2pHE(msgid, handler)
}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
//func (this *Node) SendP2pMsgHE(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) utils.ERROR {
//	return this.MessageCenter.SendP2pMsgHE(msgid, recvid, content)
//}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Node) SendP2pMsgHEWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	return this.MessageCenter.SendP2pMsgHEWaitRequest(msgid, recvid, content, timeout)
}

/*
对某个消息回复
*/
func (this *Node) SendP2pReplyMsgHE(message *message_center.MessageBase, content *[]byte) utils.ERROR {
	return this.MessageCenter.SendP2pReplyMsgHE(message, content)
}

/*
注册一个RPC接口
*/
func (this *Node) RegisterRPC(sortNumber int, rpcName string, handler any, desc string, pvs ...engine.ParamValid) utils.ERROR {
	return this.SessionEngine.RegisterRPC(sortNumber, rpcName, handler, desc, pvs...)
}

/*
设置RPC连接用户名称
*/
func (this *Node) AddRpcUser(rpcUsername, password string) utils.ERROR {
	return this.SessionEngine.AddRpcUser(rpcUsername, password)
}

/*
设置RPC连接用户名称
*/
func (this *Node) UpdateRpcUser(rpcUsername, password string) utils.ERROR {
	return this.SessionEngine.UpdateRpcUser(rpcUsername, password)
}

/*
设置RPC连接用户名称
*/
func (this *Node) DelRpcUser(rpcUsername string) utils.ERROR {
	return this.SessionEngine.DelRpcUser(rpcUsername)
}

/*
设置RPC连接用户名称
*/
func (this *Node) RunRpcMethod(method string, params map[string]interface{}) (map[string]interface{}, utils.ERROR) {
	return this.SessionEngine.RunRpcMethod(method, params)

}
