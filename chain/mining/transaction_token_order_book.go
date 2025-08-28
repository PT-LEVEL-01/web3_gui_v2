package mining

import (
	"bytes"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"sort"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/utils"
)

const (
	OrderStateNone   = iota
	OrderStateNormal //正常
	OrderStateDone   //完成
	OrderStateCancel //取消
)

var oBook *orderBookEngine

// 订单薄引擎
type orderBookEngine struct {
	Pools       *sync.Map          //订单池; key:代币对,value:*arraylist.List
	Orders      *sync.Map          //订单池; key:订单ID,value:*go_protos.OrderInfo
	NewOrders   *linkedhashmap.Map //新订单,用于判断对手价
	txCachePool *linkedhashmap.Map //撮合交易缓存池; key:自定key,value:swapTxCache
}

type swapTxCache struct {
	IsPackage bool
	*TxTokenSwapOrder
}

// 订单薄引擎启动
func OrderBookEngineSetup() {
	oBook = &orderBookEngine{
		Pools:       &sync.Map{},
		Orders:      &sync.Map{},
		NewOrders:   linkedhashmap.New(),
		txCachePool: linkedhashmap.New(),
	}

	iter := db.LevelDB.NewIteratorWithOption(util.BytesPrefix(config.DBKEY_token_order), nil)
	for iter.Next() {
		orderInfo := &go_protos.OrderInfo{}
		if err := orderInfo.Unmarshal(iter.Value()); err != nil {
			continue
		}

		//恢复订单池
		if items, ok := oBook.Pools.LoadOrStore(oBook.GetPookKey(orderInfo.TokenAID, orderInfo.TokenBID), arraylist.New(orderInfo)); ok {
			list := items.(*arraylist.List)
			list.Add(orderInfo)
		}

		//恢复订单索引
		key := config.BuildDBKeyTokenOrder(orderInfo.OrderId)
		oBook.Orders.Store(utils.Bytes2string(*key), orderInfo)
	}
	iter.Error()
	iter.Release()
}

// 新订单
func (o *orderBookEngine) NewOrder(tx *TxTokenNewOrder) []*AddrToken {
	txIds := make([][]byte, 0)
	txIds = append(txIds, tx.Hash)
	addr := tx.TokenVin.GetPukToAddr()

	lockedTokens := []*AddrToken{}

	//添加到订单池
	orderInfo := &go_protos.OrderInfo{
		OrderId:             tx.Hash,
		TokenAID:            tx.TokenAID,
		TokenAAmount:        tx.TokenAAmount.Bytes(),
		TokenBID:            tx.TokenBID,
		TokenBAmount:        tx.TokenBAmount.Bytes(),
		Address:             *addr,
		Buy:                 tx.Buy,
		Price:               tx.Price,
		State:               OrderStateNormal,
		TxIds:               txIds,
		PendingTokenAAmount: tx.TokenAAmount.Bytes(),
	}

	pookKey := o.GetPookKey(tx.TokenAID, tx.TokenBID)
	if items, ok := o.Pools.LoadOrStore(pookKey, arraylist.New(orderInfo)); ok {
		list := items.(*arraylist.List)
		list.Add(orderInfo)
	}

	//订单索引
	key := config.BuildDBKeyTokenOrder(tx.Hash)
	o.Orders.Store(utils.Bytes2string(*key), orderInfo)

	//标记新订单
	o.NewOrders.Put(utils.Bytes2string(orderInfo.OrderId), o.NewOrders.Size()+1)

	//锁定金额集合
	if tx.Buy {
		lockedTokens = append(lockedTokens, &AddrToken{
			TokenId: tx.TokenBID,
			Addr:    *addr,
			Value:   tx.TokenBAmount,
		})
	} else {
		lockedTokens = append(lockedTokens, &AddrToken{
			TokenId: tx.TokenAID,
			Addr:    *addr,
			Value:   tx.TokenAAmount,
		})
	}

	//持久化订单
	bs, _ := orderInfo.Marshal()
	db.LevelDB.Save(*key, &bs)

	return lockedTokens
}

func (o *orderBookEngine) GetPookKey(tokenAID, tokenBID []byte) string {
	if tokenAID == nil || len(tokenAID) == 0 {
		tokenAID = []byte{0}
	}
	if tokenBID == nil || len(tokenBID) == 0 {
		tokenBID = []byte{0}
	}
	pookKey := make([]byte, 0, len(tokenAID)+len(tokenBID))
	pookKey = append(tokenAID, tokenBID...)
	return utils.Bytes2string(pookKey)
}

func GetTokenOrder(orderId []byte) (*go_protos.OrderInfo, error) {
	key := config.BuildDBKeyTokenOrder(orderId)
	bs, err := db.LevelDB.Get(*key)
	if err != nil || bs == nil {
		return nil, errors.New("not found")
	}

	info := &go_protos.OrderInfo{}
	if err := proto.Unmarshal(bs, info); err != nil {
		return nil, err
	}

	return info, nil
}

