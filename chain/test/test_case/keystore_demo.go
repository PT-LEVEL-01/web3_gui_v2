package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/ripemd160"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// 钱包生成demo
// 钱包生成demo
// 钱包生成demo
// 钱包生成demo
// 钱包生成demo
// 钱包生成demo

// 一个钱包默认一个密码
var Wallet_keystore_default_pwd = "xhy19liu21@"

//收款地址前缀 默认为空
var AddrPre = ""

// 加密盐
var Salt = []byte{53, 111, 103, 103, 87, 66, 54, 103, 53, 108, 65, 81, 73, 53, 70, 43} //加密盐
const (
	version_3 = 3
)

var ERROR_password_fail = errors.New("password fail") //密码错误

// 一个钱包可以有多个地址
type Keystore struct {
	filepath     string        //keystore文件存放路径
	AddrPre      string        //收款地址前缀 默认为空
	Wallets      []*Wallet     `json:"wallets"`  //keystore中的所有钱包
	Coinbase     uint64        `json:"coinbase"` //当前默认使用的收付款地址
	DHIndex      uint64        `json:"dhindex"`  //DH密钥，指向钱包位置
	lock         *sync.RWMutex //
	MnemonicLang []string
}

type AddressCoin []byte

type Wallet struct {
	Seed      []byte         `json:"seed"`      //种子
	Key       []byte         `json:"key"`       //生成主密钥的随机数
	ChainCode []byte         `json:"chaincode"` //主KDF链编码
	IV        []byte         `json:"iv"`        //aes加密向量
	CheckHash []byte         `json:"checkhash"` //主私钥和链编码加密验证hash值 （通过种子 + 密码 + Salt（加密盐） 可以反向接触原始的seed 然后seed sha加密就能获得当前值 用于验证）
	Coinbase  uint64         `json:"coinbase"`  //当前默认使用的收付款地址
	Addrs     []*AddressInfo `json:"addrs"`     //已经生成的地址列表
	DHKey     []DHKeyPair    `json:"dhkey"`     //DH密钥
	Version   uint64         `json:"v"`         //版本号
	lock      *sync.RWMutex  `json:"-"`         //
	addrMap   *sync.Map      `json:"-"`         //key:string=收款地址;value:*AddressInfo=地址密钥等信息;
	pukMap    *sync.Map      `json:"-"`         //key:string=公钥;value:*AddressInfo=地址密钥等信息;
	addrPre   string         `json:"-"`         //
}

type AddressInfo struct {
	Index     uint64            `json:"index"`     //棘轮数量 (当前是第几个地址 默认0开始)
	Key       []byte            `json:"key"`       //密钥的随机数
	ChainCode []byte            `json:"chaincode"` //KDF链编码
	Nickname  string            `json:"nickname"`  //地址昵称
	Addr      AddressCoin       `json:"addr"`      //收款地址
	Puk       ed25519.PublicKey `json:"puk"`       //公钥
	AddrStr   string            `json:"-"`         //
	PukStr    string            `json:"-"`         //
}

/*
	私钥和公钥
*/
type Key [32]byte

/*
	公钥私钥对
*/
type KeyPair struct {
	PublicKey  Key
	PrivateKey Key
}

type DHKeyPair struct {
	Index   uint64  `json:"index"`   //棘轮数量
	KeyPair KeyPair `json:"keypair"` //
}

