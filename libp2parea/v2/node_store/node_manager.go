package nodeStore

import (
	"bytes"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"sync"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

type NodeManager struct {
	NodeSelf         *NodeInfo         //自己节点信息
	engine           *engine.Engine    //
	AreaLogicAddrWAN *AreaManager      //广域网保存域节点地址
	AreaLogicAddrLAN *AreaManager      //局域网保存域节点地址
	NodeLogicAddrWAN *LogicAddrManager //广域网保存逻辑节点地址
	NodeLogicAddrLAN *LogicAddrManager //局域网逻辑节点地址
	WhiteList        *WhiteListManager //白名单
	FreeNodeAddr     *ProxyAddrManager //游离的节点
	ProxyNodeAddr    *ProxyAddrManager //被代理的节点
	lock             *sync.RWMutex     //
	NodeIdLevel      uint              //= 256  //节点id长度
	log              *zerolog.Logger   //日志
}

/*
从新设置日志库
*/
func (this *NodeManager) SetLog(log *zerolog.Logger) {
	this.AreaLogicAddrWAN.SetLog(log)
	this.AreaLogicAddrLAN.SetLog(log)
	this.NodeLogicAddrWAN.SetLog(log)
	this.NodeLogicAddrLAN.SetLog(log)
	this.lock.Lock()
	defer this.lock.Unlock()
	this.log = log
}

func NewNodeManager(areaName [32]byte, addrPre string, key keystore.KeystoreInterface, pwd string, log *zerolog.Logger) (*NodeManager, utils.ERROR) {
	// config.Wallet_keystore_default_pwd = pwd
	//加载本地网络id
	idinfo, ERR := BuildIdinfo(addrPre, key, pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	// 如果生成的DH公钥为空, 则给出对应的提示信息
	if idinfo == nil || idinfo.CPuk == [32]byte{} {
		utils.Log.Error().Msgf("DH公钥为空!!!! keystore生成的公钥为空, 请确认keystore的用法是否正确!!!!")
	}
	nodeSelf := NewNodeInfo(areaName[:], idinfo, ulid.Make().Bytes())
	nodeManager := NodeManager{
		NodeSelf:         nodeSelf,                                     //
		AreaLogicAddrWAN: NewAreaManager(areaName[:], idinfo.Id, log),  //保存逻辑域节点地址
		AreaLogicAddrLAN: NewAreaManager(areaName[:], idinfo.Id, log),  //保存逻辑域节点地址
		NodeLogicAddrWAN: NewLogicAddrManager(nodeSelf.IdInfo.Id, log), //保存逻辑节点地址
		NodeLogicAddrLAN: NewLogicAddrManager(nodeSelf.IdInfo.Id, log), //保存逻辑节点地址
		WhiteList:        NewWhiteListManager(),                        //白名单
		FreeNodeAddr:     NewProxyAddrManager(),                        //
		ProxyNodeAddr:    NewProxyAddrManager(),                        //
		lock:             new(sync.RWMutex),                            //
		NodeIdLevel:      256,                                          //节点id长度
		log:              log,                                          //
	}
	//nodeManager.AreaLogicAddrWAN.Log = nodeManager.Log
	//nodeManager.AreaLogicAddrLAN.Log = nodeManager.Log
	//nodeManager.NodeLogicAddrWAN.Log = nodeManager.Log
	//nodeManager.NodeLogicAddrLAN.Log = nodeManager.Log
	return &nodeManager, utils.NewErrorSuccess()

}

/*
加载本地私钥生成idinfo
*/
func BuildIdinfo(addrPre string, key keystore.KeystoreInterface, pwd string) (*IdInfo, utils.ERROR) {
	pukBs, prkBs, ERR := key.GetNetAddrKeyPair(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	prk := ed25519.PrivateKey(prkBs)
	puk := ed25519.PublicKey(pukBs)
	dhKeyPair, ERR := key.GetDhAddrKeyPair(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	addr, ERR := BuildAddr(addrPre, puk)
	if ERR.CheckFail() {
		return nil, ERR
	}
	idinfo := IdInfo{
		Id:   addr,                     //id，网络地址，临时地址，由updatetime字符串产生MD5值
		EPuk: puk,                      //ed25519公钥
		CPuk: dhKeyPair.GetPublicKey(), //curve25519公钥
		V:    uint32(0),                //dh密钥版本
		// Sign: sign,                      //ed25519私钥签名
	}
	ERR = idinfo.SignDHPuk(prk)
	return &idinfo, ERR
}

// 添加一个代理节点
//func (this *NodeManager) AddProxyNode(node Node) {
//	this.ProxyNodeAddr.AddAddr(&node.IdInfo.Id)
//}

// 得到一个代理节点
//func (this *NodeManager) GetProxyNode(id string) (node *Node, ok bool) {
//	this.nodesLock.RLock()
//	node, ok = this.nodes[id]
//	if node != nil && node.Type != Node_type_proxy {
//		node = nil
//		ok = false
//	}
//	this.nodesLock.RUnlock()
//	return
//}

/*
检查是否本节点的逻辑域
检查这个域名称是否自己的逻辑域，如果是，则保存
不保存自己
@return     bool           是否需要此地址
@return     [][]byte       要删除的地址
@return     bool           地址是否已经存在
*/
func (this *NodeManager) CheckNeedAreaWan(nodeInfo *NodeInfo, save bool) (isNeed bool, removeNodes []NodeInfo, exist bool) {
	return this.AreaLogicAddrWAN.CheckOtherAreaAddr(nodeInfo, save)
}

/*
检查是否本节点的逻辑域
检查这个域名称是否自己的逻辑域，如果是，则保存
不保存自己
@return     bool           是否需要此地址
@return     [][]byte       要删除的地址
@return     bool           地址是否已经存在
*/
func (this *NodeManager) CheckNeedAreaLan(nodeInfo *NodeInfo, save bool) (isNeed bool, removeNodes []NodeInfo, exist bool) {
	return this.AreaLogicAddrLAN.CheckOtherAreaAddr(nodeInfo, save)
}

/*
检查节点是否是本节点的逻辑节点
添加一个超级节点
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
@return     bool           是否需要此地址
@return     [][]byte       要删除的地址
@return     bool           地址是否已经存在
*/
func (this *NodeManager) CheckNeedNodeWan(nodeInfo *NodeInfo, save bool) (isNeed bool, removeNodes []*NodeInfo, exist bool) {
	return this.NodeLogicAddrWAN.CheckNeedAddr(nodeInfo, save)
}

/*
检查节点是否是本节点的逻辑节点
添加一个局域网节点
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
@return     bool           是否需要此地址
@return     [][]byte       要删除的地址
@return     bool           地址是否已经存在
*/
func (this *NodeManager) CheckNeedNodeLan(nodeInfo *NodeInfo, save bool) (isNeed bool, removeNodes []*NodeInfo, exist bool) {
	return this.NodeLogicAddrLAN.CheckNeedAddr(nodeInfo, save)
}

/*
添加一个超级节点地址
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
*/
func (this *NodeManager) AddNodeInfoWAN(nodeInfo *NodeInfo) bool {
	//不添加自己
	if nodeInfo != nil && bytes.Equal(nodeInfo.IdInfo.Id.Data(), this.NodeSelf.IdInfo.Id.Data()) {
		// utils.Log.Info().Msgf("%s CheckNeedNode 111111 need:%t", this.NodeSelf.IdInfo.Id.B58String(), nodeId.B58String(), false)
		return false
	}
	//是否已经存在
	if this.NodeLogicAddrWAN.ExistAddrs(nodeInfo) || this.WhiteList.CheckWhiteListAddrExist(nodeInfo.IdInfo.Id) {
		return false
	}
	ok, _, _ := this.NodeLogicAddrWAN.CheckNeedAddr(nodeInfo, true)
	return ok
}

/*
添加内网节点地址
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
*/
func (this *NodeManager) AddNodeInfoLAN(nodeInfo *NodeInfo) bool {
	//log := *this.Log
	//this.Log.Info().Str("添加逻辑节点地址", addr.B58String()).Send()
	//不添加自己
	if nodeInfo != nil && bytes.Equal(nodeInfo.IdInfo.Id.Data(), this.NodeSelf.IdInfo.Id.Data()) {
		//this.Log.Info().Str("添加逻辑节点地址", addr.B58String()).Send()
		return false
	}
	//是否已经存在
	if this.NodeLogicAddrLAN.ExistAddrs(nodeInfo) || this.WhiteList.CheckWhiteListAddrExist(nodeInfo.IdInfo.Id) {
		//this.Log.Info().Str("添加逻辑节点地址", addr.B58String()).Send()
		return false
	}
	ok, _, exist := this.NodeLogicAddrLAN.CheckNeedAddr(nodeInfo, true)
	//this.Log.Info().Str("添加逻辑节点地址", addr.B58String()).Bool("是否需要", ok).Bool("是否已经存在", exist).Send()
	if exist {
		return false
	}
	if ok {
		//(*this.Log).Info().Str("添加节点到 LAN集合中", nodeInfo.IdInfo.Id.B58String()).Send()
		////log.Info().Str("添加逻辑节点地址", nodeInfo.IdInfo.Id.B58String()).Send()
		//for _, one := range removeNodes {
		//	(*this.Log).Info().Str("要删除的节点 LAN集合中", one.IdInfo.Id.B58String()).Send()
		//	//log.Info().Str("要删除的节点", one.IdInfo.Id.B58String()).Send()
		//}
	}
	return ok
}

/*
添加节点到游离的节点
*/
func (this *NodeManager) AddNodeInfoFree(nodeInfo *NodeInfo) bool {
	//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Send()
	//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Msgf("指针%p", nodeInfo)
	//不添加自己
	if nodeInfo != nil && bytes.Equal(nodeInfo.IdInfo.Id.Data(), this.NodeSelf.IdInfo.Id.Data()) {
		//(*this.Log).Info().Str("添加节点到 自己的节点不保存", nodeInfo.IdInfo.Id.B58String()).Send()
		// utils.Log.Info().Msgf("%s CheckNeedNode 111111 need:%t", this.NodeSelf.IdInfo.Id.B58String(), nodeId.B58String(), false)
		return false
	}
	//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Send()
	//是否已经存在
	if this.WhiteList.CheckWhiteListAddrExist(nodeInfo.IdInfo.Id) {
		//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Send()
		return false
	}
	//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Send()
	this.FreeNodeAddr.Add(nodeInfo)
	//(*this.Log).Info().Str("添加节点到 游离集合中", nodeInfo.IdInfo.Id.B58String()).Int("session数量", len(nodeInfo.GetSessions())).Send()
	return true
	//ok, _, _ := this.NodeLogicAddrWAN.CheckNeedAddr(nodeInfo, true)
	//return ok
}

/*
添加节点到代理的节点
*/
func (this *NodeManager) AddNodeInfoProxy(nodeInfo *NodeInfo) bool {
	//不添加自己
	if nodeInfo != nil && bytes.Equal(nodeInfo.IdInfo.Id.Data(), this.NodeSelf.IdInfo.Id.Data()) {
		// utils.Log.Info().Msgf("%s CheckNeedNode 111111 need:%t", this.NodeSelf.IdInfo.Id.B58String(), nodeId.B58String(), false)
		return false
	}
	//(*this.Log).Info().Str("添加节点到 代理集合中", nodeInfo.IdInfo.Id.B58String()).Send()
	//是否已经存在
	if this.WhiteList.CheckWhiteListAddrExist(nodeInfo.IdInfo.Id) {
		return false
	}
	this.ProxyNodeAddr.Add(nodeInfo)
	return true
	//ok, _, _ := this.NodeLogicAddrWAN.CheckNeedAddr(nodeInfo, true)
	//return ok
}

/*
删除一个节点，包括超级节点和代理节点
*/
//func (this *NodeManager) DelNode(id *AddressNet) {
//	idStr := utils.Bytes2string(*id) //  id.B58String()
//	// utils.Log.Info().Msgf("delete client nodeid %s", id.B58String())
//	this.nodesLock.Lock()
//	delete(this.nodes, idStr)
//	this.nodesLock.Unlock()
//}

/*
通过id查找一个节点
*/
//func (this *NodeManager) FindNode(id *AddressNet) *Node {
//	idStr := utils.Bytes2string(*id)
//	var node *Node
//	this.nodesLock.RLock()
//	node, _ = this.nodes[idStr]
//	this.nodesLock.RUnlock()
//	return node
//}
//
//func (this *NodeManager) FindNodeInLogic(id *AddressNet) *Node {
//	idStr := utils.Bytes2string(*id)
//
//	var node *Node
//	// var ok bool
//	this.nodesLock.RLock()
//	node, _ = this.nodes[idStr]
//	this.nodesLock.RUnlock()
//	if node != nil && node.Type != Node_type_logic && node.Type != Node_type_white_list {
//		return nil
//	}
//	return node
//}

/*
对比逻辑节点是否有变化
@return    bool    是否有变化。true=有变化;false=没变化;
*/
func (this *NodeManager) EqualLogicNodesWAN(nodesOld []NodeInfo) bool {
	newLogicNodes := this.NodeLogicAddrWAN.GetNodeInfos()
	if len(nodesOld) != len(newLogicNodes) {
		return true
	}
	nowNodesMap := make(map[string]bool)
	for _, one := range newLogicNodes {
		nowNodesMap[utils.Bytes2string(one.IdInfo.Id.Data())] = true
	}
	isChange := false
	for _, one := range nodesOld {
		_, ok := nowNodesMap[utils.Bytes2string(one.IdInfo.Id.Data())]
		if !ok {
			isChange = true
			break
		}
	}
	return isChange
}

/*
对比局域网中的逻辑节点是否有变化
*/
func (this *NodeManager) EqualLogicNodesLAN(nodesOld []NodeInfo) bool {
	newLogicNodes := this.NodeLogicAddrLAN.GetNodeInfos()
	if len(nodesOld) != len(newLogicNodes) {
		return true
	}
	nowNodesMap := make(map[string]bool)
	for _, one := range newLogicNodes {
		nowNodesMap[utils.Bytes2string(one.IdInfo.Id.Data())] = true
	}
	isChange := false
	for _, one := range nodesOld {
		_, ok := nowNodesMap[utils.Bytes2string(one.IdInfo.Id.Data())]
		if !ok {
			isChange = true
			break
		}
	}
	return isChange
}

/*
在超级节点中找到最近的节点，不包括代理节点，带黑名单
@nodeId       *AddressNet     要查找的节点
@outId        *AddressNet     排除一个节点
@includeSelf  bool            是否包括自己
@blockAddr    []*AddressNet   已经试过，但失败了被拉黑的节点
@return       *AddressNet     查找到的节点id，可能为空
*/
func (this *NodeManager) FindNearInWAN(nodeId, outId *AddressNet, includeSelf bool) (*AddressNet, utils.ERROR) {
	nodeInfos := this.WhiteList.GetNodeInfos()
	addrs := this.NodeLogicAddrWAN.GetNodeInfos()
	kl := NewBucket(len(addrs) + len(nodeInfos) + 1)
	if includeSelf {
		// utils.Log.Info().Msgf("添加查询逻辑节点:%s", this.NodeSelf.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id.Data()))
	}
	addSome := false
	for _, one := range append(addrs, nodeInfos...) {
		if outId != nil && bytes.Equal(one.IdInfo.Id.Data(), outId.Data()) {
			continue
		}
		addSome = true
		//utils.Log.Info().Msgf("add add %s", one.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id.Data()))
	}
	if !addSome && !includeSelf {
		return nil, utils.NewErrorSuccess()
	}
	targetIds := kl.Get(new(big.Int).SetBytes(nodeId.Data()))
	if len(targetIds) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil, utils.NewErrorSuccess()
	}
	targetIdBs := targetId.Bytes()
	targetIdBsNew := utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length)
	addr, ERR := BuildAddrByData(nodeId.GetPre(), *targetIdBsNew)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//mh := AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// utils.Log.Info().Msgf("检查id结果:%s", mh.B58String())
	return &addr, utils.NewErrorSuccess()
}

/*
在局域网中找到最近的节点，不包括代理节点，带黑名单
@nodeId       *AddressNet     要查找的节点
@outId        *AddressNet     排除一个节点
@includeSelf  bool            是否包括自己
@blockAddr    []*AddressNet   已经试过，但失败了被拉黑的节点
@return       *AddressNet     查找到的节点id，可能为空
*/
func (this *NodeManager) FindNearInLAN(nodeId, outId *AddressNet, includeSelf bool) (*AddressNet, utils.ERROR) {
	addrs := this.NodeLogicAddrLAN.GetNodeInfos()
	kl := NewBucket(len(addrs) + 1)
	if includeSelf {
		//(*this.Log).Info().Msgf("添加查询逻辑节点:%s", this.NodeSelf.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id.Data()))
	}
	addSome := false
	for _, one := range addrs {
		if outId != nil && bytes.Equal(one.IdInfo.Id.Data(), outId.Data()) {
			continue
		}
		addSome = true
		//(*this.Log).Info().Msgf("添加查询逻辑节点:%s", one.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id.Data()))
	}
	if !addSome && !includeSelf {
		return nil, utils.NewErrorSuccess()
	}
	targetIds := kl.Get(new(big.Int).SetBytes(nodeId.Data()))
	if len(targetIds) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil, utils.NewErrorSuccess()
	}
	targetIdBs := targetId.Bytes()

	targetIdBsNew := utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length)
	addr, ERR := BuildAddrByData(nodeId.GetPre(), *targetIdBsNew)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//mh := AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// utils.Log.Info().Msgf("检查id结果:%s", mh.B58String())
	return &addr, utils.NewErrorSuccess()
}

