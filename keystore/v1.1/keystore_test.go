package keystore

import (
	"fmt"
	"testing"
)

const (
	path         = "keystore.key"
	addrPre      = "ICOM"
	pwd          = "123"
	fristAddrPwd = "1234"
)

func TestMnemonic(t *testing.T) {
	//MnemonicExample()
	example_UpdatePwd()
}

func example_UpdatePwd() {
	keyPath := "D:\\keystore/keystore.key"
	//oldPwd := "123"
	oldPwd := "123"
	newPwd := "123456789"
	keystore := NewKeystore(keyPath, addrPre)
	if err := keystore.Load(); err != nil {
		panic(err)
	}
	ok, err := keystore.UpdatePwd(oldPwd, newPwd)
	if err != nil {
		panic(err)
	}
	fmt.Println("修改种子密码", ok)

	addrInfo := keystore.GetAddr()
	for _, one := range addrInfo {
		ok, err := keystore.UpdateAddrPwd(one.GetAddrStr(), oldPwd, newPwd)
		if err != nil {
			panic(err)
		}
		fmt.Println("修改收款地址密码", ok)
	}
	ok, err = keystore.UpdateNetAddrPwd(oldPwd, newPwd)
	if err != nil {
		panic(err)
	}
	fmt.Println("修改网络地址密码", ok)

	ok, err = keystore.UpdateDHKeyPwd(oldPwd, newPwd)
	if err != nil {
		panic(err)
	}
	fmt.Println("修改DH密码", ok)

}

func MnemonicExample() {
	keystore := NewKeystore(path, addrPre)
	if err := keystore.Load(); err != nil {
		keystore.CreateNewKeystore(pwd)
	}

	//_, _, err := keystore.GetNetAddr(fristAddrPwd)
	//if err != nil {
	//	panic("GetNetAddr error:" + err.Error())
	//}
	//
	//words, err := keystore.ExportMnemonic(pwd)
	//if err != nil {
	//	panic(err)
	//}
	words := "salt slush prepare manual labor van snack person achieve recycle clever join"
	fmt.Println("打印助记词:", words)
	keystore = NewKeystore(path, addrPre)
	err := keystore.ImportMnemonic(words, pwd, fristAddrPwd, fristAddrPwd)
	if err != nil {
		panic(err)
	}
	words, err = keystore.ExportMnemonic(pwd)
	if err != nil {
		panic(err)
	}
	fmt.Println("打印助记词:", words)
}