// 模拟一个钱包的生成过程
func main() {

	//hkdf := hkdf.New(sha256.New, master, salt, nil)
	//hashSeed := make([]byte, 64)

	var priv [5]byte

	//用随机数填满私钥
	buf := bytes.NewBuffer([]byte("hello"))
	n, err := buf.Read(priv[:])
	fmt.Println(n)
	fmt.Println(Bytes2string(priv[:]))
	return

	fmt.Println("111111111")
	fmt.Println(Bytes2string([]byte(Wallet_keystore_default_pwd)))

	// 密码hash值
	pwd := sha256.Sum256([]byte(Wallet_keystore_default_pwd))
	fmt.Printf("%x\n", pwd)
	// 1c0dff2030d50d1e89761cb66c7d7e4f060c80532eeeb29d6d1a09017cf9b4c3

	// 生成一个随机字符串
	seed, err := Rand32Byte()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	seedBs := seed[:]
	fmt.Println(seedBs)
	fmt.Println(fmt.Printf("%x\n", seedBs))
	// 1f0b4b7dd9066657dadc234cd9969c0c272afacd0b74ade6074693b72df6d044

	// 创建一个新的钱包种子
	wallet, err := NewWallet(AddrPre, &seedBs, nil, nil, nil, &pwd)
	fmt.Println(wallet)

	if err != nil {
		return
	}
	wallet.Version = version_3

	// 保存在key里面
	key := Keystore{}

	//key.lock.Lock()
	key.Wallets = append(key.Wallets, wallet)
	//key.lock.Unlock()
	fmt.Printf("%+v\n", key)

	// 最后就是把key转为json存在本地指定路径的文件里面
	//key.Save()

}

/*
	创建一个新的钱包种子
*/
func NewWallet(addrPre string, seed *[]byte, key, code *[32]byte, iv *[16]byte, pwd *[32]byte) (*Wallet, error) {
	// 创建一个钱包结构体
	wallet := Wallet{}
	wallet.addrPre = addrPre
	if seed != nil {

		// 用cbc加密算法进行加密 生成钱包的种子
		seedSec, err := EncryptCBC(*seed, (*pwd)[:], Salt)
		if err != nil {
			return nil, err
		}

		checkHash := sha256.Sum256(*seed)

		wallet.Seed = seedSec

		wallet.CheckHash = checkHash[:]

	}
	//else {
	//
	//	keySec, err := crypto.EncryptCBC(key[:], pwd[:], iv[:])
	//	if err != nil {
	//		return nil, err
	//	}
	//	codeSec, err := crypto.EncryptCBC(code[:], pwd[:], iv[:])
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	hash := sha256.New()
	//	hash.Write(append(key[:], code[:]...))
	//	checkHash := hash.Sum(pwd[:])
	//
	//	wallet.Key = keySec
	//	wallet.ChainCode = codeSec
	//	wallet.IV = iv[:]
	//	wallet.CheckHash = checkHash
	//}
	wallet.Addrs = make([]*AddressInfo, 0)
	wallet.Coinbase = 0
	wallet.DHKey = make([]DHKeyPair, 0)
	wallet.lock = new(sync.RWMutex)
	wallet.addrMap = new(sync.Map)
	wallet.pukMap = new(sync.Map)

	// wallet := Wallet{
	// 	Key:       keySec,                  //生成主密钥的随机数
	// 	ChainCode: codeSec,                 //主KDF链编码
	// 	IV:        iv[:],                   //aes加密向量
	// 	CheckHash: checkHash,               //主私钥和链编码加密验证hash值
	// 	Addrs:     make([]*AddressInfo, 0), //已经生成的地址列表
	// 	Coinbase:  0,                       //当前默认使用的收付款地址
	// 	DHKey:     make([]DHKeyPair, 0),    //dh密钥对
	// 	lock:      new(sync.RWMutex),       //
	// 	addrMap:   new(sync.Map),           //
	// 	pukMap:    new(sync.Map),           //
	// }
	//生成第一个地址
	// if len(wallet.Addrs) == 0 {
	wallet.GetNewAddr(*pwd)
	// }

	//生成第一个DH密钥对 (双方确定对称密钥
	//)
	wallet.GetNewDHKey(*pwd)

	return &wallet, nil
}

