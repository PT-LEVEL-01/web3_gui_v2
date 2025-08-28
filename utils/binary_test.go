package utils

import (
	"testing"
)

func TestFullHighPositionZero(t *testing.T) {
	//先生成一个hash长度不够的字节
	randBs, err := Rand32Byte()
	if err != nil {
		t.Errorf("error:%s", err.Error())
	}
	bs := Hash_SHA3_256(randBs[:])
	Log.Info().Hex("生成hash", bs).Int("生成hash长度", len(bs)).Send()
}
