package rpc

import (
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
)

func OfflineTxSendToAddress(params *map[string]interface{}, key_store_path, srcaddress, address string, amount, nonce, currentHeight,
	frozen_height, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	tx, hash, ERR := chain_plus.OfflineTxCreateSendAddress(key_store_path, srcaddress, address, pwd, comment, amount, gas, frozen_height,
		nonce, currentHeight, "", 0)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["hash"] = hash
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = tx
	return *pr
}

func OfflineTxCreateContract(params *map[string]interface{}, key_store_path, srcaddress, address string, amount, nonce, currentHeight,
	frozen_height, gas, gas_price uint64, abi, source, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	tx, hash, addressContract, ERR := chain_plus.OfflineTxCreateContract(key_store_path, srcaddress, address, pwd, comment,
		amount, gas, frozen_height, gas_price, nonce, currentHeight, "", 0, abi, source)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["hash"] = hash
	pr.Data["contract_address"] = addressContract
	pr.Data["tx"] = tx
	return *pr
}

func OfflineTxCommunityDepositIn(params *map[string]interface{}, walletPath, witness, address string, rate, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	witnessItr := (*params)["witness1"]
	witnessAddr := witnessItr.(coin_address.AddressCoin)
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteIn, ERR := chain_plus.OfflineTxCommunityDepositIn(walletPath, witnessAddr, addr, uint16(rate), amount, gas,
		nonce, currentHeight, "", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = voteIn.GetHash()
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteIn
	return *pr
}

func OfflineTxCommunityDepositOut(params *map[string]interface{}, walletPath, address string, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteOut, ERR := chain_plus.OfflineTxCommunityDepositOut(walletPath, addr, amount, gas, nonce, currentHeight,
		"", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = hash
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteOut
	return *pr
}

func OfflineTxLightDepositIn(params *map[string]interface{}, walletPath, address string, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteIn, ERR := chain_plus.OfflineTxLightDepositIn(walletPath, addr, 0, amount, gas,
		nonce, currentHeight, "", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = voteIn.GetHash()
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteIn
	return *pr
}

func OfflineTxLightDepositOut(params *map[string]interface{}, walletPath, address string, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteOut, ERR := chain_plus.OfflineTxLightDepositOut(walletPath, addr, amount, gas, nonce, currentHeight,
		"", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = hash
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteOut
	return *pr
}

func OfflineTxLightVoteIn(params *map[string]interface{}, walletPath, community, address string, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	communityAddressItr := (*params)["community1"]
	communityAddr := communityAddressItr.(coin_address.AddressCoin)
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteIn, ERR := chain_plus.OfflineTxLightVoteIn(walletPath, communityAddr, addr, 0, amount, gas,
		nonce, currentHeight, "", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = voteIn.GetHash()
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteIn
	return *pr
}

func OfflineTxLightVoteOut(params *map[string]interface{}, walletPath, address string, amount,
	nonce, currentHeight, gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	voteOut, ERR := chain_plus.OfflineTxLightVoteOut(walletPath, addr, amount, gas, nonce, currentHeight,
		"", 0, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//pr.Data["hash"] = hash
	//pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = voteOut
	return *pr
}
