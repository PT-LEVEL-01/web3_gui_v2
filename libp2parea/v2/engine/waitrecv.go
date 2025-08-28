package engine

import (
	"sync"
	"time"
	"web3_gui/utils"
)

var (
	waitRequest = new(sync.Map)
)

/*
注册一个消息，立即返回，另外一个方法去取消息
*/
func RegisterRequestKey(major, minor []byte) {
	keyBs := make([]byte, 0, len(major)+len(minor)+2+2)
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(major)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, major...)
	keyLenBs = utils.Uint16ToBytesByBigEndian(uint16(len(minor)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, minor...)
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	tag := utils.Bytes2string(keyBs)
	//RegisterRequest(tag)
	_, ok := waitRequest.Load(tag)
	if ok {
		// utils.Log.Info().Msgf("111111111")
		return
	}
	c := make(chan interface{}, 1)
	waitRequest.Store(tag, c)
}

/*
删除一个消息
*/
func RemoveRequestKey(major, minor []byte) {
	keyBs := make([]byte, 0, len(major)+len(minor)+2+2)
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(major)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, major...)
	keyLenBs = utils.Uint16ToBytesByBigEndian(uint16(len(minor)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, minor...)
	tag := utils.Bytes2string(keyBs)
	waitRequest.Delete(tag)
}

/*
等待返回消息内容
*/
func WaitResponseByteKey(major, minor []byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	itr, ERR := WaitResponseItrKey(major, minor, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	bs := itr.(*[]byte)
	return bs, utils.NewErrorSuccess()
}

/*
等待返回消息内容
*/
func WaitResponseItrKey(major, minor []byte, timeout time.Duration) (interface{}, utils.ERROR) {
	keyBs := make([]byte, 0, len(major)+len(minor)+2+2)
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(major)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, major...)
	keyLenBs = utils.Uint16ToBytesByBigEndian(uint16(len(minor)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, minor...)
	tag := utils.Bytes2string(keyBs)
	//id := ulid.Make().Bytes()
	//utils.Log.Info().Hex("超时id", id).Str("超时时间", timeout.String()).Send()
	//
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return nil, utils.NewErrorBus(ERROR_code_timeout_recv, "")
	}
	c := cItr.(chan interface{})
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		//utils.Log.Info().Hex("超时id", id).Str("超时时间", timeout.String()).Send()
		waitRequest.Delete(tag)
		close(c)
		//utils.Log.Info().Hex("超时id", id).Str("超时时间", timeout.String()).Send()
		return nil, utils.NewErrorBus(ERROR_code_timeout_recv, "")
	case itr := <-c:
		//utils.Log.Info().Hex("超时id", id).Str("超时时间", timeout.String()).Send()
		//waitRequest.Delete(tag)
		timer.Stop()
		return itr, utils.NewErrorSuccess()
	}
}

/*
有消息内容返回了
*/
func ResponseByteKey(major, minor []byte, bs *[]byte) bool {
	return ResponseItrKey(major, minor, bs)
}

/*
有消息内容返回了
*/
func ResponseItrKey(major, minor []byte, itr interface{}) bool {
	keyBs := make([]byte, 0, len(major)+len(minor)+2+2)
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(major)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, major...)
	keyLenBs = utils.Uint16ToBytesByBigEndian(uint16(len(minor)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, minor...)
	tag := utils.Bytes2string(keyBs)
	//
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return ok
	}
	c := cItr.(chan interface{})
	select {
	case c <- itr:
	default:
	}
	return ok
}
