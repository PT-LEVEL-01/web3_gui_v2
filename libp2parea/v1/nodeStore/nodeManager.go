package nodeStore

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/utils"
)

type NodeClass int

const (
	Node_min = 0 //每个节点的最少连接数量

	Node_type_all        NodeClass = 0 //包含所有类型
	Node_type_logic      NodeClass = 1 //自己需要的逻辑节点
	Node_type_client     NodeClass = 2 //保存其他逻辑节点连接到自己的节点，都是超级节点
	Node_type_proxy      NodeClass = 3 //被代理的节点
	Node_type_other      NodeClass = 4 //每个节点有最少连接数量
	Node_type_white_list NodeClass = 5 //连接白名单
	Node_type_oneByone   NodeClass = 6 //onebyone规则连接类型
)

type NodeManager struct {
	NodeSelf         *Node            //
	nodesLock        *sync.RWMutex    //
	nodes            map[string]*Node //保存所有类型的节点，通过Type参数区分开
	OutFindNode      chan *AddressNet //= make(chan *AddressNet, 1000) //需要查询的逻辑节点
	OutCloseConnName chan *AddressNet //= make(chan *AddressNet, 1000) //废弃的nodeid，需要询问是否关闭
	SuperPeerId      *AddressNet      //超级节点名称
	NodeIdLevel      uint             //= 256  //节点id长度
	HaveNewNode      chan *Node       //= make(chan *Node, 100) //当添加新的超级节点时，给他个信号
	WhiteList        *sync.Map        //new(sync.Map) //连接白名单
	NetType          config.NetModel  //网络类型:正式网络release/测试网络test
	superPeerIdLock  *sync.RWMutex    //超级节点id锁
	IsHeadTailModl   bool
	AreaNameSelf     []byte // 本节点大区名
}

var (
	NodeIdLevel uint = 256 //节点id长度
)

/*
设置网络环境
*/
func (this *NodeManager) SetNetType(t config.NetModel) {
	this.NetType = t
}

/*
获取超级节点id
*/
func (this *NodeManager) GetSuperPeerId() (sAddr *AddressNet) {
	this.superPeerIdLock.RLock()
	sAddr = this.SuperPeerId
	this.superPeerIdLock.RUnlock()
	return
}

/*
设置超级节点id
*/
func (this *NodeManager) SetSuperPeerId(sAddr *AddressNet) {
	this.superPeerIdLock.Lock()
	this.SuperPeerId = sAddr
	this.superPeerIdLock.Unlock()
}

func NewNodeManager(key keystore.KeystoreInterface, pwd string) (*NodeManager, error) {
	// config.Wallet_keystore_default_pwd = pwd
	//加载本地网络id
	idinfo, err := BuildIdinfo(key, pwd)
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}

	// 如果生成的DH公钥为空, 则给出对应的提示信息
	if idinfo == nil || idinfo.CPuk == [32]byte{} {
		utils.Log.Error().Msgf("DH公钥为空!!!! keystore生成的公钥为空, 请确认keystore的用法是否正确!!!!")
	}

	nodeSelf := Node{
		IdInfo: *idinfo, //节点id信息，id字符串以16进制显示
		// IsSuper              uint32    `json:"issuper"` //是不是超级节点，超级节点有外网ip地址，可以为其他节点提供代理服务
		// Addr                 string    `json:"addr"`    //外网ip地址
		// TcpPort              uint16    `json:"tcpport"` //TCP端口
		// IsApp                bool      `json:"isapp"`   //是不是手机端节点
		// MachineID: utils.GetRandomOneInt64(), //每个节点启动的时候生成一个随机数，用作判断多个节点使用同一个key连入网络的情况
		Version: config.Version_0, //版本号
		// lastContactTimestamp time.Time `json:"-"`       //最后检查的时间戳
		// Type                 NodeClass `json:"-"`       //
	}

	nodeManager := NodeManager{
		NodeSelf:         &nodeSelf,
		nodesLock:        new(sync.RWMutex),            //
		nodes:            make(map[string]*Node),       //保存所有类型的节点，通过Type参数区分开
		OutFindNode:      make(chan *AddressNet, 1000), //需要查询的逻辑节点
		OutCloseConnName: make(chan *AddressNet, 1000), //废弃的nodeid，需要询问是否关闭
		// SuperPeerId      *AddressNet      //超级节点名称
		NodeIdLevel:     256,                   //节点id长度
		HaveNewNode:     make(chan *Node, 100), //当添加新的超级节点时，给他个信号
		WhiteList:       new(sync.Map),         //连接白名单
		superPeerIdLock: new(sync.RWMutex),     //超级节点id锁
	}
	return &nodeManager, nil

}

