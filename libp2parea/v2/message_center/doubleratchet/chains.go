package doubleratchet

import (
	"web3_gui/keystore/v2"
)

// KDFer为链执行键派生函数。
type KDFer interface {
	// KdfRK 返回一对密钥(根密钥key, 链密钥key)
	KdfRK(rk, dhOut keystore.Key) (rootKey, chainKey, newHeaderKey keystore.Key) //

	// KdfCK returns a pair (32-byte chain key, 32-byte message key) as the output of applying
	// a KDF keyed by a 32-byte chain key ck to some constant.
	//通过链密钥ck作为常量，作为kdf的密钥，返回一对密钥（链密钥，消息密钥）
	KdfCK(ck keystore.Key) (chainKey, msgKey keystore.Key)
}

type kdfChain struct {
	Crypto KDFer

	//链密钥
	CK keystore.Key

	//链上消息总数。
	N uint32
}

// 步骤执行对称棘轮步骤并返回新的消息密钥。
func (this *kdfChain) step() keystore.Key {
	var mk keystore.Key
	this.CK, mk = this.Crypto.KdfCK(this.CK)
	this.N++
	return mk
}

type kdfRootChain struct {
	Crypto KDFer

	//kdf链密钥
	CK keystore.Key
}

// 步骤执行对称棘轮步骤，并返回新的链和新的头键。
func (this *kdfRootChain) step(kdfInput keystore.Key) (ch kdfChain, nhk keystore.Key) {
	ch = kdfChain{
		Crypto: this.Crypto,
	}
	this.CK, ch.CK, nhk = this.Crypto.KdfRK(this.CK, kdfInput)
	return ch, nhk
}
