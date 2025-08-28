package libp2parea

import (
	"errors"
	ma "github.com/multiformats/go-multiaddr"
	"strconv"
	"time"
	"web3_gui/keystore/adapter"
	kv2 "web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/libp2parea/v2"
	message_center_new "web3_gui/libp2parea/v2/message_center"
	ns2 "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type Area struct {
	Keystore      keystore.KeystoreInterface
	node          *libp2parea.Node
	NodeManager   *NodeManager
	SessionEngine *SessionEngine
}

func NewArea(areaName [32]byte, pre string, key keystore.KeystoreInterface, pwd string) (*Area, utils.ERROR) {
	//itr := key.(kv2.KeystoreInterface)
	node, ERR := libp2parea.NewNode(areaName, pre, key, pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	nm := NodeManager{
		node: node,
	}
	se := SessionEngine{
		node: node,
	}
	area := Area{
		Keystore:      key,
		node:          node,
		NodeManager:   &nm,
		SessionEngine: &se,
	}
	return &area, utils.NewErrorSuccess()
}

func BuildArea(node *libp2parea.Node, pwd string) *Area {
	//addrPre := node.Keystore
	keyst := node.Keystore.(*kv2.Keystore)
	k := keystore.NewKeystoreItrByKeystore(keyst, pwd)
	nm := NodeManager{
		node: node,
	}
	se := SessionEngine{
		node: node,
	}
	area := Area{
		Keystore:      k,
		node:          node,
		NodeManager:   &nm,
		SessionEngine: &se,
	}
	return &area
}

/*
注
*/
func (this *Area) GetNode() *libp2parea.Node {
	return this.node
}

/*
注册邻居节点消息，不转发
*/
func (this *Area) Register_neighbor(msgid uint64, handler message_center.MsgHandler) {
	this.node.Register_neighbor(msgid, func(message *message_center_new.MessageBase) {
		var c engine.Controller = &engine.ControllerImpl{}
		sessionOld := engine.SessionImpl{Session: message.GetPacket().Session}
		var msg engine.Packet = engine.Packet{}
		msg.Session = &sessionOld
		var messageOld *message_center.Message = message_center.BuildMessage(message)
		handler(c, msg, messageOld)
	})
}

/*
对某个消息回复
*/
func (this *Area) SendNeighborReplyMsg(message *message_center.Message, msgid uint64, content *[]byte, session engine.Session) error {
	ERR := this.node.SendNeighborReplyMsg(message.Message, content)
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	return nil
}

/*
发送一个新邻居节点消息
发送出去，等返回
*/
func (this *Area) SendNeighborMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, error) {
	recvidCoin := coin_address.AddressCoin(*recvid)
	recvIdNew := ns2.NewAddressNet(recvidCoin)
	bs, ERR := this.node.SendNeighborMsgWaitRequest(msgid, recvIdNew, content, timeout)
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	return bs, nil
	//return this.MessageCenter.SendNeighborMsgWaitRequest(msgid, recvid, content, timeout)
}

/*
注册广播消息
*/
func (this *Area) Register_multicast(msgid uint64, handler message_center.MsgHandler) {
	this.node.Register_multicast(msgid, func(message *message_center_new.MessageBase) {
		var c engine.Controller = &engine.ControllerImpl{}
		sessionOld := engine.SessionImpl{Session: message.GetPacket().Session}
		var msg engine.Packet = engine.Packet{}
		msg.Session = &sessionOld
		var messageOld *message_center.Message = message_center.BuildMessage(message)
		handler(c, msg, messageOld)
	})
}

/*
发送一个新的广播消息
*/
func (this *Area) SendMulticastMsg(msgid uint64, content *[]byte) error {
	ERR := this.node.MessageCenter.SendMulticastMsg(msgid, content, 0)
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	return nil
}

func (this *Area) BroadcastsAll(msgid, p2pMsgid uint64, whiltlistNodes, superNodes, proxyNodes []nodeStore.AddressNet, hash *[]byte) error {
	ERR := this.node.MessageCenter.SendMulticastMsg(msgid, hash, 0)
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	return nil
}

/*
注册点对点通信消息
*/
func (this *Area) Register_p2p(msgid uint64, handler message_center.MsgHandler) {
	this.node.Register_p2p(msgid, func(message *message_center_new.MessageBase) {
		var c engine.Controller = &engine.ControllerImpl{}
		sessionOld := engine.SessionImpl{Session: message.GetPacket().Session}
		var msg engine.Packet = engine.Packet{}
		msg.Session = &sessionOld
		var messageOld *message_center.Message = message_center.BuildMessage(message)
		handler(c, msg, messageOld)
	})
}

/*
发送一个新消息
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Area) SendP2pMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, bool, bool, error) {
	recvidCoin := coin_address.AddressCoin(*recvid)
	recvIdNew := ns2.NewAddressNet(recvidCoin)
	msgNew, ERR := this.node.SendP2pMsg(msgid, recvIdNew, content, 0)
	if ERR.CheckFail() {
		return nil, false, false, errors.New(ERR.String())
	}
	msgOld := message_center.BuildMessage(msgNew)
	return msgOld, false, false, nil
}

/*
对某个消息回复
*/
func (this *Area) SendP2pReplyMsg(message *message_center.Message, msgid uint64, content *[]byte) error {
	ERR := this.node.SendNeighborReplyMsg(message.Message, content)
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	return nil
}

