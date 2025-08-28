package keystore

import (
	"fmt"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"sync"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
BIP44，全称是Bitcoin Improvement Proposal 44，中文名为多币种和多账户的规范，是比特币的一项提议。

BIP44基于种子（由BIP39生成）和BIP32路径，为确定性钱包定义了一个逻辑层次结构，并在此基础上定义了五层的树状路径。BIP44允许生成和管理多个币种和多个账户，每个账户有自己的接收和更改地址。

跟随这个规范，HD钱包能根据一个种子生成多元化的钱包系统，即你可以使用一个私钥种子生成并管理不同币种的账户与地址。

BIP44定义的路径结构是：m / purpose’ / coin_type’ / account’ / change / address_index：

Purpose (目的)：在BIP44中，目的一直被设置为44'。
Coin Type (币种)：这一层用于区别不同的数字货币，比如0代表比特币，1代表测试网比特币，60代表以太坊等等。完整的币种列表地址在这里。
Account (账户)：把不同的账户地址分开，可以更好地管理资金。这一层使得用户可以在同一个软件下生成和管理多个独立的账户，使得账本可以透明化，而不会全部混在一起。
Change (找零)：用于区别找零地址和接收地址。通常情况下，0代表外部地址，1代表找零地址
Address Index (地址序号)：标识生成的第n个地址。
使用BIP44的好处是，只要记住种子和BIP44的这个路径定义，就能在任何符合BIP44规范的钱包上生成和恢复你需要管理的币种和地址。
*/

/*
实现bip44协议的种子密钥树状生成器
*/
type KeyTreeGenerator struct {
	mnemonic   string                //助记词
	passphrase string                //密码
	masterKey  *bip32.Key            //
	lock       *sync.Mutex           //锁
	keys       map[string]*bip32.Key //已经生成的密钥
}

func NewKeyTreeGeneratorByMnemonic(mnemonic string) (*KeyTreeGenerator, utils.ERROR) {
	hashedSeed := bip39.NewSeed(mnemonic, "")
	masterKey, err := bip32.NewMasterKey(hashedSeed)
	if err == bip32.ErrInvalidPrivateKey {
		return nil, utils.NewErrorBus(config.ERROR_code_unusable_seed, "")
	}
	ktg := &KeyTreeGenerator{
		mnemonic:   mnemonic,
		passphrase: "",
		masterKey:  masterKey,
		lock:       &sync.Mutex{},
		keys:       make(map[string]*bip32.Key),
	}
	return ktg, utils.NewErrorSuccess()
}

func NewKeyTreeGeneratorBySeed(seeds []byte, passphrase string, mnemonicLang []string) (*KeyTreeGenerator, error) {
	bip39.SetWordList(mnemonicLang)
	//通过原生的seed推出助记词
	mnemonic, err := bip39.NewMnemonic(seeds)
	if err != nil {
		return nil, err
	}
	ktg := &KeyTreeGenerator{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		lock:       &sync.Mutex{},
		keys:       make(map[string]*bip32.Key),
	}
	return ktg, nil
}

// 找根私钥 返构造bip32.key
func (this *KeyTreeGenerator) getKey(path string) (*bip32.Key, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	key, ok := this.keys[path]
	return key, ok
}

// 构造对应路径的key
func (km *KeyTreeGenerator) setKey(path string, key *bip32.Key) {
	km.lock.Lock()
	defer km.lock.Unlock()
	km.keys[path] = key
}

// GetMnemonic 构造助记词返回
func (km *KeyTreeGenerator) GetMnemonic() string {
	return km.mnemonic
}

// GetPassphrase 构造助记词密码返回
func (km *KeyTreeGenerator) GetPassphrase() string {
	return km.passphrase
}

// GetSeed 获取助记词的seed
func (km *KeyTreeGenerator) GetSeed() []byte {
	return bip39.NewSeed(km.GetMnemonic(), km.GetPassphrase())
}

// GetMasterKey 找根key
func (km *KeyTreeGenerator) GetMasterKey() (*bip32.Key, error) {
	path := "m"
	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}
	key, err := bip32.NewMasterKey(km.GetSeed())
	if err != nil {
		return nil, err
	}
	km.setKey(path, key)
	return key, nil
}

// GetPurposeKey 根据协议获取key
func (km *KeyTreeGenerator) GetPurposeKey(purpose uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'`, purpose)
	key, ok := km.getKey(path)
	if ok {
		//utils.Log.Info().Interface("key", key).Send()
		return key, nil
	}
	parent, err := km.GetMasterKey()
	if err != nil {
		//utils.Log.Info().Interface("parent", parent).Send()
		return nil, err
	}
	key, err = parent.NewChildKey(purpose)
	if err != nil {
		//utils.Log.Info().Interface("key", key).Send()
		return nil, err
	}
	km.setKey(path, key)
	return key, nil
}

// GetCoinTypeKey 根据不同币种获取key
func (km *KeyTreeGenerator) GetCoinTypeKey(purpose, coinType uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'`, purpose, coinType)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetPurposeKey(purpose)
	if err != nil {
		return nil, err
	}
	//utils.Log.Info().Interface("parent", parent).Send()
	key, err = parent.NewChildKey(coinType)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

// GetAccountKey 根据account获取key
func (km *KeyTreeGenerator) GetAccountKey(purpose, coinType, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'`, purpose, coinType, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetCoinTypeKey(purpose, coinType)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(account)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

// GetChangeKey 获取指定路径的bip32.key
func (km *KeyTreeGenerator) GetChangeKey(purpose, coinType, account, change uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose, coinType, account, change)
	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}
	parent, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(change)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

// 兼容一下批量创建的时候
func (km *KeyTreeGenerator) GetIndexKey(purpose, coinType, account, change, index uint32) (string, *bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d/%d`, purpose, coinType, account, change, index)
	//fmt.Println(path)
	key, ok := km.getKey(path)
	if ok {
		return path, key, nil
	}
	parent, err := km.GetChangeKey(purpose, coinType, account, change)
	if err != nil {
		return "", nil, err
	}
	key, err = parent.NewChildKey(index)
	if err != nil {
		return "", nil, err
	}
	km.setKey(path, key)
	return path, key, nil
}
