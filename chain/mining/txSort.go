package mining

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"

	"web3_gui/chain/config"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/utils"
)

/*
区块中的交易价值比
*/
type TransactionRatio struct {
	tx        TxItr    //交易
	size      uint64   //交易总大小
	gas       uint64   //手续费
	Ratio     *big.Int //价值比
	spend     uint64   //交易花费金额
	spendLock uint64   //花费锁定的金额
	gasUsed   uint64   //使用的gas
}

/*
区块中的交易价值比排序
*/
type RatioSort struct {
	txRatio []TransactionRatio
}

func (this *RatioSort) Len() int {
	return len(this.txRatio)
}

func (this *RatioSort) Less(i, j int) bool {
	if this.txRatio[i].Ratio.Cmp(this.txRatio[j].Ratio) < 0 {
		return false
	} else {
		return true
	}
}

func (this *RatioSort) Swap(i, j int) {
	this.txRatio[i], this.txRatio[j] = this.txRatio[j], this.txRatio[i]
}

//-----------------------------------
/*
	相同地址的交易缓存,价值比等信息
*/
type TransactionsRatio struct {
	lock       *sync.RWMutex
	trs        []*TransactionRatio //交易
	size       uint64              //交易总大小
	gas        uint64              //手续费
	Ratio      *big.Int            //价值比
	spend      uint64              //交易花费金额
	tokenspend uint64              //交易花费金额
	spendLock  uint64              //花费锁定的金额
	gasUsed    uint64              //总gas
}

/*
获取交易的数量
*/
func (this *TransactionsRatio) TrsLen() int {
	this.lock.RLock()
	n := len(this.trs)
	this.lock.RUnlock()
	return n
}

/*
获取交易的数量
*/
func (this *TransactionsRatio) GetTrs() []*TransactionRatio {
	this.lock.RLock()
	trs := make([]*TransactionRatio, len(this.trs))
	copy(trs, this.trs)
	this.lock.RUnlock()
	return trs
}

/*
查询交易是否存在
*/
func (this *TransactionsRatio) FindTxByHash(hs *[]byte) bool {
	exist := false
	this.lock.RLock()
	for _, one := range this.trs {
		if bytes.Equal(*one.tx.GetHash(), *hs) {
			exist = true
			break
		}
	}
	this.lock.RUnlock()
	return exist
}

/*
查询最新nonce
*/
func (this *TransactionsRatio) GetNonce() big.Int {
	this.lock.RLock()
	trOne := this.trs[len(this.trs)-1]
	vins := trOne.tx.GetVin()
	nonce := (*vins)[0].Nonce
	this.lock.RUnlock()
	return nonce
}

/*
查询最新nonce
*/
func (this *TransactionsRatio) DelTx(hs *[]byte) {
	this.lock.Lock()
	for i, one := range this.trs {
		if bytes.Equal(*one.tx.GetHash(), *hs) {
			this.gas -= one.gas
			this.size -= one.size
			this.spend -= one.spend
			this.spendLock -= one.spendLock
			this.gasUsed -= one.gasUsed
			temp := this.trs[:i]
			this.trs = append(temp, this.trs[i+1:]...)
			break
		}
	}
	this.lock.Unlock()
}

/*
区块中的交易价值比排序
*/
type AddrRatioSort struct {
	txRatio []TransactionsRatio
}

func (this *AddrRatioSort) Len() int {
	return len(this.txRatio)
}

func (this *AddrRatioSort) Less(i, j int) bool {
	if this.txRatio[i].Ratio.Cmp(this.txRatio[j].Ratio) < 0 {
		return false
	} else {
		return true
	}
}

func (this *AddrRatioSort) Swap(i, j int) {
	this.txRatio[i], this.txRatio[j] = this.txRatio[j], this.txRatio[i]
}

type UnpackedTransaction struct {
	addrs      *sync.Map //key:string=地址;value:*TransactionsRatio=交易;
	addrsToken *sync.Map //key:string=tokenid+addr;value:*AddrToken; 用于计算Token未打包交易金额 ,不要做快照
	// txs   *sync.Map //key:string=交易hash;value:*TxItr=交易;
}

