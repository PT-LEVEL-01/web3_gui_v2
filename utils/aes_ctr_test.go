package utils

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"
)

func TestAesCtr(*testing.T) {
	//fileEncrypt()
	//bytesEncryptAndDecrypt()
}

/*
文件加密
*/
func fileEncrypt() {
	//待加密的文件路径
	plainFilePath := "D:/迅雷下载/7z2107-x64.exe"
	cipherDirPath := "D:\\test\\temp"
	//计算文件hash
	hashBs, err := FileSHA3_256(plainFilePath)
	if err != nil {
		panic("计算文件hash 错误:" + err.Error())
	}
	newFileName := hex.EncodeToString(hashBs)
	fmt.Println("加密前文件hash", newFileName)
	cipherFilePath := filepath.Join(cipherDirPath, newFileName)
	key := []byte("fjadsklfjdklafjkdls;afa")
	iv := []byte("gurefdjafhduafdksoafhjdks")
	err = AesCTR_Encrypt_File(key, iv, plainFilePath, cipherFilePath)
	if err != nil {
		panic("加密文件 错误:" + err.Error())
	}
	newHashBs, err := FileSHA3_256(cipherFilePath)
	if err != nil {
		panic("计算文件hash 错误:" + err.Error())
	}
	fmt.Println("加密后的文件hash", hex.EncodeToString(newHashBs))

	//开始解密
	fileName := filepath.Base(plainFilePath)
	newFilePath := filepath.Join(cipherDirPath, fileName)
	fileDecrypt(key, iv, cipherFilePath, newFilePath)
}

/*
文件解密
*/
func fileDecrypt(key, iv []byte, cipherDirPath, plainFilePath string) {
	//计算文件hash
	hashBs, err := FileSHA3_256(cipherDirPath)
	if err != nil {
		panic("计算文件hash 错误:" + err.Error())
	}
	newFileName := hex.EncodeToString(hashBs)
	fmt.Println("解密前文件hash", newFileName)

	err = AesCTR_Decrypt_File(key, iv, cipherDirPath, plainFilePath)
	if err != nil {
		panic("解密文件 错误:" + err.Error())
	}
	newHashBs, err := FileSHA3_256(plainFilePath)
	if err != nil {
		panic("计算文件hash 错误:" + err.Error())
	}
	fmt.Println("解密后的文件hash", hex.EncodeToString(newHashBs))
}

/*
字节加密
*/
func bytesEncryptAndDecrypt() {
	key := []byte("fdasfasd")
	iv := []byte("jutfndhhsd")
	plainText := []byte("hello bob!")
	fmt.Println("加密前字节:", hex.EncodeToString(plainText))
	cipherText, err := AesCTR_Decrypt(key, iv, plainText)
	if err != nil {
		panic("加密 错误:" + err.Error())
	}
	fmt.Println("加密后字节:", hex.EncodeToString(cipherText))

	//开始解密
	plainText, err = AesCTR_Decrypt(key, iv, cipherText)
	if err != nil {
		panic("解密 错误:" + err.Error())
	}
	fmt.Println("解密后字节:", hex.EncodeToString(plainText))
}
