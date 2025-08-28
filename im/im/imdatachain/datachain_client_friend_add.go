package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_addFriend, ParseAddFriendFactory)
}

type DatachainAddrFriend struct {
	DataChainClientBase
	Nickname string //发送自己的昵称
}

func (this *DatachainAddrFriend) CheckCmd() bool {
	return true
}

func (this *DatachainAddrFriend) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainAgreeFriend{
		Base:       base,
		TokenLocal: []byte(this.Nickname),
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
func ParseAddFriendFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainAgreeFriend{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainAddrFriend{
		DataChainClientBase: *clientBase,
		Nickname:            string(base.TokenLocal),
	}
	return &addFriend, nil
}

func NewDatachainAddFriend(addrSelf, addrFirend nodeStore.AddressNet, dhPukInfo []byte, nickname string) *DatachainAddrFriend {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	forward.DhPuk = dhPukInfo
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_addFriend, addrSelf, addrFirend)
	addFriend := DatachainAddrFriend{DataChainClientBase: *clientBase, Nickname: nickname}
	forward.clientItr = &addFriend
	addFriend.SetProxyItr(forward)
	return &addFriend
}
