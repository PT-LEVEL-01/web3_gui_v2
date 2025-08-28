package config

import (
	"bytes"
	"crypto/pbkdf2"
	"crypto/sha256"
	btcecv2 "github.com/btcsuite/btcd/btcec/v2"
	"math/big"
	"time"
	"web3_gui/utils"
)

const (
	VERSION_v1 = uint64(1) //旧地址升级而来
	VERSION_v2 = uint64(2) //新创建的钱包

	version_1     = 1
	roundsDefault = 10000
)

const (
	Address_string_encoding_type_custom = 1 //自定义编码
	Address_string_encoding_type_base58 = 2 //base58编码地址
	Address_string_encoding_type_hex16  = 3 //16进制编码地址
)

var (
	ShortAddress = false //是否使用ripemd160生成的短地址
)

/*
计算时间超过1秒钟的hash值
*/
func GetCheckHashForDifficult(plantText string, salt []byte) (rounds uint64, bs []byte, err error) {
	rounds = roundsDefault
	start := time.Now()
	bs, err = GetCheckHashByRounds(plantText, salt, roundsDefault)
	end := time.Now()
	x := end.Sub(start)
	if x == 0 {
		x = 1
	}
	if time.Second > x {
		rounds = uint64((time.Second/x)+1) * roundsDefault
		bs, err = GetCheckHashByRounds(plantText, salt, rounds)
	}
	return
}

/*
计算时间超过1秒钟的hash值
*/
func GetCheckHashForEasy(plantText string, salt []byte) (rounds uint64, bs []byte, err error) {
	rounds = uint64(1)
	bs, err = GetCheckHashByRounds(plantText, salt, rounds)
	return
}

/*
计算指定round次数的checkHash
*/
func GetCheckHashByRounds(plantText string, salt []byte, rounds uint64) ([]byte, error) {
	return pbkdf2.Key(sha256.New, plantText, salt, int(rounds), 32)
}

/*
验证密码是否正确
*/
func ValidateCheckHash(passwordHash string, checkHash []byte, salt []byte, rounds uint64) (bool, error) {
	bs, err := GetCheckHashByRounds(passwordHash, salt, rounds)
	if err != nil {
		return false, err
	}
	if bytes.Equal(bs, checkHash) {
		return true, nil
	}
	return false, nil
}

/*
检查种子是否可用
*/
func ValidateUnusableSeed(secretKey []byte) utils.ERROR {
	secretKeyNum := new(big.Int).SetBytes(secretKey)
	if len(secretKey) != 32 || secretKeyNum.Cmp(btcecv2.S256().N) >= 0 || secretKeyNum.Sign() == 0 {
		return utils.NewErrorBus(ERROR_code_unusable_seed, "")
	}
	return utils.NewErrorSuccess()
}