//
//// 在节点中找到最近的节点，包括代理节点
//func (this *NodeManager) FindNearNodeId(nodeId, outId *AddressNet, includeSelf bool) *AddressNet {
//	addrs := this.NodeLogicAddrWAN.GetAddrs()
//	addrProxy := this.ProxyNodeAddr.GetAddrs()
//	kl := NewBucket(len(addrs) + len(addrProxy) + 1)
//	if includeSelf {
//		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id))
//	}
//
//	this.nodesLock.RLock()
//
//	for _, one := range this.nodes {
//		if one.Type != Node_type_logic && one.Type != Node_type_proxy && one.Type != Node_type_white_list {
//			continue
//		}
//		if outId != nil && bytes.Equal(one.IdInfo.Id, *outId) {
//			continue
//		}
//		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
//	}
//
//	this.nodesLock.RUnlock()
//
//	// Nodes.Range(func(k, v interface{}) bool {
//	// 	if k.(string) == outIdStr {
//	// 		return true
//	// 	}
//	// 	value := v.(*Node)
//	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
//	// 	return true
//	// })
//	// //代理节点
//	// Proxys.Range(func(k, v interface{}) bool {
//	// 	if k.(string) == outIdStr {
//	// 		return true
//	// 	}
//	// 	value := v.(*Node)
//	// 	//过滤APP节点
//	// 	if value.IsApp {
//	// 		return true
//	// 	}
//	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
//	// 	return true
//	// })
//
//	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
//	if len(targetIds) == 0 {
//		return nil
//	}
//	targetId := targetIds[0]
//	if targetId == nil {
//		return nil
//	}
//	targetIdBs := targetId.Bytes()
//	mh := AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
//	// mh := AddressNet(targetId.Bytes())
//	return &mh
//}

