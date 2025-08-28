/*
 * Copyright 2020 The SealEVM Authors
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package evmutils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/common"

	"golang.org/x/crypto/sha3"
	"web3_gui/keystore/adapter/crypto"
)

const (
	hashLength    = 32
	AddressLength = 20
)

type Address [AddressLength]byte

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}
func (a *Address) GetBytes() []byte {
	return a[:]
}

var (
	BlankHash   = make([]byte, hashLength)
	ZeroHash    = Keccak256(nil)
	ZeroAddress = common.Address(BytesToAddress(nil))
)

// EVMIntToHashBytes returns the absolute value of x as a big-endian fixed length byte slice.
func EVMIntToHashBytes(i *Int) [hashLength]byte {
	iBytes := i.Bytes()
	iLen := len(iBytes)

	var hash [hashLength]byte
	if iLen > hashLength {
		copy(hash[:], iBytes[iLen-hashLength:])
	} else {
		copy(hash[hashLength-iLen:], iBytes)
	}

	return hash
}

// EthHashBytesToEVMInt EVMIntToHashBytes reverse
func EthHashBytesToEVMInt(hash [hashLength]byte) (*Int, error) {
	return HashBytesToEVMInt(hash[:])
}

// HashBytesToEVMInt byte to bigInt
func HashBytesToEVMInt(hash []byte) (*Int, error) {
	i := New(0)
	i.SetBytes(hash[:])
	return i, nil
}

// BytesDataToEVMIntHash fixed length bytes
func BytesDataToEVMIntHash(data []byte) *Int {
	var hashBytes []byte
	srcLen := len(data)
	if srcLen < hashLength {
		hashBytes = LeftPaddingSlice(data, hashLength)
	} else {
		hashBytes = data[:hashLength]
	}

	i := New(0)
	i.SetBytes(hashBytes)

	return i
}

func GetDataFrom(src []byte, offset uint64, size uint64) []byte {
	ret := make([]byte, size)
	dLen := uint64(len(src))
	if dLen < offset {
		return ret
	}

	end := offset + size
	if dLen < end {
		end = dLen
	}

	copy(ret, src[offset:end])
	return ret
}

func LeftPaddingSlice(src []byte, toSize int) []byte {
	sLen := len(src)
	if toSize <= sLen {
		return src
	}

	ret := make([]byte, toSize)
	copy(ret[toSize-sLen:], src)

	return ret
}

func RightPaddingSlice(src []byte, toSize int) []byte {
	sLen := len(src)
	if toSize <= sLen {
		return src
	}

	ret := make([]byte, toSize)
	copy(ret, src)

	return ret
}

func Keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}
func Keccak256Hash(data []byte) common.Hash {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	h := common.Hash{}
	h.SetBytes(hasher.Sum(nil))
	return h
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

// MakeAddressFromHex any hex str make a evm.Int
func MakeAddressFromHex(str string) (*Int, error) {
	data, err := FromHex(str)
	if err != nil {
		return nil, err
	}
	return MakeAddress(data), nil
}

// MakeAddressFromString any str make a evm.Int
func MakeAddressFromString(str string) (*Int, error) {
	return MakeAddress([]byte(str)), nil
}

// MakeAddress any byte make a evm.Int
func MakeAddress(data []byte) *Int {
	//计算公钥hash
	address := Keccak256(data)

	addr := hex.EncodeToString(address)[24:]
	return FromHexString(addr)
}

// BytesToAddress any byte set to a evm address
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// addresscoin转换成20字节的address
// 使用该方法需要注意空数组处理
func AddressCoinToAddress(b []byte) Address {
	//去掉前3个然后截取20字节即可
	//fmt.Println("前缀：", config.AddrPre)
	// engine.Log.Error("----------------AddressCoinToAddress------------------------")
	// engine.Log.Error(config.AddrPre)
	// engine.Log.Error("----------------AddressCoinToAddress------------------------")

	if config.AddrPre == "" {
		config.AddrPre = "IM"
	}
	var a Address
	start := len(config.AddrPre)
	end := start + 20

	//NOTE：end超出会报错,直接返回空数组
	if end > len(b) {
		return a
	}

	a.SetBytes(b[start:end])
	return a
}

// 20字节地址转换成addresscoin
func AddressToAddressCoin(publicRIPEMD160 []byte) crypto.AddressCoin {
	if config.AddrPre == "" {
		config.AddrPre = "IM"
	}
	pre := config.AddrPre
	buf := bytes.NewBuffer([]byte(pre))
	buf.Write(publicRIPEMD160)
	//第四步，计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(buf.Bytes())
	//第五步，再次计算上一步结果的SHA-256哈希值
	temp = sha256.Sum256(temp[:])
	//第六步，取上一步结果的前4个字节（8位十六进制数）D61967F6，把这4个字节加在第三步结果的后面，作为校验
	bs := make([]byte, 0, len(pre)+len(publicRIPEMD160)+4+1)
	bs = append(bs, pre...)
	bs = append(bs, publicRIPEMD160...)
	bs = append(bs, temp[:4]...)
	preLen := len([]byte(pre))
	bs = append(bs, byte(preLen))
	return bs
}

// StringToAddress any string make a evm address
func StringToAddress(s string) (Address, error) {
	addrInt, err := MakeAddressFromString(s)
	if err != nil {
		var a Address
		return a, err
	}
	return BigToAddress(addrInt), nil
}

// HexToAddress any hex make a evm address
func HexToAddress(s string) (Address, error) {
	addrInt, err := MakeAddressFromHex(s)
	if err != nil {
		var a Address
		return a, err
	}
	return BigToAddress(addrInt), nil
}

// BigToAddress math.bigInt to evm address
func BigToAddress(b *Int) Address {
	return BytesToAddress(b.Bytes())
}

func FromHex(s string) ([]byte, error) {
	if Has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex.DecodeString(s)
}

func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}
func NameHash(inputName string) common.Hash {
	node := common.Hash{}
	if len(inputName) > 0 {
		labels := strings.Split(inputName, ".")
		for i := len(labels) - 1; i >= 0; i-- {
			label := Keccak256Hash([]byte(labels[i]))
			node = Keccak256Hash(append(node.Bytes(), label.Bytes()...))
		}
	}
	return node
}

func Keccak256ByDatas(data ...[]byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	for _, b := range data {
		hasher.Write(b)
	}
	return hasher.Sum(nil)
}
