package main

import (
	"crypto/sha256"
	"math/big"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/utils"
)

func main() {
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")
	Startzyz()
}

var sstt int
var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("zyz111"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	host       = "127.0.0.1"
	basePort   = 19960
)

/*
启动所有节点
*/
func Startzyz() {
	n := 7

	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)

		area.area.OpenVnode()
		area.area.Vm.AddVnode()
		area.area.Vm.AddVnode()
		area.area.Vm.AddVnode()

		vnodeinfos := area.area.Vm.GetVnodeSelf()
		for _, vnode := range vnodeinfos {
			utils.Log.Info().Msgf("虚拟地址:%s %d", vnode.Vid.B58String(), vnode.Index)
		}
		time.Sleep(time.Second)
	}

	kad := new(nodeStore.IdDESC)

	for _, v := range areaPeers {
		utils.Log.Info().Msgf("NetId: %s VnodeId : %s",
			v.area.GetNetId().B58String(),
			v.area.GetVnodeId().B58String())
		*kad = append(*kad, new(big.Int).SetBytes([]byte(v.area.GetNetId())))
	}
	var small string
	var snum int
	var lnum int
	var large string

	sort.Sort(kad)
	for i := 0; i < len(*kad); i++ {

		// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
		one := (*kad)[i]
		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		utils.Log.Info().Msgf("排序 %d ： %s \n", i, nodeStore.AddressNet(*IdBsP).B58String())
		if i == 0 {
			large = nodeStore.AddressNet(*IdBsP).B58String()
		}
		if i == len(*kad)-1 {
			small = nodeStore.AddressNet(*IdBsP).B58String()
		}
	}

	// for i := 0; i < len(areaPeers); i++ {
	// 	if areaPeers[i].area.GetNetId().B58String() == small {
	// 		snum = i
	// 		utils.Log.Info().Msgf("ssss %s", areaPeers[i].area.GetNetId().B58String())
	// 	}

	// 	if areaPeers[i].area.GetNetId().B58String() == large {
	// 		lnum = i
	// 		utils.Log.Info().Msgf("llll %s", areaPeers[i].area.GetNetId().B58String())
	// 	}
	// 	// if areaPeers[i].area.GetNetId().B58String() == "BYito3Uzea3wHeHmHpi9ckE98AtVQVXFWxSG6cAmXxQA" {
	// 	// 	sstt = i
	// 	// }

	// 	// utils.Log.Info().Msgf("地址:%s  虚拟节点 : %s", areaPeers[i].area.GetNetId().B58String(), areaPeers[i].area.Vm.GetVnodeSelf()[1].Vid.B58String())
	// 	logicNodes := areaPeers[i].area.Vm.FindLogicInVnodeSelf(areaPeers[i].area.Vm.GetVnodeSelf()[0].Vid)

	// 	for _, one := range logicNodes {
	// 		utils.Log.Info().Msgf("逻辑节点:%s %d %s", one.Nid.B58String(), one.Index, one.Vid.B58String())
	// 	}
	// }

	utils.Log.Info().Msgf("--------------------------")
	kadV := new(nodeStore.IdDESC)

	for _, v := range areaPeers {
		//*kadV = append(*kadV, new(big.Int).SetBytes([]byte(v.Vm.DiscoverVnodes.Vnode.Vid)))
		for _, vInfo := range v.area.Vm.VnodeMap {
			if vInfo.Vnode.Index == 0 {
				continue
			}
			utils.Log.Info().Msgf("Vnode.Vid : %s", vInfo.Vnode.Vid.B58String())
			*kadV = append(*kadV, new(big.Int).SetBytes([]byte(vInfo.Vnode.Vid)))
		}
	}

	sort.Sort(kadV)
	for i := 0; i < len(*kadV); i++ {
		// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
		one := (*kadV)[i]
		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		utils.Log.Info().Msgf("Vnode 排序 %d ： %s \n", i, nodeStore.AddressNet(*IdBsP).B58String())
	}

	{
		//排序 打印
		kad := new(nodeStore.IdDESC)

		for _, v := range areaPeers {
			/*			utils.Log.Info().Msgf("NetId: %s VnodeId : %s",
						v.GetNetId().B58String(),
						v.GetVnodeId().B58String())
			*/
			*kad = append(*kad, new(big.Int).SetBytes([]byte(v.area.GetNetId())))
		}

		sort.Sort(kad)
		for i := 0; i < len(*kad); i++ {
			// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
			one := (*kad)[i]
			IdBs := one.Bytes()
			IdBsP := utils.FullHighPositionZero(&IdBs, 32)
			utils.Log.Info().Msgf("真实排序 %d ： %s \n", i, nodeStore.AddressNet(*IdBsP).B58String())
		}

	}
	utils.Log.Info().Msgf("开始等待节点自治")
	//等待各个节点都准备好
	for i, one := range areaPeers {
		utils.Log.Info().Msgf("----------------------------> start %d", i)
		one.area.WaitAutonomyFinish()
		// utils.Log.Info().Msgf("----------------------------> mid %d", i)
		// one.area.WaitAutonomyFinishVnode()
		utils.Log.Info().Msgf("----------------------------> end %d", i)
	}
	for i, one := range areaPeers {
		// utils.Log.Info().Msgf("----------------------------> start %d", i)
		// one.area.WaitAutonomyFinish()
		utils.Log.Info().Msgf("----------------------------> mid %d", i)
		one.area.WaitAutonomyFinishVnode()
		utils.Log.Info().Msgf("----------------------------> end %d", i)
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")
	utils.Log.Info().Msgf("", snum, lnum, small, large)
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("发送消息测试开始")
	// utils.Log.Info().Msgf("最小节点: %d  最大节点: %d", snum, lnum)
	// tt := time.NewTicker(5 * time.Second)
	// <-tt.C
	// utils.Log.Info().Msgf("删除areaPeers[2]虚拟节点 %s self %s", areaPeers[2].area.Vm.GetVnodeSelf()[1].Vid.B58String(),
	// 	areaPeers[2].area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	// areaPeers[2].area.DelSelfVnodeByAddress(nodeStore.AddressNet(areaPeers[2].area.Vm.GetVnodeSelf()[1].Vid))
	// utils.Log.Info().Msgf("删除vnode完成")
	//time.Sleep(60 * time.Second)
	// sendMsgLoop(areaPeers[0].area, areaPeers, lnum, snum)
	// utils.Log.Info().Msgf("--------------------------")
	// utils.Log.Info().Msgf("--------------------------")
	//<-time.NewTicker(5 * time.Second).C
	// for i := 0; i < 50000; i++ {
	// 	searchVnoideId(areaPeers[0].area, 0)
	// }
	// utils.Log.Info().Msgf("获取真实节点1111")
	// toNetidStr := "E2AH5YrEwr27GmGbxkzt4CZfy9mrtsLggjmqAxLFf2N1"
	// id := nodeStore.AddressFromB58String(toNetidStr)
	// utils.Log.Info().Msgf("sender : %s", areaPeers[2].area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	{
		utils.Log.Info().Msgf("普通消息发送 开始")
		con := []byte("heeeeellllllllooooooo")
		for i := 0; i < len(areaPeers); i++ {
			if i == 6 {
				continue
			}
			for n := len(areaPeers) - 1; n >= 0; n-- {
				if n == i {
					continue
				}
				utils.Log.Error().Msgf("zyz zyz 1111 self %s to %s", areaPeers[i].area.NodeManager.NodeSelf.IdInfo.Id.B58String(), areaPeers[n].area.NodeManager.NodeSelf.IdInfo.Id.B58String())
				_, _, _, err := areaPeers[i].area.SendP2pMsgWaitRequest(msg_id_p2p, &areaPeers[n].area.NodeManager.NodeSelf.IdInfo.Id, &con, 2*time.Second)
				if err != nil {
					utils.Log.Error().Msgf("zyz zyz 2222 error : %s", err.Error())
					continue
				}
			}

		}
		utils.Log.Info().Msgf("普通消息发送 结束")
	}

	{
		searchAddr := []string{
			"EdbKNMnG9nSarP76X73K8qnZyu3C6oqtithkrfZYcNFV",
			"EdbKNMnG9nSarP76X73K8qnZyu3C6oqtith4rfZYcNF1",

			"9b7LU2QW9ED9gX55vVTmT33kC9YJWKPECpVbWY5k2BB1",
			"7YuHM6tydLbjsnBSrb6jN3KrA4TqNBeH3FurmC4hpYg1",
			"6av3TYMGUyuMA31WfV36fzUoFGd2EE8WjwcKYr8AFGn1",
			"aXTebYN68gs2rsAdZ35zJMMZfA56nruSUpMwQwkxXk1",
			"DdnVnUFSdfQBrJ43kirMqQEARUVy8hsi8DP5EApx8D1",
			"CEEJz55QSokCmW5GxyZNuysQN7EQQv4EbmWymmuy6H1",
		}
		utils.Log.Info().Msgf("searchnetaddr 开始")
		for i := 0; i < len(areaPeers); i++ {
			if i == 6 {
				for u := range searchAddr {
					utils.Log.Info().Msgf("")
					utils.Log.Info().Msgf("======================代理使用============================================")
					utils.Log.Info().Msgf("sender %s  searchaddr %s", areaPeers[i].area.NodeManager.NodeSelf.IdInfo.Id.B58String(), searchAddr[u])
					s := nodeStore.AddressFromB58String(searchAddr[u])
					re, err := areaPeers[i].area.SearchNetAddrOneByOneAuto(&s, 3)
					if err != nil {
						utils.Log.Error().Msgf("SearchNetAddrOneByOneAuto err : %s", err.Error())
						continue
					}
					if len(re) != 3 {
						utils.Log.Error().Msgf("SearchNetAddrOneByOneAuto Num : %d", len(re))
						continue
					}
					for n := 0; n < len(re); n++ {
						utils.Log.Info().Msgf("index %d addr %s", n, re[n].B58String())
					}
				}
			}

			for u := range searchAddr {
				utils.Log.Info().Msgf("")
				utils.Log.Info().Msgf("==================================================================")
				utils.Log.Info().Msgf("sender %s  searchaddr %s", areaPeers[i].area.NodeManager.NodeSelf.IdInfo.Id.B58String(), searchAddr[u])
				s := nodeStore.AddressFromB58String(searchAddr[u])
				re, err := areaPeers[i].area.SearchNetAddrWithNum(&s, 3)
				if err != nil {
					utils.Log.Error().Msgf("SearchNetAddrWithNum err : %s", err.Error())
					continue
				}
				if len(re) != 3 {
					utils.Log.Error().Msgf("SearchNetAddrWithNum Num : %d", len(re))
					continue
				}
				for n := 0; n < len(re); n++ {
					utils.Log.Info().Msgf("index %d addr %s", n, re[n].B58String())
				}
			}
		}
		utils.Log.Info().Msgf("searchnetaddr 结束")
	}

	// var wg sync.WaitGroup
	// var successCnt, failedCnt int32
	// startTime := time.Now().UnixMilli()

	// for i := 0; i < 40000; i++ {
	// 	wg.Add(1)

	// 	go func() {
	// 		defer wg.Done()
	// 		r1, err := areaPeers[2].area.SearchNetAddrOneByOne(&id, 0)
	// 		if err != nil {
	// 			atomic.AddInt32(&failedCnt, 1)
	// 			utils.Log.Warn().Msgf(err.Error())
	// 			return
	// 		}
	// 		if len(r1) == 0 {
	// 			atomic.AddInt32(&failedCnt, 1)
	// 			utils.Log.Info().Msgf("失败")
	// 			return
	// 		}
	// 		atomic.AddInt32(&successCnt, 1)
	// 	}()
	// }
	// wg.Wait()
	// endTime := time.Now().UnixMilli()
	// useTime := endTime - startTime

	// utils.Log.Info().Msgf("--------------------------------------------------")
	// utils.Log.Info().Msgf("")
	// utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	// utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	// utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	// utils.Log.Warn().Msgf("TPS:%v", (successCnt*10000)/int32(useTime))
	// utils.Log.Info().Msgf("--------------------------------------------------")
	// utils.Log.Info().Msgf("")
	// for _, one := range r1 {
	// 	utils.Log.Info().Msgf("7777 :%s", one.B58String())
	// }

	// utils.Log.Info().Msgf("获取真实节点2222")

	// utils.Log.Info().Msgf("sender : %s", areaPeers[2].area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	// r2, err := areaPeers[2].area.SearchNetAddrOneByOne(&id, 0)
	// if err != nil {
	// 	utils.Log.Warn().Msgf(err.Error())
	// }
	// for _, one := range r2 {
	// 	utils.Log.Info().Msgf("8888 :%s", one.B58String())
	// }
	// utils.Log.Info().Msgf("sender : %s", areaPeers[lnum].area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	// r2, err := areaPeers[lnum].area.SearchNetAddrOneByOne(&id, 0)
	// if err != nil {
	// 	utils.Log.Warn().Msgf(err.Error())
	// }
	// for _, one := range r2 {
	// 	utils.Log.Info().Msgf("9999 :%s", one.B58String())
	// }
	utils.Log.Info().Msgf("发送消息测试完成")
	select {}
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

	//area.OpenVnode()

	//serverHost
	if i == 6 {
		area.SetAreaGodAddr(host, basePort)
		area.SetPhoneNode()
	}

	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))

	if i == 6 {
		ok, addr := area.SetAreaGodAddr(host, basePort)
		utils.Log.Info().Msgf("==============================设置代理 %s 设置代理成功 %t self %s", addr, ok, area.NodeManager.NodeSelf.IdInfo.Id.B58String())
		area.SetPhoneNode()
	}

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

