package mining

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"

	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

//保存网络中的交易
//var unpackedTransactions = new(sync.Map) //未打包的交易,key:string=交易hahs id；value=&TxItr

type TransactionManager struct {
	witnessBackup *WitnessBackup //候选见证人
	// depositin                  *sync.Map            //见证人缴押金,key:string=见证人公钥；value:*TransactionRatio=;见证人不能有重复，因此单独管理
	unpackedNonceDiscontinuity *sync.Map            //未打包的交易,key:string=交易hash id；value:*TransactionRatio=;
	unpackedTransaction        *UnpackedTransaction //
	tempTxLock                 *sync.RWMutex        //锁
	tempTx                     []TransactionRatio   //未验证的交易。有的交易验证需要花很长时间，导致阻塞，因此，先保存在这里，再一条一条的验证。
	tempTxsignal               chan bool            //有未验证的交易时，发送一个信号。
	checkedTxCache             *sync.Map            //已经检查过的交易缓存,超过锁定高度就清除掉 hash:lockheight
}

/*
添加一个未打包的交易
*/
//func (this *TransactionManager) AddTxOld(txItr TxItr) bool {
//	txItr.BuildHash()
//	if len((*txItr.GetVin())[0].Nonce.Bytes()) == 0 {
//		engine.Log.Info("txid:%s nonce is nil", hex.EncodeToString(*txItr.GetHash()))
//		return false
//	}
//	// engine.Log.Info("添加一个未打包的交易 111111111")
//	//已有的交易不重复保存，避免把交易的BlockHash字段覆盖掉
//	txhashkey := config.BuildBlockTx(*txItr.GetHash())
//	if ok, _ := db.LevelDB.CheckHashExist(txhashkey); !ok {
//		//把交易放到数据库，并且标记为未确认的交易。
//		err := SaveTempTx(txItr)
//		if err != nil {
//			engine.Log.Info("txid:%s save error:%s", txItr.GetHash(), err.Error())
//			return false
//		}
//	}
//
//	spend := txItr.GetSpend()
//	// spend := txItr.GetGas()
//	// for _, vout := range *txItr.GetVout() {
//	// 	spend += vout.Value
//	// }
//
//	txbs := txItr.Serialize()
//
//	div1 := new(big.Int).Mul(big.NewInt(int64(txItr.GetGas())), big.NewInt(100000000))
//	div2 := big.NewInt(int64(len(*txbs)))
//
//	ratioValue := new(big.Int).Div(div1, div2)
//	// ratioValue := txItr.GetGas() / uint64(len(*txbs))
//	ratio := TransactionRatio{
//		tx:    txItr,              //交易
//		size:  uint64(len(*txbs)), //交易总大小
//		gas:   txItr.GetGas(),     //手续费
//		Ratio: ratioValue,         //价值比
//	}
//	if txItr.Class() == config.Wallet_tx_type_voting_reward {
//		ratio.spendLock = spend
//	} else {
//		ratio.spend = spend
//	}
//
//	this.tempTxLock.Lock()
//	this.tempTx = append(this.tempTx, ratio)
//	this.tempTxLock.Unlock()
//	// engine.Log.Info("添加一个未打包的交易 222222222222")
//	select {
//	case this.tempTxsignal <- false:
//		// engine.Log.Info("添加一个未打包的交易 3333333333333")
//	default:
//		// engine.Log.Info("添加一个未打包的交易 4444444444444")
//	}
//	// engine.Log.Info("添加一个未打包的交易 555555555555555")
//	return true
//}

/*
将未验证的交易，一条一条的验证后，再添加到未打包的交易中去。
此方法异步执行
*/
func (this *TransactionManager) loopCheckTxs() {
	utils.Go(func() {

		// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		// engine.Log.Info("add Goroutine:%s", goroutineId)
		// defer engine.Log.Info("del Goroutine:%s", goroutineId)
		for range this.tempTxsignal {
			// engine.Log.Info("开始验证交易 1111111111111")
			this.tempTxLock.Lock()
			if len(this.tempTx) <= 0 {
				this.tempTxLock.Unlock()
				continue
			}
			//取出交易
			temp := this.tempTx
			//删除交易
			this.tempTx = make([]TransactionRatio, 0)
			this.tempTxLock.Unlock()
			//engine.Log.Info("开始验证交易 2222222222222    %d", len(temp))
			//开始验证交易
			for i, _ := range temp {
				one := temp[i]

				//engine.Log.Debug("开始验证交易 %s", hex.EncodeToString(*one.tx.GetHash()))
				//验证失败的交易
				err := this.checkTx(&one)
				if err != nil {
					engine.Log.Warn("交易验证失败 %s %s", hex.EncodeToString(*one.tx.GetHash()), err.Error())
					DelCacheTxAndLimit(hex.EncodeToString(*one.tx.GetHash()), one.tx)

					//if err != config.ERROR_tx_limiter_full {
					//	DelCacheTxAndLimit(hex.EncodeToString(*one.tx.GetHash()), one.tx)
					//}

					//清除本地缓存
					forks.GetLongChain().Balance.DelLockTx(one.tx)

					// engine.Log.Debug("Transaction validation failed %s %s", hex.EncodeToString(*one.tx.GetHash()), err.Error())
					// bs, _ := json.Marshal(one.GetVOJSON())
					// engine.Log.Debug("%s", string(bs))
					continue
				}
				this.checkedTxCache.Store(utils.Bytes2string(*one.tx.GetHash()), one.tx.GetLockHeight())
				// engine.Log.Debug("结束验证交易 %s", hex.EncodeToString(*one.tx.GetHash()))
			}
		}
	}, nil)
}

