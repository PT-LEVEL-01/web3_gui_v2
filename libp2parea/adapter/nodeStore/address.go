package nodeStore

import (
	"bytes"
	"crypto/sha256"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/utils"
)

type AddressNet []byte

func (this AddressNet) B58String() string {
	if len(this) <= 0 {
		return ""
	}
	lastByte := (this)[len(this)-1:]
	lastStr := string(base58.Encode(lastByte))
	if len(lastByte) == 0 {
		return ""
	}
	preLen := int(lastByte[0])
	preStr := string((this)[:preLen])
	centerStr := string(base58.Encode((this)[preLen : len(this)-1]))
	return preStr + centerStr + lastStr
}

func (this AddressNet) GetPre() string {
	if len(this) <= 0 {
		return ""
	}
	lastByte := (this)[len(this)-1:]
	if len(lastByte) == 0 {
		return ""
	}
	preLen := int(lastByte[0])
	preStr := string((this)[:preLen])
	return preStr
}

/*
有效数据部分
*/
func (this AddressNet) Data() []byte {
	if len(this) == 0 {
		return nil
	}
	lastByte := (this)[len(this)-1:]
	if len(lastByte) == 0 {
		return nil
	}
	preLen := int(lastByte[0])
	return (this)[preLen : len(this)-(4+1)]
}

func AddressFromB58String(str string) AddressNet {
	if str == "" {
		return nil
	}
	lastStr := str[len(str)-1:]
	lastByte := base58.Decode(lastStr)
	if len(lastByte) == 0 {
		return nil
	}
	preLen := int(lastByte[0])
	if preLen > len(str) {
		return nil
	}
	preStr := str[:preLen]
	preByte := []byte(preStr)
	centerByte := base58.Decode(str[preLen : len(str)-1])
	bs := make([]byte, 0, len(preByte)+len(centerByte)+len(lastByte))
	bs = append(bs, preByte...)
	bs = append(bs, centerByte...)
	bs = append(bs, lastByte...)
	return AddressNet(bs)
}

// 节点地址
//type AddressNet []byte
//
//func (this AddressNet) B58String() string {
//	return string(base58.Encode(this))
//}
//
//func AddressFromB58String(str string) AddressNet {
//	return AddressNet(base58.Decode(str))
//}

/*
通过公钥生成网络节点地址，将公钥两次hash得到网络节点地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pubKey []byte) AddressNet {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(publicSHA256[:])
	return temp[:]
}

/*
检查公钥生成的地址是否一样
@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddr(pubKey []byte, addr AddressNet) bool {
	tagAddr := BuildAddr(pubKey)
	return bytes.Equal(tagAddr, addr)
}

/*
去除重复地址
*/
func RemoveDuplicateAddress(addrs []*AddressNet) []*AddressNet {
	m := make(map[string]*AddressNet)
	for i, one := range addrs {
		// m[hex.EncodeToString(*one)] = addrs[i]
		m[utils.Bytes2string(*one)] = addrs[i]
	}
	dstAddrs := make([]*AddressNet, 0)
	for _, v := range m {
		dstAddrs = append(dstAddrs, v)
	}
	return dstAddrs
}