func sendMsgOneSearchSuper(area *libp2parea.Area, toArea *libp2parea.Area) {
	utils.Log.Info().Msgf("===============sendMsgOneSearchSuper> 节点:%s 发送消息给:%s 开始！！", area.GetNodeSelf().IdInfo.Id.B58String(), toArea.Vm.GetVnodeDiscover().Vnode.Nid.B58String())
	utils.Log.Info().Msgf("")
	_, err := area.SendSearchSuperMsgWaitRequest(msg_id_searchSuper, &toArea.Vm.GetVnodeDiscover().Vnode.Nid, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return
	}
	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendMsgOneSearchSuper< 节点:%s 发送消息给:%s 结束！！", area.GetNodeSelf().IdInfo.Id.B58String(), toArea.Vm.GetVnodeDiscover().Vnode.Nid.B58String())
}

func sendVnodeP2pMsgHE(area *libp2parea.Area, toArea *libp2parea.Area) {
	//真实节点 from : GVi4gfzEVzGBH5uNHocZTLPG8ywMi5oZbME7rPWbKDcY  to : 6SnYLyD8oPoMyKLKr7MKmZ5Wqjis2hLFw6dbKSDVsMUE
	con := []byte("123")
	toVnodes := area.Vm.GetVnodeSelf()

	utils.Log.Info().Msgf("===============sendVnodeP2pMsgHE> 虚拟节点:%s 发送消息给 虚拟节点:%s || 真实节点 from : %s  to : %s 开始！！", toVnodes[1].Vid.B58String(), toArea.Vm.GetVnodeSelf()[1].Vid.B58String(),
		toVnodes[1].Nid.B58String(), toArea.NodeManager.NodeSelf.IdInfo.Id.B58String())
	utils.Log.Info().Msgf("")

	_, err := area.SendVnodeP2pMsgHEWaitRequest(msg_id_vnode_p2p, &toVnodes[1].Vid, &toArea.Vm.GetVnodeSelf()[1].Vid, &toArea.NodeManager.NodeSelf.IdInfo.Id, &con, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送VnodeP2p消息失败:%s", err.Error())
		return
	}

	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendVnodeP2pMsgHE< 虚拟节点:%s 发送消息给 虚拟节点:%s 结束！！", toVnodes[1].Vid.B58String(), toArea.Vm.GetVnodeSelf()[1].Vid.B58String())

}

