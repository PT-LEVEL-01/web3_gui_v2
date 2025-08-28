package utils

import (
	"fmt"
	"testing"
)

func TestLevelDBMapInList(t *testing.T) {
	// startTestLeveldbMapInList()
}

func startTestLeveldbMapInList() {
	cleanLeveldb()
	createleveldb()

	leveldbExample_SaveMapInList() //
	ldb.PrintAll()
}

/*
保存查询豇豆藤集合测试
*/
func leveldbExample_SaveMapInList() {
	dbkey := Uint64ToBytesByBigEndian(3)
	key := Uint64ToBytesByBigEndian(1)
	_, err := ldb.SaveMapInList(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	key = Uint64ToBytesByBigEndian(2)
	_, err = ldb.SaveMapInList(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	items, err := ldb.FindMapInListAllKeyToList(dbkey)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的豇豆藤集合结果:", one.Key, one.Value)
	}

	key = Uint64ToBytesByBigEndian(2)
	_, err = ldb.SaveMapInList(dbkey, key, key)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	key = Uint64ToBytesByBigEndian(3)
	_, err = ldb.SaveMapInList(dbkey, key, key)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	key = Uint64ToBytesByBigEndian(3)
	index, err := ldb.SaveMapInList(dbkey, key, key)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	items, err = ldb.FindMapInListAllKeyToList(dbkey)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的豇豆藤集合结果:", one.Key, one.Value)
	}

	listItem, err := ldb.FindMapInListKeyList(dbkey, key)
	if err != nil {
		panic(err)
	}
	for _, one := range listItem {
		fmt.Println("查询key=3的豇豆藤集合结果:", one.Index, one.Value)
	}

	//开始修改内容
	err = ldb.UpdateMapInListByIndex(dbkey, key, index, []byte("hello"))
	if err != nil {
		panic(err)
	}
	listItem, err = ldb.FindMapInListKeyList(dbkey, key)
	if err != nil {
		panic(err)
	}
	for _, one := range listItem {
		fmt.Println("查询修改后的key=3的豇豆藤集合结果:", one.Index, one.Value)
	}

	//查询指定一条记录
	value, err := ldb.FindMapInListByIndex(dbkey, key, index)
	if err != nil {
		panic(err)
	}
	fmt.Println("查询指定一条记录:", string(value))

	//删除
	dbkey = Uint64ToBytesByBigEndian(2)
	key = Uint64ToBytesByBigEndian(1)
	index, err = ldb.SaveMapInList(dbkey, key, []byte("nihao"))
	if err != nil {
		panic(err)
	}
	index, err = ldb.SaveMapInList(dbkey, key, []byte("nihaoa"))
	if err != nil {
		panic(err)
	}
	err = ldb.RemoveMapInListByIndex(dbkey, key, index)
	if err != nil {
		panic(err)
	}
	listItem, err = ldb.FindMapInListKeyList(dbkey, key)
	if err != nil {
		panic(err)
	}
	for _, one := range listItem {
		fmt.Println("查询删除后的dbkey=2的豇豆藤集合结果:", one.Index, one.Value)
	}
	err = ldb.RemoveMapInListByKey(dbkey, key)
	if err != nil {
		panic(err)
	}
	listItem, err = ldb.FindMapInListKeyList(dbkey, key)
	if err != nil {
		panic(err)
	}
	for _, one := range listItem {
		fmt.Println("查询删除后的dbkey=2的豇豆藤集合结果:", one.Index, one.Value)
	}

	count, err := ldb.FindMapInListCount(Uint64ToBytesByBigEndian(3), Uint64ToBytesByBigEndian(3))
	if err != nil {
		panic(err)
	}
	fmt.Println("查询列表记录总个数:", count)

}