/*
加载本地私钥生成idinfo
*/
func BuildIdinfo(key keystore.KeystoreInterface, pwd string) (*IdInfo, error) {
	keyPair := key.GetDHKeyPair()
	prk, puk, err := key.GetNetAddr(pwd)
	if err != nil {
		return nil, err
	}
	addrNet := BuildAddr(puk)

	idinfo := IdInfo{
		Id:   addrNet,                   //id，网络地址，临时地址，由updatetime字符串产生MD5值
		EPuk: puk,                       //ed25519公钥
		CPuk: keyPair.KeyPair.PublicKey, //curve25519公钥
		V:    uint32(keyPair.Index),     //dh密钥版本
		// Sign: sign,                      //ed25519私钥签名
	}
	idinfo.SignDHPuk(prk)
	return &idinfo, nil
}

// 添加一个代理节点
func (this *NodeManager) AddProxyNode(node Node) {
	// Proxys.Store(node.IdInfo.Id.B58String(), node)
	// utils.Log.Info().Msgf("add proxys nodeid:%s", node.IdInfo.Id.B58String())
	node.Type = Node_type_proxy
	this.nodesLock.Lock()
	this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
	this.nodesLock.Unlock()
	// Proxys.Store(utils.Bytes2string(node.IdInfo.Id), &node)
}

// 得到一个代理节点
func (this *NodeManager) GetProxyNode(id string) (node *Node, ok bool) {
	this.nodesLock.RLock()
	node, ok = this.nodes[id]
	if node != nil && node.Type != Node_type_proxy {
		node = nil
		ok = false
	}
	this.nodesLock.RUnlock()
	return
}

/*
检查节点是否是本节点的逻辑节点
只检查，不保存
*/
func (this *NodeManager) CheckNeedNode(nodeId *AddressNet) (isNeed bool, removeIDs [][]byte) {
	// utils.Log.Info().Msgf("add node:%s", node.IdInfo.Id.B58String())
	//不添加自己
	if bytes.Equal(*nodeId, this.NodeSelf.IdInfo.Id) {
		// utils.Log.Info().Msgf("%s CheckNeedNode 111111 need:%t", this.NodeSelf.IdInfo.Id.B58String(), nodeId.B58String(), false)
		return false, nil
	}

	//检查是否在白名单中
	if this.FindWhiteList(nodeId) {
		return true, nil
	}

	idm := NewIds(this.NodeSelf.IdInfo.Id, NodeIdLevel)

	this.nodesLock.Lock()

	//检查是否够最少连接数量
	total := 0
	for _, one := range this.nodes {
		if one.Type != Node_type_logic {
			continue
		}
		total += 1
		idm.AddId(one.IdInfo.Id)
	}
	//不够就保存
	if total < Node_min {
		return true, nil
	}

	ok, removeIDs := idm.AddId(*nodeId)
	this.nodesLock.Unlock()
	// utils.Log.Info().Msgf("%s CheckNeedNode 22222 target:%s need:%t", this.NodeSelf.IdInfo.Id.B58String(), nodeId.B58String(), ok)
	return ok, removeIDs
}

