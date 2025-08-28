package main

import (
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/term"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"
	"web3_gui/chain/boot"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter"
	"web3_gui/libp2parea/adapter/engine"
)

func main() {
	// NOTE 监控Cpu/Mem
	go profileCpu()
	go profileMem()

	StartChainWitnessVote("")
}

func StartChainWitnessVote(passwd string) {
	config.ParseConfig()
	if !config.EnableStartupWeb {
		if err := parseWalletPassword(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

func parseWalletPassword() error {
	flag := existsKeyStoreFile()
	var pwd string
	var err error
	for t := 1; t <= 3; t++ {
		if flag {
			if pwd, err = parsePwd(); err != nil {
				fmt.Println(err.Error())
				continue
			}
			break
		} else {
			if pwd, err = parseNewPwd(); err != nil {
				fmt.Println(err.Error())
				continue
			}
			break
		}
	}
	if err != nil {
		return errors.New("密码错误，程序退出!")
	}

	config.Wallet_keystore_default_pwd = pwd
	return nil
}

func parsePwd() (string, error) {
	return getPwd("请输入密码：")
}

func parseNewPwd() (string, error) {
	pwd1, err := getPwd("请输入密码：")
	if err != nil {
		return "", fmt.Errorf("读取输入密码错误:%s", err.Error())
	}
	pwd2, err := getPwd("请确认密码：")
	if err != nil {
		return "", fmt.Errorf("读取确认密码错误:%s", err.Error())
	}

	if pwd1 != pwd2 {
		return "", errors.New("确认密码错误!")
	}
	return pwd1, nil
}

func getPwd(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	pwdStr := string(bytePwd)
	if len(pwdStr) == 0 || len(pwdStr) > 2048 {
		return "", errors.New("密码长度错误")
	}
	return string(bytePwd), nil
}

func existsKeyStoreFile() bool {
	return keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre).Load() == nil
}
