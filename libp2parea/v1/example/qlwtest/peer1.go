package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hyahm/golog"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/example/qlwtest/unit"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

func main() {
	utils.PprofMem(time.Minute * 2)

	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	// unit.Simulation()
	// return
	StartAllPeer()
}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	host       = "127.0.0.1"
	basePort   = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 50
	areas := make([]*libp2parea.Area, 0, n)
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)
		areas = append(areas, area.area)
		nsm.AddNodeSuperIDs(area.area.GetNetId())
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始等待节点自治")
	//等待各个节点都准备好
	for _, one := range areaPeers {
		one.area.WaitAutonomyFinish()
		one.area.WaitAutonomyFinishVnode()
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	//测试磁力节点消息到达是否准确
	SearchPeerTest(areas, nsm)

	//测试虚拟节点之间的磁力消息到达是否准确
	// SearchVnodeTest(areas)

	//--------------------------
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始发送消息")
	//异步发消息
	group := new(sync.WaitGroup)
	group.Add(len(areaPeers))
	// <-time.NewTimer(time.Second * 5).C
	for i, _ := range areaPeers {
		one := areaPeers[i]
		// one.sendMsgLoop(one.area, areaPeers, group)

		// 测试带有代理节点的消息发送
		one.sendMsgProxyLoop(one.area, areaPeers, group)
	}
	group.Wait()
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("发送消息完成")

	//测试给未知节点发送消息
	randID := sha256.Sum256([]byte("randNetId"))
	randNetId := virtual_node.AddressNetExtend(randID[:])
	areaPeers[0].area.SendVnodeP2pMsgHE(9999999, &(areaPeers[0].area.Vm.GetVnodeDiscover().Vnode.Vid), &randNetId, nil, nil)

	randNetIdB := nodeStore.AddressNet([]byte("ABC"))
	areaPeers[0].area.SendSearchSuperMsg(9999999, &randNetIdB, nil)

	// addrs := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&areaPeers[0].area.Vm.GetVnodeDiscover().Vnode.Vid)
	// for i, _ := range addrs {
	// 	logicNetid, _ := areaPeers[0].area.SearchVnodeId(addrs[i])
	// 	utils.Log.Info().Msgf("查找的逻辑节点地址:%s", logicNetid.B58String())
	// }
	// utils.Log.Info().Msgf("--------------------------")
	// utils.Log.Info().Msgf("查找逻辑节点地址完成")

	//-------------------
	// utils.Log.Info().Msgf("--------------------------")
	// utils.Log.Info().Msgf("下线一个节点")
	// areaPeers[len(areaPeers)-1].area.Destroy()
	// areaPeers = areaPeers[:len(areaPeers)-1]

	//循环检查网络状态
	// DisconnectionReconnection(areaPeers)

	for i := 0; i < 1; i++ {
		time.Sleep(time.Second * 2)
		countNum := 1
		for k, v := range MsgCount {
			utils.Log.Info().Msgf("%d节点：%s 收到:%d", countNum, k, v)
			countNum++
		}
		utils.Log.Info().Msgf("-----------------------")
	}

	select {}

}

