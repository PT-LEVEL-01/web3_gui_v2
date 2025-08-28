package mining

import (
	"web3_gui/chain/config"
	"web3_gui/chain/db"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/utils"
)

//region todo 未使用，暂时注释

/*
	添加到冻结高度
*/
//func AddFrozenHeight(addr *crypto.AddressCoin, height uint64, value uint64) error {
//	if height < config.Wallet_frozen_time_min {
//		// return AddZSetFrozenHeightForHeight(addr, height, value)
//		return AddFrozenHeightForAddrValue(addr, height, value)
//	} else {
//		return AddZSetFrozenHeightForTime(addr, int64(height), value)
//	}
//}
//
//func GetFrozenHeight(height uint64, time int64) (*[]*TxItem, error) {
//	txItems, err := GetZSetFrozenHeightForHeight(height)
//	if err != nil {
//		return nil, err
//	}
//	txItemsTime, err := GetZSetFrozenHeightForTime(time)
//	if err != nil {
//		return nil, err
//	}
//
//	*txItems = append(*txItems, *txItemsTime...)
//	return txItems, nil
//}
//
//func RemoveFrozenHeight(height uint64, time int64) error {
//	sps, err := db.LevelTempDB.GetZSetPage(&config.DBKEY_zset_frozen_time, 0, time, int(time))
//	if err != nil {
//		return err
//	}
//	for i := 0; i < len(*sps); i++ {
//		one := (*sps)[i]
//		childKey := append(config.DBKEY_zset_frozen_time_children, utils.Uint64ToBytes(uint64(one.Score))...)
//		err = db.LevelTempDB.DelZSetAll(&childKey)
//		// err := db.LevelDB.AddZSetAutoincrId(&childKey, txItemBs)
//		if err != nil {
//			return err
//		}
//	}
//	err = db.LevelTempDB.DelZSet(&config.DBKEY_zset_frozen_time, 0, time)
//	if err != nil {
//		return err
//	}
//
//	sps, err = db.LevelTempDB.GetZSetPage(&config.DBKEY_zset_frozen_height, 0, int64(height), int(height))
//	if err != nil {
//		return err
//	}
//	for i := 0; i < len(*sps); i++ {
//		one := (*sps)[i]
//		childKey := append(config.DBKEY_zset_frozen_height_children, utils.Uint64ToBytes(uint64(one.Score))...)
//		err = db.LevelTempDB.DelZSetAll(&childKey)
//		// err := db.LevelTempDB.AddZSetAutoincrId(&childKey, txItemBs)
//		if err != nil {
//			return err
//		}
//	}
//	err = db.LevelTempDB.DelZSet(&config.DBKEY_zset_frozen_height, 0, int64(height))
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
///*
//	添加到冻结高度
//*/
//func AddZSetFrozenHeightForHeight(addr *crypto.AddressCoin, height uint64, value uint64) error {
//	txItemBs := UtilsSerializeTxItem(addr, value)
//	childKey := append(config.DBKEY_zset_frozen_height_children, utils.Uint64ToBytes(height)...)
//	// engine.Log.Info("添加冻结key:%s %d", hex.EncodeToString(childKey), height)
//	err := db.LevelTempDB.AddZSetAutoincrId(&childKey, *txItemBs, 0)
//	if err != nil {
//		return err
//	}
//	// engine.Log.Info("添加冻结key:%s", hex.EncodeToString(config.DBKEY_zset_frozen_height))
//	//重复的height并不会添加到数据库
//	err = db.LevelTempDB.AddZSet(&config.DBKEY_zset_frozen_height, &childKey, int64(height))
//	return err
//}
//
//func GetZSetFrozenHeightForHeight(height uint64) (*[]*TxItem, error) {
//	sps, err := db.LevelTempDB.GetZSetPage(&config.DBKEY_zset_frozen_height, 0, int64(height), int(height))
//	if err != nil {
//		return nil, err
//	}
//	// engine.Log.Info("获取冻结的余额:%d", len(*sps))
//	addrMap := make(map[string]uint64)
//	for i := 0; i < len(*sps); i++ {
//		one := (*sps)[i]
//		// engine.Log.Info("查询冻结的余额:%d", one.Score)
//		childKey := append(config.DBKEY_zset_frozen_height_children, utils.Uint64ToBytes(uint64(one.Score))...)
//		spsChild, err := db.LevelTempDB.GetZSetAll(&childKey)
//		if err != nil {
//			return nil, err
//		}
//		// engine.Log.Info("11111:%d", len(*spsChild))
//		for j := 0; j < len(*spsChild); j++ {
//			childOne := (*spsChild)[j]
//			addr, value := UtilsParseTxItem(&childOne.Member)
//			oldvalue, _ := addrMap[utils.Bytes2string(*addr)]
//			oldvalue += value
//			addrMap[utils.Bytes2string(*addr)] = oldvalue
//		}
//	}
//	txItems := make([]*TxItem, 0)
//	for addrStr, value := range addrMap {
//		addr := crypto.AddressCoin([]byte(addrStr))
//		// engine.Log.Info("111111:%s %d", addr.B58String(), value)
//		item := TxItem{
//			Addr:  &addr,
//			Value: value,
//		}
//		txItems = append(txItems, &item)
//	}
//	// engine.Log.Info("返回数量:%d %d", len(txItems), len(addrMap))
//	return &txItems, nil
//}
//
///*
//	添加到冻结高度
//*/
//func AddZSetFrozenHeightForTime(addr *crypto.AddressCoin, height int64, value uint64) error {
//	txItemBs := UtilsSerializeTxItem(addr, value)
//	childKey := append(config.DBKEY_zset_frozen_time_children, utils.Uint64ToBytes(uint64(height))...)
//	err := db.LevelTempDB.AddZSetAutoincrId(&childKey, *txItemBs, 0)
//	if err != nil {
//		return err
//	}
//	//重复的height并不会添加到数据库
//	err = db.LevelTempDB.AddZSet(&config.DBKEY_zset_frozen_time, &childKey, int64(height))
//	return err
//}
//
//func GetZSetFrozenHeightForTime(height int64) (*[]*TxItem, error) {
//	sps, err := db.LevelTempDB.GetZSetPage(&config.DBKEY_zset_frozen_time, 0, int64(height), int(height))
//	if err != nil {
//		return nil, err
//	}
//	addrMap := make(map[string]uint64)
//	for i := 0; i < len(*sps); i++ {
//		one := (*sps)[i]
//		childKey := append(config.DBKEY_zset_frozen_time_children, utils.Uint64ToBytes(uint64(one.Score))...)
//		spsChild, err := db.LevelTempDB.GetZSetAll(&childKey)
//		if err != nil {
//			return nil, err
//		}
//		for j := 0; j < len(*spsChild); j++ {
//			childOne := (*spsChild)[j]
//			addr, value := UtilsParseTxItem(&childOne.Member)
//			oldvalue, _ := addrMap[utils.Bytes2string(*addr)]
//			oldvalue += value
//			addrMap[utils.Bytes2string(*addr)] = oldvalue
//		}
//	}
//	txItems := make([]*TxItem, 0)
//	for addrStr, value := range addrMap {
//		addr := crypto.AddressCoin([]byte(addrStr))
//		item := TxItem{
//			Addr:  &addr,
//			Value: value,
//		}
//		txItems = append(txItems, &item)
//	}
//	return &txItems, nil
//}

