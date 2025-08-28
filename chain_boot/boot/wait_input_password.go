package boot

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"syscall"
	chainconfig "web3_gui/chain/config"
	ks2 "web3_gui/keystore/v2"
	ks2config "web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
检查钱包文件是否存在，存在则输入密码。
不存在，则输入两次初始密码。
*/
func ParseWalletPassword(oldFilePath, newFilePath, addrPre string) (string, error) {
	//检查是否存在新版密钥文件
	have := existsKeyStoreFileNew(newFilePath, addrPre)
	if !have {
		//utils.Log.Info().Str("没有新版密钥文件", "").Send()
		//检查是否存在旧版密钥文件
		have = existsKeyStoreFileOld(oldFilePath, addrPre)
		//utils.Log.Info().Str("没有旧版密钥文件", oldFilePath).Bool("have", have).Send()
	}
	var pwd string
	var err error
	for t := 1; t <= 3; t++ {
		if have {
			if pwd, err = WaitInputPwd(); err != nil {
				fmt.Println(err.Error())
				continue
			}
			break
		} else {
			if pwd, err = WaitInputInitPwd(); err != nil {
				fmt.Println(err.Error())
				continue
			}
			break
		}
	}
	if err != nil {
		return "", errors.New("Password verification failed!")
	}
	//config.Wallet_keystore_default_pwd = pwd
	return pwd, nil
}

/*
等待输入密码
*/
func WaitInputPwd() (string, error) {
	return getInputPwd("password:")
}

/*
等待输入两次初始密码
*/
func WaitInputInitPwd() (string, error) {
	pwd1, err := getInputPwd("new password:")
	if err != nil {
		return "", fmt.Errorf("error:%s", err.Error())
	}
	pwd2, err := getInputPwd("retype new password:")
	if err != nil {
		return "", fmt.Errorf("error:%s", err.Error())
	}
	if pwd1 != pwd2 {
		return "", errors.New("The password entered twice is inconsistent!")
	}
	return pwd1, nil
}

func getInputPwd(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	pwdStr := string(bytePwd)
	if len(pwdStr) > 2048 {
		return "", errors.New("password length error")
	}
	if len(pwdStr) == 0 {
		pwdStr = chainconfig.Wallet_keystore_default_pwd
	}
	return pwdStr, nil
}

func existsKeyStoreFileOld(filePath, addrPre string) bool {
	//ks1.NewKeystore(filePath, addrPre).Load() == nil
	ok, err := utils.PathExists(filePath)
	if err != nil {
		return true
	}
	return ok
}
func existsKeyStoreFileNew(filePath, addrPre string) bool {
	wallet := ks2.NewWallet(filePath, addrPre)
	ERR := wallet.Load()
	if ERR.CheckFail() {
		if ERR.Code == ks2config.ERROR_code_wallet_file_not_exist {
			return false
		}
	}
	return true
}
