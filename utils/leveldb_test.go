package utils

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	dbpath = "D:/test/db/leveldbtest"
)

var ldbLock = new(sync.Mutex)
var ldb *LevelDB

func TestLevelDB(t *testing.T) {
	// startTestLeveldb()
}

func startTestLeveldb() {
	leveldbExample_buildLeveldbKey()
	// return
	cleanLeveldb()
	createleveldb()
	// leveldbExample_ShowSort() //插入并打印key，看排序规律
	// return
	// buildRandKeyLeveldb() //构建数据
	// ldb.PrintAll()
	// leveldbExample_Snapshot()
	// leveldbExample_Iterator()
	// leveldbExample_Transaction() //测试事务处理
	// leveldbExample_QueryEfficiency()
	leveldbExample_SaveList() //测试保存List
	leveldbExample_SaveMap()  //测试保存Map
	// leveldbExample_SaveCowpeaVine() //
	ldb.PrintAll()
}

/*
测试构建数据库key
*/
func leveldbExample_buildLeveldbKey() {
	oldKey := []byte("hello")
	newKey, err := NewLeveldbKey(oldKey)
	if err != nil {
		fmt.Println("创建key错误:", err.Error())
		panic(err)
	}
	oldKey[0] = 0
	fmt.Println(oldKey, newKey)
}

