package utilsleveldb

import (
	"fmt"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestLevelDBMapInList(t *testing.T) {
	startTestLeveldbMapInList()
}

func startTestLeveldbMapInList() {
	// cleanLeveldb()
	// createleveldb()
	// leveldbExample_SaveMapInList() //
	// ldb.PrintAll()
	// closeLeveldb()

	cleanLeveldb()
	createleveldb()
	//leveldbExample_LoopSaveMapInList() //
	leveldbExample_SaveAdnFindMapInList() //
	closeLeveldb()

}

/*
保存和查询
*/
func leveldbExample_SaveAdnFindMapInList() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	index := []byte{1}
	ERR := ldb.SaveOrUpdateMapInListByIndex(*dbkey, *key, index, index, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	index = []byte{2}
	ERR = ldb.SaveOrUpdateMapInListByIndex(*dbkey, *key, index, index, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	items, ERR := ldb.FindMapInListRangeByKeyOut(*dbkey, nil, 1, true)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("查询的结果", items)

	total, start, end, ERR := ldb.FindMapInListTotal(*dbkey, *key)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("查询的结果", total, start, end, ERR)
}

/*
保存查询豇豆藤集合测试
*/
func leveldbExample_SaveMapInList() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	_, ERR := ldb.SaveMapInList(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	_, ERR = ldb.SaveMapInList(*dbkey, *key, nil, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	items, ERR := ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range items {
		fmt.Println("查询的豇豆藤集合结果:", one.Key, one.Value)
	}

	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	_, ERR = ldb.SaveMapInList(*dbkey, *key, key.key, nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	_, ERR = ldb.SaveMapInList(*dbkey, *key, key.key, nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	index, ERR := ldb.SaveMapInList(*dbkey, *key, key.key, nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	items, ERR = ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range items {
		fmt.Println("查询的豇豆藤集合结果:", one.Key, one.Value)
	}

	listItem, ERR := ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range listItem {
		fmt.Println("查询key=3的豇豆藤集合结果:", one.Index, one.Value)
	}

	//开始修改内容
	ERR = ldb.SaveOrUpdateMapInListByIndex(*dbkey, *key, index, []byte("hello"), nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	listItem, ERR = ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range listItem {
		fmt.Println("查询修改后的key=3的豇豆藤集合结果:", one.Index, one.Value)
	}

	//查询指定一条记录
	item, ERR := ldb.FindMapInListByIndex(*dbkey, *key, index)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("查询指定一条记录:", string(item.Value))

	//删除
	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	index, ERR = ldb.SaveMapInList(*dbkey, *key, []byte("nihao"), nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	index, ERR = ldb.SaveMapInList(*dbkey, *key, []byte("nihaoa"), nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	ERR = ldb.RemoveMapInListByIndex(*dbkey, *key, index, nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	listItem, ERR = ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range listItem {
		fmt.Println("查询删除后的dbkey=2的豇豆藤集合结果:", one.Index, one.Value)
	}
	ERR = ldb.RemoveMapInListByIndex(*dbkey, *key, index, nil)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	listItem, ERR = ldb.FindMapInListAll(*dbkey)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range listItem {
		fmt.Println("查询删除后的dbkey=2的豇豆藤集合结果:", one.Index, one.Value)
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	count, _, _, ERR := ldb.FindMapInListTotal(*dbkey, *key)
	if ERR.CheckFail() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("查询列表记录总个数:", count)
}

/*
循环保存，测试性能
*/
func leveldbExample_LoopSaveMapInList() {
	n := 100
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	start := time.Now()
	for i := 0; i < n; i++ {
		key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		_, ERR := ldb.SaveMapInList(*dbkey, *key, nil, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}
	fmt.Println("循环保存", n, "次MapInList，耗时:", time.Now().Sub(start))
}