func GetTokenOrderPool() map[string][]interface{} {
	data := map[string][]interface{}{}
	oBook.Pools.Range(func(key, value any) bool {
		list := value.(*arraylist.List)
		values := list.Values()
		if item, ok := data[key.(string)]; ok {
			item = append(item, values...)
		} else {
			data[key.(string)] = values
		}
		return true
	})

	return data
}

/*
更新订单:
OrderID1        //买订单ID
OrderID2       //卖订单ID
tokenAAmount=双方最小TokenA
tokenBAmount=按卖单价计算TokenB
tokenBRefund=买单返还TokenB

处理订单:
买单: TokenAAmount=Old-tokenAAmount; TokenBAmount=Old-(tokenBAmount+tokenBRefund)
卖单: TokenAAmount=Old-tokenAAmount; TokenBAmount=Old-tokenBAmount

余额处理
买单: TokenA:+tokenAAmount; TokenB:+(tokenBAmount+tokenBRefund)
卖单: TokenAAmount=Old-tokenAAmount; TokenBAmount=Old-tokenBAmount

TokenRefund=amount*(price1-price)
*/
//func (o *orderBookEngine) UpdateOrder(swapTx *TxTokenSwapOrder) ([]*AddrToken, []*AddrToken) {
//	updateTokens := []*AddrToken{}   //+-余额
//	unlockedTokens := []*AddrToken{} //-锁定
//
//	buyInfoItr, ok := o.Orders.Load(utils.Bytes2string(*config.BuildDBKeyTokenOrder(swapTx.OrderID1)))
//	if !ok {
//		return updateTokens, unlockedTokens
//	}
//	buyInfo := buyInfoItr.(*go_protos.OrderInfo)
//
//	sellInfoItr, ok := o.Orders.Load(utils.Bytes2string(*config.BuildDBKeyTokenOrder(swapTx.OrderID2)))
//	if !ok {
//		return updateTokens, unlockedTokens
//	}
//	sellInfo := sellInfoItr.(*go_protos.OrderInfo)
//
//	tokenAAmount, tokenBAmount, tokenBRefund, err := CalcSwapTokenOrderAmount(
//		new(big.Int).SetBytes(buyInfo.TokenAAmount),
//		new(big.Int).SetBytes(buyInfo.TokenBAmount),
//		new(big.Int).SetBytes(sellInfo.TokenAAmount),
//		new(big.Int).SetBytes(sellInfo.TokenBAmount),
//	)
//	if err != nil {
//		return updateTokens, unlockedTokens
//	}
//
//	//============= 处理买单 =============
//	{
//		//更新订单
//		buyInfo.TokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(buyInfo.TokenAAmount), tokenAAmount).Bytes()
//		updateTokens = append(updateTokens, &AddrToken{
//			TokenId: buyInfo.TokenAID,
//			Addr:    buyInfo.Address,
//			Value:   tokenAAmount,
//		})
//
//		tmpTokenBAmount := new(big.Int).Add(tokenBAmount, tokenBRefund)
//		buyInfo.TokenBAmount = new(big.Int).Sub(new(big.Int).SetBytes(buyInfo.TokenBAmount), tmpTokenBAmount).Bytes()
//		updateTokens = append(updateTokens, &AddrToken{
//			TokenId: buyInfo.TokenBID,
//			Addr:    buyInfo.Address,
//			Value:   new(big.Int).Neg(tmpTokenBAmount),
//		})
//		unlockedTokens = append(unlockedTokens, &AddrToken{
//			TokenId: buyInfo.TokenBID,
//			Addr:    buyInfo.Address,
//			Value:   tmpTokenBAmount,
//		})
//
//		//返还
//		if tokenBRefund.Cmp(big.NewInt(0)) == 1 {
//			updateTokens = append(updateTokens, &AddrToken{
//				TokenId: buyInfo.TokenBID,
//				Addr:    buyInfo.Address,
//				Value:   tokenBRefund,
//			})
//		}
//
//		//检查订单是否完成,更新订单状态
//		buyInfo.State = OrderStateNormal
//		if new(big.Int).SetBytes(buyInfo.TokenAAmount).Cmp(big.NewInt(0)) == 0 || new(big.Int).SetBytes(buyInfo.TokenBAmount).Cmp(big.NewInt(0)) == 0 {
//			buyInfo.State = OrderStateDone
//			remainTokenB := new(big.Int).SetBytes(buyInfo.TokenBAmount)
//			//订单完成有剩余;返还Token
//			if remainTokenB.Cmp(big.NewInt(0)) == 1 {
//				updateTokens = append(updateTokens, &AddrToken{
//					TokenId: buyInfo.TokenBID,
//					Addr:    buyInfo.Address,
//					Value:   remainTokenB,
//				})
//			}
//			//移除完成的订单
//			o.Orders.Delete(utils.Bytes2string(*config.BuildDBKeyTokenOrder(buyInfo.OrderId)))
//		}
//
//		//持久化订单
//		buyInfo.TxIds = append(buyInfo.TxIds, swapTx.Hash)
//		bs, _ := buyInfo.Marshal()
//		key := config.BuildDBKeyTokenOrder(buyInfo.OrderId)
//		db.LevelDB.Save(*key, &bs)
//	}
//
//	//============= 处理卖单 =============
//	{
//		//更新订单
//		sellInfo.TokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(sellInfo.TokenAAmount), tokenAAmount).Bytes()
//		updateTokens = append(updateTokens, &AddrToken{
//			TokenId: sellInfo.TokenAID,
//			Addr:    sellInfo.Address,
//			Value:   new(big.Int).Neg(tokenAAmount),
//		})
//
//		unlockedTokens = append(unlockedTokens, &AddrToken{
//			TokenId: sellInfo.TokenAID,
//			Addr:    sellInfo.Address,
//			Value:   tokenAAmount,
//		})
//
//		sellInfo.TokenBAmount = new(big.Int).Sub(new(big.Int).SetBytes(sellInfo.TokenBAmount), tokenBAmount).Bytes()
//		updateTokens = append(updateTokens, &AddrToken{
//			TokenId: sellInfo.TokenBID,
//			Addr:    sellInfo.Address,
//			Value:   tokenBAmount,
//		})
//
//		//检查订单是否完成,更新订单状态
//		sellInfo.State = OrderStateNormal
//		if new(big.Int).SetBytes(sellInfo.TokenAAmount).Cmp(big.NewInt(0)) == 0 || new(big.Int).SetBytes(sellInfo.TokenBAmount).Cmp(big.NewInt(0)) == 0 {
//			sellInfo.State = OrderStateDone
//			//移除完成的订单
//			o.Orders.Delete(utils.Bytes2string(*config.BuildDBKeyTokenOrder(sellInfo.OrderId)))
//		}
//
//		//持久化订单
//		sellInfo.TxIds = append(sellInfo.TxIds, swapTx.Hash)
//		bs, _ := sellInfo.Marshal()
//		key := config.BuildDBKeyTokenOrder(sellInfo.OrderId)
//		db.LevelDB.Save(*key, &bs)
//	}
//
//	return updateTokens, unlockedTokens
//}

