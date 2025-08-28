package main

import (
	"bytes"
	"encoding/hex"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"web3_gui/chain/db"
	"web3_gui/chain/mining"
	_ "web3_gui/chain/mining/token/payment"
	_ "web3_gui/chain/mining/token/publish"
	_ "web3_gui/chain/mining/tx_name_in"
	_ "web3_gui/chain/mining/tx_name_out"

	"web3_gui/chain/config"
	"web3_gui/chain/sqlite3_db"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	txItem_status_notSpent = int32(0) //未花费的交易余额，可以正常支付
	txItem_status_frozen   = int32(1) //锁仓,区块达到指定高度才能使用
	txItem_status_lock     = int32(2) //冻结高度，指定高度还未上链，则转为未花费的交易

	// addrOne = "MMSKLRCKRLDzBFXJUxT9LJm8VaSBaZmirVDo4"
	addrOne = "MMS3E1LDivNFVBrqoPMcb3CXM291ikAuPoYJ4"
)

// 截至到现在高度，计算释放的余额
var maxHeight = uint64(0)

var addrSpecial = crypto.AddressFromB58String(addrOne)

func main() {
	Start()
}

/*
统计链上所有地址余额
保存在数据库里面
*/
func Start() {
	// utils.PprofMem()
	// sqlite.Init()
	// path :=

	engine.SetLogPath("logs/log.txt")
	config.AddrPre = "MMS"
	CountAddrBalance(filepath.Join("wallet", "data"))

	// PrintBalanceMap(m, tim, timOld)

	//等待3秒钟再关闭，让sqlite数据库处理完
	time.Sleep(time.Second * 3)
}

/*
统计链上所有地址余额
*/
func CountAddrBalance(dir string) map[string]*map[string]*mining.TxItem {

	db.InitDB(config.DB_path, config.DB_path_temp)
	sqlite3_db.Init()
	// timOld := mining.NewTxItemManagerOld()
	// tim := mining.NewTxItemManager()

	// NotSpentBalance := make(map[string]*map[string]*mining.TxItem) //保存各个状态的txitem，解锁、冻结等只是状态的改变。key:string=收款地址;value:*sync.Map(key:string=[txid]_[voutIndex];value:*TxItem=TxItem;)=;
	// NotSpentBalanceHex := make(map[string]*map[string]*mining.TxItem) //保存各个状态的txitem，解锁、冻结等只是状态的改变。key:string=收款地址;value:*sync.Map(key:string=[txid]_[voutIndex];value:*TxItem=TxItem;)=;

	// txItemKey := make(map[string]uint64)

	// balanceMap := make(map[string]uint64)
	balanceNotSpend := make(map[string]uint64)
	balanceFrozen := make(map[string]uint64)

	// db := utils.CreateLevelDB(dir)

	// db.InitDB(dir)
	beforBlockHash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return nil
	}

	maxHeight = db.GetHighstBlock()
	// maxHeight = 93888
	engine.Log.Info("统计到区块高度:%d", maxHeight)

	rewardTotal := uint64(0)             //挖矿奖励总合
	balanceCirculationTotal := uint64(0) //流通量
	balanceFrozenTotal := uint64(0)      //锁仓量

	// beforGroupHeight := uint64(0)

	for beforBlockHash != nil {
		runtime.GC()
		bhvo, err := mining.LoadBlockHeadVOByHash(beforBlockHash)
		if err != nil {
			engine.Log.Info("查询第 个块错误:%s", err.Error())
			return nil
		}

		engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))

		for _, txBase := range bhvo.Txs {

			isExclude := false
			//排除上链但不合法的交易
			for _, one := range config.Exclude_Tx {
				if bhvo.BH.Height != one.Height {
					continue
				}

				if bytes.Equal(one.TxByte, *txBase.GetHash()) {
					// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
					// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
					isExclude = true
					break
				}
			}
			if isExclude {
				continue
			}

			var payAddr *crypto.AddressCoin
			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				for _, one := range *txBase.GetVout() {
					rewardTotal += one.Value
				}
				engine.Log.Info("区块奖励累计 %d", rewardTotal)
			} else {
				//计算所有交易中的地址余额
				// engine.Log.Info("本区块高度:%d", bhvo.BH.Height)
				//第一个区块中的见证人质押没有vin
				if bhvo.BH.Height != 1 {
					vinOne := (*txBase.GetVin())[0]
					payAddr = vinOne.GetPukToAddr()
					// engine.Log.Info("扣款地址:%v", payAddr)
				}
			}

			valueTotal := uint64(0)
			//计算输出中的余额
			for voutIndex, _ := range *txBase.GetVout() {
				one := (*txBase.GetVout())[voutIndex]
				valueTotal += one.Value
				if one.Value > 100000000*100000000 {
					panic("交易hash:" + hex.EncodeToString(*txBase.GetHash()) + " voutIndex:" + strconv.Itoa(voutIndex) + " value:" + strconv.Itoa(int(one.Value)))
				}
				if one.FrozenHeight > maxHeight {
					balanceFrozenTotal += one.Value
					oldValue, ok := balanceFrozen[utils.Bytes2string(one.Address)]
					if ok {
						balanceFrozen[utils.Bytes2string(one.Address)] = oldValue + one.Value
					} else {
						balanceFrozen[utils.Bytes2string(one.Address)] = one.Value
					}
				} else {
					balanceCirculationTotal += one.Value
					oldValue, ok := balanceNotSpend[utils.Bytes2string(one.Address)]
					if ok {
						balanceNotSpend[utils.Bytes2string(one.Address)] = oldValue + one.Value
					} else {
						balanceNotSpend[utils.Bytes2string(one.Address)] = one.Value
					}
				}
			}
			// engine.Log.Info("扣款地址:%v", payAddr)
			//余额中减去
			if payAddr != nil {
				oldValue, ok := balanceNotSpend[utils.Bytes2string(*payAddr)]
				if ok {
					engine.Log.Info("余额中减去:%s", payAddr.B58String())
					balanceNotSpend[utils.Bytes2string(*payAddr)] = oldValue + valueTotal
				} else {
					panic("未找到这个地址的余额:" + payAddr.B58String())
				}
			}

		}

		if bhvo.BH.Height == maxHeight {
			break
		}

		if bhvo.BH.Nextblockhash != nil && len(bhvo.BH.Nextblockhash) > 0 {
			beforBlockHash = &bhvo.BH.Nextblockhash
		} else {
			beforBlockHash = nil
		}

	}
	engine.Log.Info("挖矿奖励总量:%d 链上共流通币数量:%d", rewardTotal, balanceCirculationTotal)

	for key, value := range balanceNotSpend {

		frozenValue := balanceFrozen[key]

		addr := crypto.AddressCoin([]byte(key))
		engine.Log.Info("地址:%s  可用余额:%d  锁定:%d", addr.B58String(), value, frozenValue)
	}

	// tim.FindBalanceAll()

	return nil
}

