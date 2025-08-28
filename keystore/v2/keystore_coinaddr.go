package keystore

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	kv1 "web3_gui/keystore/v1"
	"web3_gui/keystore/v1/derivation"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
创建一个新的收款地址
@param      seedPassword          种子密码
@param      coinAddrPassword      新地址密码
@return     crypto.AddressCoin    新创建的地址
@return     error                 错误
*/
func (this *Keystore) CreateCoinAddr(nickname, seedPassword, coinAddrPassword string) (coin_address.AddressCoin, utils.ERROR) {
	//utils.Log.Info().Uint64("版本号", this.Version).Send()
	if this.Version == config.VERSION_v1 {
		return nil, utils.NewErrorBus(config.ERROR_code_version_old, "")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	//查找用过的最高的棘轮数量
	index := uint32(0)
	if len(this.Addrs) > 0 {
		addrInfo := this.Addrs[len(this.Addrs)-1]
		index = addrInfo.Index + 1
	}
	ctafi := &coin_address.AddressCustomBuilder{}
	coinType, _ := coin_address.GetCoinTypeCustom()
	addrInfo, ERR := this.buildCoinAddrByCoinType(ctafi, nickname, seedPassword, coinAddrPassword, coinType, index)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Str("新地址", addrInfo.GetAddrStr()).Str("新地址2", addrInfo.Addr.B58String()).Send()
	addr := coin_address.AddressCoin(addrInfo.Addr.Bytes())
	// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
	this.Addrs = append(this.Addrs, addrInfo)
	//utils.Log.Info().Hex("地址key", addr.Bytes()).Str("地址字符串", addr.B58String()).Send()
	this.addrMap.Store(utils.Bytes2string(addrInfo.Addr.Bytes()), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return addr, ERR
}

/*
创建一个新的收款地址
@param      seedPassword          种子密码
@param      coinAddrPassword      新地址密码
@return     crypto.AddressCoin    新创建的地址
@return     error                 错误
*/
func (this *Keystore) buildCoinAddrByCoinType(ctafi coin_address.CoinTypeAddressFactoryInterface, nickname, seedPassword,
	coinAddrPassword string, coinType, index uint32) (*CoinAddressInfo, utils.ERROR) {
	//utils.Log.Info().Uint32("创建一个币种地址", coinType).Send()

	mnemonic, ERR := this.exportMnemonic(seedPassword)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	coinAddrPasswordBs, err := pbkdf2.Key(sha256.New, coinAddrPassword, this.Salt, 1, 32)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	checkHash, err := config.GetCheckHashByRounds(coinAddrPassword, this.Salt, this.Rounds)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	checkHashTemp, err := config.GetCheckHashByRounds(coinAddrPassword, this.Salt, this.RoundsTemp)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//密码验证通过

	keyTree, ERR := NewKeyTreeGeneratorByMnemonic(mnemonic)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	_, key, err := keyTree.GetIndexKey(config.PurposeBIP44, coinType, config.ZeroQuote, 0, index)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}

	//config.Coin_address_pre = this.addrPre
	prk, puk, ERR := ctafi.BuildPukAndPrk(key)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}

	addr, ERR := ctafi.BuildAddress(this.addrPre, puk)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	//utils.Log.Error().Str("收款地址", addr.String()).Send()

	pukCrypted, ERR := EncryptCBC(puk, coinAddrPasswordBs[:], this.Salt) //对子地址加密码
	if ERR.CheckFail() {
		return nil, ERR
	}
	prkCrypted, ERR := EncryptCBC(prk, coinAddrPasswordBs[:], this.Salt) //对子地址加密码
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	addrInfo := &CoinAddressInfo{
		keystore:      this,
		Purpose:       config.PurposeBIP44, //固定为 44'，表示 BIP-44 标准。
		CoinType:      coinType,            //指定加密货币的类型（例如，0' 表示 Bitcoin，2' 表示 Litecoin 等）。
		Account:       config.ZeroQuote,    //表示账户索引，允许用户在同一钱包中有多个账户。
		Change:        0,                   //表示地址类型（0 表示接收地址，1 表示找零地址）。
		Index:         index,               //表示地址索引，用于生成每个账户的具体地址。
		Nickname:      nickname,            //
		Addr:          addr,                //收款地址
		CryptedPuk:    pukCrypted,          //公钥
		CryptedPrk:    prkCrypted,          //
		CheckHash:     checkHash,           //
		CheckHashTemp: checkHashTemp,       //
		//SubKey:    subKeySec,                 //
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return addrInfo, utils.NewErrorSysSelf(err)
}

/*
修改地址的昵称
@param name 新昵称
@param password 地址密码
@param addr 地址
@return error
*/
func (this *Keystore) UpdateCoinAddrNickname(addr coin_address.AddressCoin, password string, nickname string) utils.ERROR {
	v, ok := this.addrMap.Load(utils.Bytes2string(addr.Bytes()))
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := v.(*CoinAddressInfo)
	ok, ERR := addrInfo.CheckPassword(password)
	if ERR.CheckFail() {
		return ERR
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_coinAddr_password_fail, "")
	}
	addrInfo.Nickname = nickname
	ERR = this.Save()
	return ERR
}

