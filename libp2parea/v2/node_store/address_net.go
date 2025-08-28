package nodeStore

import (
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

// 节点地址
//type AddressNet []byte

type AddressNet struct {
	addr    coin_address.AddressCoin //地址数据部分
	pre     string                   //地址前缀
	data    []byte                   //地址数据段
	addrStr string                   //完整地址base58字符串
}

func NewAddressNet(addr coin_address.AddressCoin) *AddressNet {
	addrNet := AddressNet{addr: addr}
	return &addrNet
}

func (this *AddressNet) B58String() string {
	if this.addrStr != "" {
		return this.addrStr
	}
	this.addrStr = this.addr.B58String()
	return this.addrStr
}
func (this AddressNet) GetAddr() coin_address.AddressCoin {
	return this.addr
}

func (this *AddressNet) GetPre() string {
	if this.pre != "" {
		return this.pre
	}
	this.pre = this.addr.GetPre()
	return this.pre
}

/*
有效数据部分
*/
func (this *AddressNet) Data() []byte {
	return this.addr.Data()
}

func AddressFromB58String(str string) AddressNet {
	return *NewAddressNet(keystore.AddressFromB58String(str))
}

/*
通过公钥生成网络节点地址，将公钥两次hash得到网络节点地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) (*AddressNet, utils.ERROR) {
	addr, ERR := coin_address.BuildAddr(pre, pubKey)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return NewAddressNet(addr), utils.NewErrorSuccess()
}

/*
通过公钥生成网络节点地址，将公钥两次hash得到网络节点地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddrByData(pre string, data []byte) (AddressNet, utils.ERROR) {
	addr, ERR := keystore.BuildAddrByData(pre, data)
	if ERR.CheckFail() {
		return AddressNet{}, ERR
	}
	//utils.Log.Info().Int("地址长度", len(addr)).Send()
	return *NewAddressNet(addr), utils.NewErrorSuccess()
}

/*
检查公钥生成的网络地址是否一样
@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddrNet(pubKey []byte, addr *AddressNet) (bool, utils.ERROR) {
	return keystore.CheckPukAddrNet(addr.GetPre(), pubKey, addr.addr)
}

/*
去除重复地址
*/
func RemoveDuplicateAddress(addrs []*AddressNet) []*AddressNet {
	m := make(map[string]*AddressNet)
	for i, one := range addrs {
		// m[hex.EncodeToString(*one)] = addrs[i]
		m[utils.Bytes2string(one.Data())] = addrs[i]
	}
	dstAddrs := make([]*AddressNet, 0)
	for _, v := range m {
		dstAddrs = append(dstAddrs, v)
	}
	return dstAddrs
}