func SearchPeerTest(areas []*libp2parea.Area, nsm nodeStore.NodeSimulationManager) {
	utils.Log.Info().Msgf("Simulation start")
	//构建标准的逻辑节点
	nsm.BuildNodeLogicIDs()
	//模拟节点发现过程
	nsm.Discover()
	//打印各个自定义节点的逻辑节点
	nsm.PrintlnLogicNodesNew(true)
	//发送P2P消息
	// nsm.MsgPingNodesP2P()
	utils.Log.Info().Msgf("---------------------------")
	//发送搜索磁力节点消息
	// nsm.MsgPingNodesSearch()

	//对比标准节点和自定义节点，各方保存的逻辑节点差异
	nodeStore.EqualNodes(&nsm)

	utils.Log.Info().Msgf("Simulation end")

	// nsm.BuildNode(n, 0)
	// nsm.BuildNodeLogicIDs()
	// nsm.Discover()
	// nsm.PrintlnLogicNodesNew()
	// // nsm.MsgPingNodesP2P()
	// nsm.MsgPingNodesSearch()

	//对比仿真的逻辑节点和真实的逻辑节点是否相同
	utils.Log.Info().Msgf("对比仿真逻辑节点和真实逻辑节点是否相同")
	nodeMap := nsm.GetCustomNodes()
	for _, areaOne := range areas {
		ns, ok := nodeMap[utils.Bytes2string(areaOne.GetNetId())]
		if !ok {
			utils.Log.Info().Msgf("未找到节点")
			panic("未找到节点！")
		}
		logicNodes := areaOne.NodeManager.GetLogicNodes()
		ids := make([][]byte, 0, len(logicNodes))
		for _, one := range logicNodes {
			ids = append(ids, one)
		}
		isChange, _ := nodeStore.EqualIds(ids, ns.Logic)
		if isChange {
			utils.Log.Info().Msgf("--------------------------------------")
			utils.Log.Info().Msgf("逻辑节点不相等:%s", areaOne.GetNetId().B58String())
			for _, one := range ns.Logic {
				utils.Log.Info().Msgf("仿真逻辑节点:%s", nodeStore.AddressNet(one).B58String())
			}
			utils.Log.Info().Msgf("")
			for _, one := range logicNodes {
				utils.Log.Info().Msgf("真实逻辑节点:%s", one.B58String())
			}
		}
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("对比逻辑节点结束")

	//多次查询逻辑节点并对比结果是否正确
	// time.Sleep(time.Second * 30)
	unit.LoopSendSearch(areas)

	// for _, one := range areaPeers {
	// 	utils.Log.Info().Msgf("==本节点地址:%s", one.area.GetNetId().B58String())
	// 	for _, one := range one.area.NodeManager.GetLogicNodes() {
	// 		utils.Log.Info().Msgf("--逻辑节点:%s", one.B58String())
	// 	}
	// }

	// PrintLogicID(areaPeers)
	// CheckLogicID(areaPeers)
}

func SearchVnodeTest(areas []*libp2parea.Area) {

	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}
	for _, area := range areas {
		vnodeinfos := area.Vm.GetVnodeSelf()
		for _, vnode := range vnodeinfos {
			nsm.AddNodeSuperIDs(vnode.Vid)
		}
	}

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("虚拟节点：打印每个节点的虚拟节点地址")
	for _, area := range areas {
		vnodeinfos := area.Vm.GetVnodeSelf()
		for _, vnode := range vnodeinfos {
			utils.Log.Info().Msgf("虚拟地址:%s %d", vnode.Vid.B58String(), vnode.Index)
		}
		utils.Log.Info().Msgf("--")
	}

	nsm.BuildNodeLogicIDs()

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("虚拟节点：打印标准逻辑节点")
	nsm.PrintlnStandardLogicNodesNew(true)

	//多次启动来对比各个节点的逻辑节点差异
	// unit.EquleLogicID(areas)
	for _, area := range areas {
		vnodes := area.Vm.GetVnodeSelf()
		for _, one := range vnodes {
			utils.Log.Info().Msgf("本虚拟节点地址:%s", one.Vid.B58String())
			vnode := area.Vm.FindVnodeInSelf(one.Vid)
			vnode.LogicalNode.Range(func(k, v interface{}) bool {
				vnodeInfo := v.(virtual_node.Vnodeinfo)
				utils.Log.Info().Msgf("  逻辑虚拟节点:%s", vnodeInfo.Vid.B58String())
				return true
			})
		}
	}

	//对比仿真的逻辑节点和真实的逻辑节点是否相同
	utils.Log.Info().Msgf("对比仿真逻辑虚拟节点和真实逻辑虚拟节点是否相同")
	nodeMap := nsm.GetStandardNodes()
	for _, areaOne := range areas {

		vnodeinfos := areaOne.Vm.GetVnodeSelf()
		for _, vnodeinfo := range vnodeinfos {

			ns, ok := nodeMap[utils.Bytes2string(vnodeinfo.Vid)]
			if !ok {
				utils.Log.Info().Msgf("未找到节点")
				panic("未找到节点！")
			}
			vnode := areaOne.Vm.FindVnodeInSelf(vnodeinfo.Vid)

			logicNodes := make([]virtual_node.AddressNetExtend, 0) // vnode.lo areaOne.NodeManager.GetLogicNodes()
			vnode.LogicalNode.Range(func(k, v interface{}) bool {
				vnodeInfo := v.(virtual_node.Vnodeinfo)
				logicNodes = append(logicNodes, vnodeInfo.Vid)
				return true
			})
			ids := make([][]byte, 0, len(logicNodes))
			for _, one := range logicNodes {
				ids = append(ids, one)
			}
			isChange, _ := nodeStore.EqualIds(ids, ns.Logic)
			if isChange {
				utils.Log.Info().Msgf("--------------------------------------")
				utils.Log.Info().Msgf("逻辑虚拟节点不相等:%s", areaOne.GetNetId().B58String())
				for _, one := range ns.Logic {
					utils.Log.Info().Msgf("仿真逻辑虚拟节点:%s", nodeStore.AddressNet(one).B58String())
				}
				utils.Log.Info().Msgf("")
				for _, one := range logicNodes {
					utils.Log.Info().Msgf("真实逻辑虚拟节点:%s", one.B58String())
				}
			}
		}
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("对比逻辑节点结束")

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始发送虚拟节点磁力消息")
	//多次查询磁力虚拟节点，并对比结果是否正确
	unit.LoopSendSearchVnode(areas)
}

type TestPeer struct {
	area *libp2parea.Area
}

func StartOnePeer(i int) *TestPeer {
	keyPath1 := filepath.Join("conf", "keystore"+strconv.Itoa(i)+".key")

	key1 := keystore.NewKeystore(keyPath1, addrPre)
	err := key1.Load()
	if err != nil {
		//没有就创建
		err = key1.CreateNewKeystore(keyPwd)
		if err != nil {
			panic("创建key1错误:" + err.Error())
		}
	}

	if key1.NetAddr == nil {
		_, _, err = key1.CreateNetAddr(keyPwd, keyPwd)
		if err != nil {
			panic("创建NetAddr错误:" + err.Error())
		}
	}
	if len(key1.GetAddr()) < 1 {
		_, err = key1.GetNewAddr(keyPwd, keyPwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}
	if len(key1.GetDHKeyPair().SubKey) < 1 {
		_, err = key1.GetNewDHKey(keyPwd, keyPwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}

	area, err := libp2parea.NewArea(areaName, key1, keyPwd)
	if err != nil {
		panic(err.Error())
	}
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	area.SetNetTypeToTest()

	// area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))
	// area.CloseVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	area.Vm.AddVnodeByIndex(100)
	area.Vm.AddVnodeByIndex(200)
	area.Vm.AddVnodeByIndex(300)
	area.Vm.AddVnodeByIndex(400)

	// vnodeinfo := area.Vm.AddVnodeByIndex(100)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(200)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(300)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(400)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

var rand = utils.GetHashForDomain(utils.GetRandomDomain())
var magneticIds []nodeStore.AddressNet
var magneticVnodeIds []virtual_node.AddressNetExtend

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	//发送p2p消息
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}

	//判断各个节点查找的磁力节点地址是否一样。
	// randId := nodeStore.AddressNet(rand)
	// utils.Log.Info().Msgf("randid:%s", randId.B58String())
	// magneticids := nodeStore.GetQuarterLogicAddrNetByAddrNet(&randId)
	// for i, one := range magneticids {
	// 	utils.Log.Info().Msgf("磁力id:%s", one.B58String())
	// 	magnetic, err := area.SearchNetAddr(one)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("发送SearchNetAddr消息失败:%s", err.Error())
	// 		continue
	// 	}
	// 	if len(magneticIds) < 4 {
	// 		magneticIds = append(magneticIds, *magnetic)
	// 	} else {
	// 		if !bytes.Equal(magneticIds[i], *magnetic) {
	// 			utils.Log.Info().Msgf("磁力节点不相等 %d:%s", i, magnetic.B58String())
	// 		} else {
	// 			utils.Log.Info().Msgf("磁力节点 %d:%s", i, magnetic.B58String())
	// 		}
	// 	}
	// }

	// fromNetid := area.GetNetId()
	// logicids := nodeStore.GetQuarterLogicAddrNetByAddrNet(&fromNetid)
	// for _, one := range logicids {
	// 	_, err := area.SendSearchSuperMsgWaitRequest(msg_id_searchSuper, one, nil, time.Second*10)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("发送searchSuper消息失败:%s", err.Error())
	// 		return err
	// 	}
	// }

	//发送虚拟节点p2p消息
	// toVnodes := toArea.Vm.GetVnodeSelf()
	// for _, one := range toVnodes {
	// 	utils.Log.Info().Msgf("发送虚拟节点id:%s to:%s", area.Vm.GetVnodeDiscover().Vnode.Vid.B58String(), one.Vid.B58String())
	// 	_, err := area.SendVnodeP2pMsgHEWaitRequest(msg_id_vnode_p2p, &area.Vm.GetVnodeDiscover().Vnode.Vid, &one.Vid, nil, time.Second*2)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("发送VnodeP2p消息失败:%s", err.Error())
	// 		return err
	// 	}
	// }

	// //发送搜索虚拟节点消息
	// randVnodeId := virtual_node.AddressNetExtend(rand)
	// addrs := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&randVnodeId)
	// for i, two := range addrs {
	// 	// utils.Log.Info().Msgf("寻找磁力虚拟节点地址self:%s find:%s", area.GetNetId().B58String(), two.B58String())
	// 	magneticVnodeid, err := area.SearchNetAddrVnode(two)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("发送SearchNetAddr消息失败:%s", err.Error())
	// 		continue
	// 	}
	// 	utils.Log.Info().Msgf("磁力虚拟节点地址 %d:%s", i, magneticVnodeid.B58String())
	// 	if len(magneticVnodeIds) < 4 {
	// 		magneticVnodeIds = append(magneticVnodeIds, *magneticVnodeid)
	// 	} else {
	// 		if !bytes.Equal(magneticVnodeIds[i], *magneticVnodeid) {
	// 			utils.Log.Error().Msgf("磁力虚拟节点不相等")
	// 		}
	// 	}
	// }

	// addrs := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&area.Vm.GetVnodeDiscover().Vnode.Vid)
	// for _, one := range addrs {
	// 	utils.Log.Info().Msgf("===--------- start")
	// 	// utils.Log.Info().Msgf("先打印逻辑节点")
	// 	// logicNodes := area.Vm.FindLogicInVnodeSelf(area.Vm.GetVnodeSelf()[0].Vid)
	// 	// for _, one := range logicNodes {
	// 	// 	utils.Log.Info().Msgf("逻辑节点:%s %d %s", one.Nid.B58String(), one.Index, one.Vid.B58String())
	// 	// }
	// 	_, err := area.SendVnodeSearchMsgWaitRequest(msg_id_vnode_search, &area.Vm.GetVnodeSelf()[0].Vid, one, nil, time.Second*10)
	// 	if err != nil {
	// 		utils.Log.Info().Msgf("发送VnodeSearchMsg error:%s", err.Error())
	// 		return err
	// 	}
	// 	utils.Log.Info().Msgf("===--------- end")
	// }

	//发送广播消息
	// utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	// err = area.SendMulticastMsg(msg_id_multicast, nil)
	// if err != nil {
	// 	utils.Log.Error().Msgf("发送广播消息失败:%s", err.Error())
	// }

	return nil
}

// 发送代理消息
func sendMsgOneProxy(area *libp2parea.Area, proxyNetid nodeStore.AddressNet, toArea *libp2parea.Area) error {
	//发送p2p消息
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s, 代理节点:%s -----begin", area.GetNetId().B58String(), toArea.GetNetId().B58String(), proxyNetid.B58String())
	toNetid := toArea.GetNetId()
	// sendMsgProxyMsg(area, &toNetid, &proxyNetid, 1) // p2p代理消息测试
	// sendMsgProxyMsg(area, &toNetid, &proxyNetid, 2) // p2p search代理消息测试
	sendMsgProxyMsg(area, &toNetid, &proxyNetid, 3) // p2pHE代理消息测试

	utils.Log.Info().Msgf("节点:%s 发送消息给:%s, 代理节点:%s -----end", area.GetNetId().B58String(), toArea.GetNetId().B58String(), proxyNetid.B58String())
	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("")

	return nil
}

func sendMsgProxyMsg(area *libp2parea.Area, toNetid, proxyNetid *nodeStore.AddressNet, msgType int) error {
	if msgType == 1 {
		utils.Log.Info().Msgf("qlw---没有代理节点的发送-------")
		_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, toNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p消息失败1:%s", err.Error())
			return err
		}
		utils.Log.Info().Msgf("qlw----分割线-------")
		utils.Log.Info().Msgf("qlw---：含有代理节点的发送-------")
		_, _, _, err = area.SendP2pMsgProxyWaitRequest(msg_id_p2p, toNetid, nil, proxyNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p消息失败2:%s", err.Error())
			return err
		}
	} else if msgType == 2 {
		utils.Log.Info().Msgf("qlw---没有代理节点的发送-------")
		_, err := area.SendSearchSuperMsgWaitRequest(msg_id_p2p, toNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p Search消息失败1:%s", err.Error())
			return err
		}
		utils.Log.Info().Msgf("qlw----分割线-------")
		utils.Log.Info().Msgf("qlw---：含有代理节点的发送-------")
		_, err = area.SendSearchSuperMsgProxyWaitRequest(msg_id_p2p, toNetid, nil, proxyNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p Search消息失败2:%s", err.Error())
			return err
		}
	} else if msgType == 3 {
		utils.Log.Info().Msgf("qlw---没有代理节点的发送-------")
		_, _, _, err := area.SendP2pMsgHEWaitRequest(msg_id_p2p, toNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p加密消息失败1:%s", err.Error())
			return err
		}
		utils.Log.Info().Msgf("qlw----分割线-------")
		utils.Log.Info().Msgf("qlw---：含有代理节点的发送-------")
		_, _, _, err = area.SendP2pMsgHEProxyWaitRequest(msg_id_p2p, toNetid, nil, proxyNetid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p加密消息失败:%s", err.Error())
			return err
		}
	}

	return nil
}

func (this *TestPeer) sendMsgLoop(area *libp2parea.Area, toAddrs []*TestPeer, group *sync.WaitGroup) {
	utils.Log.Info().Msgf("开始一个节点的发送")
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.area.GetNetId()) {
			continue
		}
		done := false
		for !done {
			err := sendMsgOne(area, one.area)
			if err != nil {
				//当网络发送失败，等待1秒钟后，继续发送
				time.Sleep(time.Second)
			} else {
				done = true
				time.Sleep(time.Second / 100)
			}
		}
	}
	group.Done()
	utils.Log.Info().Msgf("发送完一个节点")
}

