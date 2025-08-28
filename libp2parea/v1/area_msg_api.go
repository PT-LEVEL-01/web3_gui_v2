package libp2parea

import (
	"time"
	"web3_gui/utils"

	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/libp2parea/v1/virtual_node"
)

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
func (this *Area) Register_neighbor(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_neighbor(msgid, handler, false)
}

/*
发送一个新邻居节点消息
@return    bool    消息是否发送成功
*/
func (this *Area) SendNeighborMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendNeighborMsg(msgid, recvid, content)
}

/*
对某个消息回复
*/
func (this *Area) SendNeighborReplyMsg(message *message_center.Message, msgid uint64, content *[]byte, session engine.Session) error {
	return this.MessageCenter.SendNeighborReplyMsg(message, msgid, content, session)
}

/*
发送一个新邻居节点消息
发送出去，等返回
*/
func (this *Area) SendNeighborMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet,
	content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendNeighborMsgWaitRequest(msgid, recvid, content, timeout)
}

/*
注册广播消息
*/
func (this *Area) Register_multicast(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_multicast(msgid, handler, false)
}

/*
发送一个新的广播消息
*/
func (this *Area) SendMulticastMsg(msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendMulticastMsg(msgid, content)
}

/*
发送一个新的广播消息
*/
func (this *Area) SendMulticastMsgWaitRequest(msgid uint64, content *[]byte, timeout time.Duration) (interface{}, error) {
	return this.MessageCenter.SendMulticastMsgWaitRequest(msgid, content, timeout)
}

func (this *Area) BroadcastsAll(msgid, p2pMsgid uint64, whiltlistNodes, superNodes, proxyNodes []nodeStore.AddressNet, hash *[]byte) error {
	return this.MessageCenter.BroadcastsAll(msgid, p2pMsgid, whiltlistNodes, superNodes, proxyNodes, hash)
}

/*
注册消息
从超级节点中搜索目标节点
*/
func (this *Area) Register_search_super(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_search_super(msgid, handler, false)
}

/*
发送一个新的查找超级节点消息
*/
func (this *Area) SendSearchSuperMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendSearchSuperMsg(msgid, recvid, content)
}

/*
发送一个新的查找超级节点消息
*/
func (this *Area) SendSearchSuperMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendSearchSuperMsgWaitRequest(msgid, recvid, content, timeout, false)
}

/*
对某个消息回复
*/
func (this *Area) SendSearchSuperReplyMsg(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendSearchSuperReplyMsg(message, msgid, content)
}

/*
注册消息
从所有节点中搜索节点，包括普通节点
*/
func (this *Area) Register_search_all(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_search_super(msgid, handler, false)
}

/*
发送一个新的查找超级节点消息
*/
func (this *Area) SendSearchAllMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendSearchSuperMsg(msgid, recvid, content)
}

/*
对某个消息回复
*/
func (this *Area) SendSearchAllReplyMsg(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendSearchSuperReplyMsg(message, msgid, content)
}

/*
注册点对点通信消息
*/
func (this *Area) Register_p2p(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_p2p(msgid, handler, false)
}

/*
发送一个新消息
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Area) SendP2pMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsg(msgid, recvid, content)
}

/*
给指定节点发送一个消息
@return    *[]byte      返回的内容
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Area) SendP2pMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgWaitRequest(msgid, recvid, content, timeout)
}

/*
发送一个新消息
是SendP2pMsg方法的定制版本，多了recvSuperId参数。
*/
func (this *Area) SendP2pMsgEX(msgid uint64, recvid, recvSuperId *nodeStore.AddressNet, content *[]byte, hash *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendP2pMsgEX(msgid, recvid, recvSuperId, content, hash)
}

/*
对某个消息回复
*/
func (this *Area) SendP2pReplyMsg(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendP2pReplyMsg(message, msgid, content)
}

