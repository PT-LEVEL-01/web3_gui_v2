package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

func main() {
	utils.PprofMem(time.Minute * 2)

	// golog.InitLogger("logother.txt", 0, true)
	// golog.Infof("start %s", "log")
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

	n := 31
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

	for _, one := range areaPeers {
		one.areas = areas
		for _, two := range areaPeers {
			one.area.AddWhiteList(two.area.GetNetId())
			one.area.AddConnect(two.area.GetNodeSelf().Addr, two.area.GetNodeSelf().TcpPort)
		}
	}

	// sleepTime := time.Second * 3
	// utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	// time.Sleep(sleepTime)

	//--------------------------
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始发送消息")

	total := 2
	//异步发消息
	for i := 0; i < total; i++ {
		peerOne := areaPeers[i%31]
		sendMsgOne(peerOne.area, areaPeers, uint64(i))
		time.Sleep(time.Second * 2)
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("发送消息完成")

	time.Sleep(time.Second * 10)

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("打印收到的消息")
	for i := 0; i < total; i++ {
		for _, one := range areaPeers {
			// utils.Log.Info().Msgf("print block map:%s p:%p", one.area.GetNetId().B58String(), one.block)
			addrs, ok := one.block.Load(uint64(i))
			if ok {
				startNano := int64(0)
				endNano := int64(0)

				m := addrs.(*sync.Map)
				m.Range(func(k, v interface{}) bool {
					// addrStr := k.(string)
					now := v.(int64)
					// addr := nodeStore.AddressNet([]byte(addrStr))
					// utils.Log.Info().Msgf("地址:%s 时间:%s", addr.B58String(), time.Unix(0, int64(now)))
					if startNano == 0 {
						startNano = now
						endNano = now
					} else {
						if startNano > now {
							startNano = now
						}
						if endNano < now {
							endNano = now
						}
					}
					return true
				})
				utils.Log.Info().Msgf("地址:%s 开始时间:%s 结束时间:%s 间隔时间:%s", one.area.GetNetId().B58String(), time.Unix(0, int64(startNano)), time.Unix(0, int64(endNano)), time.Duration(endNano-startNano))
			} else {
				utils.Log.Info().Msgf("未找到这个高度的区块:%d", i)
			}
		}
	}

	// select {}

}

type TestPeer struct {
	area  *libp2parea.Area
	block *sync.Map //key:uint64=区块高度;value:=;

	areas []*libp2parea.Area
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

	area.CloseVnode()

	//serverHost
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))
	// area.CloseVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	vnodeinfo := area.Vm.AddVnodeByIndex(100)
	utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	vnodeinfo = area.Vm.AddVnodeByIndex(200)
	utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	vnodeinfo = area.Vm.AddVnodeByIndex(300)
	utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	vnodeinfo = area.Vm.AddVnodeByIndex(400)
	utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())

	peer := TestPeer{
		area:  area,
		block: new(sync.Map),
	}

	peer.InitHandler(area)
	return &peer
}

var rand = utils.GetHashForDomain(utils.GetRandomDomain())
var magneticIds []nodeStore.AddressNet
var magneticVnodeIds []virtual_node.AddressNetExtend

func sendMsgOne(area *libp2parea.Area, toAddrs []*TestPeer, height uint64) error {
	heightBs := utils.Uint64ToBytes(height)
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.area.GetNetId()) {
			continue
		}
		netId := one.area.GetNetId()
		area.SendP2pMsg(msg_id_p2p_body, &netId, &heightBs)
	}
	return nil
}

const msg_id_p2p_start = 1001
const msg_id_p2p_body = 1002
const msg_id_p2p_end = 1003

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p_start, this.RecvP2PMsgHandlerStart)
	area.Register_p2p(msg_id_p2p_body, this.RecvP2PMsgHandlerBody)
	area.Register_p2p(msg_id_p2p_end, this.RecvP2PMsgHandlerEnd)
}

func (this *TestPeer) RecvP2PMsgHandlerStart(c engine.Controller, msg engine.Packet, message *mc.Message) {

}

func (this *TestPeer) RecvP2PMsgHandlerBody(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("body")
	for _, one := range this.areas {
		if bytes.Equal(this.area.GetNetId(), one.GetNetId()) {
			continue
		}
		netid := one.GetNetId()
		this.area.SendP2pMsg(msg_id_p2p_end, &netid, message.Body.Content)
	}
}

func (this *TestPeer) RecvP2PMsgHandlerEnd(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("end")
	height := utils.BytesToUint64(*message.Body.Content)
	newMap := new(sync.Map)
	newMap.Store(utils.Bytes2string(*message.Head.Sender), time.Now().UnixNano())
	// utils.Log.Info().Msgf("save block map:%p", this.block)
	// utils.Log.Info().Msgf("save block map:%s p:%p", this.area.GetNetId().B58String(), this.block)
	m, ok := this.block.LoadOrStore(height, newMap)
	if ok {
		// utils.Log.Info().Msgf("end 2222222222222 %d", height)
		newMap := m.(*sync.Map)
		newMap.Store(utils.Bytes2string(*message.Head.Sender), time.Now().UnixNano())
	} else {
		// utils.Log.Info().Msgf("end 33333333333333 %d", height)
	}

	// this.block.
}
