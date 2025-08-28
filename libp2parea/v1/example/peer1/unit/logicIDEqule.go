package unit

import (
	"encoding/json"
	"io/ioutil"

	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/utils"
)

const (
	filepath = "logicIDS.txt"
)

/*
多次启动来对比各个节点的逻辑节点差异
*/
func EquleLogicID(areas []*libp2parea.Area) {

	//查询各个节点的逻辑节点
	newNodes := make([]*Node, 0)
	for _, oneArea := range areas {
		nodeOne := Node{
			NetID: oneArea.GetNetId().B58String(),
		}
		logics := oneArea.NodeManager.GetLogicNodes()
		for _, logicOne := range logics {
			nodeOne.LogicID = append(nodeOne.LogicID, logicOne.B58String())
		}
		newNodes = append(newNodes, &nodeOne)
	}

	//检查文件是否存在
	exist, err := utils.PathExists(filepath)
	if err != nil {
		utils.Log.Error().Msgf("检查文件错误:%s", err.Error())
		return
	}

	if exist {
		bs, err := ioutil.ReadFile(filepath)
		if err != nil {
			utils.Log.Error().Msgf("读文件错误:%s", err.Error())
			return
		}
		oldNodes := make([]*Node, 0)
		err = json.Unmarshal(bs, &oldNodes)
		if err != nil {
			utils.Log.Error().Msgf("解析文件错误:%s", err.Error())
			return
		}
	} else {
		utils.SaveJsonFile(filepath, newNodes)
	}
}

func equleLogicID(oldNodes, newNodes []*Node) {
	oldNodeMap := make(map[string]*map[string]string)
	newNodeMap := make(map[string]*map[string]string)
	for _, one := range oldNodes {
		mOne := make(map[string]string)
		for _, two := range one.LogicID {
			mOne[two] = two
		}
		oldNodeMap[one.NetID] = &mOne
	}
	for _, one := range newNodes {
		mOne := make(map[string]string)
		for _, two := range one.LogicID {
			mOne[two] = two
		}
		oldNodeMap[one.NetID] = &mOne
	}
	for k, v := range oldNodeMap {
		newV := newNodeMap[k]
		if len(*v) != len(*newV) {
			utils.Log.Error().Msgf("逻辑节点不一样nodeid:%s", k)
		} else {
			for oldLogic, _ := range *v {
				_, ok := (*newV)[oldLogic]
				if !ok {
					utils.Log.Error().Msgf("逻辑节点不一样nodeid:%s", k)
				}
			}
		}
	}
}

type Node struct {
	NetID   string
	LogicID []string
}