/*
注册点对点通信消息
*/
func (this *Area) Register_p2pHE(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_p2pHE(msgid, handler, false)
}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Area) SendP2pMsgHE(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHE(msgid, recvid, content)
}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *Area) SendP2pMsgHEWaitRequest(msgid uint64, recvid *nodeStore.AddressNet,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHEWaitRequest(msgid, recvid, content, timeout)
}

/*
对某个消息回复
*/
func (this *Area) SendP2pReplyMsgHE(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendP2pReplyMsgHE(message, msgid, content)
}

/*
注册虚拟节点搜索节点消息
*/
func (this *Area) Register_vnode_search(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_vnode_search(msgid, handler, false)
}

/*
发送虚拟节点搜索节点消息
*/
func (this *Area) SendVnodeSearchMsg(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend,
	content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendVnodeSearchMsg(msgid, sendVnodeid, recvVnodeid, nil, nil, content)
}

/*
发送虚拟节点搜索节点消息
*/
func (this *Area) SendVnodeSearchMsgWaitRequest(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend,
	content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendVnodeSearchMsgWaitRequest(msgid, sendVnodeid, recvVnodeid, nil, nil, content, timeout, false)
}

/*
对发送虚拟节点搜索节点消息回复
*/
func (this *Area) SendVnodeSearchReplyMsg(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendVnodeSearchReplyMsg(message, msgid, content, false)
}

/*
注册虚拟节点之间点对点加密消息
*/
func (this *Area) Register_vnode_p2pHE(msgid uint64, handler message_center.MsgHandler) {
	this.MessageCenter.Register_vnode_p2pHE(msgid, handler, false)
}

/*
 * 发送虚拟节点之间点对点消息
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNodeId		*AddressNet			接收方真实地址，可以传nil，如果传nil，则会先根据recvVnode找到nid，再进行p2p转发，否则将会直接发送p2p消息
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeP2pMsgHE(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvNodeId *nodeStore.AddressNet,
	content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendVnodeP2pMsgHE(msgid, sendVnodeid, recvVnodeid, recvNodeId, nil, nil, "", content)
}

/*
 * 发送虚拟节点之间点对点消息, 等待消息返回
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNodeId		*AddressNet			接收方真实地址，可以传nil，如果传nil，则会先根据recvVnode找到nid，再进行p2p转发，否则将会直接发送p2p消息
 * @param	timeout			time.Duration		超时时间
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeP2pMsgHEWaitRequest(msgid uint64, sendVnodeid,
	recvVnodeid *virtual_node.AddressNetExtend, recvNodeId *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendVnodeP2pMsgHEWaitRequest(msgid, sendVnodeid, recvVnodeid, recvNodeId, nil, nil, "", content, timeout)
}

/*
对发送虚拟节点之间点对点消息回复
*/
func (this *Area) SendVnodeP2pReplyMsgHE(message *message_center.Message, msgid uint64, content *[]byte) error {
	return this.MessageCenter.SendVnodeP2pReplyMsgHE(message, msgid, content, false)
}

/*
网络中查询一个逻辑节点地址的真实地址
*/
func (this *Area) SearchVnodeId(vnodeid *virtual_node.AddressNetExtend) (*virtual_node.AddressNetExtend, error) {
	nodeBs, err := this.Vc.SearchVnodeId(vnodeid, nil, nil, false, 1)
	if err != nil {
		return nil, err
	} else if nodeBs == nil || len(*nodeBs) == 0 {
		return nil, err
	}

	virtualNode := virtual_node.AddressNetExtend(*nodeBs)
	return &virtualNode, err
}

func (this *Area) SearchVnodeIdOnebyone(vnodeid *virtual_node.AddressNetExtend, num uint16) ([]virtual_node.Vnodeinfo, error) {
	var rbs []virtual_node.Vnodeinfo

	bs, err := this.Vc.SearchVnodeId(vnodeid, nil, nil, true, num)
	if err != nil || bs == nil {
		return nil, err
	}

	vrp := new(go_protobuf.VnodeinfoRepeated)
	err = proto.Unmarshal(*bs, vrp)
	if err != nil {
		utils.Log.Info().Msgf("SearchVnodeIdOnebyone 不能解析proto:%s", err.Error())
		return nil, err
	}

	for _, one := range vrp.Vnodes {
		rbs = append(rbs, virtual_node.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		})
	}
	return rbs, nil
}

