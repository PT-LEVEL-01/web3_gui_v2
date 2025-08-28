package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"time"
	"web3_gui/config"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
保存一个群信息到列表中
*/
func ImProxyClient_SaveGroupList(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet,
	group *imdatachain.DataChainCreateGroup, batch *leveldb.Batch) utils.ERROR {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(group.GroupID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := group.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//indexBs := big.NewInt(group.CreateTime).Bytes()
	utils.Log.Info().Msgf("添加到群列表:%+v %+v %+v", key, *addrKey, *groupIdKey)
	return config.LevelDB.SaveMapInMap(key, *addrKey, *groupIdKey, *bs, batch)
}

/*
从群信息列表中删除一个群
*/
func ImProxyClient_RemoveGroupList(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet,
	groupId []byte, batch *leveldb.Batch) utils.ERROR {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	utils.Log.Info().Msgf("删除群列表:%+v %+v %+v", key, *addrKey, *groupIdKey)
	err := config.LevelDB.RemoveMapInMapByKeyIn(key, *addrKey, *groupIdKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//清理群数据
	ERR = cleanGroupDatachain(addrSelf, groupId, batch)
	return ERR
}

/*
查询群列表
*/
func ImProxyClient_FindGroupList(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet) (*[]imdatachain.DataChainCreateGroup, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	items, ERR := config.LevelDB.FindMapInMapByKeyOut(key, *addrKey)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//for _, one := range items {
	//	utils.Log.Info().Msgf("解析群信息:%d %+v", len(one.Value), one.Value)
	//}
	groups := make([]imdatachain.DataChainCreateGroup, 0, len(items))
	for _, one := range items {
		//utils.Log.Info().Msgf("解析群信息:%d %+v", len(one.Value), one.Value)
		item, err := imdatachain.ParseDataChainCreateGroupFactory(one.Value)
		if err != nil {
			utils.Log.Error().Msgf("解析群信息错误:%s", err.Error())
			return nil, utils.NewErrorSysSelf(err)
		}
		groupOne := item.(*imdatachain.DataChainCreateGroup)
		groups = append(groups, *groupOne)
	}
	return &groups, utils.NewErrorSuccess()
}

/*
查询一个群是否在列表中
*/
func ImProxyClient_FindGroupListExist(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte) (bool, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	bs, err := config.LevelDB.FindMapInMapByKeyIn(key, *addrKey, *groupIdKey)
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return false, utils.NewErrorSuccess()
	}
	return true, utils.NewErrorSuccess()
}

/*
列表中查询群信息
*/
func ImProxyClient_FindGroupInfo(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte) (*imdatachain.DataChainCreateGroup, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	utils.Log.Info().Msgf("查询群信息:%+v %+v %+v", key, *addrKey, *groupIdKey)
	bs, err := config.LevelDB.FindMapInMapByKeyIn(key, *addrKey, *groupIdKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	proxyItr, err := imdatachain.ParseDataChainCreateGroupFactory(*bs)
	if err != nil {
		utils.Log.Error().Msgf("解析群信息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
	return createGroup, utils.NewErrorSuccess()
}

/*
在所有列表中查找，直到找到群
*/
func FindGroupInfoAllList(addrSelf nodeStore.AddressNet, groupId []byte) (*imdatachain.DataChainCreateGroup, utils.ERROR) {
	//当各个群列表都不存在的时候，就可以清理群数据链记录了
	groupInfo, ERR := ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_create, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if groupInfo != nil {
		return groupInfo, utils.NewErrorSuccess()
	}
	groupInfo, ERR = ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_join, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if groupInfo != nil {
		return groupInfo, utils.NewErrorSuccess()
	}
	groupInfo, ERR = ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_knit, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if groupInfo != nil {
		return groupInfo, utils.NewErrorSuccess()
	}
	groupInfo, ERR = ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_minor, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if groupInfo != nil {
		return groupInfo, utils.NewErrorSuccess()
	}
	groupInfo, ERR = ImProxyClient_FindDissolveGroupInfo(addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if groupInfo != nil {
		return groupInfo, utils.NewErrorSuccess()
	}
	return nil, utils.NewErrorSuccess()
}

/*
保存到解散群列表
*/
func ImProxyClient_SaveDissolveGroupList(addrSelf nodeStore.AddressNet, groupId []byte, batch *leveldb.Batch) utils.ERROR {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_list_dissolve, *addrKey)
	groupKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	index := utils.Uint64ToBytesByBigEndian(uint64(time.Now().Unix()))
	return config.LevelDB.SaveMap(*joinKey, *groupKey, index, batch)
}

/*
删除解散群列表中的一个群
*/
func ImProxyClient_RemoveDissolveGroupList(addrSelf nodeStore.AddressNet, groupId []byte, batch *leveldb.Batch) utils.ERROR {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_list_dissolve, *addrKey)
	groupKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	err := config.LevelDB.RemoveMapByKey(*joinKey, *groupKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//删除群成员
	err = config.LevelDB.RemoveMapInMapByKeyOut(*config.DBKEY_improxy_group_members_knit, *groupKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	ERR = cleanGroupDatachain(addrSelf, groupId, batch)
	return ERR
}

func cleanGroupDatachain(addrSelf nodeStore.AddressNet, groupId []byte, batch *leveldb.Batch) utils.ERROR {
	//当各个群列表都不存在的时候，就可以清理群数据链记录了
	ok, ERR := ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_create, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if ok {
		return utils.NewErrorSuccess()
	}
	ok, ERR = ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_join, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if ok {
		return utils.NewErrorSuccess()
	}
	ok, ERR = ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_knit, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if ok {
		return utils.NewErrorSuccess()
	}
	ok, ERR = ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_minor, addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if ok {
		return utils.NewErrorSuccess()
	}
	ok, ERR = ImProxyClient_FindDissolveGroupExist(addrSelf, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if ok {
		return utils.NewErrorSuccess()
	}
	//所有列表都不存在的情况，删除链的所有数据
	groupKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//删除群数据链
	err := config.LevelDB.RemoveMapInListAll(*config.DBKEY_improxy_user_datachain, *groupKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//删除群聊天记录
	//删除索引
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*addrKey, *groupKey)
	err = config.LevelDB.RemoveMapInMapByKeyOut(*config.DBKEY_improxy_user_message_history_sendID, *joinKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//删除记录
	err = config.LevelDB.RemoveMapInListAll(*config.DBKEY_improxy_user_message_history, *joinKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//删除每个成员刚加入群的index
	ERR = ImProxyClient_RemoveGroupMemberStartIndex(addrSelf, groupId, batch)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("保存记录")
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询解散的群列表
*/
func ImProxyClient_FindDissolveGroupList(addrSelf nodeStore.AddressNet) ([]utilsleveldb.DBItem, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_list_dissolve, *addrKey)
	items, err := config.LevelDB.FindMapAllToList(*joinKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return items, utils.NewErrorSuccess()
}

/*
解散群列表中查询群是否存在
*/
func ImProxyClient_FindDissolveGroupExist(addrSelf nodeStore.AddressNet, groupId []byte) (bool, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_list_dissolve, *addrKey)
	groupKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return false, ERR
	}

	items, err := config.LevelDB.FindMapByKeys(*joinKey, *groupKey)
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	if items != nil && len(items) > 0 {
		return true, utils.NewErrorSuccess()
	}
	return false, utils.NewErrorSuccess()
}

/*
解散群列表中查询群信息
*/
func ImProxyClient_FindDissolveGroupInfo(addrSelf nodeStore.AddressNet, groupId []byte) (*imdatachain.DataChainCreateGroup, utils.ERROR) {
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_list_dissolve, *addrKey)
	groupKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	items, err := config.LevelDB.FindMapByKeys(*joinKey, *groupKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if items == nil || len(items) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	proxyItr, err := imdatachain.ParseDataChainCreateGroupFactory(items[0].Value)
	if err != nil {
		utils.Log.Error().Msgf("解析群信息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
	return createGroup, utils.NewErrorSuccess()
}

/*
保存快照高度
@index    []byte    快照高度
@return    []byte    记录最后的index
@return    error     错误
*/
func ImProxyClient_SaveGroupShot(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId, bs []byte, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("加载快照高度:%+v", addr)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	joinKey := utilsleveldb.JoinDbKey(*addrKey, *groupIdKey)
	ERR = config.LevelDB.SaveMap(key, *joinKey, bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//index := itr.GetIndex()
	return utils.NewErrorSuccess()
}

/*
加载自己聊天记录的快照高度
@return    []byte    快照高度
@return    []byte    记录最后的index
@return    error     错误
*/
func ImProxyClient_LoadGroupShot(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte) ([]byte, utils.ERROR) {
	//utils.Log.Info().Msgf("加载快照高度:%+v", addr)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	joinKey := utilsleveldb.JoinDbKey(*addrKey, *groupIdKey)
	bs, err := config.LevelDB.FindMap(key, *joinKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return bs, utils.NewErrorSuccess()

	//indexBs, err := config.LevelDB.FindMapInMapByKeyIn(config.DBKEY_improxy_user_group_shot_index_parser, *addrKey, *groupIdKey)
	//if err != nil {
	//	return nil, nil, utils.NewErrorSysSelf(err)
	//}
	//itr, ERR := ImProxyClient_FindDataChainLast(addr)
	//if !ERR.CheckSuccess() {
	//	//utils.Log.Info().Msgf("加载快照高度 错误:%s", ERR.String())
	//	return nil, nil, ERR
	//}
	////index := itr.GetIndex()
	//return *indexBs, itr, utils.NewErrorSuccess()
}

/*
保存群成员信息
*/
func ImProxyClient_SaveGroupMembers(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte,
	members []model.UserInfo, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)

	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//清空原来的好友列表
	err := config.LevelDB.RemoveMapInMapByKeyOut(*dbKey, *groupIdKey, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存新成员
	for _, one := range members {
		utils.Log.Info().Msgf("保存群成员:%+v %+v", key, one.Addr.B58String())
		addrKey, ERR := utilsleveldb.BuildLeveldbKey(one.Addr.GetAddr())
		if !ERR.CheckSuccess() {
			return ERR
		}
		bsOne, err := one.Proto()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		ERR = config.LevelDB.SaveMapInMap(*dbKey, *groupIdKey, *addrKey, *bsOne, batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	return utils.NewErrorSuccess()
}

/*
查询群成员信息
*/
func ImProxyClient_FindGroupMembers(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte) (*[]model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)

	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	items, ERR := config.LevelDB.FindMapInMapByKeyOut(*dbKey, *groupIdKey)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	users := make([]model.UserInfo, 0, len(items))
	for _, one := range items {
		userinfo, err := model.ParseUserInfo(&one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		//如果列表中没有详细信息，则查询一次数据库
		if userinfo.Nickname == "" {
			ui, ERR := ImProxyClient_FindUserinfo(addrSelf, userinfo.Addr)
			if ERR.CheckSuccess() && ui != nil {
				userinfo.Nickname = ui.Nickname
			}
		}
		users = append(users, *userinfo)
	}
	return &users, utils.NewErrorSuccess()
}

/*
在群成员列表中查询一个群的成员信息
*/
func ImProxyClient_FindGroupMember(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte,
	addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey, *groupIdKey, *addrKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	userinfo, err := model.ParseUserInfo(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
删除一个群成员
*/
func ImProxyClient_RemoveGroupMember(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, groupId []byte,
	addr nodeStore.AddressNet, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(key, *addrSelfKey)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	err := config.LevelDB.RemoveMapInMapByKeyIn(*dbKey, *groupIdKey, *addrKey, batch)
	//bs, err := config.LevelDB.FindMapInMapByKeyIn(config.DBKEY_improxy_user_group_members, *groupIdKey, *addrKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
保存群数据链
*/
func ImProxyClient_SaveGroupDataChain(addrSelf nodeStore.AddressNet, itr imdatachain.DataChainProxyItr, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_id, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain, *addrSelfKey)
	dbKey3 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_sendid, *addrSelfKey)
	//utils.Log.Info().Msgf("保存群数据链:%+v", itr)
	bs, err := itr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("消息序列化:%+v", bs)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetBase().GroupID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	idKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetID())
	if !ERR.CheckSuccess() {
		return ERR
	}
	index := itr.GetIndex()

	utils.Log.Info().Msgf("保存消息唯一ID:%+v %+v %+v %+v", *dbKey1, *groupIdKey, *idKey, index.Bytes())
	//保存这条消息的唯一ID，用于去重
	ERR = config.LevelDB.SaveMapInMap(*dbKey1, *groupIdKey, *idKey, index.Bytes(), batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存这条消息的sendID，用于去重
	if itr.GetSendID() != nil && len(itr.GetID()) != 0 {
		sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetSendID())
		if !ERR.CheckSuccess() {
			return ERR
		}
		ERR = config.LevelDB.SaveMapInMap(*dbKey3, *groupIdKey, *sendIdKey, index.Bytes(), batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//utils.Log.Info().Msgf("保存消息唯一ID:%+v %+v %+v %+v", config.DBKEY_improxy_user_datachain_id, *addrKey, *idKey, index.Bytes())
	//保存消息内容
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*dbKey2, *groupIdKey, index.Bytes(), *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("保存消息内容:%+v %+v %+v %+v", config.DBKEY_improxy_user_datachain, *addrKey, index.Bytes(), *bs)
	return utils.NewErrorSuccess()
}

/*
查询一条数据链记录是否存在
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindGroupDataChainByID(addrSelf nodeStore.AddressNet, groupId []byte, id []byte) (imdatachain.DataChainProxyItr, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_id, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain, *addrSelfKey)
	//dbKey3 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_group_datachain_sendid, *addrSelfKey)
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v", addr, id)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	idKey, ERR := utilsleveldb.BuildLeveldbKey(id)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v %+v", *dbKey1, *groupIdKey, *idKey)
	index, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey1, *groupIdKey, *idKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if index == nil {
		//utils.Log.Info().Msgf("未找到记录:%+v %+v", addr, id)
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("查询数据链记录id:%+v %+v %+v", config.DBKEY_improxy_user_datachain_id, *addrKey, *idKey)
	//注意：因为保存时考虑性能未使用事务，则可能存在出错时，只保存了数据链id，而未保存数据内容的情况。
	item, ERR := config.LevelDB.FindMapInListByIndex(*dbKey2, *groupIdKey, *index)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("查询数据链数据:%+v %+v %+v", config.DBKEY_improxy_user_datachain, *addrKey, index)
	itr, ERR := imdatachain.ParseDataChain(item.Value)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return itr, utils.NewErrorSuccess()
}

/*
查询一条数据链记录是否存在
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindGroupDataChainBySendID(addrSelf nodeStore.AddressNet, groupId, sendId []byte) (bool, utils.ERROR) {
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v", addr, id)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_sendid, *addrSelfKey)
	//dbKey2 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_user_datachain, *addrSelfKey)

	addrKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(sendId)
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v %+v", config.DBKEY_improxy_user_datachain_id, *addrKey, *idKey)
	index, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey1, *addrKey, *sendIdKey)
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	if index == nil {
		//utils.Log.Info().Msgf("未找到记录:%+v %+v", addr, id)
		return false, utils.NewErrorSuccess()
	}
	return true, utils.NewErrorSuccess()
}

/*
通过index查询一条记录
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindGroupDataChainByIndex(addrSelf nodeStore.AddressNet, groupId []byte, index []byte) (imdatachain.DataChainProxyItr, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_group_datachain_id, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain, *addrSelfKey)
	//dbKey3 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_group_datachain_sendid, *addrSelfKey)

	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	item, ERR := config.LevelDB.FindMapInListByIndex(*dbKey2, *groupIdKey, index)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("解析的数据:%+v", item)
	itr, ERR := imdatachain.ParseDataChain(item.Value)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return itr, utils.NewErrorSuccess()
}

/*
通过index查询范围内的记录
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindGroupDataChainRange(addrSelf nodeStore.AddressNet, groupId, indexStart []byte) ([]imdatachain.DataChainProxyItr, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_group_datachain_id, *addrSelfKey)
	dbKey2 := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain, *addrSelfKey)
	//dbKey3 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_group_datachain_sendid, *addrSelfKey)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	items, ERR := config.LevelDB.FindMapInListRangeByKeyIn(*dbKey2, *groupIdKey, indexStart, config.IMPROXY_sync_total_once, true)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(items))
	for _, one := range items {
		itr, ERR := imdatachain.ParseDataChain(one.Value)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		itrs = append(itrs, itr)
	}
	return itrs, utils.NewErrorSuccess()
}

/*
保存发送者的index
*/
func ImProxyClient_SaveGroupSendIndex(key utilsleveldb.LeveldbKey, addrFrom nodeStore.AddressNet, groupId []byte,
	indexBs []byte, batch *leveldb.Batch) utils.ERROR {
	keyDB, ERR := utilsleveldb.NewLeveldbKeyJoin(addrFrom.GetAddr(), groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(key, *keyDB, indexBs, batch)
}

/*
查询发送者的index
*/
func ImProxyClient_FindGroupSendIndex(key utilsleveldb.LeveldbKey, addrFrom nodeStore.AddressNet, groupId []byte) (*big.Int, utils.ERROR) {
	keyDB, ERR := utilsleveldb.NewLeveldbKeyJoin(addrFrom.GetAddr(), groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(key, *keyDB)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return big.NewInt(0), utils.NewErrorSuccess()
	}
	index := new(big.Int).SetBytes(bs)
	return index, utils.NewErrorSuccess()
}

/*
保存群成员加入群的index
*/
func ImProxyClient_SaveGroupMemberStartIndex(addrSelf nodeStore.AddressNet, groupId []byte, addrMember nodeStore.AddressNet,
	index []byte, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_member_start, *addrSelfKey)

	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	addrMemberKey, ERR := utilsleveldb.BuildLeveldbKey(addrMember.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMapInMap(*dbKey, *groupIdKey, *addrMemberKey, index, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询群成员加入群的index
*/
func ImProxyClient_FindGroupMemberStartIndex(addrSelf nodeStore.AddressNet, groupId []byte, addrMember nodeStore.AddressNet) (*[]byte, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_member_start, *addrSelfKey)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	addrMemberKey, ERR := utilsleveldb.BuildLeveldbKey(addrMember.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	index, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey, *groupIdKey, *addrMemberKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return index, utils.NewErrorSuccess()
}

/*
删除群所有成员加入群的index
*/
func ImProxyClient_RemoveGroupMemberStartIndex(addrSelf nodeStore.AddressNet, groupId []byte, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_group_datachain_member_start, *addrSelfKey)
	groupIdKey, ERR := utilsleveldb.BuildLeveldbKey(groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	err := config.LevelDB.RemoveMapInListAll(*dbKey, *groupIdKey, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
保存发送文件列表
@addrSelf         nodeStore.AddressNet          属于谁的数据
@addr             nodeStore.AddressNet          好友地址或者群地址
@datachainFile    *imdatachain.DatachainFile    待发送文件
*/
func ImProxyClient_SaveSendFileList(addrSelf nodeStore.AddressNet, sendFileInfo *imdatachain.SendFileInfo) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	var addr nodeStore.AddressNet
	if sendFileInfo.IsGroup {
		addr = sendFileInfo.Addr
	} else {
		addr = addrSelf
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_send_file_list, *addrSelfKey, *addrKey)

	indexBs := append(utils.Int64ToBytesByBigEndian(sendFileInfo.SendTime), sendFileInfo.Hash...)

	//fileHashKey, ERR := utilsleveldb.NewLeveldbKeyJoin(datachainFile.Hash, utils.Int64ToBytesByBigEndian(datachainFile.Time))
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	bs, err := sendFileInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	ERR = config.LevelDB.SaveOrUpdateListByIndex(*dbKey, indexBs, *bs, nil)
	//_, ERR = config.LevelDB.SaveMapInList(*dbKey, *fileHashKey, *bs, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询文件发送列表
@addrSelf         nodeStore.AddressNet          属于谁的数据
@addr             nodeStore.AddressNet          好友地址或者群地址
@return    *imdatachain.DatachainFile    待发送文件
*/
func ImProxyClient_FindSendFileList(addrSelf, addr nodeStore.AddressNet) (*imdatachain.SendFileInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_send_file_list, *addrSelfKey, *addrKey)

	var fileInfo *imdatachain.SendFileInfo
	items, ERR := config.LevelDB.FindListRange(*dbKey, nil, 1, true)
	for _, item := range items {
		var err error
		fileInfo, err = imdatachain.ParseSendFileInfo(item.Value)
		//fileOne, err := imdatachain.ParseDatachainFileFactory(item.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		break
	}
	return fileInfo, utils.NewErrorSuccess()
}

/*
查询文件发送列表
@addrSelf         nodeStore.AddressNet          属于谁的数据
@addr             nodeStore.AddressNet          好友地址或者群地址
@return    *imdatachain.DatachainFile    待发送文件
*/
func ImProxyClient_DelSendFileList(addrSelf nodeStore.AddressNet, datachainFile *imdatachain.SendFileInfo) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	var addr nodeStore.AddressNet
	if datachainFile.IsGroup {
		addr = datachainFile.Addr
	} else {
		addr = addrSelf
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_send_file_list, *addrSelfKey, *addrKey)
	indexBs := append(utils.Int64ToBytesByBigEndian(datachainFile.SendTime), datachainFile.Hash...)
	ERR = config.LevelDB.RemoveListByIndex(*dbKey, indexBs, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存接收文件列表
@addrSelf         nodeStore.AddressNet          属于谁的数据
@addr             nodeStore.AddressNet          好友地址或者群地址
@datachainFile    *imdatachain.DatachainFile    待发送文件
*/
func ImProxyClient_SaveRecvFileList(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet,
	datachainFile *imdatachain.DatachainFile) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*config.DBKEY_improxy_send_file_list, *addrSelfKey, *addrKey)
	fileHashKey, ERR := utilsleveldb.NewLeveldbKeyJoin(datachainFile.Hash, utils.Int64ToBytesByBigEndian(datachainFile.Time))
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := datachainFile.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	_, ERR = config.LevelDB.SaveMapInList(*dbKey, *fileHashKey, *bs, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}
