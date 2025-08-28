package nodeStore

import (
	"testing"
	"web3_gui/utils"
)

func TestSimulation(t *testing.T) {
	Simulation()
}

func Simulation() *NodeSimulationManager {
	utils.Log.Info().Msgf("Simulation start")
	nsm := NodeSimulationManager{IDdepth: 32 * 8} //
	nsm.BuildNode(40, 0)
	// utils.Log.Info().Msgf("1111111111")
	//构建标准的逻辑节点
	nsm.BuildNodeLogicIDs()
	// utils.Log.Info().Msgf("1111111111")
	//模拟节点发现过程
	nsm.Discover()
	// utils.Log.Info().Msgf("1111111111")
	//打印各个自定义节点的逻辑节点
	nsm.PrintlnLogicNodesNew(false)
	//发送P2P消息
	// nsm.MsgPingNodesP2P()
	utils.Log.Info().Msgf("---------------------------")
	//发送搜索磁力节点消息
	// nsm.MsgPingNodesSearch()

	//对比标准节点和自定义节点，各方保存的逻辑节点差异
	EqualNodes(&nsm)

	utils.Log.Info().Msgf("Simulation end")
	return &nsm

}