func (o *orderBookEngine) UpdateOrderV3(swapTx *TxTokenSwapOrder) ([]*AddrToken, []*AddrToken) {
	updateTokens := []*AddrToken{}   //+-余额
	unlockedTokens := []*AddrToken{} //-锁定

	order1Itr, ok := o.Orders.Load(utils.Bytes2string(*config.BuildDBKeyTokenOrder(swapTx.OrderID1)))
	if !ok {
		return updateTokens, unlockedTokens
	}
	order1 := order1Itr.(*go_protos.OrderInfo)
	buyInfo := &go_protos.OrderInfo{}
	sellInfo := &go_protos.OrderInfo{}
	if order1.Buy {
		buyInfo = order1
	} else {
		sellInfo = order1
	}

	order2Itr, ok := o.Orders.Load(utils.Bytes2string(*config.BuildDBKeyTokenOrder(swapTx.OrderID2)))
	if !ok {
		return updateTokens, unlockedTokens
	}
	order2 := order2Itr.(*go_protos.OrderInfo)
	if order2.Buy {
		buyInfo = order2
	} else {
		sellInfo = order2
	}

	tokenAAmount := new(big.Int).Set(swapTx.TokenAAmount)
	tokenBAmount := new(big.Int).Set(swapTx.TokenBAmount)

	//============= 处理买单 =============
	{
		//更新订单
		buyInfo.TokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(buyInfo.TokenAAmount), tokenAAmount).Bytes()
		updateTokens = append(updateTokens, &AddrToken{
			TokenId: buyInfo.TokenAID,
			Addr:    buyInfo.Address,
			Value:   tokenAAmount,
		})

		buyInfo.TokenBAmount = new(big.Int).Sub(new(big.Int).SetBytes(buyInfo.TokenBAmount), tokenBAmount).Bytes()
		updateTokens = append(updateTokens, &AddrToken{
			TokenId: buyInfo.TokenBID,
			Addr:    buyInfo.Address,
			Value:   new(big.Int).Neg(tokenBAmount),
		})
		unlockedTokens = append(unlockedTokens, &AddrToken{
			TokenId: buyInfo.TokenBID,
			Addr:    buyInfo.Address,
			Value:   tokenBAmount,
		})

		//检查订单是否完成,更新订单状态
		//if new(big.Int).SetBytes(buyInfo.TokenAAmount).Cmp(big.NewInt(0)) <= 0 || new(big.Int).SetBytes(buyInfo.TokenBAmount).Cmp(big.NewInt(0)) <= 0 {
		if new(big.Int).SetBytes(buyInfo.TokenAAmount).Cmp(big.NewInt(0)) <= 0 {
			remainTokenB := new(big.Int).SetBytes(buyInfo.TokenBAmount)
			//订单完成有剩余;返还Token
			if remainTokenB.Cmp(big.NewInt(0)) == 1 {
				unlockedTokens = append(unlockedTokens, &AddrToken{
					TokenId: buyInfo.TokenBID,
					Addr:    buyInfo.Address,
					Value:   remainTokenB,
				})
			}
			buyInfo.State = OrderStateDone
			//移除完成的订单
			o.Orders.Delete(utils.Bytes2string(*config.BuildDBKeyTokenOrder(buyInfo.OrderId)))
		}

		//持久化订单
		buyInfo.TxIds = append(buyInfo.TxIds, swapTx.Hash)
		bs, _ := buyInfo.Marshal()
		key := config.BuildDBKeyTokenOrder(buyInfo.OrderId)
		db.LevelDB.Save(*key, &bs)
	}

	//============= 处理卖单 =============
	{
		//更新订单
		sellInfo.TokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(sellInfo.TokenAAmount), tokenAAmount).Bytes()
		updateTokens = append(updateTokens, &AddrToken{
			TokenId: sellInfo.TokenAID,
			Addr:    sellInfo.Address,
			Value:   new(big.Int).Neg(tokenAAmount),
		})

		unlockedTokens = append(unlockedTokens, &AddrToken{
			TokenId: sellInfo.TokenAID,
			Addr:    sellInfo.Address,
			Value:   tokenAAmount,
		})

		sellInfo.TokenBAmount = new(big.Int).Sub(new(big.Int).SetBytes(sellInfo.TokenBAmount), tokenBAmount).Bytes()
		updateTokens = append(updateTokens, &AddrToken{
			TokenId: sellInfo.TokenBID,
			Addr:    sellInfo.Address,
			Value:   tokenBAmount,
		})

		//检查订单是否完成,更新订单状态
		//if new(big.Int).SetBytes(sellInfo.TokenAAmount).Cmp(big.NewInt(0)) <= 0 || new(big.Int).SetBytes(sellInfo.TokenBAmount).Cmp(big.NewInt(0)) <= 0 {
		if new(big.Int).SetBytes(sellInfo.TokenAAmount).Cmp(big.NewInt(0)) <= 0 {
			remainTokenA := new(big.Int).SetBytes(sellInfo.TokenAAmount)
			//订单完成有剩余;返还Token
			if remainTokenA.Cmp(big.NewInt(0)) == 1 {
				unlockedTokens = append(unlockedTokens, &AddrToken{
					TokenId: sellInfo.TokenAID,
					Addr:    sellInfo.Address,
					Value:   remainTokenA,
				})
			}
			sellInfo.State = OrderStateDone

			//移除完成的订单
			o.Orders.Delete(utils.Bytes2string(*config.BuildDBKeyTokenOrder(sellInfo.OrderId)))
		}

		//持久化订单
		sellInfo.TxIds = append(sellInfo.TxIds, swapTx.Hash)
		bs, _ := sellInfo.Marshal()
		key := config.BuildDBKeyTokenOrder(sellInfo.OrderId)
		db.LevelDB.Save(*key, &bs)
	}

	return updateTokens, unlockedTokens
}

