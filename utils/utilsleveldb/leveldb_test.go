package utilsleveldb

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"web3_gui/utils"
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

	cleanLeveldb()
	createleveldb()
	leveldbExample_QueryEfficiency() //
	ldb.PrintAll()
	closeLeveldb()
	cleanLeveldb()

	// leveldbExample_ShowSort() //插入并打印key，看排序规律

	// buildRandKeyLeveldb() //构建数据
	// ldb.PrintAll()
	// leveldbExample_Snapshot()
	// leveldbExample_Iterator()
	// leveldbExample_Transaction() //测试事务处理
	// leveldbExample_SaveCowpeaVine() //

	createleveldb()
	leveldbExample_findAllCount()
	closeLeveldb()
	cleanLeveldb()
}

/*
测试构建数据库key
*/
func leveldbExample_buildLeveldbKey() {
	oldKey := []byte("hello")
	newKey, ERR := BuildLeveldbKey(oldKey)
	if ERR.CheckSuccess() {
		fmt.Println("创建key错误:", ERR.String())
		panic(ERR.String())
	}
	oldKey[0] = 0
	fmt.Println(oldKey, newKey)
}

/*
普通单个查询，和迭代器批量查询效率对比
*/
func leveldbExample_QueryEfficiency() {
	n := 1000000
	//乱序查找
	keys := make([]LeveldbKey, 0)
	for i := 0; i < n; i++ {
		randNum := utils.GetRandNum(1e8)
		keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(randNum)))
		keys = append(keys, *keyOne)
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
	keys = make([]LeveldbKey, 0)
	for i := 1e7; i < 1e7+float64(n); i++ {
		keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		keys = append(keys, *keyOne)
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
	iter.Seek(keys[0].key)
	for i := 0; i < n; {
		if !iter.Next() {
			break
		}
		i++
		total++
	}
	iter.Release()
	fmt.Println("顺序迭代器查询花费时间:", time.Now().Sub(start), "查询到的次数:", total)
}

/*
迭代器，测试保存相同的key，能否用迭代器查询出来。
结果是查不出来
*/
func leveldbExample_Iterator() {
	//测试保存相同的key，能否用Iterator查询出来
	key := []byte("nihao")
	ldb.GetDB().Put(key, key, nil)
	ldb.GetDB().Put(key, key, nil)
	ldb.GetDB().Put(key, key, nil)
	ldb.GetDB().Put(key, key, nil)

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
	index := utils.BytesToUint64ByBigEndian(indexBs)
	index++
	indexBs = utils.Uint64ToBytesByBigEndian(index)
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
		ldb.GetDB().Put([]byte("hello"), nil, nil)
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
快照，在快照中做查询。
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
	ldb.GetDB().Put(key, key, nil)

	value, err := snapshot.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		fmt.Println("error:", err.Error())
		panic(err)
	}
	fmt.Println("查询结果:", string(value))
}

/*
构建随机数据
*/
func buildRandKeyLeveldb() {
	for i := 0; i < 1e8; i++ {
		if i%1e7 == 0 {
			fmt.Println("插入到:", i)
		}
		key := utils.Uint64ToBytesByBigEndian(uint64(i))
		ldb.GetDB().Put(key, key, nil)
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

func closeLeveldb() {
	ldbLock.Lock()
	defer ldbLock.Unlock()
	ldb.Close()
}

func cleanLeveldb() {
	os.RemoveAll(dbpath)
}

/*
插入并打印key，看排序规律
*/
func leveldbExample_ShowSort() {
	n := 1
	key := utils.Uint64ToBytesByBigEndian(uint64(n))
	ldb.GetDB().Put(key, nil, nil)

	n = 255
	key = utils.Uint64ToBytesByBigEndian(uint64(n))
	ldb.GetDB().Put(key, nil, nil)

	n = 256
	key = utils.Uint64ToBytesByBigEndian(uint64(n))
	ldb.GetDB().Put(key, nil, nil)

	iter := ldb.GetDB().NewIterator(nil, nil)
	for iter.Next() {
		fmt.Println("key", iter.Key(), "value", iter.Value())
	}
	iter.Release()
	iter.Error()
}

/*
测试查询总量的速度
*/
func leveldbExample_findAllCount() {
	fmt.Println("开始测试查询总量速度")
	n := 100000
	//
	keys := make([]LeveldbKey, 0)
	for i := 0; i < n; i++ {
		// randNum := utils.GetRandNum(1e8)
		keyOne, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
		keys = append(keys, *keyOne)
	}
	start := time.Now()
	for _, one := range keys {
		ldb.Save(one, &one.key)
	}
	fmt.Println("保存数据耗时:", time.Now().Sub(start))

	//查询总量
	start = time.Now()
	total := 0
	iter := ldb.GetDB().NewIterator(nil, nil)
	iter.Seek(keys[0].key)
	for i := 0; i < n; {
		if !iter.Next() {
			break
		}
		i++
		total++
	}
	iter.Release()
	fmt.Println("查询数据总量耗时:", time.Now().Sub(start), "查询到的次数:", total)

}
