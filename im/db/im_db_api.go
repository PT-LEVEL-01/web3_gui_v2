package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
修改自己的个人信息
*/
func SaveBatch(batch *leveldb.Batch) error {
	return config.LevelDB.SaveBatch(batch)
}

/*
修改自己的个人信息
*/
func UpdateSelfInfo(addrSelf nodeStore.AddressNet, uinfo *model.UserInfo) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_self_userinfo, *addrSelfKey)
	bs, err := uinfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	ERR = config.LevelDB.Save(*dbKey, bs)
	if ERR.CheckFail() {
		return ERR
	}
	//往用户信息里面放一份
	ERR = ImProxyClient_SaveUserinfo(addrSelf, uinfo)
	return ERR
}

/*
获取自己的个人信息
*/
func GetSelfInfo(addrSelf nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_self_userinfo, *addrSelfKey)
	bs, err := config.LevelDB.Find(*dbKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseUserInfo(&bs.Value)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	// utils.Log.Info().Msgf("用户信息:%+v %+v", userinfo, bs.Value)
	return userinfo, utils.NewErrorSuccess()
}

/*
保存到好友列表中
*/
func SaveUserList(addrSelf nodeStore.AddressNet, userInfo *model.UserInfo, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("保存好友列表:%+v %+v %+v", key, addrSelf, userInfo)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_friend_userlist, *addrSelfKey)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(userInfo.Addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := userInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return config.LevelDB.SaveMap(*dbKey, *addrKey, *bs, batch)
}