func (o *orderBookEngine) CancelOrder(tx *TxTokenCancelOrder) []*AddrToken {
	unlockedTokens := []*AddrToken{} //-锁定
	for _, orderID := range tx.OrderIDs {
		//订单
		if item, ok := o.Orders.Load(utils.Bytes2string(*config.BuildDBKeyTokenOrder(orderID))); ok {
			orderInfo := item.(*go_protos.OrderInfo)
			if orderInfo.State != OrderStateNormal {
				continue
			}
			if orderInfo.Buy {
				unlockedTokens = append(unlockedTokens, &AddrToken{
					TokenId: orderInfo.TokenBID,
					Addr:    orderInfo.Address,
					Value:   new(big.Int).SetBytes(orderInfo.TokenBAmount),
				})
			} else {
				unlockedTokens = append(unlockedTokens, &AddrToken{
					TokenId: orderInfo.TokenAID,
					Addr:    orderInfo.Address,
					Value:   new(big.Int).SetBytes(orderInfo.TokenAAmount),
				})
			}

			//更新订单状态
			orderInfo.State = OrderStateCancel

			//移除完成的订单
			o.Orders.Delete(utils.Bytes2string(*config.BuildDBKeyTokenOrder(orderID)))

			//持久化订单
			orderInfo.TxIds = append(orderInfo.TxIds, tx.Hash)
			bs, _ := orderInfo.Marshal()
			key := config.BuildDBKeyTokenOrder(orderInfo.OrderId)
			db.LevelDB.Save(*key, &bs)
		}
	}
	return unlockedTokens
}

/*
运行订单薄撮合引擎
*/
func (o *orderBookEngine) Matching(chain *Chain) {
	if !chain.SyncBlockFinish {
		return
	}

	//处理挂单
	swapTxs := o.matchingRuleV3()

	//将撮合好交易放入缓存池中
	for _, swapTx := range swapTxs {
		key := swapTx.BuildSwapTxKey()
		_, ok := o.txCachePool.Get(key) //将撮合好交易放入缓存池中
		if !ok {
			o.txCachePool.Put(key, &swapTxCache{
				IsPackage:        false,
				TxTokenSwapOrder: swapTx,
			})
		}
	}
}

// 上链成功,删除缓存
func (o *orderBookEngine) RemoveTxCache(swapTx *TxTokenSwapOrder) {
	o.txCachePool.Remove(swapTx.BuildSwapTxKey())
}

