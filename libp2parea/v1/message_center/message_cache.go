package message_center

import (
	"errors"
	"sync/atomic"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
添加消息到缓存
*/
func AddMessageCacheByBytes(hash, head, body *[]byte, levelDB *utilsleveldb.LevelDB) error {
	msgBytes := go_protobuf.MessageMulticast{
		Head: *head,
		Body: *body,
	}
	return AddMessageCache(hash, &msgBytes, levelDB)
}

/*
添加消息到缓存
*/
func AddMessageCache(hash *[]byte, msgBytes *go_protobuf.MessageMulticast, levelDB *utilsleveldb.LevelDB) error {
	bs, err := msgBytes.Marshal()
	if err != nil {
		return err
	}

	key := atomic.LoadUint64(&config.CurBroadcastAddValue)
	keyOut := utils.Uint64ToBytesByBigEndian(key)
	newHashKey, ERR := utilsleveldb.BuildLeveldbKey(*hash)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	newKeyOut, ERR := utilsleveldb.BuildLeveldbKey(keyOut)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	// utils.Log.Info().Msgf("hash:%v bs:%v", hash, bs)
	newBroadcastKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = levelDB.SaveMapInMap(*newBroadcastKey, *newKeyOut, *newHashKey, bs, nil)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
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

	newBroadcastKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	newAddValueKey, ERR := utilsleveldb.BuildLeveldbKey(addValueKey)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	newHashKey, ERR := utilsleveldb.BuildLeveldbKey(hash)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	bs, err := levelDB.FindMapInMapByKeyIn(*newBroadcastKey, *newAddValueKey, *newHashKey)
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
		bs, err = levelDB.FindMapInMapByKeyIn(*newBroadcastKey, *newAddValueKey2, *newHashKey)
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
	//if err != nil {
	//	utils.Log.Info().Msgf("开始查询消息 错误:%s", err.Error())
	//	if err == leveldb.ErrNotFound {
	//		// 如果没有找到数据, 则会降低一次addValue, 再次查找一次
	//		if addValue > 1 {
	//			addValue--
	//			addValueKey2 := utils.Uint64ToBytesByBigEndian(addValue)
	//			newAddValueKey2, ERR := utilsleveldb.NewLeveldbKey(addValueKey2)
	//			if !ERR.CheckSuccess() {
	//				return nil, errors.New(ERR.Msg)
	//			}
	//			bs, err = levelDB.FindMapInMapByKeyIn(*newBroadcastKey, *newAddValueKey2, *newHashKey)
	//		}
	//		if err != nil {
	//			utils.Log.Info().Msgf("开始查询消息 错误:%s", err.Error())
	//			return nil, err
	//		}
	//	} else {
	//		utils.Log.Info().Msgf("开始查询消息 错误:%s", err.Error())
	//		return nil, err
	//	}
	//}
	//utils.Log.Info().Msgf("开始查询消息")
	//return bs, nil
}
