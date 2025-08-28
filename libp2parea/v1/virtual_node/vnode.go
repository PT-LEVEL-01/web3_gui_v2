package virtual_node

import (
	"bytes"
	"context"
	"github.com/gogo/protobuf/proto"
	"math/big"
	"sort"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

type AutonomyFinish interface {
	WaitAutonomyFinish()
}

/*
虚拟节点
*/
type Vnode struct {
	Vnode               Vnodeinfo            //自己的虚拟节点
	LogicalNode         *sync.Map            //逻辑节点 key:string=AddressNetExtend;value:Vnodeinfo=;
	LogicalNodeIndexNID *sync.Map            //逻辑节点 key:string=AddressNet;value:map[AddressNetExtend]Vnodeinfo=;
	clientVnode         *sync.Map            //被连接的逻辑节点 key:string=AddressNetExtend;value:Vnodeinfo=;
	findVnodeChan       chan *[]*FindVnodeVO //让外部网络接口查找逻辑节点的管道
	contextRoot         context.Context      //
	canceRoot           context.CancelFunc   //
	findVnodeTag        chan bool            //让本节点触发查询邻居节点的信号
	autonomyFinish      AutonomyFinish       //
	autonomyFinishChan  chan bool            //本虚拟节点是否完成自治管理
	closeVnodeChan      chan *Vnodeinfo      //让外部网络接口广播删除虚拟逻辑节点的管道
	UpVnodeInfo         *sync.Map            //虚拟节点连接的大于自己的节点信息
	DownVnodeInfo       *sync.Map            //虚拟节点连接的小于自己的节点信息
	HeadTail            bool
	Lock                *sync.RWMutex //调整up down
}

/*
创建一个虚拟节点
*/
func NewVnode(index uint64, addrNet nodeStore.AddressNet, findNearVnodeChan chan *[]*FindVnodeVO, c context.Context, af AutonomyFinish, closeVnodeChan chan *Vnodeinfo) *Vnode {
	ctx, cancel := context.WithCancel(c)
	vnodeInfo := BuildNodeinfo(index, addrNet)
	// utils.Log.Info().Msgf("创建一个虚拟节点:%s %d %s", addrNet.B58String(), index, vnodeInfo.Vid.B58String())
	vnode := Vnode{
		Vnode:               *vnodeInfo,         //自己的虚拟节点
		LogicalNode:         new(sync.Map),      //逻辑节点
		LogicalNodeIndexNID: new(sync.Map),      //
		clientVnode:         new(sync.Map),      //
		findVnodeChan:       findNearVnodeChan,  //
		contextRoot:         ctx,                //
		canceRoot:           cancel,             //
		findVnodeTag:        make(chan bool, 1), //
		autonomyFinish:      af,                 //
		autonomyFinishChan:  make(chan bool, 1), //
		closeVnodeChan:      closeVnodeChan,     // 删除虚拟节点对外告知通道
		Lock:                new(sync.RWMutex),
		UpVnodeInfo:         new(sync.Map),
		DownVnodeInfo:       new(sync.Map),
	}
	vnode.Run()
	return &vnode
}

func (this *Vnode) Run() {
	// utils.Log.Info().Msgf("运行VNODE run:%s", this.Vnode.Vid.B58String())
	// go this.printLogicNode()
	go this.SearchVnode()
}

/*
销毁这个虚拟节点
*/
func (this *Vnode) Destroy() {
	this.canceRoot()
	this.closeVnodeChan <- &this.Vnode
}

/*
等待网络自治完成
*/
func (this *Vnode) WaitAutonomyFinish() {
	// 超时自动完成
	timeout := time.NewTimer(time.Second * 60)
	select {
	case <-this.autonomyFinishChan:
		timeout.Stop()
	case <-timeout.C:
	case <-this.contextRoot.Done():
		// area销毁,如果vnode自治还没有成功，则直接返回
		timeout.Stop()
		return
	}
	select {
	case this.autonomyFinishChan <- false:
	default:
	}
}

/*
定时打印
*/
func (this *Vnode) printLogicNode() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		utils.Log.Info().Msgf("打印节点:%s", this.Vnode.Vid.B58String())
		this.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeInfo, ok := v.(Vnodeinfo)
			if !ok {
				return false
			}
			utils.Log.Info().Msgf("ID:%s LogicID:%s index:%d", this.Vnode.Vid.B58String(), vnodeInfo.Vid.B58String(), vnodeInfo.Index)
			return true
		})
	}
}

