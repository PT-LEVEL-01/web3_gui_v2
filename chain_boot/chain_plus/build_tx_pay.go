package chain_plus

import (
	"golang.org/x/crypto/ed25519"
	"math/big"
	"strconv"
	chainconfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	chainbootconfig "web3_gui/chain_boot/config"
	gconfig "web3_gui/config"
	ka "web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	keystconfig "web3_gui/keystore/v2/config"
	engineconfig "web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func GetKeystore() (*keystore.Keystore, utils.ERROR) {
	keystAapter, ok := gconfig.Node.Keystore.(*ka.Keystore)
	if ok {
		keys, ERR := keystAapter.GetV2Keystore()
		if ERR.CheckFail() {
			return keys, ERR
		}
		return keys, ERR
	} else {
		keys := gconfig.Node.Keystore.(*keystore.Keystore)
		return keys, utils.NewErrorSuccess()
	}
}

/*
创建一个转款交易
*/
func CreateTxPay(srcAddress, address *coin_address.AddressCoin, amount, gas, frozenHeight uint64, pwd string, comment []byte,
	domain string, domainType uint64) (*mining.Tx_Pay, utils.ERROR) {
	//src := crypto.AddressCoin(*srcAddress)
	//dst := crypto.AddressCoin(*address)
	//txpay, err := mining.CreateTxPay(nil, &dst, amount, gas, frozenHeight, pwd, "", domain, domainType)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//return txpay, utils.NewErrorSuccess()

	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}
	currentHeight := chain.GetCurrentBlock()

	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}

	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}

	//查找余额
	vins := make([]*mining.Vin, 0)
	var srcAddr *crypto.AddressCoin
	if srcAddress != nil && len(*srcAddress) > 0 {
		addrInfo := keyst.FindCoinAddr(srcAddress)
		if addrInfo == nil {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_address_not_exist, "srcAddress")
		}
		srcAddrTemp := crypto.AddressCoin(*srcAddress)
		srcAddr = &srcAddrTemp
	}

	total, item := chain.Balance.BuildPayVinNew(srcAddr, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
	}
	addr := coin_address.AddressCoin(*item.Addr)
	addrInfo := keyst.FindCoinAddr(&addr)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//puk := keyPair.GetPublicKey()

	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := mining.Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk[:], //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	vout := mining.Vout{
		Value:        amount,                       //输出金额 = 实际金额 * 100000000
		Address:      crypto.AddressCoin(*address), //钱包地址
		FrozenHeight: frozenHeight,                 //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *mining.Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := mining.TxBase{
			Type:       chainconfig.Wallet_tx_type_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                                    //输入交易数量
			Vin:        vins,                                                 //交易输入
			Vout_total: uint64(len(vouts)),                                   //输出交易数量
			Vout:       vouts,                                                //交易输出
			Gas:        gas,                                                  //交易手续费
			LockHeight: currentHeight + chainconfig.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                             //
			Comment:    comment,
		}
		pay = &mining.Tx_Pay{
			TxBase: base,
		}
		//给输出签名，防篡改
		for i, one := range pay.Vin {
			addrCoin, ERR := keystore.BuildAddr(chainconfig.AddrPre, one.Puk)
			if ERR.CheckFail() {
				return nil, ERR
			}
			addrInfo := keyst.FindCoinAddr(&addrCoin)
			_, prkBs, ERR := addrInfo.DecryptPrkPuk(pwd)
			if ERR.CheckFail() {
				return nil, ERR
			}
			prk := ed25519.PrivateKey(prkBs[:])
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
			sign := pay.GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}
		// engine.Log.Info("给输出签名 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		pay.BuildHash()
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	//utils.Log.Info().Interface("支付交易", pay).Send()
	chain.Balance.AddLockTx(pay)
	return pay, utils.NewErrorSuccess()
}

