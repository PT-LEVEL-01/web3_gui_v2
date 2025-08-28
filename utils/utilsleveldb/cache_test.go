package utilsleveldb

import (
	"strconv"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestCache(t *testing.T) {
	cleanLeveldb()
	createleveldb()
	leveldbExample_cache_save_rand()
	//leveldbExample_cache_saveMore()
	//leveldbExample_cache_find()
	ldb.PrintAll()

	closeLeveldb()
	cleanLeveldb()
}

func leveldbExample_cache_save_rand() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	bs1 := []byte("key1")
	key1 := dbkey.JoinByte(&bs1).Byte()
	value1 := []byte("value1")
	ldb.Cache_Save(&key1, &value1)
	//ldb.Cache_Commit()

	bs2 := []byte("key2")
	key2 := dbkey.JoinByte(&bs2).Byte()
	value2 := []byte("value2")
	ldb.Cache_Save(&key2, &value2)

	key3, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	ldb.Cache_Set_Save(dbkey, key3, &bs1)
	key4, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(4))
	ldb.Cache_Set_Save(dbkey, key4, &bs1)

	err := ldb.Cache_Commit()
	if err != nil {
		panic(err)
	}
}

func leveldbExample_cache_saveMore() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	kvps := make([]KVMore, 0)
	saveMap := make(map[string]*[]byte)
	for i := range 10 {
		keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		key := []byte("key" + strconv.Itoa(i))
		value := []byte("value" + strconv.Itoa(i))
		saveMap[utils.Bytes2string(key)] = &value
		kvP := KVMore{
			Key:   *dbkey.JoinKey(keyOne),
			Value: value,
		}
		kvps = append(kvps, kvP)
	}
	start := time.Now()
	//for k, v := range saveMap {
	//	key := []byte(k)
	//	ldb.Cache_Save(&key, v)
	//}
	//ldb.Cache_Commit()
	utils.Log.Info().Msgf("1 spend time:%s", time.Now().Sub(start))

	start = time.Now()
	ldb.Cache_SaveMore(kvps...)
	ldb.Cache_Commit()
	utils.Log.Info().Msgf("2 spend time:%s", time.Now().Sub(start))
}

func leveldbExample_cache_find() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	kvps := make([]KVMore, 0)
	saveMap := make(map[string]*[]byte)
	for i := range 10 {
		keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		key := []byte("key" + strconv.Itoa(i))
		value := []byte("value" + strconv.Itoa(i))
		saveMap[utils.Bytes2string(key)] = &value
		kvP := KVMore{
			Key:   *dbkey.JoinKey(keyOne),
			Value: value,
		}
		kvps = append(kvps, kvP)
	}
	ldb.Cache_SaveMore(kvps...)
	err := ldb.Cache_Commit()
	if err != nil {
		panic(err)
	}
	findkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	keyBs := dbkey.JoinKey(findkey).Byte() //append(dbkey.Byte(), findkey.Byte()...)
	//ldb.Cache_Map_FindMore()
	value, err := ldb.Cache_Find(&keyBs)
	if err != nil {
		panic(err)
	}
	utils.Log.Info().Msgf("查询结果 value:%v", value)
}