/*
添加一个超级节点
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
*/
func (this *NodeManager) AddNode(node Node) bool {

	ok := false
	var removeIDs [][]byte

	nodeType := Node_type_white_list

	//检查是否在白名单中
	if this.FindWhiteList(&node.IdInfo.Id) {
		nodeType = Node_type_white_list
		ok = true
	} else {
		nodeType = Node_type_logic
		ok, removeIDs = this.CheckNeedNode(&node.IdInfo.Id)
		if !ok {
			nodeType = Node_type_oneByone
			ok = true
		}
	}

	node.lastContactTimestamp = time.Now()

	this.nodesLock.Lock()
	// utils.Log.Info().Msgf("Nodes super add:%s %d", node.IdInfo.Id.B58String(), len(removeIDs))
	nodeLoad, ok := this.nodes[utils.Bytes2string(node.IdInfo.Id)]
	if ok {
		nodeLoad.Type = nodeType
	} else {
		node.Type = nodeType
		this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
	}

	// Nodes.Store(utils.Bytes2string(node.IdInfo.Id), &node)
	// utils.Log.Info().Msgf("node info:%+v", node)
	select {
	case this.HaveNewNode <- &node:
	default:
	}

	if nodeType == Node_type_oneByone {
		this.nodesLock.Unlock()
		return true
	}
	//修改超级节点，普通节点经常切换影响网络
	// addrNet := AddressNet(idm.GetIndex(0))
	// this.SuperPeerId = &addrNet
	this.SetSuperPeerId(&node.IdInfo.Id)

	//删除被替换的id
	for _, one := range removeIDs {
		//			idOne := hex.EncodeToString(one)
		//			delete(Nodes, idOne)
		//			OutCloseConnName <- idOne
		addrNet := AddressNet(one)
		// utils.Log.Info().Msgf("%s 应该删除 %s", this.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String())
		addrStr := utils.Bytes2string(addrNet)
		//如果自己是对方节点的逻辑节点，则不删除，保存到ClientNodes中

		nodeLoad, ok := this.nodes[addrStr]
		if ok {
			if nodeLoad.Type == Node_type_white_list {
				continue
			}
			// utils.Log.Info().Msgf("add client nodeid:%s", addrNet.B58String())
			nodeLoad.Type = Node_type_client
		}
		if !ok {
			continue
		}

		// utils.Log.Info().Msgf("Nodes super del:%s", addrNet.B58String())

		//如果排除的节点在自己的白名单中，就不询问关闭连接了
		if this.FindWhiteList(&addrNet) {
			continue
		}
		//询问对方，自己是否是对方的逻辑节点，如果是则保留连接，如果不是，则关闭连接
		select {
		case this.OutCloseConnName <- &addrNet:
			// utils.Log.Info().Msgf("Nodes super del2:%s", addrNet.B58String())
		default:
		}
		//TODO 以上是询问对方的方式自治网络，需要对方配合，如果对方不配合，可以直接关闭连接，让对方重新连接自己
	}

	this.nodesLock.Unlock()
	//	fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("add node end:%s %t", node.IdInfo.Id.B58String(), ok)
	return true

}

