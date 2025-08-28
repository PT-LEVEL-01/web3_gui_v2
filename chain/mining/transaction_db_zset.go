package mining

// import (
// 	"web3_gui/chain/db"
// 	"web3_gui/chain/config"

// 	"web3_gui/keystore/adapter/crypto"
// 	"web3_gui/utils"
// )

// /*
// 	添加到冻结高度
// */
// func AddFrozenHeight(addr *crypto.AddressCoin, height uint64, value uint64) error {
// 	if height < config.Wallet_frozen_time_min {
// 		return AddZSetFrozenHeightForHeight(addr, height, value)
// 	} else {
// 		return AddZSetFrozenHeightForTime(addr, int64(height), value)
// 	}
// }

// func GetFrozenHeight(height uint64, time int64) (*[]*TxItem, error) {
// 	txItems, err := GetZSetFrozenHeightForHeight(height)
// 	if err != nil {
// 		return nil, err
// 	}
// 	txItemsTime, err := GetZSetFrozenHeightForTime(time)
// 	if err != nil {
// 		return nil, err
// 	}
// 	*txItems = append(*txItems, *txItemsTime...)
// 	return txItems, nil
// }

// func RemoveFrozenHeight(height uint64, time int64) error {
// 	sps, err := db.LevelDB.GetZSetPage(&config.DBKEY_zset_frozen_time, 0, time, int(time))
// 	if err != nil {
// 		return err
// 	}
// 	for i := 0; i < len(*sps); i++ {
// 		one := (*sps)[i]
// 		childKey := append(config.DBKEY_zset_frozen_time_children, utils.Uint64ToBytes(uint64(one.Score))...)
// 		err = db.LevelDB.DelZSetAll(&childKey)
// 		// err := db.LevelDB.AddZSetAutoincrId(&childKey, txItemBs)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = db.LevelDB.DelZSet(&config.DBKEY_zset_frozen_time, 0, time)
// 	if err != nil {
// 		return err
// 	}

// 	sps, err = db.LevelDB.GetZSetPage(&config.DBKEY_zset_frozen_height, 0, int64(height), int(height))
// 	if err != nil {
// 		return err
// 	}
// 	for i := 0; i < len(*sps); i++ {
// 		one := (*sps)[i]
// 		childKey := append(config.DBKEY_zset_frozen_height_children, utils.Uint64ToBytes(uint64(one.Score))...)
// 		err = db.LevelDB.DelZSetAll(&childKey)
// 		// err := db.LevelDB.AddZSetAutoincrId(&childKey, txItemBs)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = db.LevelDB.DelZSet(&config.DBKEY_zset_frozen_height, 0, int64(height))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// /*
// 	添加到冻结高度
// */
// func AddZSetFrozenHeightForHeight(addr *crypto.AddressCoin, height uint64, value uint64) error {
// 	txItemBs := UtilsSerializeTxItem(addr, value)
// 	childKey := append(config.DBKEY_zset_frozen_height_children, utils.Uint64ToBytes(height)...)
// 	err := db.LevelDB.AddZSetAutoincrId(&childKey, txItemBs)
// 	if err != nil {
// 		return err
// 	}
// 	//重复的height并不会添加到数据库
// 	err = db.LevelDB.AddZSet(&config.DBKEY_zset_frozen_height, &childKey, int64(height))
// 	return err
// }

// func GetZSetFrozenHeightForHeight(height uint64) (*[]*TxItem, error) {
// 	sps, err := db.LevelDB.GetZSetPage(&config.DBKEY_zset_frozen_height, 0, int64(height), int(height))
// 	if err != nil {
// 		return nil, err
// 	}
// 	addrMap := make(map[string]uint64)
// 	for i := 0; i < len(*sps); i++ {
// 		one := (*sps)[i]
// 		childKey := append(config.DBKEY_zset_frozen_height_children, one.Member...)
// 		spsChild, err := db.LevelDB.GetZSetAll(&childKey)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for j := 0; j < len(*spsChild); j++ {
// 			childOne := (*spsChild)[j]
// 			addr, value := UtilsParseTxItem(&childOne.Member)
// 			oldvalue, _ := addrMap[utils.Bytes2string(*addr)]
// 			oldvalue += value
// 			addrMap[utils.Bytes2string(*addr)] = oldvalue
// 		}
// 	}
// 	txItems := make([]*TxItem, len(addrMap))
// 	for addrStr, value := range addrMap {
// 		addr := crypto.AddressCoin([]byte(addrStr))
// 		item := TxItem{
// 			Addr:  &addr,
// 			Value: value,
// 		}
// 		txItems = append(txItems, &item)
// 	}
// 	return &txItems, nil
// }

// /*
// 	添加到冻结高度
// */
// func AddZSetFrozenHeightForTime(addr *crypto.AddressCoin, height int64, value uint64) error {
// 	txItemBs := UtilsSerializeTxItem(addr, value)
// 	childKey := append(config.DBKEY_zset_frozen_time_children, utils.Uint64ToBytes(uint64(height))...)
// 	err := db.LevelDB.AddZSetAutoincrId(&childKey, txItemBs)
// 	if err != nil {
// 		return err
// 	}
// 	//重复的height并不会添加到数据库
// 	err = db.LevelDB.AddZSet(&config.DBKEY_zset_frozen_time, &childKey, int64(height))
// 	return err
// }

// func GetZSetFrozenHeightForTime(height int64) (*[]*TxItem, error) {
// 	sps, err := db.LevelDB.GetZSetPage(&config.DBKEY_zset_frozen_time, 0, int64(height), int(height))
// 	if err != nil {
// 		return nil, err
// 	}
// 	addrMap := make(map[string]uint64)
// 	for i := 0; i < len(*sps); i++ {
// 		one := (*sps)[i]
// 		childKey := append(config.DBKEY_zset_frozen_time_children, one.Member...)
// 		spsChild, err := db.LevelDB.GetZSetAll(&childKey)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for j := 0; j < len(*spsChild); j++ {
// 			childOne := (*spsChild)[j]
// 			addr, value := UtilsParseTxItem(&childOne.Member)
// 			oldvalue, _ := addrMap[utils.Bytes2string(*addr)]
// 			oldvalue += value
// 			addrMap[utils.Bytes2string(*addr)] = oldvalue
// 		}
// 	}
// 	txItems := make([]*TxItem, len(addrMap))
// 	for addrStr, value := range addrMap {
// 		addr := crypto.AddressCoin([]byte(addrStr))
// 		item := TxItem{
// 			Addr:  &addr,
// 			Value: value,
// 		}
// 		txItems = append(txItems, &item)
// 	}
// 	return &txItems, nil
// }
