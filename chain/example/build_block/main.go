package main

import (
	jsoniter "github.com/json-iterator/go"
	"math/big"
	"os"
	"path/filepath"
	"time"
	chainConfig "web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/db/leveldb"
	"web3_gui/chain/example/build_block/build"
	"web3_gui/chain/example/build_block/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/tx_name_in"
	"web3_gui/chain/rpc/limiter"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	firstBlockTime int64 = 0
)

func init() {
	t, err := time.Parse("2006-01-02 15:04:05", "2024-09-25 15:35:51")
	if err != nil {
		panic(err)
	}
	firstBlockTime = t.Unix()
	firstBlockTime = time.Now().Unix() - 1000
}

func main() {
	logPath := filepath.Join(config.RootDir, "buildblocklog.txt")
	engine.SetLogPath(filepath.Join(config.RootDir, "log.txt"))
	os.Remove(logPath)
	//os.RemoveAll(filepath.Join(config.WalletDirPath, config.WalletDBName))
	utils.LogBuildDefaultFile(logPath)
	utils.LogZero.Info().Str("start", "").Send()
	chainConfig.AddrPre = config.AddrPre
	chainConfig.Wallet_keystore_default_pwd = config.KeyPwd
	//example_write()
	//example_read()
	//return
	ldb := build.BuildDB(config.Peer_total, config.WalletDirPath, config.WalletDBName)
	areas := build.BuildAreas(config.Peer_total, config.AreaName, config.AddrPre, config.KeyPwd, config.KeystoreDirPath)
	db.LevelDB = ldb
	db.LevelTempDB = ldb
	BuildBlock(areas, ldb)

}

func BuildBlock(areas []*libp2parea.Area, ldb *leveldb.LevelDB) {
	mining.Area = areas[0]

	// 订单薄引擎启动
	mining.OrderBookEngineSetup()

	limiter.InitRpcLimiter()
	events := BuildEvent(areas)

	//检查leveldb数据库中是否有创始区块
	bhvo := mining.LoadStartBlock()
	if bhvo == nil {
		var err error
		bhvo, err = build.BuildFirstBlock(areas[0], ldb, firstBlockTime)
		if err != nil {
			panic(err)
		}
		//utils.LogZero.Info().Interface("bhvo", bhvo).Send()
		ok, err := mining.SaveBlockHead(bhvo)
		if err != nil {
			panic(err)
		}
		utils.LogZero.Info().Bool("ok", ok).Send()
	}

	//ldb.PrintAll()

	//build.PrintBHVO(bhvo)

	//find_block_print.FindNextBlock()
	chain := build.BuildChain(bhvo)
	build.LoopLoadBlock(chain, bhvo)
	utils.LogZero.Info().Uint64("加载完成", chain.GetCurrentBlock()).Send()
	build.LoopBuildBlock(chain, areas, events)
	utils.LogZero.Info().Str("构建完成", "").Send()

	//查询余额
	for _, one := range areas {
		b1, b2, b3 := mining.GetNotspendByAddrOther(chain, one.Keystore.GetCoinbase().Addr)
		utils.LogZero.Info().Str("地址", one.Keystore.GetCoinbase().Addr.B58String()).Uint64("b1", b1).Uint64("b2", b2).Uint64("b3", b3).Send()
	}

}

