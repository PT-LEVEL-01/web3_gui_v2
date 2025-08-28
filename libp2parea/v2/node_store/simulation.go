/*
仿真测试
*/
package nodeStore

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"web3_gui/utils"
)

/*
多个节点管理器
*/
type NodeSimulationManager struct {
	AddrPre           string                     //
	IDdepth           uint                       //id有多少位
	NodeSuperIDs      []AddressNet               //所有已知的超级节点id
	NodeStandardSuper map[string]*NodeSimulation //标准的逻辑节点
	NodeCustomSuper   map[string]*NodeSimulation //通过节点发现构建的
	Bucket            *Bucket                    //
}

/*
获取标准节点
*/
func (this *NodeSimulationManager) GetStandardNodes() map[string]*NodeSimulation {
	return this.NodeStandardSuper
}

/*
获取自定义节点
*/
func (this *NodeSimulationManager) GetCustomNodes() map[string]*NodeSimulation {
	return this.NodeCustomSuper
}

/*
随机数构建足够数量的节点
@superLen     int
@clientLen    int
*/
func (this *NodeSimulationManager) BuildNode(superLen, clientLen int) utils.ERROR {
	var ERR utils.ERROR
	if this.IDdepth == 0 {
		this.IDdepth = 32 * 8
	}
	this.NodeSuperIDs = make([]AddressNet, 0)
	this.NodeStandardSuper = make(map[string]*NodeSimulation)
	this.NodeCustomSuper = make(map[string]*NodeSimulation)
	this.Bucket = NewBucket(superLen)
	for i := 0; i < superLen; i++ {
		var id AddressNet
		for j := 0; j < 100; j++ {
			name := utils.GetRandomDomain()
			idData := utils.Hash_SHA3_256([]byte(name))
			if this.IDdepth < 8 {
				idBig := new(big.Int).SetBytes(idData[:1])
				idBig = idBig.Rsh(idBig, 8-this.IDdepth)
				idData = idBig.Bytes()
				if len(idData) == 0 {
					continue
					idData = []byte{0}
				}
			}
			id, ERR = BuildAddrByData(this.AddrPre, idData)
			if ERR.CheckFail() {
				return ERR
			}
			//判断id是否重复，重复则重新生成
			if _, ok := this.NodeCustomSuper[utils.Bytes2string(idData)]; !ok {
				break
			}
		}
		this.NodeSuperIDs = append(this.NodeSuperIDs, id)
		if this.IDdepth < 8 {
			utils.Log.Info().Msgf("随机超级节点id NO:%d 16进制:%s 10进制:%d", i, hex.EncodeToString(id.Data()), id.Data()[0])
		} else {
			utils.Log.Info().Msgf("随机超级节点id NO:%d 16进制:%s", i, hex.EncodeToString(id.Data()))
		}
		nodeOne := &NodeSimulation{
			Id:     id,
			Logic:  make([]AddressNet, 0),
			Client: make([]AddressNet, 0),
		}
		// utils.Log.Info().Msgf("BuildNode one:%s %p", AddressNet(id).B58String(), nodeOne)
		this.NodeStandardSuper[utils.Bytes2string(id.Data())] = nodeOne
		nodeTemp := *nodeOne
		// utils.Log.Info().Msgf("BuildNode one:%s %p", AddressNet(id).B58String(), &nodeTemp)
		this.NodeCustomSuper[utils.Bytes2string(id.Data())] = &nodeTemp
		this.Bucket.Add(new(big.Int).SetBytes(id.Data()))
	}
	// fmt.Println("---------------------------")

	// for i := 0; i < superLen; i++ {
	// 	name := utils.GetRandomDomain()
	// 	id := utils.Hash_SHA3_256([]byte(name))
	// 	idOrdinary = append(idOrdinary, id)
	// 	utils.Log.Info().Msgf("随机普通节点id %d %s", i, hex.EncodeToString(id))
	// 	nodeOne := Node{
	// 		id:     id,
	// 		logic:  make([][]byte, 0),
	// 		client: make([][]byte, 0),
	// 	}
	// 	nodes[utils.Bytes2string(id)] = &nodeOne

	// 	newNodeOne := nodeOne
	// 	newNodes[utils.Bytes2string(id)] = &newNodeOne
	// }
	return utils.NewErrorSuccess()
}

