package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_group_accept, ParseGroupAcceptFactory)
}

/*
用户接受入群
*/
type DatachainGroupAccept struct {
	DataChainClientBase
	GroupID       []byte //群ID
	AcceptTime    int64  //同意入群时间
	Sign          []byte //确认加入群聊的签名
	SignPuk       []byte //签名使用的公钥
	DHPuk         []byte //协商密钥用公钥
	Token         []byte //本地令牌
	Nickname      string //好友昵称
	GroupNickname string //群昵称
}

func (this *DatachainGroupAccept) CheckCmd() bool {
	return true
}

func (this *DatachainGroupAccept) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainGroupAccept{
		Base:          base,
		GroupID:       this.GroupID,               //群ID
		AcceptTime:    this.AcceptTime,            //
		Sign:          this.Sign,                  //确认加入群聊的签名
		SignPuk:       this.SignPuk,               //签名使用的公钥
		DHPuk:         this.DHPuk,                 //协商密钥用公钥
		Token:         this.Token,                 //
		Nickname:      []byte(this.Nickname),      //
		GroupNickname: []byte(this.GroupNickname), //
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
func ParseGroupAcceptFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainGroupAccept{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainGroupAccept{
		DataChainClientBase: *clientBase,
		GroupID:             base.GroupID,               //
		AcceptTime:          base.AcceptTime,            //
		Sign:                base.Sign,                  //
		SignPuk:             base.SignPuk,               //签名使用的公钥
		DHPuk:               base.DHPuk,                 //协商密钥用公钥
		Token:               base.Token,                 //
		Nickname:            string(base.Nickname),      //
		GroupNickname:       string(base.GroupNickname), //
	}
	return &addFriend, nil
}

func NewDatachainGroupAccept(addrSelf, addrFirend nodeStore.AddressNet, groupId []byte, acceptTime int64, sign, signPuk,
	dhPuk, tokenLocal []byte, nickname, groupNickname string) *DatachainGroupAccept {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_group_accept, addrSelf, addrFirend)
	groupAccept := DatachainGroupAccept{DataChainClientBase: *clientBase, GroupID: groupId, AcceptTime: acceptTime,
		Sign: sign, SignPuk: signPuk, DHPuk: dhPuk, Token: tokenLocal, Nickname: nickname, GroupNickname: groupNickname}
	forward.clientItr = &groupAccept
	groupAccept.SetProxyItr(forward)
	return &groupAccept
}
