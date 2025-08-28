package imdatachain

import (
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_forward, ParseDatachainForwardFactory)
}

/*
需要转发的消息
*/
type ProxyForward struct {
	DataChainProxyBase //
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *ProxyForward) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

func (this *ProxyForward) Forward() bool {
	return true
}

/*
 */
func (this *ProxyForward) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDatachainForwardFactory(bs []byte) (DataChainProxyItr, error) {
	proxyBase, err := ParseDatachainProxyBase(bs)
	if err != nil {
		return nil, err
	}
	fli := &ProxyForward{*proxyBase}
	return fli, nil
}

func NewDataChainProxyForward(addrSelf, addrFriend nodeStore.AddressNet) *ProxyForward {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_forward, addrSelf, addrFriend, nodeStore.AddressNet{})
	f := ProxyForward{DataChainProxyBase: *proxyBase}
	return &f
}