/*
	根据节点id得到一个距离最短节点的信息，不包括代理节点
	@nodeId         要查找的节点
	@includeSelf    是否包括自己
	@outId          排除一个节点
	@return         查找到的节点id，可能为空
*/
//func Get(nodeId string, includeSelf bool, outId string) *Node {
//	nodeIdInt, b := new(big.Int).SetString(nodeId, IdStrBit)
//	if !b {
//		fmt.Println("节点id格式不正确，应该为十六进制字符串:")
//		fmt.Println(nodeId)
//		return nil
//	}
//	kl := NewKademlia()
//	if includeSelf {
//		//		temp := new(big.Int).SetBytes(Root.IdInfo.Id)
//		kl.add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
//	}
//	for key, value := range Nodes {
//		if outId != "" && key == outId {
//			continue
//		}
//		kl.add(new(big.Int).SetBytes(value.IdInfo.Id))
//	}
//	// TODO 不安全访问
//	targetId := kl.get(nodeIdInt)[0]

//	if targetId == nil {
//		return nil
//	}
//	if hex.EncodeToString(targetId.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id) {
//		return NodeSelf
//	}
//	return Nodes[hex.EncodeToString(targetId.Bytes())]
//}

/*
在连接中获得白名单节点
*/
func (this *NodeManager) GetWhiltListNodes() []NodeInfo {
	return this.WhiteList.GetNodeInfos()
}

