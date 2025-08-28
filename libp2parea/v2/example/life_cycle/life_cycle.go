package main

import (
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
	n := 10
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
	time.Sleep(time.Second * 50)

	//计算理论逻辑节点，并打印
	example.BuildLogicNodes(nodes)
	//打印实际逻辑节点
	example.PrintActualNodes(nodes)
	//打印已经存在的连接
	example.PrintSessionNodes(nodes)
	//发送消息
	example.SendMsg(nodes)
	example.SendMsgHE(nodes)

	for _, one := range nodes {
		one.Log.Info().Str("发送消息完成", "----------------------------").Send()
	}

	//把节点0关闭
	nodes[0].Destroy()
	time.Sleep(time.Second * 10)
	//打印已经存在的连接
	example.PrintSessionNodes(nodes)

	//模拟节点重启
	example.Restart(nodes, 0)
	//多等待一下，让各个节点自动组网完成
	time.Sleep(time.Second * 50)
	//计算理论逻辑节点，并打印
	example.BuildLogicNodes(nodes)
	//打印实际逻辑节点
	example.PrintActualNodes(nodes)
	//打印已经存在的连接
	example.PrintSessionNodes(nodes)

}