/*
定时搜索其他虚拟节点
*/
func (this *Vnode) SearchVnode() {
	//等待逻辑节点自治完成
	this.autonomyFinish.WaitAutonomyFinish()
	// utils.Log.Info().Msgf("SearchVnode 节点自治完成")
	timer := time.NewTimer(time.Second)
	for {
		timer.Reset(time.Minute * 100)
		select {
		case <-timer.C:
			// case <-time.After(time.Minute * 100):
			// utils.Log.Info().Msgf("SearchVnode 11111111111111111:%s", this.Vnode.Nid.B58String())
		case <-this.contextRoot.Done():
			// utils.Log.Info().Msgf("SearchVnode done!")
			timer.Stop()
			return
		case <-this.findVnodeTag:
			// utils.Log.Info().Msgf("SearchVnode 1111111")
			// utils.Log.Info().Msgf("SearchVnode 11111 self:%s", this.Vnode.Nid.B58String())
			// total = 0
			timer.Stop()
		}
		fvVO := make([]*FindVnodeVO, 0)
		//从自己的逻辑节点查找，创建一个Target为空的FindVnodeVO表示从逻辑节点查找
		fvv := &FindVnodeVO{
			Self: this.Vnode,
			// Target: nil,
		}
		fvVO = append(fvVO, fvv)
		//再给自己的虚拟邻居节点发送查询消息
		this.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeinfo, ok := v.(Vnodeinfo)
			if !ok {
				return false
			}
			fvv := &FindVnodeVO{
				Self:   this.Vnode,
				Target: vnodeinfo,
			}
			fvVO = append(fvVO, fvv)
			return true
		})
		// utils.Log.Info().Msgf("SearchVnode 22222 self:%s", this.Vnode.Nid.B58String())
		select {
		case this.findVnodeChan <- &fvVO:
		case <-this.contextRoot.Done():
			return
		}
	}
}

/*
对比逻辑节点是否有变化
*/
func (this *Vnode) EqualLogicNodes(oldNodeinfo []Vnodeinfo) bool {
	for _, one := range oldNodeinfo {
		_, ok := this.LogicalNode.Load(utils.Bytes2string(one.Vid))
		if !ok {
			return true
		}
	}
	return false
}

/*
设置自治完成
*/
func (this *Vnode) SetAutonomyFinish() {
	select {
	case this.autonomyFinishChan <- false:
	default:
	}
}

/*
查看是否自治成功
*/
func (this *Vnode) CheckAutonomyFinish() bool {
	if len(this.autonomyFinishChan) != 0 {
		return true
	}
	return false
}

/*
添加一个完整的Vnodes集合，可以判断此节点是否自治完成
@return    bool    是否有改变true=有改变;false=无改变。
*/
func (this *Vnode) AddLogicVnodeinfosCheckChange(vnodes *[]Vnodeinfo) (change bool) {
	change = this.AddLogicVnodeinfos(vnodes)
	if change {
		select {
		case this.findVnodeTag <- false:
			// utils.Log.Info().Msgf("有虚拟节点变动:%s", this.Vnode.Vid.B58String())
		default:
		}
	}
	return
}

/*
添加一个完整的Vnodes集合，可以判断此节点是否自治完成
@return    bool    是否有改变true=有改变;false=无改变。
*/
func (this *Vnode) AddLogicVnodeinfos(vnodes *[]Vnodeinfo) (change bool) {
	change = false
	for _, one := range *vnodes {
		if this.addLogicVnodeinfo(one) {
			change = true
		}
	}
	return
}

