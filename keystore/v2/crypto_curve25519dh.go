package keystore

import (
	"bytes"
	"encoding/hex"
	"golang.org/x/crypto/curve25519"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
私钥和公钥
*/
type Key [32]byte

// 桁条接口符合性。
func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

/*
DH秘钥交换结构体
*/
type DHPair struct {
	privateKey Key
	publicKey  Key
}

func (this DHPair) GetPrivateKey() Key {
	return this.privateKey
}
func (this DHPair) GetPublicKey() Key {
	return this.publicKey
}

/*
创建一个DH密钥对
*/
func NewDHPair(prk, puk Key) DHPair {
	return DHPair{
		privateKey: prk,
		publicKey:  puk,
	}
}

/*
公钥私钥对
*/
type KeyPair struct {
	PublicKey  Key
	PrivateKey Key
}

/*
创建一个DH密钥对
*/
func NewKeyPair(prk, puk []byte) *KeyPair {
	prkBs := [32]byte{}
	copy(prkBs[:], prk)
	pukBs := [32]byte{}
	copy(pukBs[:], puk)
	return &KeyPair{
		PrivateKey: prkBs,
		PublicKey:  pukBs,
	}
}

func (this *KeyPair) GetPrivateKey() Key {
	return this.PrivateKey
}
func (this *KeyPair) GetPublicKey() Key {
	return this.PublicKey
}

/*
生成公钥私钥对
*/
func GenerateKeyPair(rand []byte) (KeyPair, utils.ERROR) {
	size := 32
	if len(rand) < size {
		//随机数长度不够
		return KeyPair{}, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	var priv [32]byte
	//用随机数填满私钥
	buf := bytes.NewBuffer(rand)
	n, err := buf.Read(priv[:])
	if err != nil {
		return KeyPair{}, utils.NewErrorSysSelf(err)
	}
	if n != size {
		//读取私钥长度不够
		return KeyPair{}, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &priv)
	return KeyPair{
		PrivateKey: priv,
		PublicKey:  pubKey,
	}, utils.NewErrorSuccess()
}

/*
协商密钥
*/
func KeyExchange(dh DHPair) ([32]byte, error) {
	var (
		sharedSecret [32]byte
		priv         [32]byte = dh.GetPrivateKey()
		pub          [32]byte = dh.GetPublicKey()
	)
	temp, err := curve25519.X25519(priv[:], pub[:])
	if err != nil {
		return sharedSecret, err
	}
	copy(sharedSecret[:], temp)
	return sharedSecret, nil
}