/*
保存Map集合测试
*/
func leveldbExample_SaveMap() {
	dbkey := Uint64ToBytesByBigEndian(4)
	key := Uint64ToBytesByBigEndian(1)
	err := ldb.SaveMap(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(6)
	key = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMap(dbkey, key, Uint64ToBytesByBigEndian(2))
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}
	dbkey = Uint64ToBytesByBigEndian(6)
	key = Uint64ToBytesByBigEndian(2)
	err = ldb.SaveMap(dbkey, key, Uint64ToBytesByBigEndian(3))
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(8)
	key = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMap(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(10)
	key = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMap(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	dbkey = Uint64ToBytesByBigEndian(257)
	key = Uint64ToBytesByBigEndian(1)
	err = ldb.SaveMap(dbkey, key, nil)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	items, err := ldb.FindMapAllToList(Uint64ToBytesByBigEndian(6))
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的Map结果:", one.Key, one.Value)
	}

	kvp := make([]KVPair, 0)
	dbkey = Uint64ToBytesByBigEndian(6)
	kvpOne, err := NewKVPair(true, Uint64ToBytesByBigEndian(5), nil)
	kvp = append(kvp, *kvpOne)
	kvpOne, err = NewKVPair(false, Uint64ToBytesByBigEndian(5), nil)
	kvp = append(kvp, *kvpOne)
	err = ldb.SaveMapMore(dbkey, kvp...)
	if err != nil {
		fmt.Println("保存Map集合错误:", err.Error())
		panic(err)
	}

	//循环插入n条记录
	start := time.Now()
	dbkey = Uint64ToBytesByBigEndian(60)
	num := 1000000
	for i := 0; i < num; i++ {
		key = Uint64ToBytesByBigEndian(uint64(i))
		ldb.SaveMap(dbkey, key, key)
	}
	fmt.Println("循环保存map集合花费时间:", time.Now().Sub(start))

	//删除一条数据
	err = ldb.RemoveMapByKey(dbkey, Uint64ToBytesByBigEndian(1))
	if err != nil {
		fmt.Println("删除Map集合错误:", err.Error())
		panic(err)
	}
	//删除n条记录
	start = time.Now()
	err = ldb.RemoveMapByDbKey(dbkey, uint64(num+1), 0)
	if err != nil {
		fmt.Println("删除Map集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("删除map集合花费时间:", time.Now().Sub(start))
}

/*
保存List集合测试
*/
func leveldbExample_SaveList() {
	dbkey := Uint64ToBytesByBigEndian(5)
	index, err := ldb.SaveList(dbkey, dbkey)
	if err != nil {
		fmt.Println("保存list集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("保存", index)

	dbkey = Uint64ToBytesByBigEndian(7)
	index, err = ldb.SaveList(dbkey, Uint64ToBytesByBigEndian(2))
	if err != nil {
		fmt.Println("保存list集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("保存", index)
	dbkey = Uint64ToBytesByBigEndian(7)
	index, err = ldb.SaveList(dbkey, Uint64ToBytesByBigEndian(3))
	if err != nil {
		fmt.Println("保存list集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("保存", index)

	dbkey = Uint64ToBytesByBigEndian(9)
	index, err = ldb.SaveList(dbkey, dbkey)
	if err != nil {
		fmt.Println("保存list集合错误:", err.Error())
		panic(err)
	}
	fmt.Println("保存", index)

	dbkey = Uint64ToBytesByBigEndian(7)
	items, err := ldb.FindList(dbkey)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的list结果:", one.Index, one.Value)
	}
}

/*
普通单个查询，和迭代器批量查询效率对比
*/
func leveldbExample_QueryEfficiency() {
	//乱序查找
	keys := make([][]byte, 0)
	for i := 0; i < 10000; i++ {
		randNum := GetRandNum(1e8)
		keyOne := Uint64ToBytesByBigEndian(uint64(randNum))
		keys = append(keys, keyOne)
	}
	//第一种方法，一次一次的查询
	start := time.Now()
	total := 0
	// resultKeys := make([][]byte, 0)
	for _, one := range keys {
		_, err := ldb.Find(one)
		if err != nil {
			if err.Error() == leveldb.ErrNotFound.Error() {
				continue
			}
			panic(err)
		}
		total++
	}
	fmt.Println("乱序查询花费时间:", time.Now().Sub(start), "查询到的次数:", total)

	//顺序查找
	keys = make([][]byte, 0)
	for i := 1e7; i < 1e7+10000; i++ {
		keyOne := Uint64ToBytesByBigEndian(uint64(i))
		keys = append(keys, keyOne)
	}
	//第一种方法，一次一次的查询
	start = time.Now()
	total = 0
	for _, one := range keys {
		_, err := ldb.Find(one)
		if err != nil {
			if err.Error() == leveldb.ErrNotFound.Error() {
				continue
			}
			panic(err)
		}
		total++
	}
	fmt.Println("顺序查询花费时间:", time.Now().Sub(start), "查询到的次数:", total)

	//顺序迭代器查找
	// keys = make([][]byte, 0)
	// for i := 1e7; i < 1e7+10000; i++ {
	// 	keyOne := Uint64ToBytesByBigEndian(uint64(i))
	// 	keys = append(keys, keyOne)
	// }
	//第一种方法，一次一次的查询
	start = time.Now()
	total = 0
	iter := ldb.GetDB().NewIterator(nil, nil)
	iter.Seek(keys[0])
	for i := 0; i < 10000; iter.Next() {
		i++
		total++
	}
	iter.Release()
	fmt.Println("顺序迭代器查询花费时间:", time.Now().Sub(start), "查询到的次数:", total)
}

/*
迭代器，测试保存相同的key，能否用迭代器查询出来。结果是查不出来
*/
func leveldbExample_Iterator() {
	//测试保存相同的key，能否用Iterator查询出来
	key := []byte("nihao")
	ldb.Save(key, &key)
	ldb.Save(key, &key)
	ldb.Save(key, &key)
	ldb.Save(key, &key)

	iter := ldb.GetDB().NewIterator(nil, nil)
	count := 0
	for iter.Next() {
		count++
		fmt.Println(iter.Key())
	}
	fmt.Println("查询总数量:", count)
}

/*
事务
事务未提交之前，不能做其他保存操作
*/
func leveldbExample_Transaction() {
	dbkey := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	value := []byte{}
	tr, err := ldb.GetDB().OpenTransaction()
	if err != nil {
		return
	}
	//查询这个dbkey的最大index是多少
	indexDBKEY := append(dbkey, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	indexBs, err := tr.Get(indexDBKEY, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			indexBs = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			tr.Discard()
			return
		}
	}
	index := BytesToUint64ByBigEndian(indexBs)
	index++
	indexBs = Uint64ToBytesByBigEndian(index)
	//保存最新的index
	err = tr.Put(indexDBKEY, indexBs, nil)
	if err != nil {
		tr.Discard()
		return
	}
	//保存记录
	err = tr.Put(append(dbkey, indexBs...), value, nil)
	if err != nil {
		tr.Discard()
		return
	}
	// tr.Commit()

	c := make(chan bool, 1)
	go func() {
		start := time.Now()
		ldb.Save([]byte("hello"), nil)
		c <- false
		fmt.Println("保存成功", time.Now().Sub(start))
	}()
	time.Sleep(time.Second * 3)
	tr.Discard()
	<-c

	ldb.PrintAll()
	return
}

/*
快照，在快照中做操作，只会改变快照中的结果，适合做临时操作。
*/
func leveldbExample_Snapshot() {
	start := time.Now()
	snapshot, err := ldb.GetDB().GetSnapshot()
	if err != nil {
		fmt.Println("error:", err.Error())
		panic(err)
	}
	fmt.Println("建立快照花时间:", time.Now().Sub(start))
	fmt.Print(snapshot.String())

	key := []byte("key")
	ldb.Save(key, &key)

	value, err := snapshot.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		fmt.Println("error:", err.Error())
		panic(err)
	}
	fmt.Println("查询结果:", string(value))

	// createleveldb()

	// ldb.GetDB().GetSnapshot()

	// snapshot.Release()
}

/*
构建随机数据
*/
func buildRandKeyLeveldb() {
	for i := 0; i < 1e8; i++ {
		if i%1e7 == 0 {
			fmt.Println("插入到:", i)
		}
		key := Uint64ToBytesByBigEndian(uint64(i))
		ldb.Save(key, &key)
	}
}

func createleveldb() {
	ldbLock.Lock()
	defer ldbLock.Unlock()
	if ldb != nil {
		ldb.Close()
		ldb = nil
	}
	var err error
	ldb, err = CreateLevelDB(dbpath)
	if err != nil {
		fmt.Println("error:", err.Error())
		panic(err)
	}
}

func cleanLeveldb() {
	os.RemoveAll(dbpath)
}

/*
插入并打印key，看排序规律
*/
func leveldbExample_ShowSort() {
	n := 1
	key := Uint64ToBytesByBigEndian(uint64(n))
	ldb.Save(key, nil)

	n = 255
	key = Uint64ToBytesByBigEndian(uint64(n))
	ldb.Save(key, nil)

	n = 256
	key = Uint64ToBytesByBigEndian(uint64(n))
	ldb.Save(key, nil)

	iter := ldb.GetDB().NewIterator(nil, nil)
	for iter.Next() {
		fmt.Println("key", iter.Key(), "value", iter.Value())
	}
	iter.Release()
	iter.Error()
}