/*
得到所有逻辑节点，不包括本节点，也不包括代理节点
*/
func (this *NodeManager) GetLogicNodesWAN() []NodeInfo {
	return this.NodeLogicAddrWAN.GetNodeInfos()
}

/*
获取局域网中逻辑节点地址
*/
func (this *NodeManager) GetLogicNodesLAN() []NodeInfo {
	return this.NodeLogicAddrLAN.GetNodeInfos()
}

/*
获取游离的节点信息
*/
func (this *NodeManager) GetFreeNodes() []NodeInfo {
	return this.FreeNodeAddr.GetNodeInfos()
}

/*
获取代理的节点信息
*/
func (this *NodeManager) GetProxyNodes() []NodeInfo {
	return this.ProxyNodeAddr.GetNodeInfos()
}

/*
查询外网节点信息
*/
func (this *NodeManager) FindNodeInfoAllByAddr(addr *AddressNet) *NodeInfo {
	if nodeInfoRemote := this.FindNodeInfoLANByAddr(addr); nodeInfoRemote != nil {
		//(*this.Log).Info().Str("这里找到的", "").Send()
		return nodeInfoRemote
	}
	if nodeInfoRemote := this.FindNodeInfoWANByAddr(addr); nodeInfoRemote != nil {
		//(*this.Log).Info().Str("这里找到的", "").Send()
		return nodeInfoRemote
	}
	if nodeInfoRemote := this.FindNodeInfoWhiteByAddr(addr); nodeInfoRemote != nil {
		//(*this.Log).Info().Str("这里找到的", "").Send()
		return nodeInfoRemote
	}
	if nodeInfoRemote := this.FindNodeInfoFreeByAddr(addr); nodeInfoRemote != nil {
		//(*this.Log).Info().Str("这里找到的", "").Send()
		return nodeInfoRemote
	}
	if nodeInfoRemote := this.FindNodeInfoProxyByAddr(addr); nodeInfoRemote != nil {
		//(*this.Log).Info().Str("这里找到的", "").Send()
		return nodeInfoRemote
	}
	return nil
}

