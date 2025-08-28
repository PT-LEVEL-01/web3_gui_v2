package main

import (
	"bytes"
	"strconv"
	"sync"
	"time"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/example"
	"web3_gui/utils"
)

/*
本示例测试单机上，多节点自动组网后，各个节点的逻辑连接是否正确
直接运行看日志结果
对比理论逻辑节点、实际逻辑节点、会话列表是否一致。
*/
func main() {
	nodes := make([]*libp2parea.Node, 0)
	n := 2
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		wg.Add(1)
		if i == 0 {
			//utils.Log.Info().Str("开始等待", "1111").Send()
			//等待首节点启动完成，再其他其他节点
			node, ERR := example.StartNodeOne(example.AreaName, example.AddrPre, i, wg)
			if ERR.CheckFail() {
				return
			}

			example.RegisterMsg(node)

			//utils.Log.Info().Str("开始等待", "1111").Send()
			node.WaitAutonomyFinish()
			//utils.Log.Info().Str("开始等待", "1111").Send()
			nodes = append(nodes, node)
			continue
		}
		node, ERR := example.StartNodeOne(example.AreaName, example.AddrPre, i, wg)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			return
		}
		example.RegisterMsg(node)
		nodes = append(nodes, node)
	}
	wg.Wait()
	//utils.Log.Info().Str("开始等待", "1111").Send()

	//多等待一下，让各个节点自动组网完成
	time.Sleep(time.Second * 5)

	//计算理论逻辑节点，并打印
	example.BuildLogicNodes(nodes)
	//打印实际逻辑节点
	example.PrintActualNodes(nodes)

	example.PrintSessionNodes(nodes)
	//

	example.SendMsg(nodes)
	//example.SendMsgHE(nodes)
	//example.SendMsgHE(nodes)

	//发送个加密消息
	nodes[0].Log.Info().Str("-------start----------", "").Send()
	ERR := example.HelloHE_o(nodes[0], nodes[1].GetNetId())
	if ERR.CheckFail() {
		nodes[0].Log.Error().Str("发送加密消息失败", ERR.String()).Send()
	}
	nodes[0].Log.Info().Str("-------end----------", "").Send()

	//发送个加密消息
	nodes[1].Log.Info().Str("-------start----------", "").Send()
	ERR = example.HelloHE_o(nodes[1], nodes[0].GetNetId())
	if ERR.CheckFail() {
		nodes[1].Log.Error().Str("发送加密消息失败", ERR.String()).Send()
	}
	nodes[1].Log.Info().Str("-------end----------", "").Send()

}

/*
发送加密消息
*/
func SendMsgHE(sender *libp2parea.Node, nodes []*libp2parea.Node) {
	sender.Log.Info().Str("开始发送加密消息", "------------------------------ "+strconv.Itoa(len(nodes))).Send()
	//one.Log.Info().Int("开始打印各个会话节点", len(sss)).Send()
	for _, nodeRecv := range nodes {
		if bytes.Equal(sender.GetNetId().GetAddr(), nodeRecv.GetNetId().GetAddr()) {
			continue
		}
		//发送个加密消息
		ERR := example.HelloHE_o(sender, nodeRecv.GetNetId())
		if ERR.CheckFail() {
			sender.Log.Error().Str("发送加密消息失败", ERR.String()).Send()
		}
	}
}
