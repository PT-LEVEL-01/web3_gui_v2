package utilsleveldb

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestLevelDBMapInMap(t *testing.T) {
	//startTestLeveldbMapInMap()
}

func startTestLeveldbMapInMap() {
	//key := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	//newKey, _ := LeveldbBuildKey(key)
	//fmt.Println(key, newKey)

	//cleanLeveldb()
	//createleveldb()
	//mapInMapLoopSave() //
	//closeLeveldb()
	//cleanLeveldb()

	createleveldb()
	mapInMapOpertion() //
	closeLeveldb()
	cleanLeveldb()

	// createleveldb()
	// mapInMapRemoveOpertion() //
	// closeLeveldb()
	// cleanLeveldb()

	// createleveldb()
	// mapInMapRemoveOpertionMoreLoop() //
	// closeLeveldb()
	// cleanLeveldb()

}

func mapInMapLoopSave() {
	start := time.Now()
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(300))
	n := 10000
	for i := 0; i < n; i++ {
		keyOut, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}
	fmt.Println("循环保存MapInMap", n, "次耗时:", time.Now().Sub(start))
}

func mapInMapOpertion() {

	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(300))
	keyOut, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(300))
	keyOut, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	keyIn, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	ERR = ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存Map集合错误:", ERR.String())
		panic(ERR.String())
	}

	//dbkey, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(300))
	//keyOut, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(2))
	//keyIn, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	//ERR = ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
	//if !ERR.CheckSuccess() {
	//	fmt.Println("保存Map集合错误:", ERR.String())
	//	panic(ERR.String())
	//}
	//
	//dbkey, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(300))
	//keyOut, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(3))
	//keyIn, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	//ERR = ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
	//if !ERR.CheckSuccess() {
	//	fmt.Println("保存Map集合错误:", ERR.String())
	//	panic(ERR.String())
	//}
	//
	//value, err := ldb.FindMapInMapByKeyIn(*dbkey, *keyOut, *keyIn)
	//if err != nil {
	//	fmt.Println("查询Map集合错误:", err.Error())
	//	panic(err)
	//}
	//fmt.Println("查询MapInMapKeyIn集合:", value)

	items, ERR := ldb.FindMapInMapByKeyOut(*dbkey, *keyOut)
	if !ERR.CheckSuccess() {
		fmt.Println("查询Map集合错误:", ERR.String())
		panic(ERR.String())
	}
	for _, one := range items {
		fmt.Println("查询MapInMapKeyOut集合:", one.Key, one.Value)
	}
}

func mapInMapRemoveOpertion() {

	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(301))
	keyOut, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(1))
	for i := 0; i < 10000; i++ {
		keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}

	err := ldb.RemoveMapInMapByKeyOutInterval(*dbkey, *keyOut, 1, time.Microsecond)
	if err != nil {
		fmt.Println("查询Map集合错误:", err.Error())
		panic(err)
	}
	// fmt.Println("查询MapInMap集合:", value)
}

//----------------------------------
/*
测试一定量数据，指定删除
*/
func mapInMapRemoveOpertionMoreLoop() {
	for i := 0; i < 100; i++ {
		mapInMapRemoveOpertionMore()
	}
}
func mapInMapRemoveOpertionMore() {

	fmt.Println("开始插入一定量数据")
	dbkey, _ := BuildLeveldbKey([]byte{1, 1})
	keyOut, _ := BuildLeveldbKey([]byte{0, 0, 0, 0, 0, 0, 0, 9})
	for i := 0; i < 800000; i++ {
		keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}
	fmt.Println("开始插入一定量数据")
	keyOut, _ = BuildLeveldbKey([]byte{0, 0, 0, 0, 0, 0, 0, 10})
	for i := 0; i < 800; i++ {
		keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}

	c := make(chan bool, 1)
	go func() {
		log.Println("开始删除指定数据")
		//
		keyOut, _ := BuildLeveldbKey([]byte{0, 0, 0, 0, 0, 0, 0, 9})
		err := ldb.RemoveMapInMapByKeyOutInterval(*dbkey, *keyOut, 10000000, time.Second)
		if err != nil {
			fmt.Println("删除Map集合错误:", err.Error())
			panic(err)
		}
		log.Println("删除完成")

		value, err := ldb.FindMapInMapByKeyIn(*dbkey, *keyOut, *keyOut)
		if err != nil {
			if !errors.Is(err, leveldb.ErrNotFound) {
				fmt.Println("查询Map集合错误:", err.Error())
				panic(err)
			}
		}
		if value != nil && len(*value) > 0 {
			panic("找到了")
		}
		c <- false
	}()

	// time.Sleep(time.Second * 3)

	fmt.Println("开始插入一定量数据")
	keyOut, _ = BuildLeveldbKey([]byte{0, 0, 0, 0, 0, 0, 0, 11})
	for i := 0; i < 800000; i++ {
		keyIn, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := ldb.SaveMapInMap(*dbkey, *keyOut, *keyIn, keyIn.key, nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存Map集合错误:", ERR.String())
			panic(ERR.String())
		}
	}

	// time.Sleep(time.Minute * 3)
	<-c

}
