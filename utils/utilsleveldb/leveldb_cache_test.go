package utilsleveldb

import (
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
	"web3_gui/utils"
)

func TestLeveldbCache(t *testing.T) {
	cleanLeveldb()
	createleveldb()
	//leveldbExample_cache_saveMore()
	//ldb.PrintAll()
	//leveldbExample_cache_set_flash()
	//leveldbExample_set_findAll()
	//leveldbExample_set_findRange()
	//leveldbExample_cache_saveMore()
	//leveldbExample_cache_find()
	leveldbExample_SetExample()

	closeLeveldb()
	cleanLeveldb()
}

func leveldbExample_set_findAll() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	items, ERR := ldb.Cache_Set_FindRange(dbkey, nil, 0, true)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	for _, one := range items {
		utils.Log.Info().Hex("查询结果 key", one.Key.Byte()).Hex("查询结果 value", one.Value).Send()
	}
}
func leveldbExample_set_findRange() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	startKey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	items, ERR := ldb.Cache_Set_FindRange(dbkey, startKey, 0, false)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	for _, one := range items {
		utils.Log.Info().Hex("查询结果 key", one.Key.Byte()).Hex("查询结果 value", one.Value).Send()
	}
}

func leveldbExample_cache_set_flash() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	key5, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	value5 := []byte("hello")
	ldb.Cache_Set_Save(dbkey, key5, &value5)
	key6, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	value6 := []byte("hello")
	ldb.Cache_Set_Save(dbkey, key6, &value6)
	//ldb.Cache_Commit()
	//添加一些干扰数据
	dbkey3, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	key7, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	value7 := []byte("hello")
	ldb.Cache_Set_Save(dbkey3, key7, &value7)

	dbkey9, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(9))
	ldb.Cache_Set_Save(dbkey9, key7, &value7)
}

func TestSortSlice(t *testing.T) {
	list := make([][]byte, 0)
	list = append(list, []byte{0, 0, 2})
	list = append(list, []byte{0, 0, 1})
	list = append(list, []byte{0, 0, 0, 3})
	list = append(list, []byte{0, 0, 0, 4})
	slices := SliceSort{false, list}
	sort.Sort(slices)
	for i, one := range slices.list {
		fmt.Println("排序结果", i, hex.EncodeToString(one))
	}
}

func leveldbExample_SetExample() {
	dbKey1, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1)) //config.DBKEY_improxy_client_orders
	dbKey2, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2)) //
	//utils.Log.Info().Msgf("保存未支付订单key:%s", hex.EncodeToString(dbKey2.Byte()))

	number := utils.Uint64ToBytesByBigEndian(999)
	numberDBKey, _ := BuildLeveldbKey(number)
	cache := NewCache()
	//保存订单
	utils.Log.Info().Hex("保存订单 key", dbKey1.JoinKey(numberDBKey).Byte()).Send()
	cache.Set_Save(dbKey1, numberDBKey, nil)
	//客户端正在服务的订单列表
	utils.Log.Info().Hex("保存订单number key", dbKey2.JoinKey(numberDBKey).Byte()).Send()
	cache.Set_Save(dbKey2, numberDBKey, &number)
	err := ldb.Cache_CommitCache(cache)
	if err != nil {
		panic(err)
	}
	//utils.Log.Info().Msgf("保存未支付订单:%s", hex.EncodeToString(form.Number))

	//dbKey1 := JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_orders)
	fromDbKey := dbKey2
	toDbKey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	cache = NewCache()
	//刷新订单
	cache.Set_Save(dbKey1, numberDBKey, nil)
	//从未支付订单中删除
	cache.Set_Remove(fromDbKey, numberDBKey)
	//保存到已支付订单列表
	cache.Set_Save(toDbKey, numberDBKey, &number)
	err = ldb.Cache_CommitCache(cache)
	if err != nil {
		panic(err)
	}
	dbKey2 = toDbKey
	//
	items, ERR := ldb.Cache_Set_FindRange(dbKey2, nil, 0, false)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Int("未支付订单 返回数量", len(items)).Send()
	keys := make([]*LeveldbKey, 0, len(items))
	for _, one := range items {
		utils.Log.Info().Hex("查询出来的订单number", one.Value).Send()
		keyOne, ERR := BuildLeveldbKey(one.Value)
		if ERR.CheckFail() {
			panic(ERR.String())
		}
		utils.Log.Info().Hex("查询订单 key", keyOne.Byte()).Send()
		keys = append(keys, keyOne)
	}
	items, ERR = ldb.Cache_Set_FindMore(dbKey1, keys, true)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Int("未支付订单 返回数量", len(items)).Send()

	//items, err := config.LevelDB.FindMapByKeys(*dbKey2, keys...)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}

}
