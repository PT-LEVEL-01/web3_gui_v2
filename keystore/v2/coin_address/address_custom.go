package coin_address

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/ripemd160"
	"hash"
	"sync"
	"web3_gui/keystore/v2/base58"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

var (
	coinTypeLock         = new(sync.RWMutex)
	coinType_coin uint32 = config.ZeroQuote + 85379471 //自定义币种编码
	//CoinType_net  uint32 = 0x80000000 + 85379471 //自定义币种编码
	coinType_dh uint32 = config.ZeroQuote + 85379472 //自定义币种编码

	coinName_custom = "TEST" //自定义地址前缀
)

func GetCoinTypeCustom() (uint32, string) {
	coinTypeLock.RLock()
	defer coinTypeLock.RUnlock()
	return coinType_coin, coinName_custom
}

func SetCoinTypeCustom(coinType uint32, coinName string) {
	coinTypeLock.Lock()
	defer coinTypeLock.Unlock()
	coinType_coin = coinType
	coinName_custom = coinName
	RegisterCoinTypeAddress(coinName, coinType_coin, &AddressCustomBuilder{})
}

func GetCoinTypeCustomDH() uint32 {
	coinTypeLock.RLock()
	defer coinTypeLock.RUnlock()
	return coinType_dh
}

func SetCoinTypeCustomDH(coinType uint32) {
	coinTypeLock.Lock()
	defer coinTypeLock.Unlock()
	coinType_dh = coinType
	RegisterCoinTypeAddress("", coinType_dh, &AddressCustomDHBuilder{})
}

func init() {
	RegisterCoinTypeAddress(coinName_custom, coinType_coin, &AddressCustomBuilder{})
}

type AddressCustomBuilder struct {
}

/*
创建地址对象
*/
func (this *AddressCustomBuilder) NewAddress(data []byte) AddressInterface {
	addr := AddressCoin(data)
	return &addr
}

/*
@return    []byte    私钥
@return    []byte    公钥
@return    error     错误
*/
func (this *AddressCustomBuilder) BuildPukAndPrk(key *bip32.Key) (prk []byte, puk []byte, ERR utils.ERROR) {
	puk, prk = ParseKey(key)
	ERR = utils.NewErrorSuccess()
	return
}

func (this *AddressCustomBuilder) BuildAddress(addrPre string, puk []byte) (addr AddressInterface, ERR utils.ERROR) {
	var temp AddressCoin
	if config.ShortAddress {
		temp, ERR = BuildAddr_old(addrPre, puk)
	} else {
		temp, ERR = BuildAddr(addrPre, puk)
	}
	if ERR.CheckFail() {
		return nil, ERR
	}
	//temp, ERR := BuildAddr(addrPre, puk)
	//if ERR.CheckFail() {
	//	return nil, ERR
	//}
	//addr = a
	//utils.Log.Info().Str("新地址", temp.String()).Str("pre", addrPre).Send()
	return &temp, ERR
}

func (this *AddressCustomBuilder) GetAddressStr(data []byte) string {
	return string(base58.Encode(data))
}

type AddressCoin []byte

/*
数据部分
*/
func (this *AddressCoin) Bytes() []byte {
	return *this
}

func (this *AddressCoin) B58String() string {
	if len(*this) <= 0 {
		return ""
	}
	lastByte := (*this)[len(*this)-1:]
	lastStr := string(base58.Encode(lastByte))
	if len(lastByte) == 0 {
		return ""
	}
	preLen := int(lastByte[0])
	preStr := string((*this)[:preLen])
	centerStr := string(base58.Encode((*this)[preLen : len(*this)-1]))
	return preStr + centerStr + lastStr
}

func (this *AddressCoin) GetPre() string {
	if len(*this) <= 0 {
		return ""
	}
	lastByte := (*this)[len(*this)-1:]
	if len(lastByte) == 0 {
		return ""
	}
	preLen := int(lastByte[0])
	preStr := string((*this)[:preLen])
	return preStr
}