/*
添加Vnode
@return    bool    是否有改变true=有改变;false=无改变。
*/
func (this *Vnode) addLogicVnodeinfo(vnode Vnodeinfo) (change bool) {
	// 所有的节点都不添加真实节点地址信息
	if vnode.Index == 0 {
		// utils.Log.Error().Msgf("qlw-----addLogicVnodeinfo add vid is invalid cur:%s, vid:%s, nid:%s, index:%d", this.Vnode.Nid.B58String(), vnode.Vid.B58String(), vnode.Nid.B58String(), vnode.Index)
		return false
	}

	//不能添加自己
	if bytes.Equal(vnode.Vid, this.Vnode.Vid) {
		// utils.Log.Info().Msgf("自己不能添加自己:%s %s", this.Vnode.Vid.B58String(), vnode.Vid.B58String())
		return false
	}
	// utils.Log.Info().Msgf("虚拟节点 添加逻辑节点:%s %s", this.Vnode.Vid.B58String(), vnode.Vid.B58String())

	idm := nodeStore.NewIds(this.Vnode.Vid, nodeStore.NodeIdLevel)
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		idm.AddId(vnodeinfo.Vid)
		return true
	})
	ok, removeIDs := idm.AddId(vnode.Vid)
	if ok {
		this.LogicalNode.Store(utils.Bytes2string(vnode.Vid), vnode)
		nidMapItr, ok := this.LogicalNodeIndexNID.Load(utils.Bytes2string(vnode.Nid))
		if ok {
			nidMap := nidMapItr.(*sync.Map)
			nidMap.Store(utils.Bytes2string(vnode.Vid), &vnode)
		} else {
			nidMap := new(sync.Map)
			nidMap.Store(utils.Bytes2string(vnode.Vid), &vnode)
			this.LogicalNodeIndexNID.Store(utils.Bytes2string(vnode.Nid), nidMap)
		}
		//删除被替换的id
		for _, one := range removeIDs {
			addrNetExtend := AddressNetExtend(one)
			vnodeItr, ok := this.LogicalNode.Load(utils.Bytes2string(addrNetExtend))
			this.LogicalNode.Delete(utils.Bytes2string(addrNetExtend))
			if !ok {
				continue
			}
			vnode := vnodeItr.(Vnodeinfo)
			nidMapItr, ok := this.LogicalNodeIndexNID.Load(utils.Bytes2string(vnode.Nid))
			if ok {
				nidMap := nidMapItr.(*sync.Map)
				nidMap.Delete(utils.Bytes2string(vnode.Vid))
			}
			isZore := true
			this.LogicalNodeIndexNID.Range(func(k, v interface{}) bool {
				isZore = false
				return false
			})
			if isZore {
				continue
			}
			this.LogicalNodeIndexNID.Delete(utils.Bytes2string(vnode.Nid))
		}
		// utils.Log.Info().Msgf("虚拟节点 添加逻辑节点:%s %s 有变化:%t", this.Vnode.Vid.B58String(), vnode.Vid.B58String(), true)
		return true
	} else {
		// utils.Log.Info().Msgf("虚拟节点 添加逻辑节点:%s %s 有变化:%t", this.Vnode.Vid.B58String(), vnode.Vid.B58String(), false)
		return false
	}
}

/*
获得自己的vnodeinfo
*/
func (this *Vnode) GetSelfVnodeinfo() Vnodeinfo {
	return this.Vnode
}

/*
获得自己节点的所有逻辑节点，不包括自己节点
*/
func (this *Vnode) GetVnodeinfoAllNotSelf() []Vnodeinfo {
	// this.lock.RLock()
	// defer this.lock.RUnlock()
	vns := make([]Vnodeinfo, 0)
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		vns = append(vns, vnodeinfo)
		return true
	})
	return vns

}

