package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
节点发现仿真测试
*/
func main() {
	engine.SetLogPath("log.txt")
	example()
	// BuildIds()
}

func example() {
	//初始化n个节点
	InitIds()
	//使用查询邻居节点的方式，获得逻辑节点
	Discover()
	//打印查询逻辑节点后的结果
	PrintlnLogicNodes()
	utils.Log.Info().Msgf("=================================")

	//打印每个节点的逻辑节点
	PrintlnLogicNodesNew()

	MsgPingNodes()
	//-----------------------------------
	// for _, nodeOne := range nodes {
	// 	nodeOne.logic = make([][]byte, 0)
	// 	nodeOne.client = make([][]byte, 0)
	// }
	//构建各个节点的标准逻辑节点
	BuildLogicNodes()
	fmt.Println("---------------------------")
	//打印每个节点的逻辑节点
	PrintlnLogicNodesNew()
	fmt.Println("---------------------------")
	MsgPingNodes()
	// EqualNodes()
	fmt.Println("end")
}

// 超级节点
var idSuperLength = 40
var idSuper = make([][]byte, 0, idSuperLength)

// 内网中的普通节点
var idOrdinaryLength = 0
var idOrdinary = make([][]byte, 0, idOrdinaryLength)

// 所有节点信息
var nodes = make(map[string]*Node)

var newNodes = make(map[string]*Node)

/*
初始化N个节点
*/
func InitIds() {
	for i := 0; i < idSuperLength; i++ {
		name := utils.GetRandomDomain()
		id := utils.Hash_SHA3_256([]byte(name))
		idSuper = append(idSuper, id)
		utils.Log.Info().Msgf("随机超级节点id %d %s", i, hex.EncodeToString(id))
		nodeOne := Node{
			id:     id,
			logic:  make([][]byte, 0),
			client: make([][]byte, 0),
		}
		nodes[utils.Bytes2string(id)] = &nodeOne

		newNodeOne := nodeOne
		newNodes[utils.Bytes2string(id)] = &newNodeOne
	}
	fmt.Println("---------------------------")

	for i := 0; i < idOrdinaryLength; i++ {
		name := utils.GetRandomDomain()
		id := utils.Hash_SHA3_256([]byte(name))
		idOrdinary = append(idOrdinary, id)
		utils.Log.Info().Msgf("随机普通节点id %d %s", i, hex.EncodeToString(id))
		nodeOne := Node{
			id:     id,
			logic:  make([][]byte, 0),
			client: make([][]byte, 0),
		}
		nodes[utils.Bytes2string(id)] = &nodeOne

		newNodeOne := nodeOne
		newNodes[utils.Bytes2string(id)] = &newNodeOne
	}

}

/*
节点发现，找到自己的逻辑节点
超级节点列表中第一个节点为引导节点
一个节点一个节点的加入网络
*/
func Discover() {
	for i, superAddr := range idSuper {
		if i == 0 {
			continue
		}
		AddNetForSuperNode(superAddr, idSuper[0], idSuper[:i])
	}

	for _, clientAddr := range idOrdinary {
		AddNetForClientNode(clientAddr, idSuper[0])
	}
	MardSuperNodeByClient()
}

/*
一个超级节点加入网络
会引起节点震荡，所有在网的节点都要重新找逻辑节点
@newAddrOne       []byte    新加入的节点
@regAddr          []byte    加入网络时，首个连入的节点
@nowSuperAddrs    [][]byte  已经组网的节点
*/
func AddNetForSuperNode(newAddrOne, regAddr []byte, nowSuperAddrs [][]byte) {
	//新节点添加引导节点为逻辑节点
	newNode := nodes[utils.Bytes2string(newAddrOne)]
	newNode.logic = append(newNode.logic, regAddr)
	//引导节点判断是否添加新节点为逻辑节点
	regNode := nodes[utils.Bytes2string(regAddr)]
	newLogicIds := LogicNodes(regAddr, append(regNode.logic, newAddrOne), nil)
	regNode.logic = newLogicIds

	notChangeCount := 0
	//未改变的次数大于2，则退出循环
	for notChangeCount < 3 {
		change := LoopDiscoverLogicNode(append(nowSuperAddrs, newAddrOne))
		if change {
			notChangeCount = 0
		} else {
			notChangeCount++
		}
	}

}

/*
循环发现自己的逻辑节点
@return    bool    是否有逻辑节点改变
*/
func LoopDiscoverLogicNode(superAddr [][]byte) bool {
	utils.Log.Info().Msgf("循环查找逻辑节点数量:%d", len(superAddr))
	change := false
	for _, one := range superAddr {
		nodeOne := nodes[utils.Bytes2string(one)]
		for _, logicId := range nodeOne.logic {
			logicNode := nodes[utils.Bytes2string(logicId)]
			//获得邻居节点的逻辑节点后，计算自己新的逻辑节点
			tempLogicIds := LogicNodes(one, logicNode.logic, nil)

			newLogicIds := LogicNodes(one, nodeOne.logic, tempLogicIds)
			oldLogicIds := nodeOne.logic
			nodeOne.logic = newLogicIds

			//对比新旧逻辑节点，看是否有改变
			changeOne := EqualIds(oldLogicIds, newLogicIds)
			if change {
				continue
			}
			change = changeOne
		}
	}
	return change
}