func BuildEvent(areas []*libp2parea.Area) []*build.Event {
	events := make([]*build.Event, 0)
	{
		//转账
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, 107, func() mining.TxItr {
			//keyPath := filepath.Join(config.KeystoreDirPath, "keystore"+strconv.Itoa(0)+".key")
			key := areas[0].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := mining.CreateTxPayV2(key, &key.GetCoinbase().Addr, &areas[1].Keystore.GetCoinbase().Addr, 2000000*1e8,
				1*1e8, 0, config.KeyPwd, "", nonceInt.Uint64(), 107, "", 1)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	{
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, 111, func() mining.TxItr {
			//keyPath := filepath.Join(config.KeystoreDirPath, "keystore"+strconv.Itoa(0)+".key")
			key := areas[0].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := mining.CreateTxPayV2(key, &key.GetCoinbase().Addr, &areas[2].Keystore.GetCoinbase().Addr, 2000000*1e8,
				1*1e8, 0, config.KeyPwd, "", nonceInt.Uint64(), 111, "", 1)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	{
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, 114, func() mining.TxItr {
			//keyPath := filepath.Join(config.KeystoreDirPath, "keystore"+strconv.Itoa(0)+".key")
			key := areas[0].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := mining.CreateTxPayV2(key, &key.GetCoinbase().Addr, &areas[3].Keystore.GetCoinbase().Addr, 2000000*1e8,
				1*1e8, 0, config.KeyPwd, "", nonceInt.Uint64(), 114, "", 1)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	{
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, 118, func() mining.TxItr {
			//keyPath := filepath.Join(config.KeystoreDirPath, "keystore"+strconv.Itoa(0)+".key")
			key := areas[0].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := mining.CreateTxPayV2(key, &key.GetCoinbase().Addr, &areas[4].Keystore.GetCoinbase().Addr, 2000000*1e8,
				1*1e8, 0, config.KeyPwd, "", nonceInt.Uint64(), 118, "", 1)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//第2个见证人质押
	{
		operationHeight := uint64(206)
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			key := areas[1].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := CreateTxDepositIn(key, &key.GetCoinbase().Addr, &key.GetCoinbase().Addr, chainConfig.Mining_deposit,
				1*1e8, config.KeyPwd, nonceInt.Uint64(), operationHeight, "witness2", 50)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//第3个见证人质押
	{
		operationHeight := uint64(284)
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			key := areas[2].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := CreateTxDepositIn(key, &key.GetCoinbase().Addr, &key.GetCoinbase().Addr, chainConfig.Mining_deposit,
				1*1e8, config.KeyPwd, nonceInt.Uint64(), operationHeight, "witness3", 50)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}

	//第4个见证人质押
	{
		operationHeight := uint64(306)
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			key := areas[3].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := CreateTxDepositIn(key, &key.GetCoinbase().Addr, &key.GetCoinbase().Addr, chainConfig.Mining_deposit,
				1*1e8, config.KeyPwd, nonceInt.Uint64(), operationHeight, "witness4", 50)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}

	//第5个见证人质押
	{
		operationHeight := uint64(389)
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			key := areas[4].Keystore.(*keystore.Keystore)
			nonceInt, err := mining.GetAddrNonce(&key.GetCoinbase().Addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := CreateTxDepositIn(key, &key.GetCoinbase().Addr, &key.GetCoinbase().Addr, chainConfig.Mining_deposit,
				1*1e8, config.KeyPwd, nonceInt.Uint64(), operationHeight, "witness5", 50)
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}

	//注册角色1奖励
	{
		rolen := 0
		operationHeight := uint64(389)
		operationHeight += uint64(13 * (rolen + 1))
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			utils.LogZero.Info().Str("开始注册角色", chainConfig.RoleRewardList[rolen].RoleName).Send()
			key := areas[0].Keystore.(*keystore.Keystore)
			coinbaseAddr := key.GetCoinbase().Addr
			addr := key.GetAddr()[rolen+1].Addr
			nonceInt, err := mining.GetAddrNonce(&addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := tx_name_in.BuildTx_namein_offline(key, coinbaseAddr, addr, nonceInt.Uint64(), operationHeight,
				chainConfig.RoleRewardList[rolen].RoleName, nil, []crypto.AddressCoin{addr},
				chainConfig.Mining_name_deposit_min, 1*1e8, config.KeyPwd, "")
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//注册角色2奖励
	{
		rolen := 1
		operationHeight := uint64(389)
		operationHeight += uint64(13 * (rolen + 1))
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			utils.LogZero.Info().Str("开始注册角色", chainConfig.RoleRewardList[rolen].RoleName).Send()
			key := areas[0].Keystore.(*keystore.Keystore)
			coinbaseAddr := key.GetCoinbase().Addr
			addr := key.GetAddr()[rolen+1].Addr
			nonceInt, err := mining.GetAddrNonce(&addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := tx_name_in.BuildTx_namein_offline(key, coinbaseAddr, addr, nonceInt.Uint64(), operationHeight,
				chainConfig.RoleRewardList[rolen].RoleName, nil, []crypto.AddressCoin{addr},
				chainConfig.Mining_name_deposit_min, 1*1e8, config.KeyPwd, "")
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//注册角色3奖励
	{
		rolen := 2
		operationHeight := uint64(389)
		operationHeight += uint64(13 * (rolen + 1))
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			utils.LogZero.Info().Str("开始注册角色", chainConfig.RoleRewardList[rolen].RoleName).Send()
			key := areas[0].Keystore.(*keystore.Keystore)
			coinbaseAddr := key.GetCoinbase().Addr
			addr := key.GetAddr()[rolen+1].Addr
			nonceInt, err := mining.GetAddrNonce(&addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := tx_name_in.BuildTx_namein_offline(key, coinbaseAddr, addr, nonceInt.Uint64(), operationHeight,
				chainConfig.RoleRewardList[rolen].RoleName, nil, []crypto.AddressCoin{addr},
				chainConfig.Mining_name_deposit_min, 1*1e8, config.KeyPwd, "")
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//注册角色4奖励
	{
		rolen := 3
		operationHeight := uint64(389)
		operationHeight += uint64(13 * (rolen + 1))
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			utils.LogZero.Info().Str("开始注册角色", chainConfig.RoleRewardList[rolen].RoleName).Send()
			key := areas[0].Keystore.(*keystore.Keystore)
			coinbaseAddr := key.GetCoinbase().Addr
			addr := key.GetAddr()[rolen+1].Addr
			nonceInt, err := mining.GetAddrNonce(&addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := tx_name_in.BuildTx_namein_offline(key, coinbaseAddr, addr, nonceInt.Uint64(), operationHeight,
				chainConfig.RoleRewardList[rolen].RoleName, nil, []crypto.AddressCoin{addr},
				chainConfig.Mining_name_deposit_min, 1*1e8, config.KeyPwd, "")
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	//注册角色5奖励
	{
		rolen := 4
		operationHeight := uint64(389)
		operationHeight += uint64(13 * (rolen + 1))
		eventPay := build.CreateEventBlockHeight(build.Event_type_pay, operationHeight, func() mining.TxItr {
			utils.LogZero.Info().Str("开始注册角色", chainConfig.RoleRewardList[rolen].RoleName).Send()
			key := areas[0].Keystore.(*keystore.Keystore)
			coinbaseAddr := key.GetCoinbase().Addr
			addr := key.GetAddr()[rolen+1].Addr
			nonceInt, err := mining.GetAddrNonce(&addr)
			if err != nil {
				panic(err)
			}
			utils.LogZero.Info().Str("地址", key.GetCoinbase().GetAddrStr()).Uint64("查询地址nonce", nonceInt.Uint64()).Send()
			pay, err := tx_name_in.BuildTx_namein_offline(key, coinbaseAddr, addr, nonceInt.Uint64(), operationHeight,
				chainConfig.RoleRewardList[rolen].RoleName, nil, []crypto.AddressCoin{addr},
				chainConfig.Mining_name_deposit_min, 1*1e8, config.KeyPwd, "")
			if err != nil {
				panic(err)
			}
			//mining.BuildOfflineTx(keyPath, config.KeyPwd, 1, 107, 0, 0, "", "", "")
			return pay
		})
		events = append(events, &eventPay)
	}
	return events
}

/*
创建一个见证人押金交易
@amount    uint64    押金额度
*/
func CreateTxDepositIn(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, amount, gas uint64,
	pwd string, nonceInt uint64, currentHeight uint64, payload string, rate uint16) (*mining.Tx_deposit_in, error) {
	// utils.LogZero.Debug().Msgf("创建见证人押金交易 111")
	if amount != chainConfig.Mining_deposit {
		// fmt.Println("交押金数量最少", config.Mining_deposit)
		//押金太少
		return nil, chainConfig.ERROR_deposit_witness
	}
	//chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	//currentHeight := chain.GetCurrentBlock()

	//key := Area.Keystore.GetCoinbase()

	//查找余额
	vins := make([]*mining.Vin, 0)

	//total, item := chain.Balance.BuildPayVinNew(nil, amount+gas)
	//if total < amount+gas {
	//	//资金不够
	//	return nil, config.ERROR_not_enough // errors.New("余额不足")
	//}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }
	// for _, item := range items {
	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		return nil, chainConfig.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	//nonce := chain.GetBalance().FindNonce(item.Addr)
	nonce := big.NewInt(int64(nonceInt))
	vin := mining.Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := mining.Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: *address, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	// if total > amount {
	// 	vout := Vout{
	// 		Value:   total - (amount + gas), //输出金额 = 实际金额 * 100000000
	// 		Address: key.Addr,               //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	var txin *mining.Tx_deposit_in
	//for i := uint64(0); i < 10000; i++ {
	//
	base := mining.TxBase{
		Type:       chainConfig.Wallet_tx_type_deposit_in,            //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  uint64(len(vouts)),                               //输入交易数量
		Vin:        vins,                                             //交易输入
		Vout_total: uint64(len(vouts)),                               //
		Vout:       vouts,                                            //
		Gas:        gas,                                              //交易手续费
		LockHeight: currentHeight + chainConfig.Wallet_tx_lockHeight, //锁定高度
		// CreateTime: config.TimeNow().Unix(),                //创建时间
		Payload: []byte(payload),
	}
	txin = &mining.Tx_deposit_in{
		TxBase: base,
		Puk:    puk,
		Rate:   rate,
	}
	//给输出签名，防篡改
	for i, one := range txin.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return nil, err
		}

		// utils.LogZero.Info().Msgf("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

		sign := txin.GetSign(&prk, uint64(i))
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		txin.Vin[i].Sign = *sign
		// for _, key := range keystore.GetAddr() {
		// 	puk, ok := keystore.GetPukByAddr(key.Addr)
		// 	if !ok {
		// 		return nil, config.ERROR_public_key_not_exist
		// 	}
		// 	if bytes.Equal(puk, one.Puk) {
		// 		_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)
		// 		// prk, err := key.GetPriKey(pwd)
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
		// 		//				sign := txin.GetVoutsSign(prk, uint64(i))
		// 		txin.Vin[i].Sign = *sign
		// 	}
		// }
	}
	txin.BuildHash()
	//if txin.CheckHashExist() {
	//	txin = nil
	//	continue
	//} else {
	//	break
	//}
	//}
	// chain.Balance.Frozen(items, txin)
	//chain.Balance.AddLockTx(txin)
	return txin, nil
}