/*
在好友列表中查找用户信息
*/
func FindUserListByAddr(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_friend_userlist, *addrSelfKey)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	bs, err := config.LevelDB.FindMap(*dbKey, *addrKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseUserInfo(&bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
在好友列表中删除好友
*/
func DelUserListByAddr(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_friend_userlist, *addrSelfKey)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	err := config.LevelDB.RemoveMapByKey(*dbKey, *addrKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
获取已添加的好友列表
*/
func FindUserListAll(addrSelf nodeStore.AddressNet) ([]model.UserInfo, utils.ERROR) {
	//utils.Log.Info().Msgf("查询用户列表:%+v", key)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_friend_userlist, *addrSelfKey)
	items, err := config.LevelDB.FindMapAllToList(*dbKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	users := make([]model.UserInfo, 0, len(items))
	for _, one := range items {
		if one.Value != nil && len(one.Value) > 0 {
			//utils.Log.Info().Msgf("查询到的用户列表:%+v", one.Value)
			userOne, err := model.ParseUserInfo(&one.Value)
			if err != nil {
				return nil, utils.NewErrorSysSelf(err)
			}
			users = append(users, *userOne)
		} else {
			addrBs, ERR := one.Key.BaseKey()
			if !ERR.CheckSuccess() {
				return nil, ERR
			}
			userOne := model.NewUserInfo(*nodeStore.NewAddressNet(addrBs)) //model.UserInfo{Addr: addrBs}
			users = append(users, *userOne)
		}
	}
	return users, utils.NewErrorSuccess()
}

/*
保存自己主动发出的申请列表
*/
func SaveUserListLocalApply(addrSelf nodeStore.AddressNet, userInfo *model.UserInfo, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("保存好友列表:%+v %+v %+v", key, addrSelf, userInfo)
	return saveUserListApply(*config.DBKEY_apply_local_userlist, *config.DBKEY_apply_local_userlist_index,
		*config.DBKEY_apply_local_userlist_addr, addrSelf, userInfo, batch)
}

/*
保存对方申请添加自己为好友的列表
*/
func SaveUserListRemoteApply(addrSelf nodeStore.AddressNet, userInfo *model.UserInfo, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("保存好友列表:%+v %+v %+v", key, addrSelf, userInfo)
	return saveUserListApply(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index,
		*config.DBKEY_apply_remote_userlist_addr, addrSelf, userInfo, batch)
}

/*
保存到好友申请列表
好友申请列表有两个，一个是我主动发出的申请，一个是别人申请加我好友
*/
func saveUserListApply(key1, key2, key3 utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, userInfo *model.UserInfo, batch *leveldb.Batch) utils.ERROR {
	//当有好友申请的时候，对方会提供公钥，此时统一保存公钥
	if userInfo.GroupDHPuk != nil && len(userInfo.GroupDHPuk) > 0 {
		//查询并保存公钥
		user, ERR := ImProxyClient_FindUserinfo(addrSelf, userInfo.Addr)
		if ERR.CheckFail() {
			return ERR
		}
		if user == nil {
			user = model.NewUserInfo(userInfo.Addr)
		}
		if userInfo.GroupDHPuk != nil && len(userInfo.GroupDHPuk) > 0 {
			user.GroupDHPuk = userInfo.GroupDHPuk
		}
		ERR = ImProxyClient_SaveUserinfo(addrSelf, user)
		if ERR.CheckFail() {
			return ERR
		}
	}
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(key1, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(key2, *addrSelfKey)
	dbKey3 := utilsleveldb.JoinDbKey(key3, *addrSelfKey)
	tokenKey, ERR := utilsleveldb.BuildLeveldbKey(userInfo.Token)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return ERR
	}
	bs, err := userInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//查询是否存在，存在则是修改操作，覆盖之前的记录
	indexBs, err := config.LevelDB.FindMap(*dbKey2, *tokenKey)
	if err != nil {
		utils.Log.Info().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if indexBs != nil && len(indexBs) > 0 {
		//已经存在，修改操作
		ERR = config.LevelDB.SaveOrUpdateListByIndex(*dbKey1, indexBs, *bs, batch)
	} else {
		//utils.Log.Info().Msgf("保存好友申请记录:%+v", dbKey1)
		//不存在，保存新的
		indexBs, ERR = config.LevelDB.SaveList(*dbKey1, *bs, batch)
		if ERR.CheckFail() {
			utils.Log.Info().Msgf("错误:%s", ERR.String())
			return ERR
		}
		ERR = config.LevelDB.SaveMap(*dbKey2, *tokenKey, indexBs, batch)
		if ERR.CheckFail() {
			utils.Log.Info().Msgf("错误:%s", ERR.String())
			return ERR
		}
		addrKey, ERR := utilsleveldb.BuildLeveldbKey(userInfo.Addr.GetAddr())
		if !ERR.CheckSuccess() {
			return ERR
		}
		if userInfo.IsGroup {
			//是群，则使用地址和群id作联合主键
			addrKey, ERR = utilsleveldb.NewLeveldbKeyJoin(userInfo.Addr.GetAddr(), userInfo.GroupId)
			if !ERR.CheckSuccess() {
				return ERR
			}
		}
		ERR = config.LevelDB.SaveMap(*dbKey3, *addrKey, indexBs, batch)
	}
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
在好友申请列表中通过index查找好友
*/
func FindUserListApplyByToken(key1, key2 utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, token []byte) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(key1, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(key2, *addrSelfKey)
	tokenKey, ERR := utilsleveldb.BuildLeveldbKey(token)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	indexBs, err := config.LevelDB.FindMap(*dbKey2, *tokenKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	item, ERR := config.LevelDB.FindListByIndex(*dbKey1, indexBs)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseUserInfo(&item.Value)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
在好友申请列表中通过地址查找好友
*/
func FindUserListApplyByAddr(key1, key2 utilsleveldb.LeveldbKey, addrSelf, addr nodeStore.AddressNet, groupId []byte) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Str("error", ERR.String()).Send()
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(key1, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(key2, *addrSelfKey)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Str("error", ERR.String()).Send()
		return nil, ERR
	}
	if len(groupId) > 0 {
		addrKey, ERR = utilsleveldb.NewLeveldbKeyJoin(addr.GetAddr(), groupId)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Str("error", ERR.String()).Send()
			return nil, ERR
		}
	}
	indexBs, err := config.LevelDB.FindMap(*dbKey2, *addrKey)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		utils.Log.Error().Str("error", ERR.String()).Send()
		return nil, ERR
	}
	if indexBs == nil || len(indexBs) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	item, ERR := config.LevelDB.FindListByIndex(*dbKey1, indexBs)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Str("error", ERR.String()).Send()
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseUserInfo(&item.Value)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		utils.Log.Error().Str("error", ERR.String()).Send()
		return nil, ERR
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
在不同的好友申请列表中查找好友
*/
func FindUserListApplyAll(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet) ([]model.UserInfo, utils.ERROR) {
	//utils.Log.Info().Msgf("查询用户列表:%+v", key)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)
	//utils.Log.Info().Msgf("查询的用户列表:%+v", dbKey)
	items, ERR := config.LevelDB.FindListRange(*dbKey, nil, 0, false)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	users := make([]model.UserInfo, 0, len(items))
	for _, one := range items {
		if one.Value != nil && len(one.Value) > 0 {
			//utils.Log.Info().Msgf("查询到的用户列表:%+v", one.Value)
			userOne, err := model.ParseUserInfo(&one.Value)
			if err != nil {
				return nil, utils.NewErrorSysSelf(err)
			}
			users = append(users, *userOne)
		} else {
			addrBs, ERR := one.Key.BaseKey()
			if !ERR.CheckSuccess() {
				return nil, ERR
			}
			userOne := model.NewUserInfo(*nodeStore.NewAddressNet(addrBs)) //model.UserInfo{Addr: addrBs}
			users = append(users, *userOne)
		}
	}
	return users, utils.NewErrorSuccess()
}

/*
添加好友到列表
*/
//var addUserListLock = new(sync.Mutex)
//
//func AddUserList(userInfo *model.UserInfo, key utilsleveldb.LeveldbKey) utils.ERROR {
//	addUserListLock.Lock()
//	defer addUserListLock.Unlock()
//	item, err := config.LevelDB.Find(key)
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	var userinfoList *model.UserInfoList
//	if item != nil {
//		userinfoList, err = model.ParseUserList(&item.Value)
//		if err != nil {
//			return utils.NewErrorSysSelf(err)
//		}
//	}
//	if userinfoList == nil {
//		userinfoList = model.NewUserList()
//	}
//	for _, one := range userinfoList.UserList {
//		if bytes.Equal(one.Addr, userInfo.Addr) {
//			return utils.NewErrorSuccess()
//		}
//	}
//	userinfoList.UserList = append(userinfoList.UserList, userInfo)
//	bs, err := userinfoList.Proto()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	ERR := config.LevelDB.Save(key, bs)
//	return ERR
//}

/*
从列表中删除好友
*/
//var delUserListLock = new(sync.Mutex)
//
//func DelUserList(addr nodeStore.AddressNet, key utilsleveldb.LeveldbKey) utils.ERROR {
//	delUserListLock.Lock()
//	defer delUserListLock.Unlock()
//	item, err := config.LevelDB.Find(key)
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	var userinfoList *model.UserInfoList
//	if item != nil {
//		userinfoList, err = model.ParseUserList(&item.Value)
//		if err != nil {
//			return utils.NewErrorSysSelf(err)
//		}
//	}
//	for i, one := range userinfoList.UserList {
//		if bytes.Equal(one.Addr, addr) {
//			temp := userinfoList.UserList[:i]
//			userinfoList.UserList = append(temp, userinfoList.UserList[i+1:]...)
//			break
//		}
//	}
//	bs, err := userinfoList.Proto()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	ERR := config.LevelDB.Save(key, bs)
//	return ERR
//}

/*
修改列表中的好友信息
*/
//var updateUserListLock = new(sync.Mutex)
//
//func UpdateUserList(userInfo model.UserInfo, key utilsleveldb.LeveldbKey) utils.ERROR {
//	updateUserListLock.Lock()
//	defer updateUserListLock.Unlock()
//	item, err := config.LevelDB.Find(key)
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	var userinfoList *model.UserInfoList
//	if item != nil {
//		userinfoList, err = model.ParseUserList(&item.Value)
//		if err != nil {
//			return utils.NewErrorSysSelf(err)
//		}
//	}
//	for i, one := range userinfoList.UserList {
//		if bytes.Equal(one.Addr, userInfo.Addr) {
//			userinfoList.UserList[i] = &userInfo
//			break
//		}
//	}
//	bs, err := userinfoList.Proto()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	ERR := config.LevelDB.Save(key, bs)
//	return ERR
//}

/*
查询好友列表
*/
//func GetUserList(key utilsleveldb.LeveldbKey) (*model.UserInfoList, utils.ERROR) {
//	users, ERR := FindUserListAll(key)
//	if !ERR.CheckSuccess() {
//		return nil, ERR
//	}
//	userinfoList := model.NewUserList()
//	if users != nil && len(users) > 0 {
//		for _, one := range users {
//			userinfoList.UserList = append(userinfoList.UserList, &one)
//		}
//	}
//	return userinfoList, utils.NewErrorSuccess()
//
//	//item, err := config.LevelDB.Find(key)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//var userinfoList *model.UserInfoList
//	//if item != nil {
//	//	userinfoList, err = model.ParseUserList(&item.Value)
//	//}
//	//return userinfoList, err
//}

//var messageHistoryLock = new(sync.Mutex)

/*
添加好友聊天记录
*/
func AddMessageHistory_old(msgContent *model.MessageContent) (*model.MessageContent, utils.ERROR) {
	//utils.Log.Info().Msgf("保存记录:%+v", *msgContent)
	//构造数据库key
	selfAddr := msgContent.To
	remoteAddr := msgContent.From
	if msgContent.FromIsSelf {
		selfAddr = msgContent.From
		remoteAddr = msgContent.To
	}
	key := config.BuildKeyUserMsgIndex(selfAddr.GetAddr(), remoteAddr.GetAddr())
	// utils.Log.Info().Msgf("查询的key:%s", hex.EncodeToString(key))
	dbKeyTo, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.From.GetAddr(), msgContent.To.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	sendIDKey, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.From.GetAddr(), msgContent.To.GetAddr(), msgContent.SendID)
	//sendIDKey, err := utilsleveldb.BuildLeveldbKey(msgContent.SendID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	var recvIDKey *utilsleveldb.LeveldbKey
	if msgContent.RecvID != nil && len(msgContent.RecvID) > 0 {
		recvIDKey, ERR = utilsleveldb.NewLeveldbKeyJoin(msgContent.From.GetAddr(), msgContent.To.GetAddr(), msgContent.RecvID)
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			return nil, ERR
		}
	}
	//var ERR utils.ERROR
	//开启事务
	err := config.LevelDB.OpenTransaction()
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		return nil, utils.NewErrorSysSelf(err)
	}
	defer func() {
		if ERR.CheckSuccess() {
			//事务提交
			err = config.LevelDB.Commit()
			if err != nil {
				config.LevelDB.Discard()
				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
				return
			}
			return
		}
		//事务回滚
		config.LevelDB.Discard()
	}()

	//未发送消息总数不能超过10条，需要判断
	if msgContent.State == config.MSG_GUI_state_not_send {
		//查询未发送消息总量
		totalBs, err := config.LevelDB.FindMap(*config.DBKEY_message_undelivered_total, *dbKeyTo)
		if err != nil {
			//utils.Log.Info().Msgf("保存记录")
			ERR = utils.NewErrorSysSelf(err)
			return nil, ERR
		}
		if totalBs == nil {
			totalBs = utils.Uint64ToBytes(0)
		}
		total := utils.BytesToUint64(totalBs)
		//判断是否超出未发送消息总数量
		if total >= config.MsgChanMaxLength {
			//utils.Log.Info().Msgf("保存记录")
			ERR = utils.NewErrorBus(config.ERROR_CODE_IM_too_many_undelivered_messages, "")
			return nil, ERR
		}
		total++
		totalBs = utils.Uint64ToBytes(total)
		//保存新的未发送消息总数量
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_message_undelivered_total, *dbKeyTo, totalBs)
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			//ERR = utils.NewErrorSysSelf(err)
			return nil, ERR
		}
	}

	//查询index
	var indexBs *[]byte
	item, err := config.LevelDB.Find(key)
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		ERR = utils.NewErrorSysSelf(err)
		return nil, ERR
	}
	if item == nil {
		temp := utils.Uint64ToBytes(0)
		indexBs = &temp
	} else {
		indexBs = &item.Value
	}
	//utils.Log.Info().Msgf("保存记录")
	index := new(big.Int).Add(new(big.Int).SetBytes(*indexBs), big.NewInt(1))
	newIndexBs := index.Bytes()
	//index := utils.BytesToUint64(*indexBs)
	// utils.Log.Info().Msgf("查询到的index:%d", index)
	//index++
	//newIndexBs := utils.Uint64ToBytes(index)
	msgContent.Index = newIndexBs
	//保存记录到数据库
	msgKey := config.BuildKeyUserMsgContent(selfAddr.GetAddr(), remoteAddr.GetAddr(), newIndexBs)
	// utils.Log.Info().Msgf("保存的msgkey:%s", hex.EncodeToString(msgKey))
	bs, err := msgContent.Proto()
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		return nil, utils.NewErrorSysSelf(err)
	}
	//保存消息内容
	ERR = config.LevelDB.Save_Transaction(msgKey, bs)
	//err = config.LevelDB.Save(msgKey, bs)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		//ERR = utils.NewErrorSysSelf(err)
		return nil, ERR
	}
	//保存最新的index
	ERR = config.LevelDB.Save_Transaction(key, &newIndexBs)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		//ERR = utils.NewErrorSysSelf(err)
		return nil, ERR
	}
	//是自己发出的消息，则保存发送id对应的消息index，方便发送失败后再次发送时更新状态
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_message_send_id, *sendIDKey, newIndexBs)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		//ERR = utils.NewErrorSysSelf(err)
		return nil, ERR
	}
	if recvIDKey != nil {
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_message_recv_id, *recvIDKey, newIndexBs)
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			//ERR = utils.NewErrorSysSelf(err)
			return nil, ERR
		}
	}
	//utils.Log.Info().Msgf("保存记录")
	ERR = utils.NewErrorSuccess()
	return msgContent, ERR
}

