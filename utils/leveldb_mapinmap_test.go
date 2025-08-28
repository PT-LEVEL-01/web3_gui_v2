package utils

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestLevelDBMapInMap(t *testing.T) {
	// startTestLeveldbMapInMap()
}

func startTestLeveldbMapInMap() {
	key := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	newKey, _ := LeveldbBuildKey(key)
	fmt.Println(key, newKey)

	// mapInMapOpertion()
	// mapInMapRemoveOpertion()
	// ldb.PrintAll()

	// mapInMapRemoveOpertionMoreLoop()
	mapInMapRemoveOpertionMore()
	// ldb.PrintAll()
}

func mapInMapOpertion() {
	cleanLeveldb()
	createleveldb()

	dbkey := Uint64ToBytesByBigEndian(300)
	keyOut := Uint64ToBytesByBigEndian(1)
	keyIn := Uint64ToBytesByBigEndian(1)
	err := ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(300)
	keyOut = Uint64ToBytesByBigEndian(1)
	keyIn = Uint64ToBytesByBigEndian(2)
	err = ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(300)
	keyOut = Uint64ToBytesByBigEndian(2)
	keyIn = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(300)
	keyOut = Uint64ToBytesByBigEndian(3)
	keyIn = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	value, err := ldb.FindMapInMapByKeyIn(dbkey, keyOut, keyIn)
	if err != nil {
		fmt.Println("查询Map集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("查询MapInMap集合:", value)
}

func mapInMapRemoveOpertion() {
	// cleanLeveldb()
	// createleveldb()

	dbkey := Uint64ToBytesByBigEndian(301)
	keyOut := Uint64ToBytesByBigEndian(1)

	for i := 0; i < 10000; i++ {
		keyIn := Uint64ToBytesByBigEndian(uint64(i))

		err := ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
		if err != nil {
			fmt.Println("保存Map集合错误:", err.Error())
			panic(err)
		}
	}

	err := ldb.RemoveMapInMapByKeyOutInterval(dbkey, keyOut, 1, time.Microsecond)
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
	cleanLeveldb()
	createleveldb()

	fmt.Println("开始插入一定量数据")
	dbkey := []byte{1, 1}
	keyOut := []byte{0, 0, 0, 0, 0, 0, 0, 9}
	for i := 0; i < 800000; i++ {
		keyIn := Uint64ToBytesByBigEndian(uint64(i))
		err := ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
		if err != nil {
			fmt.Println("保存Map集合错误:", err.Error())
			panic(err)
		}
	}
	fmt.Println("开始插入一定量数据")
	keyOut = []byte{0, 0, 0, 0, 0, 0, 0, 10}
	for i := 0; i < 800; i++ {
		keyIn := Uint64ToBytesByBigEndian(uint64(i))
		err := ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
		if err != nil {
			fmt.Println("保存Map集合错误:", err.Error())
			panic(err)
		}
	}

	c := make(chan bool, 1)
	go func() {
		log.Println("开始删除指定数据")
		//
		keyOut := []byte{0, 0, 0, 0, 0, 0, 0, 9}
		err := ldb.RemoveMapInMapByKeyOutInterval(dbkey, keyOut, 10000000, time.Second)
		if err != nil {
			fmt.Println("删除Map集合错误:", err.Error())
			panic(err)
		}
		log.Println("删除完成")

		value, err := ldb.FindMapInMapByKeyIn(dbkey, keyOut, keyOut)
		if err != nil {
			if !errors.Is(err, leveldb.ErrNotFound) {
				fmt.Println("查询Map集合错误:", err.Error())
				panic(err)
			}
		}
		if value != nil && len(value) > 0 {
			panic("找到了")
		}
		c <- false
	}()

	// time.Sleep(time.Second * 3)

	fmt.Println("开始插入一定量数据")
	keyOut = []byte{0, 0, 0, 0, 0, 0, 0, 11}
	for i := 0; i < 800000; i++ {
		keyIn := Uint64ToBytesByBigEndian(uint64(i))
		err := ldb.SaveMapInMap(dbkey, keyOut, keyIn, keyIn)
		if err != nil {
			fmt.Println("保存Map集合错误:", err.Error())
			panic(err)
		}
	}

	// time.Sleep(time.Minute * 3)
	<-c

}