func (this *UnpackedTransaction) AddTx(tr *TransactionRatio) {
	//处理Token未打包交易统计花费
	//config.Wallet_tx_type_token_publish 部署Token不需要处理
	addr := (*tr.tx.GetVin())[0].GetPukToAddr()
	addrStr := utils.Bytes2string(*addr)
	tsrItr, ok := this.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		if tsr.TrsLen() >= config.Wallet_addr_tx_count_max || tsr.size+tr.size > config.Block_size_max {
			//删除本地缓存
			forks.GetLongChain().Balance.DelLockTx(tr.tx)
			DelTxLimit(tr.tx)
			return
		}
		tsr.lock.Lock()
		//engine.Log.Info("AddTx:1111111111111111")
		tsr.gas += tr.gas
		tsr.size += tr.size
		tsr.trs = append(tsr.trs, tr)
		tsr.spend += tr.spend
		tsr.spendLock += tr.spendLock
		tsr.gasUsed += tr.gasUsed
		div1 := new(big.Int).Mul(big.NewInt(int64(tsr.gas)), big.NewInt(100000000))
		div2 := big.NewInt(int64(tsr.size))
		tsr.Ratio = new(big.Int).Div(div1, div2)
		tsr.lock.Unlock()
	} else {
		tsr := TransactionsRatio{
			lock:      new(sync.RWMutex),
			trs:       make([]*TransactionRatio, 0), //交易
			size:      tr.size,                      //交易总大小
			gas:       tr.gas,                       //手续费
			Ratio:     tr.Ratio,                     //价值比
			spend:     tr.spend,                     //交易花费金额
			spendLock: tr.spendLock,                 //花费锁定的金额
			gasUsed:   tr.gasUsed,                   //总gas
		}
		tsr.trs = append(tsr.trs, tr)
		this.addrs.Store(addrStr, &tsr)
	}

	// AddTx处理代币交易燃料费(iCom)
	// addTx4TokenPay处理代币金额
	this.addTx4TokenPay(tr.tx)
}

// 添加Token的未打包交易金额
func (this *UnpackedTransaction) addTx4TokenPay(tx TxItr) {
	switch tx.Class() {
	case config.Wallet_tx_type_token_publish:
		//note: 不处理
	case config.Wallet_tx_type_token_payment:
		tokenTx := tx.(*TxTokenPay)
		addr := tokenTx.Token_Vin[0].GetPukToAddr()
		key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.Token_txid, *addr))
		value := big.NewInt(0)
		for _, vout := range tokenTx.Token_Vout {
			value.Add(value, new(big.Int).SetUint64(vout.Value))
		}

		if addrToken, ok := this.addrsToken.LoadOrStore(key, &AddrToken{
			TokenId: tokenTx.Token_txid,
			Addr:    *addr,
			Value:   value,
		}); ok {
			addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
		}
	case config.Wallet_tx_type_token_new_order:
		newOrder := tx.(*TxTokenNewOrder)
		//当Buy=true代表买入A，vin用于证明B资产。
		//当Buy=false代表卖出A，vin用于证明A资产。
		addr := newOrder.TokenVin.GetPukToAddr()
		if newOrder.Buy {
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(newOrder.TokenBID, *addr))
			value := new(big.Int).Set(newOrder.TokenBAmount)
			if addrToken, ok := this.addrsToken.LoadOrStore(key, &AddrToken{
				TokenId: newOrder.TokenBID,
				Addr:    *addr,
				Value:   value,
			}); ok {
				addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
			}
		} else {
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(newOrder.TokenAID, *addr))
			value := new(big.Int).Set(newOrder.TokenAAmount)
			if addrToken, ok := this.addrsToken.LoadOrStore(key, &AddrToken{
				TokenId: newOrder.TokenAID,
				Addr:    *addr,
				Value:   value,
			}); ok {
				addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
			}
		}

	case config.Wallet_tx_type_token_swap_order:
		//note: 不处理
	case config.Wallet_tx_type_swap:
		tokenTx := tx.(*TxSwap)
		//只冻结挂单
		if len(tokenTx.TokenVinRecv) == 0 {
			addr := crypto.BuildAddr(config.AddrPre, tokenTx.PukPromoter)
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.TokenTxidOut, addr))
			value := new(big.Int).Set(tokenTx.AmountOut)

			if addrToken, ok := this.addrsToken.LoadOrStore(key, &AddrToken{
				TokenId: tokenTx.TokenTxidOut,
				Addr:    addr,
				Value:   value,
			}); ok {
				addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
			} else {
			}
			return
		}

		//只冻结吃单
		addr := tokenTx.Vin[0].GetPukToAddr()
		key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.TokenTxidIn, *addr))
		value := new(big.Int).Set(tokenTx.AmountDeal)

		if addrToken, ok := this.addrsToken.LoadOrStore(key, &AddrToken{
			TokenId: tokenTx.TokenTxidIn,
			Addr:    *addr,
			Value:   value,
		}); ok {
			addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
		} else {
		}
		return

	}
}

