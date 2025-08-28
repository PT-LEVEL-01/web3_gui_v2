package subscription

import (
	"sync"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

// 保存自己的个人信息
var userSelfLock = new(sync.RWMutex)
var userSelf *model.UserInfo

// 保存已经添加的好友个人信息
var friendListLock = new(sync.RWMutex)
var friendList = make(map[string]*model.UserInfo)

var userInfoCache = utils.NewCache(10000)

//func GetUserSelf_cache() model.UserInfo {
//	userSelfLock.RLock()
//	defer userSelfLock.RUnlock()
//	return *userSelf
//}

//func SetUserSelf_cache(userinfo *model.UserInfo) {
//	userSelfLock.RLock()
//	userSelf = userinfo
//	userSelfLock.RUnlock()
//	SetFriendOne_cache(userinfo)
//}

//func GetFriendInfo_cache(addr nodeStore.AddressNet) model.UserInfo {
//	friendListLock.RLock()
//	defer friendListLock.RUnlock()
//	userinfo := friendList[utils.Bytes2string(addr)]
//	if userinfo == nil {
//		return *model.NewUserInfo(addr)
//	}
//	return *userinfo
//}
//func SetFriendList_cache(userinfos []*model.UserInfo) {
//	userMap := make(map[string]*model.UserInfo)
//	for _, userinfo := range userinfos {
//		userMap[utils.Bytes2string(userinfo.Addr)] = userinfo
//	}
//	//把自己的信息也放进去，方便查询
//	userSelf := GetUserSelf_cache()
//	userMap[utils.Bytes2string(userSelf.Addr)] = &userSelf
//	friendListLock.Lock()
//	defer friendListLock.Unlock()
//	friendList = userMap
//}

//func SetFriendOne_cache(userinfo *model.UserInfo) {
//	friendListLock.Lock()
//	defer friendListLock.Unlock()
//	friendList[utils.Bytes2string(userinfo.Addr)] = userinfo
//}

/*
从缓存中获取用户基本信息
*/
func CacheGetUserInfo(addr *nodeStore.AddressNet) *model.UserInfo {
	userItr, ok := userInfoCache.Get(utils.Bytes2string(addr.GetAddr()))
	if !ok {
		//未查到，则到数据库查询
		userinfo, ERR := db.ImProxyClient_FindUserinfo(*config.Node.GetNetId(), *addr)
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("查询 错误:%s", ERR.String())
			return nil
		}
		if userinfo == nil {
			return nil
		}
		userInfoCache.Add(utils.Bytes2string(userinfo.Addr.GetAddr()), userinfo)
		return userinfo
	}
	userinfo := userItr.(*model.UserInfo)
	return userinfo
}

/*
用户基本信息放入缓存
*/
func CacheSetUserInfo(userinfo *model.UserInfo) {
	userInfoCache.Add(utils.Bytes2string(userinfo.Addr.GetAddr()), userinfo)
}

/*
查询指定好友在不在列表中
是否在好友列表中
*/
//func FindUserInFriendList(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
//	user, ERR := db.FindUserListByAddr(config.DBKEY_friend_userlist, Node.GetNetId(), addr)
//	if !ERR.CheckSuccess() {
//		return nil, ERR
//	}
//	return user, utils.NewErrorSuccess()
//}

/*
查询指定好友在不在列表中
对方添加自己的好友申请列表
*/
//func FindUserInApplyRemoteList(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
//	user, ERR := db.FindUserListByAddr(config.DBKEY_apply_remote_userlist, Node.GetNetId(), addr)
//	if !ERR.CheckSuccess() {
//		return nil, ERR
//	}
//	return user, utils.NewErrorSuccess()
//}

/*
查询指定好友在不在列表中
自己主动添加对方的好友申请列表
*/
//func FindUserInApplyLocalList(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
//	user, ERR := db.FindUserListByAddr(config.DBKEY_apply_local_userlist, Node.GetNetId(), addr)
//	if !ERR.CheckSuccess() {
//		return nil, ERR
//	}
//	return user, utils.NewErrorSuccess()
//}
