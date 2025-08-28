package utilsleveldb

import (
	"fmt"
	"testing"
	"time"

	"web3_gui/utils"
)

/*
 */
func TestMap(t *testing.T) {
	//startMap()
}

func startMap() {
	cleanLeveldb()
	createleveldb()
	//leveldbExample_Save16M()
	leveldbExample_InitMap()
	leveldbExample_FindMapAll()
	//ldb.PrintAll()
	//leveldbExample_FindMap()
	//leveldbExample_SaveMap()
	//leveldbExample_DeleteMap()
	closeLeveldb()
	//cleanLeveldb()
}

/*
初始化Map集合测试
*/
func leveldbExample_InitMap() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(4))
	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	ERR := ldb.SaveMap(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	ERR = ldb.SaveMap(*dbkey, *key, utils.Uint64ToBytesByBigEndian(2), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	ERR = ldb.SaveMap(*dbkey, *key, utils.Uint64ToBytesByBigEndian(3), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(8))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	ERR = ldb.SaveMap(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(10))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	ERR = ldb.SaveMap(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(257))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	ERR = ldb.SaveMap(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
}

/*
保存Map集合测试
*/
func leveldbExample_FindMapAll() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	items, _ := ldb.FindMapAllToList(*dbkey)
	for _, item := range items {
		fmt.Println(item)
	}
}

/*
保存Map集合测试
*/
func leveldbExample_SaveMap() {

	//kvp := make([]KVPair, 0)
	//dbkey, _ := NewLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	//key := utils.Uint64ToBytesByBigEndian(5)
	//kvpOne, ERR := NewKVPair(true, key, nil)
	//if !ERR.CheckSuccess() {
	//	panic(ERR.String())
	//}
	//kvp = append(kvp, *kvpOne)
	//kvpOne, ERR = NewKVPair(true, key, nil)
	//kvp = append(kvp, *kvpOne)
	//err := ldb.SaveMapMore(*dbkey, kvp...)
	//if err != nil {
	//	fmt.Println("保存Map集合错误:", err.Error())
	//	panic(err)
	//}

}

/*
查询Map集合测试
*/
func leveldbExample_FindMap() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	items, err := ldb.FindMapAllToList(*dbkey)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的Map结果:", one.Key, one.Value)
	}

	keys := make([]LeveldbKey, 0)
	keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	keys = append(keys, *keyOne)
	keyOne, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	keys = append(keys, *keyOne)
	keyOne, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	keys = append(keys, *keyOne)
	items, err = ldb.FindMapByKeys(*dbkey, keys...)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("多条查询的Map结果:", one.Key, one.Value)
	}
}

/*
删除Map集合测试
*/
func leveldbExample_DeleteMap() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	//删除一条数据
	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	err := ldb.RemoveMapByKey(*dbkey, *key, nil)
	if err != nil {
		fmt.Println("删除Map集合错误:", err.Error())
		panic(err)
	}

	//循环插入n条记录
	start := time.Now()
	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(60))
	num := 1000000
	for i := 0; i < num; i++ {
		key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ldb.SaveMap(*dbkey, *key, utils.Uint64ToBytesByBigEndian(uint64(i)), nil)
	}
	fmt.Println("循环保存map集合花费时间:", time.Now().Sub(start))

	//删除n条记录
	start = time.Now()
	err = ldb.RemoveMapByDbKey(*dbkey, uint64(num+1), 0)
	if err != nil {
		fmt.Println("删除Map集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("删除map集合花费时间:", time.Now().Sub(start))
}

/*
删除Map集合测试
*/
func leveldbExample_Save16M() {

	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(6))
	for i := uint64(0); i < 100; i++ {
		bs16M := make([]byte, 1024*1024*16)
		copy(bs16M, utils.Uint64ToBytesByBigEndian(i))
		//删除一条数据
		key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(i + 1))

		ERR := ldb.SaveMap(*dbkey, *key, bs16M, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}

	ldb.PrintAll()

	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	_, err := ldb.FindMap(*dbkey, *key)
	if err != nil {
		fmt.Println("查询Map集合错误:", err.Error())
		panic(err)
	}

	//fmt.Println("删除map集合花费时间:", time.Now().Sub(start))
}