func sendMsgOneP2pMsg(area *libp2parea.Area, toArea *libp2parea.Area, ll int) {
	toNetid := toArea.GetNetId()

	// //主动发地址
	// toNetidStr := "AuWrFLnPttkfypro3WMStx9QiFRHRx1HGqHsiAxwTFqq"
	// toNetid = nodeStore.AddressNet(engine.AddressFromB58String(toNetidStr))

	cont := []byte("00")

	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendMsgOneP2pMsg> 节点:%s 发送消息给:%s 开始！！", area.GetNodeSelf().IdInfo.Id.B58String(), toNetid.B58String())
	utils.Log.Info().Msgf("")

	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, &cont, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return
	}

	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendMsgOneP2pMsg< 节点:%s 发送消息给:%s 结束！！", area.GetNodeSelf().IdInfo.Id.B58String(), toNetid.B58String())
	utils.Log.Info().Msgf("")

}

func sendVnodeSearchMsg(area *libp2parea.Area, toArea *libp2parea.Area) {
	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendVnodeSearchMsg> 节点:%s 发送消息给:%s 开始！！", area.Vm.GetVnodeSelf()[1].Vid.B58String(), toArea.Vm.GetVnodeSelf()[1].Vid.B58String())
	utils.Log.Info().Msgf("")

	_, err := area.SendVnodeSearchMsgWaitRequest(msg_id_vnode_search, &area.Vm.GetVnodeSelf()[1].Vid, &toArea.Vm.GetVnodeSelf()[1].Vid, nil, time.Second*10)
	if err != nil {
		utils.Log.Info().Msgf("发送VnodeSearchMsg error:%s", err.Error())
		return
	}

	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("===============sendVnodeSearchMsg< 节点:%s 发送消息给:%s 结束！！", area.Vm.GetVnodeSelf()[1].Vid.B58String(), toArea.Vm.GetVnodeSelf()[1].Vid.B58String())
	utils.Log.Info().Msgf("")
}

