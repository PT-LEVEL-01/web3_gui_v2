package coin_address

import (
	"crypto/ed25519"
	"testing"
	"web3_gui/utils"
)

func TestAddressCustom(t *testing.T) {
	key, err := utils.Rand32Byte()
	if err != nil {
		panic(err)
	}
	privk := ed25519.NewKeyFromSeed(key[:]) //私钥
	pub := privk.Public().(ed25519.PublicKey)

	builder := AddressCustomBuilder{}
	addr, ERR := builder.BuildAddress("TEST", pub)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Str("生成地址", addr.B58String()).Send()
}
