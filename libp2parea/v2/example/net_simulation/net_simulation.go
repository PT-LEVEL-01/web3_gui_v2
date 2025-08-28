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
测试消息经过的节点转发的路径
测试消息可达性
*/
func main() {
	example()
}

/*
仿真模拟节点之间发送消息
*/
func example() {
	utils.LogBuildDefaultFile("D:\\test\\temp\\logs/log.txt")
	//RandIds()
	InitIds()
	BuildLogicNodes()
	utils.Log.Info().Msgf("---------------------------")
	PrintlnLogicNodes()
	utils.Log.Info().Msgf("---------------------------")
	MsgPingNodes()
	utils.Log.Info().Msgf("end")
}

var idSuperLength = 10 //外网节点数量
var idSuper = make([][]byte, 0, idSuperLength)
var idOrdinaryLength = 5 //代理节点数量
var idOrdinary = make([][]byte, 0, idOrdinaryLength)

/*
随机数构建指定数量的节点id
*/
func RandIds() {
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
预设指定数量的地址
*/
func InitIds() {
	addr := nodeStore.AddressFromB58String("D5UpxhSAW4gG4B1fdEA2afmV4yzFLjhNcQgZu9eAENGA")
	idSuper = append(idSuper, addr)
	nodeInfo := NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("FtpZxwdRo5DxLaagDC4YKTLVFzZj8JfXTcc7M9vwL5zV")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("7451t9dxbhttK7H9NUJefE26GrXLxuJhs6wjzvhSi12N")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("wnSZMB3oWgCc59834WdS2nDExG1Nri2zCdKjt72FqNZ")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("5X3gAETTXNfYVyYWjqep4fQnYXbiZg8FmsLhQXvygfdN")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("HZcbNKLWXrzujiwbmHJahXprVNjXphboaqSEtnZQJwvb")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("HxGS1TrPjpVyg657aG4B7T62anUNKgaBtSSnSiuLKon8")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("SdrN89X12TYYqRDKo4SBJiiEq4xnjuTFofSv46VMU7h")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("HduKPWLikv3TR4a2E4EcePKKmgxup2CmZ6dNjdzF2phU")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
	addr = nodeStore.AddressFromB58String("87YS2GK7nXDKZWisLKZnnhDStTMX9sV7fVMgejNL933d")
	idSuper = append(idSuper, addr)
	nodeInfo = NewNode(addr)
	nodes[utils.Bytes2string(addr)] = nodeInfo
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
		utils.Log.Info().Msgf("逻辑节点:%d", i)
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
		utils.Log.Info().Msgf("普通节点:%d", i)
	}
}

func PrintlnLogicNodes() {
	// fmt.Println("计算逻辑节点")
	for _, v := range nodes {
		utils.Log.Info().Msgf("节点id为:%s 逻辑节点数:%d 代理节点数:%d", nodeStore.AddressNet(v.id).B58String(), len(v.logic), len(v.client))
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
	utils.Log.Info().Msgf("发送者:%s 接收者:%s", nodeStore.AddressNet(fromId).B58String(), nodeStore.AddressNet(toId).B58String())
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

func NewNode(id []byte) *Node {
	nodeOne := Node{
		id:     id,
		logic:  make([][]byte, 0),
		client: make([][]byte, 0),
	}
	return &nodeOne
}
