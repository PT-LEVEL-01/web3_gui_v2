package keystore

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
创建一个网络地址
@param password 种子密码
@param netAddrPassword 网络地址的密码
@return prk 地址私钥
@return puk 地址公钥
@return err
*/
func (this *Keystore) CreateNetAddr_old(password, netAddrPassword string) (*CoinAddressInfo, utils.ERROR) {
	//dhPasswordBs := sha256.Sum256([]byte(netAddrPassword)) //密码
	dhPasswordBs, err := pbkdf2.Key(sha256.New, netAddrPassword, this.Salt, 1, 32)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	checkHash, err := config.GetCheckHashByRounds(netAddrPassword, this.Salt, this.Rounds)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	checkHashTemp, err := config.GetCheckHashByRounds(netAddrPassword, this.Salt, this.RoundsTemp)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	ok, err := this.CheckSeedPassword(password)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	//密码验证通过
	//查找用过的最高的棘轮数量
	index := uint32(0)
	if len(this.NetAddrs) > 0 {
		addrInfo := this.NetAddrs[len(this.NetAddrs)-1]
		index = addrInfo.Index + 1
	}
	pwdHash, err := pbkdf2.Key(sha256.New, password, this.Salt, 1, 32)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//先用密码解密种子
	seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//通过bip44推导出公钥.
	km, err := NewKeyTreeGeneratorBySeed(seedBs, password, this.MnemonicLang)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	coinType, _ := coin_address.GetCoinTypeCustom()
	//通过原生助记词推出私钥
	_, key, err := km.GetIndexKey(config.PurposeBIP44, coinType, 0, 0, index)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//根据私钥推导公钥生成地址.
	puk, prk := coin_address.ParseKey(key)
	pukCrypted, ERR := EncryptCBC(puk, dhPasswordBs[:], this.Salt) //对子地址加密码
	if ERR.CheckFail() {
		return nil, ERR
	}
	prkCrypted, ERR := EncryptCBC(prk, dhPasswordBs[:], this.Salt) //对子地址加密码
	if ERR.CheckFail() {
		return nil, ERR
	}
	addrInfo := &CoinAddressInfo{
		keystore: this,
		Purpose:  config.PurposeBIP44, //固定为 44'，表示 BIP-44 标准。
		CoinType: coinType,            //指定加密货币的类型（例如，0' 表示 Bitcoin，2' 表示 Litecoin 等）。
		Account:  config.ZeroQuote,    //表示账户索引，允许用户在同一钱包中有多个账户。
		Change:   0,                   //表示地址类型（0 表示接收地址，1 表示找零地址）。
		Index:    index,               //表示地址索引，用于生成每个账户的具体地址。
		//Addr:          addr,                      //收款地址
		CryptedPuk:    pukCrypted,    //公钥
		CryptedPrk:    prkCrypted,    //
		CheckHash:     checkHash,     //
		CheckHashTemp: checkHashTemp, //
		//SubKey:    subKeySec,                 //
	}
	// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
	// fmt.Println("保存PUK", hex.EncodeToString(puk))
	this.NetAddrs = append(this.NetAddrs, addrInfo)
	//this.addrMap.Store(utils.Bytes2string(addr), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return nil, ERR
	}
	return addrInfo, utils.NewErrorSysSelf(err)
}

/*
创建一个网络地址
@param password 种子密码
@param netAddrPassword 网络地址的密码
@return prk 地址私钥
@return puk 地址公钥
@return err
*/
func (this *Keystore) CreateNetAddr(seedPassword, coinAddrPassword string) (coin_address.AddressCoin, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()
	//查找用过的最高的棘轮数量
	index := uint32(0)
	if len(this.NetAddrs) > 0 {
		addrInfo := this.NetAddrs[len(this.NetAddrs)-1]
		index = addrInfo.Index + 1
	}
	ctafi := &coin_address.AddressCustomBuilder{}
	coinType, _ := coin_address.GetCoinTypeCustom()
	addrInfo, ERR := this.buildCoinAddrByCoinType(ctafi, "", seedPassword, coinAddrPassword, coinType, index)
	if ERR.CheckFail() {
		return nil, ERR
	}
	addr := coin_address.AddressCoin(addrInfo.Addr.Bytes())
	// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
	//this.Addrs = append(this.Addrs, addrInfo)
	//utils.Log.Info().Hex("地址key", addr.Data()).Send()
	this.NetAddrs = append(this.NetAddrs, addrInfo)
	//this.addrMap.Store(utils.Bytes2string(addr.Data()), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return addr, ERR
}

/*
修改网络地址密码
*/
func (this *Keystore) UpdateNetAddrPwd(oldPassword, newPassword string) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if len(this.NetAddrs) == 0 {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.NetAddrs[len(this.NetAddrs)-1]
	ok, ERR := addrInfo.CheckPassword(oldPassword)
	if ERR.CheckFail() {
		return ERR
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_netAddr_password_fail, "")
	}
	ERR = addrInfo.UpdatePassword(oldPassword, newPassword)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.Save()
	return ERR
}

/*
获取地址列表，包括导入的钱包地址
*/
func (this *Keystore) GetNetAddrAll() []CoinAddressInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrInfoList := make([]CoinAddressInfo, 0, len(this.NetAddrs))
	for _, one := range this.NetAddrs {
		addrInfoList = append(addrInfoList, *one)
	}
	return addrInfoList
}

/*
获取网络地址
当有多个时，获取最后一个
@param password 网络地址的密码
@return prk 地址私钥
@return puk 地址公钥
@return err
*/
func (this *Keystore) GetNetAddrInfo(password string) (*CoinAddressInfo, utils.ERROR) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if len(this.NetAddrs) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.NetAddrs[len(this.NetAddrs)-1]
	ok, ERR := addrInfo.CheckPassword(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	if !ok {
		//密码错误
		return nil, utils.NewErrorBus(config.ERROR_code_netAddr_password_fail, "")
	}
	//密码正确
	//if addrInfo.Addr == nil || len(addrInfo.Addr) == 0 {
	//	pwdBs, err := pbkdf2.Key(sha256.New, password, this.Salt, 1, 32)
	//	if err != nil {
	//		return nil, utils.NewErrorSysSelf(err)
	//	}
	//	//解密公钥
	//	puk, ERR := DecryptCBC(addrInfo.CryptedPuk, pwdBs[:], this.Salt)
	//	if ERR.CheckFail() {
	//		return nil, ERR
	//	}
	//	bs := sha256.Sum256(puk)
	//	addrInfo.Addr = bs[:]
	//}
	temp := *addrInfo
	return &temp, utils.NewErrorSuccess()
}

/*
获取网络地址
当有多个时，获取最后一个
@param password 网络地址的密码
@return prk 地址私钥
@return puk 地址公钥
@return err
*/
func (this *Keystore) GetNetAddrKeyPair(password string) (puk []byte, prk []byte, ERR utils.ERROR) {
	addrInfo, ERR := this.GetNetAddrInfo(password)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	pwdBs, err := pbkdf2.Key(sha256.New, password, this.Salt, 1, 32)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	//解密公钥
	puk, ERR = DecryptCBC(addrInfo.CryptedPuk, pwdBs[:], this.Salt)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	//解密私钥
	prk, ERR = DecryptCBC(addrInfo.CryptedPrk, pwdBs[:], this.Salt)
	return
}
