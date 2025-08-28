package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
加密
*/
func EncryptCBC(plantText, key, iv []byte) ([]byte, utils.ERROR) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, utils.NewErrorBus(config.ERROR_code_salt_size_too_small, "")
	}
	block, err := aes.NewCipher(key) //
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, utils.NewErrorSuccess()
}

/*
PKCS #7 填充字符串由一个字节序列组成，每个字节填充该字节序列的长度。
下面的示例演示这些模式的工作原理。假定块长度为 8，数据长度为 9，则填充用八位字节数等于 7，数据等于 FF FF FF FF FF FF FF FF FF：
数据： FF FF FF FF FF FF FF FF FF
PKCS7 填充： FF FF FF FF FF FF FF FF FF 07 07 07 07 07 07 07
*/
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

/*
解密
*/
func DecryptCBC(ciphertext, key, iv []byte) ([]byte, utils.ERROR) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, utils.NewErrorBus(config.ERROR_code_salt_size_too_small, "")
	}
	keyBytes := key
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	return PKCS7UnPadding(plantText)
	// return plantText, nil
}

func PKCS7UnPadding(plantText []byte) ([]byte, utils.ERROR) {
	length := len(plantText)
	if length == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_salt_size_too_small, "") // errors.New("plantText Len is 0")
	}
	unpadding := int(plantText[length-1])
	if unpadding >= length {
		return plantText, utils.NewErrorSuccess()
	}
	//截取填充段
	return plantText[:(length - unpadding)], utils.NewErrorSuccess()
}
