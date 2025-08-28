package rpc

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"net/http"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

func LaunchSwap(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	tokenIdOutItr, ok := rj.Get("tokenidout")
	if !ok {
		res, err = model.Errcode(model.NoField, "tokenidout")
		return
	}
	tokenIdOutStr := tokenIdOutItr.(string)
	tokenIdOut, err := hex.DecodeString(tokenIdOutStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "tokenidout")
		return
	}

	tokenIdInItr, ok := rj.Get("tokenidin")
	if !ok {
		res, err = model.Errcode(model.NoField, "tokenidin")
		return
	}
	tokenIdInStr := tokenIdInItr.(string)
	tokenIdIn, err := hex.DecodeString(tokenIdInStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "tokenidin")
		return
	}

	if bytes.Equal(tokenIdOut, tokenIdIn) {
		res, err = model.Errcode(model.Nomarl, "tokenidout and tokenidin is equal")
		return
	}

	amountOutItr, ok := rj.Get("amountout")
	if !ok {
		res, err = model.Errcode(5002, "amountout")
		return
	}
	amountOut, ok := new(big.Int).SetString(amountOutItr.(string), 10)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "amountout")
		return
	}

	amountInItr, ok := rj.Get("amountin")
	if !ok {
		res, err = model.Errcode(5002, "amountin")
		return
	}
	amountIn, ok := new(big.Int).SetString(amountInItr.(string), 10)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "amountin")
		return
	}

	lockhightPromoter := toUint64(0)
	lockhightPromoterItr, ok := rj.Get("lockhightpromoter")
	if ok {
		lockhightPromoter = toUint64(lockhightPromoterItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	txpay, e := mining.SwapTxPromoter(&src, tokenIdOut, tokenIdIn, amountOut, amountIn, lockhightPromoter, pwd)
	if e != nil {
		if e.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if e.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(NotEnough)
			return
		}
		if e.Error() == config.ERROR_name_exist.Error() {
			res, err = model.Errcode(model.Exist)
			return
		}
		return model.Errcode(model.Nomarl, e.Error())
	}

	return model.Tojson(txpay.GetVOJSON())
}

func SwapPool(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	chain := mining.GetLongChain().GetTransactionSwapManager()
	var state int8
	stateItr, ok := rj.Get("state")
	if ok {
		state = int8(stateItr.(float64))
	}
	pool := chain.GetSwapPool(mining.Swap_Step(state))

	txs := make([]mining.SwapTransaction_V0, 0, len(pool))
	for _, v := range pool {
		tx := v.GetVOJSON().(mining.SwapTransaction_V0)
		txs = append(txs, tx)
	}
	return model.Tojson(txs)
	//result := make([]*mining.TxSwap_VO, 0, len(pool))
	//for _, v := range pool {
	//	tx := v.GetVOJSON().(mining.TxSwap_VO)
	//	result = append(result, &tx)
	//}
	//return model.Tojson(result)
}

func AccomplishSwap(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	txItr, ok := rj.Get("txid")
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txidStr := txItr.(string)
	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	tx := mining.GetLongChain().GetTransactionSwapManager().GetSwapTx(txid, mining.Swap_Step_Promoter)
	if tx == nil {
		res, err = model.Errcode(model.NotExist, "txid")
		return
	}

	//先验证签名
	if err := tx.CheckSignStep(mining.Swap_Step_Promoter); err != nil {
		return model.Errcode(model.Nomarl, "签名验证失败")
	}

	amountDealItr, ok := rj.Get("amountdeal")
	if !ok {
		res, err = model.Errcode(5002, "amountdeal")
		return
	}
	amountDeal, ok := new(big.Int).SetString(amountDealItr.(string), 10)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "amountdeal")
		return
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	if gas < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas)
	if total < gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	txpay, e := mining.SwapTxReceiver(&src, tx, amountDeal.Uint64(), gas, pwd)
	if e != nil {
		if e.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if e.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(NotEnough)
			return
		}
		if e.Error() == config.ERROR_name_exist.Error() {
			res, err = model.Errcode(model.Exist)
			return
		}
		return model.Errcode(model.Nomarl, e.Error())
	}

	return model.Tojson(txpay.GetVOJSON())
}

