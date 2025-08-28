package config

import (
	"fmt"
	"testing"
)

func TestBftMajorityPrinciple(t *testing.T) {
	for i := 0; i < 33; i++ {
		successTotal := BftMajorityPrinciple(i)
		fmt.Println("节点总数量:", i, "允许成功数量:", successTotal)
	}
}
