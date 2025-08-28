package main

import (
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/cake/country"
	"web3_gui/libp2parea/v2/config"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func main() {
	utils.Log.Info().Msgf("start")

	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

var areaInfo = `{
	"address": {
		"47.109.16.70:19965": "FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n",
		"47.109.16.70:19966": "4MTCfHur4XPmRhpUByGC7Z9ve3pbezDyiYuYM6cA5Lmj",
		"47.109.16.70:19967": "6Wrco6JKC5c7pg8qp5Z5mNnayHqdW1eyLExHboihCY4c",
		"47.109.16.70:19968": "5o1oZJwByLAgT36BXXbBRDJ8UDhUUY4JLJbjR3aG6ZEp",
		"47.109.16.70:19969": "qUX2PsQATzSaofU6Q7sGmipu1g2V1H2KwgdzcqZsbN7",
		"47.109.16.70:19970": "GQHsVDddC44AM6z4LmS6f1brkJn1cNYhrms75ZEwchAG"
	},
	"country": {
		"CN": "CN",
		"HK": "HK",
		"UK":"UK"
	},
	"nodes2": {
		"CN": {
			"FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n": "47.109.16.70:19965",
			"4MTCfHur4XPmRhpUByGC7Z9ve3pbezDyiYuYM6cA5Lmj": "47.109.16.70:19966",
			"6Wrco6JKC5c7pg8qp5Z5mNnayHqdW1eyLExHboihCY4c": "47.109.16.70:19967"
			
		},
		"HK": {
			"qUX2PsQATzSaofU6Q7sGmipu1g2V1H2KwgdzcqZsbN7": "47.109.16.70:19969",
			"GQHsVDddC44AM6z4LmS6f1brkJn1cNYhrms75ZEwchAG": "47.109.16.70:19970"
		},
		"UK":{
			"5o1oZJwByLAgT36BXXbBRDJ8UDhUUY4JLJbjR3aG6ZEp": "47.109.16.70:19968"
		},
		"default": {
			"FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n": "47.109.16.70:19965"
		}
	},
	"nodes": {
		"CN": "47.109.16.70:19965,47.109.16.70:19966",
		"MY": "47.109.16.70:19967,47.109.16.70:19968",
		"SG": "47.109.16.70:19969,47.109.16.70:19970",
		"default": "47.109.16.70:19965"
	},
	"version": 23
}`

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 6
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)
		ct := country.NewAreaCountry(area.area, i == 0)
		area.ct = ct
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

	for i := range areaPeers {
		areaPeers[i].ct.Start()
		if i == 0 {
			areaPeers[i].ct.SetData(areaInfo)
		}
	}

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	// for i := 1; i < len(areaPeers); i++ {
	// 	go func(nIndex int) {
	// 		for {
	// 			value, err := areaPeers[nIndex].ct.GetNodeData()
	// 			if err != nil {
	// 				utils.Log.Error().Msgf("[%s] ct.GetNodeData() err:%s", areaPeers[nIndex].area.GetNetId().B58String(), err)
	// 			} else {
	// 				utils.Log.Info().Msgf("[%s] ct.GetNodeData() 获取到的数据为: %s", areaPeers[nIndex].area.GetNetId().B58String(), value)
	// 			}
	// 			time.Sleep(time.Second * 15)
	// 		}
	// 	}(i)
	// }

	cnt := 0
	for {
		utils.Log.Info().Msgf("")
		utils.Log.Info().Msgf("---------------------------------------------------")
		utils.Log.Info().Msgf("")

		time.Sleep(time.Second * 30)
		cnt++
		// value := fmt.Sprintf("哈哈哈 这是 %d 次的消息", cnt)
		// areaPeers[0].ct.SetData(value)

		if cnt >= 6 {
			cnt = 0
		}

		nodeIds, err := areaPeers[cnt].ct.GetAreaSaveNodeIds()
		if err != nil {
			utils.Log.Error().Msgf("GetAreaSaveNodeIds err:%s", err)
			continue
		}

		if len(nodeIds) == 0 {
			utils.Log.Error().Msgf("GetAreaSaveNodeIds 获取到的结果为空!!!!")
		}

		for i := range nodeIds {
			utils.Log.Error().Msgf("[%d] area node id:%s", i, nodeIds[i].B58String())
		}
	}
}

type TestPeer struct {
	area *libp2parea.Area
	ct   *country.AreaCountry
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

	area.SetMachineID("area_machine_id_" + strconv.Itoa(i))

	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))

	peer := TestPeer{
		area: area,
	}

	return &peer
}