// 上链超时,删除缓存,并恢复订单的数量
func (o *orderBookEngine) DelTxCache(swapTx *TxTokenSwapOrder) {
	o.txCachePool.Remove(swapTx.BuildSwapTxKey())

	//更新订单
	if item, ok := oBook.Orders.Load(utils.Bytes2string(swapTx.OrderID1)); ok {
		order1 := item.(*go_protos.OrderInfo)
		order1.PendingTokenAAmount = new(big.Int).Add(new(big.Int).SetBytes(order1.PendingTokenAAmount), swapTx.TokenAAmount).Bytes()
	}
	if item, ok := oBook.Orders.Load(utils.Bytes2string(swapTx.OrderID2)); ok {
		order2 := item.(*go_protos.OrderInfo)
		order2.PendingTokenAAmount = new(big.Int).Add(new(big.Int).SetBytes(order2.PendingTokenAAmount), swapTx.TokenAAmount).Bytes()
	}
}

/*
移除取消/完成的订单,池内交易按价格排序
*/
func (o *orderBookEngine) handlePool() {
	o.Pools.Range(func(key, value any) bool {
		list := value.(*arraylist.List)
		iter := list.Iterator()
		for iter.Next() {
			oi := iter.Value().(*go_protos.OrderInfo)
			if oi.State != OrderStateNormal {
				list.Remove(iter.Index())
			}
		}

		if list.Size() == 0 {
			o.Pools.Delete(key)
		}

		// 池内交易按价格稳定排序
		list.Sort(o.byPrice)

		return true
	})
}

/*
运行订单薄撮合引擎:
*/
func (o *orderBookEngine) SwapTxPackage() ([]TxItr, [][]byte) {
	txItrs := []TxItr{}
	ids := [][]byte{}

	//严格按顺序打包
	itr := o.txCachePool.Iterator() //打包
	for itr.Next() {
		txCache := itr.Value().(*swapTxCache)
		if !txCache.IsPackage {
			txItrs = append(txItrs, txCache.TxTokenSwapOrder)
			ids = append(ids, txCache.TxTokenSwapOrder.Hash)
			txCache.IsPackage = true
		}
	}

	return txItrs, ids
}

/*
==== 新订单 ====
交易对永远是TokenA-TokenB。
当Buy=true代表买入A，vin用于证明B资产。
当Buy=false代表卖出A，vin用于证明A资产。
价格计算方法：TokenBAmount/TokenAAmount。

==== 撮合订单 ====
1.买单价格>=卖单价格
2.成交数量: amount=双方最小TokenAAmount
3.成交价格: price=卖方价格
4.打包
5.统计
更新订单:
订单1: TokenAAmount=Old-amount; TokenBAmount=Old-(amount*price1)
订单2: TokenAAmount=Old-amount; TokenBAmount=Old-(amount*price)
更新账户:
订单1:锁定金额(B)=Old-(amount*price1); TokenBRefund=amount*(price1-price)
订单2:锁定金额(A)=Old-amount,TokenBAmount=Old-(amount*price)
*/
//func (o *orderBookEngine) matchingRule() []*TxTokenSwapOrder {
//	swapTxs := []*TxTokenSwapOrder{}
//	o.Pools.Range(func(key, value any) bool {
//		list := value.(*arraylist.List)
//		buys := arraylist.New()  //买单高到低
//		sells := arraylist.New() //卖单低到高
//		iter := list.Iterator()
//		for iter.Next() {
//			oi := iter.Value().(*go_protos.OrderInfo)
//			if oi.State == OrderStateNormal {
//				if oi.Buy {
//					buys.Insert(0, oi)
//				} else {
//					sells.Add(oi)
//				}
//			}
//		}
//
//		// ============ 开始撮合交易 ============
//		sellIter := sells.Iterator()
//		buyIter := buys.Iterator()
//		for buyIter.Next() {
//			if sellIter.Next() {
//				buy := buyIter.Value().(*go_protos.OrderInfo)
//				sell := sellIter.Value().(*go_protos.OrderInfo)
//				//买单价格>=卖单价格
//				if buy.Price >= sell.Price {
//					//匹配
//					tokenAAmount, tokenBAmount, _, err := CalcSwapTokenOrderAmount(
//						new(big.Int).SetBytes(buy.TokenAAmount),
//						new(big.Int).SetBytes(buy.TokenBAmount),
//						new(big.Int).SetBytes(sell.TokenAAmount),
//						new(big.Int).SetBytes(sell.TokenBAmount),
//					)
//					if err != nil {
//						continue
//					}
//					buildSwapTx, err := BuildSwapTokenOrderByWitness(buy.OrderId, sell.OrderId, tokenAAmount, tokenBAmount)
//					if err != nil {
//						continue
//					}
//					swapTxs = append([]*TxTokenSwapOrder{
//						buildSwapTx,
//					}, swapTxs...)
//				} else {
//					break //不可能有合适的价格了
//				}
//			} else {
//				break //没有更多的卖单了
//			}
//		}
//
//		return true
//	})
//
//	return swapTxs
//}
//

