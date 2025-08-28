package keystore

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/keystore/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

type CoinAddressInfo struct {
	keystore      *Keystore                     //
	Purpose       uint32                        //固定为 44'，表示 BIP-44 标准。
	CoinType      uint32                        //指定加密货币的类型（例如，0' 表示 Bitcoin，2' 表示 Litecoin 等）。
	Account       uint32                        //表示账户索引，允许用户在同一钱包中有多个账户。
	Change        uint32                        //表示地址类型（0 表示接收地址，1 表示找零地址）。
	Index         uint32                        //表示地址索引，用于生成每个账户的具体地址。
	Nickname      string                        //地址昵称
	Addr          coin_address.AddressInterface //地址
	CryptedPuk    []byte                        //加密后公钥
	CryptedPrk    []byte                        //加密后私钥
	AddrStr       string                        //
	PukStr        string                        //
	CheckHash     []byte                        //单独设置密码的验证凭证，保存到文件的复杂验证
	CheckHashTemp []byte                        //单独设置密码的验证凭证，保存到内存的简单验证
}

func (this *CoinAddressInfo) GetAddrStr() string {
	if this.AddrStr != "" {
		return this.AddrStr
	}
	//if this.CoinType == config.CoinType_coin {
	//	//addr := coin_address.AddressCoin(this.Addr)
	//	this.AddrStr = this.Addr.B58String() // addr.B58String()
	//} else {
	//	ctafi := coin_address.GetCoinTypeFactory(this.CoinType)
	//	if ctafi == nil {
	//		return ""
	//	}
	//	this.AddrStr = ctafi.GetAddressStr(this.Addr)
	//}
	//addr := AddressCoin(this.Addr)
	this.AddrStr = this.Addr.B58String() // addr.B58String()
	return this.AddrStr
}

/*
验证密码
@return    bool    密码是否正确.true=正确;false=错误;
*/
func (this *CoinAddressInfo) CheckPassword(password string) (bool, utils.ERROR) {
	if this.CheckHashTemp == nil || len(this.CheckHashTemp) == 0 {
		ok, err := config.ValidateCheckHash(password, this.CheckHash, this.keystore.Salt, this.keystore.Rounds)
		if err != nil {
			return false, utils.NewErrorSysSelf(err)
		}
		if !ok {
			return false, utils.NewErrorSuccess()
		}
		//密码验证成功
		//计算一个容易验证checkHash放在内存，软件运行中需要快速的验证
		this.CheckHashTemp, err = config.GetCheckHashByRounds(password, this.keystore.Salt, this.keystore.RoundsTemp)
		if err != nil {
			return true, utils.NewErrorSysSelf(err)
		}
		return true, utils.NewErrorSuccess()
	}
	ok, err := config.ValidateCheckHash(password, this.CheckHashTemp, this.keystore.Salt, this.keystore.RoundsTemp)
	if err != nil {
		return true, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return false, utils.NewErrorSuccess()
	}
	return true, utils.NewErrorSuccess()
}

/*
解密公钥
@return    []byte    公钥信息
*/
func (this *CoinAddressInfo) GetDHKeyPair(password string) (*KeyPair, utils.ERROR) {
	ok, ERR := this.CheckPassword(password)
	if ERR.CheckFail() {
		return nil, ERR
	}
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_code_password_fail, "")
	}
	pwdHash, err := pbkdf2.Key(sha256.New, password, this.keystore.Salt, 1, 32)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	pukBs, ERR := DecryptCBC(this.CryptedPuk, pwdHash, this.keystore.Salt)
	if ERR.CheckFail() {
		return nil, ERR
	}
	prkBs, ERR := DecryptCBC(this.CryptedPrk, pwdHash, this.keystore.Salt)
	if ERR.CheckFail() {
		return nil, ERR
	}
	keyPair := NewKeyPair(prkBs, pukBs)
	return keyPair, utils.NewErrorSuccess()
}

