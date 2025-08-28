package db

import (
	"web3_gui/chain/config"
)

// 保存地址绑定关系
func SaveAddressBind(addr1, addr2 []byte) error {
	if _, err := LevelTempDB.HSet(addr1, addr2, []byte{1}); err != nil {
		return err
	}

	return LevelTempDB.Set(config.BuildAddressTxBindKey(addr2), addr1)
}

// 检查地址绑定关系
func CheckAddressBind(addr1, addr2 []byte) bool {
	v, err := LevelTempDB.HGet(addr1, addr2)
	if v == nil && err == nil {
		return false
	}
	return err == nil
}

// 移除地址绑定关系
func RemAddressBind(addr1, addr2 []byte) error {
	if _, err := LevelTempDB.HDel(addr1, addr2); err != nil {
		return err
	}
	if _, err := LevelTempDB.Del(config.BuildAddressTxBindKey(addr2)); err != nil {
		return err
	}
	return nil
}

// 查询地址是否被绑定过
func AddressIsBind(addr []byte) bool {
	if exists, _ := LevelTempDB.Exists(config.BuildAddressTxBindKey(addr)); exists > 0 {
		return true
	}

	return false
}

// 地址冻结
func SaveAddressFrozen(addr []byte) error {
	return LevelTempDB.Set(config.BuildAddressTxFrozenKey(addr), []byte{1})
}

// 地址解冻
func SaveAddressUnFrozen(addr []byte) error {
	_, err := LevelTempDB.Del(config.BuildAddressTxFrozenKey(addr))
	return err
}

// 查询地址是否被冻结
func AddressIsFrozen(addr []byte) bool {
	if exists, _ := LevelTempDB.Exists(config.BuildAddressTxFrozenKey(addr)); exists > 0 {
		return true
	}

	return false
}