func searchVnoideId(area *libp2parea.Area, num uint16) {
	// utils.Log.Info().Msgf("===============> 节点 %d", num)

	//	toNetidStr := "75Q8RGVmHfE7qDt616oJn5zBtcvkciM5N2zm97hyDa1Z2"
	toNetidStr := "FntW7YmLqoq1xV7rabu6GASUHXNhLhuLQtaJpc1rL531"
	id := nodeStore.AddressFromB58String(toNetidStr)
	vid := virtual_node.AddressNetExtend(id)
	//utils.Log.Info().Msgf("sender %s", area.Vm.VnodeMap[1].Vnode.Vid.B58String())
	bs1, err := area.SearchVnodeIdOnebyone(&vid, num)
	if err != nil {
		utils.Log.Info().Msgf("发送VnodeSearchMsg error:%s", err.Error())
		return
	}
	// for _, v := range bs1 {
	// 	utils.Log.Info().Msgf("66666 v:%s n:%s", v.Vid.B58String(), v.Nid.B58String())
	// }
	if len(bs1) < 5 {
		utils.Log.Warn().Msgf("出错 出错 出错 出错 出错")
	}
	// bs2, err := area.SearchVnodeId(&vid)
	// if err != nil {
	// 	utils.Log.Info().Msgf("发送VnodeSearchMsg error:%s", err.Error())
	// 	return
	// }
	// utils.Log.Info().Msgf("7777", bs2.B58String())
	// // updown := make([]string, 0)
	// // err = json.Unmarshal(*bs, &updown)
	// if err != nil {
	// 	utils.Log.Info().Msgf("marsh err : %s", err.Error())
	// }
	// for i := 0; i <= len(*bs); {
	// 	v := i + 64

	//
	// }

	// utils.Log.Info().Msgf("===============< 节点 %d", num)
}