/*
添加一个超级节点
检查这个节点是否是自己的逻辑节点，如果是，则保存
不保存自己
*/
func (this *NodeManager) AddNode_old(node Node) bool {
	// utils.Log.Info().Msgf("add node:%s", node.IdInfo.Id.B58String())
	//不添加自己
	if bytes.Equal(node.IdInfo.Id, this.NodeSelf.IdInfo.Id) {
		return false
	}

	//检查是否够最少连接数量
	// total := 0
	// OtherNodes.Range(func(k, v interface{}) bool {
	// 	total++
	// 	return true
	// })
	// //不够就保存
	// if total < Node_min {
	// 	//保存
	// 	OtherNodes.Store(utils.Bytes2string(node.IdInfo.Id), node)
	// }

	//检查是否在白名单中
	if this.FindWhiteList(&node.IdInfo.Id) {
		// utils.Log.Info().Msgf("add node:%s", node.IdInfo.Id.B58String())
		this.nodesLock.Lock()
		node.Type = Node_type_white_list
		this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		this.nodesLock.Unlock()
		return true
	}

	idm := NewIds(this.NodeSelf.IdInfo.Id, NodeIdLevel)
	// ids := GetLogicNodes()

	this.nodesLock.Lock()

	//检查是否够最少连接数量
	total := 0
	for _, one := range this.nodes {
		if one.Type != Node_type_logic {
			continue
		}
		total += 1
		// ids = append(ids, one.IdInfo.Id)
		idm.AddId(one.IdInfo.Id)
	}
	//不够就保存
	if total < Node_min {
		// utils.Log.Info().Msgf("add node:%s", node.IdInfo.Id.B58String())
		node.Type = Node_type_logic
		this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		this.nodesLock.Unlock()
		return true
	}

	// for _, one := range ids {
	// 	idm.AddId(one)
	// }

	ok, removeIDs := idm.AddId(node.IdInfo.Id)
	if ok {
		//		fmt.Println("添加成功", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())
		node.lastContactTimestamp = time.Now()
		// utils.Log.Info().Msgf("Nodes super add: %s %d", node.IdInfo.Id.B58String(), len(removeIDs))
		// utils.Log.Info().Msgf("Nodes super add:%s %d", node.IdInfo.Id.B58String(), len(removeIDs))

		nodeLoad, ok := this.nodes[utils.Bytes2string(node.IdInfo.Id)]
		if ok {
			nodeLoad.Type = Node_type_logic
		} else {
			node.Type = Node_type_logic
			this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		}

		// Nodes.Store(utils.Bytes2string(node.IdInfo.Id), &node)
		// utils.Log.Info().Msgf("node info:%+v", node)
		select {
		case this.HaveNewNode <- &node:
		default:
		}

		//修改超级节点，普通节点经常切换影响网络
		addrNet := AddressNet(idm.GetIndex(0))
		this.superPeerIdLock.Lock()
		this.SuperPeerId = &addrNet
		this.superPeerIdLock.Unlock()

		//删除被替换的id
		for _, one := range removeIDs {
			//			idOne := hex.EncodeToString(one)
			//			delete(Nodes, idOne)
			//			OutCloseConnName <- idOne
			addrNet := AddressNet(one)
			// utils.Log.Info().Msgf("应该删除self:%s %s", this.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String())
			addrStr := utils.Bytes2string(addrNet)
			//如果自己是对方节点的逻辑节点，则不删除，保存到ClientNodes中

			nodeLoad, ok := this.nodes[addrStr]
			if ok {
				if nodeLoad.Type == Node_type_white_list {
					continue
				}
				// utils.Log.Info().Msgf("add client nodeid:%s", addrNet.B58String())
				nodeLoad.Type = Node_type_client
			}
			if !ok {
				continue
			}

			// nodeOne, ok := Nodes.Load(addrStr)
			// if !ok {
			// 	continue
			// }
			// Nodes.Delete(addrStr)
			// NodesClient.Store(addrStr, nodeOne)

			// utils.Log.Info().Msgf("Nodes super del:%s", addrNet.B58String())

			//如果排除的节点在自己的白名单中，就不询问关闭连接了
			if this.FindWhiteList(&addrNet) {
				continue
			}
			//询问对方，自己是否是对方的逻辑节点，如果是则保留连接，如果不是，则关闭连接
			select {
			case this.OutCloseConnName <- &addrNet:
				// utils.Log.Info().Msgf("Nodes super del2:%s", addrNet.B58String())
			default:
			}
			//TODO 以上是询问对方的方式自治网络，需要对方配合，如果对方不配合，可以直接关闭连接，让对方重新连接自己
		}
	}
	this.nodesLock.Unlock()
	//	fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("add node end:%s %t", node.IdInfo.Id.B58String(), ok)
	return ok

}

/*
删除一个节点，包括超级节点和代理节点
*/
func (this *NodeManager) DelNode(id *AddressNet) {
	idStr := utils.Bytes2string(*id) //  id.B58String()
	// utils.Log.Info().Msgf("delete client nodeid %s", id.B58String())
	this.nodesLock.Lock()
	delete(this.nodes, idStr)
	this.nodesLock.Unlock()
}

/*
通过id查找一个节点
*/
func (this *NodeManager) FindNode(id *AddressNet) *Node {
	idStr := utils.Bytes2string(*id)
	var node *Node
	this.nodesLock.RLock()
	node, _ = this.nodes[idStr]
	this.nodesLock.RUnlock()
	return node
}

func (this *NodeManager) FindNodeInLogic(id *AddressNet) *Node {
	idStr := utils.Bytes2string(*id)

	var node *Node
	// var ok bool
	this.nodesLock.RLock()
	node, _ = this.nodes[idStr]
	this.nodesLock.RUnlock()
	if node != nil && node.Type != Node_type_logic && node.Type != Node_type_white_list {
		return nil
	}
	return node

	// v, ok := Nodes.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }
	// return nil
}

