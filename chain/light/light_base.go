package light

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/mining"
	"web3_gui/chain/rpc"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

var Area *libp2parea.Area

func RegisterBaseMsg(area *libp2parea.Area) {
	Area = area
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_ACCOUNT_INFO, GetAccountInfoOfLightNode)     //获取账号信息接口
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_ACCOUNT_MSG, GetAccountMsgOfLightNode)       //获取账号详情接口
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_HISTORY_TX_RECORD, GetTxHistoryRecord)       //获取交易历史记录
	Area.Register_p2p(config.MSGID_LIGHT_NODE_CHECK_TX_ONLINE, CheckTxIsOnline)                //查询交易上链
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_BLOCK_INFO, FindBlock)                       //获取区块详情by height
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_BY_HASH, FindBlockByHash)         //获取区块详情by hash
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO, FindRangeBlock)            //获取范围区块详情
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_PROTO, FindRangeBlockProto) //获取范围区块详情proto
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_BLOCK_DEAL_TX_TIME, GetBlockTime)            //查询区块处理交易时间
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_DEPOSIT_NUM, GetDepositNum)                  //获取节点质押数量
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_LIGHT_LIST, GetLightList)                    //获取所有轻节点列表
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GET_NODE_NUM, GetNodeNum)                        //获取节点数量
	Area.Register_p2p(config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS, GetRoleAddr)               //查询一个地址角色
	Area.Register_p2p(config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS, GetAccountOther)        //查询一个地址余额

	Area.Register_p2p(config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS, HandleGetNewAddress)   //创建新地址
	Area.Register_p2p(config.MSGID_LIGHT_NODE_HANDLELISTALLADDRESS, HandleListAllAddress) //列出节点私钥存储的所有地址及余额
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GETBALANCEOFADDR, GetBalanceOfAddr)         //获取地址余额
	Area.Register_p2p(config.MSGID_LIGHT_NODE_SENDTOADDRESS, SendToAddress)               //转账
	Area.Register_p2p(config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE, SendToAddressMore)       //批量转账
	Area.Register_p2p(config.MSGID_LIGHT_NODE_GETNONCE, GetNonce)                         //获取某个地址的nonce
}

func HandleGetNewAddress(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	if !rj.VerifyType("password", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(model.TypeWrong, "password"))
		return
	}
	password, ok := rj.Get("password")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(model.NoField, "password"))
		return
	}

	addr, err := config.Area.Keystore.GetNewAddr(password.(string), password.(string))
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(model.FailPwd, ""))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(rpc.SystemError, ""))
		return
	}
	getnewadress := model.GetNewAddress{Address: addr.B58String()}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLEGETNEWADDRESS_REV, pkg(model.Success, getnewadress))
}

// 列出节点私钥存储的所有地址及余额
func HandleListAllAddress(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLELISTALLADDRESS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	tokenidStr := ""
	tokenidItr, ok := rj.Get("token_id")
	if ok {
		if !rj.VerifyType("token_id", "string") {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLELISTALLADDRESS_REV, pkg(model.TypeWrong, "token_id"))
			return
		}
		tokenidStr = tokenidItr.(string)
	}
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10000)
	}
	pageSizeInt := int(pageSize.(float64))
	vos := make([]rpc.AccountVO, 0)
	list, _ := rj.Get("addr")
	list1 := list.([]interface{})
	total := len(list1)
	start := (pageInt - 1) * pageSizeInt
	end := start + pageSizeInt
	if start > total {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLELISTALLADDRESS_REV, pkg(model.Success, vos))
		return
	}
	if end > total {
		end = total
	}

	for i, val := range list1[start:end] {
		var ba, fba, baLockup uint64
		val1 := MapChangeAddressInfo(val)
		if tokenidStr == "" {
			ba, fba, baLockup = mining.GetBalanceForAddrSelf(val1.Addr)
		} else {

		}
		vo := rpc.AccountVO{
			Index:       i + start,
			AddrCoin:    val1.GetAddrStr(),
			Type:        mining.GetAddrState(val1.Addr),
			Value:       ba,       //可用余额
			ValueFrozen: fba,      //冻结余额
			ValueLockup: baLockup, //
		}
		vos = append(vos, vo)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_HANDLELISTALLADDRESS_REV, pkg(model.Success, vos))
}

func GetBalanceOfAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETBALANCEOFADDR_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETBALANCEOFADDR_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETBALANCEOFADDR_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETBALANCEOFADDR_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	value, valueFrozen, _ := mining.GetBalanceForAddrSelf(addrCoin)
	// fmt.Println(addr)
	getaccount := model.GetAccount{
		Balance:       value,
		BalanceFrozen: valueFrozen,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETBALANCEOFADDR_REV, pkg(model.Success, getaccount))
}

func SendToAddress(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	if mining.CheckOutOfMemory() {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.Timeout, ""))
		return
	}

	config.GetRpcRate(rj.Method, true)

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
		}
	}
	addrItr, ok = rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)

	dst := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.NoField, "amount"))
		return
	}
	amount := uint64(amountItr.(float64))
	if amount <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.AmountIsZero, "amount"))
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.NoField, "gas"))
		return
	}
	gas := uint64(gasItr.(float64))

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.CommentOverLengthMax, "comment"))
		return
	}

	temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
	temp = new(big.Int).Div(temp, big.NewInt(1024))
	if gas < temp.Uint64() {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.GasTooLittle, "gas"))
		return
	}
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas)
	if total < amount+gas {
		//资金不够
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.BalanceNotEnough, ""))
		return
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := uint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = uint64(domainTypeItr.(float64))
	}

	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(src.B58String(), domain, dst.B58String(), new(big.Int).SetUint64(domainType)) {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.Nomarl, "domain name resolution failed"))
			return
		}
	}

	txpay, err := mining.SendToAddress(&src, &dst, amount, gas, frozenHeight, pwd, comment, domain, domainType)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.FailPwd, ""))
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(rpc.AmountIsZero, "amount"))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.Nomarl, err.Error()))
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESS_REV, pkg(model.Success, result))
}

func SendToAddressMore(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var src crypto.AddressCoin
	srcAddrStr := ""
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcAddrStr = addrItr.(string)
		if srcAddrStr != "" {
			src = crypto.AddressFromB58String(srcAddrStr)
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
		}
	}

	addrItr, ok = rj.Get("addresses")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.NoField, "addresses"))
		return
	}

	bs, err := json.Marshal(addrItr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.TypeWrong, "addresses"))
		return
	}

	addrs := make([]rpc.PayNumber, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&addrs)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.TypeWrong, "addresses"))
		return
	}
	//给多人转账，可是没有地址
	if len(addrs) <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.NoField, "addresses"))
		return
	}

	amount := uint64(0)

	addr := make([]mining.PayNumber, 0)
	for _, one := range addrs {
		dst := crypto.AddressFromB58String(one.Address)
		//验证地址前缀
		if !crypto.ValidAddr(config.AddrPre, dst) {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(rpc.ContentIncorrectFormat, "addresses"))
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
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.NoField, "gas"))
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount+gas > value {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(rpc.BalanceNotEnough, ""))
		return
	}

	txpay, err := mining.SendToMoreAddress(&src, addr, gas, pwd, comment)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.FailPwd, ""))
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(rpc.AmountIsZero, "amount"))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result, err := utils.ChangeMap(txpay)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDTOADDRESSMORE_REV, pkg(model.Success, result))
}

func GetNonce(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETNONCE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //地址
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETNONCE_REV, pkg(model.NoField, "address"))
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}
	if addr == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETNONCE_REV, pkg(model.NoField, "address"))
		return
	}

	nonceInt, e := mining.GetAddrNonce(addr)
	if e != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETNONCE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	out := make(map[string]interface{})
	out["nonce"] = nonceInt.Uint64()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETNONCE_REV, pkg(model.Success, out))
}