/*
创建一个转款交易
*/
func CreateTxPayMore(srcAddress coin_address.AddressCoin, address []coin_address.AddressCoin, amount, frozenHeight []uint64,
	domain []string, domainType []uint64, gas uint64, pwd string, comment []byte) (*mining.Tx_Pay, utils.ERROR) {
	if len(address) != len(amount) {
		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "amount")
	}
	if len(address) != len(frozenHeight) {
		return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "frozenHeight")
	}
	if len(domain) > 0 {
		if len(address) != len(domain) {
			return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "domain")
		}
	} else {
		domain = make([]string, len(address))
	}
	if len(domainType) > 0 {
		if len(address) != len(domainType) {
			return nil, utils.NewErrorBus(engineconfig.ERROR_code_rpc_param_length_fail, "domainType")
		}
	} else {
		domainType = make([]uint64, len(address))
	}

	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}

	//src := crypto.AddressCoin(*srcAddress)
	//dst := crypto.AddressCoin(*address)
	//txpay, err := mining.CreateTxPay(nil, &dst, amount, gas, frozenHeight, pwd, "", domain, domainType)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//return txpay, utils.NewErrorSuccess()
	//keystAapter := config.Node.Keystore.(*ka.Keystore)
	keys, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//keys := config.Node.Keystore.(*keystore.Keystore)
	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}
	currentHeight := chain.GetCurrentBlock()
	//查找余额
	vins := make([]*mining.Vin, 0)
	var srcAddr *crypto.AddressCoin
	if srcAddress != nil && len(srcAddress) > 0 {
		addrInfo := keys.FindCoinAddr(&srcAddress)
		if addrInfo == nil {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_address_not_exist, "srcAddress")
		}
		srcAddrTemp := crypto.AddressCoin(srcAddress)
		srcAddr = &srcAddrTemp
	}

	amountTotal := gas
	for _, one := range amount {
		amountTotal += one
	}

	total, item := chain.Balance.BuildPayVinNew(srcAddr, amountTotal)
	if total < amountTotal {
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
	}
	addr := coin_address.AddressCoin(*item.Addr)
	addrInfo := keys.FindCoinAddr(&addr)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//puk := keyPair.GetPublicKey()

	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := mining.Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk[:], //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	for i, one := range address {
		vout := mining.Vout{
			Value:        amount[i],               //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressCoin(one), //钱包地址
			FrozenHeight: frozenHeight[i],         //
			Domain:       []byte(domain[i]),
			DomainType:   domainType[i],
		}
		vouts = append(vouts, &vout)
	}

	var pay *mining.Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := mining.TxBase{
			Type:       chainconfig.Wallet_tx_type_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                                    //输入交易数量
			Vin:        vins,                                                 //交易输入
			Vout_total: uint64(len(vouts)),                                   //输出交易数量
			Vout:       vouts,                                                //交易输出
			Gas:        gas,                                                  //交易手续费
			LockHeight: currentHeight + chainconfig.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                             //
			Comment:    comment,
		}
		pay = &mining.Tx_Pay{
			TxBase: base,
		}
		//给输出签名，防篡改
		for i, one := range pay.Vin {
			addrCoin, ERR := keystore.BuildAddr(chainconfig.AddrPre, one.Puk)
			if ERR.CheckFail() {
				return nil, ERR
			}
			addrInfo := keys.FindCoinAddr(&addrCoin)
			_, prkBs, ERR := addrInfo.DecryptPrkPuk(pwd)
			if ERR.CheckFail() {
				return nil, ERR
			}
			prk := ed25519.PrivateKey(prkBs[:])
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
			sign := pay.GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}
		// engine.Log.Info("给输出签名 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		pay.BuildHash()
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	//utils.Log.Info().Interface("支付交易", pay).Send()
	chain.Balance.AddLockTx(pay)
	return pay, utils.NewErrorSuccess()
}

/*
创建交易基本信息
*/
func NewTxBase(txType uint64, vins *[]*mining.Vin, vouts *[]*mining.Vout, gas, currentHeight, itr uint64) *mining.TxBase {
	base := mining.TxBase{
		Type:       txType,                                                 //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  uint64(len(*vins)),                                     //输入交易数量
		Vin:        *vins,                                                  //交易输入
		Vout_total: uint64(len(*vouts)),                                    //
		Vout:       *vouts,                                                 //
		Gas:        gas,                                                    //交易手续费
		LockHeight: currentHeight + chainconfig.Wallet_tx_lockHeight + itr, //锁定高度
		// CreateTime: config.TimeNow().Unix(),                //创建时间
		//Payload: []byte(nickname),
		//Comment: []byte{},
	}
	return &base
}

