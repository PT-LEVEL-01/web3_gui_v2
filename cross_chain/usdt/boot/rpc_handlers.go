package boot

import (
	"web3_gui/cross_chain/usdt/wallet"
	"web3_gui/libp2parea/v2/engine"
)

/*
创建新地址
*/
func RPC_CreateNewAddress(params *map[string]interface{}, coinType uint64) engine.PostResult {
	result := engine.NewPostResult()
	coinTypeItr := (*params)["coinType1"]
	coinType1 := coinTypeItr.(uint32)
	coinAddrStr, ERR := wallet.CreateNewAddress(coinType1)
	if ERR.CheckFail() {
		result.Code = ERR.Code
		result.Msg = ERR.Msg
		return *result
	}
	result.Data["info"] = coinAddrStr
	return *result
}
