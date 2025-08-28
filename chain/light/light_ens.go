package light

import (
	"encoding/hex"
	"encoding/json"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/rpc"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

// import (
//
//	"bytes"
//	"encoding/hex"
//	"encoding/json"
//	"web3_gui/keystore/adapter/crypto"
//	"web3_gui/libp2parea/adapter/engine"
//	"web3_gui/libp2parea/adapter/message_center"
//	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
//	"web3_gui/utils"
//	"web3_gui/chain/config"
//	"web3_gui/chain/db"
//	"web3_gui/chain/evm/common"
//	"web3_gui/chain/evm/common/evmutils"
//	"web3_gui/chain/evm/precompiled/ens"
//	"web3_gui/chain/mining"
//	"web3_gui/chain/rpc"
//	"math/big"
//	"sort"
//	"strings"
//	"sync"
//	"time"
//
// )
//
// //todo:设计到keystore从本地加载地址数据库，轻节点调用可能会存在一定问题
func RegisterEnsMsg() {
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR, DelayBaseRegister)           //部署注册器
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_ADDDOMAIN, AddDomain)                           //添加域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_ABORTADDDOMAIN, AbortAddDomain)                 //终止添加域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN, ModifyAddDomain)               //修改添加域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SETDOMAINMANGER, SetDomainManger)               //设置控制器合约为根域名管理员，同时设置解析器
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER, SetDomainImResolver)       //解析主链地址，同时设置解析器
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER, SetDomainOtherResolver) //设置其他币种解析
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_DOMAINTRANSFER, DomainTransfer)                 //转让
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_DOMAINWITHDRAW, DomainWithDraw)                 //提现
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_REGISTERDOMAIN, RegisterDomain)                 //注册域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_RENEWDOMAIN, ReNewDomain)                       //续费域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS, GetLaunchDomains)             //获取投放的域名列表
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETHOLDERNUM, GetHolderNum)                     //获取持有人数量
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINEXP, GetDomainExp)                     //获取域名过期时间
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINOWNER, GetDomainOwner)                 //获取域名持有人
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINDETAIL, GetDomainDetail)               //获取域名详情
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER, GetDomainResolver)           //获取域名的解析
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER, GetDomainOtherResolver) //获取其他币种解析
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETMYROOTDOMAIN, GetMyRootDomain)               //获取我名下的根域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETMYSUBDOMAIN, GetMySubDomain)                 //获取我名下的子域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETENSCONTRACT, GetEnsContract)                 //获取平台域名的基础合约地址
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETROOTINCOME, GetRootInCome)                   //获取域名收入
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1, FindBlockRangeV1) //获取一定区块高度范围内的区块新版本
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_LAUNCHDOMAIN, LaunchDomain)                     //投放域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN, AbortLaunchDomain)           //终止投放域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN, ModifyLaunchDomain)         //修改投放域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SETLOCKDOMAIN, SetLockDomain)                   //锁定域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_UNLOCKDOMAIN, UnLockDomain)                     //解锁域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETLOCKDOMAINS, GetLockDomains)                 //获取锁定域名列表
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDOMAINCOST, GetDomainCost)                   //获取域名费用
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST, GetSingleDomainCost)       //获取单个域名费用
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETHOTROOTDOMAIN, GetHotRootDomain)             //获取热度高的根域名
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME, SetReverseResolverName) //反向解析设置名称
	//	Area.Register_neighbor(config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER, DelDomainImResolver)       //删除解析
}

