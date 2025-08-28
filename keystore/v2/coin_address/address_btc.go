package coin_address

import (
	btcecv2 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/tyler-smith/go-bip32"
	"web3_gui/keystore/v2/base58"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

func init() {
	RegisterCoinTypeAddress(config.BTC, config.COINTYPE_BTC, &AddressBTCBuilder{})
}

type AddressBTCBuilder struct {
}

/*
创建地址对象
*/
func (this *AddressBTCBuilder) NewAddress(data []byte) (addr AddressInterface) {
	btc := AddressBTC(data)
	return &btc
}

/*
@return    []byte    私钥
@return    []byte    公钥
@return    error     错误
*/
func (this *AddressBTCBuilder) BuildPukAndPrk(key *bip32.Key) (prk []byte, puk []byte, ERR utils.ERROR) {
	//utils.Log.Info().Str("助记词", mnemonic).Send()
	//var keyTree *keystore.KeyTreeGenerator
	//keyTree, ERR = keystore.NewKeyTreeGeneratorByMnemonic(mnemonic)
	//if ERR.CheckFail() {
	//	return
	//}
	//_, key, err := keyTree.GetIndexKey(purpose, coinType, account, change, index)
	//if err != nil {
	//	return nil, nil, utils.NewErrorSysSelf(err)
	//}
	btcecPrk, btcecPuk := btcecv2.PrivKeyFromBytes(key.Key)
	prk = btcecPrk.Serialize()
	puk = btcecPuk.SerializeCompressed()
	//utils.Log.Info().Hex("公钥", puk).Send()
	ERR = utils.NewErrorSuccess()
	return
}

func (this *AddressBTCBuilder) BuildAddress(addrPre string, puk []byte) (addr AddressInterface, ERR utils.ERROR) {
	//utils.Log.Info().Hex("公钥1111", puk).Send()
	//pubKeyBs := btcecPuk.SerializeCompressed()
	pkHash := btcutil.Hash160(puk)
	addrHash, err := btcutil.NewAddressPubKeyHash(pkHash, &chaincfg.MainNetParams)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	addrStr := addrHash.EncodeAddress()
	addrBs := base58.Decode(addrStr)
	a := AddressBTC(addrBs)
	addr = &a
	//utils.Log.Info().Str("生成的btc地址", addrStr).Send()
	ERR = utils.NewErrorSuccess()
	return
}

func (this *AddressBTCBuilder) GetAddressStr(data []byte) string {
	return string(base58.Encode(data))
}

type AddressBTC []byte

/*
数据部分
*/
func (this *AddressBTC) Bytes() []byte {
	return *this
}

func (this *AddressBTC) B58String() string {
	if len(*this) <= 0 {
		return ""
	}
	return string(base58.Encode(*this))
}

/*
有效数据部分
*/
func (this *AddressBTC) Data() []byte {
	if len(*this) == 0 {
		return nil
	}
	return *this
}

func (this *AddressBTC) String() string {
	return this.B58String()
}
