package keystore

import (
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
创建DHKey协商秘钥
@param password 钱包密码
@param newDHKeyPassword DHKey密码
@return *dh.KeyPair
@return error
*/
func (this *Keystore) CreateDHKey(seedPassword string, dhKeyPassword string) (*CoinAddressInfo, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()

	//查找用过的最高的棘轮数量
	index := uint32(0)
	if len(this.DHAddrs) > 0 {
		addrInfo := this.DHAddrs[len(this.DHAddrs)-1]
		index = addrInfo.Index + 1
	}
	ctafi := &coin_address.AddressCustomDHBuilder{}
	coinType, _ := coin_address.GetCoinTypeCustom()
	addrInfo, ERR := this.buildCoinAddrByCoinType(ctafi, "", seedPassword, dhKeyPassword, coinType+1, index)
	if ERR.CheckFail() {
		return nil, ERR
	}

	//addr := coin_address.AddressCoin(addrInfo.Addr.Bytes())
	// fmt.Println("保存公钥", hex.EncodeToString(addrInfo.Puk), index)
	// fmt.Println("保存PUK", hex.EncodeToString(puk))
	this.DHAddrs = append(this.DHAddrs, addrInfo)
	//this.addrMap.Store(utils.Bytes2string(addr), addrInfo)
	//this.pukMap.Store(utils.Bytes2string(puk), addrInfo)
	ERR = this.Save()
	if ERR.CheckFail() {
		return nil, ERR
	}
	return addrInfo, ERR
}

/*
获取地址列表，包括导入的钱包地址
*/
func (this *Keystore) GetDHAddrInfoAll() []CoinAddressInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	addrInfoList := make([]CoinAddressInfo, 0, len(this.DHAddrs))
	for _, one := range this.DHAddrs {
		addrInfoList = append(addrInfoList, *one)
	}
	return addrInfoList
}

/*
获取DH公钥
当有多个时，获取最后一个
*/
func (this *Keystore) GetDHAddrInfo(password string) (*CoinAddressInfo, utils.ERROR) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if len(this.DHAddrs) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.DHAddrs[len(this.DHAddrs)-1]
	ok, ERR := addrInfo.CheckPassword(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	if ok {
		temp := *addrInfo
		return &temp, utils.NewErrorSuccess()
	}
	return nil, utils.NewErrorBus(config.ERROR_code_dhkey_password_fail, "")
}

/*
修改DHKey密碼
@param index
@param oldpwd 旧密码
@param newpwd 新密码
@return ok
@return err
*/
func (this *Keystore) UpdateDHKeyPwd(oldPassword, newPassword string) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if len(this.DHAddrs) == 0 {
		return utils.NewErrorBus(config.ERROR_code_addr_not_found, "")
	}
	addrInfo := this.DHAddrs[len(this.DHAddrs)-1]
	ok, ERR := addrInfo.CheckPassword(oldPassword)
	if ERR.CheckFail() {
		return ERR
	}
	if !ok {
		return utils.NewErrorBus(config.ERROR_code_dhkey_password_fail, "")
	}
	ERR = addrInfo.UpdatePassword(oldPassword, newPassword)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.Save()
	return ERR
}

/*
导出助记词
@param pwd 钱包密码
@return string 助词记
@return error
*/
func (this *Keystore) GetDhAddrKeyPair(password string) (*KeyPair, utils.ERROR) {
	addrInfo, ERR := this.GetDHAddrInfo(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	keyPair, ERR := addrInfo.GetDHKeyPair(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return keyPair, ERR
}
