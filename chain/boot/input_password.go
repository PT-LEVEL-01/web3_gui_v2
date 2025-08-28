package boot

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"path/filepath"
	"syscall"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter"
	"web3_gui/utils"
)

func GetInputPasswordForStartup() (string, error) {
	startupfile := filepath.Join(config.Path_configDir, config.Core_startup_key)
	ok, _ := utils.PathExists(startupfile)
	if !ok {
		password, err := InputPwd("请输入启动WEB密码:")
		if err != nil {
			return "", err
		}
		repeatPassword, err := InputPwd("请重复输入启动WEB密码:")
		if err != nil {
			return "", err
		}
		if password != repeatPassword {
			return "", errors.New("输入密码不一致")
		}
		return password, nil
	}
	return "", nil
}

func GetInputPasswordForKeystore() (string, error) {
	password, err := InputPwd("请输入钱包密码:")
	if err != nil {
		return "", err
	}
	if !existsKeyStoreFile() {
		repeatPassword, err := InputPwd("请重复输入钱包密码:")
		if err != nil {
			return "", err
		}
		if password != repeatPassword {
			return "", errors.New("输入密码不一致")
		}
	}
	return password, nil
}

func existsKeyStoreFile() bool {
	return keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre).Load() == nil
}

func InputPwd(prompt string) (string, error) {
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