func (this *UnpackedTransaction) FindAddrNonce(addr *crypto.AddressCoin) big.Int {
	// engine.Log.Info("FindAddrNonce")
	tsrItr, ok := this.addrs.Load(utils.Bytes2string(*addr))
	if ok {
		// engine.Log.Info("FindAddrNonce")
		tsr := tsrItr.(*TransactionsRatio)
		// // engine.Log.Info("FindAddrNonce:%d", len(tsr.trs))
		// trOne := tsr.trs[len(tsr.trs)-1]
		// // engine.Log.Info("FindAddrNonce")
		// // vins := tsr.trs[len(tsr.trs)-1].tx.GetVin()
		// vins := trOne.tx.GetVin()
		// // engine.Log.Info("FindAddrNonce:%d", len(tsr.trs))
		// nonce := (*vins)[0].Nonce
		// // engine.Log.Info("FindAddrNonce:%d", len(tsr.trs))
		return tsr.GetNonce()
	} else {
		// engine.Log.Info("FindAddrNonce")
		return big.Int{}
	}
}

/*
在这些临时交易中查找某个地址花费金额
*/
func (this *UnpackedTransaction) FindAddrSpend(addr *crypto.AddressCoin) (uint64, uint64) {
	tsrItr, ok := this.addrs.Load(utils.Bytes2string(*addr))
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		return tsr.spend, tsr.spendLock
	} else {
		return 0, 0
	}
}

/*
在这些未打包交易中查找某个地址Token冻结金额
*/
func (this *UnpackedTransaction) getTokenFrozenBalance(tokenId []byte, addr crypto.AddressCoin) uint64 {
	key := config.BuildDBKeyTokenAddrFrozenValue(tokenId, addr)
	addrToken, ok := this.addrsToken.Load(utils.Bytes2string(*key))
	if ok {
		return addrToken.(*AddrToken).Value.Uint64()
	} else {
		return 0
	}
}

/*
删除一个交易
*/
func (this *UnpackedTransaction) DelTx(tx TxItr) {
	addr := (*tx.GetVin())[0].GetPukToAddr()
	addrStr := utils.Bytes2string(*addr)
	tsrItr, ok := this.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		tsr.DelTx(tx.GetHash())
		//删除Token的交易金额
		this.delTx4TokenPay(tx)

		//清除缓存和限流
		DelCacheTxAndLimit(utils.Bytes2string(*tx.GetHash()), tx)
		// tsr.lock.Lock()
		// for i, one := range tsr.trs {
		// 	if bytes.Equal(*one.tx.GetHash(), *tx.GetHash()) {
		// 		tsr.gas -= one.gas
		// 		tsr.size -= one.size
		// 		tsr.spend -= one.spend
		// 		tsr.spendLock -= one.spendLock
		// 		temp := tsr.trs[:i]
		// 		tsr.trs = append(temp, tsr.trs[i+1:]...)
		// 	}
		// }
		// tsr.lock.Unlock()
		//如果交易数量为0，则删除这个地址的记录
		if tsr.TrsLen() <= 0 {
			this.addrs.Delete(addrStr)
		}
		//return
	} else {
		//return
	}
}