/*
查找Vnodeinfo
*/
func (this *Vnode) FindVnodeinfo(vid AddressNetExtend) *Vnodeinfo {
	value, ok := this.LogicalNode.Load(utils.Bytes2string(vid))
	if ok {
		vnodeinfo := value.(Vnodeinfo)
		return &vnodeinfo
	}
	// utils.Log.Info().Msgf("未找到:%s %s", vid.B58String(), this.Vnode.Vid.B58String())
	// this.LogicalNode.Range(func(k, v interface{}) bool {
	// 	keyStr := k.(string)
	// 	addrExt := AddressNetExtend([]byte(keyStr))
	// 	// vnodeinfo := v.(Vnodeinfo)
	// 	utils.Log.Info().Msgf("打印逻辑节点:%s", addrExt.B58String())
	// 	return true
	// })
	return nil
}

/*
 * 检查是否包含指定真实的节点信息
 * 	@param	nid		真实节点地址
 *	@return bool	是否包含标识
 */
func (this *Vnode) CheckVnodeinfoExistByNid(nid nodeStore.AddressNet) bool {
	if _, exist := this.LogicalNodeIndexNID.Load(utils.Bytes2string(nid)); exist {
		return true
	}

	bExist := false
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeInfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		if bytes.Equal(vnodeInfo.Nid, nid) {
			bExist = true
			return false
		}
		return true
	})

	return bExist
}

/*
当一个节点下线，删除一个节点
查找一个节点是否包含在本虚拟节点的逻辑节点中
@return    bool    是否存在
*/
func (this *Vnode) DeleteNid(nid nodeStore.AddressNet) bool {
	bChange := false
	nidMapItr, ok := this.LogicalNodeIndexNID.Load(utils.Bytes2string(nid))
	if ok {
		nidMap := nidMapItr.(*sync.Map)
		nidMap.Range(func(k, v interface{}) bool {
			vnodeinfo, ok := v.(*Vnodeinfo)
			if !ok {
				return false
			}
			this.LogicalNode.Delete(utils.Bytes2string(vnodeinfo.Vid))
			return true
		})
		this.LogicalNodeIndexNID.Delete(utils.Bytes2string(nid))
		bChange = true
	}

	// 删除nid对应的所有vid信息
	// 因为存在删除了LogicalNodeIndexNID中的nid，却没有删除LogicalNode中数据的情况，所以需要再次删除LogicalNode中的数据
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeInfo, ok := v.(Vnodeinfo)
		if !ok {
			return false
		}
		if bytes.Equal(vnodeInfo.Nid, nid) {
			this.LogicalNode.Delete(utils.Bytes2string(vnodeInfo.Vid))
			bChange = true
		}
		return true
	})

	if bChange {
		select {
		case this.findVnodeTag <- false:
		default:
		}
	}

	return ok
}

/*
 * 当一个虚拟节点下线，删除对应记录信息
 * 查找一个节点是否包含在本虚拟节点的逻辑节点中
 * @param	nid		虚拟节点的真实id
 * @param	vid		虚拟节点的虚拟id
 */
func (this *Vnode) DeleteVid(vnodeInfo *Vnodeinfo) bool {
	if vnodeInfo == nil || vnodeInfo.Nid == nil || len(vnodeInfo.Nid) == 0 || vnodeInfo.Vid == nil || len(vnodeInfo.Vid) == 0 {
		return false
	}

	// 如果删除的虚拟节点是节点的真实节点，则走删除真实节点处理流程
	// if bytes.Equal(vnodeInfo.Nid, vnodeInfo.Vid) {
	// 	return this.DeleteNid(vnodeInfo.Nid)
	// }

	bChange := false
	nidMapItr, ok := this.LogicalNodeIndexNID.Load(utils.Bytes2string(vnodeInfo.Nid))
	if ok {
		nidMap := nidMapItr.(*sync.Map)
		bHaveOtherValue := false // 记录map中是否还有其它节点的信息
		nidMap.Range(func(k, v interface{}) bool {
			vnodeinfo, ok := v.(*Vnodeinfo)
			if !ok {
				return false
			}
			if bytes.Equal(vnodeinfo.Vid, vnodeInfo.Vid) {
				this.LogicalNode.Delete(utils.Bytes2string(vnodeinfo.Vid))
				nidMap.Delete(utils.Bytes2string(vnodeinfo.Vid))
			} else {
				bHaveOtherValue = true
			}
			// utils.Log.Info().Msgf("qlw---real delete vnode: curNode:%s, vid:%s, nid:%s", this.Vnode.Vid.B58String(), vnodeinfo.Vid.B58String(), vnodeinfo.Nid.B58String())
			return true
		})
		if !bHaveOtherValue {
			this.LogicalNodeIndexNID.Delete(utils.Bytes2string(vnodeInfo.Nid))
		}
		bChange = true
	}

	// 因为存在删除了LogicalNodeIndexNID中的nid，却没有删除LogicalNode中数据的情况，所以需要再次删除LogicalNode中的数据
	_, exist := this.LogicalNode.LoadAndDelete(utils.Bytes2string(vnodeInfo.Vid))
	if exist {
		bChange = true
	}

	if bChange {
		select {
		case this.findVnodeTag <- false:
			// utils.Log.Info().Msgf("有虚拟节点变动")
		default:
		}
	}

	return ok
}

