package keystore

import (
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
通过cointype创建一个新的收款地址
@param      password              种子密码
@param      coinAddrPassword      新地址密码
@return     crypto.AddressCoin    新创建的地址
@return     error                 错误
*/
func (this *Keystore) CreateAddressCoinType(nickname, seedPassword, coinAddrPassword string, coinType uint32) (*CoinAddressInfo, utils.ERROR) {
	if this.Version == config.VERSION_v1 {
		return nil, utils.NewErrorBus(config.ERROR_code_version_old, "")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	//utils.Log.Info().Str("CreateAddressCoinType", "3333").Send()
	//密码验证通过
	//查找用过的最高的棘轮数量
	index := uint32(0)
	for _, one := range this.AddrsOther {
		if one.CoinType == coinType {
			index = one.Index + 1
		}
	}

	builder := coin_address.GetCoinTypeFactory(coinType)
	if builder == nil {
		return nil, utils.NewErrorBus(config.ERROR_code_coin_type_addr_not_achieve, "")
	}

	addrInfo, ERR := this.buildCoinAddrByCoinType(builder, nickname, seedPassword, coinAddrPassword, coinType, index)
	if ERR.CheckFail() {
		return nil, ERR
	}

	// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
	this.AddrsOther = append(this.AddrsOther, addrInfo)
	this.addrMap.Store(utils.Bytes2string(addrInfo.Addr.Bytes()), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
	return addrInfo, ERR
}

/*
通过cointype创建一个新的收款地址
@param      password              种子密码
@param      coinAddrPassword      新地址密码
@return     crypto.AddressCoin    新创建的地址
@return     error                 错误
*/
//func (this *Keystore) GetAddressCoinType(coinType, account, index uint32, seedPassword string) (coin_address.AddressInterface, utils.ERROR) {
//	ctafi := coin_address.GetCoinTypeFactory(coinType)
//	if ctafi == nil {
//		return nil, utils.NewErrorBus(ERROR_code_coin_type_addr_not_achieve, "")
//	}
//	//utils.Log.Info().Uint64("版本号", this.Version).Send()
//	if this.Version == VERSION_v1 {
//		return nil, utils.NewErrorBus(ERROR_code_version_old, "")
//	}
//
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	ok, err := this.CheckSeedPassword(seedPassword)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	if !ok {
//		return nil, utils.NewErrorBus(ERROR_code_seed_password_fail, "")
//	}
//	//密码验证通过
//	//utils.Log.Info().Uint32("开始创建收款地址", index).Send()
//	pwdHash, err := pbkdf2.Key(sha256.New, seedPassword, this.Salt, 1, 32)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	//先用密码解密种子
//	seedBs, ERR := DecryptCBC(this.CryptedSeed, pwdHash, this.Salt)
//	if ERR.CheckFail() {
//		return nil, ERR
//	}
//	//utils.Log.Info().Hex("种子原始字节", seedBs).Send()
//	//通过bip44推导出公钥.
//	km, err := NewKeyTreeGeneratorBySeed(seedBs, seedPassword, this.MnemonicLang)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
//	//通过原生助记词推出私钥
//	_, key, err := km.GetIndexKey(PurposeBIP44, coinType, account, 0, index)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	//utils.Log.Info().Str("开始创建收款地址", "11111").Send()
//	_, _, addr, err := ctafi.BuildAddress(key)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	return addr, utils.NewErrorSysSelf(err)
//}