/*
发布一个token
*/
func TokenPublish(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok = rj.Get("address") //押金冻结的地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(ContentIncorrectFormat, "address")
				return
			}
		}
	}

	amount := toUint64(0)

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(GasTooLittle, "gas")
		return
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	name := ""
	nameItr, ok := rj.Get("name") //token的名称
	if ok {
		name = nameItr.(string)
	}
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(model.NoField, "name")
		return
	}

	//symbol token单位
	symbol := ""
	symbolItr, ok := rj.Get("symbol")
	if ok {
		symbol = symbolItr.(string)
	}
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if strings.Contains(symbol, ".") || strings.Contains(symbol, " ") {
		res, err = model.Errcode(model.NoField, "symbol")
		return
	}

	//token发行总量
	supplyItr, ok := rj.Get("supply") //token发行总量
	if !ok {
		res, err = model.Errcode(model.NoField, "supply")
		return
	}
	supply, ok := new(big.Int).SetString(supplyItr.(string), 10)
	if !ok || supply.Cmp(config.Witness_token_supply_min) < 0 {
		res, err = model.Errcode(BalanceTooLittle, config.ERROR_token_min_fail.Error())
		return
	}

	if !supply.IsUint64() {
		res, err = model.Errcode(BalanceTooBig, "limit supply")
		return
	}

	//精度最大18位
	accuracyItr, ok := rj.Get("accuracy")
	if !ok {
		res, err = model.Errcode(model.NoField, "accuracy")
		return
	}
	accuracy := toUint64(accuracyItr.(float64))
	if !(accuracy >= 0 && accuracy <= 18) {
		res, err = model.Errcode(ParamError, "limit accuracy 0..18")
		return
	}

	// supply := toUint64(supplyItr.(float64))
	// if supply < config.Witness_token_supply_min {
	// 	res, err = model.Errcode(BalanceTooLittle, config.ERROR_token_min_fail.Error())
	// 	return
	// }

	var owner crypto.AddressCoin
	ownerItr, ok := rj.Get("owner") //押金冻结的地址
	if ok {
		ownerStr := ownerItr.(string)
		if ownerStr != "" {
			ownerMul := crypto.AddressFromB58String(ownerStr)
			owner = ownerMul
		}

		if ownerStr != "" {
			dst := crypto.AddressFromB58String(ownerStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(ContentIncorrectFormat, "owner")
				return
			}
		}
	}

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @name    string    Token名称全称
	// @symbol    string    Token单位，符号
	// @supply    uint64    发行总量
	// @owner    crypto.AddressCoin    所有者
	txItr, e := mining.PublishToken(&src, addr, amount, gas, frozenHeight, pwd, comment, name, symbol, supply, accuracy, owner)
	if e == nil {
		//// res, err = model.Tojson("success")
		//result, e := utils.ChangeMap(txItr)
		//if e != nil {
		//	res, err = model.Errcode(model.Nomarl, err.Error())
		//	return
		//}
		//result["hash"] = hex.EncodeToString(*txItr.GetHash())
		res, err = model.Tojson(txItr.GetVOJSON())
		return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
使用token支付
*/
func TokenPay(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok = rj.Get("address") //押金冻结的地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(ContentIncorrectFormat, "address")
				return
			}
		}
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(GasTooLittle, "gas")
		return
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	// var owner crypto.AddressCoin
	txidItr, ok := rj.Get("txid") //发布token的交易id
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txid := txidItr.(string)

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @txid    string    发布token的交易id
	txItr, e := mining.TokenPay(&src, addr, amount, gas, frozenHeight, pwd, comment, txid)
	if e == nil {
		res, err = model.Tojson(txItr.GetVOJSON())
		return

		// res, err = model.Tojson("success")
		// return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
使用token多人转账
*/
func TokenPayMore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	srcAddrStr := ""
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcAddrStr = addrItr.(string)
		if srcAddrStr != "" {
			src = crypto.AddressFromB58String(srcAddrStr)
			//判断地址前缀是否正确
			// if !crypto.ValidAddr(config.AddrPre, src) {
			// 	res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			//判断地址是否包含在keystone里面
			// if !keystore.FindAddress(src) {
			// 	res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrItr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	addrs := make([]PayNumber, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&addrs)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	//给多人转账，可是没有地址
	if len(addrs) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	amount := toUint64(0)

	addr := make([]mining.PayNumber, 0)
	for _, one := range addrs {
		dst := crypto.AddressFromB58String(one.Address)
		//验证地址前缀
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(ContentIncorrectFormat, "addresses")
			return
		}
		pnOne := mining.PayNumber{
			Address:      dst,              //转账地址
			Amount:       one.Amount,       //转账金额
			FrozenHeight: one.FrozenHeight, //
		}
		addr = append(addr, pnOne)
		amount += one.Amount
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(GasTooLittle, "gas")
		return
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	// var owner crypto.AddressCoin
	txidItr, ok := rj.Get("txid") //发布token的交易id
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txid, e := hex.DecodeString(txidItr.(string))
	if e != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @txid    string    发布token的交易id
	txItr, e := mining.TokenPayMore(nil, src, addr, gas, pwd, comment, txid)
	if e == nil {
		result, e := utils.ChangeMap(txItr)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())

		res, err = model.Tojson(result)
		return

		// res, err = model.Tojson("success")
		// return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
token信息
*/
func TokenInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	tokenId := []byte{}
	tokenidItr, ok := rj.Get("token_id")
	if !ok {
		res, err = model.Errcode(model.NoField, "token_id")
		return
	}

	if !rj.VerifyType("token_id", "string") {
		res, err = model.Errcode(model.TypeWrong, "token_id")
		return
	}
	tokenidStr := tokenidItr.(string)
	tokenId, err = hex.DecodeString(tokenidStr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "decode error")
		return
	}

	tokeninfo, err := mining.FindTokenInfo(tokenId)
	if err != nil {
		res, err = model.Errcode(model.NotExist, "not exist")
		return
	}
	res, err = model.Tojson(mining.ToTokenInfoV0(tokeninfo))
	return
}

