package imdatachain

import (
	"time"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_msglog_add, ParseDatachainMsgLogFactory)
}

/*
需要记录的消息
*/
type MsgLog struct {
	DataChainProxyBase //
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *MsgLog) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrTo
}

func (this *MsgLog) Forward() bool {
	return false
}

/*
 */
func (this *MsgLog) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDatachainMsgLogFactory(bs []byte) (DataChainProxyItr, error) {
	proxyBase, err := ParseDatachainProxyBase(bs)
	if err != nil {
		return nil, err
	}
	fli := &MsgLog{DataChainProxyBase: *proxyBase}
	return fli, nil
}

func NewDataChainProxyMsgLog(proxyItr DataChainProxyItr) *MsgLog {
	clientItr := proxyItr.GetClientItr()
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_msglog_add, proxyItr.GetAddrFrom(), proxyItr.GetAddrTo(), nodeStore.AddressNet{})
	proxyBase.SendID = proxyItr.GetID()
	proxyBase.SendIndex = proxyItr.GetBase().SendIndex
	proxyBase.clientItr = clientItr
	proxyBase.EncryptType = proxyItr.GetBase().EncryptType
	proxyBase.Content = proxyItr.GetBase().Content
	proxyBase.DhPuk = proxyItr.GetBase().DhPuk
	proxyBase.SendTime = proxyItr.GetSendTime()
	proxyBase.RecvTime = time.Now().Unix()
	f := MsgLog{DataChainProxyBase: *proxyBase}
	if clientItr != nil {
		clientItr.SetProxyItr(&f)
	}
	return &f
}