/*
获取本节点地址
*/
func (this *Area) GetNetId() nodeStore.AddressNet {
	return nodeStore.AddressNet(this.node.NodeManager.NodeSelf.IdInfo.Id.GetAddr())
}

func (this *Area) SetNetTypeToTest() {

}

func (this *Area) GetNodeSelf() *ns2.NodeInfo {
	return this.node.NodeManager.NodeSelf
}

func (this *Area) GetAddrAndPort() (string, uint16) {
	nodeInfo := this.node.NodeManager.NodeSelf
	ms := nodeInfo.GetMultiaddrWAN()
	if len(ms) <= 0 {
		return "", nodeInfo.Port
	}
	return ms[0].Addr, nodeInfo.Port
}

/*
添加一个地址到白名单
*/
func (this *Area) AddWhiteList(addr nodeStore.AddressNet) bool {
	return this.node.AddWhiteListByAddr(*ns2.NewAddressNet([]byte(addr)))
}

/*
添加一个连接
*/
func (this *Area) AddConnect(ip string, port uint16) (engine.Session, error) {
	a, err := ma.NewMultiaddr("/ip4/" + ip + "/tcp/" + strconv.Itoa(int(port)) + "/ws")
	if err != nil {
		return nil, err
	}
	_, ERR := this.node.AddConnect(a)
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	return nil, nil
}

/*
删除一个地址到白名单
*/
func (this *Area) RemoveWhiteList(addr nodeStore.AddressNet) bool {
	return this.node.RemoveWhiteList(*ns2.NewAddressNet([]byte(addr)))
}

func (this *Area) WaitAutonomyFinish() {
	this.node.WaitAutonomyFinish()
}

func (this *Area) GetNodeManager() *NodeManager {
	return this.NodeManager
}

func (this *Area) StartUP(port uint16) utils.ERROR {
	return this.node.StartUP(port)
}

type NodeManager struct {
	node *libp2parea.Node
}

/*
在连接中获得白名单节点
*/
func (this *NodeManager) GetWhiltListNodes() []nodeStore.AddressNet {
	addrInfos := this.node.NodeManager.GetWhiltListNodes()
	addrs := make([]nodeStore.AddressNet, 0, len(addrInfos))
	for _, one := range addrInfos {
		addrOne := nodeStore.AddressNet(one.IdInfo.Id.GetAddr())
		addrs = append(addrs, addrOne)
	}
	return addrs
}

// 得到所有逻辑节点，不包括本节点，也不包括代理节点
func (this *NodeManager) GetLogicNodes() []nodeStore.AddressNet {
	addrInfosWAN := this.node.NodeManager.GetLogicNodesWAN()
	addrInfosLAN := this.node.NodeManager.GetLogicNodesLAN()
	addrInfosFree := this.node.NodeManager.GetFreeNodes()
	addrInfosProxy := this.node.NodeManager.GetProxyNodes()
	addrInfosWhilt := this.node.NodeManager.GetWhiltListNodes()
	addrInfos := append(addrInfosWAN, addrInfosLAN...)
	addrInfos = append(addrInfos, addrInfosFree...)
	addrInfos = append(addrInfos, addrInfosProxy...)
	addrInfos = append(addrInfos, addrInfosWhilt...)
	addrs := make([]nodeStore.AddressNet, 0, len(addrInfos))
	for _, one := range addrInfos {
		addrOne := nodeStore.AddressNet(one.IdInfo.Id.GetAddr())
		//addrV2 := one.IdInfo.Id.GetAddr()
		//utils.Log.Info().Str("地址转换错误1", addrOne.B58String()).Str("地址转换错误2", addrV2.B58String()).Send()
		addrs = append(addrs, addrOne)
	}
	return addrs
}

/*
获取本机被其他当作逻辑节点的连接
*/
func (this *NodeManager) GetNodesClient() []nodeStore.AddressNet {
	addrInfos := this.node.NodeManager.GetFreeNodes()
	addrs := make([]nodeStore.AddressNet, 0, len(addrInfos))
	for _, one := range addrInfos {
		addrOne := nodeStore.AddressNet(one.IdInfo.Id.GetAddr())
		addrs = append(addrs, addrOne)
	}
	return addrs
}

/*
获取本机被其他当作逻辑节点的连接
*/
func (this *NodeManager) GetProxyAll() []nodeStore.AddressNet {
	addrInfos := this.node.NodeManager.GetProxyNodes()
	addrs := make([]nodeStore.AddressNet, 0, len(addrInfos))
	for _, one := range addrInfos {
		addrOne := nodeStore.AddressNet(one.IdInfo.Id.GetAddr())
		addrs = append(addrs, addrOne)
	}
	return addrs
}

func (this *NodeManager) GetSuperNodeIps() []string {
	nodeInfos := this.node.NodeManager.GetLogicNodesWAN()
	nodeInfos = append(nodeInfos, this.node.NodeManager.GetFreeNodes()...)
	ips := make([]string, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, v := range one.GetMultiaddrWAN() {
			ips = append(ips, v.Addr)
		}
	}
	return ips
}

type SessionEngine struct {
	node *libp2parea.Node
}

// 获得session
func (this *SessionEngine) GetSession(areaName []byte, name string) (engine.Session, bool) {
	ss := this.node.SessionEngine.GetSession([]byte(name))
	if ss == nil {
		return nil, false
	}
	ssImple := engine.SessionImpl{ss}
	return &ssImple, true
}
