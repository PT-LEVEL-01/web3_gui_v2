package message_center

import (
	"errors"
	"sync/atomic"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
添加消息到缓存
*/
func AddMessageCacheByBytes(hash, content *[]byte, levelDB *utilsleveldb.LevelDB) utils.ERROR {
	return AddMessageCache(hash, content, levelDB)
}

/*
添加消息到缓存
*/
func AddMessageCache(hash *[]byte, msgBytes *[]byte, levelDB *utilsleveldb.LevelDB) utils.ERROR {
	key := atomic.LoadUint64(&config.CurBroadcastAddValue)
	keyOut := utils.Uint64ToBytesByBigEndian(key)
	newHashKey, ERR := utilsleveldb.BuildLeveldbKey(*hash)
	if !ERR.CheckSuccess() {
		return ERR
	}
	newKeyOut, ERR := utilsleveldb.BuildLeveldbKey(keyOut)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = levelDB.SaveMapInMap(*config.DBKEY_broadcast_msg, *newKeyOut, *newHashKey, *msgBytes, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
通过hash查询一条消息是否存在
*/
func FindMessageCacheByHashExist(hash []byte, levelDB *utilsleveldb.LevelDB) (bool, error) {
	bs, err := FindMessageCacheByHash(hash, levelDB)
	if err != nil {
		return false, err
	}
	if bs == nil {
		return false, nil
	}
	return true, nil
}

/*
通过hash查询一条消息
*/
func FindMessageCacheByHash(hash []byte, levelDB *utilsleveldb.LevelDB) (*[]byte, error) {
	addValue := atomic.LoadUint64(&config.CurBroadcastAddValue)
	addValueKey := utils.Uint64ToBytesByBigEndian(addValue)

	//newBroadcastKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg)
	//if !ERR.CheckSuccess() {
	//	return nil, errors.New(ERR.Msg)
	//}
	newAddValueKey, ERR := utilsleveldb.BuildLeveldbKey(addValueKey)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	newHashKey, ERR := utilsleveldb.BuildLeveldbKey(hash)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	bs, err := levelDB.FindMapInMapByKeyIn(*config.DBKEY_broadcast_msg, *newAddValueKey, *newHashKey)
	if err != nil {
		return nil, err
	}
	//utils.Log.Info().Msgf("开始查询消息:%+v", bs)
	if bs != nil {
		return bs, nil
	}
	// 如果没有找到数据, 则会降低一次addValue, 再次查找一次
	if addValue > 1 {
		addValue--
		addValueKey2 := utils.Uint64ToBytesByBigEndian(addValue)
		newAddValueKey2, ERR := utilsleveldb.BuildLeveldbKey(addValueKey2)
		if !ERR.CheckSuccess() {
			return nil, errors.New(ERR.Msg)
		}
		bs, err = levelDB.FindMapInMapByKeyIn(*config.DBKEY_broadcast_msg, *newAddValueKey2, *newHashKey)
		if err != nil {
			//utils.Log.Info().Msgf("开始查询消息 错误:%s", err.Error())
			return nil, err
		}
		//utils.Log.Info().Msgf("开始查询消息:%+v", bs)
		if bs != nil {
			return bs, nil
		}
		return nil, nil
	} else {
		return nil, nil
	}
}
