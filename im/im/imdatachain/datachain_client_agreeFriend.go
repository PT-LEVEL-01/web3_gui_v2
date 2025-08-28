package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_agreeFriend, ParseAgreeFriendFactory)
}

type DatachainAgreeFriend struct {
	DataChainClientBase
	TokenLocal  []byte //自己的令牌
	TokenRemote []byte //对方的令牌
	DhPukInfo   []byte //公钥信息
}

func (this *DatachainAgreeFriend) CheckCmd() bool {
	return true
}

func (this *DatachainAgreeFriend) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainAgreeFriend{
		Base:        base,
		TokenLocal:  this.TokenLocal,
		TokenRemote: this.TokenRemote,
		DhPukInfo:   this.DhPukInfo,
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
func ParseAgreeFriendFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainAgreeFriend{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainAgreeFriend{
		DataChainClientBase: *clientBase,
		TokenLocal:          base.TokenLocal,
		TokenRemote:         base.TokenRemote,
		DhPukInfo:           base.DhPukInfo,
	}
	return &addFriend, nil
}

func NewDatachainAgreeFriend(addrSelf, addrFirend nodeStore.AddressNet, tokenLocal, tokenRemote, dhPukInfo []byte) *DatachainAgreeFriend {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	//forward.DhPuk = dhPukInfo
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_agreeFriend, addrSelf, addrFirend)
	addFriend := DatachainAgreeFriend{DataChainClientBase: *clientBase, TokenLocal: tokenLocal, TokenRemote: tokenRemote, DhPukInfo: dhPukInfo}
	forward.clientItr = &addFriend
	addFriend.SetProxyItr(forward)
	return &addFriend
}
