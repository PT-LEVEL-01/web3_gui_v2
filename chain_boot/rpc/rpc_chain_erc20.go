package rpc

import (
	"web3_gui/chain/db"
	"web3_gui/chain/mining"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func ERC20Info(params *map[string]interface{}, address string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//addrItr := (*params)["address1"]
	//addr := crypto.AddressCoin(addrItr.(coin_address.AddressCoin))

	info := db.GetErc20Info(address)
	if info.Address == "" {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	type erc20Info struct {
		Address     string `json:"address"`
		From        string `json:"from"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimals    uint8  `json:"decimals"`
		TotalSupply uint64 `json:"totalSupply"`
	}

	data := erc20Info{
		Address:     info.Address,
		From:        info.From,
		Name:        info.Name,
		Symbol:      info.Symbol,
		Decimals:    info.Decimals,
		TotalSupply: info.TotalSupply.Uint64(),
	}

	//res, err = model.Tojson(data)
	pr.Data["data"] = data
	return *pr
}
