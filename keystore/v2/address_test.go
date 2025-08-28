package keystore

import (
	"crypto/rand"
	"golang.org/x/crypto/ed25519"
	"testing"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

func TestAddr(t *testing.T) {
	ParseAddr()
}

func TestValidAddr(t *testing.T) {
	preStr := "TEST"
	testCases := []struct {
		input    string
		expected bool
	}{
		//旧版短地址
		{"TEST6d7g4uwGdeKF226zijoQKxE3BEMw7Nzqh5", false},
		//新版长地址，但是效验码算法不对
		{"TESTYaMrNEDH8sSdqPx7EtYjzqefZDxk4f6hQCdQdPjFdWKGG8phw5", false},
		//新版长地址
		{"TEST2RmL3J3YyZX9TWirbYAavw6RvqH5QEzEwb9KfUmSvs27JkXnds5", true},
	}

	for _, tc := range testCases {
		addr := coin_address.AddressFromB58String(tc.input)
		actual := coin_address.ValidAddrCoin(preStr, addr)
		if actual != tc.expected {
			t.Errorf("IsDNS(%q) = %v, expected %v", tc.input, actual, tc.expected)
		}
	}
}

func ParseAddr() {
	//一个不完整的地址
	addStr := "SELFKN1RzSSzNQ9KdC3rH255SKp"
	addr := coin_address.AddressFromB58String(addStr)
	utils.Log.Info().Str("解析后的地址", addr.B58String()).Send()
	//完整地址
	addStr = "TESTYaMrNEDH8sSdqPx7EtYjzqefZDxk4f6hQCdQdPjFdWKGG8phw5"
	addr = coin_address.AddressFromB58String(addStr)
	utils.Log.Info().Str("解析后的地址", addr.B58String()).Send()
}

func TestValidAddrSimple(t *testing.T) {
	//一个不完整的地址
	addStr := "HIVEiLUNAsUUjRmdFwT8L1idzaH7y2uyqkEZemNo7bPweBBWDSL565"
	addr := coin_address.AddressFromB58String(addStr)
	utils.Log.Info().Str("解析后的地址", addr.B58String()).Send()
	ok := coin_address.ValidAddrCoin("HIVE", addr)
	utils.Log.Info().Bool("地址是否合法", ok).Send()
	var src = coin_address.AddressCoin(addr)
	utils.Log.Info().Str("解析后的地址", src.B58String()).Send()

}

func TestBuildAddrExample(t *testing.T) {
	puk, _, _ := ed25519.GenerateKey(rand.Reader)
	preStr := "TEST"
	addr, ERR := coin_address.BuildAddr(preStr, puk)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("创建的地址", addr.B58String()).Send()
	utils.Log.Info().Hex("创建的地址", addr).Send()
	utils.Log.Info().Hex("有效数据", addr.Data()).Send()

	ok := coin_address.ValidAddrCoin(preStr, addr)
	utils.Log.Info().Bool("是否验证通过", ok).Send()

	//
	addrCoin := coin_address.AddressFromB58String("123456")
	utils.Log.Info().Str("地址", addrCoin.GetPre()).Send()
	if !coin_address.ValidAddrCoin(addrCoin.GetPre(), addrCoin) {
		utils.Log.Info().Str("地址", "不合法").Send()
	}
}

func TestBuildAddrByData(t *testing.T) {
	puk, _, _ := ed25519.GenerateKey(rand.Reader)
	preStr := "TEST"
	addr, ERR := coin_address.BuildAddr(preStr, puk)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("创建的地址", addr.B58String()).Send()

	addr, ERR = coin_address.BuildAddrByData(preStr, addr.Data())
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("重新构建的地址", addr.B58String()).Send()
}