/*
轻节点获取账号信息
*/
func GetAccountInfoOfLightNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	/*
		var param = make([]*keystore.AddressInfo, 0)
		err := json.Unmarshal(*message.Body.Content, &param)
		if err != nil {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_ACCOUNT_INFO_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
			return
		}
		chain := mining.GetLongChain()
		var notspend, frozen, lock uint64
		for _, one := range param {
			n, f, l := mining.GetBalanceForAddrSelf(one.Addr)
			notspend += n
			frozen += f
			lock += l
		}
		tbs := token.GetAllTokenInfos()
		tbVOs := make([]token.TokenBalanceVO, 0)
		for _, one := range tbs {
			tbVO := token.TokenBalanceVO{
				TokenId: hex.EncodeToString([]byte(one.Txid)),
				Name:    one.Name,
				Symbol:  one.Symbol,
				Supply:  one.Supply.Text(10),
				//Balance:       one.Balance.Text(10),
				//BalanceFrozen: one.BalanceFrozen.Text(10),
				//BalanceLockup: one.BalanceLockup.Text(10),
			}
			tbVOs = append(tbVOs, tbVO)
		}

		currentBlock := uint64(0)
		startBlock := uint64(0)
		heightBlock := uint64(0)
		pulledStates := uint64(0)
		startBlockTime := uint64(0)

		if chain != nil {
			currentBlock = chain.GetCurrentBlock()
			startBlock = chain.GetStartingBlock()
			heightBlock = mining.GetHighestBlock()
			pulledStates = chain.GetPulledStates()
			startBlockTime = chain.GetStartBlockTime()
		}

		info := model.Getinfo{
			Netid:          []byte(config.AddrPre),   //
			TotalAmount:    config.Mining_coin_total, //
			Balance:        notspend,                 //
			BalanceFrozen:  frozen,                   //
			BalanceLockup:  lock,                     //
			Testnet:        true,                     //
			Blocks:         currentBlock,             //
			Group:          0,                        //
			StartingBlock:  startBlock,               //区块开始高度
			StartBlockTime: startBlockTime,           //
			HighestBlock:   heightBlock,              //所链接的节点的最高高度
			CurrentBlock:   currentBlock,             //已经同步到的区块高度
			PulledStates:   pulledStates,             //正在同步的区块高度
			//BlockTime:      config.Mining_block_time,       //出块时间
			BlockTime:      uint64(config.Mining_block_time.Nanoseconds()), //出块时间 pl time
			LightNode:      config.Mining_light_min,                        //轻节点押金数量
			CommunityNode:  config.Mining_vote,                             //社区节点押金数量
			WitnessNode:    config.Mining_deposit,                          //见证人押金数量
			NameDepositMin: config.Mining_name_deposit_min,                 //
			AddrPre:        config.AddrPre,                                 //
			//TokenBalance:   tbVOs,                                          //
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_ACCOUNT_INFO_REV, pkg(model.Success, info))
	*/
}

/*
轻节点获取账号详情
*/
type AccountV1 struct {
	AddrCoin             string //收款地址
	Value                uint64 //可用余额
	ValueFrozen          uint64 //冻结余额
	ValueLockup          uint64 //
	ValueFrozenWitness   uint64 //见证人节点冻结奖励
	ValueFrozenCommunity uint64 //社区节点冻结奖励
	ValueFrozenLight     uint64 //轻节点冻结奖励
	DepositIn            uint64 //质押数量
	Type                 int    //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

func GetAccountMsgOfLightNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	addrStr := utils.Bytes2string(*message.Body.Content)
	addr := crypto.AddressFromB58String(addrStr)
	ba, fba, baLockup := mining.GetBalanceForAddrSelf(addr)

	depositIn := uint64(0)
	//社区地址
	var cAddr crypto.AddressCoin
	ty := mining.GetAddrState(addr)
	chain := mining.GetLongChain().GetBalance()

	switch ty {
	case 1:
		depositIn = config.Mining_deposit
	case 2:
		depositIn = config.Mining_vote
		cAddr = addr
	case 3:
		depositIn = config.Mining_light_min
		light := precompiled.GetLightList([]crypto.AddressCoin{addr})
		if len(light) > 0 && light[0].Score.Uint64() != 0 {
			if light[0].C != evmutils.ZeroAddress {
				cAddr = evmutils.AddressToAddressCoin(light[0].C.Bytes())
			}
		}
	}
	var wValue, cValue, lValue uint64 = 0, 0, 0
	if cAddr != nil {
		cRewardPool := precompiled.GetCommunityRewardPool(cAddr)
		if cRewardPool.Cmp(big.NewInt(0)) > 0 {
			cRate, err := precompiled.GetRewardRatio(cAddr)
			if err != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_ACCOUNT_MSG_REV, pkg(rpc.SystemError, "address"))
				return
			}

			lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
			cBigValue := new(big.Int).Sub(cRewardPool, lTotal)
			cValue = cBigValue.Uint64()
			if depositIn == config.Mining_light_min {
				community := precompiled.GetCommunityList([]crypto.AddressCoin{cAddr})
				if len(community) == 0 {
					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_ACCOUNT_MSG_REV, pkg(rpc.SystemError, "address"))
					return
				}
				//当前轻节点身份，不需要查出社区身份时候的未体现余额，新需求取消质押会提现所有奖励池
				cValue = 0
				v := new(big.Int).Mul(new(big.Int).SetUint64(chain.GetDepositVote(&addr).Value), new(big.Int).SetInt64(1e8))
				ratio := new(big.Int).Quo(v, community[0].Vote)
				lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
				lValue = lBigValue.Uint64()
			}
		}
	}

	vo := AccountV1{
		AddrCoin:    addrStr,
		Type:        mining.GetAddrState(addr),
		Value:       ba,       //可用余额
		ValueFrozen: fba,      //冻结余额
		ValueLockup: baLockup, //
		//ValueFrozenWitness: precompiled.GetMyWitFrozenReward(addr),
		//ValueFrozenCommunity: precompiled.GetMyCommunityFrozenReward(addr),
		//ValueFrozenLight:     precompiled.GetMyLightFrozenReward(addr),
		ValueFrozenWitness:   wValue,
		ValueFrozenCommunity: cValue,
		ValueFrozenLight:     lValue,
		DepositIn:            depositIn,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_ACCOUNT_MSG_REV, pkg(model.Success, vo))
}