/*
查询外网节点信息
*/
func (this *NodeManager) FindNodeInfoWANByAddr(addr *AddressNet) *NodeInfo {
	return this.NodeLogicAddrWAN.FindNodeInfoByAddr(addr)
}

/*
查询局域节点信息
*/
func (this *NodeManager) FindNodeInfoLANByAddr(addr *AddressNet) *NodeInfo {
	return this.NodeLogicAddrLAN.FindNodeInfoByAddr(addr)
}

/*
查询白名单节点信息
*/
func (this *NodeManager) FindNodeInfoWhiteByAddr(addr *AddressNet) *NodeInfo {
	return this.WhiteList.FindNodeInfoByAddr(addr)
}

/*
查询游离节点信息
*/
func (this *NodeManager) FindNodeInfoFreeByAddr(addr *AddressNet) *NodeInfo {
	return this.FreeNodeAddr.FindNodeInfoByAddr(addr)
}

/*
查询代理节点信息
*/
func (this *NodeManager) FindNodeInfoProxyByAddr(addr *AddressNet) *NodeInfo {
	return this.ProxyNodeAddr.FindNodeInfoByAddr(addr)
}

/*
得到所有逻辑节点，不包括本节点，也不包括代理节点
*/
//func (this *NodeManager) GetLogicNodeInfo() []*Node {
//	ids := make([]*Node, 0)
//	this.nodesLock.RLock()
//	for _, one := range this.nodes {
//		// utils.Log.Info().Msgf("打印所有连接:%s", one.IdInfo.Id.B58String())
//		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
//			continue
//		}
//		ids = append(ids, one)
//	}
//	this.nodesLock.RUnlock()
//	return ids
//}