// // 部署注册器
//
//	func DelayBaseRegister(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		ensAddr, ok := rj.Get("ens")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.NoField, "ens"))
//			return
//		}
//		addr := ensAddr.(string)
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		lockNameAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.LOCKNAME_ADDR).Bytes())
//		comment := ens.BuildDelayBaseRegistarInput(addr, lockNameAddr.B58String(), name)
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, nil, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		result["contract_address"] = txpay.Vout[0].Address.B58String()
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELAYBASEREGISTAR_REV, pkg(model.Success, result))
//	}
//
// // 添加域名
//
//	func AddDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//		//域名名字
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		priceItr, ok := rj.Get("price")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "price"))
//			return
//		}
//		price := int64(priceItr.(float64))
//
//		openTimeItr, ok := rj.Get("open_time")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "open_time"))
//			return
//		}
//		openTime := int64(openTimeItr.(float64))
//
//		foreverPriceItr, ok := rj.Get("forever_price")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.NoField, "forever_price"))
//			return
//		}
//		foreverPrice := int64(foreverPriceItr.(float64))
//
//		comment := ens.BuildAddDomainInput(name, big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice), "", "")
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDDOMAIN_REV, pkg(model.Success, result))
//	}
//
// //终止添加域名
//
//	func AbortAddDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//		//域名名字
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//
//		comment := ens.BuildAbortAddDomainInput(name)
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTADDDOMAIN_REV, pkg(model.Success, result))
//	}
//
// //修改添加域名
//
//	func ModifyAddDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//		//域名名字
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		priceItr, ok := rj.Get("price")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "price"))
//			return
//		}
//		price := int64(priceItr.(float64))
//
//		openTimeItr, ok := rj.Get("open_time")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "open_time"))
//			return
//		}
//		openTime := int64(openTimeItr.(float64))
//
//		foreverPriceItr, ok := rj.Get("forever_price")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.NoField, "forever_price"))
//			return
//		}
//		foreverPrice := int64(foreverPriceItr.(float64))
//
//		comment := ens.BuildModifyAddDomainInput(name, big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice), "", "")
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYADDDOMAIN_REV, pkg(model.Success, result))
//	}
//
// // 设置域名管理员为某注册器合约
//
//	func SetDomainManger(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		//节点名称
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		//注册器地址
//		registarItr, ok := rj.Get("registar")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.NoField, "registar"))
//			return
//		}
//		registar := registarItr.(string)
//		comment := ""
//		if name == "" {
//			comment = ens.BuildNodeOwnerInput(name, registar)
//		} else {
//			comment = ens.BuildSubNodeRecordInput("", name, registar)
//		}
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINMANGER_REV, pkg(model.Success, result))
//	}
//
// // 设置域名解析主
//
//	func SetDomainImResolver(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//		contractAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.PUBLIC_RESOLVER_ADDR).Bytes())
//
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		//节点名称
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		//节点名称
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//
//		if root == "" && sub == "" {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, ""))
//			return
//		}
//
//		//要解析到的Im地址
//		imItr, ok := rj.Get("im_address")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.NoField, "im_address"))
//			return
//		}
//		imAddr := imItr.(string)
//		//判断域名持有人是否是当前账户
//		domainOwner := ens.GetDomainOwner(src.B58String(), root, sub)
//		if domainOwner != src.B58String() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.Nomarl, "srcaddress must be domain owner"))
//			return
//		}
//
//		node := sub + "." + root
//		if root == "" {
//			node = sub
//		} else if sub == "" {
//			node = root
//		}
//
//		comment := ens.BuildImResolverInput(node, imAddr)
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINIMRESOLVER_REV, pkg(model.Success, result))
//	}
//
// // 设置其他币种解析
//
//	func SetDomainOtherResolver(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//		contractAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.PUBLIC_RESOLVER_ADDR).Bytes())
//
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		//节点名称
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		//节点名称
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//
//		if root == "" && sub == "" {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, ""))
//			return
//		}
//
//		coinTypeItr, ok := rj.Get("coin_type")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "coin_type"))
//			return
//		}
//		coinType := int64(coinTypeItr.(float64))
//		//要解析到的其他链地址
//		imItr, ok := rj.Get("other_address")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "im_address"))
//			return
//		}
//		imAddr := imItr.(string)
//		//判断域名持有人是否是当前账户
//		domainOwner := ens.GetDomainOwner(src.B58String(), root, sub)
//		if domainOwner != src.B58String() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.Nomarl, "srcaddress must be domain owner"))
//			return
//		}
//		node := sub + "." + root
//		if root == "" {
//			node = sub
//		} else if sub == "" {
//			node = root
//		}
//		comment := ens.BuildOtherResolverInput(node, imAddr, big.NewInt(coinType))
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETDOMAINOTHERRESOLVER_REV, pkg(model.Success, result))
//	}
//
// // 根域名持有人和平台持有人提现
//
//	func DomainTransfer(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		fromItr, ok := rj.Get("from")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "owner"))
//			return
//		}
//		from := fromItr.(string)
//		toItr, ok := rj.Get("to")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "to"))
//			return
//		}
//		to := toItr.(string)
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		comment := ens.BuildTransferInput(from, to, name)
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINTRANSFER_REV, pkg(model.Success, result))
//	}
//
// // 根域名持有人和平台持有人提现
//
//	func DomainWithDraw(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		amountItr, ok := rj.Get("amount")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.NoField, "amount"))
//			return
//		}
//		amount := uint64(amountItr.(float64))
//
//		comment := ens.BuildWithDrawInput(new(big.Int).SetUint64(amount))
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//		if total < gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DOMAINWITHDRAW_REV, pkg(model.Success, result))
//	}
//
// //注册域名
//
//	func RegisterDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		amountItr, ok := rj.Get("amount")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "amount"))
//			return
//		}
//		amount := uint64(amountItr.(float64))
//		if amount < 0 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//
//		//节点名称
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		forever := false
//		foreverItr, ok := rj.Get("forever")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "forever"))
//			return
//		}
//		forever = foreverItr.(bool)
//
//		//如果合约地址是平台基础注册器，则域名为根域名注册，默认为永久注册
//		baseRegistarAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.BASE_REGISTAR_ADDR).Bytes())
//		if addr == baseRegistarAddr.B58String() {
//			forever = true
//		}
//
//		duration := uint64(0)
//		durationItr, ok := rj.Get("duration")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "duration"))
//			return
//		}
//		duration = uint64(durationItr.(float64))
//		if duration < 365*24*60*60 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Nomarl, "min duration is 31536000"))
//			return
//		}
//		if !forever && duration%31536000 != 0 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Nomarl, "duration must multiple 31536000"))
//			return
//		}
//		//如果是永久，时间设置为1万年
//		if forever {
//			duration = 31536000 * 10000
//		}
//		ownerItr, ok := rj.Get("owner")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.NoField, "owner"))
//			return
//		}
//		owner := ownerItr.(string)
//		secret32, _ := crypto.Rand32Byte()
//		comment := ens.BuildRegisterInput(name, owner, big.NewInt(int64(duration)), secret32)
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
//		if total < amount+gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_REGISTERDOMAIN_REV, pkg(model.Success, result))
//	}
//
// // 续费域名
//
//	func ReNewDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		var src crypto.AddressCoin
//		addrItr, ok := rj.Get("srcaddress")
//		if ok {
//			srcaddr := addrItr.(string)
//			if srcaddr != "" {
//				src = crypto.AddressFromB58String(srcaddr)
//				//判断地址前缀是否正确
//				if !crypto.ValidAddr(config.AddrPre, src) {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//				_, ok := config.Area.Keystore.FindAddress(src)
//				if !ok {
//					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//					return
//				}
//			}
//		}
//
//		addrItr, ok = rj.Get("contractaddress")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//			return
//		}
//		addr := addrItr.(string)
//
//		contractAddr := crypto.AddressFromB58String(addr)
//		if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//			return
//		}
//		gasItr, ok := rj.Get("gas")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "gas"))
//			return
//		}
//		gas := uint64(gasItr.(float64))
//
//		gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//		gasPriceItr, ok := rj.Get("gas_price")
//		if ok {
//			gasPrice = uint64(gasPriceItr.(float64))
//			if gasPrice < config.DEFAULT_GAS_PRICE {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//				return
//			}
//		}
//		frozenHeight := uint64(0)
//		frozenHeightItr, ok := rj.Get("frozen_height")
//		if ok {
//			frozenHeight = uint64(frozenHeightItr.(float64))
//		}
//
//		pwdItr, ok := rj.Get("pwd")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "pwd"))
//			return
//		}
//		pwd := pwdItr.(string)
//
//		amountItr, ok := rj.Get("amount")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "amount"))
//			return
//		}
//		amount := uint64(amountItr.(float64))
//		if amount < 0 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//
//		//节点名称
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		forever := false
//		foreverItr, ok := rj.Get("forever")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "forever"))
//			return
//		}
//		forever = foreverItr.(bool)
//		duration := uint64(0)
//		durationItr, ok := rj.Get("duration")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.NoField, "duration"))
//			return
//		}
//		duration = uint64(durationItr.(float64))
//		if duration < 365*24*60*60 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Nomarl, "min duration is 31536000"))
//			return
//		}
//		if !forever && duration%31536000 != 0 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Nomarl, "duration must multiple 31536000"))
//			return
//		}
//		//如果是永久，时间设置为1万年
//		if forever {
//			duration = 31536000 * 10000
//		}
//
//		comment := ens.BuildReNewInput(name, big.NewInt(int64(duration)))
//
//		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
//		if total < amount+gas*gasPrice {
//			//资金不够
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//			return
//		}
//		txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//		if err != nil {
//			if err.Error() == config.ERROR_password_fail.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.FailPwd, ""))
//				return
//			}
//			if err.Error() == config.ERROR_amount_zero.Error() {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//				return
//			}
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//
//		result, err := utils.ChangeMap(txpay)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//			return
//		}
//		result["hash"] = hex.EncodeToString(*txpay.GetHash())
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RENEWDOMAIN_REV, pkg(model.Success, result))
//	}
//
// // 获取域名
//
//	func GetLaunchDomains(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		ensAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.ENS_CONTRACT_ADDR).Bytes())
//		addr := ensAddr.B58String()
//		nameItr, ok := rj.Get("name")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.NoField, "name"))
//			return
//		}
//		name := nameItr.(string)
//		if name != "" && !strings.Contains(name, ".") {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.Nomarl, "name must be null or .xx or xx.xx"))
//			return
//		}
//		nameArray := strings.Split(name, ".")
//		if len(nameArray) > 2 {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.Nomarl, "name must be null or .xx or xx.xx"))
//			return
//		}
//		if len(nameArray) == 2 && nameArray[1] == "" {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.Nomarl, "name must be null or .xx or xx.xx"))
//			return
//		}
//
//		page, ok := rj.Get("page")
//		if !ok {
//			page = float64(1)
//		}
//		pageInt := int(page.(float64))
//
//		pageSize, ok := rj.Get("page_size")
//		if !ok {
//			pageSize = float64(10)
//		}
//		pageSizeInt := int(pageSize.(float64))
//
//		list := ens.GetOpenDomain(config.Area.Keystore.GetCoinbase().Addr.B58String(), name, addr)
//		total := len(list)
//		start := (pageInt - 1) * pageSizeInt
//		end := start + pageSizeInt
//		if start > total {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.Success, list))
//			return
//		}
//		if end > total {
//			end = total
//		}
//		data := make(map[string]interface{})
//		data["total"] = total
//		data["list"] = list[start:end]
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLAUNCHDOMAINS_REV, pkg(model.Success, data))
//	}
//
// // 获取持有人数量
//
//	func GetHolderNum(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOLDERNUM_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		registarAddr, ok := rj.Get("contract")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOLDERNUM_REV, pkg(model.NoField, "contract"))
//			return
//		}
//		addr := registarAddr.(string)
//		num := ens.GetHolderNum(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOLDERNUM_REV, pkg(model.Success, num))
//	}
//
// // 获取域名过期时间
//
//	func GetDomainExp(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINEXP_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINEXP_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINEXP_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//		num := ens.GetDomainExp(config.Area.Keystore.GetCoinbase().Addr.B58String(), root, sub)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINEXP_REV, pkg(model.Success, num))
//	}
//
// // 获取域名持有人
//
//	func GetDomainOwner(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOWNER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOWNER_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOWNER_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//		num := ens.GetDomainOwner(config.Area.Keystore.GetCoinbase().Addr.B58String(), root, sub)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOWNER_REV, pkg(model.Success, num))
//	}
//
// // 获取域名详情
//
//	func GetDomainDetail(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINDETAIL_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINDETAIL_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINDETAIL_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//		owner, domainExp, manager, canRegister := ens.GetDomainDetail(config.Area.Keystore.GetCoinbase().Addr.B58String(), root, sub)
//		data := make(map[string]interface{})
//		data["owner"] = owner
//		data["domain_exp"] = domainExp
//		data["manager"] = manager
//		data["can_register"] = canRegister
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINDETAIL_REV, pkg(model.Success, data))
//	}
//
// // 获取域名解析
//
//	func GetDomainResolver(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//
//		if root == "" && sub == "" {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER_REV, pkg(model.NoField, ""))
//			return
//		}
//
//		node := sub + "." + root
//		if root == "" {
//			node = sub
//		} else if sub == "" {
//			node = root
//		}
//		num := ens.GetDomainResolver(config.Area.Keystore.GetCoinbase().Addr.B58String(), node)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINRESOLVER_REV, pkg(model.Success, num))
//	}
//
// // 获取其他币种域名解析
//
//	func GetDomainOtherResolver(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		subItr, ok := rj.Get("sub")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "sub"))
//			return
//		}
//		sub := subItr.(string)
//
//		if root == "" && sub == "" {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, ""))
//		}
//
//		coinTypeItr, ok := rj.Get("coin_type")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(model.NoField, "coin_type"))
//			return
//		}
//		coinType := int64(coinTypeItr.(float64))
//		node := sub + "." + root
//		if root == "" {
//			node = sub
//		} else if sub == "" {
//			node = root
//		}
//		num := ens.GetDomainOtherResolver(config.Area.Keystore.GetCoinbase().Addr.B58String(), node, big.NewInt(coinType))
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINOTHERRESOLVER_REV, pkg(model.Success, num))
//	}
//
// // 获取我名下的根域名
//
//	func GetMyRootDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMYROOTDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		addrList := config.Area.Keystore.GetAddr()
//		addrMap := make(map[string]struct{})
//		for _, v := range addrList {
//			addrMap[v.Addr.B58String()] = struct{}{}
//		}
//		list := ens.GetMyRootDomain(addrList[0].Addr.B58String(), addrMap)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMYROOTDOMAIN_REV, pkg(model.Success, list))
//	}
//
// // 获取我名下的子域名
//
//	func GetMySubDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMYSUBDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		addrList := config.Area.Keystore.GetAddr()
//		addrMap := make(map[string]struct{})
//		for _, v := range addrList {
//			addrMap[v.Addr.B58String()] = struct{}{}
//		}
//		list := ens.GetMySubDomain(addrList[0].Addr.B58String(), addrMap)
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMYSUBDOMAIN_REV, pkg(model.Success, list))
//	}
//
// // 获取域名合约地址
//
//	func GetEnsContract(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETENSCONTRACT_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		data := make(map[string]interface{})
//		ensAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.ENS_CONTRACT_ADDR).Bytes())
//		data["ens"] = ensAddr.B58String()
//		registarAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.BASE_REGISTAR_ADDR).Bytes())
//		data["registar"] = registarAddr.B58String()
//		resolverAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.PUBLIC_RESOLVER_ADDR).Bytes())
//		data["resolver"] = resolverAddr.B58String()
//		locknameAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.LOCKNAME_ADDR).Bytes())
//		data["lockname"] = locknameAddr.B58String()
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETENSCONTRACT_REV, pkg(model.Success, data))
//	}
//
// // 获取根域名收入
//
//	func GetRootInCome(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//		rj := new(model.RpcJson)
//		err := json.Unmarshal(*message.Body.Content, &rj)
//		if err != nil {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETROOTINCOME_REV, pkg(rpc.SystemError, err.Error()))
//			return
//		}
//		rootItr, ok := rj.Get("root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETROOTINCOME_REV, pkg(model.NoField, "root"))
//			return
//		}
//		root := rootItr.(string)
//		income, withdraw, address := ens.GetIncome(config.Area.Keystore.GetCoinbase().Addr.B58String(), root)
//		data := make(map[string]interface{})
//		data["all_income"] = income
//		data["curr_withdraw_income"] = withdraw
//		data["now_income"] = 0
//		if income != big.NewInt(0) {
//			contract := crypto.AddressFromB58String(address)
//			_, balance := db.GetNotSpendBalance(&contract)
//			data["now_income"] = balance
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETROOTINCOME_REV, pkg(model.Success, data))
//	}
//
// //通过一定范围的区块高度查询多个区块详细信息新版本
func FindBlockRangeV1(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(model.NoField, "startHeight"))
		return
	}
	startHeight := uint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(model.NoField, "endHeight"))
		return
	}
	endHeight := uint64(endHeightItr.(float64))

	if endHeight < startHeight {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(model.NoField, "endHeight"))
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
	}
	//待返回的区块
	start := config.TimeNow()
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {

		bhvo := BlockHeadOut{}
		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		txResChan := make(chan interface{}, len(bh.Tx))
		txMap := make(map[string]interface{})
		var wg sync.WaitGroup
		for _, one := range bh.Tx {
			wg.Add(1)
			go func(hash []byte) {
				defer wg.Done()
				txItrJson, code, txItr := mining.FindTxJsonVoV1(hash)
				if txItr == nil {
					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(model.Nomarl, "交易不存在"))
					return
				}
				item := rpc.JsonMethod(txItrJson)
				item["block_height"] = bh.Height
				item["timestamp"] = bh.Time
				item["blockhash"] = hex.EncodeToString(bh.Hash)
				txClass, vins, vouts := rpc.DealTxInfoV2(txItr)
				if txClass > 0 {
					item["type"] = txClass
					if vins != nil {
						item["vin"] = vins
					}
					if vouts != nil {
						item["vout"] = vouts
						item["vout_total"] = len(vouts.([]mining.VoutVO))
					}
				}
				//合约交易是20
				var (
					gasUsed  uint64
					gasLimit uint64
					gasPrice uint64
				)

				if txItr.Class() == config.Wallet_tx_type_contract {
					tx := txItr.(*mining.Tx_Contract)
					gasUsed = tx.GetGasLimit()
					gasLimit = config.EVM_GAS_MAX
					gasPrice = tx.GasPrice
				} else {
					gasUsed = txItr.GetGas()
					gasLimit = gasUsed
					gasPrice = config.DEFAULT_GAS_PRICE
				}
				item["free"] = txItr.GetGas()
				item["gas_used"] = gasUsed
				item["gas_limit"] = gasLimit
				item["gas_price"] = gasPrice
				item["upchaincode"] = code
				txResChan <- item
			}(one)

		}
		go func() {
			defer close(txResChan)
			wg.Wait()
		}()
		for v := range txResChan {
			tx := v.(map[string]interface{})
			hash := tx["hash"].(string)
			txMap[hash] = tx
		}
		for _, one := range bh.Tx {
			txhash := hex.EncodeToString(one)
			if _, ok := txMap[txhash]; ok {
				value := txMap[txhash].(map[string]interface{})
				bhvo.Txs = append(bhvo.Txs, value)
			}

			if _, ok := txMap[txhash+"-1"]; ok {
				value := txMap[txhash+"-1"].(map[string]interface{})
				newTx := append([]map[string]interface{}{value}, bhvo.Txs...)
				bhvo.Txs = newTx
			}
		}
		bhvos = append(bhvos, bhvo)
	}
	end := config.TimeNow().Sub(start)
	engine.Log.Info("3221行%s,%s", end, config.TimeNow().Sub(start))
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FINDBLOCKRANGEV1_REV, pkg(model.Success, bhvos))

}

