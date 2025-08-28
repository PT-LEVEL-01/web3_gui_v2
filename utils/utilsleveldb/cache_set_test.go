package utilsleveldb

import (
	"testing"
	"web3_gui/utils"
)

func TestCacheMapSave(t *testing.T) {
	cleanLeveldb()
	createleveldb()
	leveldbExample_cache_map_findMore()
	ldb.PrintAll()
	closeLeveldb()
	cleanLeveldb()
}

func leveldbExample_cache_map_findMore() {
	//
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	key1, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	value1 := []byte{1}
	key2, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	value2 := []byte{2}
	key3, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	value3 := []byte{3}
	ldb.Cache_Set_Save(dbkey, key1, &value1)
	ldb.Cache_Set_Save(dbkey, key2, &value2)
	ldb.Cache_Set_Save(dbkey, key3, &value3)
	err := ldb.Cache_Commit()
	if err != nil {
		panic(err)
	}

	keys := make([]*LeveldbKey, 0)
	keys = append(keys, key1)
	keys = append(keys, key2)
	keys = append(keys, key3)
	items, ERR := ldb.Cache_Set_FindMore(dbkey, keys, true)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Int("查询结果数量", len(items)).Send()

	ldb.Cache_Set_Remove(dbkey, key1)
	key4, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(4))
	value4 := []byte{4}
	ldb.Cache_Set_Save(dbkey, key4, &value4)
	err = ldb.Cache_Commit()
	if err != nil {
		panic(err)
	}
}
