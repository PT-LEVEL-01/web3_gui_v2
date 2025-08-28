package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_group_addMember, ParseGroupAddMemberFactory)
}

/*
管理员同意用户入群
*/
type DatachainGroupAddMember struct {
	DataChainClientBase
	GroupID []byte //群ID
	Token   []byte //令牌
}

func (this *DatachainGroupAddMember) CheckCmd() bool {
	return true
}

func (this *DatachainGroupAddMember) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainGroupAccept{
		Base:    base,
		GroupID: this.GroupID, //群ID
		Token:   this.Token,
	}
	bs, err := sendText.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
解析
*/
func ParseGroupAddMemberFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainGroupAccept{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainGroupAddMember{
		DataChainClientBase: *clientBase,
		GroupID:             base.GroupID, //
		Token:               base.Token,
	}
	return &addFriend, nil
}

func NewDatachainGroupAddMember(addrSelf, addrFirend nodeStore.AddressNet, groupId, token []byte) *DatachainGroupAddMember {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_group_addMember, addrSelf, addrFirend)
	groupAccept := DatachainGroupAddMember{DataChainClientBase: *clientBase, GroupID: groupId, Token: token}
	forward.clientItr = &groupAccept
	groupAccept.SetProxyItr(forward)
	return &groupAccept
}
