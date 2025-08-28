package imdatachain

import (
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_group_dissolve, ParseProxyGroupDissolveFactory)
}

/*
群成员主动退出群聊
*/
type ProxyGroupDissolve struct {
	DataChainProxyBase //
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *ProxyGroupDissolve) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

func (this *ProxyGroupDissolve) Forward() bool {
	return true
}

/*
 */
func (this *ProxyGroupDissolve) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseProxyGroupDissolveFactory(bs []byte) (DataChainProxyItr, error) {
	proxyBase, err := ParseDatachainProxyBase(bs)
	if err != nil {
		return nil, err
	}
	fli := &ProxyGroupDissolve{*proxyBase}
	return fli, nil
}

func NewProxyGroupDissolve(addrSelf nodeStore.AddressNet, groupId []byte) *ProxyGroupDissolve {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_group_dissolve, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	proxyBase.GroupID = groupId
	f := ProxyGroupDissolve{DataChainProxyBase: *proxyBase}
	return &f
}