// func PrintBalanceMap(notSpentBalance map[string]*map[string]*mining.TxItem, tim *mining.TxItemManager, timOld *mining.TxItemManagerOld) {
// 	// if balances == nil {
// 	// 	engine.Log.Info("打印结果为空")
// 	// 	return
// 	// }
// 	// for k, v := range balances {
// 	// 	addrCoin := crypto.AddressCoin([]byte(k))
// 	// 	engine.Log.Info("%s %d", addrCoin.B58String(), v)
// 	// }

// 	for addrStr, items := range notSpentBalance {
// 		addrCoin := crypto.AddressCoin([]byte(addrStr))

// 		notSpend := uint64(0)
// 		frozen := uint64(0)
// 		for _, item := range *items {
// 			if item.Status == txItem_status_frozen {
// 				notSpend += item.Value
// 			} else {
// 				frozen += item.Value
// 			}
// 		}

// 		//查询对比
// 		notSpend2 := uint64(0)
// 		frozen2 := uint64(0)
// 		items := tim.FindBalanceNotSpent(addrCoin)
// 		for _, one := range items {
// 			notSpend2 += one.Value
// 		}
// 		items = tim.FindBalanceFrozen(addrCoin)
// 		for _, one := range items {
// 			frozen2 += one.Value
// 		}

// 		engine.Log.Info("地址:%s 可用:%d 锁定:%d  可用2:%d 锁定2:%d", addrCoin.B58String(), notSpend, frozen, notSpend2, frozen2)
// 	}

// 	// engine.Log.Info("打印已经解锁的余额")

// 	// if balanceNotSpend == nil {
// 	// 	engine.Log.Info("打印结果为空")
// 	// 	return
// 	// }
// 	// for k, v := range balanceNotSpend {
// 	// 	addrCoin := crypto.AddressCoin([]byte(k))
// 	// 	engine.Log.Info("%s %d", addrCoin.B58String(), v)
// 	// }