/*
	生成dh key
*/
func (this *Wallet) GetNewDHKey(password [32]byte) (*KeyPair, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	ok, key, code, err := this.Decrypt(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ERROR_password_fail
	}

	//查找用过的最高的棘轮数量
	index := uint64(0)
	if len(this.Addrs) > 0 {
		addrInfo := this.Addrs[len(this.Addrs)-1]
		index = addrInfo.Index
	}
	if len(this.DHKey) > 0 {
		dhKey := this.DHKey[len(this.DHKey)-1]
		if index < dhKey.Index {
			index = dhKey.Index
		}
	}
	index = index + 1

	if this.Seed != nil && len(this.Seed) > 0 {
		//密码验证通过，生成新的地址
		var keyNew *[]byte
		var err error
		// if this.Version == version_3 {
		// 	keyNew, _, err = crypto.HkdfChainCodeNewV3(key, code, index)
		// } else {
		keyNew, _, err = HkdfChainCodeNew(key, code, index)
		// }
		// keyNew, _, err := crypto.HkdfChainCodeNew(key, code, index)
		if err != nil {
			return nil, err
		}
		key = *keyNew
	} else {
		//密码验证通过，生成新的地址
		key, _, err = GetHkdfChainCode(key, code, index)
		if err != nil {
			return nil, err
		}
	}

	keyPair, err := GenerateKeyPair(key)
	if err != nil {
		return nil, err
	}
	dhKey := DHKeyPair{
		Index:   index,
		KeyPair: keyPair,
	}
	this.DHKey = append(this.DHKey, dhKey)
	return &keyPair, nil
}

/*
	生成公钥私钥对
*/
func GenerateKeyPair(rand []byte) (KeyPair, error) {
	size := 32
	if len(rand) != size {
		//私钥长度不够
		return KeyPair{}, errors.New("Insufficient length of private key")
	}
	var priv [32]byte

	//用随机数填满私钥
	buf := bytes.NewBuffer(rand)
	n, err := buf.Read(priv[:])
	if n != size {
		//读取私钥长度不够
		return KeyPair{}, errors.New("Insufficient length to read private key")
	}
	if err != nil {
		return KeyPair{}, err
	}

	// _, err := rand.Reader.Read(priv[:])
	// if err != nil {
	// 	return KeyPair{}, err
	// }

	// priv[0] &= 248
	// priv[31] &= 127
	// priv[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &priv)

	return KeyPair{
		PrivateKey: priv,
		PublicKey:  pubKey,
	}, nil

}

func hkdfChainCode(master, salt []byte) (key, chainCode []byte, err error) {
	hkdf := hkdf.New(sha256.New, master, salt, nil)
	keys := make([][]byte, 2)
	for i := 0; i < len(keys); i++ {
		keys[i] = make([]byte, 32)
		n, err := io.ReadFull(hkdf, keys[i])
		if n != len(keys[i]) {
			return nil, nil, errors.New("hkdf chain read hash fail")
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return keys[0], keys[1], nil
}

/*
	获取hkdf链编码
	@master    []byte    随机数
	@salt      []byte    盐
	@index     uint64    索引，棘轮数
*/
func GetHkdfChainCode(master, salt []byte, index uint64) (key, chainCode []byte, err error) {
	key, chainCode = master, salt
	for i := 0; i <= int(index); i++ {
		key, chainCode, err = hkdfChainCode(key, chainCode)
		if err != nil {
			return nil, nil, err
		}
		// fmt.Println("----", len(key), len(chainCode))
	}
	return
}

/*
	生成32字节（256位）的随机数
*/
func Rand32Byte() ([32]byte, error) {
	k := [32]byte{}
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return k, err
	}
	for i, _ := range k {
		k[i] = key[i]
	}
	return k, nil
}

// cbc加密算法代码
func EncryptCBC(plantText, key, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, errors.New("VI length error(" + strconv.Itoa(len(iv)) + ")，aes cbc IV length should be " + strconv.Itoa(aes.BlockSize))
	}
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
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
func DecryptCBC(ciphertext, key, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, errors.New("VI length error(" + strconv.Itoa(len(iv)) + ")，aes cbc IV length should be" + strconv.Itoa(aes.BlockSize))
	}
	keyBytes := key
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	return PKCS7UnPadding(plantText, block.BlockSize())
	// return plantText, nil
}