//func (o *orderBookEngine) matchingRuleV2() []*TxTokenSwapOrder {
//	swapTxs := []*TxTokenSwapOrder{}
//	poolKeys := []string{}
//	o.Pools.Range(func(key, value any) bool {
//		poolKeys = append(poolKeys, key.(string))
//		return true
//	})
//
//	//订单池排序
//	sort.Strings(poolKeys)
//
//	//严格按顺序撮合交易
//	for _, poolKey := range poolKeys {
//		value, ok := o.Pools.Load(poolKey)
//		if !ok {
//			continue
//		}
//
//		list := value.(*arraylist.List)
//		buys := arraylist.New()  //买单高到低
//		sells := arraylist.New() //卖单低到高
//		iter := list.Iterator()
//		for iter.Next() {
//			oi := iter.Value().(*go_protos.OrderInfo)
//			if oi.State == OrderStateNormal {
//				if oi.Buy {
//					//订单的副本
//					buys.Insert(0, oi)
//				} else {
//					//订单的副本
//					sells.Add(oi)
//				}
//			}
//		}
//
//		// ============ 开始撮合交易 ============
//		buyIter := buys.Iterator()
//		for buyIter.Next() {
//			buy := buyIter.Value().(*go_protos.OrderInfo)
//			sellIter := sells.Iterator()
//			for sellIter.Next() {
//				sell := sellIter.Value().(*go_protos.OrderInfo)
//				//过滤价格
//				if buy.Price < sell.Price {
//					continue
//				}
//
//				//计算
//				pendingBuyTokenAAmount := new(big.Int).SetBytes(buy.TokenAAmount)
//				pendingBuyTokenBAmount := new(big.Int).SetBytes(buy.TokenBAmount)
//				pendingSellTokenAAmount := new(big.Int).SetBytes(sell.TokenAAmount)
//				pendingSellTokenBAmount := new(big.Int).SetBytes(sell.TokenBAmount)
//				buyReplica := &orderReplica{
//					PendingTokenAAmount: new(big.Int),
//					PendingTokenBAmount: new(big.Int),
//				}
//				if riter, ok := o.Replica.LoadOrStore(utils.Bytes2string(buy.OrderId), buyReplica); ok {
//					r := riter.(*orderReplica)
//					pendingBuyTokenAAmount.Sub(pendingBuyTokenAAmount, r.PendingTokenAAmount)
//					pendingBuyTokenBAmount.Sub(pendingBuyTokenBAmount, r.PendingTokenBAmount)
//				}
//				if riter, ok := o.Replica.Load(utils.Bytes2string(sell.OrderId)); ok {
//					r := riter.(*orderReplica)
//					pendingSellTokenAAmount.Sub(pendingSellTokenAAmount, r.PendingTokenAAmount)
//					pendingSellTokenBAmount.Sub(pendingSellTokenBAmount, r.PendingTokenBAmount)
//				}
//
//				tokenAAmount, tokenBAmount, tokenBRefund, err := CalcSwapTokenOrderAmount(
//					pendingBuyTokenAAmount,
//					pendingBuyTokenBAmount,
//					pendingSellTokenAAmount,
//					pendingSellTokenBAmount,
//				)
//				if err != nil {
//					continue
//				}
//
//				//构建撮合交易
//				buildSwapTx, err := BuildSwapTokenOrderByWitness(buy.OrderId, sell.OrderId, tokenAAmount, tokenBAmount)
//				if err != nil {
//					continue
//				}
//				swapTxs = append(swapTxs, buildSwapTx)
//
//				//更新买单副本
//				buy.TmpTokenAAmount = new(big.Int).Sub(buy.TmpTokenAAmount, tokenAAmount)
//				tmpTokenBAmount := new(big.Int).Add(tokenBAmount, tokenBRefund)
//				buy.TmpTokenBAmount = new(big.Int).Sub(buy.TmpTokenBAmount, tmpTokenBAmount)
//				buy.OrderInfo.State = OrderStatePending
//
//				//更新卖单副本
//				sell.TmpTokenAAmount = new(big.Int).Sub(sell.TmpTokenAAmount, tokenAAmount)
//				sell.TmpTokenBAmount = new(big.Int).Sub(sell.TmpTokenBAmount, tokenBAmount)
//				sell.OrderInfo.State = OrderStatePending
//			}
//		}
//	}
//
//	return swapTxs
//}