/*
查找虚拟节点
*/
type FindVnodeVO struct {
	Self   Vnodeinfo //自己节点
	Target Vnodeinfo //目标节点
}

func (this *FindVnodeVO) Proto() ([]byte, error) {
	self := go_protobuf.Vnodeinfo{
		Nid:   this.Self.Nid,
		Index: this.Self.Index,
		Vid:   this.Self.Vid,
	}

	target := go_protobuf.Vnodeinfo{
		Nid:   this.Target.Nid,
		Index: this.Target.Index,
		Vid:   this.Target.Vid,
	}

	fvp := go_protobuf.FindVnodeVO{
		Self:   &self,
		Target: &target,
	}
	return fvp.Marshal()
}

func ParseFindVnodeVO(bs []byte) (*FindVnodeVO, error) {
	fvp := new(go_protobuf.FindVnodeVO)
	err := proto.Unmarshal(bs, fvp)
	if err != nil {
		return nil, err
	}
	self := Vnodeinfo{
		Nid:   fvp.Self.Nid,
		Index: fvp.Self.Index,
		Vid:   fvp.Self.Vid,
	}

	target := Vnodeinfo{
		Nid:   fvp.Target.Nid,
		Index: fvp.Target.Index,
		Vid:   fvp.Target.Vid,
	}

	fvVO := FindVnodeVO{
		Self:   self,
		Target: target,
	}
	return &fvVO, nil
}

// 解析vnodeinfo
func ParseVnodeinfo(bs []byte) (*Vnodeinfo, error) {
	vn := new(go_protobuf.Vnodeinfo)
	err := proto.Unmarshal(bs, vn)
	if err != nil {
		return nil, err
	}

	vnode := &Vnodeinfo{
		Nid:   vn.Nid,
		Vid:   vn.Vid,
		Index: vn.Index,
	}

	return vnode, nil
}

/*
 * 虚拟节点地址按10进制从大到小排序
 */
func (this *Vnode) GetSortAddressNetExtend(ss []VnodeinfoS, self AddressNetExtend) ([]VnodeinfoS, int) {
	var selfAddrBI *big.Int
	selfAddrBI = new(big.Int).SetBytes(self)
	var sortedS []VnodeinfoS
	sortLargeNum := 0
	sortM := make(map[string]VnodeinfoS)
	onebyone := new(nodeStore.IdDESC)

	for _, one := range ss {
		sortM[one.Vid.B58String()] = one
		*onebyone = append(*onebyone, new(big.Int).SetBytes(one.Vid))
	}

	sort.Sort(onebyone)

	for i := 0; i < len(*onebyone); i++ {
		// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
		one := (*onebyone)[i]
		if one.Cmp(selfAddrBI) == 1 {
			sortLargeNum += 1
		}

		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		sortedS = append(sortedS, sortM[nodeStore.AddressNet(*IdBsP).B58String()])
	}

	return sortedS, sortLargeNum
}

/*
 * 加锁，删除本Vnode中的某个虚拟地址所有信息
 */