/*
获取交易历史记录
*/
func GetTxHistoryRecord(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		ID    *big.Int
		Total int
	}
	p := new(param)
	err := json.Unmarshal(*message.Body.Content, p)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_HISTORY_TX_RECORD_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	//如果是见证人，需要时间间隔控制
	if ok, _, _, _, _ := mining.GetWitnessStatus(); ok {
		utils.SetTimeToken(config.TIMETOKEN_GetTransactionHistoty, time.Second*5)
		if allow := utils.GetTimeToken(config.TIMETOKEN_GetTransactionHistoty, false); !allow {
			return
		}
	}

	hivos := make([]rpc.HistoryItemVO, 0)
	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_HISTORY_TX_RECORD_REV, pkg(model.Success, hivos))
		return
	}
	his := chain.GetHistoryBalance(p.ID, p.Total)
	for _, one := range his {
		hivo := rpc.HistoryItemVO{
			GenerateId: one.GenerateId.String(),      //
			IsIn:       one.IsIn,                     //资金转入转出方向，true=转入;false=转出;
			Type:       one.Type,                     //交易类型
			InAddr:     make([]string, 0),            //输入地址
			OutAddr:    make([]string, 0),            //输出地址
			Value:      one.Value,                    //交易金额
			Txid:       hex.EncodeToString(one.Txid), //交易id
			Height:     one.Height,                   //区块高度
			Payload:    string(one.Payload),          //
		}

		for _, two := range one.InAddr {
			hivo.InAddr = append(hivo.InAddr, two.B58String())
		}

		for _, two := range one.OutAddr {
			hivo.OutAddr = append(hivo.OutAddr, two.B58String())
		}

		hivos = append(hivos, hivo)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_HISTORY_TX_RECORD_REV, pkg(model.Success, hivos))
}

/*
检查交易是否上链
*/
func CheckTxIsOnline(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		TxID    []byte
		IsType1 bool
	}
	p := &param{}
	err := json.Unmarshal(*message.Body.Content, p)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECK_TX_ONLINE_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	outMap := make(map[string]interface{})
	txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(p.TxID)
	var blockHashStr string
	if blockHash != nil {
		blockHashStr = hex.EncodeToString(*blockHash)
	}
	tx, err := mining.LoadTxBase(p.TxID)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECK_TX_ONLINE_REV, pkg(model.TypeWrong, ""))
		return
	}
	item := rpc.JsonMethod(txItr)
	txClass, vins, vouts := rpc.DealTxInfoV2(tx)
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
	if p.IsType1 {
		returnVouts := []mining.VoutVO{}
		vouts := tx.GetVout()
		outs := *vouts
		returnVouts = append(returnVouts, mining.VoutVO{Address: precompiled.RewardContract.B58String(), Value: outs[0].Value})
		item["type"] = config.Wallet_tx_type_mining
		item["vout"] = returnVouts
		item["vout_total"] = 1
		item["vin"] = []mining.VinVO{}
		item["vin_total"] = 0
		item["hash"] = common.Bytes2Hex(*tx.GetHash()) + "-1"
	}
	outMap["txinfo"] = item
	outMap["upchaincode"] = code
	outMap["blockheight"] = blockHeight
	outMap["blockhash"] = blockHashStr
	outMap["timestamp"] = timestamp
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECK_TX_ONLINE_REV, pkg(model.Success, outMap))
}