//func (o *orderBookEngine) matchingRuleV3_Old() []*TxTokenSwapOrder {
//	swapTxs := []*TxTokenSwapOrder{}
//	poolKeys := []string{}
//	o.Pools.Range(func(key, value any) bool {
//		poolKeys = append(poolKeys, key.(string))
//		return true
//	})
//
//	//订单池子排序
//	sort.Strings(poolKeys)
//
//	//严格按顺序撮合交易
//	for _, poolKey := range poolKeys {
//		value, ok := o.Pools.Load(poolKey)
//		if !ok {
//			continue
//		}
//
//		list := value.(*arraylist.List)
//		iter := list.Iterator()
//		for iter.Next() {
//			//约定order1就是对手订单
//			order1 := iter.Value().(*go_protos.OrderInfo)
//			if order1.State != OrderStateNormal {
//				continue
//			}
//
//			list2 := list.Select(func(index int, value interface{}) bool {
//				if index > iter.Index() {
//					return true
//				}
//				return false
//			})
//			iter2 := list2.Iterator()
//			for iter2.Next() {
//				order2 := iter2.Value().(*go_protos.OrderInfo)
//				if order2.State != OrderStateNormal {
//					continue
//				}
//
//				if order1.Buy == order2.Buy {
//					continue
//				}
//
//				// ============ 开始撮合交易 ============
//				//买单大于卖单
//				//buy := &go_protos.OrderInfo{}
//				//sell := &go_protos.OrderInfo{}
//				//if order1.Buy {
//				//	buy = order1
//				//	sell = order2
//				//} else {
//				//	buy = order2
//				//	sell = order1
//				//}
//
//				//计算成交量
//				tokenAAmount, tokenBAmount, err := CalcSwapTokenOrderAmountV3(order1, order2)
//				if err != nil {
//					continue
//				}
//
//				//构建撮合交易
//				buildSwapTx, err := BuildSwapTokenOrderByWitness(order1.OrderId, order2.OrderId, tokenAAmount, tokenBAmount)
//				if err != nil {
//					continue
//				}
//				swapTxs = append(swapTxs, buildSwapTx)
//
//				//更新订单副本
//				order1.PendingTokenAAmount = new(big.Int).Add(new(big.Int).SetBytes(order1.PendingTokenAAmount), tokenAAmount).Bytes()
//				order1.PendingTokenBAmount = new(big.Int).Add(new(big.Int).SetBytes(order1.PendingTokenBAmount), tokenBAmount).Bytes()
//				//更新卖单副本
//				order2.PendingTokenAAmount = new(big.Int).Add(new(big.Int).SetBytes(order2.PendingTokenAAmount), tokenAAmount).Bytes()
//				order2.PendingTokenBAmount = new(big.Int).Add(new(big.Int).SetBytes(order2.PendingTokenBAmount), tokenBAmount).Bytes()
//			}
//		}
//	}
//
//	return swapTxs
//}

func (o *orderBookEngine) matchingRuleV3() []*TxTokenSwapOrder {
	swapTxs := []*TxTokenSwapOrder{}
	poolKeys := []string{}
	o.Pools.Range(func(key, value any) bool {
		poolKeys = append(poolKeys, key.(string))
		return true
	})

	//订单池子排序
	sort.Strings(poolKeys)

	//严格按顺序撮合交易
	for _, poolKey := range poolKeys {
		value, ok := o.Pools.Load(poolKey)
		if !ok {
			continue
		}

		buys := arraylist.New()
		sells := arraylist.New()
		list := value.(*arraylist.List)
		iter := list.Iterator()
		for iter.Next() {
			order := iter.Value().(*go_protos.OrderInfo)
			if order.State != OrderStateNormal {
				continue
			}

			if order.Buy {
				buys.Insert(0, order)
			} else {
				sells.Add(order)
			}
		}

		// ============ 开始撮合交易 ============
		buyIter := buys.Iterator()
		for buyIter.Next() {
			buyOrder := buyIter.Value().(*go_protos.OrderInfo)
			sellIter := sells.Iterator()
			for sellIter.Next() {
				sellOrder := sellIter.Value().(*go_protos.OrderInfo)
				//买单大于卖单
				if !(buyOrder.Price >= sellOrder.Price) {
					break
				}

				//判断对手价
				order1 := sellOrder
				order2 := buyOrder
				{
					m := o.NewOrders.Select(func(key interface{}, value interface{}) bool {
						if key.(string) == utils.Bytes2string(buyOrder.OrderId) || key.(string) == utils.Bytes2string(sellOrder.OrderId) {
							return true
						}
						return false
					})
					if m.Size() == 1 {
						if m.Keys()[0].(string) == utils.Bytes2string(sellOrder.OrderId) {
							order1 = buyOrder
							order2 = sellOrder
						}
					} else if m.Size() == 2 {
						if m.Keys()[1].(string) == utils.Bytes2string(sellOrder.OrderId) {
							order1 = buyOrder
							order2 = sellOrder
						}
					}
				}

				//计算成交量
				tokenAAmount, tokenBAmount, err := CalcSwapTokenOrderAmountV3(order1, order2)
				if err != nil {
					continue
				}

				//构建撮合交易
				swapTx, err := BuildSwapTokenOrderByWitness(order1.OrderId, order2.OrderId, tokenAAmount, tokenBAmount)
				if err != nil {
					continue
				}
				swapTxs = append(swapTxs, swapTx)

				//更新订单副本
				order1.PendingTokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(order1.PendingTokenAAmount), tokenAAmount).Bytes()
				//更新卖单副本
				order2.PendingTokenAAmount = new(big.Int).Sub(new(big.Int).SetBytes(order2.PendingTokenAAmount), tokenAAmount).Bytes()
			}
		}
	}

	o.NewOrders.Clear()

	return swapTxs
}