/*
添加指定的超级节点ID
@superID    []byte    指定的节点id
*/
func (this *NodeSimulationManager) AddNodeSuperIDs(superID ...AddressNet) {
	if this.NodeSuperIDs == nil {
		this.NodeSuperIDs = make([]AddressNet, 0)
		this.NodeStandardSuper = make(map[string]*NodeSimulation)
		this.NodeCustomSuper = make(map[string]*NodeSimulation)
		this.Bucket = NewBucket(0)
	}
	for _, id := range superID {
		this.NodeSuperIDs = append(this.NodeSuperIDs, id)
		// utils.Log.Info().Msgf("随机超级节点id %d %s", i, hex.EncodeToString(id))
		nodeOne := &NodeSimulation{
			Id:     id,
			Logic:  make([]AddressNet, 0),
			Client: make([]AddressNet, 0),
		}
		this.NodeStandardSuper[utils.Bytes2string(id.Data())] = nodeOne
		nodeTemp := *nodeOne
		this.NodeCustomSuper[utils.Bytes2string(id.Data())] = &nodeTemp
		this.Bucket.Add(new(big.Int).SetBytes(id.Data()))
	}
}

/*
添加指定的代理节点ID
@proxyID    []byte    指定的节点id
*/
func (this *NodeSimulationManager) AddNodeProxyIDs(proxyID ...AddressNet) {
	if this.NodeSuperIDs == nil {
		this.NodeSuperIDs = make([]AddressNet, 0)
		this.NodeStandardSuper = make(map[string]*NodeSimulation)
		this.NodeCustomSuper = make(map[string]*NodeSimulation)
	}
	for _, id := range proxyID {
		this.NodeSuperIDs = append(this.NodeSuperIDs, id)
		// utils.Log.Info().Msgf("随机超级节点id %d %s", i, hex.EncodeToString(id))
		nodeOne := &NodeSimulation{
			Id:     id,
			Logic:  make([]AddressNet, 0),
			Client: make([]AddressNet, 0),
		}
		this.NodeStandardSuper[utils.Bytes2string(id.Data())] = nodeOne
		nodeTemp := *nodeOne
		this.NodeCustomSuper[utils.Bytes2string(id.Data())] = &nodeTemp
		this.Bucket.Add(new(big.Int).SetBytes(id.Data()))
	}
}

/*
构建标准的逻辑节点
*/
func (this *NodeSimulationManager) BuildNodeLogicIDs() utils.ERROR {
	// fmt.Println("计算逻辑节点")
	for _, nodeOne := range this.NodeStandardSuper {
		logicDatas := LogicNodes(nodeOne.Id, this.NodeSuperIDs, nil, this.IDdepth)
		// nodeOne := this[utils.Bytes2string(nodeOne.id)]
		logicNodes := make([]AddressNet, 0, len(logicDatas))
		for _, one := range logicDatas {
			logicAddr, ERR := BuildAddrByData(this.AddrPre, one)
			if ERR.CheckFail() {
				return ERR
			}
			logicNodes = append(logicNodes, logicAddr)
		}
		nodeOne.Logic = logicNodes
		// utils.Log.Info().Msgf("逻辑节点 %d", i)
	}
	// fmt.Println("计算普通节点")
	// for i, one := range idOrdinary {
	// 	logicNodes := LogicNodes(one, idSuper, nil)
	// 	// fmt.Println("普通节点id为:", hex.EncodeToString(one))
	// 	for _, two := range logicNodes {
	// 		// fmt.Println("++++逻辑节点:", hex.EncodeToString(two))
	// 		nodeOne := newNodes[utils.Bytes2string(two)]
	// 		nodeOne.client = append(nodeOne.client, one)
	// 	}
	// 	nodeOne := newNodes[utils.Bytes2string(one)]
	// 	nodeOne.logic = logicNodes
	// 	utils.Log.Info().Msgf("普通节点 %d", i)
	// }
	return utils.NewErrorSuccess()
}

