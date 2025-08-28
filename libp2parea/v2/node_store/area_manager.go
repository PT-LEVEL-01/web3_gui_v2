package nodeStore

import (
	"bytes"
	"github.com/rs/zerolog"
	"math/big"
	"sync"
	"web3_gui/utils"
)

/*
保存自己域的逻辑域名称，以及逻辑域下的节点地址
保持每个域内有两个节点地址
*/
type AreaManager struct {
	addrSelfArea  AddressArea                      //自己的域名称
	addrSelfNet   *AddressNet                      //自己的节点名称
	lock          *sync.RWMutex                    //锁
	logicAreaName *LogicAreaManager                //逻辑域名称
	areaNodeAddr  map[string]*OtherAreaAddrManager //每个域下的逻辑节点地址，保持每个域2个节点连接。key:string=域名称;value:*OtherAreaAddrManager=逻辑节点;
	log           *zerolog.Logger                  //日志
}

func NewAreaManager(areaName AddressArea, addrSelf *AddressNet, log *zerolog.Logger) *AreaManager {
	//(*log).Info().Msgf("日志指针:%p", *log)
	am := AreaManager{
		addrSelfArea:  areaName,
		addrSelfNet:   addrSelf,
		lock:          &sync.RWMutex{},
		logicAreaName: NewLogicAreaManager(areaName, log),
		areaNodeAddr:  make(map[string]*OtherAreaAddrManager),
		log:           log,
	}
	return &am
}

/*
检查一个域地址是否是需要的
@areaName    AddressArea    域名称
@addr        AddressNet     地址
@save        bool           是否保存
@return      bool           是否需要此地址
@return      [][]byte       要删除的地址
@return      bool           地址是否已经存在
*/
func (this *AreaManager) CheckOtherAreaAddr(nodeInfo *NodeInfo, save bool) (bool, []NodeInfo, bool) {
	if save {
		this.lock.Lock()
		defer this.lock.Unlock()
	} else {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}
	areaName := nodeInfo.AreaName
	//和自己的域相同，则不添加
	need, removeAreaNames, exist := this.logicAreaName.CheckNeedAddr(areaName, save)
	//if save {
	//	this.log.Info().Hex("要添加的域", areaName).Str("要添加的节点", nodeInfo.IdInfo.Id.B58String()).Bool("是否存在",
	//		exist).Bool("是否需要", need).Interface("要删除的节点", removeAreaNames).Send()
	//}
	//已存在的域，判断是否替换节点
	if exist {
		logicAddr, _ := this.areaNodeAddr[utils.Bytes2string(areaName)]
		need, removeNodeInfo, exist := logicAddr.CheckAreaAddr(*nodeInfo, save)
		//this.log.Info().Msgf("日志指针:%p", *this.Log)
		//this.log.Info().Bool("是否存在", exist).Bool("是否需要", need).Interface("要删除的节点", removeNodeInfo).Send()
		removeNodeInfos := make([]NodeInfo, 0, 1)
		if removeNodeInfo != nil {
			removeNodeInfos = append(removeNodeInfos, *removeNodeInfo)
		}
		return need, removeNodeInfos, exist
	}
	//需要这个域
	if need {
		oaam, ok := this.areaNodeAddr[utils.Bytes2string(areaName)]
		if !ok {
			oaam = NewOtherAreaAddrManager(nodeInfo.AreaName, nodeInfo.IdInfo.Id)
			this.areaNodeAddr[utils.Bytes2string(areaName)] = oaam
		}
		need, removeNodeInfo, exist := oaam.CheckAreaAddr(*nodeInfo, save)
		//
		removeNodeInfos := make([]NodeInfo, 0, 1)
		if removeNodeInfo != nil {
			removeNodeInfos = append(removeNodeInfos, *removeNodeInfo)
		}
		//要删除的其他域的节点
		for _, areaNameOne := range removeAreaNames {
			//if save {
			//	this.log.Info().Hex("CheckOtherAreaAddr要删除的域名称", areaNameOne).Send()
			//}
			oaam, ok := this.areaNodeAddr[utils.Bytes2string(areaNameOne)]
			if !ok {
				continue
			}
			for _, nodeOne := range oaam.GetNodeInfos() {
				removeNodeInfos = append(removeNodeInfos, nodeOne)
			}
			delete(this.areaNodeAddr, utils.Bytes2string(areaNameOne))
		}
		return need, removeNodeInfos, exist
	}
	return false, nil, false
}

/*
获取其他域节点地址
@return      []AddressNet       要删除的地址
*/
func (this *AreaManager) GetNodeInfos() []NodeInfo {
	addrs := make([]NodeInfo, 0)
	this.lock.RLock()
	defer this.lock.RUnlock()
	for _, oaamOne := range this.areaNodeAddr {
		addrs = append(addrs, oaamOne.GetNodeInfos()...)
	}
	return addrs
}

/*
对比节点是否有变化
@return    bool    是否有变化。true=有变化;false=没变化;
*/
func (this *AreaManager) EqualLogicNodes(nodesOld []NodeInfo) bool {
	newAddrs := this.GetNodeInfos()
	if len(nodesOld) != len(newAddrs) {
		return true
	}
	//放入map，去重后方便查找
	isChange := false
	oldMap := make(map[string]bool)
	for _, one := range nodesOld {
		oldMap[utils.Bytes2string(one.IdInfo.Id.Data())] = false
	}
	//查找对比
	for _, one := range newAddrs {
		_, ok := oldMap[utils.Bytes2string(one.IdInfo.Id.Data())]
		if !ok {
			isChange = true
			break
		}
	}
	return isChange
}

