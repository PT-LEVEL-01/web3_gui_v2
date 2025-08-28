package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
计算各个节点的标准逻辑节点
*/
func main() {
	example()
	// BuildIds()
}

func example() {
	InitIds()
	BuildLogicNodes()
	fmt.Println("---------------------------")
	PrintlnLogicNodes()
	fmt.Println("---------------------------")
	MsgPingNodes()
	fmt.Println("end")
}

var idSuperLength = 1000
var idSuper = make([][]byte, 0, idSuperLength)
var idOrdinaryLength = 5
var idOrdinary = make([][]byte, 0, idOrdinaryLength)

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
	}
}

/*
构建各个节点的逻辑节点
*/
func BuildLogicNodes() {
	// fmt.Println("计算逻辑节点")
	for i, one := range idSuper {
		logicNodes := LogicNodes(one, idSuper, nil)
		// fmt.Println("逻辑节点id为:", hex.EncodeToString(one))
		// for _, two := range logicNodes {
		// 	// fmt.Println("++++逻辑节点:", hex.EncodeToString(two))
		// 	nodeOne := nodes[utils.Bytes2string(two)]
		// 	nodeOne.client = append(nodeOne.client, one)
		// }
		nodeOne := nodes[utils.Bytes2string(one)]
		nodeOne.logic = logicNodes
		utils.Log.Info().Msgf("逻辑节点 %d", i)
	}
	// fmt.Println("计算普通节点")
	for i, one := range idOrdinary {
		logicNodes := LogicNodes(one, idSuper, nil)
		// fmt.Println("普通节点id为:", hex.EncodeToString(one))
		for _, two := range logicNodes {
			// fmt.Println("++++逻辑节点:", hex.EncodeToString(two))
			nodeOne := nodes[utils.Bytes2string(two)]
			nodeOne.client = append(nodeOne.client, one)
		}
		nodeOne := nodes[utils.Bytes2string(one)]
		nodeOne.logic = logicNodes
		utils.Log.Info().Msgf("普通节点 %d", i)
	}
}

func PrintlnLogicNodes() {
	// fmt.Println("计算逻辑节点")
	for _, v := range nodes {
		utils.Log.Info().Msgf("节点id为: %s 逻辑节点数：%d %d", hex.EncodeToString(v.id), len(v.logic), len(v.client))
		// for _, two := range v.logic {
		// 	fmt.Println("++++逻辑节点:", hex.EncodeToString(two))
		// }
		// for _, two := range v.client {
		// 	fmt.Println("++++客户端节点:", hex.EncodeToString(two))
		// }
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
	fromNode, ok := nodes[utils.Bytes2string(fromId)]
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

var nodes = make(map[string]*Node)

type Node struct {
	id     []byte
	logic  [][]byte
	client [][]byte
}
