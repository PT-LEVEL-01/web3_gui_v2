package flood

import (
	"bytes"
	"errors"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
)

const waitRequstTime = 30 //超时时间设置为60秒

var (
	waitRequest = new(sync.Map)
)

type HttpRequestWait struct {
	tagMap *sync.Map
}

// /*
// 	等待请求返回
// */
// func WaitRequest(class, tag string, timeout int64) (*[]byte, error) {
// 	if timeout <= 0 {
// 		timeout = waitRequstTime
// 	}

// 	// fmt.Println("1111111111111", class, tag)
// 	// utils.Log.Info().Msgf("WaitRequest %s %s", class, tag)
// 	rwItr, ok := waitRequest.Load(class) //[class]
// 	if !ok {
// 		c := make(chan *[]byte, 1)
// 		hrw := HttpRequestWait{
// 			tagMap: new(sync.Map), //make(map[string]chan *[]byte),
// 		}
// 		hrw.tagMap.Store(tag, c)       //[tag] = c
// 		waitRequest.Store(class, &hrw) //[class] = &hrw
// 		ticker := time.NewTicker(time.Second * time.Duration(timeout))
// 		// timer.NewTimer().

// 		select {
// 		case <-ticker.C:
// 			hrw.tagMap.Delete(tag)
// 			return nil, config.ERROR_wait_msg_timeout
// 		case bs := <-c:
// 			ticker.Stop()
// 			return bs, nil
// 		}

// 	}
// 	rw := rwItr.(*HttpRequestWait)
// 	cItr, ok := rw.tagMap.Load(tag) // [tag]
// 	if !ok {
// 		c := make(chan *[]byte, 1)
// 		rw.tagMap.Store(tag, c) // [tag] = c

// 		ticker := time.NewTicker(time.Second * time.Duration(timeout))
// 		select {
// 		case <-ticker.C:
// 			rw.tagMap.Delete(tag)
// 			return nil, config.ERROR_wait_msg_timeout
// 		case bs := <-c:
// 			ticker.Stop()
// 			return bs, nil
// 		}
// 	}
// 	c := cItr.(chan *[]byte)

// 	ticker := time.NewTicker(time.Second * time.Duration(timeout))
// 	select {
// 	case <-ticker.C:
// 		rw.tagMap.Delete(tag)
// 		return nil, config.ERROR_wait_msg_timeout
// 	case bs := <-c:
// 		ticker.Stop()
// 		return bs, nil
// 	}
// }

// /*
// 	返回等待
// */
// func ResponseWait(class, tag string, bs *[]byte) {
// 	// fmt.Println("ResponseWait", class, tag)
// 	// utils.Log.Info().Msgf("ResponseWait %s %s", class, tag)
// 	rwItr, ok := waitRequest.Load(class) // [class]
// 	if !ok {
// 		return
// 	}
// 	rw := rwItr.(*HttpRequestWait)
// 	cItr, ok := rw.tagMap.Load(tag) // [tag]
// 	if !ok {
// 		return
// 	}
// 	c := cItr.(chan *[]byte)

// 	select {
// 	case c <- bs:
// 		return
// 	default:
// 	}
// }

/*
注册一个消息，立即返回，另外一个方法去取消息
*/
func RegisterRequest(tag string) {
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	_, ok := waitRequest.Load(tag)
	if ok {
		// utils.Log.Info().Msgf("111111111")
		return
	}
	c := make(chan interface{}, 1)
	waitRequest.Store(tag, c)
	// utils.Log.Info().Msgf("111111111")
}

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
	RegisterRequest(tag)
}

/*
删除一个消息
*/
func RemoveRequest(tag string) {
	waitRequest.Delete(tag)
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
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	tag := utils.Bytes2string(keyBs)
	RemoveRequest(tag)
}

/*
查询一个消息，看是否存在
*/
func FindRequest(tag string) bool {
	_, ok := waitRequest.Load(tag)
	return ok
}

/*
有消息内容返回了
@return    bool    是否有对应的等待
*/
func ResponseBytes(tag string, bs *[]byte) bool {
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return ok
	}
	c := cItr.(chan interface{})
	select {
	case c <- bs:
		// utils.Log.Info().Msgf("111111111")
	default:
		// utils.Log.Info().Msgf("111111111")
	}
	return ok
}

/*
有消息内容返回了
*/
func ResponseItr(tag string, itr interface{}) bool {
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return ok
	}
	c := cItr.(chan interface{})
	select {
	case c <- itr:
		// utils.Log.Info().Msgf("111111111")
	default:
		// utils.Log.Info().Msgf("111111111")
	}
	return ok
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
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	tag := utils.Bytes2string(keyBs)
	return ResponseItr(tag, itr)
}

/*
等待返回消息内容
*/
func WaitResponse(tag string, timeout time.Duration) (*[]byte, error) {
	if timeout < time.Second {
		utils.Log.Warn().Msgf("WaitResponse Timeout time is less than 1 second")
	}
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	}
	c := cItr.(chan interface{})
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		waitRequest.Delete(tag)
		close(c)
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	case itr := <-c:
		waitRequest.Delete(tag)
		timer.Stop()
		bs := itr.(*[]byte)
		//len([]byte(config.CLASS_router_err)）= 16
		if bs != nil && len(*bs) == 16 {
			if bytes.Equal(*bs, []byte(config.CLASS_router_err)) {
				return nil, errors.New("Routing message to recv node err")
			}
		}
		return bs, nil
	}
}

/*
等待返回消息内容
*/
func WaitResponseItr(tag string, timeout time.Duration) (interface{}, error) {
	if timeout < time.Second {
		utils.Log.Warn().Msgf("WaitResponseItr Timeout time is less than 1 second")
	}
	cItr, ok := waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	}
	c := cItr.(chan interface{})
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		waitRequest.Delete(tag)
		close(c)
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	case itr := <-c:
		waitRequest.Delete(tag)
		timer.Stop()
		if itb, ok := itr.(*[]byte); ok {
			//len([]byte(config.CLASS_router_err)）= 16
			if itb != nil && len(*itb) == 16 {
				if bytes.Equal(*itb, []byte(config.CLASS_router_err)) {
					return nil, errors.New("Routing message to recv node err")
				}
			}
		}
		// utils.Log.Info().Msgf("111111111")
		return itr, nil
	}
}

/*
等待返回消息内容
*/
func WaitResponseItrKey(major, minor []byte, timeout time.Duration) (interface{}, error) {
	keyBs := make([]byte, 0, len(major)+len(minor)+2+2)
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(major)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, major...)
	keyLenBs = utils.Uint16ToBytesByBigEndian(uint16(len(minor)))
	keyBs = append(keyBs, keyLenBs...)
	keyBs = append(keyBs, minor...)
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	tag := utils.Bytes2string(keyBs)
	return WaitResponseItr(tag, timeout)
}