/*
对比逻辑节点是否有变化
*/
func (this *NodeManager) EqualLogicNodes(ids []AddressNet) bool {
	newLogicNodes := this.GetLogicNodes()
	if len(ids) != len(newLogicNodes) {
		return true
	}
	isChange := false
	this.nodesLock.RLock()
	for _, one := range ids {
		idStr := utils.Bytes2string(one)
		// _, ok := Nodes.Load(idStr)
		_, ok := this.nodes[idStr]
		if !ok {
			isChange = true
			break
			// return true
		}
	}
	this.nodesLock.RUnlock()
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
func (this *NodeManager) FindNearInSuper(nodeId, outId *AddressNet, includeSelf bool, blockAddr map[string]int) *AddressNet {
	kl := NewKademlia(len(this.nodes) + 1)
	if includeSelf {
		// utils.Log.Info().Msgf("添加查询逻辑节点:%s", this.NodeSelf.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id))
	}

	addSome := false
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list && one.Type != Node_type_client {
			continue
		}
		if outId != nil && bytes.Equal(one.IdInfo.Id, *outId) {
			continue
		}
		//去除黑名单地址
		if blockAddr[utils.Bytes2string(one.IdInfo.Id)] > engine.MaxRetryCnt {
			continue
		}
		addSome = true
		//utils.Log.Info().Msgf("add add %s", one.IdInfo.Id.B58String())
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
		// utils.Log.Info().Msgf("添加查询逻辑节点:%s", one.IdInfo.Id.B58String())
	}
	this.nodesLock.RUnlock()

	if !addSome && !includeSelf {
		return nil
	}
	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	targetIdBs := targetId.Bytes()
	mh := AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// utils.Log.Info().Msgf("检查id结果:%s", mh.B58String())
	return &mh
}

/*
在超级节点中找到最近的节点，不包括代理节点，带黑名单，在onebyone规则下使用
@nodeId       *AddressNet     要查找的节点
@outId        *AddressNet     排除一个节点
@blockAddr    []*AddressNet   已经试过，但失败了被拉黑的节点
@return       *AddressNet     查找到的节点id，可能为空
*/
func (this *NodeManager) FindNearInSuperOnebyone(nodeId, outId *AddressNet, blockAddr map[string]int, sessionEngine *engine.Engine) *AddressNet {

	var gs []engine.Session

	//判断需要中转的地址比自己大还是小，确定方向
	if new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id).Cmp(new(big.Int).SetBytes(*nodeId)) == -1 {
		//向大地址传
		for _, one := range sessionEngine.GetAllUpSession(this.AreaNameSelf) {
			if outId != nil && bytes.Equal([]byte(one.GetName()), *outId) {
				continue
			}
			//去掉黑名单地址
			if blockAddr[one.GetName()] > engine.MaxRetryCnt {
				continue
			}
			gs = append(gs, one)
		}

	} else {
		//向小地址传
		for _, one := range sessionEngine.GetAllDownSession(this.AreaNameSelf) {
			if outId != nil && bytes.Equal([]byte(one.GetName()), *outId) {
				continue
			}
			//去掉黑名单地址
			if blockAddr[one.GetName()] > engine.MaxRetryCnt {
				continue
			}
			gs = append(gs, one)
		}
	}

	if len(gs) <= 1 {
		if len(gs) == 0 {
			return nil
		}
		target := gs[0]
		if target == nil {
			return nil
		}
		rs := AddressNet(target.GetName())
		return &rs
	}

	onebyone := new(IdDESC)
	gsm := make(map[*big.Int]engine.Session)
	for i, _ := range gs {
		space := new(big.Int).Abs(new(big.Int).Sub(new(big.Int).SetBytes([]byte(gs[i].GetName())), new(big.Int).SetBytes(*nodeId)))
		gsm[space] = gs[i]
		*onebyone = append(*onebyone, space)
	}

	sort.Sort(onebyone)
	targetInt := (*onebyone)[len(*onebyone)-1]
	target := AddressNet(gsm[targetInt].GetName())
	return &target

}

