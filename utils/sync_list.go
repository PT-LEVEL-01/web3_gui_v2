package utils

import (
	"sync"
)

type SyncList struct {
	lock *sync.RWMutex
	list []interface{}
}

func (this *SyncList) Add(obj interface{}) {
	this.lock.Lock()
	this.list = append(this.list, obj)
	this.lock.Unlock()
}

func (this *SyncList) Get(index int) (n interface{}) {
	this.lock.RLock()
	n = this.list[index]
	this.lock.RUnlock()
	return
}

func (this *SyncList) Remove(index int) (ok bool) {
	this.lock.Lock()
	if len(this.list) < index+1 {
		ok = false
		this.lock.Unlock()
		return
	}
	temp := this.list[:index]
	this.list = append(temp, this.list[index+1:]...)
	ok = true
	this.lock.Unlock()
	return
}

func (this *SyncList) GetAll() []interface{} {
	this.lock.RLock()
	l := make([]interface{}, len(this.list))
	copy(l, this.list)
	this.lock.RUnlock()
	return l
}

//func (this *SyncList) Range(fn func(i int, v interface{}) bool) {
//	this.lock.Lock()
//	for i, one := range this.list {
//		if !fn(i, one) {
//			break
//		}
//	}
//	this.lock.Unlock()
//}

//func (this *SyncList) Find() {
//	for _,one := range this.list{
//		if
//	}
//}

func NewSyncList() *SyncList {
	return &SyncList{
		lock: new(sync.RWMutex),
		list: make([]interface{}, 0),
	}
}
