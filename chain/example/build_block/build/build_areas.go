package build

import (
	"path/filepath"
	"strconv"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter"
	"web3_gui/libp2parea/adapter"
	"web3_gui/utils"
)

func BuildAreas(n int, areaName [32]byte, addrPre, keyPwd, keystoreDirName string) []*libp2parea.Area {
	err := utils.CheckCreateDir(keystoreDirName)
	if err != nil {
		panic(err)
	}
	areas := make([]*libp2parea.Area, 0, n)
	//fmt.Println("start")
	for i := 0; i < n; i++ {
		keyPath1 := filepath.Join(keystoreDirName, "keystore"+strconv.Itoa(i)+".key")

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
		if len(key1.GetAddr()) < len(config.RoleRewardList) {
			n := len(config.RoleRewardList) - len(key1.GetAddr()) + 1
			for i := 0; i < n; i++ {
				utils.LogZero.Info().Int("已有地址数量", len(key1.GetAddr())).Int("需要地址数量", len(config.RoleRewardList)).Int("创建地址数量", i).Send()
				_, err = key1.GetNewAddr(keyPwd, keyPwd)
				if err != nil {
					panic("创建Addr错误:" + err.Error())
				}
			}
		}
		if len(key1.GetDHKeyPair().SubKey) < 1 {
			_, err = key1.GetNewDHKey(keyPwd, keyPwd)
			if err != nil {
				panic("创建Addr错误:" + err.Error())
			}
		}

		area, err := libp2parea.NewArea(areaName, key1, keyPwd)
		//area.SetDiscoverPeer(serverHost + ":" + strconv.Itoa(basePort))
		//area.StartUP(false, clientHost, uint16(basePort+i))
		//area.Destroy()
		//time.Sleep(time.Second)

		//utils.LogZero.Info().Msgf("协程数量 %d", runtime.NumGoroutine())
		areas = append(areas, area)
	}
	return areas
}
