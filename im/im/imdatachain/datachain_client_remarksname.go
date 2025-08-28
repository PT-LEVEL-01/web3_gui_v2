package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_remarksname, ParseClientRemarksnameFactory)
}

type ClientRemarksname struct {
	DataChainClientBase
	Addr        nodeStore.AddressNet //好友地址
	Remarksname string               //备注昵称
}

func (this *ClientRemarksname) CheckCmd() bool {
	return true
}

func (this *ClientRemarksname) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainAgreeFriend{
		Base:        base,
		TokenLocal:  this.Addr.GetAddr(),
		TokenRemote: []byte(this.Remarksname),
	}
	bs, err := sendText.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseClientRemarksnameFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainAgreeFriend{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := ClientRemarksname{
		DataChainClientBase: *clientBase,
		Addr:                *nodeStore.NewAddressNet(base.TokenLocal),
		Remarksname:         string(base.TokenRemote),
	}
	return &addFriend, nil
}

func NewClientRemarksname(addrSelf, addrFirend nodeStore.AddressNet, remarksName string) *ClientRemarksname {
	proxySetup := NewProxySetup(addrSelf)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_remarksname, addrSelf, nodeStore.AddressNet{})
	addFriend := ClientRemarksname{DataChainClientBase: *clientBase, Addr: addrFirend, Remarksname: remarksName}
	proxySetup.clientItr = &addFriend
	addFriend.SetProxyItr(proxySetup)
	return &addFriend
}