var AddMessageHistoryLock = new(sync.Mutex)

/*
添加好友聊天记录
*/
func AddMessageHistoryV2(selfAddr nodeStore.AddressNet, msgContent *model.MessageContent, batch *leveldb.Batch) (*model.MessageContent, utils.ERROR) {
	sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(msgContent.SendID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	AddMessageHistoryLock.Lock()
	defer AddMessageHistoryLock.Unlock()
	//utils.Log.Info().Msgf("保存聊天记录:%+v", *msgContent)
	//构造数据库key
	localAddr := msgContent.To
	remoteAddr := msgContent.From
	if msgContent.IsGroup {
		localAddr = selfAddr
		remoteAddr = msgContent.To
	} else if msgContent.FromIsSelf {
		localAddr = msgContent.From
		remoteAddr = msgContent.To
	}
	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(localAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	//查询最后一条记录的index
	_, _, endIndex, ERR := config.LevelDB.FindMapInListTotal(*config.DBKEY_improxy_user_message_history, *joinKey)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	indexBs := new(big.Int).Add(new(big.Int).SetBytes(endIndex), big.NewInt(1)).Bytes()
	msgContent.Index = indexBs
	bs, err := msgContent.Proto()
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		return nil, utils.NewErrorSysSelf(err)
	}
	//保存这个索引，方便修改其状态
	ERR = config.LevelDB.SaveMapInMap(*config.DBKEY_improxy_user_message_history_sendID, *joinKey, *sendIdKey, msgContent.Index, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//保存这个记录
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*config.DBKEY_improxy_user_message_history, *joinKey, indexBs, *bs, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//是保存的文件，则保存文件hash索引
	if msgContent.FileHash != nil && len(msgContent.FileHash) > 0 {
		key := config.DBKEY_improxy_user_message_history_fileHash_recv
		if msgContent.FromIsSelf {
			key = config.DBKEY_improxy_user_message_history_fileHash_send
		}
		addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(selfAddr.GetAddr())
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		dbKey := utilsleveldb.JoinDbKey(*key, *addrSelfKey)
		fileHashKey, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.FileHash, utils.Uint64ToBytesByBigEndian(uint64(msgContent.FileSendTime)))
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			return nil, ERR
		}

		//utils.Log.Info().Msgf("保存文件索引的key:%+v %+v %+v %+v", *dbKey, *joinKey, *fileHashKey, indexBs)
		//保存这个文件索引，方便查找
		ERR = config.LevelDB.SaveMapInMap(*dbKey, *joinKey, *fileHashKey, indexBs, batch)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
	}
	return msgContent, utils.NewErrorSuccess()
}

