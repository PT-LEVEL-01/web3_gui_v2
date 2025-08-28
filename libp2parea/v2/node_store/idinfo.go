package nodeStore

import (
	"bytes"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/ed25519"
	"web3_gui/keystore/v2"
	keysconfig "web3_gui/keystore/v2/config"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

// Id信息
type IdInfo struct {
	Id   *AddressNet       `json:"id"`   //id，节点网络地址
	EPuk ed25519.PublicKey `json:"epuk"` //ed25519公钥，身份密钥的公钥
	CPuk keystore.Key      `json:"cpuk"` //curve25519公钥,DH公钥
	V    uint32            `json:"v"`    //DH公钥版本，低版本将被弃用，用于自动升级更换DH公钥协议
	Sign []byte            `json:"sign"` //ed25519私钥签名,Sign(V + CPuk)
	// Ctype string           `json:"ctype"` //签名方法 如ecdsa256 ecdsa512
}

/*
给idInfo签名
*/
func (this *IdInfo) SignDHPuk(prk ed25519.PrivateKey) utils.ERROR {
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.LittleEndian, this.V)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	n, err := buf.Write(this.CPuk[:])
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if n != len(this.CPuk) {
		return utils.NewErrorBus(keysconfig.ERROR_code_size_too_small, "")
	}
	this.Sign = ed25519.Sign(prk, buf.Bytes())
	return utils.NewErrorSuccess()
}

/*
验证签名
*/
func (this *IdInfo) CheckSignDHPuk() (bool, utils.ERROR) {
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.LittleEndian, this.V)
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	n, err := buf.Write(this.CPuk[:])
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	if n != len(this.CPuk) {
		return false, utils.NewErrorBus(config.ERROR_code_read_or_write_size_fail, "CheckSignDHPuk")
	}
	return ed25519.Verify(this.EPuk, buf.Bytes(), this.Sign), utils.NewErrorSuccess()
}

func (this *IdInfo) Conver() *go_protobuf.IdInfoV2 {
	idinfo := go_protobuf.IdInfoV2{
		Id:   this.Id.addr,
		EPuk: this.EPuk,
		CPuk: this.CPuk[:],
		V:    this.V,
		Sign: this.Sign,
	}
	return &idinfo
}

func (this *IdInfo) Proto() ([]byte, error) {
	idinfo := this.Conver()
	return idinfo.Marshal()
}

func ConverIdInfo(idInfoV2 *go_protobuf.IdInfoV2) *IdInfo {
	var cpuk keystore.Key = [32]byte{}
	copy(cpuk[:], idInfoV2.CPuk)
	idInfo := IdInfo{
		Id:   NewAddressNet(idInfoV2.Id),
		EPuk: idInfoV2.EPuk,
		CPuk: cpuk,
		V:    idInfoV2.V,
		Sign: idInfoV2.Sign,
	}
	return &idInfo
}

/*
检查idInfo是否合法
1.地址生成合法
2.签名正确
@return   true:合法;false:不合法;
*/
func CheckIdInfo(idInfo IdInfo) (bool, utils.ERROR) {
	//验证签名
	ok, ERR := idInfo.CheckSignDHPuk()
	if ERR.CheckFail() {
		return false, ERR
	}
	if !ok {
		return false, utils.NewErrorSuccess()
	}
	//验证地址
	return CheckPukAddrNet(idInfo.EPuk, idInfo.Id)
}

func ParseIdInfo(bs []byte) (*IdInfo, error) {
	iip := new(go_protobuf.IdInfoV2)
	err := proto.Unmarshal(bs, iip)
	if err != nil {
		return nil, err
	}
	var cpuk keystore.Key = [32]byte{}
	copy(cpuk[:], iip.CPuk)
	idInfo := IdInfo{
		Id:   NewAddressNet(iip.Id), //id，节点网络地址
		EPuk: iip.EPuk,              //ed25519公钥，身份密钥的公钥
		CPuk: cpuk,                  //curve25519公钥,DH公钥
		V:    iip.V,                 //DH公钥版本，低版本将被弃用，用于自动升级更换DH公钥协议
		Sign: iip.Sign,              //ed25519私钥签名,Sign(V + CPuk)
	}
	return &idInfo, nil
}
