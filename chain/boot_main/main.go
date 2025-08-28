package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
	"web3_gui/chain/boot"
	"web3_gui/chain/boot/startup_web"
	selfConfig "web3_gui/chain/boot_main/config"
	"web3_gui/chain/config"
	"web3_gui/libp2parea/adapter/engine"
)

func main() {
	selfConfig.InitConfig()
	// NOTE 监控Cpu/Mem
	go profileCpu()
	go profileMem()

	StartChainWitnessVote("")
}

func StartChainWitnessVote(passwd string) {
	config.ParseConfig()
	if config.EnableStartupWeb {
		//获取本地公私钥
		if startupPwd, err := boot.GetInputPasswordForStartup(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			//startup web
			if err := startup_web.Start(startupPwd); err != nil {
				fmt.Println("Startup:", err)
				os.Exit(1)
			}
		}
	} else {
		//if pwd, err := boot.GetInputPasswordForKeystore(); err != nil {
		//	fmt.Println("Input:", err)
		//	os.Exit(1)
		//} else {
		//	config.Wallet_keystore_default_pwd = pwd
		//}
	}
	boot.StartWithArea("")
}

func profileMem() {
	base := "./profiles"
	os.MkdirAll(base, 0644)
	limit := float64(70) //大于70%记录mem pprof
	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case t := <-ticker.C:
			memv := float64(0)
			if v, err := mem.VirtualMemory(); err == nil {
				memv = v.UsedPercent
			}
			if memv < limit {
				continue
			}

			filename := filepath.Join(base, fmt.Sprintf("mem_t%d_p%d.pprof", t.Unix(), int(memv)))
			f, err := os.Create(filename)
			if err != nil {
				engine.Log.Error("Mem:%v", err)
				continue
			}
			if err := pprof.WriteHeapProfile(f); err != nil {
				engine.Log.Error("Mem:%v", err)
				continue
			}
			f.Close()
		}
	}
}

func profileCpu() {
	base := "./profiles"
	os.MkdirAll(base, 0644)
	limit := float64(70) //大于70%记录cpu pprof
	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case t := <-ticker.C:
			cpuv := float64(0)
			if v, _ := cpu.Percent(10*time.Second, false); len(v) > 0 {
				cpuv = v[0]
			}
			if cpuv < limit {
				continue
			}
			filename := filepath.Join(base, fmt.Sprintf("cpu_t%d_p%d.pprof", t.Unix(), int(cpuv)))
			f, err := os.Create(filename)
			if err != nil {
				continue
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				continue
			}
			time.Sleep(10 * time.Second)
			pprof.StopCPUProfile()
			f.Close()
		}
	}
}
