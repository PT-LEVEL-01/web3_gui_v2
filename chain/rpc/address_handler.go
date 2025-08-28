package rpc

import (
	"math/big"
	"net/http"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/mining"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

func AddressBind(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	bindType := toUint64(0)
	bindTypeItr, ok := rj.Get("bind_type")
	if ok {
		bindType = toUint64(bindTypeItr.(float64))
	}

	bindItr, ok := rj.Get("bindaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "bindaddress")
		return
	}
	bindAddr := bindItr.(string)
	bind := crypto.AddressFromB58String(bindAddr)
	if !crypto.ValidAddr(config.AddrPre, bind) {
		res, err = model.Errcode(ContentIncorrectFormat, "bindaddress")
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

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}
	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}
	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(src.B58String(), domain, bind.B58String(), new(big.Int).SetUint64(domainType)) {
			return model.Errcode(model.Nomarl, "domain name resolution failed")
		}
	}

	// 获取comment
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok {
		comment = commentItr.(string)
	}

	txpay, e := mining.AddressBind(&src, &bind, bindType, gas, frozenHeight, pwd, comment, domain, domainType)
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

func AddressTransfer(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	addrItr, ok = rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	dst := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	payItr, ok := rj.Get("payaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "payaddress")
		return
	}
	payAddr := payItr.(string)
	pay := crypto.AddressFromB58String(payAddr)
	if !crypto.ValidAddr(config.AddrPre, pay) {
		res, err = model.Errcode(ContentIncorrectFormat, "payaddress")
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	//if gas < config.Wallet_tx_gas_min {
	//	return model.Errcode(GasTooLittle, "gas")
	//}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas)
	if total < gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	payTotal, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&pay, amount)
	if amount == 0 {
		amount = payTotal
	}
	if payTotal < amount {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}
	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}
	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(src.B58String(), domain, dst.B58String(), new(big.Int).SetUint64(domainType)) {
			return model.Errcode(model.Nomarl, "domain name resolution failed")
		}
	}

	// 获取comment
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok {
		comment = commentItr.(string)
	}

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	if e := mining.CheckTxPayFreeGasWithParams(config.Wallet_tx_type_address_transfer, src, amount, gas, comment); e != nil {
		res, err = model.Errcode(GasTooLittle, "gas too little")
		return
	}

	txpay, e := mining.AddressTransfer(&src, &dst, &pay, amount, gas, frozenHeight, pwd, comment, domain, domainType)
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

func AddressFrozen(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	frozenType := toUint64(0)
	frozenTypeItr, ok := rj.Get("frozen_type")
	if ok {
		frozenType = toUint64(frozenTypeItr.(float64))
	}

	frozenAddrItr, ok := rj.Get("frozenaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "frozenaddress")
		return
	}
	frozenAddr := crypto.AddressFromB58String(frozenAddrItr.(string))
	if !crypto.ValidAddr(config.AddrPre, frozenAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "frozenaddress")
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

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}
	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}
	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(src.B58String(), domain, frozenAddr.B58String(), new(big.Int).SetUint64(domainType)) {
			return model.Errcode(model.Nomarl, "domain name resolution failed")
		}
	}

	// 获取comment
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok {
		comment = commentItr.(string)
	}

	txpay, e := mining.AddressFrozen(&src, &frozenAddr, frozenType, gas, frozenHeight, pwd, comment, domain, domainType)
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
