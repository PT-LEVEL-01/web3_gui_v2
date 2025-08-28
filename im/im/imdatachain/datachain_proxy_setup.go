package imdatachain

import (
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_setup, ParseProxySetupFactory)
}

/*
需要记录的消息
*/
type ProxySetup struct {
	DataChainProxyBase //
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *ProxySetup) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

func (this *ProxySetup) Forward() bool {
	return false
}

/*
 */
func (this *ProxySetup) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseProxySetupFactory(bs []byte) (DataChainProxyItr, error) {
	proxyBase, err := ParseDatachainProxyBase(bs)
	if err != nil {
		return nil, err
	}
	fli := &ProxySetup{DataChainProxyBase: *proxyBase}
	return fli, nil
}

func NewProxySetup(addrSelf nodeStore.AddressNet) *ProxySetup {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_setup, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	f := ProxySetup{DataChainProxyBase: *proxyBase}
	return &f
}
