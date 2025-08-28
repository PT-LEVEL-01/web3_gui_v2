package virtual_node

import (
	"bytes"
	"context"
	"sync"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

type VnodeManager struct {
	lock              *sync.RWMutex          //
	DiscoverVnodes    *Vnode                 //每个节点启动一个发现者虚拟节点，用于保存逻辑节点
	VnodeMap          map[uint64]*Vnode      //保存自己的虚拟节点key:uint64=index;value:*Vnode=自己的虚拟节点。
	clientVnode       *sync.Map              //所有被连接的逻辑节点 key:string=AddressNetExtend;value:Vnodeinfo=;
	findNearVnodeChan chan *[]*FindVnodeVO   //需要查找的虚拟节点
	nodeManager       *nodeStore.NodeManager //
	contextRoot       context.Context        //
	autonomyFinish    AutonomyFinish
	isClose           bool            //
	closeVnodeChan    chan *Vnodeinfo //删除虚拟逻辑节点的广播管道
}

func NewVnodeManager(nodeManager *nodeStore.NodeManager, c context.Context, af AutonomyFinish) *VnodeManager {
	vm := &VnodeManager{
		lock:              new(sync.RWMutex),
		VnodeMap:          make(map[uint64]*Vnode),
		clientVnode:       new(sync.Map),
		findNearVnodeChan: make(chan *[]*FindVnodeVO, 1000),
		nodeManager:       nodeManager,
		contextRoot:       c,
		autonomyFinish:    af,
		closeVnodeChan:    make(chan *Vnodeinfo, 100),
	}
	//添加一个发现者虚拟节点，用于发现其他虚拟节点，以及保存逻辑节点
	vm.DiscoverVnodes = NewVnode(0, nodeManager.NodeSelf.IdInfo.Id, vm.findNearVnodeChan, vm.contextRoot, af, vm.closeVnodeChan)
	vm.VnodeMap[0] = vm.DiscoverVnodes
	return vm
}

/*
关闭虚拟节点管理功能
*/
func (this *VnodeManager) IsClose() {
	this.isClose = true
}

/*
打开虚拟节点管理功能
*/
func (this *VnodeManager) IsOpen() {
	this.isClose = false
}

/*
虚拟节点管理读锁 锁操作
*/
func (this *VnodeManager) RLock() {
	this.lock.RLock()
}

/*
虚拟节点管理读锁 解锁操作
*/
func (this *VnodeManager) RUnlock() {
	this.lock.RUnlock()
}

/*
等待虚拟节点网络自治完成
*/
func (this *VnodeManager) WaitAutonomyFinish() {
	if this.isClose {
		return
	}
	this.DiscoverVnodes.WaitAutonomyFinish()
	this.lock.RLock()
	vnodesSelf := make([]*Vnode, 0, len(this.VnodeMap))
	for _, v := range this.VnodeMap {
		vnodesSelf = append(vnodesSelf, v)
	}
	this.lock.RUnlock()
	for _, one := range vnodesSelf {
		one.WaitAutonomyFinish()
	}
}

/*
获取查询vnode管道
*/
func (this *VnodeManager) GetFindVnodeChan() chan *[]*FindVnodeVO {
	return this.findNearVnodeChan
}

/*
添加一个发现者虚拟节点，用于发现其他虚拟节点，以及保存逻辑节点
*/
func (this *VnodeManager) GetVnodeDiscover() (discoverVnode *Vnode) {
	this.lock.Lock()
	discoverVnode = this.DiscoverVnodes
	this.lock.Unlock()
	return
}

/*
添加一个虚拟节点
*/
func (this *VnodeManager) AddVnode() Vnodeinfo {
	this.lock.Lock()
	defer this.lock.Unlock()
	//找到不连续的index添加
	index := uint64(0)
	ok := true
	for ok {
		index++
		_, ok = this.VnodeMap[index]
	}

	nodeOne := NewVnode(index, this.nodeManager.NodeSelf.IdInfo.Id, this.findNearVnodeChan, this.contextRoot, this.autonomyFinish, this.closeVnodeChan)
	// this.Vnodes = append(this.Vnodes, *nodeOne)
	this.VnodeMap[index] = nodeOne
	return nodeOne.Vnode
}

/*
添加一个指定下标的虚拟节点
*/
func (this *VnodeManager) AddVnodeByIndex(index uint64) *Vnodeinfo {
	this.lock.Lock()
	defer this.lock.Unlock()
	if vNode, ok := this.VnodeMap[index]; ok && vNode != nil {
		return &vNode.Vnode
	}

	nodeOne := NewVnode(index, this.nodeManager.NodeSelf.IdInfo.Id, this.findNearVnodeChan, this.contextRoot, this.autonomyFinish, this.closeVnodeChan)
	// this.Vnodes = append(this.Vnodes, *nodeOne)
	this.VnodeMap[index] = nodeOne

	//添加虚拟节点后，触发它寻找逻辑节点
	nCurVnodeLen := len(this.VnodeMap)
	if nCurVnodeLen > 1 {
		vnodes := make([]Vnodeinfo, 0, len(this.VnodeMap)-1)
		for _, vnode := range this.VnodeMap {
			if bytes.Equal(vnode.Vnode.Vid, nodeOne.Vnode.Vid) {
				continue
			}

			vnodes = append(vnodes, vnode.Vnode)
		}
		nodeOne.AddLogicVnodeinfosCheckChange(&vnodes)
	}

	return &nodeOne.Vnode
}

/*
删除一个指定下标的虚拟节点
*/
func (this *VnodeManager) DelVnodeByIndex(index uint64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	// nodeOne := NewVnode(index, this.nodeManager.NodeSelf.IdInfo.Id, this.findNearVnodeChan, this.contextRoot)
	// this.Vnodes = append(this.Vnodes, *nodeOne)
	if vnode, ok := this.VnodeMap[index]; ok {
		vnode.Destroy()
	}
	delete(this.VnodeMap, index)

}

/*
 * 指定删除一个自己创建的虚拟地址
 * 同时删除本节点其他vnode中存此地址的信息
 */
func (this *VnodeManager) DelSelfVnodeByAddr(vid nodeStore.AddressNet) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i, _ := range this.VnodeMap {
		if bytes.Equal(this.VnodeMap[i].Vnode.Vid, vid) && this.VnodeMap[i].Vnode.Index != 0 {
			this.VnodeMap[i].Destroy()
			delete(this.VnodeMap, i)
		} else {
			this.VnodeMap[i].VnodeDelAddr(AddressNetExtend(vid))
		}
	}
}