/*
 * 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 */
func (this *Area) SendSearchSuperMsgProxy(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendSearchSuperMsgProxy(msgid, recvid, recvProxyId, senderProxyId, content)
}

/*
 * 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 */
func (this *Area) SendSearchSuperMsgProxyWaitRequest(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendSearchSuperMsgProxyWaitRequest(msgid, recvid, recvProxyId, senderProxyId, content, timeout, false)
}

/*
 * 发送一个新消息，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 * @return    		*Message     返回的消息
 * @return    		bool         是否发送成功
 * @return    		bool         消息是发给自己
 */
func (this *Area) SendP2pMsgProxy(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgProxy(msgid, recvid, recvProxyId, senderProxyId, "", content)
}

/*
 * 给指定节点发送一个消息，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 *	@return    *[]byte      返回的内容
 *	@return    bool         是否发送成功
 *	@return    bool         消息是发给自己
 */
func (this *Area) SendP2pMsgProxyWaitRequest(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgProxyWaitRequest(msgid, recvid, recvProxyId, senderProxyId, "", content, timeout)
}

/*
 * 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 *	@return    *Message     返回的消息
 *	@return    bool         是否发送成功
 *	@return    bool         消息是发给自己
 */
func (this *Area) SendP2pMsgHEProxy(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHEProxy(msgid, recvid, recvProxyId, senderProxyId, "", content)
}

/*
 * 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点
 * @msgid			uint64		消息号
 * @recvid			addressnet	接收者节点id
 * @recvProxyId		addressnet	接收者代理节点id
 * @senderProxyId	addressnet	发送者代理节点id
 * @content			[]byte		发送内容
 *	@return    *Message     返回的消息
 *	@return    bool         是否发送成功
 *	@return    bool         消息是发给自己
 */
func (this *Area) SendP2pMsgHEProxyWaitRequest(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHEProxyWaitRequest(msgid, recvid, recvProxyId, senderProxyId, "", content, timeout)
}

/*
 * 对某个消息回复，可以指定发送端的代理节点
 */
func (this *Area) SendSearchSuperReplyMsgProxy(message *message_center.Message, msgid uint64, content *[]byte, senderProxyId *nodeStore.AddressNet) error {
	return this.MessageCenter.SendSearchSuperReplyMsg(message, msgid, content)
}

/*
 * 对某个消息回复，可以指定发送端的代理节点
 */
func (this *Area) SendP2pReplyMsgProxy(message *message_center.Message, msgid uint64, content *[]byte, senderProxyId *nodeStore.AddressNet) error {
	return this.MessageCenter.SendP2pReplyMsgProxy(message, msgid, content, senderProxyId)
}

/*
 * 对某个消息回复，可以指定发送端的代理节点
 */
func (this *Area) SendP2pReplyMsgHEProxy(message *message_center.Message, msgid uint64, content *[]byte, senderProxyId *nodeStore.AddressNet) error {
	return this.MessageCenter.SendP2pReplyMsgHEProxy(message, msgid, content, senderProxyId)
}

