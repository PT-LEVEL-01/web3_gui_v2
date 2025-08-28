package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"os"
)

const AES_CTR_DEFAULT_KEY = "AES_CTR_DEFAULT_KEY" //默认密码

/*
加密文件
@key               []byte    加密密码
@iv                []byte    加密向量
@plainFilePath     string    待加密的文件路径
@cipherFilePath    string    加密后的文件保存路径
*/
func AesCTR_Encrypt_File(key, iv []byte, plainFilePath, cipherFilePath string) error {
	//判断用户传过来的key和iv是否符合16字节，如果不符合16字节加以处理
	if key == nil || len(key) == 0 {
		key = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	}

	keylen := len(key)
	//if keylen == 0 { //如果用户传入的密钥为空那么就用默认密钥
	//	key = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	//}
	if keylen > 0 && keylen < 16 { //如果密钥长度在0到16之间，那么用0补齐剩余的
		key = append(key, bytes.Repeat([]byte{0}, (16-keylen))...)
	} else if keylen > 16 {
		key = key[:16]
	}
	if iv == nil || len(iv) == 0 {
		iv = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	}
	keylen = len(iv)
	//if keylen == 0 { //如果用户传入的密钥为空那么就用默认密钥
	//	iv = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	//}
	if keylen > 0 && keylen < 16 { //如果密钥长度在0到16之间，那么用0补齐剩余的
		iv = append(iv, bytes.Repeat([]byte{0}, (16-keylen))...)
	} else if keylen > 16 {
		iv = iv[:16]
	}

	//打开待加密文件
	inFile, err := os.Open(plainFilePath)
	if err != nil {
		return err
	}
	defer inFile.Close()
	//创建加密后文件
	outFile, err := os.Create(cipherFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	//1.指定使用的加密aes算法
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	//2.不需要填充,直接获取ctr分组模式的stream
	// 返回一个计数器模式的、底层采用block生成key流的Stream接口，初始向量iv的长度必须等于block的块尺寸。
	stream := cipher.NewCTR(block, iv)
	writer := cipher.StreamWriter{
		S: stream,
		W: outFile,
	}
	// 将文件内容加密后写入新文件
	if _, err := io.Copy(writer, inFile); err != nil {
		return err
	}
	return nil
}

/*
解密文件
*/
func AesCTR_Decrypt_File(key, iv []byte, cipherFilePath, plainFilePath string) error {
	return AesCTR_Encrypt_File(key, iv, cipherFilePath, plainFilePath)
}

/*
加密
@key               []byte    加密密码
@iv                []byte    加密向量
@plainText         []byte    待加密字节
*/
func AesCTR_Encrypt(key, iv, plainText []byte) ([]byte, error) {
	//判断用户传过来的key和iv是否符合16字节，如果不符合16字节加以处理
	if key == nil || len(key) == 0 {
		key = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	}
	keylen := len(key)
	//if keylen == 0 { //如果用户传入的密钥为空那么就用默认密钥
	//	key = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	//}
	if keylen > 0 && keylen < 16 { //如果密钥长度在0到16之间，那么用0补齐剩余的
		key = append(key, bytes.Repeat([]byte{0}, (16-keylen))...)
	} else if keylen > 16 {
		key = key[:16]
	}
	if iv == nil || len(iv) == 0 {
		iv = []byte(AES_CTR_DEFAULT_KEY)
	}
	keylen = len(iv)
	//if keylen == 0 { //如果用户传入的密钥为空那么就用默认密钥
	//	iv = []byte(AES_CTR_DEFAULT_KEY) //默认密钥
	//}
	if keylen > 0 && keylen < 16 { //如果密钥长度在0到16之间，那么用0补齐剩余的
		iv = append(iv, bytes.Repeat([]byte{0}, (16-keylen))...)
	} else if keylen > 16 {
		iv = iv[:16]
	}
	//1.指定使用的加密aes算法
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//2.不需要填充,直接获取ctr分组模式的stream
	// 返回一个计数器模式的、底层采用block生成key流的Stream接口，初始向量iv的长度必须等于block的块尺寸。
	stream := cipher.NewCTR(block, iv)
	//3.加密操作
	cipherText := make([]byte, len(plainText))
	stream.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

/*
解密
@key               []byte    加密密码
@iv                []byte    加密向量
@cipherText         []byte    待解密字节
*/
func AesCTR_Decrypt(key, iv, cipherText []byte) ([]byte, error) {
	return AesCTR_Encrypt(key, iv, cipherText)
}