/*
token信息
*/
func TokenList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	infos := []*mining.TokenInfoV0{}
	list, err := mining.ListTokenInfos()
	if err != nil {
		res, err = model.Tojson(infos)
		return
	}

	for _, l := range list {
		infos = append(infos, mining.ToTokenInfoV0(l))
	}

	res, err = model.Tojson(infos)
	return
}

/*
发布一个token订单
*/
func NewTokenOrder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	srcAddress, res, err := getParamString(rj, "srcaddress")
	if err != nil {
		return res, err
	}
	address, res, err := getParamString(rj, "address")
	if err != nil {
		return res, err
	}
	tokenAIDStr, res, err := getParamString(rj, "tokenaid")
	if err != nil {
		return res, err
	}
	tokenAID, err := hex.DecodeString(tokenAIDStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}
	tokenAAmount, res, err := getParamUint64(rj, "tokenaamount")
	if err != nil {
		return res, err
	}
	tokenBIDStr, res, err := getParamString(rj, "tokenbid")
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}
	tokenBID, err := hex.DecodeString(tokenBIDStr)
	if err != nil {
		return nil, err
	}
	tokenBAmount, res, err := getParamUint64(rj, "tokenbamount")
	if err != nil {
		return res, err
	}
	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}
	if gas < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	buy, res, err := getParamBool(rj, "buy")
	if err != nil {
		return res, err
	}

	if !(tokenAAmount > 0 && tokenBAmount > 0) {
		res, err = model.Errcode(model.Nomarl, "tokena/tokenb amount error")
		return
	}

	if bytes.Equal(tokenAID, tokenBID) {
		res, err = model.Errcode(model.Nomarl, "tokena/tokenb error")
		return
	}

	srcAddr := crypto.AddressFromB58String(srcAddress)
	addr := crypto.AddressFromB58String(address)
	txItr, err := mining.CreateTokenOrder(&srcAddr, &addr, tokenAID, tokenAAmount, tokenBID, tokenBAmount, buy, gas, pwd)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
