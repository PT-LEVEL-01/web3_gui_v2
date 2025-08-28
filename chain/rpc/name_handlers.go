package rpc

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"strings"

	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	namenet "web3_gui/chain/mining/name"
	"web3_gui/chain/mining/tx_name_in"
	"web3_gui/chain/mining/tx_name_out"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

/*
域名注册
*/
func NameInReg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src *crypto.AddressCoin
	srcItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := srcItr.(string)
		if srcaddr != "" {
			addrSrc := crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, addrSrc) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(addrSrc)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			src = &addrSrc
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
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

	amountItr, ok := rj.Get("amount") //转账金额
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < config.Mining_name_deposit_min {
		res, err = model.Errcode(model.Nomarl, config.ERROR_name_deposit.Error())
		return
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

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

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	//域名解析的节点地址参数
	ids := make([]nodeStore.AddressNet, 0)
	netIdsItr, ok := rj.Get("netids") //名称解析的网络id
	if ok {
		netIds := netIdsItr.([]interface{})
		for _, one := range netIds {
			netidOne := one.(string)
			if crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(netidOne)) {
				res, err = model.Errcode(ContentIncorrectFormat, "invalid netids")
				return
			}
			idOne := nodeStore.AddressFromB58String(netidOne)
			if netidOne != "" && idOne.B58String() == "" {
				res, err = model.Errcode(ContentIncorrectFormat, netidOne)
				return
			}

			ids = append(ids, idOne)
		}
	}

	//收款地址参数
	coins := make([]crypto.AddressCoin, 0)
	addrcoinsItr, ok := rj.Get("addrcoins") //名称解析的收款地址
	if ok {
		addrcoins := addrcoinsItr.([]interface{})
		for _, one := range addrcoins {
			addrcoinOne := one.(string)
			idOne := crypto.AddressFromB58String(addrcoinOne)
			if !crypto.ValidAddr(config.AddrPre, idOne) {
				res, err = model.Errcode(ContentIncorrectFormat, "invalid addrcoins")
				return
			}
			coins = append(coins, idOne)
		}
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	//判断域名是否已经注册
	nameinfo := namenet.FindNameToNet(name)
	if nameinfo != nil {
		res, err = model.Errcode(model.Nomarl, "name already register")
		return
	}

	txpay, err := tx_name_in.NameIn(src, addr, amount, gas, frozenHeight, pwd, comment, mining.NameInActionReg, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		//result, e := utils.ChangeMap(txpay)
		//if e != nil {
		//	res, err = model.Errcode(model.Nomarl, e.Error())
		//	return
		//}
		//result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(txpay.GetVOJSON())

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
域名转让
*/
func NameInTransfer(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src *crypto.AddressCoin
	srcItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := srcItr.(string)
		if srcaddr != "" {
			addrSrc := crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, addrSrc) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(addrSrc)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			src = &addrSrc
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
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
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

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

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	nameinfo := namenet.FindName(name)
	if nameinfo == nil {
		res, err = model.Errcode(model.Nomarl, "not found name")
		return
	}

	//域名解析的节点地址参数
	ids := nameinfo.NetIds
	//收款地址参数
	coins := nameinfo.AddrCoins

	amount := nameinfo.Deposit

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	//查找域名是否属于自己
	if val := namenet.FindName(name); val == nil {
		res, err = model.Errcode(model.Nomarl, "not owner")
		return
	}

	txpay, err := tx_name_in.NameIn(src, addr, amount, gas, frozenHeight, pwd, comment, mining.NameInActionTransfer, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		//result, e := utils.ChangeMap(txpay)
		//if e != nil {
		//	res, err = model.Errcode(model.Nomarl, e.Error())
		//	return
		//}
		//result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(txpay.GetVOJSON())

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
域名续期
*/
func NameInRenew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src *crypto.AddressCoin
	srcItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := srcItr.(string)
		if srcaddr != "" {
			addrSrc := crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, addrSrc) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(addrSrc)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			src = &addrSrc
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
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
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

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

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	nameinfo := namenet.FindNameToNet(name)
	if nameinfo == nil {
		res, err = model.Errcode(model.Nomarl, "not found name")
		return
	}

	//域名解析的节点地址参数
	ids := nameinfo.NetIds
	//收款地址参数
	coins := nameinfo.AddrCoins

	amount := nameinfo.Deposit

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	//查找域名是否属于自己
	if val := namenet.FindName(name); !(val != nil && bytes.Equal(val.Owner, *addr)) {
		res, err = model.Errcode(model.Nomarl, "not owner")
		return
	}

	txpay, err := tx_name_in.NameIn(src, addr, amount, gas, frozenHeight, pwd, comment, mining.NameInActionRenew, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		//result, e := utils.ChangeMap(txpay)
		//if e != nil {
		//	res, err = model.Errcode(model.Nomarl, e.Error())
		//	return
		//}
		//result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(txpay.GetVOJSON())

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
域名修改
*/
func NameInUpdate(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src *crypto.AddressCoin
	srcItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := srcItr.(string)
		if srcaddr != "" {
			addrSrc := crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, addrSrc) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(addrSrc)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			src = &addrSrc
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
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
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

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

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	//查找域名是否属于自己
	if val := namenet.FindName(name); !(val != nil && bytes.Equal(val.Owner, *addr)) {
		res, err = model.Errcode(model.Nomarl, "not owner")
		return
	}

	//域名解析的节点地址参数
	ids := make([]nodeStore.AddressNet, 0)
	netIdsItr, ok := rj.Get("netids") //名称解析的网络id
	if ok {
		netIds := netIdsItr.([]interface{})
		for _, one := range netIds {
			netidOne := one.(string)
			if crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(netidOne)) {
				res, err = model.Errcode(ContentIncorrectFormat, "invalid netids")
				return
			}
			idOne := nodeStore.AddressFromB58String(netidOne)
			if netidOne != "" && idOne.B58String() == "" {
				res, err = model.Errcode(ContentIncorrectFormat, netidOne)
				return
			}
			ids = append(ids, idOne)
		}
	}

	//收款地址参数
	coins := make([]crypto.AddressCoin, 0)
	addrcoinsItr, ok := rj.Get("addrcoins") //名称解析的收款地址
	if ok {
		addrcoins := addrcoinsItr.([]interface{})
		for _, one := range addrcoins {
			addrcoinOne := one.(string)
			idOne := crypto.AddressFromB58String(addrcoinOne)
			if !crypto.ValidAddr(config.AddrPre, idOne) {
				res, err = model.Errcode(ContentIncorrectFormat, "invalid addrcoins")
				return
			}
			coins = append(coins, idOne)
		}
	}

	//判断域名是否已经注册
	nameinfo := namenet.FindNameToNet(name)
	if nameinfo == nil {
		res, err = model.Errcode(model.Nomarl, "name not register")
		return
	}

	amount := nameinfo.Deposit

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_in.NameIn(src, addr, amount, gas, frozenHeight, pwd, comment, mining.NameInActionUpdate, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		//result, e := utils.ChangeMap(txpay)
		//if e != nil {
		//	res, err = model.Errcode(model.Nomarl, e.Error())
		//	return
		//}
		//result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(txpay.GetVOJSON())

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
域名注销
*/
func NameOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// isReg := false //转账类型，true=注册名称；false=注销名称；

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金退还地址
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
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

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

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_out.NameOut(nil, addr, 0, gas, frozenHeight, pwd, comment, name)
	if err == nil {
		// res, err = model.Tojson("success")

		result, e := utils.ChangeMap(txpay)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_not_exist.Error() {
		res, err = model.Errcode(model.NotExist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())
	return

}

/*
获取自己注册的域名列表
*/
func GetNames(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	nameinfoVOs := make([]NameinfoVO, 0)

	names := namenet.GetNameList()
	for _, one := range names {
		nets := make([]string, 0)
		for _, two := range one.NetIds {
			nets = append(nets, two.B58String())
		}
		addrs := make([]string, 0)
		for _, two := range one.AddrCoins {
			addrs = append(addrs, two.B58String())
		}
		voOne := NameinfoVO{
			Name:           one.Name,              //域名
			Owner:          one.Owner.B58String(), //
			NetIds:         nets,                  //节点地址
			AddrCoins:      addrs,                 //钱包收款地址
			Height:         one.Height,            //注册区块高度，通过现有高度计算出有效时间
			NameOfValidity: one.NameOfValidity,    //有效块数量
			Deposit:        one.Deposit,
			IsMultName:     one.IsMultName,
		}
		nameinfoVOs = append(nameinfoVOs, voOne)
	}

	res, err = model.Tojson(nameinfoVOs)
	return
}

type NameinfoVO struct {
	Name           string   //域名
	Owner          string   //拥有者
	NetIds         []string //节点地址
	AddrCoins      []string //钱包收款地址
	Height         uint64   //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64   //有效块数量
	Deposit        uint64   //冻结金额
	IsMultName     bool     //是否多签域名
}

/*
查询域名
*/
func FindName(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	nameStr := nameItr.(string)
	nameinfo := namenet.FindNameToNet(nameStr)
	if nameinfo == nil || nameinfo.CheckIsOvertime(mining.GetHighestBlock()) {
		res, err = model.Errcode(model.NotExist, nameStr)
		return
	}

	//bs, err := json.Marshal(nameinfo)
	//if err != nil {
	//	res, err = model.Errcode(model.Nomarl, "find name formate failt 1")
	//	return
	//}
	//result := make(map[string]interface{})
	//// err = json.Unmarshal(bs, &result)
	//decoder := json.NewDecoder(bytes.NewBuffer(bs))
	//decoder.UseNumber()
	//err = decoder.Decode(&result)
	//if err != nil {
	//	res, err = model.Errcode(model.Nomarl, "find name formate failt 2")
	//	return
	//}
	// result["DepositMin"] = store.DepositMin
	nets := make([]string, 0)
	for _, one := range nameinfo.NetIds {
		nets = append(nets, one.B58String())
	}
	addrs := make([]string, 0)
	for _, one := range nameinfo.AddrCoins {
		addrs = append(addrs, one.B58String())
	}
	voOne := NameinfoVO{
		Name:           nameinfo.Name,                            //域名
		Owner:          nameinfo.Owner.B58String(),               //
		NetIds:         nets,                                     //节点地址
		AddrCoins:      addrs,                                    //钱包收款地址
		Height:         nameinfo.Height,                          //注册区块高度，通过现有高度计算出有效时间
		NameOfValidity: nameinfo.Height + namenet.NameOfValidity, //有效块数量
		Deposit:        nameinfo.Deposit,
		IsMultName:     nameinfo.IsMultName,
	}

	res, err = model.Tojson(voOne)
	return
}

/*
多签域名注册
*/
func MultNameInReg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	multAddressStr, res, err := getParamString(rj, "multaddress")
	if err != nil {
		return res, err
	}

	multAddress := crypto.AddressFromB58String(multAddressStr)
	amount, res, err := getParamUint64(rj, "amount")
	if err != nil {
		return res, err
	}

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	frozenHeight, res, err := getParamUint64(rj, "frozenHeight")
	if err != nil {
		frozenHeight = 0
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	account, res, err := getParamString(rj, "name")
	if err != nil {
		return res, err
	}
	netidsStr, errcode := getArrayStrParams(rj, "netids")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "netids")
		return
	}
	netids := []nodeStore.AddressNet{}
	for _, netidOne := range netidsStr {
		if crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(netidOne)) {
			res, err = model.Errcode(ContentIncorrectFormat, "invalid netids")
			return
		}
		idOne := nodeStore.AddressFromB58String(netidOne)
		if netidOne != "" && idOne.B58String() == "" {
			res, err = model.Errcode(ContentIncorrectFormat, netidOne)
			return
		}
		netids = append(netids, idOne)
	}

	addrCoinsStr, errcode := getArrayStrParams(rj, "addrcoins")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "addrcoins")
		return
	}
	addrCoins := []crypto.AddressCoin{}
	for _, addrcoinOne := range addrCoinsStr {
		idOne := crypto.AddressFromB58String(addrcoinOne)
		if !crypto.ValidAddr(config.AddrPre, idOne) {
			res, err = model.Errcode(ContentIncorrectFormat, "invalid addrcoins")
			return
		}
		addrCoins = append(addrCoins, idOne)
	}

	txItr, err := mining.BuildRequestMultsignNameTx(multAddress, amount, gas, frozenHeight, pwd, account, netids, addrCoins, mining.NameInActionReg)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
多签域名转让
*/
func MultNameInTransfer(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	multAddressStr, res, err := getParamString(rj, "multaddress")
	if err != nil {
		return res, err
	}
	multAddress := crypto.AddressFromB58String(multAddressStr)

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	frozenHeight, res, err := getParamUint64(rj, "frozenHeight")
	if err != nil {
		frozenHeight = 0
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	account, res, err := getParamString(rj, "name")
	if err != nil {
		return res, err
	}

	txItr, err := mining.BuildRequestMultsignNameTx(multAddress, 0, gas, frozenHeight, pwd, account, nil, nil, mining.NameInActionTransfer)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
多签域名续费
*/
func MultNameInRenew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	multAddressStr, res, err := getParamString(rj, "multaddress")
	if err != nil {
		return res, err
	}

	multAddress := crypto.AddressFromB58String(multAddressStr)

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	frozenHeight, res, err := getParamUint64(rj, "frozenHeight")
	if err != nil {
		frozenHeight = 0
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	account, res, err := getParamString(rj, "name")
	if err != nil {
		return res, err
	}

	txItr, err := mining.BuildRequestMultsignNameTx(multAddress, 0, gas, frozenHeight, pwd, account, nil, nil, mining.NameInActionRenew)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
多签域名更新
*/
func MultNameInUpdate(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	multAddressStr, res, err := getParamString(rj, "multaddress")
	if err != nil {
		return res, err
	}

	multAddress := crypto.AddressFromB58String(multAddressStr)

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	frozenHeight, res, err := getParamUint64(rj, "frozenHeight")
	if err != nil {
		frozenHeight = 0
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	account, res, err := getParamString(rj, "name")
	if err != nil {
		return res, err
	}
	netidsStr, errcode := getArrayStrParams(rj, "netids")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "netids")
		return
	}
	netids := []nodeStore.AddressNet{}
	for _, netidOne := range netidsStr {
		if crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(netidOne)) {
			res, err = model.Errcode(ContentIncorrectFormat, "invalid netids")
			return
		}
		idOne := nodeStore.AddressFromB58String(netidOne)
		if netidOne != "" && idOne.B58String() == "" {
			res, err = model.Errcode(ContentIncorrectFormat, netidOne)
			return
		}
		netids = append(netids, idOne)
	}

	addrCoinsStr, errcode := getArrayStrParams(rj, "addrcoins")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "addrcoins")
		return
	}
	addrCoins := []crypto.AddressCoin{}
	for _, addrcoinOne := range addrCoinsStr {
		idOne := crypto.AddressFromB58String(addrcoinOne)
		if !crypto.ValidAddr(config.AddrPre, idOne) {
			res, err = model.Errcode(ContentIncorrectFormat, "invalid addrcoins")
			return
		}
		addrCoins = append(addrCoins, idOne)
	}

	txItr, err := mining.BuildRequestMultsignNameTx(multAddress, 0, gas, frozenHeight, pwd, account, netids, addrCoins, mining.NameInActionUpdate)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
多签域名注销
*/
func MultNameOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	multAddressStr, res, err := getParamString(rj, "multaddress")
	if err != nil {
		return res, err
	}

	multAddress := crypto.AddressFromB58String(multAddressStr)

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	frozenHeight, res, err := getParamUint64(rj, "frozenHeight")
	if err != nil {
		frozenHeight = 0
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	account, res, err := getParamString(rj, "name")
	if err != nil {
		return res, err
	}

	txItr, err := mining.BuildRequestMultsignNameTx(multAddress, 0, gas, frozenHeight, pwd, account, nil, nil, mining.NameOutAction)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}
