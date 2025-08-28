package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"web3_gui/chain/config"
	db2 "web3_gui/chain/db/leveldb"
	chainbootConfig "web3_gui/chain_boot/config"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

//type TxItem struct {
//	Addr  *crypto.AddressCoin //收款地址
//	Value uint64              //余额
//
//	VoteType     uint16 //投票类型
//	LockupHeight uint64 //锁仓高度/锁仓时间
//}
//
//var notSpendBalanceLock = new(sync.RWMutex)
//var notSpendBalance = make(map[string]uint64)
//
///*
//获取一个地址的可用余额
//*/
//func GetNotSpendBalance(addr *crypto.AddressCoin) (*TxItem, uint64) {
//	dbkey := config.BuildAddrValue(*addr)
//	item := TxItem{
//		Addr:  addr,
//		Value: 0,
//	}
//	notSpendBalanceLock.Lock()
//	value, _ := notSpendBalance[utils.Bytes2string(dbkey)]
//	item.Value = value
//	notSpendBalanceLock.Unlock()
//	return &item, value
//
//	//dbkey := config.BuildAddrValue(*addr)
//	//valueBs, err := db.LevelTempDB.Find(dbkey)
//	//if err != nil {
//	//	return nil, 0
//	//}
//	//value := utils.BytesToUint64(*valueBs)
//	//item := TxItem{
//	//	Addr:  addr,
//	//	Value: value,
//	//}
//	//return &item, value
//}
//
///*
//*
//批量查询账户余额 缓存
//*/
//func GetNotSpendBalances(addrs []crypto.AddressCoin) []TxItem {
//	if len(addrs) == 0 {
//		return nil
//	}
//
//	notSpendBalanceLock.Lock()
//	defer notSpendBalanceLock.Unlock()
//
//	r := make([]TxItem, len(addrs))
//	for k, v := range addrs {
//		dbkey := config.BuildAddrValue(v)
//		r[k].Addr = &addrs[k]
//		value, _ := notSpendBalance[utils.Bytes2string(dbkey)]
//		r[k].Value = value
//	}
//
//	return r
//}
//
///*
//*
//批量设置余额 缓存
//*/
//func SetNotSpendBalances(balances map[string]uint64) {
//	if len(balances) == 0 {
//		return
//	}
//
//	//大余额的地址集合
//	setBigAddrs := make([]db2.KVPair, 0)
//
//	notSpendBalanceLock.Lock()
//
//	for k, v := range balances {
//		addr := crypto.AddressCoin(k)
//		if v > config.Mining_coin_total {
//			engine.Log.Info("大余额:%d", v)
//			setBigAddrs = append(setBigAddrs, db2.KVPair{
//				Key:   config.BuildAddrValueBig(addr),
//				Value: nil,
//			})
//		}
//
//		notSpendBalance[utils.Bytes2string(config.BuildAddrValue(addr))] = v
//	}
//
//	notSpendBalanceLock.Unlock()
//
//	LedisMultiSaves(setBigAddrs)
//}
//
///*
//*
//批量存储
//*/
//func LedisMultiSaves(vals []db2.KVPair) error {
//	if len(vals) == 0 {
//		return nil
//	}
//	return LevelTempDB.MSet(vals...)
//}
//
///*
//设置地址
//*/
//func SetAddrValueBig(addr *crypto.AddressCoin) error {
//	dbkey := config.BuildAddrValueBig(*addr)
//	return LevelTempDB.Save(dbkey, nil)
//}
//
///*
//设置一个地址的可用余额
//*/
//func SetNotSpendBalance(addr *crypto.AddressCoin, value uint64) error {
//	if value > config.Mining_coin_total {
//		engine.Log.Info("大余额:%d", value)
//		SetAddrValueBig(addr)
//	}
//
//	dbkey := config.BuildAddrValue(*addr)
//	notSpendBalanceLock.Lock()
//	notSpendBalance[utils.Bytes2string(dbkey)] = value
//	notSpendBalanceLock.Unlock()
//	return nil
//}
//
//func GetBalances() map[string]uint64 {
//	notSpendBalanceLock.Lock()
//	defer notSpendBalanceLock.Unlock()
//
//	return notSpendBalance
//}
//
//func SetBalances(balanceSnap map[string]uint64) {
//	notSpendBalanceLock.Lock()
//	defer notSpendBalanceLock.Unlock()
//
//	notSpendBalance = balanceSnap
//}

