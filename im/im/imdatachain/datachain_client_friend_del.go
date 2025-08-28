package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_del, ParseDelFriendFactory)
}

type DatachainDelFriend struct {
	DataChainClientBase
}

func (this *DatachainDelFriend) CheckCmd() bool {
	return true
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDelFriendFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainClientBase{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainDelFriend{
		DataChainClientBase: *clientBase,
	}
	return &addFriend, nil
}

func NewDatachainDelFriend(addrSelf, addrFirend nodeStore.AddressNet, dhPukInfo []byte) *DatachainDelFriend {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	forward.DhPuk = dhPukInfo
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_del, addrSelf, addrFirend)
	addFriend := DatachainDelFriend{*clientBase}
	forward.clientItr = &addFriend
	addFriend.SetProxyItr(forward)
	return &addFriend
}