/*
删除Token的未打包交易金额,也有可能是部分金额
*/
func (this *UnpackedTransaction) delTx4TokenPay(tx TxItr) {
	switch tx.Class() {
	case config.Wallet_tx_type_token_publish:
		//note: 不处理
	case config.Wallet_tx_type_token_payment:
		tokenTx := tx.(*TxTokenPay)
		addr := tokenTx.Token_Vin[0].GetPukToAddr()
		key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.Token_txid, *addr))
		if item, ok := this.addrsToken.Load(key); ok {
			addrToken := item.(*AddrToken)
			value := uint64(0)
			for _, vout := range tokenTx.Token_Vout {
				value += vout.Value
			}

			//减去部分或删除金额
			if addrToken.Value.Uint64() > value {
				addrToken.Value.Sub(addrToken.Value, new(big.Int).SetUint64(value))
			} else {
				this.addrsToken.Delete(key)
			}
		}
	case config.Wallet_tx_type_token_new_order:
		newOrder := tx.(*TxTokenNewOrder)

		//当Buy=true代表买入A，vin用于证明B资产。
		//当Buy=false代表卖出A，vin用于证明A资产。
		addr := newOrder.TokenVin.GetPukToAddr()
		if newOrder.Buy {
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(newOrder.TokenBID, *addr))
			if item, ok := this.addrsToken.Load(key); ok {
				addrToken := item.(*AddrToken)
				value := new(big.Int).Set(newOrder.TokenBAmount)

				//减去部分或删除金额
				if addrToken.Value.Cmp(value) > 0 {
					addrToken.Value.Sub(addrToken.Value, value)
				} else {
					this.addrsToken.Delete(key)
				}
			}
		} else {
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(newOrder.TokenAID, *addr))
			if item, ok := this.addrsToken.Load(key); ok {
				addrToken := item.(*AddrToken)
				value := new(big.Int).Set(newOrder.TokenAAmount)

				//减去部分或删除金额
				if addrToken.Value.Cmp(value) > 0 {
					addrToken.Value.Sub(addrToken.Value, value)
				} else {
					this.addrsToken.Delete(key)
				}
			}
		}

	case config.Wallet_tx_type_token_swap_order:
		//tokenOrder := tx.(*TxTokenSwapOrder)
		//for _, tokenPair := range tokenOrder.TokenPairs {
		//	//持有币种
		//	addr := tokenPair.TokenVout[0].Address
		//	key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenPair.TokenId, addr))
		//	if item, ok := this.addrsToken.Load(key); ok {
		//		addrToken := item.(*AddrToken)
		//		value := uint64(0)
		//		for _, vout := range tokenPair.TokenVout {
		//			value += vout.Value
		//		}

		//		//减去部分或删除金额
		//		if addrToken.Value.Uint64() > value {
		//			addrToken.Value.Sub(addrToken.Value, new(big.Int).SetUint64(value))
		//		} else {
		//			this.addrsToken.Delete(key)
		//		}
		//	}

		//	//交换币种
		//	swapAddr := tokenPair.SwapTokenVout[0].Address
		//	swapKey := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenPair.SwapTokenId, swapAddr))
		//	if item, ok := this.addrsToken.Load(swapKey); ok {
		//		addrToken := item.(*AddrToken)
		//		value := uint64(0)
		//		for _, vout := range tokenPair.SwapTokenVout {
		//			value += vout.Value
		//		}

		//		//减去部分或删除金额
		//		if addrToken.Value.Uint64() > value {
		//			addrToken.Value.Sub(addrToken.Value, new(big.Int).SetUint64(value))
		//		} else {
		//			this.addrsToken.Delete(swapKey)
		//		}
		//	}
		//}
	case config.Wallet_tx_type_token_cancel_order:
		//tokenOrder := tx.(*TxTokenSwapOrder)
		//for _, tokenPair := range tokenOrder.TokenPairs {
		//	addr := tokenPair.TokenVin[0].GetPukToAddr()
		//	key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenPair.TokenId, *addr))
		//	this.addrsToken.Delete(key)
		//}
	case config.Wallet_tx_type_swap:
		tokenTx := tx.(*TxSwap)

		//只解冻挂单
		if len(tokenTx.TokenVinRecv) == 0 {
			addr := crypto.BuildAddr(config.AddrPre, tokenTx.PukPromoter)
			key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.TokenTxidOut, addr))
			if item, ok := this.addrsToken.Load(key); ok {
				addrToken := item.(*AddrToken)
				value := tokenTx.AmountOut.Uint64()

				//减去部分或删除金额
				if addrToken.Value.Uint64() > value {
					addrToken.Value.Sub(addrToken.Value, new(big.Int).SetUint64(value))
				} else {
					this.addrsToken.Delete(key)
				}
			}
			return
		}

		addr := crypto.BuildAddr(config.AddrPre, tokenTx.PukPromoter)
		key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.TokenTxidOut, addr))
		if item, ok := this.addrsToken.Load(key); ok {
			addrToken := item.(*AddrToken)
			value := CalcDstAmount(tokenTx.AmountOut, tokenTx.AmountIn, tokenTx.AmountDeal)

			//减去部分或删除金额
			if addrToken.Value.Uint64() > value.Uint64() {
				addrToken.Value.Sub(addrToken.Value, value)
			} else {
				this.addrsToken.Delete(key)
			}
		}

		addrRecv := tokenTx.Vin[0].GetPukToAddr()
		keyRecv := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tokenTx.TokenTxidIn, *addrRecv))
		if item, ok := this.addrsToken.Load(keyRecv); ok {
			addrToken := item.(*AddrToken)
			value := tokenTx.AmountDeal

			//减去部分或删除金额
			if addrToken.Value.Uint64() > value.Uint64() {
				addrToken.Value.Sub(addrToken.Value, value)
			} else {
				this.addrsToken.Delete(keyRecv)
			}
		}

		return
	}
}

