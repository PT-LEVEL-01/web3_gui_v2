package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_group_invitation, ParseGroupInvitationFactory)
}

/*
邀请用户入群
*/
type DatachainGroupInvitation struct {
	DataChainClientBase
	GroupID       []byte               //群ID
	AdminAddr     nodeStore.AddressNet //管理员地址
	Token         []byte               //令牌
	Nickname      string               //好友昵称
	GroupNickname string               //群昵称
}

func (this *DatachainGroupInvitation) CheckCmd() bool {
	return true
}

func (this *DatachainGroupInvitation) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainInvitation{
		Base:          base,
		GroupID:       this.GroupID,
		AdminAddr:     this.AdminAddr.GetAddr(),
		Token:         this.Token,
		Nickname:      []byte(this.Nickname),
		GroupNickname: []byte(this.GroupNickname),
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
func ParseGroupInvitationFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainInvitation{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainGroupInvitation{
		DataChainClientBase: *clientBase,
		GroupID:             base.GroupID, //
		AdminAddr:           *nodeStore.NewAddressNet(base.AdminAddr),
		Token:               base.Token,
		Nickname:            string(base.Nickname),
		GroupNickname:       string(base.GroupNickname),
	}
	return &addFriend, nil
}

func NewDatachainGroupInvitation(addrSelf, addrFirend, addrAdmin nodeStore.AddressNet, groupId []byte,
	nickname, groupNickname string) *DatachainGroupInvitation {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_group_invitation, addrSelf, addrFirend)
	addFriend := DatachainGroupInvitation{DataChainClientBase: *clientBase, GroupID: groupId, AdminAddr: addrAdmin,
		Token: nil, Nickname: nickname, GroupNickname: groupNickname}
	forward.clientItr = &addFriend
	addFriend.SetProxyItr(forward)
	return &addFriend
}