/*
查询一条好友聊天记录
*/
func FindMessageHistoryBySendIdV2(selfAddr, remoteAddr nodeStore.AddressNet, sendId []byte) (*model.MessageContent, utils.ERROR) {
	sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(sendId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("保存记录:%+v", *msgContent)
	//构造数据库key
	//selfAddr := msgContent.To
	//remoteAddr := msgContent.From
	//if msgContent.FromIsSelf {
	//	selfAddr = msgContent.From
	//	remoteAddr = msgContent.To
	//}
	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(selfAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	//查询记录的索引
	indexBs, err := config.LevelDB.FindMapInMapByKeyIn(*config.DBKEY_improxy_user_message_history_sendID, *joinKey, *sendIdKey)
	if err != nil {
		utils.Log.Info().Msgf("查询索引 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if indexBs == nil {
		return nil, utils.NewErrorSuccess()
	}
	item, ERR := config.LevelDB.FindMapInListByIndex(*config.DBKEY_improxy_user_message_history, *joinKey, *indexBs)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("查询索引 错误:%s", ERR.String())
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	msgContent, err := model.ParseMessageContent(&item.Value)
	return msgContent, utils.NewErrorSysSelf(err)
}

/*
查询聊天记录中的文件
*/
func FindMessageHistoryByFileHash(key utilsleveldb.LeveldbKey, selfAddr, remoteAddr nodeStore.AddressNet,
	fileHash []byte, sendTime int64) (*model.MessageContent, utils.ERROR) {

	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(selfAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)
	fileHashKey, ERR := utilsleveldb.NewLeveldbKeyJoin(fileHash, utils.Uint64ToBytesByBigEndian(uint64(sendTime)))
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}

	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(selfAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	//utils.Log.Info().Msgf("查询文件索引的key:%+v %+v %+v", *dbKey, *joinKey, *fileHashKey)
	//查询记录的索引
	indexBs, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey, *joinKey, *fileHashKey)
	if err != nil {
		utils.Log.Info().Msgf("查询索引 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if indexBs == nil {
		return nil, utils.NewErrorSuccess()
	}

	//utils.Log.Info().Msgf("查询的索引:%+v", indexBs)

	item, ERR := config.LevelDB.FindMapInListByIndex(*config.DBKEY_improxy_user_message_history, *joinKey, *indexBs)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("查询索引 错误:%s", ERR.String())
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	if item == nil {
		utils.Log.Info().Msgf("未找到记录:%+v", indexBs)
		return nil, utils.NewErrorSuccess()
	}
	msgContent, err := model.ParseMessageContent(&item.Value)
	msgContent.Index = *indexBs
	return msgContent, utils.NewErrorSysSelf(err)
}

/*
修改聊天记录中的文件
*/
func UpdateMessageHistoryByFileHash(selfAddr nodeStore.AddressNet, msgContent *model.MessageContent, batch *leveldb.Batch) utils.ERROR {
	//构造数据库key
	localAddr := msgContent.To
	remoteAddr := msgContent.From
	if msgContent.IsGroup {
		localAddr = selfAddr
		remoteAddr = msgContent.To
	} else if msgContent.FromIsSelf {
		localAddr = msgContent.From
		remoteAddr = msgContent.To
	}

	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(localAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return ERR
	}
	//fileHashKey, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.FileHash, utils.Uint64ToBytesByBigEndian(uint64(msgContent.FileSendTime)))
	//if !ERR.CheckSuccess() {
	//	//utils.Log.Info().Msgf("保存记录")
	//	return ERR
	//}
	//dbkey := config.DBKEY_improxy_user_message_history_fileHash_recv
	//if msgContent.FromIsSelf {
	//	dbkey = config.DBKEY_improxy_user_message_history_fileHash_send
	//}
	bs, err := msgContent.Proto()
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		return utils.NewErrorSysSelf(err)
	}

	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*config.DBKEY_improxy_user_message_history, *joinKey, msgContent.Index, *bs, batch)
	//保存这个文件索引，方便查找
	//ERR = config.LevelDB.SaveMapInMap(dbkey, *joinKey, *fileHashKey, msgContent.Index, batch)
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	return utils.NewErrorSuccess()
}

/*
修改好友聊天记录的状态
*/
func UpdateSendMessageStateV2(selfAddr, remoteAddr nodeStore.AddressNet, proxyId []byte, status int) utils.ERROR {
	//utils.Log.Info().Msgf("修改好友聊天记录的状态 查询参数:%s %s %+v", selfAddr.B58String(), remoteAddr.B58String(), proxyId)
	//status := msgContent.State
	msgContent, ERR := FindMessageHistoryBySendIdV2(selfAddr, remoteAddr, proxyId)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return ERR
	}
	msgContent.State = status
	//utils.Log.Info().Msgf("修改记录:%+v", *msgContent)
	bs, err := msgContent.Proto()
	if err != nil {
		//utils.Log.Info().Msgf("保存记录")
		return utils.NewErrorSysSelf(err)
	}
	//构造数据库key
	//selfAddr := msgContent.To
	//remoteAddr := msgContent.From
	//if msgContent.FromIsSelf {
	//	selfAddr = msgContent.From
	//	remoteAddr = msgContent.To
	//}
	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(selfAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return ERR
	}
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*config.DBKEY_improxy_user_message_history, *joinKey, msgContent.Index, *bs, nil)
	return ERR
}