// 清理已经超过锁定高度的交易缓存
func (this *TransactionManager) cleanCheckTxCache() {
	t := time.NewTicker(time.Second * config.Wallet_tx_lockHeight)
	for range t.C {
		currentBlock := GetLongChain().CurrentBlock
		this.checkedTxCache.Range(func(key, value any) bool {
			lockHeight := value.(uint64)
			if currentBlock > lockHeight {
				this.checkedTxCache.Delete(key)
			}
			return true
		})
	}
}

/*
验证一个未验证的交易
*/
func (this *TransactionManager) checkTx(ratio *TransactionRatio) error {
	//if !AddTxLimit(ratio.tx) {
	//	return config.ERROR_tx_limiter_full
	//}

	var err error
	// startTime := config.TimeNow()
	txItr := ratio.tx
	txItr.BuildHash()

	//检查地址是否是大余额，大余额需要被锁定，不能转账
	if this.CheckAddrValueBig(ratio) {
		return config.ERROR_addr_value_big
	}

	//见证人押金不能多次提交
	// if txItr.Class() == config.Wallet_tx_type_deposit_in {
	// 	vin := (*txItr.GetVin())[0]

	// 	//判断是否已经交了押金
	// 	if this.witnessBackup.haveWitness(vin.GetPukToAddr()) {
	// 		// engine.Log.Warn("已经缴纳了押金了 %s", hex.EncodeToString(*txItr.GetHash()))
	// 		engine.Log.Warn("The deposit has been paid %s", hex.EncodeToString(*txItr.GetHash()))
	// 		return config.ERROR_deposit_exist
	// 	}

	// 	//判断是否有重复的未打包的押金
	// 	_, ok := this.depositin.Load(utils.Bytes2string(vin.Puk))
	// 	if ok {
	// 		// engine.Log.Warn("有重复的未打包的押金 %s", hex.EncodeToString(*txItr.GetHash()))
	// 		engine.Log.Warn("There is a duplicate unpackaged deposit %s", hex.EncodeToString(*txItr.GetHash()))
	// 		return config.ERROR_deposit_exist
	// 	}
	// 	// engine.Log.Info("Verifying 777 Use time %s", config.TimeNow().Sub(startTime))
	// 	this.depositin.Store(utils.Bytes2string(vin.Puk), ratio)
	// 	// engine.Log.Warn("见证人押金验证通过 %s", txItr.GetHashStr())
	// 	return nil
	// }

	//检查重复的交易
	txCtrl := GetTransactionCtrl(txItr.Class())
	if txCtrl != nil {
		err = txCtrl.CheckMultiplePayments(txItr)
		if err != nil {
			return err
		}
	}
	// engine.Log.Info("验证一个交易")
	//判断余额是否足够
	err = this.CheckNotspend(ratio)
	if err != nil {
		return err
	}

	err = this.CheckNFT(ratio)
	if err != nil {
		return err
	}

	//验证域名
	if !this.CheckDomain(txItr) {
		return errors.New("checkDomain验证域名解析失败")
	}

	//验证合约
	if ratio.gasUsed == 0 && this.CheckContract(ratio) != nil {
		return errors.New("验证合约失败")
	}

	//engine.Log.Info("验证一个交易 %s", hex.EncodeToString(*txItr.GetHash()))
	//判断nonce是否连续
	ok, err := this.CheckNonce(txItr)
	if err != nil {
		return err
	}
	// engine.Log.Info("验证一个交易:%t", ok)
	if !ok {
		engine.Log.Error("nonce不连续,交易hash:%s", hex.EncodeToString(*txItr.GetHash()))
		//engine.Log.Warn("保存到不连续 1111111111 %d %s %s",
		//	(*ratio.tx.GetVin())[0].Nonce.Uint64(),
		//	(*ratio.tx.GetVin())[0].GetPukToAddr().B58String(),
		//	hex.EncodeToString(*ratio.tx.GetHash()),
		//)
		//保存到不连续nonce缓存中
		this.unpackedNonceDiscontinuity.Store(utils.Bytes2string(*txItr.GetHash()), ratio)
		// engine.Log.Info("验证一个交易")

		//暂时清除限流
		DelTxLimit(txItr)
		return nil
	}
	//engine.Log.Info("添加一个连续交易 222222222222 %d %s %s",
	//	(*ratio.tx.GetVin())[0].Nonce.Uint64(),
	//	(*ratio.tx.GetVin())[0].GetPukToAddr().B58String(),
	//	hex.EncodeToString(*ratio.tx.GetHash()),
	//)

	this.unpackedTransaction.AddTx(ratio)
	// engine.Log.Info("验证一个交易")

	addr := (*ratio.tx.GetVin())[0].GetPukToAddr()
	addrNonceMap := make(map[uint64]*TransactionRatio)
	nonceSort := make([]uint64, 0)
	this.unpackedNonceDiscontinuity.Range(func(k, v interface{}) bool {
		ratio := v.(*TransactionRatio)
		if bytes.Equal(*addr, *(*ratio.tx.GetVin())[0].GetPukToAddr()) {
			n := (*ratio.tx.GetVin())[0].Nonce
			nonceSort = append(nonceSort, n.Uint64())
			//比较gas费用
			oldRadio, ok := addrNonceMap[n.Uint64()]
			if !ok || ratio.gas > oldRadio.gas {
				addrNonceMap[n.Uint64()] = ratio
			}
		}
		return true
	})
	sort.Slice(nonceSort, func(i, j int) bool { return nonceSort[i] < nonceSort[j] })

	for _, v := range nonceSort {
		ratio := addrNonceMap[v]
		//engine.Log.Info("遍历不连续 hash:%s,nonce:%d", hex.EncodeToString(*ratio.tx.GetHash()), (*ratio.tx.GetVin())[0].Nonce.Uint64())

		//判断余额是否足够
		err = this.CheckNotspend(ratio)
		if err != nil {
			engine.Log.Warn("验证一个交易错误:%s", err.Error())
			break
		}
		//return true
		//判断nonce是否连续
		ok, err := this.CheckNonce(ratio.tx)
		if err != nil {
			engine.Log.Warn("验证一个交易错误:%s", err.Error())
			continue
		}
		if !ok {
			//保存到不连续nonce缓存中
			continue
		}

		//通过不连续来的交易，增加限流
		AddTxLimit(ratio.tx)

		this.unpackedTransaction.AddTx(ratio)
	}

	return nil

	//继续在不连续的缓存中查找连续的nonce交易
	this.unpackedNonceDiscontinuity.Range(func(k, v interface{}) bool {
		ratio := v.(*TransactionRatio)
		//判断余额是否足够
		err = this.CheckNotspend(ratio)
		if err != nil {
			engine.Log.Warn("验证一个交易错误:%s", err.Error())
			return false
		}
		//return true
		//判断nonce是否连续
		ok, err := this.CheckNonce(ratio.tx)
		if err != nil {
			engine.Log.Warn("验证一个交易错误:%s", err.Error())
			return false
		}
		if !ok {
			//保存到不连续nonce缓存中
			return false
		}
		this.unpackedTransaction.AddTx(ratio)
		return true
	})
	// engine.Log.Info("验证一个交易")
	// this.unpacked.Store(utils.Bytes2string(*txItr.GetHash()), ratio)

	//保存活动的交易输出
	// for k, v := range activiVoutIndexs {
	// 	// engine.Log.Info("添加活动的交易 %s", k)
	// 	this.ActiveVoutIndex.Store(k, v)
	// }
	// engine.Log.Info("Verifying transactions Use time %s", config.TimeNow().Sub(startTime))
	return nil
}

