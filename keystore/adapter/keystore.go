package keystore

import (
	"errors"
	"github.com/tyler-smith/go-bip39/wordlists"
	"golang.org/x/crypto/ed25519"
	"sync"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

type Keystore struct {
	filepath string                       //keystore文件存放路径
	AddrPre  string                       //
	password string                       //密码
	wallet   *keystore.Wallet             //
	NetAddr  *[]*keystore.CoinAddressInfo //

	Coinbase     uint64         `json:"coinbase"` //当前默认使用的收付款地址
	DHIndex      uint64         `json:"dhindex"`  //DH密钥，指向钱包位置
	lock         *sync.RWMutex  `json:"-"`        //
	MnemonicLang []string       `json:"-"`
	Seed         []byte         `json:"seed"`      //种子
	CheckHash    []byte         `json:"checkhash"` //主私钥和链编码加密验证hash值
	Addrs        []*AddressInfo `json:"addrs"`     //已经生成的地址列表
	//DHKey         []DHKeyPair        `json:"dhkey"`     //DH密钥
	addrMap       *sync.Map          `json:"-"` //key:string=收款地址;value:*AddressInfo=地址密钥等信息;
	pukMap        *sync.Map          `json:"-"` //key:string=公钥;value:*AddressInfo=地址密钥等信息;
	netAddrPrkTmp ed25519.PrivateKey `json:"-"` //网络地址私钥
}

func (this *Keystore) GetV2Keystore() (*keystore.Keystore, utils.ERROR) {
	return this.wallet.GetKeystoreUse()
}

func (this *Keystore) GetKeyByAddr(addr crypto.AddressCoin, password string) (rand []byte, prk ed25519.PrivateKey,
	puk ed25519.PublicKey, err error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, nil, nil, errors.New(ERR.String())
	}
	addrInfo, ERR := keyst.GetCoinAddrInfo(coin_address.AddressCoin(addr), password)
	if ERR.CheckFail() {
		return nil, nil, nil, errors.New(ERR.String())
	}
	puk, prk, ERR = keyst.Decrypt(password, addrInfo)
	//utils.Log.Info().Hex("公钥", puk).Hex("私钥", prk).Send()
	if ERR.CheckFail() {
		return nil, nil, nil, errors.New(ERR.String())
	}
	err = nil
	return
}

//func (this *Keystore) GetNetAddrKeyPair(password string) (puk []byte, prk []byte, ERR utils.ERROR) {
//	keyst, ERR := this.wallet.GetKeystoreUse()
//	if ERR.CheckFail() {
//		return nil, nil, ERR
//	}
//	puk, prk, ERR = keyst.GetNetAddrKeyPair(password)
//	if ERR.CheckSuccess() {
//		//GetPukByAddr方法旧版不传密码，新版需要密码，所以这个地方把密码保存一份
//		this.password = password
//	}
//	return
//}

func (this *Keystore) GetNetAddr(pwd string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	_, ERR = keyst.GetNetAddrInfo(pwd)
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	return nil, nil, nil
}

func (this *Keystore) FindPuk(puk []byte) (addrInfo AddressInfo, ok bool) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return AddressInfo{}, false
	}
	coinAddr, ERR := keystore.BuildAddr(this.wallet.GetAddrPre(), puk)
	if ERR.CheckFail() {
		return AddressInfo{}, false
	}
	addrInfoV2 := keyst.FindCoinAddr(&coinAddr)
	if addrInfoV2 == nil {
		return AddressInfo{}, false
	}
	//addrInfoV2 := keyst.GetCoinAddrInfoByPukNotPwd(puk)
	addrInfo = AddressInfo{
		Index:    uint64(addrInfoV2.Index), //棘轮数量
		Nickname: addrInfoV2.Nickname,      //地址昵称
		Addr:     addrInfoV2.Addr.Bytes(),  //收款地址
		//Puk      :addrInfoV2.CryptedPuk,       //公钥
		//SubKey    []byte             `json:"subKey"`    //子密钥
		//AddrStr   string             `json:"-"`         //
		//PukStr    string             `json:"-"`         //
		//CheckHash []byte             `json:"checkhash"` //主私钥和链编码加密验证hash值
		//Version   int                `json:"version"`   //地址版本
	}
	return addrInfo, true
}

func (this *Keystore) GetCoinbase() *AddressInfo {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil
	}
	addrInfoV2 := keyst.GetCoinAddrAll()[0]
	ok, err := keyst.CheckSeedPassword(this.password)
	if err != nil {
		return nil
	}
	if !ok {
		return nil
	}
	puk, _, ERR := keyst.Decrypt(this.password, &addrInfoV2)
	if ERR.CheckFail() {
		return nil
	}
	addrInfo := AddressInfo{
		Index:    uint64(addrInfoV2.Index), //棘轮数量
		Nickname: addrInfoV2.Nickname,      //地址昵称
		Addr:     addrInfoV2.Addr.Bytes(),  //收款地址
		Puk:      puk,                      //公钥
		//SubKey    []byte             `json:"subKey"`    //子密钥
		//AddrStr   string             `json:"-"`         //
		//PukStr    string             `json:"-"`         //
		//CheckHash []byte             `json:"checkhash"` //主私钥和链编码加密验证hash值
		//Version   int                `json:"version"`   //地址版本
	}
	return &addrInfo
}