/*
获取好友聊天记录，倒叙取
@startIndex    uint64    起始编号
@count         uint64    取多少条记录
*/
func GetMessageHistoryV2(startIndex []byte, count uint64, selfAddr, remoteAddr nodeStore.AddressNet) (*model.MessageContentList, utils.ERROR) {
	joinKey, ERR := utilsleveldb.NewLeveldbKeyJoin(selfAddr.GetAddr(), remoteAddr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	items, ERR := config.LevelDB.FindMapInListRangeByKeyIn(*config.DBKEY_improxy_user_message_history, *joinKey, startIndex, count, false)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return nil, ERR
	}
	mcList := model.MessageContentList{
		List: make([]*model.MessageContent, 0),
	}
	for _, item := range items {
		mcOne, err := model.ParseMessageContent(&item.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Msgf("查询到的记录:%+v", mcOne)
		mcList.List = append(mcList.List, mcOne)
	}
	return &mcList, utils.NewErrorSuccess()
}

/*
查询一条消息记录
*/
func FindMessageContentSend(selfAddr, toAddr nodeStore.AddressNet, sendID []byte) (*model.MessageContent, utils.ERROR) {
	return findMessageContent(selfAddr, toAddr, sendID)
}

/*
查询一条消息记录
*/
//func FindMessageContentRecv(selfAddr, toAddr nodeStore.AddressNet, sendID []byte) (*model.MessageContent, utils.ERROR) {
//	return findMessageContent(toAddr, selfAddr, sendID)
//}

/*
查询一条消息记录
*/
func findMessageContent(fromAddr, toAddr nodeStore.AddressNet, sendID []byte) (*model.MessageContent, utils.ERROR) {
	sendIDKey, ERR := utilsleveldb.NewLeveldbKeyJoin(fromAddr.GetAddr(), toAddr.GetAddr(), sendID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	newIndexBs, err := config.LevelDB.FindMap(*config.DBKEY_message_send_id, *sendIDKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if newIndexBs == nil {
		return nil, utils.NewErrorSuccess()
	}
	//保存记录到数据库
	msgKey := config.BuildKeyUserMsgContent(fromAddr.GetAddr(), toAddr.GetAddr(), newIndexBs)
	utils.Log.Info().Msgf("通过sendid查询数据:%+v", msgKey)
	item, err := config.LevelDB.Find(msgKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	msgContent, err := model.ParseMessageContent(&item.Value)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return msgContent, utils.NewErrorSysSelf(err)
}

/*
更新发送消息状态
*/
func UpdateSendMessageState(msgContent *model.MessageContent) utils.ERROR {
	//构造数据库key
	selfAddr := msgContent.To
	remoteAddr := msgContent.From
	if msgContent.FromIsSelf {
		selfAddr = msgContent.From
		remoteAddr = msgContent.To
	}
	sendIDKey, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.From.GetAddr(), msgContent.To.GetAddr(), msgContent.SendID)
	//sendIDKey, err := utilsleveldb.BuildLeveldbKey(msgContent.SendID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	indexBs, err := config.LevelDB.FindMap(*config.DBKEY_message_send_id, *sendIDKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	dbKeyTo, ERR := utilsleveldb.NewLeveldbKeyJoin(msgContent.From.GetAddr(), msgContent.To.GetAddr())
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//
	msgKey := config.BuildKeyUserMsgContent(selfAddr.GetAddr(), remoteAddr.GetAddr(), indexBs)
	// utils.Log.Info().Msgf("保存的msgkey:%s", hex.EncodeToString(msgKey))
	item, err := config.LevelDB.Find(msgKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)

	}
	if item == nil {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	newMsg, err := model.ParseMessageContent(&item.Value)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	newMsg.State = msgContent.State

	bs, err := newMsg.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//var ERR utils.ERROR
	//开启事务
	err = config.LevelDB.OpenTransaction()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	defer func() {
		if ERR.CheckSuccess() {
			//事务提交
			err = config.LevelDB.Commit()
			if err != nil {
				config.LevelDB.Discard()
				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
				return
			}
			return
		}
		//事务回滚
		config.LevelDB.Discard()
	}()

	//自己发送的消息，并且成功了，则减少发送失败的总量
	if msgContent.FromIsSelf && msgContent.State == config.MSG_GUI_state_success {
		//查询未发送消息总量
		totalBs, err := config.LevelDB.FindMap(*config.DBKEY_message_undelivered_total, *dbKeyTo)
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			return ERR
		}
		if totalBs == nil {
			totalBs = utils.Uint64ToBytes(0)
		}
		total := utils.BytesToUint64(totalBs)
		//判断是否超出未发送消息总数量
		if total > 0 {
			total--
			totalBs = utils.Uint64ToBytes(total)
			//保存新的未发送消息总数量
			ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_message_undelivered_total, *dbKeyTo, totalBs)
			if !ERR.CheckSuccess() {
				//ERR = utils.NewErrorSysSelf(err)
				return ERR
			}
		}
	}
	//保存新的状态
	ERR = config.LevelDB.Save_Transaction(msgKey, bs)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	ERR = utils.NewErrorSuccess()
	return ERR
}

/*
获取好友聊天记录，倒叙取
@startIndex    uint64    起始编号
@count         uint64    取多少条记录
*/
func GetMessageHistory(startIndex, count uint64, selfAddr, remoteAddr []byte) (*model.MessageContentList, error) {
	//messageHistoryLock.Lock()
	//defer messageHistoryLock.Unlock()
	//取最新编号
	if startIndex == 0 {
		key := config.BuildKeyUserMsgIndex(selfAddr, remoteAddr)
		// utils.Log.Info().Msgf("查询的key:%s", hex.EncodeToString(key))
		//查询index
		var indexBs *[]byte
		item, err := config.LevelDB.Find(key)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, nil
		} else {
			indexBs = &item.Value
		}
		startIndex = utils.BytesToUint64(*indexBs)
		// utils.Log.Info().Msgf("查询到的index:%d", startIndex)
	}
	if startIndex == 0 {
		// utils.Log.Info().Msgf("没有数据")
		//本身就没有数据
		return nil, nil
	}

	//循环取记录
	total := count
	if startIndex < count {
		total = startIndex
	}
	// utils.Log.Info().Msgf("查询记录总条数:%d", total)
	if total == 0 {
		// utils.Log.Info().Msgf("没有记录")
		return nil, nil
	}

	mcList := model.MessageContentList{
		List: make([]*model.MessageContent, 0),
	}
	for i := uint64(0); i < total; i++ {
		indexBs := utils.Uint64ToBytes(startIndex - i)
		msgKey := config.BuildKeyUserMsgContent(selfAddr, remoteAddr, indexBs)
		// utils.Log.Info().Msgf("查询的msgkey:%s", hex.EncodeToString(msgKey))
		item, err := config.LevelDB.Find(msgKey)
		if err != nil {
			return nil, err
		}
		mcOne, err := model.ParseMessageContent(&item.Value)
		if err != nil {
			return nil, err
		}
		mcList.List = append(mcList.List, mcOne)
	}
	return &mcList, nil
}