// 在节点中找到最近的节点，包括代理节点
func (this *NodeManager) FindNearNodeId(nodeId, outId *AddressNet, includeSelf bool) *AddressNet {
	kl := NewKademlia(len(this.nodes) + 1)
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id))
	}
	// outIdStr := ""
	// if outId != nil {
	// 	outIdStr = utils.Bytes2string(*outId) // outId.B58String()
	// }

	this.nodesLock.RLock()

	for _, one := range this.nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_proxy && one.Type != Node_type_white_list {
			continue
		}
		if outId != nil && bytes.Equal(one.IdInfo.Id, *outId) {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
	}

	this.nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	if k.(string) == outIdStr {
	// 		return true
	// 	}
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })
	// //代理节点
	// Proxys.Range(func(k, v interface{}) bool {
	// 	if k.(string) == outIdStr {
	// 		return true
	// 	}
	// 	value := v.(*Node)
	// 	//过滤APP节点
	// 	if value.IsApp {
	// 		return true
	// 	}
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	targetIdBs := targetId.Bytes()
	mh := AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// mh := AddressNet(targetId.Bytes())
	return &mh
}

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
func (this *NodeManager) GetWhiltListNodes() []AddressNet {
	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_white_list {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	this.nodesLock.RUnlock()
	return ids
}

// 得到所有逻辑节点，不包括本节点，也不包括代理节点
func (this *NodeManager) GetLogicNodes() []AddressNet {
	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		// utils.Log.Info().Msgf("打印所有连接:%s", one.IdInfo.Id.B58String())
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	this.nodesLock.RUnlock()
	return ids
}

// 获取所有onebyone规则连接下的节点
func (this *NodeManager) GetOneByOneNodes() []AddressNet {
	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type == Node_type_oneByone {
			ids = append(ids, one.IdInfo.Id)
		}
	}
	this.nodesLock.RUnlock()
	return ids
}

/*
得到所有逻辑节点，不包括本节点，也不包括代理节点
*/
func (this *NodeManager) GetLogicNodeInfo() []*Node {
	ids := make([]*Node, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		// utils.Log.Info().Msgf("打印所有连接:%s", one.IdInfo.Id.B58String())
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		ids = append(ids, one)
	}
	this.nodesLock.RUnlock()
	return ids
}

func (this *NodeManager) IsLogicNode(addr AddressNet) bool {
	var isLogic bool
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		if bytes.Equal(addr, one.IdInfo.Id) {
			isLogic = true
			break
		}
	}
	this.nodesLock.RUnlock()
	return isLogic
}

