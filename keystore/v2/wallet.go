package keystore

import (
	"sync"
	"sync/atomic"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
钱包中管理着多个密钥
*/
type Wallet struct {
	filepath    string        //keystore文件存放路径
	addrPre     string        //地址前缀
	lock        *sync.RWMutex //
	Keystore    []*Keystore   //保存的多个密钥
	UseKeystore int           //上次使用的密钥库索引
	CheckPwd    *atomic.Bool  //是否
	store       Store         //持久化接口
}

/*
创建一个钱包
*/
func NewWallet(filepath, addrPre string) *Wallet {
	w := Wallet{
		filepath:    filepath,
		addrPre:     addrPre,
		lock:        new(sync.RWMutex),
		Keystore:    make([]*Keystore, 0),
		UseKeystore: 0,
		store:       NewStoreFile(filepath),
	}
	return &w
}

/*
查询本钱包中的密钥库列表
*/
func (this *Wallet) List() []KeystoreIndex {
	kis := make([]KeystoreIndex, 0, len(this.Keystore))
	this.lock.RLock()
	defer this.lock.RUnlock()
	for i, k := range this.Keystore {
		kis = append(kis, KeystoreIndex{Index: i, Nickname: k.Nickname})
	}
	return kis
}

/*
查询本钱包中的密钥库列表
*/
func (this *Wallet) Use(index int, password string) (*Keystore, utils.ERROR) {
	ks, ERR := this.GetKeystore(index, password)
	if ERR.CheckFail() {
		return ks, ERR
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	this.UseKeystore = index
	return ks, ERR
}

func (this *Wallet) GetKeystore(index int, password string) (*Keystore, utils.ERROR) {
	//判断下表是否超出限制
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.Keystore) <= index {
		return nil, utils.NewErrorBus(config.ERROR_code_keystore_index_maximum, "")
	}
	ks := this.Keystore[index]
	ok, err := ks.CheckSeedPassword(password)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	return ks, utils.NewErrorSuccess()
}

/*
获取正在使用中的keystore
*/
func (this *Wallet) GetKeystoreUse() (*Keystore, utils.ERROR) {
	//判断下表是否超出限制
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.Keystore) <= this.UseKeystore {
		return nil, utils.NewErrorBus(config.ERROR_code_keystore_index_maximum, "")
	}
	ks := this.Keystore[this.UseKeystore]
	return ks, utils.NewErrorSuccess()
}

/*
添加一个随机密钥库
*/
func (this *Wallet) AddKeystoreRand(password string) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	ksOne, ERR := NewKeystoreRand(this, password)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	this.Keystore = append(this.Keystore, ksOne)
	return this.Save()
}

/*
导入助记词，并创建一个新的密钥库到钱包
@param words 助记词
@param pwd 钱包密码
@param firstCoinAddressPassword 首个钱包地址的密码
@param firstAddressPassword 首个网络地址和DHkey的密码
@return error
*/
func (this *Wallet) ImportMnemonic(words, password, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	ks := newKeystore(this.addrPre)
	ks.wallet = this
	this.lock.Lock()
	defer this.lock.Unlock()
	this.Keystore = append(this.Keystore, ks)
	return ks.ImportMnemonic(words, password, coinAddrPassword, netAddrPassword, dhPassword)
}

/*
导入助记词，并创建一个新的密钥库到钱包
@param words 助记词
@param pwd 钱包密码
@param firstCoinAddressPassword 首个钱包地址的密码
@param firstAddressPassword 首个网络地址和DHkey的密码
@return error
*/
func (this *Wallet) ImportMnemonicEncry(words, wordsPassword, seedPassword, coinAddrPassword, netAddrPassword, dhPassword string) utils.ERROR {
	ks := newKeystore(this.addrPre)
	ks.wallet = this
	this.lock.Lock()
	defer this.lock.Unlock()
	this.Keystore = append(this.Keystore, ks)
	return ks.ImportMnemonicEncry(words, wordsPassword, seedPassword, coinAddrPassword, netAddrPassword, dhPassword)
}

/*
检查钱包是否完整
*/
func (this *Wallet) CheckIntact() bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	for _, ksOne := range this.Keystore {
		if !ksOne.CheckIntact() {
			return false
		}
	}
	return true
}

func (this *Wallet) GetFilePath() string {
	return this.filepath
}

func (this *Wallet) GetAddrPre() string {
	return this.addrPre
}

type KeystoreIndex struct {
	Index    int    //索引
	Nickname string //昵称
}

func (this *Wallet) SetStore(store Store) {
	this.store = store
}

/*
加载或者初始化一个钱包，并保存到本地
*/
func LoadOrSaveWallet(filePath, addrPre, pwd string) (*Wallet, utils.ERROR) {
	//utils.Log.Info().Str("地址前缀", addrPre).Send()
	w := NewWallet(filePath, addrPre)
	//加载钱包文件
	ERR := w.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == config.ERROR_code_wallet_incomplete {
			utils.Log.Error().Str("钱包文件损坏", filePath).Send()
			return nil, ERR
		} else if ERR.Code == config.ERROR_code_wallet_file_not_exist {
			utils.Log.Error().Str("钱包文件不存在", filePath).Send()
		} else {
			utils.Log.Error().Str("加载钱包文件报错", ERR.String()).Send()
			return nil, ERR
		}
	}
	list := w.List()
	if len(list) == 0 {
		ERR = w.AddKeystoreRand(pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("添加一个随机数钱包报错", ERR.String()).Send()
			return nil, ERR
		}
	}
	keyst, ERR := w.GetKeystoreUse()
	if ERR.CheckFail() {
		utils.Log.Error().Str("使用密钥库错误", ERR.String()).Send()
		return nil, ERR
	}
	//utils.Log.Info().Msgf("地址前缀:%s", keyst.GetAddrPre())
	if len(keyst.GetCoinAddrAll()) == 0 {
		_, ERR = keyst.CreateCoinAddr("", pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	if len(keyst.GetNetAddrAll()) == 0 {
		_, ERR = keyst.CreateNetAddr(pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	if len(keyst.GetDHAddrInfoAll()) == 0 {
		_, ERR = keyst.CreateDHKey(pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	//验证一下密码
	ok, err := keyst.CheckSeedPassword(pwd)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return w, utils.NewErrorBus(config.ERROR_code_password_fail, "")
	}
	//addrList := keyst.GetCoinAddrAll()
	//for _, one := range addrList {
	//	utils.Log.Info().Str("地址列表", one.GetAddrStr()).Send()
	//}
	return w, utils.NewErrorSuccess()
}