// 测试带有代理节点的消息发送是否正确
func (this *TestPeer) sendMsgProxyLoop(area *libp2parea.Area, toAddrs []*TestPeer, group *sync.WaitGroup) {
	utils.Log.Info().Msgf("开始一个节点的发送")
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.area.GetNetId()) {
			continue
		}
		done := false
		for !done {
			logicNodes := area.NodeManager.GetLogicNodes()
			var proxyNetid = logicNodes[0]
			err := sendMsgOneProxy(area, proxyNetid, one.area)
			if err != nil {
				time.Sleep(time.Second)
			} else {
				done = true
				time.Sleep(time.Second / 100)
			}
		}
	}
	group.Done()
	utils.Log.Info().Msgf("发送完一个节点")
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_searchSuper = 1003
const msg_id_searchSuper_recv = 1004 //加密消息

const msg_id_vnode_p2p = 1005
const msg_id_vnode_p2p_recv = 1006 //加密消息

const msg_id_vnode_search = 1007      //搜索节点消息
const msg_id_vnode_search_recv = 1008 //搜索节点消息 返回

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.SearchSuperHandler)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)

	area.Register_vnode_p2pHE(msg_id_vnode_p2p, this.RecvMsgHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_p2p_recv, this.RecvMsgHEHandler)

	area.Register_vnode_search(msg_id_vnode_search, this.SearchVnodeHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_search_recv, this.SearchVnodeHandler_recv)

	area.Register_multicast(msg_id_multicast, this.MulticastMsgHandler)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("收到错误p2p消息self:%s recv:%s", this.area.GetNetId().B58String(), message.Head.RecvId.B58String())
		return
	}
	// utils.Log.Info().Msgf("收到P2P消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, nil)
}

