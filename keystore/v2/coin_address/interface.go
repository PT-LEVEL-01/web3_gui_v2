package coin_address

import (
	"github.com/tyler-smith/go-bip32"
	"sync"
	"web3_gui/utils"
)

type CoinTypeAddressFactoryInterface interface {
	NewAddress(data []byte) (addr AddressInterface)
	//BuildPukAndPrk(mnemonic string, purpose, coinType, account, change, index uint32) (prk []byte, puk []byte,
	//	ERR utils.ERROR)
	BuildPukAndPrk(key *bip32.Key) (prk []byte, puk []byte, ERR utils.ERROR)
	BuildAddress(addrPre string, puk []byte) (addr AddressInterface, ERR utils.ERROR)
	GetAddressStr(data []byte) string
}

type AddressInterface interface {
	Bytes() []byte
	B58String() string
	Data() []byte
	String() string
}

var cointypeMap = map[string]uint32{}

func GetCoinType(coinName string) uint32 {
	return cointypeMap[coinName]
}

var coinTypeAddressFactory = new(sync.Map) //map[uint32]CoinTypeAddressFactoryInterface{}

func RegisterCoinTypeAddress(coinName string, coinType uint32, coin CoinTypeAddressFactoryInterface) {
	cointypeMap[coinName] = coinType
	coinTypeAddressFactory.Store(coinType, coin) //[coinType] = coin
}

func GetCoinTypeFactory(coinType uint32) CoinTypeAddressFactoryInterface {
	itr, ok := coinTypeAddressFactory.Load(coinType)
	if !ok {
		return nil
	}
	ctfi := itr.(CoinTypeAddressFactoryInterface)
	return ctfi
}