/*
节点发现，找到自己的逻辑节点
超级节点列表中第一个节点为引导节点
一个节点一个节点的加入网络
*/
func (this *NodeSimulationManager) Discover() {
	for i, superAddr := range this.NodeSuperIDs {
		if i == 0 {
			continue
		}
		this.AddNetForSuperNode(superAddr, this.NodeSuperIDs[0], this.NodeSuperIDs[:i])
	}

	// for _, clientAddr := range idOrdinary {
	// 	AddNetForClientNode(clientAddr, this.NodeSuperIDs[0])
	// }
	this.MardSuperNodeByClient()
}

/*
把客户端节点添加到超级节点中去
*/
func (this *NodeSimulationManager) MardSuperNodeByClient() {
	//整理关系前，先把之前的关系清空
	for _, clientNode := range this.NodeCustomSuper {
		clientNode.Client = make([]AddressNet, 0)
	}
	for _, clientNode := range this.NodeCustomSuper {
		// clientNode := nodes[utils.Bytes2string(one)]
		for _, superId := range clientNode.Logic {
			superNode := this.NodeCustomSuper[utils.Bytes2string(superId.Data())]
			superNode.Client = append(superNode.Client, clientNode.Id)
		}
	}

	//整理关系前，先把之前的关系清空
	for _, clientNode := range this.NodeStandardSuper {
		clientNode.Client = make([]AddressNet, 0)
	}
	for _, clientNode := range this.NodeStandardSuper {
		// clientNode := nodes[utils.Bytes2string(one)]
		for _, superId := range clientNode.Logic {
			superNode := this.NodeStandardSuper[utils.Bytes2string(superId.Data())]
			superNode.Client = append(superNode.Client, clientNode.Id)
		}
	}
}

/*
一个超级节点加入网络
会引起节点震荡，所有在网的节点都要重新找逻辑节点
@newAddrOne       []byte    新加入的节点
@regAddr          []byte    加入网络时，首个连入的节点
@nowSuperAddrs    [][]byte  已经组网的节点
*/
func (this *NodeSimulationManager) AddNetForSuperNode(newAddrOne, regAddr AddressNet, nowSuperAddrs []AddressNet) utils.ERROR {
	// utils.Log.Info().Msgf("AddNetForSuperNode")
	//新节点添加引导节点为逻辑节点
	newNode := this.NodeCustomSuper[utils.Bytes2string(newAddrOne.Data())]
	newNode.Logic = append(newNode.Logic, regAddr)
	//引导节点判断是否添加新节点为逻辑节点
	regNode := this.NodeCustomSuper[utils.Bytes2string(regAddr.Data())]
	logicDatas := LogicNodes(regAddr, append(regNode.Logic, newAddrOne), nil, this.IDdepth)

	logicNodes := make([]AddressNet, 0, len(logicDatas))
	for _, one := range logicDatas {
		logicAddr, ERR := BuildAddrByData(this.AddrPre, one)
		if ERR.CheckFail() {
			return ERR
		}
		logicNodes = append(logicNodes, logicAddr)
	}

	regNode.Logic = logicNodes

	// for _, one := range newNode.Logic {
	// 	utils.Log.Info().Msgf("new节点逻辑ID:%s", hex.EncodeToString(one))
	// }
	// for _, one := range regNode.Logic {
	// 	utils.Log.Info().Msgf("reg节点逻辑ID:%s", hex.EncodeToString(one))
	// }

	notChangeCount := 0
	//未改变的次数大于2，则退出循环
	for notChangeCount < 3 {
		change, ERR := this.LoopDiscoverLogicNode(append(nowSuperAddrs, newAddrOne))
		if ERR.CheckFail() {
			return ERR
		}
		if change {
			notChangeCount = 0
			// utils.Log.Info().Msgf("notChangeCount")
		} else {
			notChangeCount++
		}
	}
	return utils.NewErrorSuccess()
}

