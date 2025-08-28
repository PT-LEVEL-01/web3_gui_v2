package utils

import "sync"

type Once struct {
	lock *sync.RWMutex //锁
	run  bool          //是否在运行
}

func NewOnce() *Once {
	once := Once{
		lock: new(sync.RWMutex),
		run:  false,
	}
	return &once
}

/*
尝试运行，如果是运行状态则返回false
*/
func (this *Once) TryRun(f func()) bool {
	this.lock.Lock()
	if this.run == true {
		this.lock.Unlock()
		return false
	} else {
		this.run = true
		this.lock.Unlock()
	}
	f()
	this.lock.Lock()
	this.run = false
	this.lock.Unlock()
	return true
}