/*
给交易签名
*/
func SignTxVin(tx mining.TxItr, keyst *keystore.Keystore, pwd string) utils.ERROR {
	//给输出签名，防篡改
	for i, one := range *tx.GetVin() {
		addrCoin, ERR := keystore.BuildAddr(chainconfig.AddrPre, one.Puk)
		if ERR.CheckFail() {
			return ERR
		}
		addrInfo := keyst.FindCoinAddr(&addrCoin)
		_, prkBs, ERR := addrInfo.DecryptPrkPuk(pwd)
		if ERR.CheckFail() {
			return ERR
		}
		prk := ed25519.PrivateKey(prkBs[:])
		// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		sign := tx.GetSign(&prk, uint64(i))
		(*tx.GetVin())[i].Sign = *sign
	}
	return utils.NewErrorSuccess()
}

/*
把交易加入缓存并且广播出去
*/
func CurrentLimitingAndMulticastTx(txItr mining.TxItr) utils.ERROR {
	chain := mining.GetLongChain()
	if err := chain.TransactionManager.AddTx(txItr); err != nil {
		chain.Balance.DelLockTx(txItr)
		return utils.NewErrorSysSelf(err) // errors.Wrap(err, "add tx fail!")
	}
	mining.MulticastTx(txItr)
	return utils.NewErrorSuccess()
}

