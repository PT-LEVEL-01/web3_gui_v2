package utils

import (
	"fmt"
	"github.com/ledisdb/ledisdb/ledis"
	"os"
	"testing"
	"time"
)

func TestLedisDBLink(t *testing.T) {
	// ledisdbExample()
}

func ledisdbExample() {
	//------------------------------
	fmt.Println("================================")
	dbpath := "D:/test/test/ledisdbdata"
	os.RemoveAll(dbpath)
	preKey := []byte("nihao")

	InitLedis(dbpath)

	m := make(map[string]int64)

	var startNum, min, max int64
	// max := 0
	total := 10000
	//顺序添加1
	start := time.Now()
	for i := 0; i < total; i++ {
		// orderId := GetRandomOneInt64()
		// leveldb.AddToBody(&preKey, uint64(i), nil)
		orderId := GetRandomOneInt64()
		// fmt.Printf("随机数:%d\n", orderId)
		key := Uint64ToBytes(uint64(orderId))

		if _, ok := m[Bytes2string(key)]; !ok {
			m[Bytes2string(key)] = orderId
		} else {
			fmt.Println("有相同数据")
		}

		sp := ledis.ScorePair{
			Score:  orderId,
			Member: key,
		}
		if orderId < min {
			min = orderId
		}
		if orderId > max {
			max = orderId
		}

		Quant.ZAdd(preKey, sp)
		if i == 0 {
			min = orderId
			max = orderId
			startNum = orderId
		}
	}
	fmt.Println("ledisdb顺序添加1耗时:", time.Now().Sub(start))

	fmt.Printf("第一个随机数:%d 最小:%d 最大:%d\n", startNum, min, max)

	start = time.Now()
	sps, _ := Quant.ZRangeByScoreGeneric(preKey, min, max, 1, 10, true)
	fmt.Println("ledisdb顺序查询耗时:", time.Now().Sub(start))
	fmt.Printf("%d\n", len(sps))
	fmt.Printf("%+v\n", sps)

	//删除member
	// n, _ := Quant.ZRem(preKey, Uint64ToBytes(uint64(max)))
	// fmt.Printf("ZRem:%d\n", n)

	//key/value是否存在
	n, _ := Quant.Exists(Uint64ToBytes(uint64(min)))
	fmt.Printf("Exists:%d\n", n)
	n, _ = Quant.Exists(preKey)
	fmt.Printf("Exists:%d\n", n)

	//表是否存在
	n, _ = Quant.ZKeyExists(preKey)
	fmt.Printf("ZKeyExists:%d\n", n)

	//获取下标
	n, _ = Quant.ZRank(preKey, Uint64ToBytes(uint64(max)))
	fmt.Printf("ZRank:%d\n", n)

	//通过member查询orderid
	n, _ = Quant.ZScore(preKey, Uint64ToBytes(uint64(startNum)))
	fmt.Printf("ZScore:%d\n", n)

	n, _ = Quant.ZCard(preKey)
	fmt.Printf("ZCard:%d\n", n)

	n, _ = Quant.ZCard([]byte("h"))
	fmt.Printf("ZCard:%d\n", n)

	sps, _ = Quant.ZRangeByScoreGeneric(preKey, startNum-1, max, 0, 10, false)
	fmt.Printf("%+v\n", sps)

	//
	n, _ = Quant.ZScore(preKey, Uint64ToBytes(uint64(min)))
	fmt.Printf("ZScore:%d\n", n)

	//删除zset
	n, _ = Quant.ZRemRangeByScore(preKey, min, startNum)
	fmt.Printf("ZRemRangeByScore:%d\n", n)

	key := Uint64ToBytes(uint64(0))
	sp := ledis.ScorePair{
		Score:  0,
		Member: key,
	}
	n, _ = Quant.ZAdd(preKey, sp)
	fmt.Printf("ZAdd:%d\n", n)

	n, _ = Quant.ZAdd(preKey, sp)
	fmt.Printf("ZAdd:%d\n", n)

	//删除member
	n, _ = Quant.ZRem(preKey, Uint64ToBytes(uint64(startNum)))
	fmt.Printf("ZRem:%d\n", n)

	//获取下标
	n, _ = Quant.ZRank(preKey, Uint64ToBytes(uint64(max)))
	fmt.Printf("ZRank:%d\n", n)

	key = []byte("haha")
	//存储一个数据，设置超时时间
	// ss, err := Quant.Set(key, key)
	Quant.SetEX(key, 10, key)
	value, err := Quant.Get(key)
	if err != nil {
		fmt.Println("get error:", err.Error())
	}
	fmt.Println("查询结果:", string(value))
	time.Sleep(time.Second * 12)
	value, err = Quant.Get(key)
	if err != nil {
		fmt.Println("get error:", err.Error())
	}
	fmt.Println("查询结果2:", string(value))

	fmt.Println("================================")
}