/*
循环发现自己的逻辑节点
@return    bool    是否有逻辑节点改变
*/
func (this *NodeSimulationManager) LoopDiscoverLogicNode(superAddr []AddressNet) (bool, utils.ERROR) {
	utils.Log.Info().Msgf("循环查找逻辑节点数量:%d", len(superAddr))
	change := false
	for _, one := range superAddr {
		nodeOne := this.NodeCustomSuper[utils.Bytes2string(one.Data())]
		for _, logicId := range nodeOne.Logic {
			logicNode := this.NodeCustomSuper[utils.Bytes2string(logicId.Data())]

			var candidate []AddressNet
			// utils.Log.Info().Msgf("计算应该有的逻辑节点ID:%s", hex.EncodeToString(one))
			// candidate := make([][]byte, 0)
			// candidate = append(candidate, logicNode.Id)
			// candidate = append(candidate, logicNode.Logic...)
			// candidate = append(candidate, logicNode.Client...)
			// candidate = append(candidate, nodeOne.Logic...)
			// candidate = append(candidate, nodeOne.Client...)
			// temp := LogicNodes(one, candidate, nil, this.IDdepth)
			// utils.Log.Info().Msgf("1应该有的逻辑节点ID:%s %d %d", hex.EncodeToString(one), one[0], temp)

			// utils.Log.Info().Msgf("nodeOne:%+v logicNode:%+v", nodeOne, logicNode)

			//获得邻居节点的逻辑节点后，计算自己新的逻辑节点
			candidate = make([]AddressNet, 0)
			candidate = append(candidate, logicNode.Id)
			candidate = append(candidate, logicNode.Logic...)
			candidate = append(candidate, logicNode.Client...)
			tempLogicDatas := LogicNodes(one, candidate, nil, this.IDdepth)
			// utils.Log.Info().Msgf("tempLogicIds 长度:%d %d %d", len(tempLogicIds), append(logicNode.Logic, nodeOne.Client...), one)
			tempLogicIds := make([]AddressNet, 0, len(tempLogicDatas))
			for _, one := range tempLogicDatas {
				logicAddr, ERR := BuildAddrByData(this.AddrPre, one)
				if ERR.CheckFail() {
					return false, ERR
				}
				tempLogicIds = append(tempLogicIds, logicAddr)
			}

			candidate = make([]AddressNet, 0)
			candidate = append(candidate, nodeOne.Logic...)
			candidate = append(candidate, nodeOne.Client...)
			candidate = append(candidate, tempLogicIds...)
			newLogicDatas := LogicNodes(one, candidate, nil, this.IDdepth)
			// utils.Log.Info().Msgf("newLogicIds 长度:%d", len(newLogicIds))

			newLogicIds := make([]AddressNet, 0, len(newLogicDatas))
			for _, one := range newLogicDatas {
				logicAddr, ERR := BuildAddrByData(this.AddrPre, one)
				if ERR.CheckFail() {
					return false, ERR
				}
				newLogicIds = append(newLogicIds, logicAddr)
			}

			oldLogicIds := nodeOne.Logic
			nodeOne.Logic = newLogicIds
			// utils.Log.Info().Msgf("2应该有的逻辑节点ID:%s %d %d", hex.EncodeToString(one), one[0], newLogicIds)

			// for _, one := range oldLogicIds {
			// 	utils.Log.Info().Msgf("对比旧节点ID:%s", hex.EncodeToString(one))
			// }
			// for _, one := range newLogicIds {
			// 	utils.Log.Info().Msgf("对比新节点ID:%s", hex.EncodeToString(one))
			// }
			//对比新旧逻辑节点，看是否有改变
			changeOne, _ := EqualIds(oldLogicIds, newLogicIds)
			// utils.Log.Info().Msgf("对比结果:%t", changeOne)
			if change {
				continue
			}
			change = changeOne
		}
	}
	this.MardSuperNodeByClient()
	return change, utils.NewErrorSuccess()
}

