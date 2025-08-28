package utilsleveldb

import (
	"fmt"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"web3_gui/utils"
)

func TestLeveldbBatchOrTransaction(t *testing.T) {
	//startTestLeveldbBatchOrTransaction()
}

/*
测试两种批量保存数据性能对比
*/
func startTestLeveldbBatchOrTransaction() {
	// cleanLeveldb()
	// createleveldb()
	// leveldbExample_BatchMore()
	// closeLeveldb()

	// cleanLeveldb()
	// createleveldb()
	// leveldbExample_TransactionMore()
	// closeLeveldb()

	// cleanLeveldb()
	// createleveldb()
	// leveldbExample_Batch()
	// closeLeveldb()

	cleanLeveldb()
	createleveldb()
	//leveldbExample_BatchDelete()
	leveldbExample_BatchDeleteAndSave()
	closeLeveldb()
}

const count = 1e5

/*
使用Batch批量保存数据
*/
func leveldbExample_BatchMore() {
	batch := new(leveldb.Batch)
	for i := 0; i < count; i++ {
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		batch.Put(key, key)
	}

	start := time.Now()
	err := ldb.GetDB().Write(batch, nil)
	if err != nil {
		fmt.Println("写入Batch错误", err.Error())
		panic(err)
	}
	fmt.Println("Batch批量保存耗时:", time.Now().Sub(start))
}

/*
使用Transaction批量保存数据
*/
func leveldbExample_TransactionMore() {
	keys := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		keys = append(keys, key)
	}

	start := time.Now()
	tr, err := ldb.GetDB().OpenTransaction()
	if err != nil {
		fmt.Println("写入Batch错误", err.Error())
		panic(err)
	}
	for _, one := range keys {
		tr.Put(one, one, nil)
	}
	err = tr.Commit()
	if err != nil {
		fmt.Println("写入Batch错误", err.Error())
		panic(err)
	}
	fmt.Println("Transaction批量保存耗时:", time.Now().Sub(start))
}

/*
使用Batch批量操作数据
*/
func leveldbExample_Batch() {
	batch := new(leveldb.Batch)
	for i := 0; i < 10; i++ {
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		batch.Put(key, key)
	}

	//测试batch序列化后可以暂存
	bs := batch.Dump()
	newBatch := new(leveldb.Batch)
	newBatch.Load(bs)

	// start := time.Now()
	err := ldb.GetDB().Write(newBatch, nil)
	if err != nil {
		fmt.Println("写入Batch错误", err.Error())
		panic(err)
	}
	// fmt.Println("Batch批量保存耗时:", time.Now().Sub(start))
	key := utils.Uint64ToBytesByBigEndian(5)
	value, err := ldb.GetDB().Get(key, nil)
	if err != nil {
		fmt.Println("查询错误", err.Error())
		panic(err)
	}
	fmt.Println("查询结果:", value)
}

/*
测试Batch中的删除操作是否能得到应用
*/
func leveldbExample_BatchDelete() {
	//数据库中原有1-10
	for i := 0; i < 10; i++ {
		key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		// ldb.GetDB().Put(key, key, nil)
		ldb.Save(*key, &key.key)
	}

	ldb.PrintAll()

	//batch中添加删除操作
	batch := new(leveldb.Batch)
	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(1)))
	batch.Delete(key.key)
	ldb.GetDB().Write(batch, nil)

	//查询删除那个key是否存在
	have, err := ldb.Has(*key)
	// have := ldb.CheckHashExist(*key)
	// value, err := ldb.GetDB().Get(key, nil)
	if err != nil {
		fmt.Println("查询错误", err.Error())
		panic(err)
	}
	fmt.Println("查询结果:", have)

	ldb.PrintAll()

}

/*
测试Batch中删除后再保存，是否达到预期
*/
func leveldbExample_BatchDeleteAndSave() {
	//数据库中原有1-10
	for i := 0; i < 10; i++ {
		key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		// ldb.GetDB().Put(key, key, nil)
		ldb.Save(*key, &key.key)
	}

	key, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(20)))
	value := utils.Uint64ToBytesByBigEndian(uint64(21))
	fmt.Println("value:", value)
	ldb.Save(*key, &value)
	ldb.PrintAll()

	//batch中添加删除操作
	batch := new(leveldb.Batch)
	key, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(20)))
	batch.Put(key.key, key.key) //先保存
	batch.Delete(key.key)       //再删除
	ldb.GetDB().Write(batch, nil)

	//查询删除那个key是否存在
	have, err := ldb.Has(*key)
	// have := ldb.CheckHashExist(*key)
	// value, err := ldb.GetDB().Get(key, nil)
	if err != nil {
		fmt.Println("查询错误", err.Error())
		panic(err)
	}
	fmt.Println("查询结果:", have)

	ldb.PrintAll()

}