func sendMsgLoop(area *libp2parea.Area, toAddrs []*TestPeer, ll int, ss int) {

	// //version_p2p  msgid = 5
	// for i := 0; i < len(toAddrs); i++ {
	// 	to := toAddrs[i]
	// 	sendMsgOneP2pMsg(toAddrs[len(toAddrs)-1].area, to.area, ll)
	// 	time.Sleep(3 * time.Second)
	// }

	// //version_search_super  msgid = 3
	// for i := 0; i < len(toAddrs); i++ {
	// 	sendMsgOneSearchSuper(area, toAddrs[len(toAddrs)-1].area)
	// }

	// // //version_vnode_p2pHE  msgid = 8
	// for i := 0; i < len(toAddrs); i++ {
	// 	sendVnodeP2pMsgHE(toAddrs[len(toAddrs)-1].area, toAddrs[1].area)
	// }

	// // version_vnode_search msgid = 7
	// for i := 0; i < len(toAddrs); i++ {
	// 	sendVnodeSearchMsg(toAddrs[len(toAddrs)-1].area, toAddrs[1].area)
	// }
	// con := []byte("HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH")
	// utils.Log.Info().Msgf(" -------- con %d", len(con))

	//searchVnoideId(toAddrs[1].area, 3)
	searchVnoideId(toAddrs[ll].area, uint16(3))
	searchVnoideId(toAddrs[ss].area, uint16(0))
	searchVnoideId(toAddrs[1].area, uint16(1))
	// from := toAddrs[ss].area
	// to := toAddrs[ll].area
	// utils.Log.Info().Msgf("1000条消息发送开始")
	// for i := 0; i < 1000; i++ {
	// 	sendMsgOne(from, to, &con)
	// }
	// utils.Log.Info().Msgf("1000条消息发送结束")
	// utils.Log.Info().Msgf("1000条加密消息发送开始")
	// for i := 0; i < 1000; i++ {
	// 	sendMsgOneHe(from, to, &con)
	// }
	// utils.Log.Info().Msgf("1000条加密消息发送结束")
}

