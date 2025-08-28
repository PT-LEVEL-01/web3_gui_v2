package engine

import (
	"testing"
	"time"
	"web3_gui/utils"
)

func TestWaitrecv(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	//先注册
	RegisterRequestKey(key1, key2)
	//记得删除
	defer RemoveRequestKey(key1, key2)
	go asyncResponse(key1, key2)
	//再等待
	itr, ERR := WaitResponseItrKey(key1, key2, time.Second/2)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.Msg).Send()
		return
	}
	utils.Log.Info().Interface("itr", itr).Send()
	//再次等待
	itr, ERR = WaitResponseItrKey(key1, key2, time.Second/2)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.Msg).Send()
		return
	}
	utils.Log.Info().Interface("itr", itr).Send()
}

func asyncResponse(major, minor []byte) {
	time.Sleep(time.Second / 3)
	ResponseItrKey(major, minor, []byte("hello"))
	ResponseItrKey(major, minor, []byte("hello"))
}