/*
获得所有代理节点
*/
func (this *NodeManager) GetProxyAll() []AddressNet {
	// ids := make([]string, 0)
	// Proxys.Range(func(key, value interface{}) bool {
	// 	ids = append(ids, key.(string))
	// 	return true
	// })
	// return ids

	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_proxy {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	this.nodesLock.RUnlock()
	// Proxys.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ids = append(ids, value.IdInfo.Id)
	// 	return true
	// })
	return ids
}

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
添加一个被其他节点当作逻辑节点的连接
*/
func (this *NodeManager) AddNodesClient(node Node) {
	// utils.Log.Info().Msgf("add client nodeid:%s", node.IdInfo.Id.B58String())
	node.Type = Node_type_client
	this.nodesLock.Lock()
	this.nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
	this.nodesLock.Unlock()
	// NodesClient.Store(utils.Bytes2string(node.IdInfo.Id), &node)
	select {
	case this.HaveNewNode <- &node:
	default:
	}
}

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
获取本机被其他当作逻辑节点的连接
*/
func (this *NodeManager) GetNodesClient() []AddressNet {
	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_client {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	this.nodesLock.RUnlock()
	// NodesClient.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ids = append(ids, value.IdInfo.Id)
	// 	return true
	// })
	return ids
}

/*
获得本机所有逻辑节点的ip地址
*/
func (this *NodeManager) GetSuperNodeIps() (ips []string) {
	ips = make([]string, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		// ids = append(ids, one.IdInfo.Id)
		ips = append(ips, one.Addr+":"+strconv.Itoa(int(one.TcpPort)))
	}
	this.nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ips = append(ips, value.Addr+":"+strconv.Itoa(int(value.TcpPort)))
	// 	return true
	// })

	return
}

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

// 	ids := NewIds(this.NodeSelf.IdInfo.Id, NodeIdLevel)
// 	for _, one := range this.GetLogicNodes() {
// 		ids.AddId(one)
// 	}
// 	ok, _ := ids.AddId(*nodeId)
// 	return ok
// }

/*
检查连接到本机的节点是否可以作为逻辑节点
*/
func (this *NodeManager) CheckClientNodeIsLogicNode() {
	this.nodesLock.Lock()
	for _, one := range this.nodes {
		if one.Type != Node_type_client {
			continue
		}
		//是本身节点不添加
		if bytes.Equal(one.IdInfo.Id, this.NodeSelf.IdInfo.Id) {
			continue
		}
		//在白名单中
		if this.FindWhiteList(&one.IdInfo.Id) {
			one.Type = Node_type_white_list
			continue
		}

		logicNodes := make([]AddressNet, 0)
		for _, two := range this.nodes {
			if two.Type != Node_type_logic {
				continue
			}
			logicNodes = append(logicNodes, two.IdInfo.Id)
		}
		//检查是否需要
		if len(logicNodes) == 0 {
			one.Type = Node_type_logic
			continue
		}

		ids := NewIds(this.NodeSelf.IdInfo.Id, NodeIdLevel)
		for _, two := range logicNodes {
			ids.AddId(two)
		}
		ok, _ := ids.AddId(one.IdInfo.Id)
		if ok {
			one.Type = Node_type_logic
		}
	}
	this.nodesLock.Unlock()
	// for _, one := range clientNodes {
	// 	if nodeStore.CheckNeedNode(&one) {
	// 		node := nodeStore.FindNode(&one)
	// 		nodeStore.SwitchNodesClientToLogic(*node)
	// 	}
	// }
}

type LogicNumBuider struct {
	lock  *sync.RWMutex
	id    *[]byte
	level uint
	idStr string
	ids   []*[]byte
}

/*
得到每个节点网络的网络号，不包括本节点
@id        *utils.Multihash    要计算的id
@level     int                 深度
*/
func (this *LogicNumBuider) GetNodeNetworkNum() []*[]byte {

	this.lock.RLock()
	if this.idStr != "" && this.idStr == hex.EncodeToString(*this.id) {
		this.lock.RUnlock()
		return this.ids
	}
	this.lock.RUnlock()

	this.lock.Lock()
	this.idStr = hex.EncodeToString(*this.id) // .B58String()

	root := new(big.Int).SetBytes(*this.id)

	this.ids = make([]*[]byte, 0)
	for i := 0; i < int(this.level); i++ {
		//---------------------------------
		//将后面的i位置零
		//---------------------------------
		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
		//---------------------------------
		//第i位取反
		//---------------------------------
		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

		mhbs := networkNum.Bytes()

		this.ids = append(this.ids, &mhbs)
	}
	this.lock.Unlock()

	return this.ids
}

func NewLogicNumBuider(id []byte, level uint) *LogicNumBuider {
	return &LogicNumBuider{
		lock:  new(sync.RWMutex),
		id:    &id,
		level: level,
		idStr: "",
		ids:   make([]*[]byte, 0),
	}
}

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
func (this *NodeManager) GetIdsForFar(id *AddressNet) []AddressNet {
	//计算来源的逻辑节点地址
	kl := NewKademlia(len(this.nodes) + 2)
	kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id))
	kl.Add(new(big.Int).SetBytes(*id))

	this.nodesLock.RLock()
	for _, one := range this.nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
	}
	this.nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	list := kl.Get(new(big.Int).SetBytes(*id))

	out := make([]AddressNet, 0)
	find := false
	for _, one := range list {
		oneBs := one.Bytes()
		oneNewBs := utils.FullHighPositionZero(&oneBs, config.Addr_byte_length)

		// if hex.EncodeToString(one.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id.Data()) {
		if bytes.Equal(*oneNewBs, this.NodeSelf.IdInfo.Id) {
			find = true
		} else {
			if find {
				// bs, err := utils.Encode(one.Bytes(), config.HashCode)
				// if err != nil {
				// 	// fmt.Println("编码失败")
				// 	continue
				// }
				// mh := utils.Multihash(bs)
				mh := AddressNet(*oneNewBs)
				out = append(out, mh)
			}
		}

	}

	return out
}