/*
*
验证域名
*/
func (this *TransactionManager) CheckDomain(txItr TxItr) bool {
	return txItr.CheckDomain()
}

/*
*
获取未打包交易
*/
func (this *TransactionManager) GetUnpackedTransaction() *UnpackedTransaction {
	return this.unpackedTransaction
}

/*
检查nonce是否连续
@return    bool    是否连续
@return    error    错误
*/
func (this *TransactionManager) CheckNonce(txItr TxItr) (bool, error) { // engine.Log.Info("CheckNonce")
	//判断nonce是否连续
	var err error
	fromAddr := (*txItr.GetVin())[0].GetPukToAddr()
	nonce := this.unpackedTransaction.FindAddrNonce(fromAddr)
	srcNonce, err := GetAddrNonce(fromAddr)
	if err != nil {
		return false, err
	}
	if len(nonce.Bytes()) == 0 {
		nonce = srcNonce
	}
	nonceOne := (*txItr.GetVin())[0].Nonce

	//是否超出最大交易数
	if new(big.Int).Sub(&nonceOne, &srcNonce).Cmp(big.NewInt(int64(config.Wallet_addr_tx_count_max))) > 0 {
		//删除本地缓存
		forks.GetLongChain().Balance.DelLockTx(txItr)
		//删除数据库记录
		//db.LevelDB.Remove(*txItr.GetHash())
		//engine.Log.Error("addr:%s  hash:%s || nonce store:%d  curr:%d",
		//	fromAddr.B58String(), hex.EncodeToString(*txItr.GetHash()),
		//	srcNonce.Uint64(), nonceOne.Uint64())
		engine.Log.Warn("CheckNonce: limit max addr tx count address:%s,tx_nonce:%s,db_nonce:%s,pending_nonce:%s", fromAddr.B58String(), nonceOne.String(), srcNonce.String(), nonce.String())
		return false, errors.New("tx count limited")
	}

	addrStr := utils.Bytes2string(*fromAddr)
	tsrItr, ok := this.unpackedTransaction.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		if tsr.TrsLen() >= config.Wallet_addr_tx_count_max {
			//删除本地缓存
			forks.GetLongChain().Balance.DelLockTx(txItr)
			return false, errors.New("unpackedTransaction too more")
		}
	}

	//engine.Log.Info("srcnonce:%d,nonce:%d,nonceOne:%d", srcNonce.Uint64(), nonce.Uint64(), nonceOne.Uint64())
	//nonce是否连续
	cmp := new(big.Int).Add(&nonce, big.NewInt(1)).Cmp(&nonceOne)
	if cmp > 0 {
		//engine.Log.Error("addr:%s  hash:%s || nonce store:%d  curr:%d",
		//	fromAddr.B58String(), hex.EncodeToString(*txItr.GetHash()),
		//	srcNonce.Uint64(), nonceOne.Uint64())
		engine.Log.Warn("CheckNonce: nonce too small hash:%s,address:%s,tx_nonce:%s,db_nonce:%s,pending_nonce:%s", hex.EncodeToString(*txItr.GetHash()), fromAddr.B58String(), nonceOne.String(), srcNonce.String(), nonce.String())
		return false, errors.New("nonce too small")
	} else if cmp < 0 {
		engine.Log.Warn("CheckNonce: nonce too big hash:%s,address:%s,tx_nonce:%s,db_nonce:%s,pending_nonce:%s", hex.EncodeToString(*txItr.GetHash()), fromAddr.B58String(), nonceOne.String(), srcNonce.String(), nonce.String())
		return false, nil
	}

	return true, nil
}

