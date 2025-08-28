package coin_address

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/craftto/go-tron/pkg/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"math/big"
	"web3_gui/keystore/v2/base58"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
	// AddressLengthBase58 is the expected length of the address in base58format
	AddressLengthBase58 = 34
	// TronBytePrefix is the hex prefix to address
	TronBytePrefix = byte(0x41)
)

func init() {
	RegisterCoinTypeAddress(config.TRX, config.COINTYPE_TRX, &AddressTrxBuilder{})
}

type AddressTrxBuilder struct{}

/*
创建地址对象
*/
func (this *AddressTrxBuilder) NewAddress(data []byte) AddressInterface {
	addr := AddressTrx(data)
	return &addr
}

/*
@return    []byte    私钥
@return    []byte    公钥
@return    error     错误
*/
func (this *AddressTrxBuilder) BuildPukAndPrk(key *bip32.Key) (prk []byte, puk []byte, ERR utils.ERROR) {

	prKey, _ := btcec.PrivKeyFromBytes(key.Key)
	prkECDSA := prKey.ToECDSA()
	prk = crypto.FromECDSA(prkECDSA)
	//
	pukECDSA := prkECDSA.PublicKey
	puk = crypto.FromECDSAPub(&pukECDSA)
	//utils.Log.Info().Hex("公钥", puk).Send()
	ERR = utils.NewErrorSuccess()
	return
}

func (this *AddressTrxBuilder) BuildAddress(pre string, pukBs []byte) (addr AddressInterface, ERR utils.ERROR) {
	//计算公钥hash地址
	addrData := PubBsToAddress(pukBs)

	addressTron := make([]byte, 0, 1+AddressLength+4)
	//写入前缀
	addressTron = append(addressTron, TronBytePrefix)
	//写入地址
	addressTron = append(addressTron, addrData[:]...)

	//计算“前缀+地址数据”的校验码
	code := BuildVerificationCode(addressTron)
	//写入校验码
	addressTron = append(addressTron, code...)
	//utils.Log.Info().Hex("trx 公钥转地址", addressTron).Send()
	temp := AddressTrx(addressTron)
	addr = &temp
	//utils.Log.Info().Str("生成的btc地址", addrStr).Send()
	ERR = utils.NewErrorSuccess()
	return
}

func (this *AddressTrxBuilder) GetAddressStr(data []byte) string {
	addr := this.NewAddress(data)
	return string(base58.Encode(addr.Bytes()))
}

type AddressTrx []byte

/*
数据部分
*/
func (this *AddressTrx) Bytes() []byte {
	return *this
}

func (this *AddressTrx) B58String() string {
	return string(base58.Encode(*this))
}

/*
有效数据部分
*/
func (this *AddressTrx) Data() []byte {
	if len(*this) == 0 {
		return nil
	}
	return *this
}

func (this *AddressTrx) String() string {
	return this.B58String()
}

/*
计算验证码
*/
func BuildVerificationCode(bs []byte) []byte {
	h256h0 := sha256.New()
	h256h0.Write(bs)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)
	return h1[:4]
}

// Address represents the 21 byte address of an Tron account.
type Address []byte

// String implements fmt.Stringer.
func (a Address) String() string {
	if a[0] == 0 {
		return new(big.Int).SetBytes(a.Bytes()).String()
	}
	return string(base58.EncodeCheck(a.Bytes()))
	//return a.Base58()
}

// Bytes get bytes from address
func (a Address) Bytes() []byte {
	return a[:]
}

// Hex get bytes from address in string
func (a Address) Hex() string {
	return Bytes2Hex(a[:])
}

// Base58 get base58 encoded string
func (a Address) Base58() string {
	if a[0] == 0 {
		return new(big.Int).SetBytes(a.Bytes()).String()
	}
	return EncodeBase58(a.Bytes())
}

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func Hex2Address(s string) (Address, error) {
	addr, err := common.Hex2Bytes(s)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// Base58ToAddress returns Address with byte values of s.
func Base58ToAddress(s string) (Address, error) {
	addr, err := common.DecodeBase58(s)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func PubBsToAddress(pubBs []byte) [AddressLength]byte {
	bs := crypto.Keccak256(pubBs[1:])[12:]
	if len(bs) > AddressLength {
		bs = bs[len(bs)-AddressLength:]
	}
	a := [AddressLength]byte{}
	copy(a[AddressLength-len(bs):], bs)
	return a
}

// PubkeyToAddress returns address from ecdsa public key
func PubkeyToAddress(p ecdsa.PublicKey) Address {
	address := PubkeyToAddress(p)
	addressTron := make([]byte, 0)
	addressTron = append(addressTron, TronBytePrefix)
	addressTron = append(addressTron, address.Bytes()...)
	return addressTron
}

// Has0xPrefix validates str begins with '0x' or '0X'.
func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Bytes2Hex encodes bytes as a hex string.
func Bytes2Hex(bytes []byte) string {
	encode := make([]byte, len(bytes)*2)
	hex.Encode(encode, bytes)
	return "0x" + string(encode)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func Hex2Bytes(s string) ([]byte, error) {
	if Has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex2Bytes(s)
}

func bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

func hex2Bytes(str string) ([]byte, error) {
	return hex.DecodeString(str)
}

func EncodeBase58(input []byte) string {
	h256h0 := sha256.New()
	h256h0.Write(input)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)

	inputCheck := input
	inputCheck = append(inputCheck, h1[:4]...)

	return Encode(inputCheck, BitcoinAlphabet)
}

func DecodeBase58(input string) ([]byte, error) {
	decodeCheck, err := Decode(input, BitcoinAlphabet)
	if err != nil {
		return nil, err
	}

	if len(decodeCheck) < 4 {
		return nil, fmt.Errorf("b58 check error")
	}

	decodeData := decodeCheck[:len(decodeCheck)-4]

	h256h0 := sha256.New()
	h256h0.Write(decodeData)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)

	if h1[0] == decodeCheck[len(decodeData)] &&
		h1[1] == decodeCheck[len(decodeData)+1] &&
		h1[2] == decodeCheck[len(decodeData)+2] &&
		h1[3] == decodeCheck[len(decodeData)+3] {
		return decodeData, nil
	}
	return nil, fmt.Errorf("b58 check error")
}

/*
解析16进制字符串私钥
*/
func ImportFromPrivateKey16(privateKey16 string) (Address, *ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey16)
	if err != nil {
		return nil, nil, err
	}
	return PubkeyToAddress(privateKeyECDSA.PublicKey), privateKeyECDSA, nil
}