/*
添加一个地址到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) AddWhiteList(addr AddressNet) bool {
	this.WhiteList.Store(utils.Bytes2string(addr), 0)
	node := this.FindNode(&addr)
	if node == nil {
		return false
	}
	node.Type = Node_type_white_list
	return true
}

/*
删除一个地址到白名单
@return    bool    这个连接是否已经存在
*/
func (this *NodeManager) DelWhiteList(addr AddressNet) bool {
	this.WhiteList.Delete(utils.Bytes2string(addr))
	node := this.FindNode(&addr)
	if node == nil {
		return false
	}
	//降级为逻辑节点
	node.Type = Node_type_logic
	return true
}

func (this *NodeManager) FindWhiteList(addr *AddressNet) bool {
	_, ok := this.WhiteList.Load(utils.Bytes2string(*addr))
	return ok
}

func (this *NodeManager) FindOneByOneList(addr *AddressNet) bool {
	this.nodesLock.RLock()
	defer this.nodesLock.RUnlock()
	tt, ok := this.nodes[utils.Bytes2string(*addr)]
	if ok && tt.Type == Node_type_oneByone {
		return true
	}
	return false
}

// 得到所有连接的节点信息，不包括本节点
func (this *NodeManager) GetAllNodesAddr() []AddressNet {
	ids := make([]AddressNet, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		ids = append(ids, one.IdInfo.Id)
	}
	this.nodesLock.RUnlock()
	return ids
}

// 得到所有连接的节点信息，不包括本节点
func (this *NodeManager) GetAllNodes() []*Node {
	nodes := make([]*Node, 0)
	this.nodesLock.RLock()
	for _, one := range this.nodes {
		nodes = append(nodes, one)
	}
	this.nodesLock.RUnlock()
	return nodes
}

/*
 * 关闭所有不是上帝节点的节点连接信息，根据上帝节点session的index进行判断
 *
 * @auth: qlw
 * @param sessIndex uint64		上帝节点的session连接index
 */
func (this *NodeManager) CloseNotGodNodesUseSessIndex(sessIndex uint64) {
	// utils.Log.Info().Msgf("qlw---开始关闭所有非上帝节点的节点信息 上帝节点session连接index为:%d", sessIndex)

	// 获取所有的连接节点地址信息
	this.nodesLock.Lock()
	var allNodes []*Node
	for _, v := range this.nodes {
		allNodes = append(allNodes, v)
	}
	this.nodesLock.Unlock()

	// 遍历所有节点，关闭所有session连接index不是上帝连接index的节点
	for i := range allNodes {
		if allNodes[i] == nil {
			continue
		}

		sessions := allNodes[i].GetSessions()
		// utils.Log.Info().Msgf("qlw-----关闭session, cnt:%d", len(sessions))
		for i := range sessions {
			if sessions[i].GetIndex() == sessIndex {
				continue
			}

			// utils.Log.Info().Msgf("qlw-----关闭session, index:%d, target:%s", i, AddressNet([]byte(sessions[i].GetName())).B58String())
			sessions[i].Close()
		}
	}
	// utils.Log.Info().Msgf("qlw---结束关闭所有非上帝节点的节点信息 上帝节点session连接index为:%d --------", sessIndex)
}

/*
 * SetMachineID 设置机器id
 * 	机器id由使用者控制，p2p只负责传输和保存
 *
 * @param	machineID		string		机器id
 */
func (this *NodeManager) SetMachineID(machineID string) {
	if this.NodeSelf == nil {
		utils.Log.Error().Msgf("还没初始化Node,请先初始化,再进行相应操作!!!!")
		return
	}

	// 运行期间不能修改机器id
	if this.NodeSelf.MachineIDStr != "" {
		utils.Log.Error().Msgf("还没初始化Node,请先初始化,再进行相应操作!!!!")
		return
	}

	this.NodeSelf.MachineIDStr = machineID
}

/*
 * GetMachineID 获取机器id
 *
 * @return	machineID		string		机器id
 */
func (this *NodeManager) GetMachineID() string {
	if this.NodeSelf == nil {
		utils.Log.Error().Msgf("还没初始化Node,请先初始化,再进行相应操作!!!!")
		return ""
	}

	return this.NodeSelf.MachineIDStr
}

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
