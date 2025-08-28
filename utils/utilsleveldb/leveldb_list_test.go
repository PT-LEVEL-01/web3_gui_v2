package utilsleveldb

import (
	"fmt"
	"testing"
	"time"
	"web3_gui/utils"
)

/*
 */
func TestList(t *testing.T) {
	startList()
	//startAddListSpendTime()
	//startAddListTransactionSpendTime()
}

func startList() {
	cleanLeveldb()
	createleveldb()
	//leveldbExample_SaveList()
	leveldbExample_FindList()
	//listAddSpendTime()
	//listAddMoreSpendTime()
	//ldb.PrintAll()
	closeLeveldb()
	cleanLeveldb()
}

func startAddListSpendTime() {
	cleanLeveldb()
	createleveldb()
	//leveldbExample_SaveList()
	listAddSpendTime(100000)
	//listAddMoreSpendTime()
	closeLeveldb()

	createleveldb()
	listAddSpendTime(10)
	//ldb.PrintAll()
	closeLeveldb()
	cleanLeveldb()
}

func startAddListTransactionSpendTime() {
	cleanLeveldb()
	createleveldb()
	//leveldbExample_SaveList()
	listAddTransactionSpendTime()
	//listAddMoreSpendTime()
	closeLeveldb()

	createleveldb()
	listAddTransactionSpendTime()
	//ldb.PrintAll()
	closeLeveldb()
	cleanLeveldb()
}

/*
查询List集合测试
*/
func leveldbExample_FindList() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))

	//查询记录总条数
	total, startIndex, endIndex, ERR := ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	index, ERR := ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(5), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(2), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)
	index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(3), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)

	//再查询记录总条数
	total, startIndex, endIndex, ERR = ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)

	items, ERR := ldb.FindListRange(*dbkey, nil, 0, false)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", len(items))

}

/*
保存List集合测试
*/
func leveldbExample_SaveList() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))

	//查询记录总条数
	total, startIndex, endIndex, ERR := ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(5))
	index, ERR := ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(5), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(2), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)
	index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(3), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(9))
	index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(9), nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("保存", index)

	dbkey, _ = BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	items, err := ldb.FindListAll(*dbkey)
	if err != nil {
		panic(err)
	}
	for _, one := range items {
		fmt.Println("查询的list结果:", one.Index, one.Value)
	}

	total, startIndex, endIndex, ERR = ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)

	//删除
	err = ldb.RemoveListInterval(*dbkey, 0, 0)
	if err != nil {
		panic(err)
	}
	//
	total, startIndex, endIndex, ERR = ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("删除后，查询记录总量", total, startIndex, endIndex)

	removeIndex := make([][]byte, 0)
	//插入大量数据
	for i := 0; i < 200; i++ {
		index, ERR = ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(uint64(i)), nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存list集合错误:", ERR.String())
			panic(ERR.String())
		}
		//fmt.Println("保存", index)
		if i == 6 || i == 12 {
			removeIndex = append(removeIndex, index)
		}
	}

	//删除中间的数据
	ERR = ldb.RemoveListMore(false, *dbkey, removeIndex...)
	if err != nil {
		panic(ERR.String())
	}

	//分批次查询所有
	rangeOne := uint64(10) //一次查询10条
	total, startIndex, endIndex, ERR = ldb.FindListTotal(*dbkey)
	if err != nil {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)
	var startIndexBs []byte
	for i := uint64(0); i < (total/rangeOne)+1; i++ {
		items, ERR := ldb.FindListRange(*dbkey, startIndexBs, rangeOne, true)
		if !ERR.CheckSuccess() {
			panic(ERR.String())
		}
		for _, one := range items {
			fmt.Println("查询的list结果:", one.Index, one.Value)
			startIndexBs = one.Index
		}
		fmt.Println("----")
	}

	//删除所有数据
	err = ldb.RemoveListInterval(*dbkey, 0, 0)
	if err != nil {
		panic(err)
	}

	//再查询记录总条数
	total, startIndex, endIndex, ERR = ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex)

}

func listAddSpendTime(n int) {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	start := time.Now()
	//插入大量数据
	for i := 0; i < n; i++ {
		_, ERR := ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(uint64(i)), nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存list集合错误:", ERR.String())
			panic(ERR.String())
		}
		//fmt.Println("保存", index)
	}
	fmt.Println("插入耗时", time.Now().Sub(start))

	start = time.Now()
	//给查询总量计时
	total, startIndex, endIndex, ERR := ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex, time.Now().Sub(start))
}

func listAddTransactionSpendTime() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	start := time.Now()
	err := ldb.OpenTransaction()
	if err != nil {
		panic(err)
	}
	//插入大量数据
	for i := 0; i < 200; i++ {
		_, ERR := ldb.SaveList(*dbkey, utils.Uint64ToBytesByBigEndian(uint64(i)), nil)
		if !ERR.CheckSuccess() {
			fmt.Println("保存list集合错误:", ERR.String())
			panic(ERR.String())
		}
		//fmt.Println("保存", index)
	}
	err = ldb.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Println("插入耗时", time.Now().Sub(start))

	start = time.Now()
	//给查询总量计时
	total, startIndex, endIndex, ERR := ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex, time.Now().Sub(start))
}

func listAddMoreSpendTime() {
	dbkey, _ := BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(7))
	//插入大量数据
	values := make([][]byte, 0)
	for i := 0; i < 200; i++ {
		values = append(values, utils.Uint64ToBytesByBigEndian(uint64(i)))
		//fmt.Println("保存", index)
	}
	start := time.Now()
	_, ERR := ldb.SaveListMore(*dbkey, values, nil)
	if !ERR.CheckSuccess() {
		fmt.Println("保存list集合错误:", ERR.String())
		panic(ERR.String())
	}
	fmt.Println("插入耗时", time.Now().Sub(start))

	start = time.Now()
	//给查询总量计时
	total, startIndex, endIndex, ERR := ldb.FindListTotal(*dbkey)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Println("查询记录总量", total, startIndex, endIndex, time.Now().Sub(start))
}