/*
判断余额是否足够
@return    bool    是否连续
@return    error    错误
*/
func (this *TransactionManager) CheckNotspend(ratio *TransactionRatio) error {
	fromAddr := (*ratio.tx.GetVin())[0].GetPukToAddr()
	//spend, spendLock := this.unpackedTransaction.FindAddrSpend(fromAddr)
	spend, _ := this.unpackedTransaction.FindAddrSpend(fromAddr)

	//if ratio.tx.Class() == config.Wallet_tx_type_voting_reward {
	//	rfValue := GetCommunityVoteRewardFrozen(fromAddr)
	//	if spendLock+ratio.spendLock > rfValue {
	//		// engine.Log.Info("余额不足:%s %d %d %d", fromAddr.B58String(), rfValue, spend, ratio.spend)
	//		return config.ERROR_not_enough
	//	}
	//	return nil
	//}
	// engine.Log.Info("地址：%s", fromAddr.B58String())
	notSpend, _, _ := GetNotspendByAddrOther(this.witnessBackup.chain, *fromAddr)
	if notSpend < spend+ratio.spend {
		engine.Log.Info("余额不足:%s %s %d<%d+%d", hex.EncodeToString(*ratio.tx.GetHash()), fromAddr.B58String(), notSpend, spend, ratio.spend)
		return config.ERROR_not_enough
	}
	return nil
}

/*
检查NFT是否对应账户
@return    error    错误
*/
func (this *TransactionManager) CheckNFT(ratio *TransactionRatio) error {
	if ratio.tx.Class() == config.Wallet_tx_type_nft {
		txNFT := ratio.tx.(*Tx_nft)
		if txNFT.NFT_ID == nil || len(txNFT.NFT_ID) == 0 {
			return nil
		}
		oldOwner, err := FindNFTOwner(txNFT.NFT_ID)
		if err != nil {
			return err
		}
		if bytes.Equal(*oldOwner, *txNFT.Vin[0].GetPukToAddr()) {
			return nil
		}
		return config.ERROR_tx_nft_not_our_own
	}
	if ratio.tx.Class() == config.Wallet_tx_type_nft_exchange {
		txNFT := ratio.tx.(*Tx_nft_exchange)
		oldOwner, err := FindNFTOwner(txNFT.NFT_ID_Sponsor)
		if err != nil {
			return err
		}
		if !bytes.Equal(*oldOwner, *txNFT.Vin[0].GetPukToAddr()) {
			return config.ERROR_tx_nft_not_our_own
		}

		oldOwner, err = FindNFTOwner(txNFT.NFT_ID_Recipient)
		if err != nil {
			return err
		}
		if !bytes.Equal(*oldOwner, txNFT.NFT_Recipient_addr) {
			return config.ERROR_tx_nft_not_our_own
		}
		return nil
	}
	if ratio.tx.Class() == config.Wallet_tx_type_nft_destroy {
		txNFT := ratio.tx.(*Tx_nft_destroy)
		oldOwner, err := FindNFTOwner(txNFT.NFT_ID)
		if err != nil {
			return err
		}
		if bytes.Equal(*oldOwner, *txNFT.Vin[0].GetPukToAddr()) {
			return nil
		}
		return config.ERROR_tx_nft_not_our_own
	}
	return nil
}

// 检查合约交易
func (this *TransactionManager) CheckContract(ratio *TransactionRatio) error {
	if ratio.tx.Class() == config.Wallet_tx_type_contract {
		tx := ratio.tx.(*Tx_Contract)
		return tx.PreExec()
	}
	return nil
}

/*
判断是否大余额
@return    bool    是否大余额
*/
func (this *TransactionManager) CheckAddrValueBig(ratio *TransactionRatio) bool {
	for _, one := range *ratio.tx.GetVin() {
		addr := one.GetPukToAddr()
		if GetAddrValueBig(addr) {
			return true
		}
	}
	return false
}

/*
添加一个未打包的交易
*/
func (this *TransactionManager) AddTxs(txs ...TxItr) {
	for _, one := range txs {
		this.AddTx(one)
	}
}

/*
删除一个已经打包的交易
*/
func (this *TransactionManager) DelTx(txs []TxItr) {
	for _, one := range txs {
		// str := hex.EncodeToString(*one.GetHash())
		// str := one.GetHashStr()
		// if one.Class() == config.Wallet_tx_type_deposit_in {
		// 	pukStr := (*one.GetVin())[0].Puk //.GetPukStr() // hex.EncodeToString((*one.GetVin())[0].Puk)
		// 	this.depositin.Delete(utils.Bytes2string(pukStr))
		// }

		// this.unpacked.Delete(utils.Bytes2string(*one.GetHash()))
		// engine.Log.Info("删除已经打包交易 %s", hex.EncodeToString(*one.GetHash()))
		this.unpackedNonceDiscontinuity.Delete(utils.Bytes2string(*one.GetHash()))
		this.unpackedTransaction.DelTx(one)

		db.LevelDB.Remove(config.BuildTxNotImport(*one.GetHash()))

		//删除本地缓存
		this.witnessBackup.chain.Balance.DelLockTx(one)
		//删除交易输出中的索引
		// for _, vin := range *one.GetVin() {
		// 	// keyStr := hex.EncodeToString(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
		// 	keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
		// 	// engine.Log.Info("删除活动的交易 %s", keyStr)
		// 	this.ActiveVoutIndex.Delete(keyStr)
		// }
	}
}