/*
修改地址的密码
@param name 新昵称
@param password 地址密码
@param addr 地址
@return error
*/
func (this *Keystore) UpdateCoinAddrPwd(addr coin_address.AddressCoin, oldPassword, newPassword string) utils.ERROR {
	v, ok := this.addrMap.Load(utils.Bytes2string(addr.Bytes()))
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := v.(*CoinAddressInfo)
	ok, ERR := addrInfo.CheckPassword(oldPassword)
	if ERR.CheckFail() {
		return ERR
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_coinAddr_password_fail, "")
	}
	ERR = addrInfo.UpdatePassword(oldPassword, newPassword)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.Save()
	return ERR
}

/*
验证指定地址的密码
@param addr 地址
@param pwdbs 地址密码
@return ok 密码是否正确
@return subKeyBs 子密私钥
@return err
*/
func (this *Keystore) GetCoinAddrInfo(addr coin_address.AddressCoin, password string) (*CoinAddressInfo, utils.ERROR) {
	v, ok := this.addrMap.Load(utils.Bytes2string(addr.Bytes()))
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := v.(*CoinAddressInfo)
	ok, ERR := addrInfo.CheckPassword(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	if ok {
		temp := *addrInfo
		return &temp, utils.NewErrorSuccess()
	}
	return nil, utils.NewErrorBus(config.ERROR_code_coinAddr_password_fail, "")
}

/*
获取地址列表，包括导入的钱包地址
*/
func (this *Keystore) GetCoinAddrAll() []CoinAddressInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrInfoList := make([]CoinAddressInfo, 0, len(this.Addrs))
	for _, one := range this.Addrs {
		addrInfoList = append(addrInfoList, *one)
	}
	return addrInfoList
}

/*
钱包中查找地址，判断地址是否属于本钱包
*/
func (this *Keystore) FindCoinAddr(addr coin_address.AddressInterface) *CoinAddressInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	v, ok := this.addrMap.Load(utils.Bytes2string(addr.Bytes()))
	if !ok {
		return nil
	}
	addrInfo := v.(*CoinAddressInfo)
	temp := *addrInfo
	return &temp
}

/*
钱包中查找公钥是否存在
*/
//func (this *Keystore) GetCoinAddrInfoByPukNotPwd(puk []byte) *CoinAddressInfo {
//	v, ok := this.pukMap.Load(utils.Bytes2string(puk))
//	if !ok {
//		return nil
//	}
//	addrInfo := v.(*CoinAddressInfo)
//	temp := *addrInfo
//	return &temp
//}

/*
通过公钥获取密钥
@param puk 地址公钥
@param password 地址密码
@return rand
@return prk 地址私钥
@return err
*/
//func (this *Keystore) GetCoinAddrInfoByPuk(puk []byte, password string) (*CoinAddressInfo, utils.ERROR) {
//	v, ok := this.pukMap.Load(utils.Bytes2string(puk))
//	if !ok {
//		return nil, utils.NewErrorBus(ERROR_code_addr_not_found, "")
//	}
//	addrInfo := v.(*CoinAddressInfo)
//	ok, ERR := addrInfo.CheckPassword(password)
//	if ERR.CheckFail() {
//		return nil, ERR
//	}
//	if ok {
//		temp := *addrInfo
//		return &temp, utils.NewErrorSuccess()
//	}
//	return nil, utils.NewErrorBus(ERROR_code_coinAddr_password_fail, "")
//}