/*
解密公钥
@return    []byte    公钥
@return    []byte    私钥
*/
func (this *CoinAddressInfo) DecryptPrkPuk(password string) ([]byte, []byte, utils.ERROR) {
	ok, ERR := this.CheckPassword(password)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	if !ok {
		return nil, nil, utils.NewErrorBus(config.ERROR_code_password_fail, "")
	}
	pwdHash, err := pbkdf2.Key(sha256.New, password, this.keystore.Salt, 1, 32)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	pukBs, ERR := DecryptCBC(this.CryptedPuk, pwdHash, this.keystore.Salt)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	prkBs, ERR := DecryptCBC(this.CryptedPrk, pwdHash, this.keystore.Salt)
	if ERR.CheckFail() {
		return nil, nil, ERR
	}
	return pukBs, prkBs, utils.NewErrorSuccess()
}

/*
修改密码
@return    utils.ERROR
*/
func (this *CoinAddressInfo) UpdatePassword(oldPassword, newPassword string) utils.ERROR {
	oldHash, err := pbkdf2.Key(sha256.New, oldPassword, this.keystore.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	newHash, err := pbkdf2.Key(sha256.New, newPassword, this.keystore.Salt, 1, 32)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//oldHash := sha256.Sum256([]byte(oldPassword))
	//newHash := sha256.Sum256([]byte(newPassword))
	//先用密码解密
	pukBs, ERR := DecryptCBC(this.CryptedPuk, oldHash[:], this.keystore.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	prkBs, ERR := DecryptCBC(this.CryptedPrk, oldHash[:], this.keystore.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	//再用新密码加密
	cryptedPuk, ERR := EncryptCBC(pukBs, newHash[:], this.keystore.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	cryptedPrk, ERR := EncryptCBC(prkBs, newHash[:], this.keystore.Salt)
	if ERR.CheckFail() {
		return ERR
	}
	checkHash, err := config.GetCheckHashByRounds(newPassword, this.keystore.Salt, this.keystore.Rounds)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	checkHashTemp, err := config.GetCheckHashByRounds(newPassword, this.keystore.Salt, this.keystore.RoundsTemp)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.CryptedPuk = cryptedPuk
	this.CryptedPrk = cryptedPrk
	this.CheckHash = checkHash
	this.CheckHashTemp = checkHashTemp
	return utils.NewErrorSuccess()
}

/*
转化
*/
func (this *CoinAddressInfo) Conver() *go_protobuf.CoinAddress {
	ksProto := go_protobuf.CoinAddress{}
	ksProto.Purpose = this.Purpose
	ksProto.CoinType = this.CoinType
	ksProto.Account = this.Account
	ksProto.Change = this.Change
	ksProto.Index = this.Index
	ksProto.Nickname = this.Nickname
	if this.Addr != nil {
		ksProto.Addr = this.Addr.Bytes()
	}
	ksProto.CryptedPuk = this.CryptedPuk
	ksProto.CryptedPrk = this.CryptedPrk
	ksProto.CheckHash = this.CheckHash
	return &ksProto
}

/*
格式化成[]byte
*/
func (this *CoinAddressInfo) Proto() (*[]byte, error) {
	ks := this.Conver()
	bs, err := ks.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ConverProtoCoinAddr(coinAddr *go_protobuf.CoinAddress) *CoinAddressInfo {
	coinA := CoinAddressInfo{}
	coinA.Purpose = coinAddr.Purpose
	coinA.CoinType = coinAddr.CoinType
	coinA.Account = coinAddr.Account
	coinA.Change = coinAddr.Change
	coinA.Index = coinAddr.Index
	coinA.Nickname = coinAddr.Nickname
	//coinA.Addr = coinAddr.Addr
	coinA.CryptedPuk = coinAddr.CryptedPuk
	coinA.CryptedPrk = coinAddr.CryptedPrk
	coinA.CheckHash = coinAddr.CheckHash
	//utils.Log.Info().Uint32("币种编号", coinAddr.CoinType).Hex("地址data", coinAddr.Addr).Send()
	ctafi := coin_address.GetCoinTypeFactory(coinAddr.CoinType)
	if ctafi != nil {
		//utils.Log.Info().Uint32("币种编号", coinAddr.CoinType).Hex("地址data", coinAddr.Addr).Str("找到了", "").Send()
		coinA.Addr = ctafi.NewAddress(coinAddr.Addr)
	}
	return &coinA
}