// 	// engine.Log.Info("打印 TxItemManager")
// 	// for k, v := range balances {
// 	// 	addrOne := crypto.AddressCoin([]byte(k))
// 	// 	notSpent1 := uint64(0)
// 	// 	txItems := tim.FindBalance(addrOne)
// 	// 	for _, one := range txItems {
// 	// 		notSpent1 += one.Value
// 	// 	}
// 	// 	notSpent2 := uint64(0)
// 	// 	if timOld != nil {
// 	// 		txItems = timOld.FindBalance(addrOne)
// 	// 		for _, one := range txItems {
// 	// 			notSpent2 += one.Value
// 	// 		}
// 	// 	}

// 	// 	notF1 := uint64(0)
// 	// 	txItems = tim.FindBalanceFrozen(addrOne)
// 	// 	for _, one := range txItems {
// 	// 		notF1 += one.Value
// 	// 	}
// 	// 	notF2 := uint64(0)
// 	// 	if timOld != nil {
// 	// 		txItems = timOld.FindBalanceFrozen(addrOne)
// 	// 		for _, one := range txItems {
// 	// 			notF2 += one.Value
// 	// 		}
// 	// 	}

// 	// 	engine.Log.Info("地址:%s 总余额:%d 可用余额1:%d 可用余额2:%d 锁仓余额1:%d 锁仓余额2:%d", addrOne.B58String(), v, notSpent1, notSpent2, notF1, notF2)
// 	// }

// }

// var CountAddrCoin = crypto.AddressFromB58String("MMSArGrVxQwYQDCoxWVyw9J5xGVWYZVa6ZBE4")

// func CountBalances(bhvo *mining.BlockHeadVO) mining.TxItemCount {

// 	//将txitem集中起来，一次性添加
// 	itemCount := mining.TxItemCount{
// 		Additems: make([]*mining.TxItem, 0),
// 		SubItems: make([]*mining.TxSubItems, 0),
// 		// deleteKey: make([]string, 0),
// 	}
// 	itemsChan := make(chan *mining.TxItemCount, len(bhvo.Txs))
// 	// engine.Log.Info("添加%d个交易协程", len(bhvo.Txs))
// 	wg := new(sync.WaitGroup)
// 	wg.Add(len(bhvo.Txs))
// 	go func() {
// 		for i := 0; i < len(bhvo.Txs); i++ {
// 			// engine.Log.Info("等待一个协程返回")
// 			one := <-itemsChan
// 			// engine.Log.Info("一个协程返回了")
// 			if one != nil {
// 				itemCount.Additems = append(itemCount.Additems, one.Additems...)
// 				itemCount.SubItems = append(itemCount.SubItems, one.SubItems...)
// 				// itemCount.deleteKey = append(itemCount.deleteKey, one.deleteKey...)
// 			}
// 			wg.Done()
// 		}
// 	}()

// 	//查询排除的交易
// 	// excludeTx := make([]config.ExcludeTx, 0)
// 	// for i, one := range config.Exclude_Tx {
// 	// 	if bhvo.BH.Height == one.Height {
// 	// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
// 	// 	}
// 	// }

// 	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
// 	for _, txItr := range bhvo.Txs {
// 		//排除的交易不统计
// 		// for _, two := range excludeTx {
// 		// 	if bytes.Equal(two.TxByte, *txItr.GetHash()) {
// 		// 		continue
// 		// 	}
// 		// }
// 		go countBalancesTxOne(txItr, bhvo.BH.Height, NumCPUTokenChan, itemsChan)
// 	}

// 	wg.Wait()
// 	// engine.Log.Info("协程完成统计")
// 	// start := time.Now()
// 	// this.notspentBalance.CountTxItem(itemCount)
// 	// engine.Log.Info("统计交易 耗时 %s", time.Now().Sub(start))
// 	return itemCount
// }

// /*
// 	统计单个交易余额，方便异步统计
// */
// func countBalancesTxOne(txItr mining.TxItr, height uint64, tokenCPU chan bool, itemChan chan *mining.TxItemCount) {
// 	tokenCPU <- false
// 	// engine.Log.Info("开始统计")
// 	// start := time.Now()

// 	// start := time.Now()
// 	txItr.BuildHash()
// 	// txHashStr := txItr.GetHashStr()
// 	//将之前的UTXO标记为已经使用，余额中减去。
// 	itemCount := txItr.CountTxItems(height)
// 	// engine.Log.Info("统计余额及奖励 101010 耗时 %s", time.Now().Sub(start))
// 	itemChan <- itemCount
// 	// engine.Log.Info("统单易6耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
// 	<-tokenCPU
// }