func (this *Vnode) VnodeDelAddr(delV AddressNetExtend) {
	this.Lock.Lock()
	defer this.Lock.Unlock()

	//utils.Log.Info().Msgf("删除vnode %s self %s index %d", delV.B58String(), this.Vnode.Vid.B58String(), this.Vnode.Index)
	this.DownVnodeInfo.Delete(utils.Bytes2string(delV))
	this.UpVnodeInfo.Delete(utils.Bytes2string(delV))
	this.LogicalNode.Delete(utils.Bytes2string(delV))
	this.LogicalNodeIndexNID.Delete(utils.Bytes2string(delV))
	this.clientVnode.Delete(utils.Bytes2string(delV))
}

/*
 * 锁内使用
 * 不锁Vnode获取vnode的upVnode排序信息
 */
func (this *Vnode) GetUpVnodeInfo() []VnodeinfoS {
	var rV []VnodeinfoS
	this.UpVnodeInfo.Range(func(k, v interface{}) bool {
		vnodeinfoS, ok := v.(VnodeinfoS)
		if !ok {
			return false
		}
		rV = append(rV, vnodeinfoS)
		return true
	})
	return rV
}

/*
 * 锁内使用
 * 不锁Vnode获取vnode的downVnode排序信息
 */
func (this *Vnode) GetDownVnodeInfo() []VnodeinfoS {
	var rV []VnodeinfoS
	this.DownVnodeInfo.Range(func(k, v interface{}) bool {
		vnodeinfoS, ok := v.(VnodeinfoS)
		if !ok {
			return false
		}
		rV = append(rV, vnodeinfoS)
		return true
	})
	return rV
}

/*
 * 加读锁，获取本Vnode存的up down所有Vnodeinfo以及本vnode
 */
func (this *Vnode) GetOnebyoneVnodeInfo() []VnodeinfoS {
	this.Lock.RLock()
	defer this.Lock.RUnlock()

	var rV []VnodeinfoS
	this.DownVnodeInfo.Range(func(k, v interface{}) bool {
		vnodeinfoS, ok := v.(VnodeinfoS)
		if !ok {
			return false
		}
		rV = append(rV, vnodeinfoS)
		return true
	})
	this.UpVnodeInfo.Range(func(k, v interface{}) bool {
		vn, ok := v.(VnodeinfoS)
		if !ok {
			return false
		}
		rV = append(rV, vn)
		return true
	})
	rV = append(rV, VnodeinfoS{
		Nid:   this.Vnode.Nid,
		Vid:   this.Vnode.Vid,
		Index: this.Vnode.Index,
	})
	return rV
}

/*
 * 通过地址搜索返回本Vnode中合适的地址
 */
func (this *Vnode) SearchAVnodeByOnebyone(an *[]VnodeinfoS, rr nodeStore.AddressNet) (*nodeStore.AddressNet, *nodeStore.AddressNet, bool) {
	sortedvnode, times := this.GetSortAddressNetExtend(*an, AddressNetExtend(rr))
	//onebyone正常模式

	if len(sortedvnode) == 0 {
		return nil, nil, true
	}
	if times == 0 {
		//说明onebyone所有节点都比目标小
		recvV := nodeStore.AddressNet(sortedvnode[0].Vid)
		recvN := sortedvnode[0].Nid
		if bytes.Equal(recvV, this.Vnode.Vid) {
			return &recvN, &recvV, false
		}
		return &recvN, &recvV, true
	}
	if times == len(sortedvnode) {
		//说明onebyone所有节点都比目标大
		recvV := nodeStore.AddressNet(sortedvnode[len(sortedvnode)-1].Vid)
		recvN := sortedvnode[len(sortedvnode)-1].Nid
		if bytes.Equal(recvV, this.Vnode.Vid) {
			return &recvN, &recvV, false
		}
		return &recvN, &recvV, true
	}
	//到此处说明目标RecvVnode在本Vnode的up和down之间
	targetV := nodeStore.AddressNet(sortedvnode[times-1].Vid)
	targetN := sortedvnode[times-1].Nid
	if bytes.Equal(targetV, this.Vnode.Vid) {
		//发现是本vnode的虚拟地址了，由自己处理消息的返回
		return &targetN, &targetV, false
	}
	return &targetN, &targetV, true
}