/*
 * 指定删除一个在up down排序中的别人的虚拟地址
 */
func (this *VnodeManager) DelUpdownVnodeByAddr(vid AddressNetExtend) {
	//检查，不许删除自己创建的虚拟地址
	if this.FindInVnodeSelf(vid) {
		return
	}

	//删发现节点的逻辑节点
	this.DiscoverVnodes.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		if bytes.Equal(vnodeinfo.Vid, vid) {
			this.DiscoverVnodes.LogicalNode.Delete(utils.Bytes2string(vid))
		}
		return true
	})

	//删clientVnode中的vid信息
	this.clientVnode.Delete(utils.Bytes2string(vid))

	this.lock.RLock()
	defer this.lock.RUnlock()
	//删本节点其他虚拟节点中的vid信息
	for i, _ := range this.VnodeMap {
		this.VnodeMap[i].VnodeDelAddr(vid)
	}
}

/*
删除一个虚拟节点
*/
func (this *VnodeManager) DelVnode() (vnodeinfo *Vnodeinfo) {
	this.lock.Lock()
	defer this.lock.Unlock()

	//找到不连续的index添加
	index := uint64(0)
	ok := true
	for ok {
		index++
		_, ok = this.VnodeMap[index]
	}
	if vnode, ok := this.VnodeMap[index]; ok {
		vnodeinfo = &vnode.Vnode
		vnode.Destroy()
	}
	delete(this.VnodeMap, index)

	// index := uint64(len(this.Vnodes))

	// newvnodes := make([]Vnode, 0)

	// for i, _ := range this.Vnodes {
	// 	if uint64(i+1) >= index {
	// 		nodeinfo = this.Vnodes[i].Vnode
	// 		break
	// 	}
	// 	newvnodes = append(newvnodes, this.Vnodes[i])
	// }
	// this.Vnodes = newvnodes
	return
}

