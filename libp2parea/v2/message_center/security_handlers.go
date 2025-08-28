package message_center

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
)

/*
获取节点地址和身份公钥
*/
func (this *MessageCenter) SearchAddress(message *MessageBase) {
	if this.CheckRepeatHash(message.GetBase().SendID) {
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
	this.SendP2pReplyMsg(message, &data)

}

/*
获取节点地址和身份公钥_返回
*/
func (this *MessageCenter) SearchAddress_recv(message *MessageBase) {
	//// utils.Log.Info().Msgf("%s 收到密钥对返回消息", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	//if this.CheckRepeatHash(message.Body.ReplyHash) {
	//	// utils.Log.Info().Msgf("验证返回消息错误")
	//	return
	//}
	//// fmt.Println("收到Hello消息", string(*message.Body.Content))
	//
	//// idinfo := nodeStore.Parse(*message.Body.Content)
	//idinfo, _ := nodeStore.ParseIdInfo(*message.Body.Content)
	//sni := SearchNodeInfo{
	//	Id:      message.Head.Sender,
	//	SuperId: message.Head.SenderSuperId,
	//	CPuk:    idinfo.CPuk,
	//}
	//// bs, _ := json.Marshal(sni)
	//bs, _ := sni.Proto() //json.Marshal(sni)
	//
	//// flood.ResponseWait(config.CLASS_security_searchAddr, utils.Bytes2string(message.Body.Hash), &bs)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), &bs)

}

type SearchNodeInfo struct {
	Id      *nodeStore.AddressNet //
	SuperId *nodeStore.AddressNet //
	CPuk    keystore.Key          //
}

func (this *SearchNodeInfo) Proto() ([]byte, error) {
	snip := go_protobuf.SearchNodeInfoV2{
		Id:      this.Id.GetAddr(),
		SuperId: this.SuperId.GetAddr(),
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
	snip := new(go_protobuf.SearchNodeInfoV2)
	err := proto.Unmarshal(bs, snip)
	if err != nil {
		return nil, err
	}
	var cpuk keystore.Key = [32]byte{}
	copy(cpuk[:], snip.CPuk)

	// id := nodeStore.AddressNet(snip.Id)
	// superId := nodeStore.AddressNet(snip.SuperId)
	sni := SearchNodeInfo{
		// Id:      &id,
		// SuperId: &superId,
		CPuk: cpuk,
	}
	if snip.Id != nil && len(snip.Id) > 0 {
		id := nodeStore.NewAddressNet(snip.Id)
		sni.Id = id
	}

	if snip.SuperId != nil && len(snip.SuperId) > 0 {
		superId := nodeStore.NewAddressNet(snip.SuperId) //nodeStore.AddressNet(snip.SuperId)
		sni.SuperId = superId
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
	A_DH_PUK keystore.Key     //A公钥
	B_DH_PUK keystore.Key     //B公钥
}

func (this *ShareKey) Proto() ([]byte, error) {
	// var cpuk dh.Key = [32]byte{}
	// copy(cpuk[:], this.Idinfo.CPuk)
	idinfo := go_protobuf.IdInfoV2{
		Id:   this.Idinfo.Id.GetAddr(),
		EPuk: this.Idinfo.EPuk,
		CPuk: this.Idinfo.CPuk[:],
		V:    this.Idinfo.V,
		Sign: this.Idinfo.Sign,
	}
	skp := go_protobuf.ShareKeyV2{
		Idinfo:   &idinfo,
		A_DH_PUK: this.A_DH_PUK[:],
		B_DH_PUK: this.B_DH_PUK[:],
	}
	return skp.Marshal()
}

func ParseShareKey(bs []byte) (*ShareKey, error) {
	skp := new(go_protobuf.ShareKeyV2)
	err := proto.Unmarshal(bs, skp)
	if err != nil {
		return nil, err
	}
	var cpuk keystore.Key = [32]byte{}
	copy(cpuk[:], skp.Idinfo.CPuk)

	idinfo := nodeStore.IdInfo{
		Id:   nodeStore.NewAddressNet(skp.Idinfo.Id),
		EPuk: skp.Idinfo.EPuk,
		CPuk: cpuk,
		V:    skp.Idinfo.V,
		Sign: skp.Idinfo.Sign,
	}
	var apuk, bpuk keystore.Key = [32]byte{}, [32]byte{}
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
func (this *MessageCenter) CreatePipe(message *MessageBase) {
	messageBase := message.GetBase()
	//utils.Log.Info().Msgf("收到对方公钥并创建加密通道")
	if this.CheckRepeatHash(messageBase.SendID) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}

	shareKey, err := ParseShareKey(messageBase.Content)
	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	// shareKey := new(ShareKey)
	// // err := json.Unmarshal(*message.Body.Content, shareKey)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(shareKey)
	if err != nil {
		return
	}
	keyPair, ERR := this.key.GetDhAddrKeyPair(this.pwd)
	if ERR.CheckFail() {
		return
	}
	sk, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.Idinfo.CPuk))
	if err != nil {
		return
	}
	sharedHka, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.A_DH_PUK))
	if err != nil {
		return
	}
	sharedNhkb, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.B_DH_PUK))
	if err != nil {
		return
	}
	err = this.RatchetSession.AddRecvPipe(messageBase.SenderAddr, messageBase.SenderMachineID, sk, sharedHka, sharedNhkb, shareKey.Idinfo.CPuk)
	if err != nil {
		return
	}

	//回复消息
	// data := nodeStore.NodeSelf.IdInfo.JSON()
	data, _ := this.nodeManager.NodeSelf.IdInfo.Proto()
	this.SendP2pReplyMsg(message, &data)

}

/*
获取节点地址和身份公钥_返回
*/
func (this *MessageCenter) CreatePipe_recv(message *MessageBase) {
	//if this.CheckRepeatHash(message.GetBase().ReplyID) {
	//	// utils.Log.Info().Msgf("验证返回消息错误")
	//	return
	//}
	//// flood.ResponseWait(config.CLASS_im_security_create_pipe, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
解密错误
*/
func (this *MessageCenter) Pipe_error(message *MessageBase) {
	if this.CheckRepeatHash(message.GetBase().SendID) {
		// utils.Log.Info().Msgf("验证返回消息错误")
		return
	}

	// this.RatchetSession.RemoveSendPipe(*message.Head.Sender, message.Head.SenderMachineID)
	this.CleanHEInfo(message.GetBase().SenderAddr, message.GetBase().SenderMachineID)
}