//
////域名投放
//func LaunchDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//
//	addrItr, ok = rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//
//	addr := addrItr.(string)
//
//	contractAddr := crypto.AddressFromB58String(addr)
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	lenItr, ok := rj.Get("len")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "len"))
//		return
//	}
//	length := int64(lenItr.(float64))
//
//	priceItr, ok := rj.Get("price")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "price"))
//		return
//	}
//	price := int64(priceItr.(float64))
//
//	openTimeItr, ok := rj.Get("open_time")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "open_time"))
//		return
//	}
//	openTime := int64(openTimeItr.(float64))
//
//	foreverPriceItr, ok := rj.Get("forever_price")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.NoField, "forever_price"))
//		return
//	}
//	foreverPrice := int64(foreverPriceItr.(float64))
//
//	comment := ens.BuildLaunchDomainInput(big.NewInt(length), big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice))
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_LAUNCHDOMAIN_REV, pkg(model.Success, result))
//}
//
////终止投放域名
//func AbortLaunchDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//
//	addrItr, ok = rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//
//	addr := addrItr.(string)
//
//	contractAddr := crypto.AddressFromB58String(addr)
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	lenItr, ok := rj.Get("len")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.NoField, "len"))
//		return
//	}
//	length := int64(lenItr.(float64))
//
//	comment := ens.BuildAbortLaunchDomainInput(big.NewInt(length))
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ABORTLAUNCHDOMAIN_REV, pkg(model.Success, result))
//}
//
////修改投放域名
//func ModifyLaunchDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//
//	addrItr, ok = rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//
//	addr := addrItr.(string)
//
//	contractAddr := crypto.AddressFromB58String(addr)
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	lenItr, ok := rj.Get("len")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "len"))
//		return
//	}
//	length := int64(lenItr.(float64))
//
//	priceItr, ok := rj.Get("price")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "price"))
//		return
//	}
//	price := int64(priceItr.(float64))
//
//	openTimeItr, ok := rj.Get("open_time")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "open_time"))
//		return
//	}
//	openTime := int64(openTimeItr.(float64))
//
//	foreverPriceItr, ok := rj.Get("forever_price")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.NoField, "forever_price"))
//		return
//	}
//	foreverPrice := int64(foreverPriceItr.(float64))
//
//	comment := ens.BuildModifyLaunchDomainInput(big.NewInt(length), big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice))
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_MODIFYLAUNCHDOMAIN_REV, pkg(model.Success, result))
//}
//
////域名锁定
//func SetLockDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//
//	addrItr, ok = rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//
//	addr := addrItr.(string)
//
//	contractAddr := crypto.AddressFromB58String(addr)
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//
//	globalLockNameAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.LOCKNAME_ADDR).Bytes())
//	var isRoot bool
//	if addr == globalLockNameAddr.B58String() {
//		rootItr, ok := rj.Get("is_root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.NoField, "is_root"))
//			return
//		}
//		isRoot = rootItr.(bool)
//	}
//
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	namesP, ok := rj.Get("names")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.NoField, "names"))
//		return
//	}
//
//	bs, err := json.Marshal(namesP)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.TypeWrong, "names"))
//		return
//	}
//	names := make([]string, 0)
//	decoder := json.NewDecoder(bytes.NewBuffer(bs))
//	err = decoder.Decode(&names)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.TypeWrong, "names"))
//		return
//	}
//
//	comment := ens.BuildLockNameInput(names, &contractAddr, isRoot)
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETLOCKDOMAIN_REV, pkg(model.Success, result))
//}
//
////域名解锁
//func UnLockDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//
//	addrItr, ok = rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//
//	addr := addrItr.(string)
//
//	contractAddr := crypto.AddressFromB58String(addr)
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//
//	globalLockNameAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.LOCKNAME_ADDR).Bytes())
//	var isRoot bool
//	if addr == globalLockNameAddr.B58String() {
//		rootItr, ok := rj.Get("is_root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.NoField, "is_root"))
//			return
//		}
//		isRoot = rootItr.(bool)
//	}
//
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	namesP, ok := rj.Get("names")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.NoField, "names"))
//		return
//	}
//
//	bs, err := json.Marshal(namesP)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.TypeWrong, "names"))
//		return
//	}
//	names := make([]string, 0)
//	decoder := json.NewDecoder(bytes.NewBuffer(bs))
//	err = decoder.Decode(&names)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.TypeWrong, "names"))
//		return
//	}
//
//	comment := ens.BuildUnLockNameInput(names, &contractAddr, isRoot)
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_UNLOCKDOMAIN_REV, pkg(model.Success, result))
//}
//
////获取锁定域名列表
//func GetLockDomains(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLOCKDOMAINS_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	contractItr, ok := rj.Get("contractaddress")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLOCKDOMAINS_REV, pkg(model.NoField, "contractaddress"))
//		return
//	}
//	contractStr := contractItr.(string)
//	contractAddr := crypto.AddressFromB58String(contractStr)
//
//	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLOCKDOMAINS_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
//		return
//	}
//
//	globalLockNameAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.LOCKNAME_ADDR).Bytes())
//	var isRoot bool
//	isGloabl := false
//	if contractStr == globalLockNameAddr.B58String() {
//		isGloabl = true
//		rootItr, ok := rj.Get("is_root")
//		if !ok {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLOCKDOMAINS_REV, pkg(model.NoField, "is_root"))
//			return
//		}
//		isRoot = rootItr.(bool)
//	}
//
//	result := make(map[string][]string)
//	lockNameList := make([]string, 0)
//	unLockNameList := make([]string, 0)
//	if isGloabl {
//		lockNameList = ens.GetGlobalLockNames(config.Area.Keystore.GetCoinbase().Addr, contractAddr, isRoot)
//	} else {
//		lockNameList, unLockNameList = ens.GetLockNames(config.Area.Keystore.GetCoinbase().Addr, contractAddr)
//	}
//	result["lockNameList"] = lockNameList
//	result["unLockNameList"] = unLockNameList
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLOCKDOMAINS_REV, pkg(model.Success, result))
//}
//
//// 获取域名费用
//func GetDomainCost(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINCOST_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	rootItr, ok := rj.Get("name")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINCOST_REV, pkg(model.NoField, "name"))
//		return
//	}
//	root := rootItr.(string)
//	list := ens.GetDomainCost(config.Area.Keystore.GetCoinbase().Addr.B58String(), root)
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDOMAINCOST_REV, pkg(model.Success, list))
//}
//
//// 获取单个域名费用
//func GetSingleDomainCost(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	rootItr, ok := rj.Get("root")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST_REV, pkg(model.NoField, "root"))
//		return
//	}
//	root := rootItr.(string)
//	subItr, ok := rj.Get("sub")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST_REV, pkg(model.NoField, "sub"))
//		return
//	}
//	sub := subItr.(string)
//
//	if root == "" && sub == "" {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST_REV, pkg(model.NoField, ""))
//		return
//	}
//	info := ens.GetSingleDomainCost(config.Area.Keystore.GetCoinbase().Addr.B58String(), root, sub)
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSINGLEDOMAINCOST_REV, pkg(model.Success, info))
//}
//
//// 获取热度高的根域名
//func GetHotRootDomain(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOTROOTDOMAIN_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	numItr, ok := rj.Get("num")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOTROOTDOMAIN_REV, pkg(model.NoField, "num"))
//		return
//	}
//	num := int(numItr.(float64))
//
//	list := ens.GetDomainStats(config.Area.Keystore.GetCoinbase().Addr.B58String())
//	sort.Sort(ens.DomainStatsSort(list))
//
//	if len(list) > num {
//		list = list[0:num]
//	}
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETHOTROOTDOMAIN_REV, pkg(model.Success, list))
//}
//
//// 反向解析设置名称
//func SetReverseResolverName(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//	//contractAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.PUBLIC_RESOLVER_ADDR).Bytes())
//	//
//	//gasItr, ok := rj.Get("gas")
//	//if !ok {
//	//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.NoField, "gas"))
//	//	return
//	//}
//	//gas := uint64(gasItr.(float64))
//	//
//	//gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	//gasPriceItr, ok := rj.Get("gas_price")
//	//if ok {
//	//	gasPrice = uint64(gasPriceItr.(float64))
//	//	if gasPrice < config.DEFAULT_GAS_PRICE {
//	//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.Nomarl, "gas_price is too low"))
//	//		return
//	//	}
//	//}
//	//frozenHeight := uint64(0)
//	//frozenHeightItr, ok := rj.Get("frozen_height")
//	//if ok {
//	//	frozenHeight = uint64(frozenHeightItr.(float64))
//	//}
//	//
//	//pwdItr, ok := rj.Get("pwd")
//	//if !ok {
//	//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.NoField, "pwd"))
//	//	return
//	//}
//	//pwd := pwdItr.(string)
//
//	//节点名称
//	rootItr, ok := rj.Get("root")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.NoField, "root"))
//		return
//	}
//	root := rootItr.(string)
//	//节点名称
//	subItr, ok := rj.Get("sub")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.NoField, "sub"))
//		return
//	}
//	sub := subItr.(string)
//
//	if root == "" && sub == "" {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.NoField, ""))
//		return
//	}
//	//判断域名持有人是否是当前账户
//	domainOwner := ens.GetDomainOwner(src.B58String(), root, sub)
//	if domainOwner != src.B58String() {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.Nomarl, "srcaddress must be domain owner"))
//		return
//	}
//
//	node := sub + "." + root
//	if root == "" {
//		node = sub
//	} else if sub == "" {
//		node = root
//	}
//
//	ss := ens.BuildReversSetNameInput(node, src)
//	if ss != "" {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.Success, ss))
//		return
//	}
//	comment := ens.BuildReversGetNameInput(node, src)
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SETREVERSERESOLVERNAME_REV, pkg(model.Success, comment))
//	//total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	//if total < gas*gasPrice {
//	//	//资金不够
//	//	res, err = model.Errcode(BalanceNotEnough)
//	//	return
//	//}
//	///*------------------------*/
//	//txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	//// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
//	//if err != nil {
//	//	// engine.Log.Info("创建转账交易错误 11111111")
//	//	if err.Error() == config.ERROR_password_fail.Error() {
//	//		// engine.Log.Info("创建转账交易错误 222222222222")
//	//		res, err = model.Errcode(model.FailPwd)
//	//		return
//	//	}
//	//	// engine.Log.Info("创建转账交易错误 333333333333")
//	//	if err.Error() == config.ERROR_amount_zero.Error() {
//	//		res, err = model.Errcode(AmountIsZero, "amount")
//	//		return
//	//	}
//	//	res, err = model.Errcode(model.Nomarl, err.Error())
//	//	return
//	//}
//	//
//	//result, err := utils.ChangeMap(txpay)
//	//if err != nil {
//	//	res, err = model.Errcode(model.Nomarl, err.Error())
//	//	return
//	//}
//	//result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	//
//	//res, err = model.Tojson(result)
//	//
//	//return res, err
//}
//
////删除解析
//func DelDomainImResolver(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	var src crypto.AddressCoin
//	addrItr, ok := rj.Get("srcaddress")
//	if ok {
//		srcaddr := addrItr.(string)
//		if srcaddr != "" {
//			src = crypto.AddressFromB58String(srcaddr)
//			//判断地址前缀是否正确
//			if !crypto.ValidAddr(config.AddrPre, src) {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//			_, ok := config.Area.Keystore.FindAddress(src)
//			if !ok {
//				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
//				return
//			}
//		}
//	}
//	contractAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.PUBLIC_RESOLVER_ADDR).Bytes())
//
//	gasItr, ok := rj.Get("gas")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, "gas"))
//		return
//	}
//	gas := uint64(gasItr.(float64))
//
//	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
//	gasPriceItr, ok := rj.Get("gas_price")
//	if ok {
//		gasPrice = uint64(gasPriceItr.(float64))
//		if gasPrice < config.DEFAULT_GAS_PRICE {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.Nomarl, "gas_price is too low"))
//			return
//		}
//	}
//	frozenHeight := uint64(0)
//	frozenHeightItr, ok := rj.Get("frozen_height")
//	if ok {
//		frozenHeight = uint64(frozenHeightItr.(float64))
//	}
//
//	pwdItr, ok := rj.Get("pwd")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, "pwd"))
//		return
//	}
//	pwd := pwdItr.(string)
//
//	//节点名称
//	rootItr, ok := rj.Get("root")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, "root"))
//		return
//	}
//	root := rootItr.(string)
//	//节点名称
//	subItr, ok := rj.Get("sub")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, "sub"))
//		return
//	}
//	sub := subItr.(string)
//
//	if root == "" && sub == "" {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, ""))
//		return
//	}
//
//	coinTypeItr, ok := rj.Get("coin_type")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.NoField, "coin_type"))
//		return
//	}
//	coinType := int64(coinTypeItr.(float64))
//
//	//判断域名持有人是否是当前账户
//	domainOwner := ens.GetDomainOwner(src.B58String(), root, sub)
//	if domainOwner != src.B58String() {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.Nomarl, "srcaddress must be domain owner"))
//		return
//	}
//	node := sub + "." + root
//	if root == "" {
//		node = sub
//	} else if sub == "" {
//		node = root
//	}
//	comment := ens.BuildDelResolverInput(node, big.NewInt(coinType))
//
//	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
//	if total < gas*gasPrice {
//		//资金不够
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(rpc.BalanceNotEnough, ""))
//		return
//	}
//	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice)
//	if err != nil {
//		if err.Error() == config.ERROR_password_fail.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.FailPwd, ""))
//			return
//		}
//		if err.Error() == config.ERROR_amount_zero.Error() {
//			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(rpc.AmountIsZero, "amount"))
//			return
//		}
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//
//	result, err := utils.ChangeMap(txpay)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.Nomarl, err.Error()))
//		return
//	}
//	result["hash"] = hex.EncodeToString(*txpay.GetHash())
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELDOMAINIMRESOLVER_REV, pkg(model.Success, result))
//}