//endregion

/*
将地址和余额序列化
*/
func UtilsSerializeTxItem(addr *crypto.AddressCoin, value uint64) *[]byte {
	bs := make([]byte, 0, len(*addr)+8)
	bs = append(bs, utils.Uint64ToBytes(value)...)
	bs = append(bs, *addr...)
	return &bs
}

func UtilsParseTxItem(bs *[]byte) (*crypto.AddressCoin, uint64) {
	value := utils.BytesToUint64((*bs)[:8])
	addr := crypto.AddressCoin((*bs)[8:])
	return &addr, value
}

//---------------------------
/*
	添加到冻结高度
*/
func AddFrozenHeightForAddrValue(addr *crypto.AddressCoin, height uint64, value uint64) error {
	dbname := config.BuildFrozenHeight(height)
	ledisDB := db.LevelTempDB.GetDB()
	dbValue, err := ledisDB.HGet(dbname, *addr)
	if err != nil {
		return err
	}
	oldValue := uint64(0)
	if dbValue != nil && len(dbValue) > 0 {
		oldValue = utils.BytesToUint64(dbValue)
	}
	value = oldValue + value
	valueBs := utils.Uint64ToBytes(value)
	_, err = ledisDB.HSet(dbname, *addr, valueBs)
	if err != nil {
		return err
	}
	return nil
}

/*
添加到冻结高度
*/
func GetFrozenHeightForAddrValue(height uint64) (*[]*TxItem, error) {
	dbname := config.BuildFrozenHeight(height)
	ledisDB := db.LevelTempDB.GetDB()
	fvps, err := ledisDB.HGetAll(dbname)
	if err != nil {
		return nil, err
	}

	tis := make([]*TxItem, 0, len(fvps))
	for _, one := range fvps {
		addrOne := crypto.AddressCoin(one.Field)
		valueOne := utils.BytesToUint64(one.Value)
		txItemOne := TxItem{
			Addr:  &addrOne, //收款地址
			Value: valueOne, //余额
		}
		tis = append(tis, &txItemOne)
	}
	return &tis, nil
}

/*
删除
*/
func RemoveFrozenHeightForAddrValue(height uint64, time int64) error {
	ledisDB := db.LevelTempDB.GetDB()
	dbname := config.BuildFrozenHeight(height)
	_, err := ledisDB.HClear(dbname)
	if err != nil {
		return err
	}
	return nil
}