/*
节点之间互相ping
*/
func (this *NodeSimulationManager) MsgPingNodesP2P() {
	count := 0
	idAll := this.NodeSuperIDs
	for _, fromId := range idAll {
		for _, toId := range idAll {
			utils.Log.Info().Msgf("发送给节点from:%s to:%s", fromId.B58String(), toId.B58String())
			count++
			auccess := false
			srcId := fromId
			for i := 1; i < 100; i++ {
				ok, dstId, ERR := this.LoopSendMsgP2P(srcId, toId)
				if ERR.CheckFail() {
					panic(ERR.String())
				}
				if ok {
					// fmt.Println("发送成功！跳转次数", i)
					utils.Log.Info().Msgf("%d 发送成功！跳转次数 %d", count, i)
					auccess = true
					break
				}
				srcId = *dstId
			}
			if !auccess {
				utils.Log.Info().Msgf("发送失败！")
				panic("发送失败!")
			}
		}
	}
}

func (this *NodeSimulationManager) LoopSendMsgP2P(fromId, toId AddressNet) (bool, *AddressNet, utils.ERROR) {
	fromNode, ok := this.NodeCustomSuper[utils.Bytes2string(fromId.Data())]
	if !ok {
		// fmt.Println("本次发送:", hex.EncodeToString(fromId), hex.EncodeToString(toId))
		utils.Log.Info().Msgf("本次发送:%s %s", fromId.B58String(), toId.B58String())
		panic("未找到这个id")
	}
	logicAll := make([]AddressNet, 0)
	if fromNode.Logic != nil && len(fromNode.Logic) > 0 {
		logicAll = append(logicAll, fromNode.Logic...)
	}
	if fromNode.Client != nil && len(fromNode.Client) > 0 {
		logicAll = append(logicAll, fromNode.Client...)
	}
	for _, logicOne := range logicAll {
		if bytes.Equal(logicOne.Data(), toId.Data()) {
			utils.Log.Info().Msgf("找到节点了")
			return true, nil, utils.NewErrorSuccess()
		}
	}
	kl := NewBucket(len(fromNode.Logic))
	for _, one := range fromNode.Logic {
		// utils.Log.Info().Msgf("kad add:%s", AddressNet(one).B58String())
		kl.Add(new(big.Int).SetBytes(one.Data()))
	}
	dstId := kl.Get(new(big.Int).SetBytes(toId.Data()))
	dstIdBs := dstId[0].Bytes()
	dstIdBsP := utils.FullHighPositionZero(&dstIdBs, int(this.IDdepth/8))
	addr, ERR := BuildAddrByData(fromId.GetPre(), *dstIdBsP)
	if ERR.CheckFail() {
		return false, nil, ERR
	}
	utils.Log.Info().Msgf("转发给节点:%s", addr.B58String())
	return false, &addr, utils.NewErrorSuccess()
}

