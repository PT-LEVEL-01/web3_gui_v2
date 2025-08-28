package chain_plus

import (
	"math/big"
	"strconv"
	chainconfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	engineconfig "web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

/*
发布、创建一个token
*/
func CreateTxTokenPublish(owner *coin_address.AddressCoin, tokenName, tokenSymbol string, tokenSupply *big.Int,
	tokenAccuracy, gas uint64, pwd string, comment []byte) (*mining.TxTokenPublish, utils.ERROR) {

	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}

	//验证发行总量最少
	if tokenSupply.Cmp(chainconfig.Witness_token_supply_min) < 0 {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_publish_min, "")
	}

	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}

	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}

	ownerOld := crypto.AddressCoin(*owner)
	//srcAddrOld := crypto.AddressCoin(*srcAddr)

	//拥有人为空，则设置为有余额的地址
	if owner == nil || len(*owner) == 0 {
		for _, one := range keyst.GetCoinAddrAll() {
			addrOne := crypto.AddressCoin(one.Addr.Bytes())
			total, _ := chain.Balance.BuildPayVinNew(&addrOne, gas)
			if total >= gas {
				addrOne := coin_address.AddressCoin(one.Addr.Bytes())
				owner = &addrOne
				break
			}
		}
		if owner == nil || len(*owner) == 0 {
			//资金不够
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
		}
	} else {
		total, _ := chain.Balance.BuildPayVinNew(&ownerOld, gas)
		if total < gas {
			utils.Log.Warn().Str("balance not enough", "token").Send()
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_balance_not_enough, "")
		}
	}

	tokenVout := make([]mining.Vout, 0)
	voutOne := mining.Vout{
		Value:   tokenSupply.Uint64(),       //输出金额 = 实际金额 * 100000000
		Address: crypto.AddressCoin(*owner), //钱包地址
	}
	tokenVout = append(tokenVout, voutOne)

	//total, item := chain.Balance.BuildPayVinNew(&srcAddrOld, gas)
	//if total < gas {
	//	//资金不够
	//	return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
	//}
	//addr := coin_address.AddressCoin(*item.Addr)
	addrInfo := keyst.FindCoinAddr(owner)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	nonce := chain.GetBalance().FindNonce(&ownerOld)
	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	currentHeight := chain.GetCurrentBlock()
	var txin *mining.TxTokenPublish
	for i := uint64(0); i < 10000; i++ {
		base := NewTxBase(chainconfig.Wallet_tx_type_token_publish, &vins, &vouts, gas, currentHeight, i)
		base.Comment = comment
		txin = &mining.TxTokenPublish{
			TxBase:           *base,
			Token_name:       tokenName,              //名称
			Token_symbol:     tokenSymbol,            //单位
			Token_supply:     tokenSupply,            //发行总量
			Token_accuracy:   tokenAccuracy,          //精度
			Token_Vout_total: uint64(len(tokenVout)), //输出交易数量
			Token_Vout:       tokenVout,              //交易输出
		}

		//给输出签名，防篡改
		ERR = SignTxVin(txin, keyst, pwd)
		if ERR.CheckFail() {
			return nil, ERR
		}
		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	return txin, utils.NewErrorSuccess()
}

/*
token转账
*/
func CreateTxTokenSendToAddress(tokenId []byte, srcaddress, address coin_address.AddressCoin, amount, gas, frozen_height uint64,
	pwd string, comment []byte) (*mining.TxTokenPay, utils.ERROR) {

	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}

	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}

	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}

	addrOld := crypto.AddressCoin(address)

	//没有扣款地址，就从有余额的地址扣款
	if srcaddress == nil || len(srcaddress) == 0 {
		for _, one := range keyst.GetCoinAddrAll() {
			tokenTotal, _, _ := mining.GetTokenNotSpendAndLockedBalance(tokenId, one.Addr.Bytes())
			if tokenTotal >= amount {
				srcaddress = coin_address.AddressCoin(one.Addr.Bytes())
				break
			}
		}
		if srcaddress == nil || len(srcaddress) == 0 {
			utils.Log.Warn().Str("balance not enough", "token").Send()
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_balance_not_enough, "")
		}
	}
	srcAddrOld := crypto.AddressCoin(srcaddress)
	tokenTotal, _, _ := mining.GetTokenNotSpendAndLockedBalance(tokenId, srcAddrOld)
	if tokenTotal < amount {
		utils.Log.Warn().Str("balance not enough", "token").Send()
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_balance_not_enough, "")
	}

	addr := coin_address.AddressCoin(srcAddrOld)
	addrInfo := keyst.FindCoinAddr(&addr)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	tokenVins := make([]*mining.Vin, 0)
	tokenvin := mining.Vin{
		Puk: puk, //公钥
	}
	tokenVins = append(tokenVins, &tokenvin)

	//构建交易输出
	tokenVouts := make([]*mining.Vout, 0)
	//转账token给目标地址
	tokenVout := mining.Vout{
		Value:        amount,        //输出金额 = 实际金额 * 100000000
		Address:      addrOld,       //钱包地址
		FrozenHeight: frozen_height, //
	}
	tokenVouts = append(tokenVouts, &tokenVout)

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	total, item := chain.Balance.BuildPayVinNew(&srcAddrOld, gas)
	if total < gas {
		utils.Log.Warn().Str("balance not enough", "gas").Send()
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
	}

	vins := make([]*mining.Vin, 0)
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := mining.Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)

	currentHeight := chain.GetCurrentBlock()

	var tx *mining.TxTokenPay
	for i := uint64(0); i < 10000; i++ {
		base := NewTxBase(chainconfig.Wallet_tx_type_token_payment, &vins, &vouts, gas, currentHeight, i)
		base.Comment = comment
		tx = &mining.TxTokenPay{
			TxBase:           *base,
			Token_txid:       tokenId,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
		}

		//给输出签名，防篡改
		ERR = SignTxVin(tx, keyst, pwd)
		if ERR.CheckFail() {
			return nil, ERR
		}
		tx.BuildHash()
		if tx.CheckHashExist() {
			tx = nil
			continue
		} else {
			break
		}
	}
	chain.GetBalance().AddLockTx(tx)
	return tx, utils.NewErrorSuccess()
}