func (this *Keystore) FindAddress(addr crypto.AddressCoin) (addrInfo AddressInfo, ok bool) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return AddressInfo{}, false
	}
	a := coin_address.AddressCoin(addr)
	addrInfoV2 := keyst.FindCoinAddr(&a)
	if addrInfoV2 == nil {
		return AddressInfo{}, false
	}
	addrInfo = AddressInfo{
		Index:    uint64(addrInfoV2.Index), //棘轮数量
		Nickname: addrInfoV2.Nickname,      //地址昵称
		Addr:     addrInfoV2.Addr.Bytes(),  //收款地址
		//Puk      :addrInfoV2.CryptedPuk,       //公钥
		//SubKey    []byte             `json:"subKey"`    //子密钥
		//AddrStr   string             `json:"-"`         //
		//PukStr    string             `json:"-"`         //
		//CheckHash []byte             `json:"checkhash"` //主私钥和链编码加密验证hash值
		//Version   int                `json:"version"`   //地址版本
	}
	return addrInfo, true
}

func (this *Keystore) GetAddrAll() []*AddressInfo {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		utils.Log.Info().Str("这里返回了空", "").Send()
		return nil
	}
	coinAddrs := keyst.GetCoinAddrAll()
	if coinAddrs == nil || len(coinAddrs) <= 0 {
		utils.Log.Info().Str("这里返回了空", "").Send()
		return nil
	}
	addrs := make([]*AddressInfo, 0, len(coinAddrs))
	for _, one := range coinAddrs {
		addrInfo := AddressInfo{
			Index:    uint64(one.Index), //棘轮数量
			Nickname: one.Nickname,      //地址昵称
			Addr:     one.Addr.Bytes(),  //收款地址
			//Puk      :addrInfoV2.CryptedPuk,       //公钥
			//SubKey    []byte             `json:"subKey"`    //子密钥
			//AddrStr   string             `json:"-"`         //
			//PukStr    string             `json:"-"`         //
			//CheckHash []byte             `json:"checkhash"` //主私钥和链编码加密验证hash值
			//Version   int                `json:"version"`   //地址版本
		}
		addrs = append(addrs, &addrInfo)
	}
	return addrs
}

func (this *Keystore) GetPukByAddr(addr crypto.AddressCoin) (puk ed25519.PublicKey, ok bool) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, false
	}
	//keyst.GetCoinAddrInfo()
	a := coin_address.AddressCoin(addr)
	addrInfo := keyst.FindCoinAddr(&a)
	if addrInfo == nil {
		return nil, false
	}
	puk, _, ERR = keyst.Decrypt(this.password, addrInfo)
	if ERR.CheckFail() {
		return nil, false
	}
	ok = true
	return
}

func (this *Keystore) GetKeyByPuk(puk []byte, password string) (rand []byte, prk ed25519.PrivateKey, err error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	coinAddr, ERR := keystore.BuildAddr(this.wallet.GetAddrPre(), puk)
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	//utils.Log.Info().Hex("查找公钥", puk).Send()
	//utils.Log.Info().Msgf("前缀:%s 地址:%s", this.wallet.GetAddrPre(), coinAddr.B58String())
	coinAddrInfo := keyst.FindCoinAddr(&coinAddr)
	if coinAddrInfo == nil {
		//若新版地址未找到，则尝试旧版地址
		coinAddr, ERR = keystore.BuildAddr_old(this.wallet.GetAddrPre(), puk)
		if ERR.CheckFail() {
			return nil, nil, errors.New(ERR.String())
		}
		//utils.Log.Info().Msgf("前缀:%s 地址:%s", this.wallet.GetAddrPre(), coinAddr.B58String())
		coinAddrInfo = keyst.FindCoinAddr(&coinAddr)
		if coinAddrInfo == nil {
			return nil, nil, errors.New("puk not exist")
		}
	}
	_, prk, ERR = keyst.Decrypt(password, coinAddrInfo)
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	err = nil
	return
}

