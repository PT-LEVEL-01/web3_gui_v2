package nodeStore

import (
	"bytes"
	"github.com/rs/zerolog"
	"sync"
	"web3_gui/utils"
)

/*
管理域的逻辑域名称
*/
type LogicAreaManager struct {
	self       AddressArea            //基准域名称
	lock       *sync.RWMutex          //锁
	logicAddrs map[string]AddressArea //基准域名称的逻辑域名称
	log        *zerolog.Logger        //日志
}

func NewLogicAreaManager(self AddressArea, log *zerolog.Logger) *LogicAreaManager {
	lm := &LogicAreaManager{
		self:       self,
		lock:       new(sync.RWMutex),
		logicAddrs: make(map[string]AddressArea),
		log:        log,
	}
	return lm
}

/*
检查是否需要此地址
@addr       *AddressNet    要检查的地址
@save       bool           是否保存
@return     bool           是否需要此地址
@return     [][]byte       要删除的地址
@return     bool           地址是否已经存在
*/
func (this *LogicAreaManager) CheckNeedAddr(addr AddressArea, save bool) (bool, []AddressArea, bool) {
	//不添加自己
	if bytes.Equal(addr, this.self) {
		return false, nil, false
	}
	//if save {
	//	this.log.Info().Hex("创建一个添加逻辑域", this.self).Send()
	//}
	idm := NewKademlia(this.self, NodeIdLevel, this.log)
	if save {
		this.lock.Lock()
		defer this.lock.Unlock()
	} else {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}
	//判断重复地址
	_, ok := this.logicAddrs[utils.Bytes2string(addr)]
	if ok {
		return false, nil, true
	}
	for _, addr := range this.logicAddrs {
		//if save {
		//	this.log.Info().Hex("添加逻辑域", addr).Send()
		//}
		idm.AddId(addr)
	}
	//if save {
	//	this.log.Info().Hex("添加逻辑域", addr).Send()
	//}
	ok, removeIDs := idm.AddId(addr)
	//if removeIDs != nil && len(removeIDs) > 0 {
	//	this.Log.Info().Interface("要删除的地址", AddressArea(removeIDs[0]).B58String()).Send()
	//}
	//removeIDStr := make([]string, 0, len(removeIDs))
	removeIds := make([]AddressArea, 0, len(removeIDs))
	for _, id := range removeIDs {
		removeIds = append(removeIds, AddressArea(id))
		//removeIDStr = append(removeIDStr, hex.EncodeToString(id))
	}
	//if save {
	//	this.log.Info().Interface("要删除的逻辑域", removeIDStr).Send()
	//}
	//需要保存
	if ok && save {
		this.logicAddrs[utils.Bytes2string(addr)] = addr
		for _, one := range removeIds {
			delete(this.logicAddrs, utils.Bytes2string(one))
		}
	}
	return ok, removeIds, false
}

/*
获取所有逻辑域名称
*/
func (this *LogicAreaManager) GetAddrs() []AddressArea {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ids := make([]AddressArea, 0, len(this.logicAddrs))
	for _, addr := range this.logicAddrs {
		ids = append(ids, addr)
	}
	return ids
}

/*
地址是否存在
@return    bool    true=存在;false=不存在;
*/
func (this *LogicAreaManager) ExistAddrs(addr AddressArea) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	_, ok := this.logicAddrs[utils.Bytes2string(addr)]
	return ok
}

/*
从新设置日志库
*/
func (this *LogicAreaManager) SetLog(log *zerolog.Logger) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.log = log
}
