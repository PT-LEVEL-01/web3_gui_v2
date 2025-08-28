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
type LogicAddrManager struct {
	self    *AddressNet          //基准域名称
	lock    *sync.RWMutex        //锁
	nodeMap map[string]*NodeInfo //节点信息，当T类型是AddressNet时，需要保存节点信息
	log     *zerolog.Logger      //日志
}

/*
从新设置日志库
*/
func (this *LogicAddrManager) SetLog(log *zerolog.Logger) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.log = log
}

func NewLogicAddrManager(self *AddressNet, log *zerolog.Logger) *LogicAddrManager {
	lm := &LogicAddrManager{
		self:    self,
		lock:    new(sync.RWMutex),
		nodeMap: make(map[string]*NodeInfo),
		log:     log,
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
func (this *LogicAddrManager) CheckNeedAddr(nodeInfo *NodeInfo, save bool) (bool, []*NodeInfo, bool) {
	//(*this.Log).Info().Str("保存连接", nodeInfo.IdInfo.Id.B58String()).Int("len", len(nodeInfo.GetSessions())).Send()

	//不添加自己
	if bytes.Equal(nodeInfo.IdInfo.Id.Data(), this.self.Data()) {
		return false, nil, false
	}
	idm := NewKademlia(this.self.Data(), NodeIdLevel, this.log)
	this.lock.Lock()
	defer this.lock.Unlock()
	//判断重复地址
	nodeInfoExist, ok := this.nodeMap[utils.Bytes2string(nodeInfo.IdInfo.Id.Data())]
	if ok {
		if nodeInfo != nil {
			for _, sessionOne := range nodeInfo.GetSessions() {
				nodeInfoExist.AddSession(sessionOne)
			}
			//nodeInfoExist.Sessions[utils.Bytes2string(nodeInfo.MachineID)] = nodeInfo.Sessions[utils.Bytes2string(nodeInfo.MachineID)]
		}
		return false, nil, true
	}
	for _, one := range this.nodeMap {
		idm.AddId(one.IdInfo.Id.Data())
	}
	ok, removeIDs := idm.AddId(nodeInfo.IdInfo.Id.Data())
	//if removeIDs != nil && len(removeIDs) > 0 {
	//	this.Log.Info().Interface("要删除的地址", AddressNet(removeIDs[0]).B58String()).Send()
	//}
	removeNodeInfo := make([]*NodeInfo, 0, len(removeIDs))
	//removeIds := make([]AddressNet, 0, len(removeIDs))
	for _, id := range removeIDs {

		//removeIds = append(removeIds, AddressNet(id))
		nodeInfo, ok := this.nodeMap[utils.Bytes2string(id)]
		if ok {
			removeNodeInfo = append(removeNodeInfo, nodeInfo)
		}
	}
	//需要保存
	if ok && save {
		//(*this.Log).Info().Str("保存连接", nodeInfo.IdInfo.Id.B58String()).Int("len", len(nodeInfo.GetSessions())).
		//	Hex("sid", nodeInfo.GetSessions()[0].GetId()).Send()
		this.nodeMap[utils.Bytes2string(nodeInfo.IdInfo.Id.Data())] = nodeInfo
		for _, one := range removeNodeInfo {
			delete(this.nodeMap, utils.Bytes2string(one.IdInfo.Id.Data()))
		}
	}
	return ok, removeNodeInfo, false
}

/*
获取所有逻辑节点
*/
func (this *LogicAddrManager) GetNodeInfos() []NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ids := make([]NodeInfo, 0, len(this.nodeMap))
	for _, addr := range this.nodeMap {
		ids = append(ids, *addr)
	}
	return ids
}

/*
地址是否存在
@return    bool    true=存在;false=不存在;
*/
func (this *LogicAddrManager) ExistAddrs(nodeInfo *NodeInfo) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	_, ok := this.nodeMap[utils.Bytes2string(nodeInfo.IdInfo.Id.Data())]
	return ok
}

/*
通过地址查询节点信息
*/
func (this *LogicAddrManager) FindNodeInfoByAddr(addr *AddressNet) *NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	nodeInfo, _ := this.nodeMap[utils.Bytes2string(addr.Data())]
	return nodeInfo
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *LogicAddrManager) CleanNodeInfo() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range this.nodeMap {
		if len(one.GetSessions()) == 0 {
			delete(this.nodeMap, utils.Bytes2string(one.IdInfo.Id.Data()))
		}
	}
}