发布一个token订单
*/
func NewTokenOrderV2(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	srcAddress, res, err := getParamString(rj, "srcaddress")
	if err != nil {
		return res, err
	}
	address, res, err := getParamString(rj, "address")
	if err != nil {
		return res, err
	}
	tokenAIDStr, res, err := getParamString(rj, "tokenaid")
	if err != nil {
		return res, err
	}
	tokenAID, err := hex.DecodeString(tokenAIDStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}
	tokenAAmount, res, err := getParamUint64(rj, "tokenaamount")
	if err != nil {
		return res, err
	}
	tokenBIDStr, res, err := getParamString(rj, "tokenbid")
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}
	tokenBID, err := hex.DecodeString(tokenBIDStr)
	if err != nil {
		return nil, err
	}
	price, res, err := getParamFloat64(rj, "price")
	if err != nil {
		return res, err
	}
	if price <= 0 {
		res, err = model.Errcode(model.Nomarl, "price error")
		return
	}
	tokenBAmount := uint64(float64(tokenAAmount) * price)

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}
	if gas < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}

	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	buy, res, err := getParamBool(rj, "buy")
	if err != nil {
		return res, err
	}

	if !(tokenAAmount > 0 && tokenBAmount > 0) {
		res, err = model.Errcode(model.Nomarl, "tokena/tokenb amount error")
		return
	}

	if bytes.Equal(tokenAID, tokenBID) {
		res, err = model.Errcode(model.Nomarl, "tokena/tokenb error")
		return
	}

	srcAddr := crypto.AddressFromB58String(srcAddress)
	addr := crypto.AddressFromB58String(address)
	txItr, err := mining.CreateTokenOrder(&srcAddr, &addr, tokenAID, tokenAAmount, tokenBID, tokenBAmount, buy, gas, pwd)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
取消token订单
*/
func CancelTokenOrder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	srcAddress, res, err := getParamString(rj, "srcaddress")
	if err != nil {
		return res, err
	}
	address, res, err := getParamString(rj, "address")
	if err != nil {
		return res, err
	}

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}
	if gas < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}

	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}

	orderIDStrs, errcode := getArrayStrParams(rj, "orderids")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "orderids")
		return
	}
	if len(orderIDStrs) == 0 {
		res, err = model.Errcode(model.Nomarl, "orderids empty")
		return
	}

	orderIDs := make([][]byte, 0, len(orderIDStrs))
	for _, idStr := range orderIDStrs {
		id, err2 := hex.DecodeString(idStr)
		if err2 != nil || len(id) == 0 {
			res, err = model.Errcode(model.Nomarl, "error orderid: "+idStr)
			return
		}
		orderIDs = append(orderIDs, id)
	}

	srcAddr := crypto.AddressFromB58String(srcAddress)
	addr := crypto.AddressFromB58String(address)
	txItr, err := mining.CancelTokenOrder(&srcAddr, &addr, orderIDs, gas, pwd)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
查询token订单
*/
func TokenOrderInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	orderIdStr, res, err := getParamString(rj, "orderid")
	if err != nil {
		return res, err
	}

	orderId, err := hex.DecodeString(orderIdStr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	info, err := mining.GetTokenOrder(orderId)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	data := ToOrderInfoV0(info)

	res, err = model.Tojson(data)
	return
}

/*
查询token订单池
*/
func TokenOrderPool(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	poolitems := mining.GetTokenOrderPool()
	data := map[string][]*OrderInfoV0{}
	for key, items := range poolitems {
		infos := []*OrderInfoV0{}
		for _, item := range items {
			infos = append(infos, ToOrderInfoV0(item.(*go_protos.OrderInfo)))
		}
		data[hex.EncodeToString([]byte(key))] = infos
	}

	res, err = model.Tojson(data)
	return
}