//func (o *orderBookEngine) matchingRuleV3_Old() []*TxTokenSwapOrder {
//	swapTxs := []*TxTokenSwapOrder{}
//	poolKeys := []string{}
//	o.Pools.Range(func(key, value any) bool {
//		poolKeys = append(poolKeys, key.(string))
//		return true
//	})
//
//	//订单池子排序
//	sort.Strings(poolKeys)
//
//	//严格按顺序撮合交易
//	for _, poolKey := range poolKeys {
//		value, ok := o.Pools.Load(poolKey)
//		if !ok {
//			continue
//		}
//
//		list := value.(*arraylist.List)
//		buys := arraylist.New()
//		sells := arraylist.New()
//		iter := list.Iterator()
//		for iter.Next() {
//			oi := iter.Value().(*go_protos.OrderInfo)
//			if oi.State == OrderStateNormal {
//				if oi.Buy {
//					//订单的副本
//					buys.Add(&orderReplica{
//						OrderInfo:       oi,
//						TmpTokenAAmount: new(big.Int).SetBytes(oi.TokenAAmount),
//						TmpTokenBAmount: new(big.Int).SetBytes(oi.TokenBAmount),
//					})
//				} else {
//					//订单的副本
//					sells.Add(&orderReplica{
//						OrderInfo:       oi,
//						TmpTokenAAmount: new(big.Int).SetBytes(oi.TokenAAmount),
//						TmpTokenBAmount: new(big.Int).SetBytes(oi.TokenBAmount),
//					})
//				}
//			}
//		}
//
//		// ============ 开始撮合交易 ============
//		buyIter := buys.Iterator()
//		for buyIter.Next() {
//			buy := buyIter.Value().(*orderReplica)
//			sellIter := sells.Iterator()
//			for sellIter.Next() {
//				sell := sellIter.Value().(*orderReplica)
//				//过滤价格
//				if buy.OrderInfo.Price < sell.OrderInfo.Price {
//					continue
//				}
//
//				//匹配
//				tokenAAmount, tokenBAmount, tokenBRefund, err := CalcSwapTokenOrderAmount(
//					buy.TmpTokenAAmount,
//					buy.TmpTokenBAmount,
//					sell.TmpTokenAAmount,
//					sell.TmpTokenBAmount,
//				)
//				if err != nil {
//					continue
//				}
//
//				//构建撮合交易
//				buildSwapTx, err := BuildSwapTokenOrderByWitness(buy.OrderInfo.OrderId, sell.OrderInfo.OrderId, tokenAAmount, tokenBAmount)
//				if err != nil {
//					continue
//				}
//				swapTxs = append(swapTxs, buildSwapTx)
//
//				//更新买单副本
//				buy.TmpTokenAAmount = new(big.Int).Sub(buy.TmpTokenAAmount, tokenAAmount)
//				tmpTokenBAmount := new(big.Int).Add(tokenBAmount, tokenBRefund)
//				buy.TmpTokenBAmount = new(big.Int).Sub(buy.TmpTokenBAmount, tmpTokenBAmount)
//				buy.OrderInfo.State = OrderStatePending
//
//				//更新卖单副本
//				sell.TmpTokenAAmount = new(big.Int).Sub(sell.TmpTokenAAmount, tokenAAmount)
//				sell.TmpTokenBAmount = new(big.Int).Sub(sell.TmpTokenBAmount, tokenBAmount)
//				sell.OrderInfo.State = OrderStatePending
//			}
//		}
//	}
//
//	return swapTxs
//}

// 验证撮合订单
func (o *orderBookEngine) CheckSwap(txs []TxItr) error {
	for _, tx := range txs {
		if tx.Class() != config.Wallet_tx_type_token_swap_order {
			continue
		}

		//验证交易与缓存池各项数据是否一致
		swapTx := tx.(*TxTokenSwapOrder)

		key := swapTx.BuildSwapTxKey()
		item, ok := o.txCachePool.Get(key)
		if !ok {
			return errors.New("swap not found in cache pool")
		}
		cacheTx := item.(*swapTxCache).TxTokenSwapOrder

		if !bytes.Equal(swapTx.OrderID1, cacheTx.OrderID1) {
			return errors.New("swap tx buy order id not equal")
		}
		if !bytes.Equal(swapTx.OrderID2, cacheTx.OrderID2) {
			return errors.New("swap tx sell order id not equal")
		}
		if swapTx.TokenAAmount.Cmp(cacheTx.TokenAAmount) != 0 {
			return errors.New("swap tx tokena amount not equal")
		}
		if swapTx.TokenBAmount.Cmp(cacheTx.TokenBAmount) != 0 {
			return errors.New("swap tx tokenb amount not equal")
		}
	}

	return nil
}

// 价格从低到高
func (o *orderBookEngine) byPrice(a, b interface{}) int {
	o1 := a.(*go_protos.OrderInfo)
	o2 := b.(*go_protos.OrderInfo)
	//价格相同,保持原顺序
	if o1.Price >= o2.Price {
		return 1
	} else if o1.Price == o2.Price {
		return 0
	} else {
		return -1
	}
}