/*
 * 发送虚拟节点搜索节点消息，可以指定接收端和发送端的代理节点
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeSearchMsgProxy(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendVnodeSearchMsg(msgid, sendVnodeid, recvVnodeid, recvProxyId, senderProxyId, content)
}

/*
 * 发送虚拟节点搜索节点消息, 等待消息返回，可以指定接收端和发送端的代理节点
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeSearchMsgProxyWaitRequest(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendVnodeSearchMsgWaitRequest(msgid, sendVnodeid, recvVnodeid, recvProxyId, senderProxyId, content, timeout, false)
}

/*
 * 发送虚拟节点之间点对点消息，可以指定接收端、发送端的代理节点和接收方机器id
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNodeId		*AddressNet			接收方真实地址，可以传nil，如果传nil，则会先根据recvVnode找到nid，再进行p2p转发，否则将会直接发送p2p消息
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @param	recvMachineId	string				接收方机器Id
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeP2pMsgHEProxy(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvNodeId, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendVnodeP2pMsgHE(msgid, sendVnodeid, recvVnodeid, recvNodeId, recvProxyId, senderProxyId, recvMachineId, content)
}

/*
 * 发送虚拟节点之间点对点消息, 等待消息返回, 可以指定接收端、发送端的代理节点和接收方机器id
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNodeId		*AddressNet			接收方真实地址，可以传nil，如果传nil，则会先根据recvVnode找到nid，再进行p2p转发，否则将会直接发送p2p消息
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @param	recvMachineId	string				接收方机器Id
 * @param	content			*[]byte				内容
 * @param	timeout			time.Duration		超时时间
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeP2pMsgHEProxyWaitRequest(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvNodeId, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendVnodeP2pMsgHEWaitRequest(msgid, sendVnodeid, recvVnodeid, recvNodeId, recvProxyId, senderProxyId, recvMachineId, content, timeout)
}

/*
 * 根据磁力地址查询匹配的一个虚拟地址，可以指定接收端和发送端的代理节点
 *
 * @param	vnodeId			*AddressNetExtend	虚拟磁力地址
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @return  vs				*AddressNetExtend  	获取到的虚拟地址
 * @return  err				error	         	错误信息
 */
func (this *Area) SearchVnodeIdProxy(vnodeid *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet) (*virtual_node.AddressNetExtend, error) {
	nodeBs, err := this.Vc.SearchVnodeId(vnodeid, recvProxyId, senderProxyId, false, 1)
	if err != nil {
		return nil, err
	} else if nodeBs == nil || len(*nodeBs) == 0 {
		return nil, err
	}

	virtualNode := virtual_node.AddressNetExtend(*nodeBs)
	return &virtualNode, err
}

/*
 * 根据磁力地址查询匹配的虚拟地址列表，可以指定接收端和发送端的代理节点
 *
 * @param	vnodeId			*AddressNetExtend	虚拟磁力地址
 * @param	recvProxyId		*Addressnet			接收者代理节点id
 * @param	senderProxyId	*Addressnet			发送者代理节点id
 * @param	num				uint16				需要返回的最大数量
 * @return  vs				[]VnodeInfo     	获取到的虚拟地址信息数组
 * @return  err				error	        	错误信息
 */
func (this *Area) SearchVnodeIdOnebyoneProxy(vnodeid *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet, num uint16) ([]virtual_node.Vnodeinfo, error) {
	var rbs []virtual_node.Vnodeinfo

	bs, err := this.Vc.SearchVnodeId(vnodeid, recvProxyId, senderProxyId, true, num)
	if err != nil || bs == nil {
		return nil, err
	}

	vrp := new(go_protobuf.VnodeinfoRepeated)
	err = proto.Unmarshal(*bs, vrp)
	if err != nil {
		utils.Log.Warn().Msgf("SearchVnodeIdOnebyoneProxy 不能解析proto:%s", err.Error())
		return nil, err
	}

	for _, one := range vrp.Vnodes {
		rbs = append(rbs, virtual_node.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		})
	}
	return rbs, nil
}

/*
 * 发送一个新消息，可以指定接收端、发送端的代理节点和接收方机器id
 *
 * @param	msgid			uint64		消息号
 * @param	recvid			addressnet	接收者节点id
 * @param	recvProxyId		addressnet	接收者代理节点id
 * @param	senderProxyId	addressnet	发送者代理节点id
 * @param	recvMachineId	string		接收方机器Id
 * @param	content			[]byte		发送内容
 * @return  msg				*Message    返回的消息
 * @return  sendSuccess		bool        是否发送成功
 * @return  toSelf			bool        消息是发给自己
 */
