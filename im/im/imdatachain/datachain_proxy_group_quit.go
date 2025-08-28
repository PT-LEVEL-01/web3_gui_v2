package imdatachain

import (
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_group_quit, ParseProxyGroupMemberQuitFactory)
}

/*
群成员主动退出群聊
*/
type ProxyGroupMemberQuit struct {
	DataChainProxyBase //
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *ProxyGroupMemberQuit) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

func (this *ProxyGroupMemberQuit) Forward() bool {
	return true
}

/*
 */
func (this *ProxyGroupMemberQuit) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseProxyGroupMemberQuitFactory(bs []byte) (DataChainProxyItr, error) {
	proxyBase, err := ParseDatachainProxyBase(bs)
	if err != nil {
		return nil, err
	}
	fli := &ProxyGroupMemberQuit{DataChainProxyBase: *proxyBase}
	return fli, nil
}

func NewDataChainProxyGroupMemberQuit(addrSelf nodeStore.AddressNet, groupId []byte) *ProxyGroupMemberQuit {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_group_quit, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	proxyBase.GroupID = groupId
	f := ProxyGroupMemberQuit{DataChainProxyBase: *proxyBase}
	return &f
}
