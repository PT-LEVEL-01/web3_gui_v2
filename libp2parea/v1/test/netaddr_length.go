package main

import (
	"crypto/sha256"
	"fmt"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1/nodeStore"
)

/*
网络地址长度测试
*/
func main() {
	for i := 0; i < 100000; i++ {
		example()
	}
}

func example() {
	fileAbsPath := ""
	password := "123456789"
	ks := keystore.NewKeystore(fileAbsPath)
	pwd := sha256.Sum256([]byte(password))
	err := ks.CreateNewWallet(pwd)
	if err != nil {
		return
	}

	_, puk, err := ks.GetNetAddrPuk(password)
	addrNet := nodeStore.BuildAddr(puk)
	addrStr := addrNet.B58String()
	length := len(addrStr)
	if length != 44 && length != 43 {
		fmt.Println(len(addrStr))
	}

}
