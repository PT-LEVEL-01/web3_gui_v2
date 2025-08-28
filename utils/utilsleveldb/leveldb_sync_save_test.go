package utilsleveldb

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"

	"web3_gui/utils"
)

/*
 */
func TestSyncSave(t *testing.T) {
	//startSyncSave()
}

func startSyncSave() {
	// cleanLeveldb()
	// createleveldb()
	// leveldbExample_SyncSave()
	// closeLeveldb()
	// cleanLeveldb()

	createleveldb()
	leveldbExample_SyncTransaction()
	leveldbExample_contrastTransactionAndBatch()
	closeLeveldb()
	cleanLeveldb()
}

/*
测试异步保存数据，看是否有问题
结论是能正确执行，单次操作使用普通保存速度更快。
多次操作事务，效率低。
*/
func leveldbExample_SyncSave() {
	fmt.Println("start")
	finishChan := make(chan bool, 2)

	a := 1 * 10000
	//直接保存
	go func() {
		fmt.Println("start 直接保存")
		start := time.Now()
		for i := 0; i < a; i++ {
			key := utils.Uint64ToBytesByBigEndian(uint64(i))
			err := ldb.GetDB().Put(key, key, nil)
			if err != nil {
				panic(err.Error())
			}
		}
		fmt.Println("end 直接保存", time.Now().Sub(start))
		finishChan <- false
	}()
	b := 1 * 1000
	//使用事务保存
	go func() {
		fmt.Println("start 事务保存")
		start := time.Now()
		for i := a; i < a+b; i++ {
			key := utils.Uint64ToBytesByBigEndian(uint64(i))
			tr, err := ldb.GetDB().OpenTransaction()
			if err != nil {
				panic(err.Error())
			}
			err = tr.Put(key, key, nil)
			if err != nil {
				panic(err.Error())
			}
			err = tr.Commit()
			if err != nil {
				panic(err.Error())
			}
		}
		fmt.Println("end 事务保存", time.Now().Sub(start))
		finishChan <- false
	}()

	<-finishChan
	<-finishChan

	//查询所有，看保存数量是否正确
	count := 0
	iter := ldb.GetDB().NewIterator(nil, nil)
	for iter.Next() {
		count++
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("finish")

}

/*
异步事务
*/
func leveldbExample_SyncTransaction() {
	fmt.Println("start")
	finishChan := make(chan bool, 3)

	//开启第1个事务
	go func() {
		start := time.Now()
		key := utils.Uint64ToBytesByBigEndian(uint64(1))
		tr, err := ldb.GetDB().OpenTransaction()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("start 111 事务保存")
		time.Sleep(time.Second * 2)
		err = tr.Put(key, key, nil)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("end 111 事务保存", time.Now().Sub(start))
		err = tr.Commit()
		if err != nil {
			panic(err.Error())
		}

		finishChan <- false
	}()
	//开启第2个事务
	go func() {
		time.Sleep(time.Second)

		start := time.Now()
		key := utils.Uint64ToBytesByBigEndian(uint64(1))
		tr, err := ldb.GetDB().OpenTransaction()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("start 222 事务保存")
		err = tr.Put(key, key, nil)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("end 222 事务保存", time.Now().Sub(start))
		err = tr.Commit()
		if err != nil {
			panic(err.Error())
		}

		finishChan <- false
	}()

	//开启事务后执行查询
	go func() {
		time.Sleep(time.Second / 2)
		start := time.Now()
		key := utils.Uint64ToBytesByBigEndian(uint64(1))
		ldb.GetDB().Has(key, nil)
		sn, err := ldb.GetDB().GetSnapshot()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("start 查询")
		_, err = sn.Has(key, nil)
		if err != nil {
			panic(err.Error())
		}
		sn.Release()
		fmt.Println("end 查询", time.Now().Sub(start))
		finishChan <- false
	}()

	<-finishChan
	<-finishChan
	<-finishChan

	fmt.Println("finish")
}

/*
异步事务
*/
func leveldbExample_contrastTransactionAndBatch() {
	//key := utils.Uint64ToBytesByBigEndian(uint64(1))

	//事务保存100次
	start := time.Now()
	for i := range 100 {
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		//value := utils.Uint64ToBytesByBigEndian(uint64(i))
		tr, err := ldb.GetDB().OpenTransaction()
		if err != nil {
			panic(err.Error())
		}
		//fmt.Println("start 222 事务保存")
		err = tr.Put(key, key, nil)
		if err != nil {
			panic(err.Error())
		}
		err = tr.Commit()
	}
	fmt.Println("end 222 事务保存", time.Now().Sub(start))
	start = time.Now()
	for i := range 100 {
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		batch := new(leveldb.Batch)
		batch.Put(key, key)
		err := ldb.GetDB().Write(batch, nil)
		if err != nil {
			panic(err.Error())
		}
	}
	fmt.Println("end 222 batch保存", time.Now().Sub(start))

}
