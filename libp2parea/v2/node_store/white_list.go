/*
连接白名单，P2P网络通信需要其他节点做中转，加入白名单功能，可以直接和指定节点建立连接
*/
package nodeStore

import (
	ma "github.com/multiformats/go-multiaddr"
	"sync"
	"web3_gui/utils"
)

/*
连接白名单管理
*/
type WhiteListManager struct {
	lock          *sync.RWMutex
	whiteListPort map[string]*WhiteListAddrInfo //端口白名单。key:string=ma.Multiaddr;value:*WhiteListAddrInfo=;
	whiteListAddr map[string]AddressNet         //地址白名单。key:string=AddressNet;value:AddressNet=;
	nodeInfos     map[string]*NodeInfo          //节点信息。key:string=AddressNet;value:*NodeInfo=;
}

func NewWhiteListManager() *WhiteListManager {
	return &WhiteListManager{
		lock:          new(sync.RWMutex),
		whiteListPort: make(map[string]*WhiteListAddrInfo),
		whiteListAddr: make(map[string]AddressNet),
		nodeInfos:     make(map[string]*NodeInfo),
	}
}

/*
添加白名单地址
*/
func (this *WhiteListManager) AddWhiteListAddr(addr *AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.whiteListAddr[utils.Bytes2string(addr.Data())] = *addr
}

/*
删除白名单地址
*/
func (this *WhiteListManager) DelWhiteListAddr(addr *AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.whiteListAddr, utils.Bytes2string(addr.Data()))
}

/*
检查节点地址是否在白名单中
*/
func (this *WhiteListManager) CheckWhiteListAddrExist(addr *AddressNet) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.whiteListAddr[utils.Bytes2string(addr.Data())]
	return ok
}

/*
添加白名单端口
*/
func (this *WhiteListManager) AddWhiteListAddrInfo(ip ma.Multiaddr) {
	addrInfo := NewWhiteListAddrInfo(nil, ip)
	this.lock.Lock()
	defer this.lock.Unlock()
	//直接保存，有则修改
	this.whiteListPort[ip.String()] = addrInfo
}

/*
删除白名单端口
*/
func (this *WhiteListManager) DelWhiteListAddrInfo(ip ma.Multiaddr) {
	addrInfo := NewWhiteListAddrInfo(nil, ip)
	this.lock.Lock()
	defer this.lock.Unlock()
	//直接保存，有则修改
	this.whiteListPort[ip.String()] = addrInfo
}

/*
检查端口是否在白名单中
*/
func (this *WhiteListManager) CheckWhiteListAddrInfoExist(ip ma.Multiaddr) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.whiteListPort[ip.String()]
	return ok
}

/*
获取白名单中所有地址
*/
func (this *WhiteListManager) GetAddrs() []AddressNet {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrs := make([]AddressNet, 0, len(this.whiteListAddr))
	for _, one := range this.whiteListAddr {
		addrs = append(addrs, one)
	}
	return addrs
}

/*
获取白名单中所有端口
*/
func (this *WhiteListManager) GetAddrInfo() []ma.Multiaddr {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrs := make([]ma.Multiaddr, 0, len(this.whiteListPort))
	for _, one := range this.whiteListPort {
		addrs = append(addrs, one.Ip)
	}
	return addrs
}

/*
检查是否需要这个会话
*/
func (this *WhiteListManager) CheckNeed(nodeInfo *NodeInfo, save bool) bool {
	if save {
		this.lock.Lock()
		defer this.lock.Unlock()
	} else {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}
	return false
}

/*
获取所有节点
*/
func (this *WhiteListManager) GetNodeInfos() []NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	nodeInfos := make([]NodeInfo, 0, len(this.nodeInfos))
	for _, one := range this.nodeInfos {
		nodeInfos = append(nodeInfos, *one)
	}
	return nodeInfos
}

/*
通过地址查询节点信息
*/
func (this *WhiteListManager) FindNodeInfoByAddr(addr *AddressNet) *NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	nodeInfo, _ := this.nodeInfos[utils.Bytes2string(addr.Data())]
	return nodeInfo
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *WhiteListManager) CleanNodeInfo() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range this.nodeInfos {
		if len(one.GetSessions()) == 0 {
			delete(this.nodeInfos, utils.Bytes2string(one.IdInfo.Id.Data()))
		}
	}
}

type WhiteListAddrInfo struct {
	Addr *AddressNet
	Ip   ma.Multiaddr
}

func NewWhiteListAddrInfo(addr *AddressNet, ip ma.Multiaddr) *WhiteListAddrInfo {
	wlai := new(WhiteListAddrInfo)
	wlai.Addr = addr
	wlai.Ip = ip
	return wlai
}