//func (this *NodeManager) IsLogicNode(addr AddressNet) bool {
//	var isLogic bool
//	this.nodesLock.RLock()
//	for _, one := range this.nodes {
//		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
//			continue
//		}
//		if bytes.Equal(addr, one.IdInfo.Id) {
//			isLogic = true
//			break
//		}
//	}
//	this.nodesLock.RUnlock()
//	return isLogic
//}

/*
获得所有代理节点
*/
//func (this *NodeManager) GetProxyAll() []AddressNet {
//	return this.ProxyNodeAddr.GetAddrs()
//}

/*
	获取额外的节点连接
*/
// func GetOtherNodes() []AddressNet {
// 	ids := make([]AddressNet, 0)
// 	OtherNodes.Range(func(k, v interface{}) bool {
// 		value := v.(*Node)
// 		ids = append(ids, value.IdInfo.Id)
// 		return true
// 	})
// 	return ids
// }

/*
	将节点转化为逻辑节点
*/
// func SwitchNodesClientToLogic(node Node) {
// 	node.Type = Node_type_super
// 	utils.Log.Info().Msgf("SwitchNodesClientToLogic %s", node.IdInfo.Id.B58String())
// 	nodesLock.Lock()
// 	nodeLoad, ok := nodes[utils.Bytes2string(node.IdInfo.Id)]
// 	if ok {
// 		nodeLoad.Type = Node_type_super
// 	} else {
// 		nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
// 	}
// 	nodesLock.Unlock()
// 	// NodesClient.Delete(utils.Bytes2string(node.IdInfo.Id))
// 	// Nodes.Store(utils.Bytes2string(node.IdInfo.Id), &node)
// }