/*
一个普通节点加入网络
@newAddrOne       []byte    新加入的节点
@regAddr          []byte    加入网络时，首个连入的节点
@nowSuperAddrs    [][]byte  已经组网的节点
*/
func AddNetForClientNode(newAddrOne, regAddr []byte) {
	//新节点添加引导节点为逻辑节点
	newNode := nodes[utils.Bytes2string(newAddrOne)]
	newNode.logic = append(newNode.logic, regAddr)

	notChangeCount := 0
	//未改变的次数大于2，则退出循环
	for notChangeCount < 3 {
		change := LoopDiscoverClientLogicNode(newAddrOne)
		if change {
			notChangeCount = 0
		} else {
			notChangeCount++
		}
	}
}

/*
建立关联关系
*/
func EstablishAssociation(srcNodeid, dstNodeid []byte) {
	srcNode := nodes[utils.Bytes2string(srcNodeid)]
	dstNode := nodes[utils.Bytes2string(dstNodeid)]
	//获得邻居节点的逻辑节点后，计算自己新的逻辑节点
	tempLogicIds := LogicNodes(srcNodeid, append(dstNode.logic, dstNode.client...), nil)
	newLogicIds := LogicNodes(srcNodeid, append(srcNode.logic, srcNode.client...), tempLogicIds)
	// oldLogicIds := srcNode.logic
	oldClientIds := srcNode.client
	srcNode.logic = newLogicIds

	//找出删除的clients
	clientMap := make(map[string]int)
	for _, one := range oldClientIds {
		clientMap[utils.Bytes2string(one)] = 0
	}

	// for _,

}

/*
循环发现自己的逻辑节点
@return    bool    是否有逻辑节点改变
*/
func LoopDiscoverClientLogicNode(clientId []byte) bool {
	// utils.Log.Info().Msgf("客户端节点循环查找逻辑节点数量:%d", len(superAddr))
	change := false
	clientNode := nodes[utils.Bytes2string(clientId)]
	for _, one := range clientNode.logic {
		nodeOne := nodes[utils.Bytes2string(one)]
		//获得邻居节点的逻辑节点后，计算自己新的逻辑节点
		tempLogicIds := LogicNodes(clientId, nodeOne.logic, nil)

		newLogicIds := LogicNodes(clientId, clientNode.logic, tempLogicIds)
		//对比新旧逻辑节点，看是否有改变
		changeOne := EqualIds(clientNode.logic, newLogicIds)
		clientNode.logic = newLogicIds
		if change {
			continue
		}
		change = changeOne

	}
	return change
}

/*
把客户端节点添加到超级节点中去
*/
func MardSuperNodeByClient() {
	for _, one := range idOrdinary {
		clientNode := nodes[utils.Bytes2string(one)]
		for _, superId := range clientNode.logic {
			superNode := nodes[utils.Bytes2string(superId)]
			superNode.client = append(superNode.client, one)
		}
	}
}

func PrintlnLogicNodes() {
	// fmt.Println("计算逻辑节点")
	for _, one := range idSuper {
		nodeOne := nodes[utils.Bytes2string(one)]
		utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		for _, two := range nodeOne.logic {
			utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		}
		for _, two := range nodeOne.client {
			utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		}
	}
	for _, one := range idOrdinary {
		nodeOne := nodes[utils.Bytes2string(one)]
		utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		for _, two := range nodeOne.logic {
			utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		}
		for _, two := range nodeOne.client {
			utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		}
	}

}

//====================================================
/*
	构建各个节点的逻辑节点
*/
func BuildLogicNodes() {
	// fmt.Println("计算逻辑节点")
	for i, one := range idSuper {
		logicNodes := LogicNodes(one, idSuper, nil)
		nodeOne := newNodes[utils.Bytes2string(one)]
		nodeOne.logic = logicNodes
		utils.Log.Info().Msgf("逻辑节点 %d", i)
	}
	// fmt.Println("计算普通节点")
	for i, one := range idOrdinary {
		logicNodes := LogicNodes(one, idSuper, nil)
		// fmt.Println("普通节点id为:", hex.EncodeToString(one))
		for _, two := range logicNodes {
			// fmt.Println("++++逻辑节点:", hex.EncodeToString(two))
			nodeOne := newNodes[utils.Bytes2string(two)]
			nodeOne.client = append(nodeOne.client, one)
		}
		nodeOne := newNodes[utils.Bytes2string(one)]
		nodeOne.logic = logicNodes
		utils.Log.Info().Msgf("普通节点 %d", i)
	}
}

func PrintlnLogicNodesNew() {
	// fmt.Println("计算逻辑节点")
	for _, one := range idSuper {
		nodeOne := newNodes[utils.Bytes2string(one)]
		utils.Log.Info().Msgf("逻辑节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		for _, two := range nodeOne.logic {
			utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		}
		for _, two := range nodeOne.client {
			utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		}
	}
	for _, one := range idOrdinary {
		nodeOne := newNodes[utils.Bytes2string(one)]
		utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		for _, two := range nodeOne.logic {
			utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		}
		for _, two := range nodeOne.client {
			utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		}
	}

}