/*
有效数据部分
*/
func (this *AddressCoin) Data() []byte {
	if len(*this) == 0 {
		return nil
	}
	lastByte := (*this)[len(*this)-1:]
	if len(lastByte) == 0 {
		return nil
	}
	preLen := int(lastByte[0])
	return (*this)[preLen : len(*this)-(4+1)]
}

/*
 */
func (this *AddressCoin) String() string {
	return this.B58String()
}

func ParseKey(bip32Key *bip32.Key) (ed25519.PublicKey, ed25519.PrivateKey) {
	privk := ed25519.NewKeyFromSeed(bip32Key.Key) //私钥
	pub := privk.Public().(ed25519.PublicKey)
	//addr := crypto.BuildAddr(AddrPre, pub) //公钥生成地址
	return pub, privk
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

func Hash256(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), sha256.New())
}

func AddressFromB58String(str string) AddressCoin {
	if str == "" {
		return nil
	}
	lastStr := str[len(str)-1:]
	lastByte := base58.Decode(lastStr)
	if len(lastByte) == 0 {
		return nil
	}
	preLen := int(lastByte[0])
	if preLen > len(str) {
		return nil
	}
	preStr := str[:preLen]
	preByte := []byte(preStr)
	centerByte := base58.Decode(str[preLen : len(str)-1])
	bs := make([]byte, 0, len(preByte)+len(centerByte)+len(lastByte))
	bs = append(bs, preByte...)
	bs = append(bs, centerByte...)
	bs = append(bs, lastByte...)
	return AddressCoin(bs)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) (AddressCoin, utils.ERROR) {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，再次计算SHA-256哈希值
	publicSHA256 = sha256.Sum256(publicSHA256[:])
	//utils.Log.Info().Hex("地址内容", publicSHA256[:]).Send()
	//第三步，在上一步结果之前加入地址版本号（如比特币主网版本号“0x00"）
	buf := bytes.NewBuffer([]byte(pre))
	n, err := buf.Write(publicSHA256[:])
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if n != len(publicSHA256) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	//第四步，计算上一步结果的RIPEMD-160哈希值
	RIPEMD160Hasher := ripemd160.New()
	n, err = RIPEMD160Hasher.Write(buf.Bytes())
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Int("n", n).Int("size", len(publicSHA256)).Send()
	if n != len(buf.Bytes()) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//utils.Log.Info().Hex("hash 160", publicRIPEMD160).Send()
	//temp := sha256.Sum256(buf.Bytes())
	//第五步，再次计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(publicRIPEMD160[:])
	//第六步，取上一步结果的前4个字节（8位十六进制数）D61967F6，把这4个字节加在第三步结果的后面，作为校验
	bs := make([]byte, 0, len(pre)+len(publicSHA256)+4+1)
	bs = append(bs, pre...)
	bs = append(bs, publicSHA256[:]...)
	bs = append(bs, temp[:4]...)
	preLen := len([]byte(pre))
	bs = append(bs, byte(preLen))
	//utils.Log.Info().Hex("地址", bs).Send()
	return bs, utils.NewErrorSuccess()
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddrByData(pre string, data []byte) (AddressCoin, utils.ERROR) {
	//第三步，在上一步结果之前加入地址版本号（如比特币主网版本号“0x00"）
	buf := bytes.NewBuffer([]byte(pre))
	n, err := buf.Write(data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if n != len(data) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	//第四步，计算上一步结果的RIPEMD-160哈希值
	RIPEMD160Hasher := ripemd160.New()
	n, err = RIPEMD160Hasher.Write(buf.Bytes())
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Int("n", n).Int("size", len(publicSHA256)).Send()
	if n != len(buf.Bytes()) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//utils.Log.Info().Hex("hash 160", publicRIPEMD160).Send()
	//temp := sha256.Sum256(buf.Bytes())
	//第五步，再次计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(publicRIPEMD160[:])
	//第六步，取上一步结果的前4个字节（8位十六进制数）D61967F6，把这4个字节加在第三步结果的后面，作为校验
	bs := make([]byte, 0, len(pre)+len(data)+4+1)
	bs = append(bs, pre...)
	bs = append(bs, data[:]...)
	bs = append(bs, temp[:4]...)
	preLen := len([]byte(pre))
	bs = append(bs, byte(preLen))
	//utils.Log.Info().Hex("地址", bs).Send()
	return bs, utils.NewErrorSuccess()
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr_old(pre string, pubKey []byte) (AddressCoin, utils.ERROR) {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，计算RIPEMD-160哈希值
	RIPEMD160Hasher := ripemd160.New()
	n, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Int("n", n).Int("size", len(publicSHA256)).Send()
	if n != len(publicSHA256) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//第三步，在上一步结果之间加入地址版本号（如比特币主网版本号“0x00"）
	buf := bytes.NewBuffer([]byte(pre))
	n, err = buf.Write(publicRIPEMD160)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if n != len(publicRIPEMD160) {
		return nil, utils.NewErrorBus(config.ERROR_code_size_too_small, "")
	}
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
	return bs, utils.NewErrorSuccess()
}

/*
解析前缀
*/
func ParseAddrPrefix(addr AddressCoin) string {
	if len(addr) <= 0 {
		return ""
	}
	lastByte := addr[len(addr)-1:]
	preLen := int(lastByte[0])
	preStr := string(addr[:preLen])
	return preStr
}

func ValidAddrCoin(pre string, addr AddressCoin) bool {
	if config.ShortAddress {
		return ValidAddrCoin_old(pre, addr)
	}
	return ValidLongAddrCoin(pre, addr)
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidLongAddrCoin(pre string, addr AddressCoin) bool {
	if config.COINADDR_version == config.VERSION_v1 {
		return ValidAddrCoin_old(pre, addr)
	}
	//utils.Log.Info().Interface("地址", addr).Str("前缀", pre).Send()
	//判断版本是否正确
	ok := bytes.HasPrefix(addr, []byte(pre))
	if !ok {
		return false
	}
	length := len(addr)
	if length < 5 {
		return false
	}
	preLen := int(addr[length-1])
	preStr := string(addr[:preLen])
	if pre != preStr {
		return false
	}
	//utils.Log.Info().Hex("验证地址 160", addr[:length-4-1]).Send()
	RIPEMD160Hasher := ripemd160.New()
	n, err := RIPEMD160Hasher.Write(addr[:length-4-1])
	if err != nil {
		return false
	}
	if n != length-4-1 {
		return false
	}
	//utils.Log.Info().Str("验证地址", "").Send()
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	//utils.Log.Info().Hex("hash 160", publicRIPEMD160).Send()
	//temp := sha256.Sum256(addr[:length-4-1])
	temp := sha256.Sum256(publicRIPEMD160[:])
	if bytes.Equal(addr[len(addr)-1-4:len(addr)-1], temp[:4]) {
		return true
	}
	//utils.Log.Info().Str("验证地址", "").Send()
	return false
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddrCoin_old(pre string, addr AddressCoin) bool {
	//判断版本是否正确
	ok := bytes.HasPrefix(addr, []byte(pre))
	if !ok {
		return false
	}
	length := len(addr)
	preLen := int(addr[length-1])
	preStr := string(addr[:preLen])
	if pre != preStr {
		return false
	}
	temp := sha256.Sum256(addr[:length-4-1])
	temp = sha256.Sum256(temp[:])
	ok = bytes.HasSuffix(addr[:len(addr)-1], temp[:4])
	if !ok {
		//		fmt.Println("false false")
		return false
	}
	return true
}

/*
检查公钥生成的地址是否一样
@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddrNet(pre string, pubKey []byte, addr AddressCoin) (bool, utils.ERROR) {
	tagAddr, ERR := BuildAddr(pre, pubKey)
	if ERR.CheckFail() {
		return false, ERR
	}
	return bytes.Equal(tagAddr, addr), utils.NewErrorSuccess()
}