func sendMsgOneHe(area *libp2parea.Area, toArea *libp2parea.Area, con *[]byte) error {
	//发送p2p消息

	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()

	_, _, _, err := area.SendP2pMsgHEWaitRequest(msg_id_p2p, &toNetid, con, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}
	utils.Log.Info().Msgf("结束")
	return nil
}

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area, con *[]byte) error {
	//发送p2p消息

	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, con, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}
	utils.Log.Info().Msgf("结束")
	return nil
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_searchSuper = 1003
const msg_id_searchSuper_recv = 1004 //加密消息

const msg_id_vnode_p2p = 1005
const msg_id_vnode_p2p_recv = 1006 //加密消息

const msg_id_vnode_search = 1007      //搜索节点消息
const msg_id_vnode_search_recv = 1008 //搜索节点消息 返回

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.SearchSuperHandler)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)

	area.Register_vnode_p2pHE(msg_id_vnode_p2p, this.RecvMsgHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_p2p_recv, this.RecvMsgHEHandler)

	area.Register_vnode_search(msg_id_vnode_search, this.SearchVnodeHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_search_recv, this.SearchVnodeHandler_recv)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	utils.Log.Info().Msgf("收到P2P消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	if message.Body.Content != nil {
		utils.Log.Info().Msgf("messs body %s", string(*message.Body.Content))
	}
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
	//	utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
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
	// utils.Log.Info().Msgf("收到vnode p2p消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
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
