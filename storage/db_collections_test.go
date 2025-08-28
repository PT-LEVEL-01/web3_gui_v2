package storage

import (
	"fmt"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestDbCollections(*testing.T) {
	dbCollection_AaveRandFull()
}

func dbCollection_AaveRandFull() {
	fmt.Println("start test dbCollection_AaveRandFull")

	db := CreateDBCollections()
	ERR := db.AddDBone("F:\\storageleveldb", "H:\\storageleveldb", "I:\\storageleveldb")
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	startOut := time.Now()
	startIn := time.Now()
	for i := range 9999999999 {
		if i != 0 && i%1000000 == 0 {
			fmt.Println("插入数据", i, time.Now().Sub(startIn), i/int(time.Now().Sub(startOut).Seconds()))
			startIn = time.Now()
		}
		key := utils.Hash_SHA3_256(utils.Uint64ToBytesByBigEndian(uint64(i)))
		ERR := db.SaveChunk(key, key)
		if !ERR.CheckSuccess() {
			fmt.Println("插入数据", i)
			panic(ERR.String())
		}
	}
	fmt.Println("end")
}