type TxItem struct {
	Addr         *crypto.AddressCoin //收款地址
	Value        uint64              //余额
	VoteType     uint16              //投票类型
	LockupHeight uint64              //锁仓高度/锁仓时间
}

/*
获取一个地址的可用余额
*/
func GetNotSpendBalance(addr *crypto.AddressCoin) (*TxItem, uint64) {
	dbkey := config.BuildAddrValue(*addr)
	valueBs, err := LevelTempDB.Find(dbkey)
	if err != nil {
		return nil, 0
	}
	value := utils.BytesToUint64(*valueBs)
	item := TxItem{
		Addr:  addr,
		Value: value,
	}
	return &item, value
}

/*
*
批量查询账户余额 缓存
*/
func GetNotSpendBalances(addrs []crypto.AddressCoin) []TxItem {
	if len(addrs) == 0 {
		return nil
	}

	r := make([]TxItem, len(addrs))
	for k, v := range addrs {
		addr := v
		value, _ := GetNotSpendBalance(&addr)
		r[k].Addr = value.Addr
		r[k].Value = value.Value
	}

	return r
}

/*
*
批量查询账户余额 缓存
*/
func FindNotSpendBalances(addrs []coin_address.AddressCoin) ([]uint64, utils.ERROR) {
	if len(addrs) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	values := make([]uint64, 0, len(addrs))
	for _, one := range addrs {
		dbkey := config.BuildAddrValue(one)
		valueBs, err := LevelTempDB.Find(dbkey)
		if err != nil {
			if err == leveldb.ErrNotFound {
				return nil, utils.NewErrorBus(chainbootConfig.ERROR_CODE_CHAIN_address_not_exist, "")
			}
			return nil, utils.NewErrorSysSelf(err)
		}
		value := utils.BytesToUint64(*valueBs)
		values = append(values, value)
	}
	return values, utils.NewErrorSuccess()
}

/*
*
批量设置余额 缓存
*/
func SetNotSpendBalances(balances map[string]uint64) error {
	if len(balances) == 0 {
		return nil
	}

	//大余额的地址集合
	setBigAddrs := make([]db2.KVPair, 0)

	for k, v := range balances {
		addr := crypto.AddressCoin(k)
		if v > config.Mining_coin_total {
			engine.Log.Info("大余额:%d", v)
			setBigAddrs = append(setBigAddrs, db2.KVPair{
				Key:   config.BuildAddrValueBig(addr),
				Value: nil,
			})
		} else {
			setBigAddrs = append(setBigAddrs, db2.KVPair{
				Key:   config.BuildAddrValue(addr),
				Value: utils.Uint64ToBytes(v),
			})
		}
	}

	if err := LedisMultiSaves(setBigAddrs); err != nil {
		return err
	}

	return nil
}

/*
*
批量存储
*/
func LedisMultiSaves(vals []db2.KVPair) error {
	if len(vals) == 0 {
		return nil
	}
	return LevelTempDB.MSet(vals...)
}

/*
设置地址
*/
func SetAddrValueBig(addr *crypto.AddressCoin) error {
	dbkey := config.BuildAddrValueBig(*addr)
	return LevelTempDB.Save(dbkey, nil)
}

/*
设置一个地址的可用余额
*/
func SetNotSpendBalance(addr *crypto.AddressCoin, value uint64) error {
	if value > config.Mining_coin_total {
		engine.Log.Info("大余额:%d", value)
		return SetAddrValueBig(addr)
	}

	dbkey := config.BuildAddrValue(*addr)
	if err := LevelTempDB.Set(dbkey, utils.Uint64ToBytes(value)); err != nil {
		return err
	}
	return nil
}
