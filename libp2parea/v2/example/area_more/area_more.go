package main

import (
	"crypto/sha256"
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
	n := 100
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		utils.Log.Info().Int("正在启动节点", i).Send()
		wg.Add(1)
		areaName := sha256.Sum256([]byte(example.AreaNameStr + strconv.Itoa(i)))
		if i == 0 {
			//utils.Log.Info().Str("开始等待", "1111").Send()
			//等待首节点启动完成，再启动其他节点
			node, ERR := example.StartNodeOne(areaName, example.AddrPre, i, wg)
			if ERR.CheckFail() {
				return
			}
			//utils.Log.Info().Str("开始等待", "1111").Send()
			node.WaitAutonomyFinish()
			//utils.Log.Info().Str("开始等待", "1111").Send()
			nodes = append(nodes, node)
			continue
		}
		area, ERR := example.StartNodeOne(areaName, example.AddrPre, i, wg)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			return
		}
		nodes = append(nodes, area)
	}
	wg.Wait()
	//utils.Log.Info().Str("开始等待", "1111").Send()

	//多等待一下，让各个节点自动组网完成
	time.Sleep(time.Second * 50)

	//计算理论逻辑域名称，并打印
	example.PrintLogicAreaName(nodes)
	//打印实际逻辑节点
	example.PrintActualAreaname(nodes)
	//打印各个节点的所有域节点连接
	example.PrintSessionNodeAreaname(nodes)
	//
	//select {}
}