func (this *TestPeer) RecvP2PMsgHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}

func (this *TestPeer) SearchSuperHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	// utils.Log.Info().Msgf("收到SearchSuper消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	err := this.area.SendSearchSuperReplyMsg(message, msg_id_searchSuper_recv, nil)
	if err != nil {
		utils.Log.Error().Msgf("SendSearchSuperReplyMsg error:%s", err.Error())
	}
}

func (this *TestPeer) SearchSuperHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}

func (this *TestPeer) RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// selfId := message.Head.RecvId.B58String()
	// vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode)
	if this.area.Vm.FindInVnodeSelf(*message.Head.RecvVnode) {
		utils.Log.Info().Msgf("收到vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
	} else {
		utils.Log.Error().Msgf("收到错误vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
		return
	}
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
	err := this.area.SendVnodeP2pReplyMsgHE(message, msg_id_vnode_p2p_recv, nil)
	if err != nil {
		utils.Log.Info().Msgf("SendVnodeP2pReplyMsgHE error:%s", err.Error())
	} else {

	}
}

func (this *TestPeer) RecvMsgHEHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到vnode p2p消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}

func (this *TestPeer) SearchVnodeHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// selfId := message.Head.RecvId.B58String()
	selfId := message.Head.RecvVnode.B58String()
	// utils.Log.Info().Msgf("收到 VnodeSearch 消息 self:%s from:%s", this.area.GetNetId().B58String(), message.Head.Sender.B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	err := this.area.SendVnodeSearchReplyMsg(message, msg_id_vnode_search_recv, nil)
	if err != nil {
		utils.Log.Info().Msgf("回复消息错误:%s", err.Error())
	}
}