/*
根据区块高度查询区块信息
*/
func FindBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	h := binary.BigEndian.Uint64(*message.Body.Content)
	bh := mining.LoadBlockHeadByHeight(h)
	if bh == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_REV, pkg(model.NotExist, ""))
		return
	}
	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	txs := make([]string, 0)
	gas := uint64(0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
			//reward += txItr.GetGas()
		}
	}

	reward += gas
	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	bhvo := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		//Gas:               reward - uint64(config.BLOCK_TOTAL_REWARD),
		Gas:     gas,
		Destroy: 0,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_REV, pkg(model.Success, bhvo))
}

func FindBlockByHash(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bh, err := mining.LoadBlockHeadByHash(message.Body.Content)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_BY_HASH_REV, pkg(rpc.SysTemError, ""))
		return
	}
	if bh == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_BY_HASH_REV, pkg(model.NotExist, ""))
		return
	}
	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	gas := uint64(0)
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
		}
	}

	reward += gas
	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	bhvo := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		Gas:               gas,
		Destroy:           0,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_INFO_BY_HASH_REV, pkg(model.Success, bhvo))
}

func FindRangeBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		Begin uint64
		End   uint64
	}
	p := &param{}
	err := json.Unmarshal(*message.Body.Content, p)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
	}
	//待返回的区块
	bhvos := make([]BlockHeadOut, 0, p.End-p.Begin+1)
	for i := p.Begin; i <= p.End; i++ {
		bhvo := BlockHeadOut{}
		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItrJson, code, txItr := mining.FindTxJsonVoV1(one)
			if txItr == nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_REV, pkg(rpc.Tx_not_exist, ""))
				return
			}
			item := rpc.JsonMethod(txItrJson)
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
			//如果是27，
			if txClass != config.Wallet_tx_type_reward_W {
				bhvo.Txs = append(bhvo.Txs, item)
			} else if txClass > 0 {
				if len(item["vout"].([]mining.VoutVO)) > 0 {
					bhvo.Txs = append(bhvo.Txs, item)
				}
			}
			//
			if txItr.Class() == config.Wallet_tx_type_mining {
				newiTem := make(map[string]interface{})
				for k, v := range item {
					newiTem[k] = v
				}
				returnVouts := []mining.VoutVO{}
				vouts := txItr.GetVout()
				outs := *vouts
				if bytes.Equal(outs[0].Address, precompiled.RewardContract) {
					returnVouts = append(returnVouts, mining.VoutVO{Address: precompiled.RewardContract.B58String(), Value: outs[0].Value})
					newiTem["type"] = config.Wallet_tx_type_mining
					newiTem["vout"] = returnVouts
					newiTem["vout_total"] = 1
					newiTem["vin"] = []mining.VinVO{}
					newiTem["vin_total"] = 0
					newiTem["hash"] = common.Bytes2Hex(*txItr.GetHash()) + "-1"
					newTx := append([]map[string]interface{}{newiTem}, bhvo.Txs...)
					bhvo.Txs = newTx
				}

			}

		}

		bhvos = append(bhvos, bhvo)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_REV, pkg(model.Success, bhvos))
}

func FindRangeBlockProto(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		Begin uint64
		End   uint64
	}
	p := &param{}
	err := json.Unmarshal(*message.Body.Content, p)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	//待返回的区块
	bhvos := make([]*[]byte, 0, p.End-p.Begin+1)
	for i := p.Begin; i <= p.End; i++ {
		bhvo := mining.BlockHeadVO{}
		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItr, e := mining.LoadTxBase(one)
			if e != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_PROTO_REV, pkg(rpc.SystemError, ""))
				return
			}
			bhvo.Txs = append(bhvo.Txs, txItr)
		}
		bs, e := bhvo.Proto()
		if e != nil {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_PROTO_REV, pkg(rpc.SystemError, ""))
			return
		}
		bhvos = append(bhvos, bs)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_RANGE_INFO_PROTO_REV, pkg(model.Success, bhvos))
	return
}

