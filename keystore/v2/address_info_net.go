package keystore

//
//import (
//	"web3_gui/keystore/v1/base58"
//)
//
//type AddressNet struct {
//	pre  string //地址前缀
//	data []byte //地址内容
//}
//
////type AddressNet []byte
//
//func (this *AddressNet) B58String() string {
//	if len(*this) <= 0 {
//		return ""
//	}
//	return string(base58.Encode(*this))
//}
//
//func ParseNetAddrFromB58String(str string) AddressNet {
//	if str == "" {
//		return nil
//	}
//	lastStr := str[len(str)-1:]
//	lastByte := base58.Decode(lastStr)
//	if len(lastByte) <= 0 {
//		return nil
//	}
//	preLen := int(lastByte[0])
//	if preLen > len(str) {
//		return nil
//	}
//	preStr := str[:preLen]
//	preByte := []byte(preStr)
//	centerByte := base58.Decode(str[preLen : len(str)-1])
//	bs := make([]byte, 0, len(preByte)+len(centerByte)+len(lastByte))
//	bs = append(bs, preByte...)
//	bs = append(bs, centerByte...)
//	bs = append(bs, lastByte...)
//	return AddressNet(bs)
//}
