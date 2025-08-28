package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/hyahm/golog"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

func main() {
	utils.PprofMem(time.Minute * 2)

	golog.InitLogger("logother.txt", 0, true)
	golog.Infof("start %s", "log")
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("zyz"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 10
	areas := make([]*libp2parea.Area, 0, n)
	areaPeers := make([]*TestPeer, 0, n)
	allNetId := make([]nodeStore.AddressNet, 0)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		area.area.OpenVnode()
		area.area.Vm.AddVnode()
		area.area.Vm.AddVnode()
		areaPeers = append(areaPeers, area)
		areas = append(areas, area.area)
		allNetId = append(allNetId, area.area.GetNetId())
		nsm.AddNodeSuperIDs(area.area.GetNetId())
		time.Sleep(2 * time.Second)
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始等待节点自治")
	{
		//排序 打印
		kad := new(nodeStore.IdDESC)

		for _, v := range areas {
			//*kad = append(*kad, new(big.Int).SetBytes([]byte(v.Vm.DiscoverVnodes.Vnode.Vid)))
			for _, vInfo := range v.Vm.VnodeMap {
				if vInfo.Vnode.Index == 0 {
					continue
				}
				//utils.Log.Info().Msgf("Vnode.Vid : %s", vInfo.Vnode.Vid.B58String())
				*kad = append(*kad, new(big.Int).SetBytes([]byte(vInfo.Vnode.Vid)))
			}
		}

		sort.Sort(kad)
		for i := 0; i < len(*kad); i++ {
			// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
			one := (*kad)[i]
			IdBs := one.Bytes()
			IdBsP := utils.FullHighPositionZero(&IdBs, 32)
			utils.Log.Info().Msgf("Vnode 排序 %d ： %s \n", i, nodeStore.AddressNet(*IdBsP).B58String())
		}
	}

	{
		//排序 打印
		kad := new(nodeStore.IdDESC)

		for _, v := range areas {
			/*			utils.Log.Info().Msgf("NetId: %s VnodeId : %s",
						v.GetNetId().B58String(),
						v.GetVnodeId().B58String())
			*/
			*kad = append(*kad, new(big.Int).SetBytes([]byte(v.GetNetId())))
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
	//等待各个节点都准备好
	for _, one := range areaPeers {
		one.area.WaitAutonomyFinish()
		one.area.WaitAutonomyFinishVnode()
	}

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")
	// sleepTime := time.Second * 100
	// utils.Log.Info().Msgf("等待%s后打印", sleepTime)
	// time.Sleep(sleepTime)
	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("")
	utils.Log.Info().Msgf("")

	utils.Log.Info().Msgf("CloseNet CloseNet CloseNet CloseNet")
	// from := areaPeers[0].area
	// to := areaPeers[1].area
	// for i := 0; i < 1000; i++ {
	// 	sendMsgOne(from, to)
	// }
	select {}
	//#####################关闭节点测试 start
	//单独开节点测试上下线
	// time.Sleep(2 * time.Minute)
	// utils.Log.Info().Msgf("============》节点下线测试 %s", areas[1].GetNetId().B58String())
	// for {
	// 	areas[1].Destroy()
	// 	time.Sleep(2 * time.Second)
	// }

	//#####################关闭节点测试 finish

	//#####################增加节点测试 start
	// time.Sleep(time.Minute * 2)
	// utils.Log.Info().Msgf("========ADD NODE============")

	// area := StartOnePeer(11)
	// areaPeers = append(areaPeers, area)
	// areas = append(areas, area.area)
	// nsm.AddNodeSuperIDs(area.area.GetNetId())
	// utils.Log.Info().Msgf("NetId: %s VnodeId : %s",
	// 	area.area.GetNetId().B58String(),
	// 	area.area.GetVnodeId().B58String())
	// //排序 打印
	// *kad = append(*kad, new(big.Int).SetBytes([]byte(area.area.GetNetId())))

	// sort.Sort(kad)
	// for i := 0; i < len(*kad); i++ {
	// 	// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
	// 	one := (*kad)[i]
	// 	IdBs := one.Bytes()
	// 	IdBsP := utils.FullHighPositionZero(&IdBs, 32)
	// 	utils.Log.Info().Msgf("增加节点_重新排序 %d ： %s \n", i, nodeStore.AddressNet(*IdBsP).B58String())
	// }
	//#####################临时增加节点测试 finish

}

//排序

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
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))

	peer := TestPeer{
		area: area,
	}

	//peer.InitHandler(area)
	return &peer
}

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	//发送p2p消息

	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, nil, time.Second*10)
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

// func (this *TestPeer) InitHandler(area *libp2parea.Area) {
// 	//area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
// 	area.Register_p2p(msg_id_p2p, this.FileMsg)
// 	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

// 	area.Register_p2p(msg_id_searchSuper, this.FindFileMsg)
// 	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)
// }

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
}
