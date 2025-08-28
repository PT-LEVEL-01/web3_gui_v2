package boot

import (
	"reflect"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

const (
	RPC_name_createNewAddress = "createNewAddress"
)

func RegisterRPC(eg *engine.Engine) utils.ERROR {
	ERR := eg.RegisterRPC(1, RPC_name_createNewAddress, RPC_CreateNewAddress, "创建一个收款地址",
		engine.NewParamValid_Mast_CustomFun_Panic("coinType", reflect.Uint64, ValidCoinType, "coinType1",
			"币种https://github.com/satoshilabs/slips/blob/master/slip-0044.md#registered-coin-types"))
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

func ValidCoinType(coinType interface{}) (any, utils.ERROR) {
	ct := coinType.(uint64)
	ctafi := coin_address.GetCoinTypeFactory(uint32(ct))
	if ctafi == nil {
		return nil, utils.NewErrorBus(keystore.ERROR_code_coin_type_addr_not_achieve, "")
	}
	return uint32(ct), utils.NewErrorSuccess()
}