/*
每个节点搜索逻辑节点，对比结果是否一致
*/
func (this *NodeSimulationManager) MsgPingNodesSearch() utils.ERROR {
	var ERR utils.ERROR
	randId := this.NodeSuperIDs[0]
	var logicIds []*AddressNet
	if this.IDdepth < 8 {
		logicIds, ERR = GetMagneticID16(this.AddrPre)
	} else {
		logicIds, ERR = GetMagneticID100(&randId)
	}
	if ERR.CheckFail() {
		return ERR
	}
	count := 0
	for _, toId := range logicIds {
		//toCustomId := make(map[string]string)
		for _, fromId := range this.NodeSuperIDs {

			//计算全网磁力节点
			dstId := this.Bucket.Get(new(big.Int).SetBytes(toId.Data()))
			dstIdBs := dstId[0].Bytes()
			dstIdBsP := utils.FullHighPositionZero(&dstIdBs, int(this.IDdepth/8))
			tempAddr, ERR := BuildAddrByData(this.AddrPre, *dstIdBsP)
			if ERR.CheckFail() {
				return ERR
			}
			utils.Log.Info().Msgf("搜索节点from:%s to:%s want:%s", fromId.B58String(), toId.B58String(), tempAddr.B58String())

			count++
			auccess := false
			srcId := fromId
			for i := 1; i < 10; i++ {
				ok, dstId, ERR := this.LoopSendMsgSearch(fromId, srcId, *toId)
				if ERR.CheckFail() {
					return ERR
				}
				if ok {
					// fmt.Println("发送成功！跳转次数", i)

					wantId := this.Bucket.Get(new(big.Int).SetBytes(toId.Data()))
					wantIdBs := wantId[0].Bytes()
					wantIdBsP := utils.FullHighPositionZero(&wantIdBs, int(this.IDdepth/8))
					tempAddr, ERR := BuildAddrByData(this.AddrPre, *wantIdBsP)
					if ERR.CheckFail() {
						return ERR
					}
					utils.Log.Info().Msgf("No:%d 发送成功！跳转次数:%d find:%s want:%s", count, i,
						dstId.B58String(), tempAddr.B58String())

					auccess = true

					if !bytes.Equal(*dstIdBsP, dstId.Data()) {
						utils.Log.Info().Msgf("No:%d 不相等 %s", count, dstId.B58String())
					}

					// if len(toCustomId) == 0 {
					// 	utils.Log.Info().Msgf("No:%d 收到的节点 %s", count, AddressNet(dstId).B58String())
					// 	toCustomId[utils.Bytes2string(dstId)] = ""
					// } else {
					// 	_, ok := toCustomId[utils.Bytes2string(dstId)]
					// 	if !ok {
					// 		utils.Log.Info().Msgf("No:%d 不相等 %s", count, AddressNet(dstId).B58String())
					// 		for one, _ := range toCustomId {
					// 			utils.Log.Info().Msgf("结果集:%s", AddressNet([]byte(one)).B58String())
					// 		}
					// 		toCustomId[utils.Bytes2string(dstId)] = ""
					// 	}
					// }
					break
				}
				srcId = *dstId
			}
			if !auccess {
				utils.Log.Info().Msgf("发送失败！")
				// panic("发送失败!")
			}
		}
		//for one, _ := range toCustomId {
		//	utils.Log.Info().Msgf("查询结果集:%s", AddressNet([]byte(one)).B58String())
		//}
	}
	return utils.NewErrorSuccess()
}

func (this *NodeSimulationManager) LoopSendMsgSearch(selfID, fromId, toId AddressNet) (bool, *AddressNet, utils.ERROR) {
	fromNode, ok := this.NodeCustomSuper[utils.Bytes2string(fromId.Data())]
	if !ok {
		// fmt.Println("本次发送:", hex.EncodeToString(fromId), hex.EncodeToString(toId))
		utils.Log.Info().Msgf("本次发送:%s %s", fromId.B58String(), toId.B58String())
		panic("未找到这个id")
	}
	logicAll := make([]AddressNet, 0)
	if fromNode.Logic != nil && len(fromNode.Logic) > 0 {
		logicAll = append(logicAll, fromNode.Logic...)
	}
	// if fromNode.Client != nil && len(fromNode.Client) > 0 {
	// 	logicAll = append(logicAll, fromNode.Client...)
	// }
	// for _, logicOne := range logicAll {
	// 	if bytes.Equal(logicOne, toId) {
	// 		utils.Log.Info().Msgf("找到节点了")
	// 		return true, nil
	// 	}
	// }
	kl := NewBucket(len(fromNode.Logic))
	for _, one := range logicAll {
		// if bytes.Equal(fromId, one) {
		// 	continue
		// }
		// utils.Log.Info().Msgf("kad add:%s", AddressNet(one).B58String())
		kl.Add(new(big.Int).SetBytes(one.Data()))
	}
	// if !bytes.Equal(selfID, fromId) {
	kl.Add(new(big.Int).SetBytes(fromNode.Id.Data()))
	// }
	dstId := kl.Get(new(big.Int).SetBytes(toId.Data()))
	dstIdBs := dstId[0].Bytes()
	dstIdBsP := utils.FullHighPositionZero(&dstIdBs, int(this.IDdepth/8))

	tempAddr, ERR := BuildAddrByData(this.AddrPre, *dstIdBsP)
	if ERR.CheckFail() {
		return false, nil, ERR
	}
	utils.Log.Info().Msgf("转发给节点:%s", tempAddr.B58String())
	if bytes.Equal(*dstIdBsP, fromId.Data()) {
		return true, &fromId, utils.NewErrorSuccess()
	}
	return false, &tempAddr, utils.NewErrorSuccess()
}