/*
调整云存储大小，多了的就减少，少了的就增加。
*/
func (this *VnodeManager) SetupVnodeNumber(n uint64) {
	// if n <= 0 {
	// 	return
	// }
	this.lock.Lock()
	defer this.lock.Unlock()
	//空间大小合适，不需要调整
	if uint64(len(this.VnodeMap)) == n {
		return
	}
	//空间太大，需要减少空间
	if uint64(len(this.VnodeMap)) > n {
		count := uint64(len(this.VnodeMap)) - n
		//随机删除
		for k, _ := range this.VnodeMap {
			if count == 0 {
				break
			}
			count--
			if vnode, ok := this.VnodeMap[k]; ok {
				vnode.Destroy()
			}
			delete(this.VnodeMap, k)
		}
		// newvnodes := make([]Vnode, 0)
		// for i, _ := range this.Vnodes {
		// 	if uint64(i+1) > n {
		// 		break
		// 	}
		// 	newvnodes = append(newvnodes, this.Vnodes[i])
		// 	delete(this.VnodeMap, this.Vnodes[i].Vnode.Index)
		// }
		// this.Vnodes = newvnodes
	} else {
		//空间太小，需要增加空间。
		//找到不连续的index添加
		for i := uint64(len(this.VnodeMap)); i < n; i++ {
			index := uint64(0)
			ok := true
			for ok {
				index++
				_, ok = this.VnodeMap[index]
			}
			nodeOne := NewVnode(index, this.nodeManager.NodeSelf.IdInfo.Id, this.findNearVnodeChan, this.contextRoot, this.autonomyFinish, this.closeVnodeChan)
			// this.Vnodes = append(this.Vnodes, *nodeOne)
			this.VnodeMap[index] = nodeOne
		}
		// for i := uint64(len(this.Vnodes)); i < n; i++ {
		// 	nodeOne := NewVnode(i, this.nodeManager.NodeSelf.IdInfo.Id, this.findNearVnodeChan, this.contextRoot)
		// 	this.Vnodes = append(this.Vnodes, *nodeOne)
		// 	this.VnodeMap[i] = nodeOne
		// }
	}

}

/*
查询扩展的虚拟节点数量
*/
func (this *VnodeManager) GetVnodeNumber() []Vnodeinfo {
	this.lock.RLock()
	vnodeinfo := make([]Vnodeinfo, 0, len(this.VnodeMap))
	for _, one := range this.VnodeMap {
		if one.Vnode.Index == 0 {
			continue
		}
		vnodeinfo = append(vnodeinfo, one.Vnode)
	}
	// vnodeinfo := make([]Vnodeinfo, 0, len(this.Vnodes))
	// for _, one := range this.Vnodes {
	// 	vnodeinfo = append(vnodeinfo, one.Vnode)
	// }
	this.lock.RUnlock()
	return vnodeinfo
}

/*
添加虚拟节点的逻辑节点
@return    bool    是否有改变true=有改变;false=无改变。
*/
func (this *VnodeManager) AddLogicVnodeinfo(vnodes ...Vnodeinfo) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	change := this.DiscoverVnodes.AddLogicVnodeinfosCheckChange(&vnodes)
	for _, one := range this.VnodeMap {
		//有一个改变，则返回有改变
		if one.AddLogicVnodeinfosCheckChange(&vnodes) {
			change = true
		}
	}
	return change
}

/*
添加被连接的虚拟节点
*/
func (this *VnodeManager) AddClientVnodeinfo(vnode Vnodeinfo) {
	// utils.Log.Info().Msgf("vnodeinfo:%+v", vnode)
	vnodeinfoTimeout := CreateVnodeinfoTimeout(&vnode)
	vinfotItr, ok := this.clientVnode.LoadOrStore(utils.Bytes2string(vnode.Vid), vnodeinfoTimeout)
	if ok {
		vnodeinfoTimeout = vinfotItr.(*VnodeinfoTimeout)
		vnodeinfoTimeout.FlashTime()
	}
}

/*
添加被连接的虚拟节点
*/
func (this *VnodeManager) GetClientVnodeinfo() []*Vnodeinfo {
	vnodes := make([]*Vnodeinfo, 0)
	this.clientVnode.Range(func(k, v interface{}) bool {
		vnodeinfoOne, ok := v.(*VnodeinfoTimeout)
		if !ok {
			return false
		}
		if vnodeinfoOne.CheckTimeout(config.VNODE_heartbeat_timeout) {
			return true
		}
		vnodes = append(vnodes, &vnodeinfoOne.Vnodeinfo)
		return true
	})
	return vnodes
}

/*
 * 根据真实节点删除客户端连接信息
 */
func (this *VnodeManager) DelClientVnodeinfo(nid nodeStore.AddressNet) {
	this.clientVnode.Range(func(k, v interface{}) bool {
		name, ok := k.(string)
		if !ok {
			return false
		}
		vnodeinfoOnde := v.(*VnodeinfoTimeout)
		if bytes.Equal(vnodeinfoOnde.Nid, nid) {
			this.clientVnode.Delete(name)
		}

		return true
	})
}

