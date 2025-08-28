package nodeStore

import (
	"sync"
	"web3_gui/utils"
)

type ProxyAddrManager struct {
	lock      *sync.RWMutex
	nodeInfos map[string]*NodeInfo //节点信息。key:string=AddressNet;value:*NodeInfo=;
}

func NewProxyAddrManager() *ProxyAddrManager {
	pam := ProxyAddrManager{
		lock:      new(sync.RWMutex),
		nodeInfos: make(map[string]*NodeInfo),
	}
	return &pam
}

func (this *ProxyAddrManager) Add(nodeInfo *NodeInfo) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.nodeInfos[utils.Bytes2string(nodeInfo.IdInfo.Id.Data())] = nodeInfo
}

func (this *ProxyAddrManager) Del(addr *AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.nodeInfos, utils.Bytes2string(addr.Data()))
}

/*
地址是否存在
@return    bool    true=存在;false=不存在;
*/
func (this *ProxyAddrManager) FindNodeInfoByAddr(addr *AddressNet) *NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	nodeInfo, _ := this.nodeInfos[utils.Bytes2string(addr.Data())]
	return nodeInfo
}

/*
地址是否存在
@return    bool    true=存在;false=不存在;
*/
func (this *ProxyAddrManager) ExistAddrs(addr *AddressNet) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	_, ok := this.nodeInfos[utils.Bytes2string(addr.Data())]
	return ok
}

/*
获取所有地址
*/
func (this *ProxyAddrManager) GetNodeInfos() []NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrs := make([]NodeInfo, 0, len(this.nodeInfos))
	for _, one := range this.nodeInfos {
		addrs = append(addrs, *one)
	}
	return addrs
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *ProxyAddrManager) CleanNodeInfo() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range this.nodeInfos {
		if len(one.GetSessions()) == 0 {
			delete(this.nodeInfos, utils.Bytes2string(one.IdInfo.Id.Data()))
		}
	}
}
