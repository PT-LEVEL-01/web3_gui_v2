package main

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
)

func main() {
	engine.SetLogPath("log.txt")
	life_cycle()
}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	clientHost = "127.0.0.1"
	basePort   = 19960
)

func life_cycle() {
	fmt.Println("start")
	for i := 0; i < 10; i++ {
		keyPath1 := filepath.Join("conf", "keystore"+strconv.Itoa(0)+".key")

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
		area.SetDiscoverPeer(serverHost + ":" + strconv.Itoa(basePort))
		area.StartUP(false, clientHost, uint16(basePort+i))
		area.Destroy()
		time.Sleep(time.Second)

		utils.Log.Info().Msgf("协程数量 %d", runtime.NumGoroutine())
	}

}