func (this *TestPeer) SearchVnodeHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到 VnodeSearch 消息返回 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}

func (this *TestPeer) MulticastMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {

}

/*
打印逻辑节点
*/
func PrintLogicID(peers []*TestPeer) {

	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")
	golog.Infof("节点自治完成，打印逻辑节点\n")

	allNid := make([]nodeStore.AddressNet, 0)
	allVid := make([]virtual_node.AddressNetExtend, 0)
	for _, one := range peers {
		allNid = append(allNid, one.area.GetNetId())
		vnodeSelf := one.area.Vm.GetVnodeSelf()
		for _, two := range vnodeSelf {
			allVid = append(allVid, two.Vid)
		}
	}

	//----------------------------------------
	//检查节点的逻辑节点
	for n := 0; n < len(allNid); n++ {
		utils.Log.Info().Msgf("本节点为 %s", allNid[n].B58String())
		golog.Infof("本节点为 %s\n", allNid[n].B58String())
		index := n

		idMH := allNid[index]
		idsm := nodeStore.NewIds(idMH, 256)
		for i, one := range allNid {
			if i == index {
				continue
			}
			idsm.AddId(one)
		}

		var cutsomArea *libp2parea.Area
		for i, one := range peers {
			if bytes.Equal(one.area.GetNetId(), idMH) {
				cutsomArea = peers[i].area
				break
			}
		}

		logicIDSMap := make(map[string]int)
		logicIDS := cutsomArea.GetNodeManager().GetLogicNodes()
		for _, one := range logicIDS {
			utils.Log.Info().Msgf("--真实逻辑节点 %s", one.B58String())
			golog.Infof("--真实逻辑节点 %s\n", one.B58String())
			logicIDSMap[utils.Bytes2string(one)] = 0
		}

		// standardLogicIDS := idsm.GetIds()
		// for _, one := range standardLogicIDS {
		// 	// utils.Log.Info().Msgf("--逻辑节点 %s", nodeStore.AddressNet(one).B58String())
		// 	_, ok := logicIDSMap[utils.Bytes2string(one)]
		// 	if !ok {
		// 		utils.Log.Error().Msgf("这个逻辑节点未找到:%s", nodeStore.AddressNet(one).B58String())
		// 		return
		// 	}
		// }
	}
	//----------------------------------------
	//检查虚拟节点的逻辑节点
	for n := 0; n < len(allVid); n++ {
		utils.Log.Info().Msgf("本虚拟节点为 %s", allVid[n].B58String())
		golog.Infof("本虚拟节点为 %s\n", allVid[n].B58String())
		index := n
		idMH := allVid[index]
		idsm := nodeStore.NewIds(idMH, 256)
		for i, one := range allVid {
			if i == index {
				continue
			}
			idsm.AddId(one)
		}

		var vnode *virtual_node.Vnode
		for _, one := range peers {
			vnode = one.area.Vm.FindVnodeInSelf(idMH)
			if vnode != nil {
				break
			}
		}

		logicIDSMap := make(map[string]int)
		vnodeinfos := vnode.GetVnodeinfoAllNotSelf() // cutsomArea.GetNodeManager().GetLogicNodes()
		for _, one := range vnodeinfos {
			utils.Log.Info().Msgf("--真实逻辑节点 %s", one.Vid.B58String())
			golog.Infof("--真实逻辑节点 %s\n", one.Vid.B58String())
			logicIDSMap[utils.Bytes2string(one.Vid)] = 0
		}

		// for _, one := range peers {
		// 	if !one.area.Vm.FindInVnodeSelf(idMH) {
		// 		continue
		// 	}
		// 	for _, two := range one.area.Vm.GetVnodeSelf() {
		// 		logicIDSMap[utils.Bytes2string(two.Vid)] = 0
		// 	}
		// 	break
		// }

		// standardLogicIDS := idsm.GetIds()
		// for _, one := range standardLogicIDS {
		// 	// utils.Log.Info().Msgf("--逻辑节点 %s", virtual_node.AddressNetExtend(one).B58String())
		// 	_, ok := logicIDSMap[utils.Bytes2string(one)]
		// 	if !ok {
		// 		utils.Log.Error().Msgf("这个虚拟逻辑节点未找到:%s self:%s", virtual_node.AddressNetExtend(one).B58String(), allVid[n].B58String())
		// 		// return
		// 	}
		// }
	}
}