/*
节点之间互相ping
*/
func MsgPingNodes() {
	count := 0
	idAll := append(idSuper, idOrdinary...)
	for _, fromId := range idAll {
		for _, toId := range idAll {
			count++
			auccess := false
			srcId := fromId
			for i := 1; i < 100; i++ {
				ok, dstId := LoopSendMsg(srcId, toId)
				if ok {
					// fmt.Println("发送成功！跳转次数", i)
					utils.Log.Info().Msgf("%d 发送成功！跳转次数 %d", count, i)
					auccess = true
					break
				}
				srcId = dstId
			}
			if !auccess {
				utils.Log.Info().Msgf("发送失败！")
				panic("发送失败!")
			}
		}
	}
}

func LoopSendMsg(fromId, toId []byte) (bool, []byte) {
	fromNode, ok := newNodes[utils.Bytes2string(fromId)]
	if !ok {
		// fmt.Println("本次发送:", hex.EncodeToString(fromId), hex.EncodeToString(toId))
		utils.Log.Info().Msgf("本次发送:%s %s", hex.EncodeToString(fromId), hex.EncodeToString(toId))
		panic("未找到这个id")
	}
	logicAll := make([][]byte, 0)
	if fromNode.logic != nil && len(fromNode.logic) > 0 {
		logicAll = append(logicAll, fromNode.logic...)
	}
	if fromNode.client != nil && len(fromNode.client) > 0 {
		logicAll = append(logicAll, fromNode.client...)
	}
	for _, logicOne := range logicAll {
		if bytes.Equal(logicOne, toId) {
			return true, nil
		}
	}
	kl := nodeStore.NewKademlia(len(fromNode.logic))
	for _, one := range fromNode.logic {
		kl.Add(new(big.Int).SetBytes(one))
	}
	dstId := kl.Get(new(big.Int).SetBytes(toId))
	dstIdBs := dstId[0].Bytes()
	dstIdBsP := utils.FullHighPositionZero(&dstIdBs, 32)
	return false, *dstIdBsP
}

/*
计算一个节点，应该有哪些逻辑节点
*/
func LogicNodes(id []byte, idSuper, idOrdinary [][]byte) [][]byte {
	idsm := nodeStore.NewIds(id, 256)
	for _, one := range idSuper {
		if bytes.Equal(id, one) {
			continue
		}
		idsm.AddId(one)
	}
	for _, one := range idOrdinary {
		if bytes.Equal(id, one) {
			continue
		}
		idsm.AddId(one)
	}
	return idsm.GetIds()
}

type Node struct {
	id     []byte
	logic  [][]byte
	client [][]byte
}

/*
对比两个id是否一样
@return    bool    是否改变 true=已经改变;false=未改变;
*/
func EqualIds(oldIds, newIds [][]byte) bool {
	if len(oldIds) != len(newIds) {
		return true
	}
	m := make(map[string]bool)
	for _, one := range oldIds {
		m[utils.Bytes2string(one)] = false
		// utils.Log.Info().Msgf("添加这个id %s", hex.EncodeToString(one))
	}
	for _, one := range newIds {
		_, ok := m[utils.Bytes2string(one)]
		if !ok {
			utils.Log.Info().Msgf("未找到这个id %s", hex.EncodeToString(one))
			return true
		}
	}
	return false
}

func EqualNodes() {
	for _, nodeOne := range nodes {
		newNodeOne := newNodes[utils.Bytes2string(nodeOne.id)]

		// utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(nodeOne.id), len(nodeOne.logic), len(nodeOne.client))
		// for _, two := range nodeOne.logic {
		// 	utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		// }
		// for _, two := range nodeOne.client {
		// 	utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		// }

		// utils.Log.Info().Msgf("client节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(newNodeOne.id), len(newNodeOne.logic), len(newNodeOne.client))
		// for _, two := range newNodeOne.logic {
		// 	utils.Log.Info().Msgf("++++逻辑节点:%s", hex.EncodeToString(two))
		// }
		// for _, two := range newNodeOne.client {
		// 	utils.Log.Info().Msgf("++++客户端节点:%s", hex.EncodeToString(two))
		// }

		if EqualIds(nodeOne.logic, newNodeOne.logic) {
			utils.Log.Info().Msgf("逻辑节点不一样 %s", hex.EncodeToString(nodeOne.id))
			for _, one := range newNodeOne.logic {
				utils.Log.Info().Msgf("标准逻辑节点:%s", hex.EncodeToString(one))
			}
			for _, one := range nodeOne.logic {
				utils.Log.Info().Msgf("实际逻辑节点:%s", hex.EncodeToString(one))
			}
			// panic("")
		}
		if EqualIds(nodeOne.client, newNodeOne.client) {
			utils.Log.Info().Msgf("client节点不一样 %s", hex.EncodeToString(nodeOne.id))
			// panic("")
		}
	}
}
