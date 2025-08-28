package test

import (
	"fmt"
	"testing"
	"web3_gui/chain/config"
)

func TestClacRewardForBlockHeight(t *testing.T) {
	InitConfig()
	totalAll := uint64(0)
	year1 := uint64(0)
	year2 := uint64(0)
	year3 := uint64(0)
	year4 := uint64(0)
	year5 := uint64(0)
	for i := range uint64(60 * 60 * 24 * 365 * 10) {
		n := ClacRewardForBlockHeight(i)
		totalAll += n
		if i <= 3110400*1 {
			year1 += n
		} else if i <= 3110400*2 {
			year2 += n
		} else if i <= 3110400*3 {
			year3 += n
		} else if i <= 3110400*4 {
			year4 += n
		} else if i <= 3110400*5 {
			year5 += n
		}
		if totalAll >= config.Mining_coin_total {
			fmt.Println("最后高度", i)
			break
		}
	}
	fmt.Println("周期1", year1)
	fmt.Println("周期2", year2)
	fmt.Println("周期3", year3)
	fmt.Println("周期4", year4)
	fmt.Println("周期5", year5)
	fmt.Println("总", totalAll)
}
