package flood

import (
	"bytes"
	"errors"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
)

type WaitRecvOne struct {
	waitRequest *sync.Map
}

func NewWaitRecvOne() *WaitRecvOne {
	return &WaitRecvOne{
		waitRequest: new(sync.Map),
	}
}

/*
注册一个消息，立即返回，另外一个方法去取消息
*/
func (this *WaitRecvOne) RegisterRequest(tag string) {
	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	// c := make(chan interface{}, 1)
	// this.lock.Lock()
	// _, ok := this.tags[tag]
	// if ok {
	// 	this.lock.Unlock()
	// 	return
	// }
	// this.tags[tag] = c
	// this.lock.Unlock()
	// return

	c := make(chan interface{}, 1)
	_, ok := this.waitRequest.LoadOrStore(tag, c) // Load(tag)
	if ok {
		// utils.Log.Info().Msgf("111111111")
		return
	}
	// this.waitRequest.Store(tag, c)
	// utils.Log.Info().Msgf("111111111")
}

/*
删除一个消息
*/
func (this *WaitRecvOne) RemoveRequest(tag string) {
	// this.lock.Lock()
	// delete(this.tags, tag)
	// this.lock.Unlock()
	this.waitRequest.Delete(tag)
}

/*
删除一个消息
*/
func (this *WaitRecvOne) Length() int {
	count := 0
	this.waitRequest.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

/*
获取所有等待的消息管道
*/
func (this *WaitRecvOne) GetAllChan() *[]chan interface{} {
	all := make([]chan interface{}, 0)
	this.waitRequest.Range(func(k, v interface{}) bool {
		c, ok := v.(chan interface{})
		if !ok {
			return false
		}
		all = append(all, c)
		return true
	})
	return &all
}

/*
查询一个消息，看是否存在
*/
func (this *WaitRecvOne) FindRequest(tag string) bool {
	// this.lock.RLock()
	// _, ok := this.tags[tag]
	// this.lock.RUnlock()
	// return ok
	_, ok := this.waitRequest.Load(tag)
	return ok
}

/*
有消息内容返回了
@return    bool    是否有对应的等待
*/
func (this *WaitRecvOne) ResponseBytes(tag string, bs *[]byte) bool {
	// this.lock.RLock()
	// c, ok := this.tags[tag]
	// this.lock.RUnlock()
	// if !ok {
	// 	return ok
	// }

	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	cItr, ok := this.waitRequest.Load(tag)
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
func (this *WaitRecvOne) ResponseItr(tag string, itr interface{}) bool {
	// this.lock.RLock()
	// c, ok := this.tags[tag]
	// this.lock.RUnlock()

	// utils.Log.Info().Msgf("tag:%s", hex.EncodeToString([]byte(tag)))
	cItr, ok := this.waitRequest.Load(tag)
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
等待返回消息内容
*/
func (this *WaitRecvOne) WaitResponse(tag string, timeout time.Duration) (*[]byte, error) {
	if timeout < time.Second {
		utils.Log.Warn().Msgf("WaitResponse Timeout time is less than 1 second")
	}
	cItr, ok := this.waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	}
	c := cItr.(chan interface{})
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		this.waitRequest.Delete(tag)
		close(c)
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	case itr := <-c:
		this.waitRequest.Delete(tag)
		timer.Stop()
		err, ok := itr.(error)
		if ok {
			return nil, err
		}
		bs := itr.(*[]byte)
		//len([]byte(config.CLASS_router_err)）= 16
		if bs != nil && len(*bs) == 16 {
			if bytes.Equal(*bs, []byte(config.CLASS_router_err)) {
				return nil, errors.New("Routing message to recv node err")
			}
		}
		// utils.Log.Info().Msgf("111111111", ok)
		return bs, nil
	}
}

/*
等待返回消息内容
*/
func (this *WaitRecvOne) WaitResponseItr(tag string, timeout time.Duration) (interface{}, error) {
	if timeout < time.Second {
		utils.Log.Warn().Msgf("WaitResponseItr Timeout time is less than 1 second")
	}
	cItr, ok := this.waitRequest.Load(tag)
	if !ok {
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	}
	c := cItr.(chan interface{})
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		this.waitRequest.Delete(tag)
		close(c)
		// utils.Log.Info().Msgf("111111111")
		return nil, config.ERROR_wait_msg_timeout
	case itr := <-c:
		this.waitRequest.Delete(tag)
		timer.Stop()
		err, ok := itr.(error)
		if ok {
			return nil, err
		}
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
