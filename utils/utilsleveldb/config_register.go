package utilsleveldb

import (
	"sync"
	"web3_gui/utils"
)

var dbKeyRegisterMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

/*
注册一个数据库ID，并判断是否有重复
@return    bool    是否成功
*/
func RegisterDbKey(dbKey *LeveldbKey) bool {
	_, ok := dbKeyRegisterMap.LoadOrStore(utils.Bytes2string(dbKey.Byte()), nil)
	if ok {
		utils.Log.Error().Msgf("重复注册的数据库ID:%+v", dbKey.Byte())
	}
	return !ok
}

/*
注册数据库id，有重复就panic
*/
func RegDbKeyExistPanic(dbKey *LeveldbKey) *LeveldbKey {
	if !RegisterDbKey(dbKey) {
		panic("dbKey exist")
	}
	return dbKey
}

/*
注册数据库id，有重复就panic
*/
func RegDbKeyExistPanicByByte(id []byte) *LeveldbKey {
	dbkey, ERR := BuildLeveldbKey(id)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	return RegDbKeyExistPanic(dbkey)
}

/*
注册数据库id，有重复就panic
*/
func RegDbKeyExistPanicByUint64(id uint64) *LeveldbKey {
	dbkey, ERR := BuildDbKeyByUinta64(id)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	return RegDbKeyExistPanic(dbkey)
}