/*
打包交易
控制每个区块大小，给交易手续费排序。
*/
func (this *TransactionManager) Package(reward *Tx_reward, height uint64, blocks []Block, createBlockTime int64) ([]TxItr, [][]byte) {
	// engine.Log.Info("开始打包")
	//未确认的交易
	unacknowledgedTxs := make([]TxItr, 0)
	//未确认的nonce
	// unacknowledgedNonce := make(map[string]*big.Int, 0)

	//预执行交易evm
	vmRun := evm.NewCountVmRun(nil)
	//初始化storage
	vmRun.SetStorage(nil)

	// start := config.TimeNow()
	//排除已经打包的交易

	//engine.Log.Error("package 高度：%d %d blocks:%d", height, blocks[len(blocks)-1].Height, len(blocks))
	exclude := make(map[string]string)
	for _, one := range blocks {
		// engine.Log.Info("打包加载区块信息")
		_, txs, err := one.LoadTxs()
		if err != nil {
			engine.Log.Warn("打包：未加载区块信息")
			return nil, nil
		}
		for _, txOne := range *txs {
			// exclude[hex.EncodeToString(*txOne.GetHash())] = ""
			exclude[utils.Bytes2string(*txOne.GetHash())] = ""
			unacknowledgedTxs = append(unacknowledgedTxs, txOne)

			//判断合约交易并预执行
			if !one.isCount && txOne.Class() == config.Wallet_tx_type_contract {
				//engine.Log.Error("合约在区块高度：%d", one.Height)
				from := (*txOne.GetVin())[0].GetPukToAddr()
				to := (*txOne.GetVout())[0].Address
				vmRun.SetTxContext(*from, to, *txOne.GetHash(), createBlockTime, height, nil, nil)
				if err := txOne.(*Tx_Contract).PreExecV1(vmRun); err != nil {
					engine.Log.Warn("打包：检查交易:%s,错误:合约预执行失败,%s", hex.EncodeToString(*txOne.GetHash()), err.Error())
					return nil, nil
				}
			}
		}
	}
	//engine.Log.Error("***************************************************")
	// engine.Log.Info("排除打包的交易花费耗时 %s", config.TimeNow().Sub(start))

	//打包见证人押金
	// txRatios := make([]TransactionRatio, 0)
	// this.depositin.Range(func(k, v interface{}) bool {
	// 	// engine.Log.Debug("开始打包见证人押金交易 111")
	// 	txid := k.(string)
	// 	//判断是否有排除的交易，有则不加入列表
	// 	_, ok := exclude[txid]
	// 	if ok {
	// 		// engine.Log.Debug("开始打包见证人押金交易 222")
	// 		return true
	// 	}

	// 	txItr := v.(*TransactionRatio)
	// 	//判断交易锁定高度
	// 	if err := txItr.tx.CheckLockHeight(height); err != nil {
	// 		// engine.Log.Debug("开始打包见证人押金交易 333")
	// 		return true
	// 	}
	// 	//判断余额冻结高度
	// 	// if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
	// 	// 	return true
	// 	// }

	// 	//判断交易签名是否正确
	// 	txRatios = append(txRatios, txItr)
	// 	// txids = append(txids, *txItr.GetHash())
	// 	// engine.Log.Debug("开始打包见证人押金交易 444")
	// 	return true
	// })

	tsrs := this.unpackedTransaction.FindTx()
	txs := make([]TxItr, 0)
	txids := make([][]byte, 0)
	sizeTotal := uint64(0)
	gasTotal := uint64(0)
	for _, one := range *tsrs {
		addrTxs := make([]TxItr, 0, one.TrsLen())
		delTxs := make([]TxItr, 0)
		trsSizeTotal := uint64(0)
		trsGasTotal := uint64(0)
		for j, tsr := range one.GetTrs() {
			//engine.Log.Info("打包：检查交易:%s", hex.EncodeToString(*tsr.tx.GetHash()))
			//计算gas
			if gasTotal+trsGasTotal+tsr.gasUsed > config.BLOCK_TOTAL_GAS {
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "limit block gas")
				break
			}
			//计算大小
			if sizeTotal+trsSizeTotal+tsr.size > config.Block_size_max {
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "limit block size")
				break
			}
			_, ok := exclude[utils.Bytes2string(*tsr.tx.GetHash())]
			if ok {
				//engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "block exclude tx")
				break
			}
			if err := tsr.tx.CheckLockHeight(height); err != nil {
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "check lockheight")
				break
			}
			//判断重复的交易
			if !tsr.tx.CheckRepeatedTx(append(unacknowledgedTxs, addrTxs...)...) {
				//删除这个用户的交易
				for _, r := range one.GetTrs()[j:] {
					delTxs = append(delTxs, r.tx)
				}
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "repeated tx")
				break
			}

			//判断合约交易并预执行
			if tsr.tx.Class() == config.Wallet_tx_type_contract {
				from := (*tsr.tx.GetVin())[0].GetPukToAddr()
				to := (*tsr.tx.GetVout())[0].Address
				vmRun.SetTxContext(*from, to, *tsr.tx.GetHash(), createBlockTime, height, nil, nil)
				if err := tsr.tx.(*Tx_Contract).PreExecV1(vmRun); err != nil {
					for _, r := range one.GetTrs()[j:] {
						delTxs = append(delTxs, r.tx)
					}
					engine.Log.Warn("打包：检查交易:%s,错误:合约预执行失败,%s", hex.EncodeToString(*tsr.tx.GetHash()), err.Error())
					break
				}
			}

			trsSizeTotal += tsr.size
			trsGasTotal += tsr.gasUsed
			addrTxs = append(addrTxs, tsr.tx)
		}

		if len(delTxs) > 0 {
			this.DelTx(delTxs)
		}
		if len(addrTxs) > 0 {
			txs = append(txs, addrTxs...)
			sizeTotal = sizeTotal + trsSizeTotal
			gasTotal = gasTotal + trsGasTotal
			unacknowledgedTxs = append(unacknowledgedTxs, addrTxs...)
			for _, txOne := range addrTxs {
				txids = append(txids, *txOne.GetHash())
			}
		}
	}
	//打包普通交易
	// this.unpacked.Range(func(k, v interface{}) bool {
	// 	txid := k.(string)
	// 	// engine.Log.Info("判断本次普通交易 %s", txid)
	// 	//判断是否有排除的交易，有则不加入列表
	// 	_, ok := exclude[txid]
	// 	if ok {
	// 		// engine.Log.Info("排除已经打包的交易 %s", txid)
	// 		return true
	// 	}

	// 	txItr := v.(TransactionRatio)

	// 	if err := txItr.tx.CheckLockHeight(height); err != nil {
	// 		// engine.Log.Info("排除锁定高度不正确的交易 %s %s", txid, err.Error())
	// 		return true
	// 	}

	// 	// if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
	// 	// 	// engine.Log.Info("排除冻结余额高度不正确的交易 %s %s", txid, err.Error())
	// 	// 	return true
	// 	// }

	// 	txRatios = append(txRatios, txItr)
	// 	// txids = append(txids, *txItr.GetHash())
	// 	// fmt.Println("===222打包普通交易", hex.EncodeToString(*txItr.GetHash()))
	// 	return true
	// })
	//排序，把手续费多的排前面
	// rss := &RatioSort{txRatio: txRatios}
	// sort.Sort(rss)
	// txRatios = rss.txRatio

	// //获取排在前面的
	// // start = config.TimeNow()
	// txs := make([]TxItr, 0)
	// txids := make([][]byte, 0)
	// sizeTotal := uint64(0)
	// //多个见证人的时候，有的区块没有奖励
	// if reward != nil {
	// 	sizeTotal = sizeTotal + uint64(len(*reward.Serialize()))
	// }
	// for _, one := range txRatios {
	// 	if sizeTotal+one.size > config.Block_size_max {
	// 		//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
	// 		continue
	// 	}

	// 	//判断重复的交易
	// 	if !one.tx.CheckRepeatedTx(unacknowledgedTxs...) {
	// 		// engine.Log.Info("打包时有重复的交易 %s", one.tx.GetHashStr())
	// 		continue
	// 	}

	// 	txs = append(txs, one.tx)
	// 	txids = append(txids, *one.tx.GetHash())
	// 	sizeTotal = sizeTotal + one.size
	// 	unacknowledgedTxs = append(unacknowledgedTxs, one.tx)
	// }

	// engine.Log.Info("本次打包大小 %d 交易个数 %d", sizeTotal, len(txids))

	// engine.Log.Info("获取排在前面的交易耗时 %s", config.TimeNow().Sub(start))

	return txs, txids
}

