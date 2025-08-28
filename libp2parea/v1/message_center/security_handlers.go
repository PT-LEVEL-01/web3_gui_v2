package message_center

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/keystore/v1/crypto/dh"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
获取节点地址和身份公钥
*/
func (this *MessageCenter) SearchAddress(c engine.Controller, msg engine.Packet, message *Message) {
	if this.CheckRepeatHash(message.Body.Hash) {
		return
	}

	// prk := keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	// puk := keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	// utils.Log.Info().Msgf("%s 密钥对公钥:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), hex.EncodeToString(this.nodeManager.NodeSelf.IdInfo.CPuk[:]))
	// fmt.Println("密钥对", hex.EncodeToString(prk[:]), "\n",
	// 	hex.EncodeToString(puk[:]))
	// fmt.Println("获取节点地址和身份公钥")

	//回复消息
	// data := nodeStore.NodeSelf.IdInfo.JSON()
	data, _ := this.nodeManager.NodeSelf.IdInfo.Proto()
	this.SendP2pReplyMsg(message, config.MSGID_SearchAddr_recv, &data)

}

/*
获取节点地址和身份公钥_返回
*/
func (this *MessageCenter) SearchAddress_recv(c engine.Controller, msg engine.Packet, message *Message) {
	// utils.Log.Info().Msgf("%s 收到密钥对返回消息", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	if this.CheckRepeatHash(message.Body.ReplyHash) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}
	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	// idinfo := nodeStore.Parse(*message.Body.Content)
	idinfo, _ := nodeStore.ParseIdInfo(*message.Body.Content)
	sni := SearchNodeInfo{
		Id:      message.Head.Sender,
		SuperId: message.Head.SenderSuperId,
		CPuk:    idinfo.CPuk,
	}
	// bs, _ := json.Marshal(sni)
	bs, _ := sni.Proto() //json.Marshal(sni)

	// flood.ResponseWait(config.CLASS_security_searchAddr, utils.Bytes2string(message.Body.Hash), &bs)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), &bs)

}

type SearchNodeInfo struct {
	Id      *nodeStore.AddressNet //
	SuperId *nodeStore.AddressNet //
	CPuk    dh.Key                //
}

func (this *SearchNodeInfo) Proto() ([]byte, error) {
	snip := go_protobuf.SearchNodeInfo{
		Id:      *this.Id,
		SuperId: *this.SuperId,
		CPuk:    this.CPuk[:],
	}
	return snip.Marshal()
}

// func ParserSearchNodeInfo(bs *[]byte) (*SearchNodeInfo, error) {
// 	sni := new(SearchNodeInfo)
// 	// err := json.Unmarshal(*bs, sni)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(sni)
// 	return sni, err
// }

func ParserSearchNodeInfo(bs []byte) (*SearchNodeInfo, error) {
	snip := new(go_protobuf.SearchNodeInfo)
	err := proto.Unmarshal(bs, snip)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], snip.CPuk)

	// id := nodeStore.AddressNet(snip.Id)
	// superId := nodeStore.AddressNet(snip.SuperId)
	sni := SearchNodeInfo{
		// Id:      &id,
		// SuperId: &superId,
		CPuk: cpuk,
	}
	if snip.Id != nil && len(snip.Id) > 0 {
		id := nodeStore.AddressNet(snip.Id)
		sni.Id = &id
	}

	if snip.SuperId != nil && len(snip.SuperId) > 0 {
		superId := nodeStore.AddressNet(snip.SuperId)
		sni.SuperId = &superId
	}

	return &sni, nil
	// err := json.Unmarshal(*bs, sni)
	// decoder := json.NewDecoder(bytes.NewBuffer(bs))
	// decoder.UseNumber()
	// err := decoder.Decode(sni)
	// return sni, err
}

type ShareKey struct {
	// IV_DH_PUK dh.Key//向量
	Idinfo   nodeStore.IdInfo //身份密钥公钥
	A_DH_PUK dh.Key           //A公钥
	B_DH_PUK dh.Key           //B公钥
}

