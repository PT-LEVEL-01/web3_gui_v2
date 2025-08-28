package keystore

import (
	"crypto/ed25519"
	"crypto/pbkdf2"
	"crypto/sha256"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
	"path/filepath"
	"sync"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

const (
	MnemonicLang_cn = "cn" //简体中文
	MnemonicLang_en = "en" //英文
)

type Keystore struct {
	wallet        *Wallet            //所属钱包，单独使用时没有所属钱包
	filepath      string             //单独使用时，keystore文件存放路径
	addrPre       string             //地址前缀
	Nickname      string             //密钥库昵称
	CryptedSeed   []byte             //加密保存的种子
	Salt          []byte             //盐
	CheckHash     []byte             //主私钥和链编码加密验证hash值，困难级别
	Rounds        uint64             //pbkdf2迭代次数，困难级别
	CheckHashTemp []byte             //主私钥和链编码加密验证hash值，容易级别
	RoundsTemp    uint64             //pbkdf2迭代次数，容易级别
	lock          *sync.RWMutex      //锁
	MnemonicLang  []string           //助记词语言
	Addrs         []*CoinAddressInfo //已经生成的地址列表
	addrMap       *sync.Map          //key:string=收款地址;value:*CoinAddressInfo=地址密钥等信息;
	NetAddrs      []*CoinAddressInfo //已经生成的网络地址列表
	DHAddrs       []*CoinAddressInfo //DH密钥列表
	AddrsOther    []*CoinAddressInfo //已生成其他币种的地址列表
	store         Store              //存储接口
	Version       uint64             //版本号
}

/*
创建一个单独使用的密钥库
*/
func NewKeystoreSingle(filepath, addrPre string) *Keystore {
	ks := newKeystore(addrPre)
	ks.filepath = filepath
	//ks.addrPre = addrPre
	ks.store = NewStoreFile(filepath)
	return ks
}

func NewKeystoreRand(wallet *Wallet, password string) (*Keystore, utils.ERROR) {
	ks := newKeystore(wallet.addrPre)
	ks.wallet = wallet
	//ks.addrPre = wallet.addrPre
	ERR := ks.initRand(password, password, password, password)
	return ks, ERR
}

func newKeystore(addrPre string) *Keystore {
	ks := Keystore{
		addrPre:      addrPre,                     //
		MnemonicLang: wordlists.English,           //助记词语言
		lock:         new(sync.RWMutex),           //锁
		Addrs:        make([]*CoinAddressInfo, 0), //已经生成的地址列表
		addrMap:      new(sync.Map),               //key:string=收款地址;value:*CoinAddressInfo=地址密钥等信息;
		//pukMap:       new(sync.Map),               //key:string=公钥;value:*CoinAddressInfo=地址密钥等信息;
		NetAddrs:   make([]*CoinAddressInfo, 0), //已经生成的网络地址列表
		DHAddrs:    make([]*CoinAddressInfo, 0), //DH密钥列表
		AddrsOther: make([]*CoinAddressInfo, 0), //已生成其他币种的地址列表
	}
	return &ks
}

/*
使用随机数创建一个新的种子文件
*/
func (this *Keystore) CreateRand(password, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	return this.initRand(password, coinAddrPassword, netAddrPassword, dhPassword)
}

/*
创建一个随机密钥库
会覆盖之前数据
*/
func (this *Keystore) initRand(password, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	dir, _ := filepath.Split(this.GetFilePath())
	if dir != "" {
		//判断文件夹是否存在
		err := utils.CheckCreateDir(dir)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	}
	seedBs, err := utils.Rand16Byte() //随机生成16byte
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//bip39.SetWordList(this.MnemonicLang)
	//mn, err := bip39.NewMnemonic(seedBs[:])
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}

	//utils.Log.Info().Hex("随机数种子", seed[:]).Send()
	ERR := this.initBySeed(seedBs[:], password, coinAddrPassword, netAddrPassword, dhPassword)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
导入助记词，会覆盖原有数据
@param words 助记词
@param pwd 钱包密码
@param firstCoinAddressPassword 首个钱包地址的密码
@param firstAddressPassword 首个网络地址和DHkey的密码
@return error
*/
func (this *Keystore) ImportMnemonic(words, password, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	utils.Log.Info().Str("导入助记词", words).Send()
	bip39.SetWordList(this.MnemonicLang)
	//验证助记词是否合法
	if !bip39.IsMnemonicValid(words) {
		return utils.NewErrorBus(config.ERROR_code_Invalid_mnenomic, "")
	}
	//
	seed, err := bip39.EntropyFromMnemonic(words) //12位助记词  原生的seed 这个seed是没有加密过的
	//seed, err := bip39.NewSeedWithErrorChecking(words, "")
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Hex("助记词还原的原始种子", seed).Send()
	ERR := this.initBySeed(seed, password, coinAddrPassword, netAddrPassword, dhPassword)
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Str("第一个地址", this.GetCoinAddrAll()[0].GetAddrStr()).Send()
	return utils.NewErrorSuccess()
}

/*
导出助记词
@param pwd 钱包密码
@return string 助词记
@return error
*/
func (this *Keystore) ExportMnemonic(pwd string) (string, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.exportMnemonic(pwd)
}

/*
导出助记词
@param pwd 钱包密码
@return string 助词记
@return error
*/
func (this *Keystore) exportMnemonic(pwd string) (string, utils.ERROR) {
	ok, err := this.CheckSeedPassword(pwd)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	if !ok {
		return "", utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	pwdHash, err := pbkdf2.Key(sha256.New, pwd, this.Salt, 1, 32)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	//pwdHash := sha256.Sum256([]byte(pwd))
	//先用密码解密种子
	seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
	if ERR.CheckFail() {
		return "", ERR
	}
	bip39.SetWordList(this.MnemonicLang)
	mn, err := bip39.NewMnemonic(seedBs)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	return mn, utils.NewErrorSuccess()
}

/*
导入助记词，会覆盖原有数据
@param words 助记词
@param pwd 钱包密码
@param firstCoinAddressPassword 首个钱包地址的密码
@param firstAddressPassword 首个网络地址和DHkey的密码
@return error
*/
func (this *Keystore) ImportMnemonicEncry(words, wordsPassword, seedPassword, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	bip39.SetWordList(this.MnemonicLang)
	seedSec, err := bip39.EntropyFromMnemonic(words) //12位助记词  原生的seed 这个seed加密过的
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	var seed []byte
	ERR := utils.NewErrorSuccess()
	if wordsPassword == "" {
		seed = seedSec
	} else {
		pwd, err := pbkdf2.Key(sha256.New, wordsPassword, nil, 1, 32)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		seed, ERR = DecryptCBC(seedSec, pwd, pwd)
		if ERR.CheckFail() {
			return ERR
		}
	}
	ERR = this.initBySeed(seed, seedPassword, coinAddrPassword, netAddrPassword, dhPassword)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
导出助记词
@param pwd 钱包密码
@return string 助词记
@return error
*/
func (this *Keystore) ExportMnemonicEncry(seedPassword, wordsPassword string) (string, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()
	ok, err := this.CheckSeedPassword(seedPassword)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	if !ok {
		return "", utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	pwdHash, err := pbkdf2.Key(sha256.New, seedPassword, this.Salt, 1, 32)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	//pwdHash := sha256.Sum256([]byte(pwd))
	//先用密码解密种子
	seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
	if ERR.CheckFail() {
		return "", ERR
	}
	var seedSec []byte
	if wordsPassword == "" {
		seedSec = seedBs
	} else {
		//给种子加密
		pwdHash, err = pbkdf2.Key(sha256.New, wordsPassword, nil, 1, 32)
		if err != nil {
			return "", utils.NewErrorSysSelf(err)
		}
		seedSec, ERR = EncryptCBC(seedBs, pwdHash, pwdHash)
		if ERR.CheckFail() {
			return "", ERR
		}
	}
	bip39.SetWordList(this.MnemonicLang)
	mn, err := bip39.NewMnemonic(seedSec)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	return mn, utils.NewErrorSuccess()
}

/*
导入助记词，会覆盖原有数据
@param words 助记词
@param password 种子密码
@param coinAddrPassword 钱包地址的密码
@param netAddrPassword 网络地址
@param dhPassword  DHkey的密码
@return utils.ERROR
*/
func (this *Keystore) initBySeed(seed []byte, password, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	//utils.Log.Info().Hex("初始化种子", seed).Send()

	//pwd := sha256.Sum256([]byte(password))
	salt, err := utils.Rand16Byte() //随机生成16byte
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.Salt = salt[:]

	//计算一个需要时间验证的checkHash保存到文件，作为永久存储
	this.Rounds, this.CheckHash, err = config.GetCheckHashForDifficult(password, salt[:])
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//计算一个容易验证checkHash放在内存，软件运行中需要快速的验证
	this.RoundsTemp, this.CheckHashTemp, err = config.GetCheckHashForEasy(password, salt[:])
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//种子加密存储
	pwd, err := pbkdf2.Key(sha256.New, password, salt[:], 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	var ERR utils.ERROR
	this.CryptedSeed, ERR = EncryptCBC(seed, pwd, salt[:])
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Str("通过种子初始化密钥库", "11111").Send()
	_, ERR = this.CreateCoinAddr("", password, coinAddrPassword)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Str("通过种子初始化密钥库", "11111").Send()
	_, ERR = this.CreateNetAddr(password, netAddrPassword)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Str("通过种子初始化密钥库", "11111").Send()
	_, ERR = this.CreateDHKey(password, dhPassword)
	if ERR.CheckFail() {
		return ERR
	}
	this.Version = config.VERSION_v2
	//utils.Log.Info().Str("通过种子初始化密钥库", "11111").Send()
	ERR = this.Save()
	//utils.Log.Info().Str("通过种子初始化密钥库", "11111").Send()
	return ERR
}

func (this *Keystore) GetFilePath() string {
	if this.wallet != nil {
		return this.wallet.filepath
	}
	return this.filepath
}

func (this *Keystore) GetAddrPre() string {
	if this.wallet != nil {
		//utils.Log.Info().Str("地址前缀：%s", this.wallet.addrPre).Send()
		return this.wallet.addrPre
	}
	//utils.Log.Info().Str("地址前缀：%s", this.addrPre).Send()
	return this.addrPre
}

func (this *Keystore) GetWallet() *Wallet {
	return this.wallet
}

/*
检查钱包是否完整
*/
func (this *Keystore) CheckIntact() bool {
	if this.CryptedSeed != nil && len(this.CryptedSeed) > 0 {
		if this.CheckHash == nil || len(this.CheckHash) != 32 {
			// fmt.Println("111111111111========", len(this.CheckHash))
			return false
		}
		return true
	}
	if this.CheckHash == nil || len(this.CheckHash) != 64 {
		// fmt.Println("111111111111========", len(this.CheckHash))
		return false
	}

	return true
}

/*
验证种子密码
@return    bool    密码是否正确.true=正确;false=错误;
*/
func (this *Keystore) CheckSeedPassword(password string) (bool, error) {
	//pwd, err := pbkdf2.Key(sha256.New, password, this.Salt, 1, 32)
	//if err != nil {
	//	return false, err
	//}
	//pwd := sha256.Sum256([]byte(password))
	//utils.Log.Info().Msgf("开始验证密码 111111111:%s", password)
	if this.CheckHashTemp == nil || len(this.CheckHashTemp) == 0 {
		//utils.Log.Info().Msgf("开始验证密码 111111111")
		ok, err := config.ValidateCheckHash(password, this.CheckHash, this.Salt, this.Rounds)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		//密码验证成功
		//计算一个容易验证checkHash放在内存，软件运行中需要快速的验证
		this.RoundsTemp, this.CheckHashTemp, err = config.GetCheckHashForEasy(password, this.Salt)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	//utils.Log.Info().Msgf("开始验证密码 111111111")
	ok, err := config.ValidateCheckHash(password, this.CheckHashTemp, this.Salt, this.RoundsTemp)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

/*
修改种子密码
*/
func (this *Keystore) UpdateSeedPwd(oldpwd, newpwd string) utils.ERROR {
	oldHash, err := pbkdf2.Key(sha256.New, oldpwd, this.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	newHash, err := pbkdf2.Key(sha256.New, newpwd, this.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//oldHash := sha256.Sum256([]byte(oldpwd))
	//newHash := sha256.Sum256([]byte(newpwd))
	this.lock.RLock()
	defer this.lock.RUnlock()
	ok, err := this.CheckSeedPassword(oldpwd)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	//先用密码解密种子
	seedBs, ERR := DecryptCBC(this.CryptedSeed, oldHash[:], this.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	//再用新密码加密种子
	cryptedSeed, ERR := EncryptCBC(seedBs, newHash[:], this.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	checkHash, err := config.GetCheckHashByRounds(newpwd, this.Salt, this.Rounds)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	checkHashTemp, err := config.GetCheckHashByRounds(newpwd, this.Salt, this.RoundsTemp)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.CryptedSeed = cryptedSeed
	this.CheckHash = checkHash
	this.CheckHashTemp = checkHashTemp
	ERR = this.Save()
	return ERR
}

/*
签名
*/
func Sign(prk ed25519.PrivateKey, content []byte) []byte {
	if len(prk) == 0 {
		return nil
	}
	return ed25519.Sign(prk, content)
}

/*
设置助记词语言
*/
func (this *Keystore) SetLang(lan string) {
	if lan == MnemonicLang_cn {
		this.MnemonicLang = wordlists.ChineseSimplified
	}
	if lan == MnemonicLang_en {
		this.MnemonicLang = wordlists.English
	}
}

/*
解密公私钥
@param pwd 钱包密码
*/
func (this *Keystore) Decrypt(pwd string, addrInfo *CoinAddressInfo) (puk []byte, prk []byte, ERR utils.ERROR) {
	ok, err := this.CheckSeedPassword(pwd)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return nil, nil, utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	passwordBs, err := pbkdf2.Key(sha256.New, pwd, this.Salt, 1, 32)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	plantPuk, ERR := DecryptCBC(addrInfo.CryptedPuk, passwordBs, this.Salt)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	plantPrk, ERR := DecryptCBC(addrInfo.CryptedPrk, passwordBs, this.Salt)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	return plantPuk, plantPrk, ERR
}