func (this *UnpackedTransaction) FindTx() *[]*TransactionsRatio {
	// engine.Log.Info("查找未打包的交易")
	tsrs := make([]*TransactionsRatio, 0)
	this.addrs.Range(func(k, v interface{}) bool {
		tsr := v.(*TransactionsRatio)
		// sort.Sort()
		tsrs = append(tsrs, tsr)
		return true
	})
	return &tsrs
}

/*
查询一个交易是否存在
*/
func (this *UnpackedTransaction) ExistTxByAddrTxid(tx TxItr) bool {
	fromAddr := (*tx.GetVin())[0].GetPukToAddr()
	addrStr := utils.Bytes2string(*fromAddr)
	tsrItr, ok := this.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		txhashkey := config.BuildBlockTx(*tx.GetHash())
		if tsr.FindTxByHash(&txhashkey) {
			return true
		}
		// for _, one := range tsr.trs {
		// 	if bytes.Equal(*one.tx.GetHash(), *tx.GetHash()) {
		// 		return true
		// 	}
		// }
	}
	return false
}

/*
清理超时不能上链的交易
*/
func (this *UnpackedTransaction) CleanOverTimeTx(height uint64) [][]byte {
	hashs := make([][]byte, 0)
	this.addrs.Range(func(k, v interface{}) bool {
		tsr := v.(*TransactionsRatio)
		for _, one := range tsr.GetTrs() {
			if one.tx.GetLockHeight() > height {
				continue
			}
			tsr.DelTx(one.tx.GetHash())
			//删除Token的交易金额
			this.delTx4TokenPay(one.tx)
			hashs = append(hashs, *one.tx.GetHash())

			//清除缓存和限流
			DelCacheTxAndLimit(hex.EncodeToString(*one.tx.GetHash()), one.tx)
			// temp := tsr.trs[:index]
			// tsr.trs = append(temp, tsr.trs[index+1:]...)
		}
		//如果交易数量为0，则删除这个地址的记录
		if tsr.TrsLen() <= 0 {
			this.addrs.Delete(k)
		}
		return true
	})

	//处理撮合交易超时未上链
	itr := oBook.txCachePool.Iterator()
	for itr.Next() {
		swapTx := itr.Value().(*swapTxCache).TxTokenSwapOrder
		if swapTx.GetLockHeight() <= height {
			oBook.DelTxCache(swapTx)
		}
	}

	return hashs
}

func NewUnpackedTransaction() *UnpackedTransaction {
	ut := UnpackedTransaction{
		addrs:      new(sync.Map),
		addrsToken: new(sync.Map),
		// txs:   new(sync.Map),
	}
	return &ut
}
