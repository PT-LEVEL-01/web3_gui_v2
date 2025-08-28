package file_transfer

import (
	"sync"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
白名单管理
*/
type WhiteList struct {
	classID   uint64                           //
	db        *utilsleveldb.LevelDB            //
	lock      *sync.RWMutex                    //
	whiteList map[string]*nodeStore.AddressNet //白名单地址
}

/*
创建一个白名单地址管理列表
*/
func NewWhiteList(classID uint64, db *utilsleveldb.LevelDB) *WhiteList {
	w := WhiteList{
		classID:   classID,
		db:        db,
		lock:      new(sync.RWMutex),
		whiteList: make(map[string]*nodeStore.AddressNet),
	}
	return &w
}

/*
添加一个白名单地址
*/
func (this *WhiteList) AddAddr(addr nodeStore.AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.whiteList[utils.Bytes2string(addr.GetAddr())]
	if ok {
		return
	}
	this.whiteList[utils.Bytes2string(addr.GetAddr())] = &addr
	list := make([]*nodeStore.AddressNet, 0, len(this.whiteList))
	for _, one := range this.whiteList {
		list = append(list, one)
	}
	saveWhiteList(this.db, this.classID, list)

}

/*
删除一个白名单地址
*/
func (this *WhiteList) DelAddr(addr nodeStore.AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.whiteList[utils.Bytes2string(addr.GetAddr())]
	if !ok {
		return
	}
	delete(this.whiteList, utils.Bytes2string(addr.GetAddr()))
	list := make([]*nodeStore.AddressNet, 0, len(this.whiteList))
	for _, one := range this.whiteList {
		list = append(list, one)
	}
	saveWhiteList(this.db, this.classID, list)
}

/*
验证一个白名单地址
*/
func (this *WhiteList) CheckAddr(addr nodeStore.AddressNet) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if len(this.whiteList) == 0 {
		return true
	}
	_, ok := this.whiteList[utils.Bytes2string(addr.GetAddr())]
	if ok {
		return true
	}
	return false
}