func PKCS7UnPadding(plantText []byte, blockSize int) ([]byte, error) {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	if unpadding >= length {
		return plantText, nil
	}
	//截取填充段
	return plantText[:(length - unpadding)], nil
}

/*
	通过种子生成key和chainCode
*/
func BuildKeyBySeed(seed *[]byte, salt []byte) (*[]byte, *[]byte, error) {
	hash := sha256.New

	key := [32]byte{}
	hkdf := hkdf.New(hash, *seed, salt, nil)
	_, err := io.ReadFull(hkdf, key[:])
	if err != nil {
		return nil, nil, err
	}
	code := [32]byte{}
	_, err = io.ReadFull(hkdf, code[:])
	if err != nil {
		return nil, nil, err
	}
	keyNew := key[:]
	codeNew := code[:]
	return &keyNew, &codeNew, nil
}

/*
	通过公钥生成地址
	@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) AddressCoin {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，计算RIPEMD-160哈希值
	RIPEMD160Hasher := ripemd160.New()
	RIPEMD160Hasher.Write(publicSHA256[:])

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//第三步，在上一步结果之间加入地址版本号（如比特币主网版本号“0x00"）
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

	// buf = bytes.NewBuffer([]byte(pre))
	// buf.Write(publicRIPEMD160)
	// buf.Write(temp[:4])
	// preLen := len([]byte(pre))
	// buf.WriteByte(byte(preLen))

	// return buf.Bytes()

}

/*
	获取hkdf链编码
	@master    []byte    随机数
	@salt      []byte    盐
	@index     uint64    索引，棘轮数
*/
func HkdfChainCodeNewV3(master, salt []byte, index uint64) (*[]byte, *[]byte, error) {
	// fmt.Println("获取hkdf链编码:", hex.EncodeToString(master), hex.EncodeToString(salt), index)

	hkdf := hkdf.New(sha256.New, master, salt, Uint64ToBytes(index))
	hashSeed := make([]byte, 64)
	// fmt.Println("for获取hkdf链编码:", i)
	n, err := io.ReadFull(hkdf, hashSeed)
	if n != len(hashSeed) {
		// fmt.Println("hkdf read error:", n)
		return nil, nil, errors.New("hkdf chain read hash fail")
	}
	if err != nil {
		return nil, nil, err
	}

	key := hashSeed[:32]
	chainCode := hashSeed[32:]
	return &key, &chainCode, nil
}