func (this *Area) SendP2pMsgProxyMachineID(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgProxy(msgid, recvid, recvProxyId, senderProxyId, recvMachineId, content)
}

/*
 * 给指定节点发送一个消息，可以指定接收端、发送端的代理节点和接收方机器id
 * @param	msgid			uint64		消息号
 * @param	recvid			addressnet	接收者节点id
 * @param	recvProxyId		addressnet	接收者代理节点id
 * @param	senderProxyId	addressnet	发送者代理节点id
 * @param	recvMachineId	string		接收方机器Id
 * @param	content			[]byte		发送内容
 * @param	timeout			Duration	超时时间
 * @return  bs  			*[]byte     返回的内容
 * @return  sendSuccess  	bool        是否发送成功
 * @return  toSelf  		bool        消息是发给自己
 */
func (this *Area) SendP2pMsgProxyMachineIDWaitRequest(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgProxyWaitRequest(msgid, recvid, recvProxyId, senderProxyId, recvMachineId, content, timeout)
}

/*
 * 发送一个加密消息，包括消息头也加密，可以指定接收端、发送端的代理节点和接收方机器id
 *
 * @param	msgid			uint64		消息号
 * @param	recvid			addressnet	接收者节点id
 * @param	recvProxyId		addressnet	接收者代理节点id
 * @param	senderProxyId	addressnet	发送者代理节点id
 * @param	recvMachineId	string		接收方机器Id
 * @param	content			[]byte		发送内容
 * @return	msg				*Message    返回的消息
 * @return  sendSuccess		bool        是否发送成功
 * @return  toSelf		 	bool        消息是发给自己
 */
func (this *Area) SendP2pMsgHEProxyMachineID(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string, content *[]byte) (*message_center.Message, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHEProxy(msgid, recvid, recvProxyId, senderProxyId, recvMachineId, content)
}

/*
 * 发送一个加密消息，包括消息头也加密，可以指定接收端、发送端的代理节点和接收方机器id
 *
 * @param	msgid			uint64		消息号
 * @param	recvid			addressnet	接收者节点id
 * @param	recvProxyId		addressnet	接收者代理节点id
 * @param	senderProxyId	addressnet	发送者代理节点id
 * @param	recvMachineId	string		接收方机器Id
 * @param	content			[]byte		发送内容
 * @param	timeout			Duration	超时时间
 * @return	msg				*Message    返回的消息
 * @return  sendSuccess		bool        是否发送成功
 * @return  toSelf		 	bool        消息是发给自己
 */
func (this *Area) SendP2pMsgHEProxyMachineIDWaitRequest(msgid uint64, recvid, recvProxyId, senderProxyId *nodeStore.AddressNet, recvMachineId string,
	content *[]byte, timeout time.Duration) (*[]byte, bool, bool, error) {
	return this.MessageCenter.SendP2pMsgHEProxyWaitRequest(msgid, recvid, recvProxyId, senderProxyId, recvMachineId, content, timeout)
}

/*
 * 注册节点关闭连接回调函数
 */
func (this *Area) Register_nodeClosedCallback(handler NodeEventCallbackHandler) {
	this.closedCallbackFunc = append(this.closedCallbackFunc, handler)
}

/*
 * 注册节点被设置为超级代理节点地址回调函数
 */
func (this *Area) Register_nodeBeenGodAddrCallback(handler NodeEventCallbackHandler) {
	this.beenGodAddrCallbackFunc = append(this.beenGodAddrCallbackFunc, handler)
}

/*
 * 注册节点新建连接回调函数
 */
func (this *Area) Register_nodeNewConnCallback(handler NodeEventCallbackHandler) {
	this.newConnCallbackFunc = append(this.newConnCallbackFunc, handler)
}