func (this *Keystore) GetNewAddrByName(name string, password string, newAddressPassword string) (crypto.AddressCoin, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) UpdatePwd(oldpwd, newpwd string) (ok bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) UpdateNetAddrPwd(oldpwd, newpwd string) (ok bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) UpdateAddrPwd(addr, oldpwd, newpwd string) (ok bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) ImportMnemonic(words, pwd, firstCoinAddressPassword, netAddressAndDHkeyPassword string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) ExportMnemonicEncry(seedPwd, wordsPwd string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) ImportMnemonicEncry(words, wordsPwd, seedPwd, firstCoinAddressPassword, netAddressAndDHkeyPassword string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Keystore) ExportMnemonic(pwd string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewKeystore(filepath, addrPre string) *Keystore {
	keys := Keystore{
		filepath:     filepath,          //keystore文件存放路径
		AddrPre:      addrPre,           //
		lock:         new(sync.RWMutex), //
		MnemonicLang: wordlists.English, //
	}
	return &keys
}

func NewKeystoreItr(wallet *keystore.Wallet, pwd string) (*Keystore, utils.ERROR) {
	keys := Keystore{
		filepath:     wallet.GetFilePath(), //keystore文件存放路径
		AddrPre:      wallet.GetAddrPre(),  //
		password:     pwd,                  //
		wallet:       wallet,               //
		MnemonicLang: wordlists.English,    //
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, ERR
	}
	keys.NetAddr = &keyst.NetAddrs
	return &keys, utils.NewErrorSuccess()
}

func NewKeystoreItrByKeystore(keyst *keystore.Keystore, pwd string) *Keystore {
	//utils.Log.Info().Msgf("地址前缀:%s", keyst.GetAddrPre())
	keys := Keystore{
		filepath:     keyst.GetFilePath(), //keystore文件存放路径
		AddrPre:      keyst.GetAddrPre(),  //
		password:     pwd,                 //
		wallet:       keyst.GetWallet(),   //
		MnemonicLang: wordlists.English,   //
	}
	keys.NetAddr = &keyst.NetAddrs
	return &keys
}

/*
从磁盘文件加载keystore
*/
func (this *Keystore) Load() error {
	wallet := keystore.NewWallet(this.filepath, this.AddrPre)
	ERR := wallet.Load()
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	this.NetAddr = &keyst.NetAddrs
	return nil
}

/*
创建一个新的种子文件
*/
func (this *Keystore) CreateNewKeystore(password string) error {
	wallet := keystore.NewWallet(this.filepath, this.AddrPre)
	ERR := wallet.AddKeystoreRand(password)
	if ERR.CheckFail() {
		return errors.New(ERR.String())
	}
	return nil
}

/*
CreateNetAddr
@Description: 新建网络地址
@receiver this
@param password 钱包的密码
@param netAddressPassword 网络地址的密码
@return prk 地址私钥
@return puk 地址公钥
@return err
*/
func (this *Keystore) CreateNetAddr(password, netAddressPassword string) (prk ed25519.PrivateKey, puk ed25519.PublicKey, err error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	_, ERR = keyst.CreateNetAddr(password, netAddressPassword)
	if ERR.CheckFail() {
		return nil, nil, errors.New(ERR.String())
	}
	return nil, nil, nil
}

/*
获取收款地址列表
*/
func (this *Keystore) GetAddr() (addrs []*AddressInfo) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return
	}
	coinAddrs := keyst.GetCoinAddrAll()
	addrs = make([]*AddressInfo, 0, len(coinAddrs))
	for _, one := range coinAddrs {
		addrInfo := AddressInfo{
			Index:     0,
			Nickname:  one.Nickname,
			Addr:      one.Addr.Bytes(),
			Puk:       nil,
			SubKey:    nil,
			AddrStr:   "",
			PukStr:    "",
			CheckHash: nil,
			Version:   0,
		}
		addrs = append(addrs, &addrInfo)
	}
	return
}

/*
GetNewAddr
@Description: 获取一个新的地址
@receiver this
@param password 钱包密码
@param newAddressPassword 新地址密码
@return crypto.AddressCoin
@return error
*/
func (this *Keystore) GetNewAddr(password, newAddressPassword string) (crypto.AddressCoin, error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	addrOne, ERR := keyst.CreateCoinAddr("", password, newAddressPassword)
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	return crypto.AddressCoin(addrOne), nil
}

/*
获取DH公钥
*/
func (this *Keystore) GetDHKeyPair() DHKeyPair {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return DHKeyPair{}
	}
	dhKey, ERR := keyst.GetDhAddrKeyPair(this.password)
	if ERR.CheckFail() {
		return DHKeyPair{}
	}
	return DHKeyPair{
		KeyPair: *dhKey,
	}
}

/*
GetNewDHKey
@Description: 创建DHKey协商秘钥
@receiver this
@param password 钱包密码
@param newDHKeyPassword DHKey密码
@return *dh.KeyPair
@return error
*/
func (this *Keystore) GetNewDHKey(password string, newDHKeyPassword string) (*keystore.CoinAddressInfo, error) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	addrInfo, ERR := keyst.CreateDHKey(password, newDHKeyPassword)
	if ERR.CheckFail() {
		return nil, errors.New(ERR.String())
	}
	return addrInfo, nil
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

func (this *Keystore) GetNetAddrKeyPair(password string) (puk []byte, prk []byte, ERR utils.ERROR) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	return keyst.GetNetAddrKeyPair(password)
}

func (this *Keystore) GetDhAddrKeyPair(password string) (*keystore.KeyPair, utils.ERROR) {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, ERR
	}
	return keyst.GetDhAddrKeyPair(password)
}

func (this *Keystore) GetCoinAddrAll() []keystore.CoinAddressInfo {
	keyst, ERR := this.wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil
	}
	return keyst.GetCoinAddrAll()
}