/*
添加虚拟节点的多个逻辑节点
*/
func (this *VnodeManager) AddLogicVnodeinfos(nodeVid AddressNetExtend, vnodes *[]Vnodeinfo) (change bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if bytes.Equal(this.DiscoverVnodes.Vnode.Nid, nodeVid) {
		// utils.Log.Info().Msgf("给发现节点添加逻辑节点:%s", nodeVid.B58String())
		if !this.DiscoverVnodes.AddLogicVnodeinfosCheckChange(vnodes) {
			// utils.Log.Info().Msgf("自治完成:%s", this.DiscoverVnodes.Vnode.Nid.B58String())
			//如果添加虚拟逻辑节点，未改变，则自治完成。
			this.DiscoverVnodes.SetAutonomyFinish()
		}
		for _, one := range this.VnodeMap {
			one.AddLogicVnodeinfosCheckChange(vnodes)
		}
	} else {
		// utils.Log.Info().Msgf("给虚拟节点添加逻辑节点:%s", nodeVid.B58String())
		this.DiscoverVnodes.AddLogicVnodeinfosCheckChange(vnodes)
		for _, one := range this.VnodeMap {
			if bytes.Equal(one.Vnode.Vid, nodeVid) {
				if !one.AddLogicVnodeinfosCheckChange(vnodes) {
					// utils.Log.Info().Msgf("自治完成:%s", this.DiscoverVnodes.Vnode.Nid.B58String())
					//如果添加虚拟逻辑节点，未改变，则自治完成。
					one.SetAutonomyFinish()
				}
			} else {
				one.AddLogicVnodeinfosCheckChange(vnodes)
			}
		}
	}

	// ok = this.DiscoverVnodes.AddLogicVnodeinfos(vnodes)
	// for _, one := range this.VnodeMap {
	// 	if success := one.AddLogicVnodeinfos(vnodes); success {
	// 		ok = true
	// 	}
	// }
	return
}

/*
获得所有节点，包括自己节点
*/
func (this *VnodeManager) GetVnodeAll() map[string]Vnodeinfo {

	vnodeinfoMap := make(map[string]Vnodeinfo)
	vnodeinfoMap[utils.Bytes2string(this.DiscoverVnodes.Vnode.Vid)] = this.DiscoverVnodes.Vnode
	this.DiscoverVnodes.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		vnodeinfoMap[utils.Bytes2string(vnodeinfo.Vid)] = vnodeinfo
		return true
	})

	this.lock.RLock()
	defer this.lock.RUnlock()

	for _, one := range this.VnodeMap {
		if one.Vnode.Index == 0 {
			continue
		}
		selfOne := one.GetSelfVnodeinfo()
		vnodeinfoMap[utils.Bytes2string(selfOne.Vid)] = selfOne

		one.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeinfo, ok := v.(Vnodeinfo)
			if !ok {
				return false
			}
			vnodeinfoMap[utils.Bytes2string(vnodeinfo.Vid)] = vnodeinfo
			return true
		})
	}

	//删除自己节点
	// for _, one := range this.Vnodes {
	// 	delete(vnodeinfoMap, utils.Bytes2string(one.Vnode.Vid))
	// }
	return vnodeinfoMap
}

/*
获得自己管理的节点info
*/
func (this *VnodeManager) GetVnodeSelf() []Vnodeinfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	vnodeinfo := make([]Vnodeinfo, 0, len(this.VnodeMap)+1)
	vnodeinfo = append(vnodeinfo, this.DiscoverVnodes.Vnode)
	for _, one := range this.VnodeMap {
		if one.Vnode.Index == 0 {
			continue
		}
		vnodeinfo = append(vnodeinfo, one.Vnode)
	}
	return vnodeinfo
}

/*
获得自己管理的节点
*/
func (this *VnodeManager) FindVnodeInSelf(nodeVid AddressNetExtend) *Vnode {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if bytes.Equal(this.DiscoverVnodes.Vnode.Vid, nodeVid) {
		return this.DiscoverVnodes
	}
	for _, vnode := range this.VnodeMap {
		if bytes.Equal(vnode.Vnode.Vid, nodeVid) {
			return vnode
		}
	}
	return nil
}

/*
在逻辑节点中查找Vnodeinfo
*/
func (this *VnodeManager) FindVnodeinfo(vid AddressNetExtend) *Vnodeinfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if bytes.Equal(vid, this.DiscoverVnodes.Vnode.Vid) {
		return &this.DiscoverVnodes.Vnode
	}
	v, ok := this.DiscoverVnodes.LogicalNode.Load(utils.Bytes2string(vid))
	if ok {
		vnodeinfo := v.(Vnodeinfo)
		return &vnodeinfo
	}
	for _, one := range this.VnodeMap {
		vnodeinfo := one.FindVnodeinfo(vid)
		if vnodeinfo == nil {
			continue
		}
		return vnodeinfo
	}
	return nil
}