func GetBlockTime(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		StartHeight uint64
		EndHeight   uint64
		Ty          uint64
	}
	p := &param{}
	err := json.Unmarshal(*message.Body.Content, p)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_DEAL_TX_TIME_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	if p.StartHeight == 0 {
		p.StartHeight = 1
	}

	if p.EndHeight == 0 {
		p.EndHeight = mining.GetHighestBlock()
	}

	type Blocktime struct {
		Height uint64
		Time   string
	}
	var max time.Duration
	var avg time.Duration
	type SaveBlockTime struct {
		Start  uint64
		End    uint64
		Max    string
		Avg    string
		Bigs   []Blocktime
		Blocks []Blocktime
	}
	bigs := make([]Blocktime, 0)
	blocks := make([]Blocktime, 0)
	var total time.Duration
	for i := p.StartHeight; i <= p.EndHeight; i++ {
		v := mining.SaveBlockTime[i]
		vt, err := time.ParseDuration(v)
		if err != nil {
			continue
		}

		if vt > max {
			max = vt
		}
		total += vt
		if p.Ty != 0 {
			blocks = append(blocks, Blocktime{
				i, vt.String(),
			})
		}

		if vt.Seconds() > 1 {
			bigs = append(bigs, Blocktime{
				i, vt.String(),
			})
		}
	}
	avgt := total.Nanoseconds() / int64(p.EndHeight-p.StartHeight)
	avg = time.Duration(avgt)
	resultbs := SaveBlockTime{
		p.StartHeight, p.EndHeight,
		max.String(), avg.String(), bigs, blocks,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_BLOCK_DEAL_TX_TIME_REV, pkg(model.Success, resultbs))
}

func GetDepositNum(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var addrs = make([]crypto.AddressCoin, 0)
	err := json.Unmarshal(*message.Body.Content, &addrs)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_DEPOSIT_NUM_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	witCount := uint64(0)
	communityCount := uint64(0)
	lightCount := uint64(0)
	for _, addr := range addrs {
		role := mining.GetAddrState(addr)
		switch role {
		case 1: // 见证人
			witCount++
		case 2: // 社区
			communityCount++
		case 3: // 轻节点
			lightCount++
		}
	}
	info := map[string]interface{}{
		"deposit": witCount*config.Mining_deposit + communityCount*config.Mining_vote + lightCount*config.Mining_light_min,
		// "total":   depositAmount,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_DEPOSIT_NUM_REV, pkg(model.Success, info))
}

func GetLightList(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	type param struct {
		Page     int
		PageSize int
	}
	po := &param{}
	err := json.Unmarshal(*message.Body.Content, &po)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_LIGHT_LIST_REV, pkg(rpc.JSON_UNMARSHAL_ERROR, ""))
		return
	}
	vss := mining.GetLightListSortNew(po.Page, po.PageSize)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_LIGHT_LIST_REV, pkg(model.Success, vss))
}

func GetNodeNum(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	wbg := mining.GetWitnessListSort()
	count := len(wbg.Witnesses)
	count1 := len(wbg.WitnessBackup)

	count2, count3 := precompiled.GetRoleTotal(config.Area.Keystore.GetCoinbase().Addr)

	data := make(map[string]interface{})
	data["wit_num"] = count
	data["back_wit_num"] = count1
	data["community_num"] = count2
	data["light_num"] = count3
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GET_NODE_NUM_REV, pkg(model.Success, data))
}

func GetRoleAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	out := make(map[string]interface{})
	out["type"] = mining.GetAddrState(addrCoin)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_ROLE_OF_ADDRESS_REV, pkg(model.Success, out))
}

func GetAccountOther(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	value, valueFrozen, _ := mining.GetNotspendByAddrOther(mining.GetLongChain(), addrCoin)
	// fmt.Println(addr)
	getaccount := model.GetAccount{
		Balance:       value,
		BalanceFrozen: valueFrozen,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_BALANCE_OF_ADDRESS_REV, pkg(model.Success, getaccount))
}