/*
转换为旧版本地址
@param pwd 钱包密码
*/
func (this *Keystore) ConvertOldVersionAddress(index int, pwd string) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	ok, err := this.CheckSeedPassword(pwd)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	//密码验证通过
	//查找用过的最高的棘轮数量
	if index+1 > len(this.Addrs) {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.Addrs[index]
	oldAddr := addrInfo.Addr
	//utils.Log.Info().Uint32("开始创建收款地址", index).Send()
	pwdHash, err := pbkdf2.Key(sha256.New, pwd, this.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//先用密码解密种子
	seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
	if ERR.CheckFail() {
		return ERR
	}

	//旧版问题，兼容旧版
	pwdHashOld := sha256.Sum256([]byte(pwd))
	seedSec, ERR := EncryptCBC(seedBs, pwdHashOld[:], kv1.Salt) //加密
	if ERR.CheckFail() {
		return ERR
	}

	//通过bip44推导出公钥.
	km, _ := derivation.GeneratePrivate(seedSec, pwdHashOld[:], kv1.Salt, this.MnemonicLang)
	//utils.Log.Info().Msgf("生成参数1:%x %x %x", seedSec, pwdHashOld, kv1.Salt)
	coinType, _ := coin_address.GetCoinTypeCustom()
	key, _ := km.GetKey(derivation.PurposeBIP44, coinType, 0, 0, uint32(index)) //通过原生助记词推出私钥
	//utils.Log.Info().Msgf("生成参数2:%d %d %d %+v", derivation.PurposeBIP44, derivation.CoinType, index, key.Bip32Key)
	//根据私钥推导公钥生成地址.
	addr, puk, prk := key.CreateAddr(this.GetAddrPre())
	a := coin_address.AddressCoin(addr)
	addrInfo.Addr = &a
	addrInfo.AddrStr = ""
	//更新公钥
	//utils.Log.Info().Hex("新公钥", puk).Send()
	pukSec, ERR := EncryptCBC(puk, pwdHash, this.Salt) //加密
	if ERR.CheckFail() {
		return ERR
	}
	addrInfo.CryptedPuk = pukSec
	addrInfo.PukStr = ""
	//更新私钥
	prkSec, ERR := EncryptCBC(prk, pwdHash, this.Salt) //加密
	if ERR.CheckFail() {
		return ERR
	}
	addrInfo.CryptedPrk = prkSec

	//删除旧地址
	this.addrMap.Delete(utils.Bytes2string(oldAddr.Bytes()))
	//保存新地址
	this.addrMap.Store(utils.Bytes2string(addr), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	this.Version = config.VERSION_v1
	ERR = this.Save()
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return ERR
}

/*
转换为新版本地址
@param pwd 钱包密码
*/
func (this *Keystore) ConvertNewVersionAddress(index int, pwd string) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	ok, err := this.CheckSeedPassword(pwd)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_seed_password_fail, "")
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	//密码验证通过
	//查找用过的最高的棘轮数量
	if index+1 > len(this.Addrs) {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.Addrs[index]
	oldAddr := addrInfo.Addr
	pwdHash, err := pbkdf2.Key(sha256.New, pwd, this.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	puk, ERR := DecryptCBC(addrInfo.CryptedPuk, pwdHash, this.Salt) //对子地址加密码
	if ERR.CheckFail() {
		return ERR
	}

	var addr coin_address.AddressCoin
	if config.ShortAddress {
		addr, ERR = coin_address.BuildAddr_old(this.addrPre, puk)
	} else {
		addr, ERR = coin_address.BuildAddr(this.addrPre, puk)
	}
	if ERR.CheckFail() {
		return ERR
	}
	this.Addrs[index].Addr = &addr
	this.Addrs[index].AddrStr = ""
	//

	////utils.Log.Info().Uint32("开始创建收款地址", index).Send()
	//pwdHash, err := pbkdf2.Key(sha256.New, pwd, this.Salt, 1, 32)
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	////先用密码解密种子
	//seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
	//if ERR.CheckFail() {
	//	return ERR
	//}
	////通过bip44推导出公钥.
	//km, err := NewKeyTreeGeneratorBySeed(seedBs, pwd, this.MnemonicLang)
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	////utils.Log.Info().Msgf("生成参数1:%x %s %x", seedBs, pwd, this.Salt)
	////utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	//coinType, _ := coin_address.GetCoinTypeCustom()
	////通过原生助记词推出私钥
	//_, key, err := km.GetIndexKey(config.PurposeBIP44, coinType, 0, 0, uint32(index))
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	////utils.Log.Info().Msgf("生成参数2:%d %d %d %+v", PurposeBIP44, CoinType_coin, index, key)
	////utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	////根据私钥推导公钥生成地址.
	//puk, prk := coin_address.ParseKey(key)
	////utils.Log.Info().Hex("公钥", puk).Send()
	//addr, ERR := coin_address.BuildAddr(this.GetAddrPre(), puk)
	//if ERR.CheckFail() {
	//	return ERR
	//}
	//addrInfo.Addr = &addr
	//addrInfo.AddrStr = ""
	//pukCrypted, ERR := EncryptCBC(puk, pwdHash, this.Salt) //对子地址加密码
	//if ERR.CheckFail() {
	//	return ERR
	//}
	//prkCrypted, ERR := EncryptCBC(prk, pwdHash, this.Salt) //对子地址加密码
	//if ERR.CheckFail() {
	//	return ERR
	//}
	//addrInfo.CryptedPuk = pukCrypted
	//addrInfo.CryptedPrk = prkCrypted
	//删除旧地址
	this.addrMap.Delete(utils.Bytes2string(oldAddr.Bytes()))
	//保存新地址
	this.addrMap.Store(utils.Bytes2string(addr.Bytes()), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return ERR
}