/*
	检查节点是否是本节点的逻辑节点
	只检查，不保存
*/
// func (this *NodeManager) CheckNeedNode(nodeId *AddressNet) (isNeed bool) {
// 	/*
// 		1.找到已有节点中与本节点最近的节点
// 		2.计算两个节点是否在同一个网络
// 		3.若在同一个网络，计算谁的值最小
// 	*/

// 	if len(this.GetLogicNodes()) == 0 {
// 		return true
// 	}
// 	//是本身节点不添加
// 	if bytes.Equal(*nodeId, this.NodeSelf.IdInfo.Id) {
// 		return false
// 	}

// 	ids := NewKademlia(this.NodeSelf.IdInfo.Id, NodeIdLevel)
// 	for _, one := range this.GetLogicNodes() {
// 		ids.AddId(one)
// 	}
// 	ok, _ := ids.AddId(*nodeId)
// 	return ok
// }

// var (
// 	networkIDsLock   = new(sync.RWMutex)
// 	nodeNetworkIDStr = ""
// 	networkIDs       []*AddressNet
// )

// //得到每个节点网络的网络号，不包括本节点
// func getNodeNetworkNum() []*AddressNet {
// 	networkIDsLock.RLock()
// 	if nodeNetworkIDStr != "" && nodeNetworkIDStr == NodeSelf.IdInfo.Id.B58String() {
// 		networkIDsLock.RUnlock()
// 		return networkIDs
// 	}
// 	networkIDsLock.RUnlock()

// 	// rootInt, _ := new(big.Int).SetString(, IdStrBit)
// 	networkIDsLock.Lock()
// 	nodeNetworkIDStr = NodeSelf.IdInfo.Id.B58String()

// 	root := new(big.Int).SetBytes(NodeSelf.IdInfo.Id)

// 	networkIDs = make([]*AddressNet, 0)
// 	for i := 0; i < int(NodeIdLevel); i++ {
// 		//---------------------------------
// 		//将后面的i位置零
// 		//---------------------------------
// 		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
// 		//---------------------------------
// 		//第i位取反
// 		//---------------------------------
// 		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

// 		// bs, err := utils.Encode(networkNum.Bytes(), config.HashCode)
// 		// if err != nil {
// 		// 	// fmt.Println("格式化muhash错误")
// 		// 	continue
// 		// }
// 		// mhbs := utils.Multihash(bs)

// 		mhbs := AddressNet(networkNum.Bytes())
// 		networkIDs = append(networkIDs, &mhbs)
// 	}
// 	networkIDsLock.Unlock()
// 	return networkIDs
// }

/*
获得一个节点更远的节点中，比自己更远的节点
*/
//func (this *NodeManager) GetIdsForFar(id *AddressNet) []AddressNet {
//	//计算来源的逻辑节点地址
//	kl := NewBucket(len(this.nodes) + 2)
//	kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id))
//	kl.Add(new(big.Int).SetBytes(*id))
//
//	this.nodesLock.RLock()
//	for _, one := range this.nodes {
//		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
//			continue
//		}
//		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
//	}
//	this.nodesLock.RUnlock()
//
//	// Nodes.Range(func(k, v interface{}) bool {
//	// 	value := v.(*Node)
//	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
//	// 	return true
//	// })
//
//	list := kl.Get(new(big.Int).SetBytes(*id))
//
//	out := make([]AddressNet, 0)
//	find := false
//	for _, one := range list {
//		oneBs := one.Bytes()
//		oneNewBs := utils.FullHighPositionZero(&oneBs, config.Addr_byte_length)
//
//		// if hex.EncodeToString(one.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id.Data()) {
//		if bytes.Equal(*oneNewBs, this.NodeSelf.IdInfo.Id) {
//			find = true
//		} else {
//			if find {
//				// bs, err := utils.Encode(one.Bytes(), config.HashCode)
//				// if err != nil {
//				// 	// fmt.Println("编码失败")
//				// 	continue
//				// }
//				// mh := utils.Multihash(bs)
//				mh := AddressNet(*oneNewBs)
//				out = append(out, mh)
//			}
//		}
//
//	}
//
//	return out
//}