/*
检查逻辑节点
*/
func CheckLogicID(peers []*TestPeer) {
	allNid := make([]nodeStore.AddressNet, 0)
	allVid := make([]virtual_node.AddressNetExtend, 0)
	for _, one := range peers {
		allNid = append(allNid, one.area.GetNetId())
		vnodeSelf := one.area.Vm.GetVnodeSelf()
		for _, two := range vnodeSelf {
			allVid = append(allVid, two.Vid)
		}
	}

	//----------------------------------------
	//检查节点的逻辑节点
	for n := 0; n < len(allNid); n++ {
		utils.Log.Info().Msgf("本节点为 %s", allNid[n].B58String())
		index := n

		idMH := allNid[index]
		idsm := nodeStore.NewIds(idMH, 256)
		for i, one := range allNid {
			if i == index {
				continue
			}
			idsm.AddId(one)
		}

		var cutsomArea *libp2parea.Area
		for i, one := range peers {
			if bytes.Equal(one.area.GetNetId(), idMH) {
				cutsomArea = peers[i].area
				break
			}
		}

		logicIDSMap := make(map[string]int)
		logicIDS := cutsomArea.GetNodeManager().GetLogicNodes()
		for _, one := range logicIDS {
			utils.Log.Info().Msgf("--真实逻辑节点 %s", one.B58String())
			logicIDSMap[utils.Bytes2string(one)] = 0
		}

		standardLogicIDS := idsm.GetIds()
		for _, one := range standardLogicIDS {
			utils.Log.Info().Msgf("--逻辑节点 %s", nodeStore.AddressNet(one).B58String())
			_, ok := logicIDSMap[utils.Bytes2string(one)]
			if !ok {
				utils.Log.Error().Msgf("这个逻辑节点未找到self:%s nid:%s", idMH.B58String(), nodeStore.AddressNet(one).B58String())
				// return
			}
		}
	}
	//----------------------------------------
	//检查虚拟节点的逻辑节点
	for n := 0; n < len(allVid); n++ {
		// utils.Log.Info().Msgf("本虚拟节点为 %s", allVid[n].B58String())
		index := n
		idMH := allVid[index]
		idsm := nodeStore.NewIds(idMH, 256)
		for i, one := range allVid {
			if i == index {
				continue
			}
			idsm.AddId(one)
		}

		var vnode *virtual_node.Vnode
		for _, one := range peers {
			vnode = one.area.Vm.FindVnodeInSelf(idMH)
			if vnode != nil {
				break
			}
		}

		logicIDSMap := make(map[string]int)
		vnodeinfos := vnode.GetVnodeinfoAllNotSelf() // cutsomArea.GetNodeManager().GetLogicNodes()
		for _, one := range vnodeinfos {
			// utils.Log.Info().Msgf("--真实逻辑节点 %s", one.Vid.B58String())
			logicIDSMap[utils.Bytes2string(one.Vid)] = 0
		}

		for _, one := range peers {
			if !one.area.Vm.FindInVnodeSelf(idMH) {
				continue
			}
			for _, two := range one.area.Vm.GetVnodeSelf() {
				logicIDSMap[utils.Bytes2string(two.Vid)] = 0
			}
			break
		}

		standardLogicIDS := idsm.GetIds()
		for _, one := range standardLogicIDS {
			// utils.Log.Info().Msgf("--逻辑节点 %s", virtual_node.AddressNetExtend(one).B58String())
			_, ok := logicIDSMap[utils.Bytes2string(one)]
			if !ok {
				utils.Log.Error().Msgf("这个虚拟逻辑节点未找到:%s self:%s", virtual_node.AddressNetExtend(one).B58String(), allVid[n].B58String())
				// return
			}
		}
	}
}

/*
测试断线重连
*/
func DisconnectionReconnection(areaPeers []*TestPeer) {

	//循环检查网络状态
	online := true
	reconn := false
	for {
		tempArea := areaPeers[0].area
		time.Sleep(time.Second * 5)
		online = tempArea.CheckOnline()
		utils.Log.Info().Msgf("网络状态:%t", online)
		if !online {
			reConnOk := tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()
			reConnOk = tempArea.ReconnectNet()

			utils.Log.Info().Msgf("重连:%t", reConnOk)
			reconn = true
			for {
				time.Sleep(time.Second * 1)
				tempArea.WaitAutonomyFinish()
				if reconn && tempArea.CheckOnline() {
					// sendMsgOne(tempArea, nil)
					break
				}
			}
		} else {
			reconn = false
		}
	}

}