/*
打包交易
控制每个区块大小，给交易手续费排序。
*/
func (this *TransactionManager) Package_Old(reward *Tx_reward, height uint64, blocks []Block, createBlockTime int64) ([]TxItr, [][]byte) {
	// engine.Log.Info("开始打包")
	//未确认的交易
	unacknowledgedTxs := make([]TxItr, 0)
	//未确认的nonce
	// unacknowledgedNonce := make(map[string]*big.Int, 0)

	//预执行交易evm
	vmRun := evm.NewCountVmRun(nil)
	//初始化storage
	vmRun.SetStorage(nil)

	// start := config.TimeNow()
	//排除已经打包的交易

	//engine.Log.Error("package 高度：%d %d blocks:%d", height, blocks[len(blocks)-1].Height, len(blocks))
	exclude := make(map[string]string)
	for _, one := range blocks {
		// engine.Log.Info("打包加载区块信息")
		_, txs, err := one.LoadTxs()
		if err != nil {
			engine.Log.Warn("打包：未加载区块信息")
			return nil, nil
		}
		for _, txOne := range *txs {
			// exclude[hex.EncodeToString(*txOne.GetHash())] = ""
			exclude[utils.Bytes2string(*txOne.GetHash())] = ""
			unacknowledgedTxs = append(unacknowledgedTxs, txOne)

			//判断合约交易并预执行
			if !one.isCount && txOne.Class() == config.Wallet_tx_type_contract {
				//engine.Log.Error("合约在区块高度：%d", one.Height)
				from := (*txOne.GetVin())[0].GetPukToAddr()
				to := (*txOne.GetVout())[0].Address
				vmRun.SetTxContext(*from, to, *txOne.GetHash(), createBlockTime, height, nil, nil)
				if err := txOne.(*Tx_Contract).PreExecV1(vmRun); err != nil {
					engine.Log.Warn("打包：检查交易:%s,错误:合约预执行失败,%s", hex.EncodeToString(*txOne.GetHash()), err.Error())
					return nil, nil
				}
			}
		}
	}
	//engine.Log.Error("***************************************************")
	// engine.Log.Info("排除打包的交易花费耗时 %s", config.TimeNow().Sub(start))

	//打包见证人押金
	// txRatios := make([]TransactionRatio, 0)
	// this.depositin.Range(func(k, v interface{}) bool {
	// 	// engine.Log.Debug("开始打包见证人押金交易 111")
	// 	txid := k.(string)
	// 	//判断是否有排除的交易，有则不加入列表
	// 	_, ok := exclude[txid]
	// 	if ok {
	// 		// engine.Log.Debug("开始打包见证人押金交易 222")
	// 		return true
	// 	}

	// 	txItr := v.(*TransactionRatio)
	// 	//判断交易锁定高度
	// 	if err := txItr.tx.CheckLockHeight(height); err != nil {
	// 		// engine.Log.Debug("开始打包见证人押金交易 333")
	// 		return true
	// 	}
	// 	//判断余额冻结高度
	// 	// if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
	// 	// 	return true
	// 	// }

	// 	//判断交易签名是否正确
	// 	txRatios = append(txRatios, txItr)
	// 	// txids = append(txids, *txItr.GetHash())
	// 	// engine.Log.Debug("开始打包见证人押金交易 444")
	// 	return true
	// })

	tsrs := this.unpackedTransaction.FindTx()
	txs := make([]TxItr, 0)
	txids := make([][]byte, 0)
	repeatTransaction := true
	checkRepeated := false
	sizeTotal := uint64(0)
	gasTotal := uint64(0)
	for _, one := range *tsrs {
		repeatTransaction = true
		checkRepeated = false
		addrTxs := make([]TxItr, 0, one.TrsLen())
		trsSizeTotal := uint64(0)
		trsGasTotal := uint64(0)
		for _, tsr := range one.GetTrs() {
			//engine.Log.Info("打包：检查交易:%s", hex.EncodeToString(*tsr.tx.GetHash()))
			//计算gas
			if gasTotal+trsGasTotal+tsr.gasUsed > config.BLOCK_TOTAL_GAS {
				repeatTransaction = false
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "limit block gas")
				break
			}
			//计算大小
			if sizeTotal+trsSizeTotal+tsr.size > config.Block_size_max {
				repeatTransaction = false
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "limit block size")
				break
			}
			_, ok := exclude[utils.Bytes2string(*tsr.tx.GetHash())]
			if ok {
				repeatTransaction = false
				//engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "block exclude tx")
				break
			}
			if err := tsr.tx.CheckLockHeight(height); err != nil {
				repeatTransaction = false
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "check lockheight")
				break
			}
			//判断重复的交易
			if !tsr.tx.CheckRepeatedTx(append(unacknowledgedTxs, addrTxs...)...) {
				//删除这个用户的交易
				repeatTransaction = false
				checkRepeated = true
				engine.Log.Warn("打包：检查交易:%s,错误:%s", hex.EncodeToString(*tsr.tx.GetHash()), "repeated tx")
				break
			}

			//判断合约交易并预执行
			if tsr.tx.Class() == config.Wallet_tx_type_contract {
				from := (*tsr.tx.GetVin())[0].GetPukToAddr()
				to := (*tsr.tx.GetVout())[0].Address
				vmRun.SetTxContext(*from, to, *tsr.tx.GetHash(), createBlockTime, height, nil, nil)
				if err := tsr.tx.(*Tx_Contract).PreExecV1(vmRun); err != nil {
					repeatTransaction = false
					checkRepeated = true
					engine.Log.Warn("打包：检查交易:%s,错误:合约预执行失败,%s", hex.EncodeToString(*tsr.tx.GetHash()), err.Error())
					break
				}
			}

			trsSizeTotal += tsr.size
			trsGasTotal += tsr.gasUsed
			addrTxs = append(addrTxs, tsr.tx)
		}

		if checkRepeated {
			txs := make([]TxItr, 0, one.TrsLen())
			for _, tsr := range one.GetTrs() {
				txs = append(txs, tsr.tx)
			}

			this.DelTx(txs)
			continue
		}
		if repeatTransaction {
			txs = append(txs, addrTxs...)
			sizeTotal = sizeTotal + one.size
			gasTotal = gasTotal + one.gasUsed
			unacknowledgedTxs = append(unacknowledgedTxs, addrTxs...)
			for _, txOne := range addrTxs {
				txids = append(txids, *txOne.GetHash())
			}
		}
	}
	//打包普通交易
	// this.unpacked.Range(func(k, v interface{}) bool {
	// 	txid := k.(string)
	// 	// engine.Log.Info("判断本次普通交易 %s", txid)
	// 	//判断是否有排除的交易，有则不加入列表
	// 	_, ok := exclude[txid]
	// 	if ok {
	// 		// engine.Log.Info("排除已经打包的交易 %s", txid)
	// 		return true
	// 	}

	// 	txItr := v.(TransactionRatio)

	// 	if err := txItr.tx.CheckLockHeight(height); err != nil {
	// 		// engine.Log.Info("排除锁定高度不正确的交易 %s %s", txid, err.Error())
	// 		return true
	// 	}

	// 	// if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
	// 	// 	// engine.Log.Info("排除冻结余额高度不正确的交易 %s %s", txid, err.Error())
	// 	// 	return true
	// 	// }

	// 	txRatios = append(txRatios, txItr)
	// 	// txids = append(txids, *txItr.GetHash())
	// 	// fmt.Println("===222打包普通交易", hex.EncodeToString(*txItr.GetHash()))
	// 	return true
	// })
	//排序，把手续费多的排前面
	// rss := &RatioSort{txRatio: txRatios}
	// sort.Sort(rss)
	// txRatios = rss.txRatio

	// //获取排在前面的
	// // start = config.TimeNow()
	// txs := make([]TxItr, 0)
	// txids := make([][]byte, 0)
	// sizeTotal := uint64(0)
	// //多个见证人的时候，有的区块没有奖励
	// if reward != nil {
	// 	sizeTotal = sizeTotal + uint64(len(*reward.Serialize()))
	// }
	// for _, one := range txRatios {
	// 	if sizeTotal+one.size > config.Block_size_max {
	// 		//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
	// 		continue
	// 	}

	// 	//判断重复的交易
	// 	if !one.tx.CheckRepeatedTx(unacknowledgedTxs...) {
	// 		// engine.Log.Info("打包时有重复的交易 %s", one.tx.GetHashStr())
	// 		continue
	// 	}

	// 	txs = append(txs, one.tx)
	// 	txids = append(txids, *one.tx.GetHash())
	// 	sizeTotal = sizeTotal + one.size
	// 	unacknowledgedTxs = append(unacknowledgedTxs, one.tx)
	// }

	// engine.Log.Info("本次打包大小 %d 交易个数 %d", sizeTotal, len(txids))

	// engine.Log.Info("获取排在前面的交易耗时 %s", config.TimeNow().Sub(start))

	return txs, txids
}

