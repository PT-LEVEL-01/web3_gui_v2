package flood

import (
	"sync"
	"time"

	"web3_gui/libp2parea/v1/config"
)

var GroupWaitRecv = NewWaitRecvGroup()

type WaitRecvGroup struct {
	// lock  *sync.RWMutex //
	group *sync.Map //
}

func NewWaitRecvGroup() *WaitRecvGroup {
	return &WaitRecvGroup{
		// lock:  new(sync.RWMutex),
		group: new(sync.Map),
	}
}

/*
查询一个组
*/
func (this *WaitRecvGroup) GetGroup(groupID string) *WaitRecvOne {
	waitRequestItr, ok := this.group.Load(groupID)
	if !ok {
		return nil
	}
	one := waitRequestItr.(*WaitRecvOne)
	return one
}

/*
注册一个消息，立即返回，另外一个方法去取消息
*/
func (this *WaitRecvGroup) RegisterRequest(groupID, tag string) {
	one := NewWaitRecvOne()
	// this.lock.Lock()
	oldMapItr, ok := this.group.LoadOrStore(groupID, one)
	// this.lock.Unlock()
	if ok {
		one = oldMapItr.(*WaitRecvOne)
	}
	one.RegisterRequest(tag)
	// c := make(chan interface{}, 1)
	// one.LoadOrStore(tag, c)
}

/*
删除一个消息
*/
func (this *WaitRecvGroup) RemoveRequest(groupID, tag string) {
	// this.lock.Lock()
	one := this.GetGroup(groupID)
	if one == nil {
		return
	}
	one.RemoveRequest(tag)
	// if one.Length() == 0 {
	// 	this.group.Delete(groupID)
	// }
	// this.lock.Unlock()
}

/*
删除一个组的消息
*/
func (this *WaitRecvGroup) RemoveGroup(groupID, tag string) {
	// this.lock.Lock()
	one := this.GetGroup(groupID)
	if one == nil {
		return
	}
	one.RemoveRequest(tag)
	if one.Length() == 0 {
		this.group.Delete(groupID)
	}
	// this.lock.Unlock()
}

/*
查询一个消息，看是否存在
*/
func (this *WaitRecvGroup) FindRequest(groupID, tag string) bool {
	one := this.GetGroup(groupID)
	if one == nil {
		return false
	}
	return one.FindRequest(tag)
}

/*
有消息内容返回了
@return    bool    是否有对应的等待
*/
func (this *WaitRecvGroup) ResponseBytes(groupID, tag string, bs *[]byte) bool {
	one := this.GetGroup(groupID)
	if one == nil {
		return false
	}
	return one.ResponseBytes(tag, bs)
}

/*
有消息内容返回了
*/
func (this *WaitRecvGroup) ResponseItr(groupID, tag string, itr interface{}) bool {
	one := this.GetGroup(groupID)
	if one == nil {
		return false
	}
	return one.ResponseItr(tag, itr)
}

/*
有一组消息内容返回了
*/
func (this *WaitRecvGroup) ResponseItrGroup(groupID string, itr interface{}) {
	one := this.GetGroup(groupID)
	if one == nil {
		return
	}
	this.group.Delete(groupID)
	cs := one.GetAllChan()
	for _, one := range *cs {
		select {
		case one <- itr:
			// utils.Log.Info().Msgf("111111111")
		default:
			// utils.Log.Info().Msgf("111111111")
		}
	}
}

/*
等待返回消息内容
*/
func (this *WaitRecvGroup) WaitResponse(groupID, tag string, timeout time.Duration) (*[]byte, error) {
	one := this.GetGroup(groupID)
	if one == nil {
		return nil, config.ERROR_wait_msg_timeout
	}
	return one.WaitResponse(tag, timeout)
}

/*
等待返回消息内容
*/
func (this *WaitRecvGroup) WaitResponseItr(groupID, tag string, timeout time.Duration) (interface{}, error) {
	one := this.GetGroup(groupID)
	if one == nil {
		return nil, config.ERROR_wait_msg_timeout
	}
	return one.WaitResponseItr(tag, timeout)
}
