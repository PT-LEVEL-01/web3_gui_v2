package mining

import (
	"encoding/json"

	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
	冻结交易余额
*/
// func (this *BalanceManager) Frozen(items []*TxItem, tx TxItr) {
// 	// if config.Wallet_txitem_save_db {
// 	// 	this.notspentBalanceDB.Frozen(items, tx)
// 	// } else {
// 	// 	this.notspentBalance.Frozen(items, tx)
// 	// }
// }

func (this *BalanceManager) Frozen(items []*TxItem, tx TxItr) {
	// if config.Wallet_txitem_save_db {
	// 	this.notspentBalanceDB.Frozen(items, tx)
	// } else {
	// 	this.notspentBalance.Frozen(items, tx)
	// }
}

/*
解冻回滚冻结的交易
*/
func (this *BalanceManager) Unfrozen(blockHeight uint64, blockTime int64) {
	//start := config.TimeNow()
	if unFrozenHeight(blockHeight) {
		this.UnlockByHeight(blockHeight, blockTime)
	}
	//engine.Log.Info("解冻高度:%s ,height:%d", config.TimeNow().Sub(start), blockHeight)
	return
	// if config.Wallet_txitem_save_db {
	// 	//db存储方式不需要设置解冻
	// } else {
	// 	this.notspentBalance.Unfrozen(blockHeight, blockTime)
	// }
	// engine.Log.Info("解锁高度和时间:%d %d", blockHeight, blockTime)

	// txItems, err := GetFrozenHeight(blockHeight, blockTime)
	txItems, err := GetFrozenHeightForAddrValue(blockHeight)
	if err != nil {
		engine.Log.Error("GetFrozenHeight error:%s", err.Error())
		return
	}
	for _, one := range *txItems {
		// engine.Log.Info("解锁余额:%+v", one)
		// engine.Log.Info("解锁余额:%s %d", one.Addr.B58String(), one.Value)
		//增加可用余额
		_, oldValue := db.GetNotSpendBalance(one.Addr)
		// engine.Log.Info("计算解锁余额:%d %d %d", oldValue, one.Value, oldValue+one.Value)
		oldValue += one.Value
		db.SetNotSpendBalance(one.Addr, oldValue)
		//减少锁定余额
		oldValue = GetAddrFrozenValue(one.Addr)
		// engine.Log.Info("计算解锁余额2:%s %d %d %d", one.Addr.B58String(), oldValue, one.Value, oldValue-one.Value)
		oldValue -= one.Value
		SetAddrFrozenValue(one.Addr, oldValue)
	}
	// RemoveFrozenHeight(blockHeight, blockTime)
	RemoveFrozenHeightForAddrValue(blockHeight, blockTime)

	this.UnlockByHeight(blockHeight, blockTime)
}

//region 优化解冻高度
/**
hash批量解冻高度
*/
func unFrozenHeight(blockHeight uint64) bool {
	txItems, err := GetFrozenHeightForAddrValue(blockHeight)
	if err != nil {
		engine.Log.Error("GetFrozenHeight error:%s", err.Error())
		return false
	}

	addrs := make([]crypto.AddressCoin, 0)
	frozenKeys := make([][]byte, 0)
	for _, one := range *txItems {
		addrs = append(addrs, *one.Addr)
		frozenKeys = append(frozenKeys, config.BuildAddrFrozen(*one.Addr))
	}

	//查询余额
	balances := db.GetNotSpendBalances(addrs)

	//增加可用余额
	setBalances := make(map[string]uint64)
	for k := range balances {
		addrStr := utils.Bytes2string(addrs[k])

		balances[k].Value += (*txItems)[k].Value
		setBalances[addrStr] = balances[k].Value
	}

	db.SetNotSpendBalances(setBalances)

	//查询锁定余额
	oldValues := getAddrFrozenValues(frozenKeys)

	if oldValues != nil && len(oldValues) > 0 {
		//减少锁定余额
		kvPairs := make([]db2.KVPair, 0, len(frozenKeys))
		for k, v := range frozenKeys {
			oldValues[k] -= (*txItems)[k].Value

			kvPairs = append(kvPairs, db2.KVPair{
				Key:   v,
				Value: utils.Uint64ToBytes(oldValues[k]),
			})
		}
		LedisMultiSaves(kvPairs)
	}
	//删除锁定高度的记录
	RemoveFrozenHeightForAddrValue(blockHeight, 0)
	return true
}

/*
*
json解冻高度
*/
func unFrozenHeight1(blockHeight uint64) bool {
	//获取高度下的所有信息
	frozenHeightValue := getAddrFrozenHeight(blockHeight)

	if frozenHeightValue == nil || len(frozenHeightValue) == 0 {
		return false
	}
	addrs := make([]crypto.AddressCoin, 0)
	frozenKeys := make([][]byte, 0)

	for k := range frozenHeightValue {
		//addr := crypto.AddressCoin(k)
		addr := crypto.AddressFromB58String(k)
		addrs = append(addrs, addr)
		frozenKeys = append(frozenKeys, config.BuildAddrFrozen(addr))
	}

	//查询余额
	balances := db.GetNotSpendBalances(addrs)

	//增加可用余额
	setBalances := make(map[string]uint64)
	for k := range balances {
		addrStr := utils.Bytes2string(addrs[k])

		balances[k].Value += frozenHeightValue[addrs[k].B58String()]
		setBalances[addrStr] = balances[k].Value
	}

	db.SetNotSpendBalances(setBalances)

	//查询锁定余额
	oldValues := getAddrFrozenValues(frozenKeys)

	if oldValues != nil && len(oldValues) > 0 {
		//减少锁定余额
		kvPairs := make([]db2.KVPair, 0, len(frozenKeys))
		for k, v := range frozenKeys {
			addrStr := addrs[k].B58String()
			//addrStr := utils.Bytes2string(addrs[k])

			oldValues[k] -= frozenHeightValue[addrStr]

			kvPairs = append(kvPairs, db2.KVPair{
				Key:   v,
				Value: utils.Uint64ToBytes(oldValues[k]),
			})
		}
		LedisMultiSaves(kvPairs)
	}
	//删除锁定高度的记录
	removeAddrFrozenHeight(blockHeight)
	return true
}

func getAddrFrozenHeight(height uint64) map[string]uint64 {
	key := config.BuildFrozenHeight(height)

	value, err := db.LevelTempDB.GetDB().Get(key)
	if err != nil {
		return nil
	}

	var res map[string]uint64
	json.Unmarshal(value, &res)
	return res
}

func removeAddrFrozenHeight(height uint64) error {
	key := config.BuildFrozenHeight(height)
	_, err := db.LevelTempDB.GetDB().Del(key)
	return err
}

//endregion

/*
	删除冻结的交易
	@txid         []byte    TxItem中交易txid
	@voutIndex    uint64    TxItem中的vout
*/
// func (this *BalanceManager) DelFrozen(txid []byte, voutIndex uint64) {
// 	this.notspentBalance.DelFrozen(txid, voutIndex)
// }