/*
删除所有好友聊天记录
*/
//func RemoveMessageHistoryAll(remoteAddr, selfAddr nodeStore.AddressNet) utils.ERROR {
//	//utils.Log.Info().Msgf("保存记录:%+v", *msgContent)
//	//构造数据库key
//	key := config.BuildKeyUserMsgIndex(selfAddr, remoteAddr)
//	// utils.Log.Info().Msgf("查询的key:%s", hex.EncodeToString(key))
//	dbKeyTo, err := utilsleveldb.NewLeveldbKeyJoin(selfAddr, remoteAddr)
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		return utils.NewErrorSysSelf(err)
//	}
//
//	var ERR utils.ERROR
//	//开启事务
//	err = config.LevelDB.OpenTransaction()
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		return utils.NewErrorSysSelf(err)
//	}
//	defer func() {
//		if ERR.CheckSuccess() {
//			//事务提交
//			err = config.LevelDB.Commit()
//			if err != nil {
//				config.LevelDB.Discard()
//				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
//				return
//			}
//			return
//		}
//		//事务回滚
//		config.LevelDB.Discard()
//	}()
//
//	//未发送消息总数不能超过10条
//	err = config.LevelDB.RemoveMapByKey_Transaction(config.DBKEY_message_undelivered_total, *dbKeyTo)
//	if err != nil {
//		ERR = utils.NewErrorSysSelf(err)
//		return ERR
//
//	}
//
//	//查询index
//	var indexBs *[]byte
//	item, err := config.LevelDB.Find(key)
//	if err != nil {
//		if err != leveldb.ErrNotFound {
//			utils.Log.Info().Msgf("保存记录")
//			ERR = utils.NewErrorSysSelf(err)
//			return  ERR
//		} else {
//			temp := utils.Uint64ToBytes(0)
//			indexBs = &temp
//		}
//	} else {
//		indexBs = &item.Value
//	}
//	utils.Log.Info().Msgf("保存记录")
//	index := utils.BytesToUint64(*indexBs)
//	// utils.Log.Info().Msgf("查询到的index:%d", index)
//	newIndexBs := utils.Uint64ToBytes(index)
//	for {
//
//	}
//	msgContent.Index = index
//	//保存记录到数据库
//	msgKey := config.BuildKeyUserMsgContent(selfAddr, remoteAddr, newIndexBs)
//	// utils.Log.Info().Msgf("保存的msgkey:%s", hex.EncodeToString(msgKey))
//	bs, err := msgContent.Proto()
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	//保存消息内容
//	err = config.LevelDB.Save_Transaction(msgKey, bs)
//	//err = config.LevelDB.Save(msgKey, bs)
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		ERR = utils.NewErrorSysSelf(err)
//		return nil, ERR
//	}
//	//保存最新的index
//	err = config.LevelDB.Save_Transaction(key, &newIndexBs)
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		ERR = utils.NewErrorSysSelf(err)
//		return nil, ERR
//	}
//	//是自己发出的消息，则保存发送id对应的消息index，方便发送失败后再次发送时更新状态
//	err = config.LevelDB.SaveMap_Transaction(config.DBKEY_message_send_id, *sendIDKey, newIndexBs)
//	if err != nil {
//		utils.Log.Info().Msgf("保存记录")
//		ERR = utils.NewErrorSysSelf(err)
//		return nil, ERR
//	}
//	if recvIDKey != nil {
//		err = config.LevelDB.SaveMap_Transaction(config.DBKEY_message_recv_id, *recvIDKey, newIndexBs)
//		if err != nil {
//			utils.Log.Info().Msgf("保存记录")
//			ERR = utils.NewErrorSysSelf(err)
//			return nil, ERR
//		}
//	}
//	utils.Log.Info().Msgf("保存记录")
//	ERR = utils.NewErrorSuccess()
//	return msgContent, ERR
//}