/*
打印标准逻辑节点
*/
func (this *NodeSimulationManager) PrintlnStandardLogicNodesNew(idB58 bool) {
	utils.Log.Info().Msgf("打印仿真程序中，构建的标准逻辑节点列表")
	if idB58 {
		for _, nodeOne := range this.NodeStandardSuper {
			// nodeOne := newNodes[utils.Bytes2string(one)]
			utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", nodeOne.Id.B58String(), len(nodeOne.Logic), len(nodeOne.Client))
			for _, two := range nodeOne.Logic {
				utils.Log.Info().Msgf("++++逻辑节点:%s", two.B58String())
			}
			for _, two := range nodeOne.Client {
				utils.Log.Info().Msgf("++++客户端节点:%s", two.B58String())
			}
		}
	} else {
		for _, nodeOne := range this.NodeCustomSuper {
			// nodeOne := newNodes[utils.Bytes2string(one)]
			utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.Id.Data()),
				len(nodeOne.Logic), len(nodeOne.Client))
			for _, two := range nodeOne.Logic {
				utils.Log.Info().Msgf("++++逻辑节点:%s %d", hex.EncodeToString(two.Data()), two)
			}
			for _, two := range nodeOne.Client {
				utils.Log.Info().Msgf("++++客户端节点:%s %d", hex.EncodeToString(two.Data()), two)
			}
		}
	}
}

func (this *NodeSimulationManager) PrintlnLogicNodesNew(idB58 bool) {
	if idB58 {
		utils.Log.Info().Msgf("打印仿真程序中，模拟节点发现，构建的逻辑节点列表")
		for _, nodeOne := range this.NodeCustomSuper {
			// nodeOne := newNodes[utils.Bytes2string(one)]
			utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", nodeOne.Id.B58String(), len(nodeOne.Logic), len(nodeOne.Client))
			for _, two := range nodeOne.Logic {
				utils.Log.Info().Msgf("++++逻辑节点:%s", two.B58String())
			}
			for _, two := range nodeOne.Client {
				utils.Log.Info().Msgf("++++客户端节点:%s", two.B58String())
			}
		}
		// for _, one := range idOrdinary {
		// 	nodeOne := newNodes[utils.Bytes2string(one)]
		// 	utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		// 	for _, two := range nodeOne.logic {
		// 		utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		// 	}
		// 	for _, two := range nodeOne.client {
		// 		utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		// 	}
		// }
	} else {
		utils.Log.Info().Msgf("打印仿真程序中，模拟节点发现，构建的逻辑节点列表")
		for _, nodeOne := range this.NodeCustomSuper {
			// nodeOne := newNodes[utils.Bytes2string(one)]
			utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.Id.Data()), len(nodeOne.Logic), len(nodeOne.Client))
			for _, two := range nodeOne.Logic {
				utils.Log.Info().Msgf("++++逻辑节点:%s %d", hex.EncodeToString(two.Data()), two)
			}
			for _, two := range nodeOne.Client {
				utils.Log.Info().Msgf("++++客户端节点:%s %d", hex.EncodeToString(two.Data()), two)
			}
		}
		// for _, one := range idOrdinary {
		// 	nodeOne := newNodes[utils.Bytes2string(one)]
		// 	utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		// 	for _, two := range nodeOne.logic {
		// 		utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		// 	}
		// 	for _, two := range nodeOne.client {
		// 		utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		// 	}
		// }
	}
}