//uint64转byte
func Uint64ToBytes(n uint64) []byte {
	bytesBuffer := bytes.NewBuffer(nil)
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

/*
	获取hkdf链编码
	@master    []byte    随机数
	@salt      []byte    盐
	@index     uint64    索引，棘轮数
*/
func HkdfChainCodeNew(master, salt []byte, index uint64) (*[]byte, *[]byte, error) {
	if index > 100 {
		return HkdfChainCodeNewV3(master, salt, index)
	}
	// fmt.Println("获取hkdf链编码:", hex.EncodeToString(master), hex.EncodeToString(salt), index)
	hkdf := hkdf.New(sha256.New, master, salt, nil)
	hashSeed := make([]byte, 64)
	for i := uint64(0); i < index; i++ {
		// fmt.Println("for获取hkdf链编码:", i)
		n, err := io.ReadFull(hkdf, hashSeed)
		if n != len(hashSeed) {
			fmt.Println("hkdf read error:", n)
			return nil, nil, errors.New("hkdf chain read hash fail")
		}
		if err != nil {
			return nil, nil, err
		}
	}

	key := hashSeed[:32]
	chainCode := hashSeed[32:]
	return &key, &chainCode, nil
}

//
/*
	生成一个新的地址，需要密码
*/
func (this *Wallet) GetNewAddr(password [32]byte) (AddressCoin, error) {
	// fmt.Println("创建新地址  <---------------------")
	this.lock.Lock()
	defer this.lock.Unlock()

	//验证密码是否正确
	ok, keyRoot, codeRoot, err := this.Decrypt(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ERROR_password_fail
	}
	//密码验证通过

	key, code := keyRoot, codeRoot

	//查找用过的最高的棘轮数量
	addrIndex := uint64(0)
	if len(this.Addrs) > 0 {
		addrInfo := this.Addrs[len(this.Addrs)-1]
		addrIndex = addrInfo.Index
		key = addrInfo.Key
		code = addrInfo.ChainCode
	}
	dhIndex := uint64(0)
	if len(this.DHKey) > 0 {
		dhKey := this.DHKey[len(this.DHKey)-1]
		dhIndex = dhKey.Index
	}
	index := addrIndex
	if index < dhIndex {
		index = dhIndex
	}
	index = index + 1

	if this.Seed != nil && len(this.Seed) > 0 {
		// fmt.Println("新版本生成地址")
		//密码验证通过，生成新的地址
		var keyNew *[]byte
		var err error
		// if this.Version == version_3 {
		// 	keyNew, _, err = crypto.HkdfChainCodeNewV3(keyRoot, codeRoot, index)
		// } else {
		keyNew, _, err = HkdfChainCodeNew(keyRoot, codeRoot, index)
		// }
		if err != nil {
			return nil, err
		}
		// key = *keyNew
		buf := bytes.NewBuffer(*keyNew)
		puk, _, err := ed25519.GenerateKey(buf)
		if err != nil {
			return nil, err
		}

		addr := BuildAddr(this.addrPre, puk)

		//
		// keySec, err := crypto.EncryptCBC(key, password[:], this.IV)
		// if err != nil {
		// 	return nil, err
		// }
		// codeSec, err := crypto.EncryptCBC(code, password[:], this.IV)
		// if err != nil {
		// 	return nil, err
		// }

		addrInfo := &AddressInfo{
			Index: index, //棘轮数
			// Key:       keySec,  //密钥的随机数
			// ChainCode: codeSec, //KDF链编码
			Addr: addr, //收款地址
			Puk:  puk,  //公钥
		}
		// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
		// fmt.Println("保存PUK", hex.EncodeToString(puk))
		this.Addrs = append(this.Addrs, addrInfo)
		this.addrMap.Store(Bytes2string(addr), addrInfo)
		this.pukMap.Store(Bytes2string(puk), addrInfo)

		return addr, nil

	} else {
		// 旧的
		fmt.Println(key)
		fmt.Println(code)
		addr := AddressCoin{}
		return addr, nil
	}
}

/*
	把[]byte转换为string，只要性能，不在乎可读性
*/
func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

/*
	使用密码解密种子，获得私钥和链编码
	@return    ok    bool    密码是否正确
	@return    key   []byte  生成私钥的随机数
	@return    code  []byte  链编码
*/
func (this *Wallet) Decrypt(pwdbs [32]byte) (ok bool, key, code []byte, err error) {
	//密码取hash

	if this.Seed != nil && len(this.Seed) > 0 && (this.Key == nil || len(this.Key) <= 0) {
		//先用密码解密种子
		seedBs, err := DecryptCBC(this.Seed, pwdbs[:], Salt)
		if err != nil {
			return false, nil, nil, err
		}
		//判断密码是否正确
		chackHash := sha256.Sum256(seedBs)
		if !bytes.Equal(chackHash[:], this.CheckHash) {
			return false, nil, nil, ERROR_password_fail
		}

		// hash := sha256.New
		// key := &[32]byte{}
		// hkdf := hkdf.New(hash, seedBs, salt, nil)
		// _, err = io.ReadFull(hkdf, key[:])
		// if err != nil {
		// 	return false, nil, nil, err
		// }
		// code := &[32]byte{}
		// _, err = io.ReadFull(hkdf, code[:])
		// if err != nil {
		// 	return false, nil, nil, err
		// }
		key, code, err := BuildKeyBySeed(&seedBs, Salt)
		if err != nil {
			return false, nil, nil, err
		}

		// keySec, err := crypto.EncryptCBC(key[:], pwdbs[:], salt)
		// if err != nil {
		// 	return false, nil, nil, err
		// }
		// codeSec, err := crypto.EncryptCBC(code[:], pwdbs[:], salt)
		// if err != nil {
		// 	return false, nil, nil, err
		// }

		// this.Key = keySec
		// this.ChainCode = codeSec
		// this.IV = salt
		return true, *key, *code, nil
	}

	//先用密码解密key和链编码
	keyBs, err := DecryptCBC(this.Key, pwdbs[:], this.IV)
	if err != nil {
		return false, nil, nil, ERROR_password_fail
	}
	codeBs, err := DecryptCBC(this.ChainCode, pwdbs[:], this.IV)
	if err != nil {
		return false, nil, nil, ERROR_password_fail
	}

	//验证密码是否正确
	checkHash := append(keyBs, codeBs...)
	h := sha256.New()
	n, err := h.Write(checkHash)
	if n != len(checkHash) {
		//hash 写入失败
		return false, nil, nil, errors.New("hash Write failure")
	}
	if err != nil {
		return false, nil, nil, err
	}
	checkHash = h.Sum(pwdbs[:])
	// checkHash = sha256.Sum256(checkHash)[:]
	if !bytes.Equal(checkHash, this.CheckHash) {
		return false, nil, nil, nil
	}
	return true, keyBs, codeBs, nil
}

/******************************************************文件和文件夹操作相关**********************************/
/*
	保存文件
	保存文件步骤：
	1.创建临时文件
	2.
*/
func SaveFile(name string, bs *[]byte) error {
	CheckCreateDir(name)
	//创建临时文件
	now := strconv.Itoa(int(time.Now().Unix()))
	tempname := name + "." + now
	file, err := os.OpenFile(tempname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		file.Close()
		return err
	}
	_, err = file.Write(*bs)
	if err != nil {
		file.Close()
		return err
	}
	file.Close()
	//删除旧文件
	ok, err := PathExists(name)
	if err != nil {
		// engine.Log.Info("删除旧文件失败", err)
		return err
	}
	if ok {
		err = os.Remove(name)
		if err != nil {
			return err
		}
	}

	//重命名文件
	err = os.Rename(tempname, name)
	if err != nil {
		return err
	}
	return nil
}

/*
	判断一个路径的文件是否存在
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
	检查目录是否存在，不存在则创建
*/
func CheckCreateDir(dir_path string) {
	if ok, err := PathExists(dir_path); err == nil && !ok {
		Mkdir(dir_path)
	}
}

/*
	递归创建目录
*/
func Mkdir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	//	err := os.Mkdir(path, os.ModeDir)
	if err != nil {
		//		fmt.Println("创建文件夹失败", path, err)
		return err
	}
	return nil
}

/******************************************************KeyStore相关**********************************/
/*
	从磁盘文件加载keystore
*/
func (this *Keystore) Save() error {
	// engine.Log.Info("v%", this.Wallets)

	newWallets := make([]*Wallet, 0)
	for _, one := range this.Wallets {
		walletOne := Wallet{
			Seed:      one.Seed,      //种子
			Key:       one.Key,       //生成主密钥的随机数
			ChainCode: one.ChainCode, //主KDF链编码
			IV:        one.IV,        //aes加密向量
			CheckHash: one.CheckHash, //主私钥和链编码加密验证hash值
			Coinbase:  one.Coinbase,  //当前默认使用的收付款地址
			Addrs:     one.Addrs,     //已经生成的地址列表
			DHKey:     one.DHKey,     //DH密钥
		}
		if one.Seed != nil && len(one.Seed) > 0 {
			walletOne.Key = nil
			walletOne.ChainCode = nil
		} else {
			walletOne.Seed = nil
		}
		newWallets = append(newWallets, &walletOne)
	}

	bs, err := json.Marshal(newWallets)
	if err != nil {
		return err
	}
	// fmt.Println(string(bs))
	// engine.Log.Info(string(bs))
	return SaveFile(this.filepath, &bs)
}