type OrderInfoV0 struct {
	OrderId             string   `json:"order_id"`
	TokenAID            string   `json:"tokena_id"`
	TokenAAmount        uint64   `json:"tokena_amount"`
	TokenBID            string   `json:"tokenb_id"`
	TokenBAmount        uint64   `json:"tokenb_amount"`
	Address             string   `json:"address"`
	Buy                 bool     `json:"buy"`
	State               int32    `json:"state"`
	Status              string   `json:"status"`
	Price               float64  `json:"price"`
	TxIds               []string `json:"tx_ids"`
	PendingTokenAAmount uint64   `json:"pending_tokena_amount"`
	//PendingTokenBAmount uint64   `json:"pending_tokenB_amount"`
}

func ToOrderInfoV0(in *go_protos.OrderInfo) *OrderInfoV0 {
	addr := crypto.AddressCoin(in.Address)
	txIds := []string{}
	for _, txId := range in.TxIds {
		txIds = append(txIds, hex.EncodeToString(txId))
	}

	status := "未知"
	switch in.State {
	case mining.OrderStateNormal:
		status = "Normal"
	case mining.OrderStateDone:
		status = "Done"
	case mining.OrderStateCancel:
		status = "Cancel"
	}

	return &OrderInfoV0{
		OrderId:             hex.EncodeToString(in.OrderId),
		TokenAID:            hex.EncodeToString(in.TokenAID),
		TokenAAmount:        new(big.Int).SetBytes(in.TokenAAmount).Uint64(),
		TokenBID:            hex.EncodeToString(in.TokenBID),
		TokenBAmount:        new(big.Int).SetBytes(in.TokenBAmount).Uint64(),
		Address:             addr.B58String(),
		Buy:                 in.Buy,
		State:               in.State,
		Status:              status,
		Price:               in.Price,
		TxIds:               txIds,
		PendingTokenAAmount: new(big.Int).SetBytes(in.PendingTokenAAmount).Uint64(),
		//PendingTokenBAmount: new(big.Int).SetBytes(in.PendingTokenBAmount).Uint64(),
	}
}

func getParamString(rj *model.RpcJson, name string) (string, []byte, error) {
	itemItr, ok := rj.Get(name)
	if !ok {
		res, err := model.Errcode(model.NoField, name)
		return "", res, err
	}

	if !rj.VerifyType(name, "string") {
		res, err := model.Errcode(model.TypeWrong, name)
		return "", res, err
	}

	return itemItr.(string), nil, nil
}

func getParamUint64(rj *model.RpcJson, name string) (uint64, []byte, error) {
	itemItr, ok := rj.Get(name)
	if !ok {
		res, err := model.Errcode(model.NoField, name)
		return 0, res, err
	}

	if !rj.VerifyType(name, "float64") {
		res, err := model.Errcode(model.TypeWrong, name)
		return 0, res, err
	}

	if itemItr.(float64) >= 0 {
		return toUint64(itemItr.(float64)), nil, nil
	}

	res, err := model.Errcode(model.Nomarl, name)
	return 0, res, err
}

func getParamFloat64(rj *model.RpcJson, name string) (float64, []byte, error) {
	itemItr, ok := rj.Get(name)
	if !ok {
		res, err := model.Errcode(model.NoField, name)
		return 0, res, err
	}

	if !rj.VerifyType(name, "float64") {
		res, err := model.Errcode(model.TypeWrong, name)
		return 0, res, err
	}

	return itemItr.(float64), nil, nil
}

func getParamBool(rj *model.RpcJson, name string) (bool, []byte, error) {
	itemItr, ok := rj.Get(name)
	if !ok {
		res, err := model.Errcode(model.NoField, name)
		return false, res, err
	}

	if !rj.VerifyType(name, "bool") {
		res, err := model.Errcode(model.TypeWrong, name)
		return false, res, err
	}

	return itemItr.(bool), nil, nil
}