/*
查找节点id是否是自己的节点
@return    bool    是否在
*/
func (this *VnodeManager) FindInVnodeSelf(id AddressNetExtend) bool {
	// if bytes.Equal(id, this.DiscoverVnodes.Vnode.Vid) {
	// 	return true
	// }
	for _, one := range this.GetVnodeSelf() {
		if bytes.Equal(one.Vid, id) {
			return true
		}
	}
	return false
}

/*
查找自己一个虚拟节点的逻辑节点
*/
func (this *VnodeManager) FindLogicInVnodeSelf(id AddressNetExtend) []Vnodeinfo {
	vnodeinfos := make([]Vnodeinfo, 0)
	if bytes.Equal(id, this.DiscoverVnodes.Vnode.Vid) {
		this.DiscoverVnodes.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeinfo, ok := v.(Vnodeinfo)
			if !ok {
				return false
			}
			vnodeinfos = append(vnodeinfos, vnodeinfo)
			return true
		})
		return vnodeinfos
	}
	for _, one := range this.GetVnodeSelf() {
		if bytes.Equal(one.Vid, id) {
			this.DiscoverVnodes.LogicalNode.Range(func(k, v interface{}) bool {
				vnodeinfo, ok := v.(Vnodeinfo)
				if !ok {
					return false
				}
				vnodeinfos = append(vnodeinfos, vnodeinfo)
				return true
			})
			return vnodeinfos
		}
	}
	return vnodeinfos
}

/*
查找节点id是否是自己的节点
@return    bool    是否在
*/
func (this *VnodeManager) FindInVnodeinfoSelf(id AddressNetExtend) *Vnodeinfo {
	// if bytes.Equal(id, this.DiscoverVnodes.Vnode.Vid) {
	// 	return &this.DiscoverVnodes.Vnode
	// }
	for i, one := range this.GetVnodeSelf() {
		if bytes.Equal(one.Vid, id) {
			return &this.GetVnodeSelf()[i]
		}
	}
	return nil
}

/*
查找节点id是否是自己的节点
@return    bool    是否在
*/
func (this *VnodeManager) FindVnodeInAllSelf(id nodeStore.AddressNet) *Vnode {
	// if bytes.Equal(id, this.DiscoverVnodes.Vnode.Vid) {
	// 	return &this.DiscoverVnodes.Vnode
	// }
	this.lock.Lock()
	defer this.lock.Unlock()

	for i, one := range this.VnodeMap {
		if bytes.Equal(one.Vnode.Vid, id) {
			return this.VnodeMap[i]
		}
	}
	return nil
}

/*
关闭所有虚拟节点
*/
func (this *VnodeManager) Close() {

}

/*
获取关闭vnode管道
*/
func (this *VnodeManager) GetCloseVnodeChan() chan *Vnodeinfo {
	return this.closeVnodeChan
}

/*
 * 检查是否包含指定的真实节点信息
 */
func (this *VnodeManager) CheckNodeinfoExistInSelf(nid nodeStore.AddressNet) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if bytes.Equal(nid, this.DiscoverVnodes.Vnode.Nid) {
		return true
	}
	exist := this.DiscoverVnodes.CheckVnodeinfoExistByNid(nid)
	if exist {
		return true
	}

	for _, one := range this.VnodeMap {
		exist = one.CheckVnodeinfoExistByNid(nid)
		if exist {
			return true
		}
	}

	return false
}

/*
* 检查是否是逻辑节点，虚拟节点，自己虚拟节点up down中是否包含
 */
func (this *VnodeManager) IsSelfVnodeNeed(nid nodeStore.AddressNet) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if bytes.Equal(nid, this.DiscoverVnodes.Vnode.Nid) {
		return true
	}
	exist := this.DiscoverVnodes.CheckVnodeinfoExistByNid(nid)
	if exist {
		return true
	}

	for _, one := range this.VnodeMap {
		exist = one.CheckVnodeinfoExistByNid(nid)
		if exist {
			return true
		}
		for _, one := range one.GetOnebyoneVnodeInfo() {
			if bytes.Equal(one.Nid, nid) || bytes.Equal(one.Vid, nid) {
				return true
			}
		}
	}

	return false
}

/*
 * 通过一个指定下标获取虚拟节点地址
 *
 * @param	index	uint64				下标
 * @return	vid		AddressNetExtend	节点地址
 */
func (this *VnodeManager) GetVnodeIdByIndex(index uint64) *AddressNetExtend {
	vnodeInfo := BuildNodeinfo(index, this.nodeManager.NodeSelf.IdInfo.Id)

	return &vnodeInfo.Vid
}