/*
创建一个见证人押金交易
@amount    uint64    押金额度
*/
func CreateTxWitnessDepositIn(gas uint64, pwd, nickname string, rate uint16) (*mining.Tx_deposit_in, utils.ERROR) {
	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}
	//判断押金已经缴纳
	if chain.GetBalance().GetDepositIn() != nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_witness_deposit_exist, "")
	}
	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}
	// engine.Log.Debug("创建见证人押金交易 111")
	amount := chainconfig.Mining_deposit
	//if amount != chainconfig.Mining_deposit {
	//	//押金太少
	//	return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_witness_deposit_quantity_incorrect, strconv.Itoa(int(chainconfig.Mining_deposit)))
	//}
	//chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()

	//keystAapter := config.Node.Keystore.(*ka.Keystore)
	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//keyst := config.Node.Keystore.(*keystore.Keystore)
	//key := Area.Keystore.GetCoinbase()

	coinbase := keyst.GetCoinAddrAll()[0]
	//验证密码
	puk, _, ERR := coinbase.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	witnessAddr := coinbase.Addr.(*coin_address.AddressCoin)
	witnessAddrTemp := crypto.AddressCoin(*witnessAddr)

	//查找余额
	total, item := chain.Balance.BuildPayVinNew(&witnessAddrTemp, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, witnessAddr.String())
	}

	nonce := chain.GetBalance().FindNonce(item.Addr)
	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := mining.Vout{
		Value:   amount,          //输出金额 = 实际金额 * 100000000
		Address: witnessAddrTemp, //钱包地址
	}
	vouts = append(vouts, &vout)

	var txin *mining.Tx_deposit_in
	for i := uint64(0); i < 10000; i++ {
		//
		base := NewTxBase(chainconfig.Wallet_tx_type_deposit_in, &vins, &vouts, gas, currentHeight, i)
		base.Payload = []byte(nickname)
		txin = &mining.Tx_deposit_in{
			TxBase: *base,
			Puk:    puk,
			Rate:   rate,
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
	// chain.Balance.Frozen(items, txin)
	chain.Balance.AddLockTx(txin)
	return txin, utils.NewErrorSuccess()
}

/*
创建一个退还押金交易
额度超过了押金额度，那么会从自己账户余额转账到目标账户（因为考虑到押金太少还不够给手续费的情况）
@addr      *utils.Multihash    退回到的目标账户地址
@amount    uint64              押金额度
*/
func CreateTxWitnessDepositOut(gas uint64, pwd string) (*mining.Tx_deposit_out, utils.ERROR) {
	chain := mining.GetLongChain()
	if chain == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
	}
	//判断押金已经缴纳
	txDepositIn := chain.GetBalance().GetDepositIn()
	if txDepositIn == nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_witness_deposit_not_exist, "")
	}
	//判断矿工费是否足够
	if gas < chainconfig.Wallet_tx_gas_min {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	}
	// engine.Log.Debug("创建见证人押金交易 111")
	amount := txDepositIn.Value
	currentHeight := chain.GetCurrentBlock()

	vins := make([]*mining.Vin, 0)
	//查看余额够不够
	total, item := chain.Balance.BuildPayVinNew(txDepositIn.Addr, gas)
	if total < gas {
		//资金不够
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, txDepositIn.Addr.B58String())
	}
	//keystAapter := config.Node.Keystore.(*ka.Keystore)
	keyst, ERR := GetKeystore()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//keyst := config.Node.Keystore.(*keystore.Keystore)
	newAddr := coin_address.AddressCoin(*txDepositIn.Addr)
	addrInfo := keyst.FindCoinAddr(&newAddr)
	//验证密码
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := mining.Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := mining.Vout{
		Value:   amount,            //输出金额 = 实际金额 * 100000000
		Address: *txDepositIn.Addr, //钱包地址
	}
	vouts = append(vouts, &vout)

	var txin *mining.Tx_deposit_out
	for i := uint64(0); i < 10000; i++ {
		//
		base := NewTxBase(chainconfig.Wallet_tx_type_deposit_out, &vins, &vouts, gas, currentHeight, i)
		txin = &mining.Tx_deposit_out{
			TxBase: *base,
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
	chain.Balance.AddLockTx(txin)
	return txin, utils.NewErrorSuccess()
}

/*
创建一个转款交易，并离线签名
*/
func CreateTxPayByKey(keyst *keystore.Keystore, srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64,
	pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*mining.Tx_Pay, utils.ERROR) {

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	addr := coin_address.AddressCoin(*srcAddress)
	addrInfo := keyst.FindCoinAddr(&addr)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	//查找余额
	nonce := big.NewInt(int64(nonceInt))
	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	vout := mining.Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	base := NewTxBase(chainconfig.Wallet_tx_type_pay, &vins, &vouts, gas, currentHeight, 0)
	base.Comment = commentbs

	pay := &mining.Tx_Pay{
		TxBase: *base,
	}
	//给输出签名，防篡改
	ERR = SignTxVin(pay, keyst, pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	pay.BuildHash()

	return pay, utils.NewErrorSuccess()
}

/*
创建一个社区节点质押交易
*/
func CreateTxCommunityDepositIn(rate uint64, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, gas uint64,
	pwd, comment string) utils.ERROR {
	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= chainconfig.Wallet_vote_start_height {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_community_depositin_close, "")
		return ERR
	}
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteIn(mining.VOTE_TYPE_community, uint16(rate), witnessAddr, addr, chainconfig.Mining_vote, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
创建一个社区节点取消质押交易
*/
func CreateTxCommunityDepositOut(addr coin_address.AddressCoin, gas uint64, pwd, comment string) utils.ERROR {
	address := crypto.AddressCoin(addr)
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteOut(mining.VOTE_TYPE_community, address, 0, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
创建一个轻节点质押交易
*/
func CreateTxLightDepositIn(addr crypto.AddressCoin, gas uint64, pwd, comment string) utils.ERROR {
	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= chainconfig.Wallet_vote_start_height {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_light_depositin_close, "")
		return ERR
	}
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteIn(mining.VOTE_TYPE_light, 0, nil, addr, chainconfig.Mining_light_min, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
创建一个轻节点取消质押交易
*/
func CreateTxLightDepositOut(addr coin_address.AddressCoin, gas uint64, pwd, comment string) utils.ERROR {
	address := crypto.AddressCoin(addr)
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteOut(mining.VOTE_TYPE_light, address, 0, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
创建一个轻节点投票交易
*/
func CreateTxLightVoteIn(addr, communityAddr coin_address.AddressCoin, amount, gas uint64, pwd, comment string) utils.ERROR {
	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= chainconfig.Wallet_vote_start_height {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_light_depositin_close, "")
		return ERR
	}

	voter := crypto.AddressCoin(addr)
	voteTo := crypto.AddressCoin(communityAddr)
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteIn(mining.VOTE_TYPE_vote, 0, voteTo, voter, amount, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
创建一个轻节点取消投票交易
*/
func CreateTxLightVoteOut(addr coin_address.AddressCoin, amount, gas uint64, pwd, comment string) utils.ERROR {
	address := crypto.AddressCoin(addr)
	//缴纳备用见证人押金交易
	err := mining.GetLongChain().Balance.VoteOut(mining.VOTE_TYPE_vote, address, amount, gas, pwd, comment)
	if err != nil {
		if err.Error() == chainconfig.ERROR_password_fail.Error() {
			return utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == chainconfig.ERROR_not_enough.Error() {
			return utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}