/*
token转账
*/
func CreateTxTokenSendToAddressMore(tokenId []byte, srcAddress coin_address.AddressCoin, address []coin_address.AddressCoin, amount,
	frozenHeight []uint64, gas uint64, pwd string, comment []byte) (*mining.TxTokenPay, utils.ERROR) {

	if len(address) != len(amount) {
		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "amount")
	}
	if len(address) != len(frozenHeight) {
		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "frozenHeight")
	}
	//if len(domain) > 0 {
	//	if len(address) != len(domain) {
	//		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "domain")
	//	}
	//} else {
	//	domain = make([]string, len(address))
	//}
	//if len(domainType) > 0 {
	//	if len(address) != len(domainType) {
	//		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "domainType")
	//	}
	//} else {
	//	domainType = make([]uint64, len(address))
	//}

	amountTotal := uint64(0)
	for _, one := range amount {
		amountTotal += one
	}

	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}

	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}

	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}

	//没有扣款地址，就从有余额的地址扣款
	if srcAddress == nil || len(srcAddress) == 0 {
		for _, one := range keyst.GetCoinAddrAll() {
			tokenTotal, _, _ := mining.GetTokenNotSpendAndLockedBalance(tokenId, one.Addr.Bytes())
			if tokenTotal >= amountTotal {
				srcAddress = coin_address.AddressCoin(one.Addr.Bytes())
				break
			}
		}
		if srcAddress == nil || len(srcAddress) == 0 {
			utils.Log.Warn().Str("balance not enough", "token").Send()
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_balance_not_enough, "")
		}
	}

	srcAddrOld := crypto.AddressCoin(srcAddress)
	tokenTotal, _, _ := mining.GetTokenNotSpendAndLockedBalance(tokenId, srcAddrOld)
	if tokenTotal < amountTotal {
		utils.Log.Warn().Str("balance not enough", "token").Send()
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_token_balance_not_enough, "")
	}

	addr := coin_address.AddressCoin(srcAddrOld)
	addrInfo := keyst.FindCoinAddr(&addr)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	tokenVins := make([]*mining.Vin, 0)
	tokenvin := mining.Vin{
		Puk: puk, //公钥
	}
	tokenVins = append(tokenVins, &tokenvin)

	//构建交易输出
	tokenVouts := make([]*mining.Vout, 0)
	for i, one := range address {
		//转账token给目标地址
		tokenVout := mining.Vout{
			Value:        amount[i],               //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressCoin(one), //钱包地址
			FrozenHeight: frozenHeight[i],         //
		}
		tokenVouts = append(tokenVouts, &tokenVout)
	}

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	total, item := chain.Balance.BuildPayVinNew(&srcAddrOld, gas)
	if total < gas {
		utils.Log.Warn().Str("balance not enough", "gas").Send()
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
	}

	vins := make([]*mining.Vin, 0)
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := mining.Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)

	currentHeight := chain.GetCurrentBlock()

	var tx *mining.TxTokenPay
	for i := uint64(0); i < 10000; i++ {
		base := NewTxBase(chainconfig.Wallet_tx_type_token_payment, &vins, &vouts, gas, currentHeight, i)
		base.Comment = comment
		tx = &mining.TxTokenPay{
			TxBase:           *base,
			Token_txid:       tokenId,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
		}

		//给输出签名，防篡改
		ERR = SignTxVin(tx, keyst, pwd)
		if ERR.CheckFail() {
			return nil, ERR
		}
		tx.BuildHash()
		if tx.CheckHashExist() {
			tx = nil
			continue
		} else {
			break
		}
	}
	chain.GetBalance().AddLockTx(tx)
	return tx, utils.NewErrorSuccess()
}