/*
一个节点，保存逻辑节点
*/
type NodeSimulation struct {
	Id     AddressNet
	Logic  []AddressNet
	Client []AddressNet
}

/*
计算一个节点，应该有哪些逻辑节点
@id            []byte    计算的节点
@idSuper       [][]byte    已知的逻辑节点
@idOrdinary    [][]byte    补充的逻辑节点
*/
func LogicNodes(id AddressNet, idSuper, idOrdinary []AddressNet, level uint) [][]byte {
	// utils.Log.Info().Msgf("LogicNodes:%d %d %d", len(idSuper), len(idOrdinary), level)
	idsm := NewKademlia(id.Data(), level, utils.Log)
	for _, one := range idSuper {
		if bytes.Equal(id.Data(), one.Data()) {
			continue
		}
		// utils.Log.Info().Msgf("idsm.AddId:%s", hex.EncodeToString(one))
		idsm.AddId(one.Data())
	}
	for _, one := range idOrdinary {
		if bytes.Equal(id.Data(), one.Data()) {
			continue
		}
		// utils.Log.Info().Msgf("idsm.AddId:%s", hex.EncodeToString(one))
		idsm.AddId(one.Data())
	}
	return idsm.GetIds()
}

/*
对比标准构建的逻辑节点和自动发现的逻辑节点差异
*/
func EqualNodes(nsm *NodeSimulationManager) {
	utils.Log.Info().Msgf("Simulation:打印仿真程序中，对比仿真发现的逻辑节点与标准逻辑节点对比结果")
	for _, nodeOne := range nsm.NodeStandardSuper {
		newNodeOne := nsm.NodeCustomSuper[utils.Bytes2string(nodeOne.Id.Data())]

		change, collective := EqualIds(nodeOne.Logic, newNodeOne.Logic)
		if change {
			utils.Log.Error().Msgf("逻辑节点不一样 %s", hex.EncodeToString(nodeOne.Id.Data()))
			for _, one := range nodeOne.Logic {
				for _, two := range collective {
					if bytes.Equal(one.Data(), two.Data()) {
						continue
					}
					utils.Log.Info().Msgf("✙:%s", hex.EncodeToString(one.Data()))
				}
			}
			for _, one := range newNodeOne.Logic {
				for _, two := range collective {
					if bytes.Equal(one.Data(), two.Data()) {
						continue
					}
					utils.Log.Info().Msgf("✕:%s", hex.EncodeToString(one.Data()))
				}
			}
		} else {
			utils.Log.Info().Msgf("逻辑节点一样 %s", hex.EncodeToString(nodeOne.Id.Data()))
			for _, one := range collective {
				utils.Log.Info().Msgf("实际逻辑节点:%s", hex.EncodeToString(one.Data()))
			}
		}
	}
}

/*
对比两个id是否一样
@return    bool        是否改变 true=已经改变;false=未改变;
@return    [][]byte    共有的逻辑节点
*/
func EqualIds(oldIds, newIds []AddressNet) (isChange bool, collective []AddressNet) {
	collective = make([]AddressNet, 0)
	// if len(oldIds) != len(newIds) {
	// 	isChange = true
	// }
	m := make(map[string]bool)
	for _, one := range oldIds {
		m[utils.Bytes2string(one.Data())] = false
	}
	for _, one := range newIds {
		_, ok := m[utils.Bytes2string(one.Data())]
		if ok {
			collective = append(collective, one)
		} else {
			isChange = true
		}
	}
	return
}

/*
在逻辑节点中搜索节点id
*/
func SearchNodeID(logicIDs [][]byte, findID []byte) []byte {

	netids := make([]*big.Int, 0)
	for _, one := range logicIDs {
		netids = append(netids, new(big.Int).SetBytes(one))
	}
	asc := NewIdASC(new(big.Int).SetBytes(findID), netids)
	result := asc.Sort()

	// for _, idBig := range result {
	// 	netidBs := idBig.Bytes()
	// 	utils.Log.Info().Msgf("最近的节点:%s", AddressNet(netidBs).B58String())
	// }
	return result[0].Bytes()
}