/*
查询一个域的临近域下的节点信息
@areaName    []byte         域名称
*/
func (this *AreaManager) FindNearAreaAddr(areaName AddressArea) []NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	//先查询指定的域在不在
	oaam, ok := this.areaNodeAddr[utils.Bytes2string(areaName)]
	if ok {
		return oaam.GetNodeInfos()
	}
	//指定的域不在，再查询目标的逻辑域
	areaNames := this.logicAreaName.GetAddrs()
	bucket := NewBucket(len(areaNames))
	for _, areaName := range areaNames {
		bucket.Add(new(big.Int).SetBytes(areaName))
	}
	names := bucket.Get(new(big.Int).SetBytes(areaName))
	//将逻辑域下的所有节点地址找出来
	addrs := make([]NodeInfo, 0)
	for _, areaName := range names {
		oam, ok := this.areaNodeAddr[utils.Bytes2string(areaName.Bytes())]
		if ok {
			addrs = append(addrs, oam.GetNodeInfos()...)
		}
	}
	return addrs
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *AreaManager) CleanNodeInfo() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range this.areaNodeAddr {
		one.CleanNodeInfo()
	}
}

/*
从新设置日志库
*/
func (this *AreaManager) SetLog(log *zerolog.Logger) {
	this.logicAreaName.SetLog(log)
	this.lock.Lock()
	defer this.lock.Unlock()
	this.log = log
}

/*
其他域内节点地址保存
保持一个域内有2个地址，不多不少
*/
type OtherAreaAddrManager struct {
	addrAreaSelf AddressArea          //域名称
	addrNetSelf  *AddressNet          //节点地址
	lock         *sync.RWMutex        //
	nodeMap      map[string]*NodeInfo //节点地址。key:string=域名称+AddressNet;value:*AddressNet=;
	log          *zerolog.Logger      //
}

func NewOtherAreaAddrManager(addrAreaSelf AddressArea, addrNetSelf *AddressNet) *OtherAreaAddrManager {
	am := OtherAreaAddrManager{
		addrAreaSelf: addrAreaSelf,
		addrNetSelf:  addrNetSelf,
		lock:         new(sync.RWMutex),
		nodeMap:      make(map[string]*NodeInfo),
		log:          utils.Log,
	}
	return &am
}

/*
检查一个节点地址是否是需要的
@addrSelf    AddressNet     自己节点地址
@addr        AddressNet     要检查的节点地址
@save        bool           是否保存
@return      bool           是否需要此地址
@return      *NodeInfo      要删除的节点
@return      bool           地址是否已经存在
*/
func (this *OtherAreaAddrManager) CheckAreaAddr(nodeInfo NodeInfo, save bool) (bool, *NodeInfo, bool) {
	if save {
		this.lock.Lock()
		defer this.lock.Unlock()
	} else {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}
	key := utils.Bytes2string(append(nodeInfo.AreaName, nodeInfo.IdInfo.Id.Data()...))
	//检查地址是否存在
	nodeInfoExist, ok := this.nodeMap[key]
	if ok {
		//相同节点多连接
		have := false
		for midStr, _ := range nodeInfoExist.Sessions {
			if bytes.Equal([]byte(midStr), nodeInfo.MachineID) {
				have = true
				break
			}
		}
		if !have {
			//保存多个连接
			for _, sessionOne := range nodeInfo.GetSessions() {
				nodeInfoExist.AddSession(sessionOne)
			}
			//nodeInfoExist.Sessions[utils.Bytes2string(nodeInfo.MachineID)] = nodeInfo.Sessions[utils.Bytes2string(nodeInfo.MachineID)]
		}
		return false, nil, true
	}
	bucket := NewBucket(3)
	for _, one := range this.nodeMap {
		bucket.Add(new(big.Int).SetBytes(one.IdInfo.Id.Data()))
	}
	bucket.Add(new(big.Int).SetBytes(nodeInfo.IdInfo.Id.Data()))
	ids := bucket.Get(new(big.Int).SetBytes(this.addrNetSelf.Data()))
	var removeIdData []byte
	if len(ids) > 2 {
		removeIdData = ids[2].Bytes()
	}
	//判断删除的地址中，是否有新添加的地址
	if bytes.Equal(removeIdData, nodeInfo.IdInfo.Id.Data()) {
		return false, nil, false
	}
	removeKey := utils.Bytes2string(append(nodeInfo.AreaName, removeIdData...))
	removeNodeInfo := this.nodeMap[removeKey]
	//是否保存
	if save {
		this.nodeMap[key] = &nodeInfo
		delete(this.nodeMap, removeKey)
	}
	return true, removeNodeInfo, false
}

/*
获取所有地址
@return      []AddressNet       获取的所有地址
*/
func (this *OtherAreaAddrManager) GetNodeInfos() []NodeInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrs := make([]NodeInfo, 0, len(this.nodeMap))
	for _, one := range this.nodeMap {
		addrs = append(addrs, *one)
	}
	return addrs
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *OtherAreaAddrManager) CleanNodeInfo() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range this.nodeMap {
		if len(one.GetSessions()) == 0 {
			delete(this.nodeMap, utils.Bytes2string(one.IdInfo.Id.Data()))
		}
	}
}

/*
域名称和节点地址的绑定
*/
//type AddressAreaNetBind struct {
//	AreaName AddressArea //域名称
//	Addr     AddressNet  //节点地址
//}