/*
添加一个地址到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) AddWhiteListByAddr(addr AddressNet) bool {
	this.WhiteList.AddWhiteListAddr(&addr)
	return true
}

/*
删除一个地址到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) DelWhiteListByAddr(addr AddressNet) bool {
	this.WhiteList.DelWhiteListAddr(&addr)
	return true
}

/*
查询是否存在
*/
func (this *NodeManager) FindWhiteListByAddr(addr *AddressNet) bool {
	return this.WhiteList.CheckWhiteListAddrExist(addr)
}

/*
添加一个ip到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) AddWhiteListByIP(ip ma.Multiaddr) {
	this.WhiteList.AddWhiteListAddrInfo(ip)
}

/*
添加一个ip到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) DelWhiteListByIP(ip ma.Multiaddr) bool {
	this.WhiteList.DelWhiteListAddrInfo(ip)
	return true
}

/*
查询是否存在
*/
func (this *NodeManager) FindWhiteListByIP(ip ma.Multiaddr) bool {
	return this.WhiteList.CheckWhiteListAddrInfoExist(ip)
}

/*
清理节点信息中会话数量为0的记录
*/
func (this *NodeManager) CleanNodeInfo() {
	this.ProxyNodeAddr.CleanNodeInfo()
	this.FreeNodeAddr.CleanNodeInfo()
	this.WhiteList.CleanNodeInfo()
	this.NodeLogicAddrWAN.CleanNodeInfo()
	this.NodeLogicAddrLAN.CleanNodeInfo()
	this.AreaLogicAddrLAN.CleanNodeInfo()
	this.AreaLogicAddrWAN.CleanNodeInfo()
}

// 得到所有连接的节点信息，不包括本节点
//func (this *NodeManager) GetAllNodesAddr() []AddressNet {
//	ids := make([]AddressNet, 0)
//	this.nodesLock.RLock()
//	for _, one := range this.nodes {
//		ids = append(ids, one.IdInfo.Id)
//	}
//	this.nodesLock.RUnlock()
//	return ids
//}

// 得到所有连接的节点信息，不包括本节点
//func (this *NodeManager) GetAllNodes() []*Node {
//	nodes := make([]*Node, 0)
//	this.nodesLock.RLock()
//	for _, one := range this.nodes {
//		nodes = append(nodes, one)
//	}
//	this.nodesLock.RUnlock()
//	return nodes
//}

// func init() {
// 	p1 := "2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk"
// 	AddWhiteList(AddressFromB58String(p1))
// 	p5 := "DNDywcPsJqsWq2gn7gH4yZg5GrAZbR5JvbpxoJDhyoAs"
// 	AddWhiteList(AddressFromB58String(p5))
// 	p6 := "5EontzaTP7Ad8ZQS9GviPQfZMVNMEFh5RMDhYUgZuSqB"
// 	AddWhiteList(AddressFromB58String(p6))
// 	p7 := "XXmf5vZZ7Nf7XbZsf1YN9hW2KdoKsQqyPMvpPGaxLo6"
// 	AddWhiteList(AddressFromB58String(p7))
// 	p8 := "GxCcAvSNRzymqyrRtcXFt9Mz829gFwnZ5TRNBA5Nz2Co"
// 	AddWhiteList(AddressFromB58String(p8))

// 	testp1 := "1E1ZBndCZsVDt1QXkk2gy4CTeh2sEQeVKvv7L3mP14S"
// 	AddWhiteList(AddressFromB58String(testp1))
// 	testp2 := "j9vydTsmTyC7hx6LmNUXYDtge2R4vhaggotVUF7oCPX"
// 	AddWhiteList(AddressFromB58String(testp2))
// 	testp3 := "6RKLimMBb9h2SdXuXooQTgycNdYEVcakxCsMQoWQpy7c"
// 	AddWhiteList(AddressFromB58String(testp3))
// }