func (this *ShareKey) Proto() ([]byte, error) {
	// var cpuk dh.Key = [32]byte{}
	// copy(cpuk[:], this.Idinfo.CPuk)
	idinfo := go_protobuf.IdInfo{
		Id:   this.Idinfo.Id,
		EPuk: this.Idinfo.EPuk,
		CPuk: this.Idinfo.CPuk[:],
		V:    this.Idinfo.V,
		Sign: this.Idinfo.Sign,
	}
	skp := go_protobuf.ShareKey{
		Idinfo:   &idinfo,
		A_DH_PUK: this.A_DH_PUK[:],
		B_DH_PUK: this.B_DH_PUK[:],
	}
	return skp.Marshal()
}

func ParseShareKey(bs []byte) (*ShareKey, error) {
	skp := new(go_protobuf.ShareKey)
	err := proto.Unmarshal(bs, skp)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], skp.Idinfo.CPuk)

	idinfo := nodeStore.IdInfo{
		Id:   skp.Idinfo.Id,
		EPuk: skp.Idinfo.EPuk,
		CPuk: cpuk,
		V:    skp.Idinfo.V,
		Sign: skp.Idinfo.Sign,
	}
	var apuk, bpuk dh.Key = [32]byte{}, [32]byte{}
	copy(apuk[:], skp.A_DH_PUK)
	copy(bpuk[:], skp.B_DH_PUK)

	shareKey := ShareKey{
		Idinfo:   idinfo,
		A_DH_PUK: apuk,
		B_DH_PUK: bpuk,
	}
	return &shareKey, nil
}

/*
获取节点地址和身份公钥
*/
func (this *MessageCenter) CreatePipe(c engine.Controller, msg engine.Packet, message *Message) {
	//utils.Log.Info().Msgf("收到对方公钥并创建加密通道")
	if this.CheckRepeatHash(message.Body.Hash) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}

	shareKey, err := ParseShareKey(*message.Body.Content)
	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	// shareKey := new(ShareKey)
	// // err := json.Unmarshal(*message.Body.Content, shareKey)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(shareKey)
	if err != nil {
		return
	}

	sk, err := dh.KeyExchange(dh.NewDHPair(this.key.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.Idinfo.CPuk))
	if err != nil {
		return
	}
	sharedHka, err := dh.KeyExchange(dh.NewDHPair(this.key.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.A_DH_PUK))
	if err != nil {
		return
	}
	sharedNhkb, err := dh.KeyExchange(dh.NewDHPair(this.key.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.B_DH_PUK))
	if err != nil {
		return
	}
	err = this.RatchetSession.AddRecvPipe(*message.Head.Sender, message.Head.SenderMachineID, sk, sharedHka, sharedNhkb, shareKey.Idinfo.CPuk)
	if err != nil {
		return
	}

	//回复消息
	// data := nodeStore.NodeSelf.IdInfo.JSON()
	data, _ := this.nodeManager.NodeSelf.IdInfo.Proto()
	this.SendP2pReplyMsg(message, config.MSGID_security_create_pipe_recv, &data)

}

/*
获取节点地址和身份公钥_返回
*/
func (this *MessageCenter) CreatePipe_recv(c engine.Controller, msg engine.Packet, message *Message) {
	if this.CheckRepeatHash(message.Body.ReplyHash) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}
	// flood.ResponseWait(config.CLASS_im_security_create_pipe, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
解密错误
*/
func (this *MessageCenter) Pipe_error(c engine.Controller, msg engine.Packet, message *Message) {
	if this.CheckRepeatHash(message.Body.Hash) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}

	// this.RatchetSession.RemoveSendPipe(*message.Head.Sender, message.Head.SenderMachineID)
	this.CleanHEInfo(*message.Head.Sender, message.Head.SenderMachineID)
}