/*
清除过期的交易
*/
func (this *TransactionManager) CleanIxOvertime(height uint64) {
	//清除过期的见证人押金交易
	// this.depositin.Range(func(k, v interface{}) bool {
	// 	txBase := v.(*TransactionRatio)
	// 	lockheight := txBase.tx.GetLockHeight()
	// 	if lockheight < height {
	// 		//删除
	// 		this.depositin.Delete(k.(string))
	// 		//删除交易输出中的索引和缓存
	// 		// for _, vin := range *txBase.tx.GetVin() {
	// 		// 	// txidStr := vin.GetTxidStr() //hex.EncodeToString(vin.Txid)
	// 		// 	keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
	// 		// 	// engine.Log.Info("删除过期的活动交易 111 %s", keyStr)
	// 		// 	this.ActiveVoutIndex.Delete(keyStr)
	// 		// 	//删除缓存
	// 		// 	// TxCache.RemoveTxInCache(txidStr, vin.Vout)
	// 		// }
	// 	}
	// 	return true
	// })
	//清除过期的普通交易
	//不验证的区块不清除锁定高度
	if GetHighestBlock() < config.Mining_block_start_height+config.Mining_block_start_height_jump {
		return
	}
	cleanHashs := this.unpackedTransaction.CleanOverTimeTx(height)
	for _, hash := range cleanHashs {
		this.unpackedNonceDiscontinuity.Delete(utils.Bytes2string(hash))
	}
	// this.unpacked.Range(func(k, v interface{}) bool {
	// 	txBase := v.(TransactionRatio)
	// 	lockheight := txBase.tx.GetLockHeight()
	// 	if lockheight < height {
	// 		//删除
	// 		this.unpacked.Delete(k.(string))
	// 		//删除交易输出中的索引和缓存
	// 		// for _, vin := range *txBase.tx.GetVin() {
	// 		// 	// txidStr := vin.GetTxidStr() // hex.EncodeToString(vin.Txid)
	// 		// 	keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
	// 		// 	// engine.Log.Info("删除过期的活动的交易 222 %s", hex.EncodeToString(vin.Txid))
	// 		// 	this.ActiveVoutIndex.Delete(keyStr)
	// 		// 	//删除缓存
	// 		// 	// TxCache.RemoveTxInCache(txidStr, vin.Vout)
	// 		// }
	// 	}
	// 	return true
	// })

}

/*
	查询见证人是否缴纳押金
*/
// func (this *TransactionManager) FindDeposit(puk []byte) bool {
// 	_, ok := this.depositin.Load(utils.Bytes2string(puk))
// 	return ok
// }

/*
创建交易管理
*/
func NewTransactionManager(wb *WitnessBackup) *TransactionManager {
	tm := TransactionManager{
		witnessBackup: wb, //
		// depositin:                  new(sync.Map), //见证人缴押金,key:string=交易hahs id；value=&TxItr
		unpackedNonceDiscontinuity: new(sync.Map), //未打包的交易,key:string=交易hash id；value:TransactionRatio=;
		// unpacked:            new(sync.Map),               //未打包的交易,key:string=交易hahs id；value:&TxItr=;
		unpackedTransaction: NewUnpackedTransaction(), //
		// ActiveVoutIndex:     new(sync.Map),               //
		tempTxLock:     new(sync.RWMutex),           //
		tempTx:         make([]TransactionRatio, 0), //
		tempTxsignal:   make(chan bool, 1),          //
		checkedTxCache: new(sync.Map),
	}
	tm.loopCheckTxs()
	go tm.cleanCheckTxCache()

	return &tm
}
