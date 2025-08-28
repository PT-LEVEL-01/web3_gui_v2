package im

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/sha3"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
广播上线
*/
func MulticastOnline() {
	np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	bs, err := np.Proto()
	if err != nil {
		//return nil, utils.NewErrorSysSelf(err)
		utils.Log.Error().Msgf("格式化广播消息错误:%s", err.Error())
		return
	}
	Node.SendMulticastMsg(config.MSGID_multicast_online_recv, bs)
}

/*
发送广播消息
*/
func SendMulticastMessage(content string) {
	//utils.Log.Info().Msgf("发送广播消息 长度:%d", len(content))
	//保存到管道
	msgInfo := model.NewMessageInfo(config.SUBSCRIPTION_type_msg, config.MSG_type_text, true, config.NetAddr, "",
		content, time.Now().Unix(), 0, nil, "", "", 0, nil)
	subscription.AddMuticastMsg(msgInfo)
	//utils.Log.Info().Msgf("发送广播消息")
	//bs := []byte(content)
	np := utils.NewNetParams(config.NET_protocol_version_v1, []byte(content))
	bs, err := np.Proto()
	if err != nil {
		//return nil, utils.NewErrorSysSelf(err)
		utils.Log.Error().Msgf("格式化广播消息错误:%s", err.Error())
		return
	}
	//utils.Log.Info().Msgf("发送广播消息")
	ERR := Node.SendMulticastMsg(config.MSGID_multicast_message_recv, bs)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("发送广播消息 错误:%s", ERR.String())
		return
	}
}

/*
获取好友信息
*/
func GetFriendInfoAPI(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	return SearchUserProxy(addr)
	//
	////utils.Log.Info().Msgf("获取好友信息api 11111 %s", addr.B58String())
	//np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	//bs, err := np.Proto()
	//if err != nil {
	//	//return nil, utils.NewErrorSysSelf(err)
	//	utils.Log.Error().Msgf("格式化消息错误:%s", err.Error())
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//resultBS, _, _, err := Node.SendP2pMsgHEWaitRequest(config.MSGID_get_friend_info, &addr, bs, time.Second*10)
	//if err != nil {
	//	if err == p2pconfig.ERROR_wait_msg_timeout {
	//		return nil, utils.NewErrorBus(config.ERROR_CODE_IM_user_not_exist, "") // config.ERROR_user_not_exist
	//	}
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	////utils.Log.Info().Msgf("获取好友信息api 22222")
	//if resultBS == nil || len(*resultBS) == 0 {
	//	//返回参数错误
	//	return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	//}
	//nr, err := utils.ParseNetResult(*resultBS)
	//if err != nil {
	//	utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//if !nr.CheckSuccess() {
	//	return nil, nr.ConvertERROR()
	//}
	//
	////utils.Log.Info().Msgf("公钥信息:%d %+v", len(nr.Data), nr.Data)
	//userInfo, err := model.ParseUserInfo(&nr.Data)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	////CacheSetUserInfo(userInfo)
	////utils.Log.Info().Msgf("公钥信息:%+v", userInfo)
	////解析公钥信息
	//dhPuk, ERR := config.ParseDhPukInfoV1(userInfo.GroupDHPuk)
	//if !ERR.CheckSuccess() {
	//	utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
	//	return nil, ERR
	//}
	//userInfo.GroupDHPuk = dhPuk[:]
	//return userInfo, utils.NewErrorSuccess()
}

/*
申请添加好友
*/
func AddFriendAPI(addr nodeStore.AddressNet) utils.ERROR {
	utils.Log.Info().Msgf("申请添加好友")
	//先查询对方在不在
	userinfo, ERR := SearchUserProxy(addr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//查询本地的sendindex
	sendIndex, ERR := db.ImProxyClient_FindSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_knit,
		*Node.GetNetId(), *Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if sendIndex == nil {
		//需要去远端询问自己的sendindex
		sendIndexBs, ERR := GetSendIndexForMyself(addr)
		if !ERR.CheckSuccess() {
			return ERR
		}
		if sendIndexBs != nil && len(sendIndexBs) > 0 {
			//保存这个sendIndex
			ERR = db.ImProxyClient_SaveSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_knit, *Node.GetNetId(),
				*Node.GetNetId(), addr, sendIndexBs, nil)
			if !ERR.CheckSuccess() {
				return ERR
			}
		}
	}
	//
	userinfoOld, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if userinfo.GroupDHPuk == nil || len(userinfo.GroupDHPuk) == 0 {
		userinfo.GroupDHPuk = userinfoOld.GroupDHPuk
	}
	ERR = db.ImProxyClient_SaveUserinfo(*Node.GetNetId(), userinfo)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//版本号，方便以后升级
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	dhPuk := dhKey.GetPublicKey()
	dhPukInfo, ERR := config.BuildDhPukInfoV1(dhPuk[:])
	if !ERR.CheckSuccess() {
		return ERR
	}

	//查询自己的昵称
	userSelf, ERR := db.GetSelfInfo(*Node.GetNetId())
	if !ERR.CheckSuccess() {
		return ERR
	}

	//token := ulid.Make().Bytes()

	//utils.Log.Info().Msgf("申请添加好友")
	//构建一个数据链记录
	//base := imdatachain.NewDataChainProxyBase(nil, config.IMPROXY_Command_server_forward, Node.GetNetId(), addr, addr)
	addFriend := imdatachain.NewDatachainAddFriend(*Node.GetNetId(), addr, dhPukInfo, userSelf.Nickname)
	floodKey := addFriend.GetProxyItr().GetID()
	//floodKey, ERR := config.CreateFloodKey(config.FLOOD_key_addFriend, addFriend.GetProxyItr().GetID())
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	engine.RegisterRequestKey(config.FLOOD_key_addFriend, floodKey)
	defer engine.RemoveRequestKey(config.FLOOD_key_addFriend, floodKey)
	//flood.RegisterRequestKey(config.FLOOD_key_addFriend, floodKey)
	//defer flood.RemoveRequestKey(config.FLOOD_key_addFriend, floodKey)
	utils.Log.Info().Msgf("申请添加好友")
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(addFriend.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("申请添加好友 错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("申请添加好友")

	ERRItr, ERR := engine.WaitResponseItrKey(config.FLOOD_key_addFriend, floodKey, time.Second*10)
	//ERRItr, err := flood.WaitResponseItrKey(config.FLOOD_key_addFriend, floodKey, time.Second*10)
	utils.Log.Info().Msgf("申请添加好友")
	if ERR.CheckFail() {
		return ERR
	}
	ERR = ERRItr.(utils.ERROR)
	return ERR
}

/*
好友双删
*/
func DelFriendAPI(addr nodeStore.AddressNet) utils.ERROR {
	utils.Log.Info().Msgf("删除好友")
	//先查询对方在不在好友列表中
	userinfo, ERR := db.FindUserListByAddr(*Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//不在好友列表中
	if userinfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_user_not_exist, "")
	}
	//版本号，方便以后升级
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	dhPuk := dhKey.GetPublicKey()
	dhPukInfo, ERR := config.BuildDhPukInfoV1(dhPuk[:])
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("申请添加好友")
	//构建一个数据链记录
	//base := imdatachain.NewDataChainProxyBase(nil, config.IMPROXY_Command_server_forward, Node.GetNetId(), addr, addr)
	delFriend := imdatachain.NewDatachainDelFriend(*Node.GetNetId(), addr, dhPukInfo)
	floodKey := delFriend.GetProxyItr().GetID()
	//floodKey, ERR := config.CreateFloodKey(config.FLOOD_key_addFriend, addFriend.GetProxyItr().GetID())
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	engine.RegisterRequestKey(config.FLOOD_key_addFriend, floodKey)
	defer engine.RemoveRequestKey(config.FLOOD_key_addFriend, floodKey)
	//flood.RegisterRequestKey(config.FLOOD_key_addFriend, floodKey)
	//defer flood.RemoveRequestKey(config.FLOOD_key_addFriend, floodKey)
	utils.Log.Info().Msgf("删除好友")
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(delFriend.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("申请添加好友 错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("删除好友")
	ERRItr, ERR := engine.WaitResponseItrKey(config.FLOOD_key_addFriend, floodKey, time.Second*10)
	utils.Log.Info().Msgf("删除好友")
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	ERR = ERRItr.(utils.ERROR)
	return ERR
}

/*
获取好友申请列表
*/
func GetNewFriendAPI() (*model.UserinfoListVO, utils.ERROR) {
	users, ERR := db.FindUserListApplyAll(*config.DBKEY_apply_remote_userlist, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("查询结果:%d %+v", len(users), users)
	userList := model.NewUserList()
	if users == nil {
		return model.ConverUserListVO(userList), utils.NewErrorSuccess()
	}
	for _, one := range users {
		//判断信息是否完整
		if one.Nickname == "" {
			userInfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), one.Addr)
			if ERR.CheckSuccess() {
				if userInfo != nil {
					userInfo.Status = one.Status
					userInfo.Token = one.Token
					one = *userInfo
				} else {
					//utils.Log.Info().Msgf("查询无结果")
				}
			} else {
				utils.Log.Error().Msgf("查询错误:%s", ERR.String())
			}
		}
		userList.UserList = append(userList.UserList, &one)
	}
	//utils.Log.Info().Msgf("查询结果:%d %+v", len(userList.UserList), userList)
	return model.ConverUserListVO(userList), utils.NewErrorSuccess()
}

/*
获取好友列表
*/
func GetFriendListAPI() (*model.UserinfoListVO, utils.ERROR) {
	//查询自己创建的群
	groups, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("自己创建的群列表:%+v", groups)

	//查询加入的群
	groupsJoin, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	userinfoList := model.NewUserList()
	for _, one := range append(*groups, *groupsJoin...) {
		userInfo := model.UserInfo{
			Addr:      *nodeStore.NewAddressNet(one.GroupID),
			AddrAdmin: one.AddrFrom,
			Nickname:  one.Nickname,
			//RemarksName: "",
			HeadNum: 0,
			Status:  0,
			Time:    0,
			IsGroup: true,
			GroupId: one.GroupID,
			Proxy:   new(sync.Map),
			ShoutUp: one.ShoutUp,
		}
		userInfo.Proxy.Store(utils.Bytes2string(one.ProxyMajor.GetAddr()), one.ProxyMajor)
		userinfoList.UserList = append(userinfoList.UserList, &userInfo)
	}

	users, ERR := db.FindUserListAll(*Node.GetNetId())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if users != nil && len(users) > 0 {
		for _, one := range users {
			//判断信息是否完整
			if one.Nickname == "" {
				userInfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), one.Addr)
				if !ERR.CheckSuccess() {
					utils.Log.Error().Msgf("查询 错误:%+v", ERR.String())
					continue
				}
				if userInfo != nil {
					one.Nickname = userInfo.Nickname
					one.HeadNum = userInfo.HeadNum
					one.GroupDHPuk = userInfo.GroupDHPuk
				}
			}
			//没有这个好友信息时，搜索刷新一下
			if one.Nickname == "" {
				go SearchUserProxy(one.Addr)
			}
			userinfoList.UserList = append(userinfoList.UserList, &one)
		}
	}

	//SetFriendList_cache(userinfoList.UserList)

	return model.ConverUserListVO(userinfoList), utils.NewErrorSuccess()

	//userList, err := db.GetUserList(config.DBKEY_friend_userlist)
	//if err != nil {
	//	return nil, err
	//}
	//if userList == nil {
	//	userList = model.NewUserList()
	//}
	//return model.ConverUserListVO(userList), nil
}

/*
同意好友申请
*/
func AgreeApplyFriendAPI(token []byte) utils.ERROR {
	//先查询此令牌是否在申请列表中
	user, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index,
		*Node.GetNetId(), token)
	//user, ERR := db.FindUserListApplyByIndex(config.DBKEY_apply_remote_userlist, Node.GetNetId(), uint64(createTime))
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
		return ERR
	}
	//用户不在本列表中
	if user == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_user_not_exist, "")
	}
	//版本号，方便以后升级
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	dhPuk := dhKey.GetPublicKey()
	dhPukInfo, ERR := config.BuildDhPukInfoV1(dhPuk[:])
	if !ERR.CheckSuccess() {
		return ERR
	}

	//找到发送者id，才是对方的token
	proxyItr, ERR := db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), *Node.GetNetId(), user.Token)
	if !ERR.CheckSuccess() {
		return ERR
	}

	agreeFriend := imdatachain.NewDatachainAgreeFriend(*Node.GetNetId(), user.Addr, user.Token, proxyItr.GetSendID(), dhPukInfo)
	floodKey := agreeFriend.GetProxyItr().GetID()
	engine.RegisterRequestKey(config.FLOOD_key_agreeFriend, floodKey)
	defer engine.RemoveRequestKey(config.FLOOD_key_agreeFriend, floodKey)
	//flood.RegisterRequestKey(config.FLOOD_key_agreeFriend, floodKey)
	//defer flood.RemoveRequestKey(config.FLOOD_key_agreeFriend, floodKey)
	utils.Log.Info().Msgf("同意好友申请")
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(agreeFriend.GetProxyItr())
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("保存数据链错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("同意好友申请")
	ERRItr, ERR := engine.WaitResponseItrKey(config.FLOOD_key_agreeFriend, floodKey, time.Second*10)
	utils.Log.Info().Msgf("同意好友申请")
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	ERR = ERRItr.(utils.ERROR)
	return ERR
}

/*
设置好友昵称备注
*/
func SetFriendRemarksName(addr nodeStore.AddressNet, remarksName string) utils.ERROR {
	if len(remarksName) > config.IM_nickname_length_max {
		ERR := utils.NewErrorBus(config.ERROR_CODE_IM_nickname_over_size, "remarksName")
		return ERR
	}
	//先查询对方在不在好友列表中
	userinfo, ERR := db.FindUserListByAddr(*Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//不在好友列表中
	if userinfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_user_not_exist, "")
	}

	clientRemarksname := imdatachain.NewClientRemarksname(*Node.GetNetId(), addr, remarksName)
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(clientRemarksname.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询聊天历史记录
*/
func GetChatHistoryAPI(startIndex []byte, count uint64, remoteAddr nodeStore.AddressNet) (*model.MessageContentListVO, utils.ERROR) {
	mcList, ERR := db.GetMessageHistoryV2(startIndex, count, *Node.GetNetId(), remoteAddr)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if mcList == nil {
		mcList = model.NewMessageContentList()
	}
	//判断base64编码图片
	for _, mcOne := range mcList.List {
		//是base64编码的图片，则查询并拼接
		ERR = JoinImageBase64(mcOne)
		if ERR.CheckFail() {
			return nil, ERR
		}
	}
	//给聊天记录加昵称
	for _, one := range mcList.List {
		info := subscription.CacheGetUserInfo(&one.From)
		if info == nil {
			continue
		}
		one.Nickname = info.Nickname
	}
	list := model.ConverMessageContentListVO(mcList)
	return list, utils.NewErrorSuccess()
}

/*
拼接图片
*/
func JoinImageBase64(mcOne *model.MessageContent) utils.ERROR {
	//传送完才拼接，未传送完的不拼接
	if mcOne.FileType != config.FILE_type_image_base64 || mcOne.FileBlockTotal != mcOne.FileBlockIndex+1 {
		return utils.NewErrorSuccess()
	}
	var sharekey []byte
	var ERR utils.ERROR
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	dhPrk := dhKey.GetPrivateKey()
	//dhPrk := Node.Keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	if mcOne.IsGroup {
		//获取群共享密钥
		sharekey, ERR = StaticProxyClientManager.groupParserManager.GetShareKey(mcOne.To.GetAddr())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("拼接图片:%+v %+v", mcOne.To, sharekey)
	} else {
		addrRemote := mcOne.From
		if mcOne.FromIsSelf {
			addrRemote = mcOne.To
		}
		dhPuk, ERR := db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), addrRemote)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("加密用公私密钥:%+v %+v", dhPrk, dhPuk)
		//生成共享密钥sharekey
		shareKey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, *dhPuk))
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		sharekey = shareKey[:]
	}

	//utils.Log.Info().Msgf("计算协商密钥 私钥:%s 公钥:%s key:%s", hex.EncodeToString(dhPrk[:]), hex.EncodeToString((*dhPuk)[:]), hex.EncodeToString(sharekey[:]))
	for _, sendId := range mcOne.FileContent {
		var proxyItrOne imdatachain.DataChainProxyItr
		if mcOne.IsGroup {
			//utils.Log.Info().Msgf("查询群文件 拼接图片:%+v %+v", mcOne.To, sendId)
			proxyItrOne, ERR = db.ImProxyClient_FindGroupDataChainByID(*Node.GetNetId(), mcOne.To.GetAddr(), sendId)
		} else {
			proxyItrOne, ERR = db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), *Node.GetNetId(), sendId)
		}
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			return ERR
		}
		//utils.Log.Info().Msgf("共享密钥:%p", proxyItrOne)
		ERR = proxyItrOne.DecryptContent(sharekey)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		clientItr := proxyItrOne.GetClientItr()
		file := clientItr.(*imdatachain.DatachainFile)

		//imgBasea64Str, ERR := config.ParseImgBase64InfoV1(file.Block)
		//if !ERR.CheckSuccess() {
		//	utils.Log.Error().Msgf("错误:%s", ERR.String())
		//	return ERR
		//}
		//content := []byte(imgBasea64Str)
		mcOne.Content = append(mcOne.Content, file.Block...)
	}
	imgBasea64Str := config.BuildImgBase64(mcOne.FileMimeType, mcOne.Content)
	mcOne.Content = []byte(imgBasea64Str)
	return utils.NewErrorSuccess()
}

/*
拼接文件
*/
func JoinFile(mcOne *model.MessageContent, filePath string) (string, utils.ERROR) {
	dirName := ""
	if mcOne.IsGroup {
		dirName = mcOne.To.B58String()
	} else {
		addrRemote := mcOne.From
		if mcOne.FromIsSelf {
			addrRemote = mcOne.To
		}
		dirName = addrRemote.B58String()
	}

	if mcOne.FileType != config.FILE_type_file {
		return "", utils.NewErrorSuccess()
	}
	var ERR utils.ERROR
	var sharekey []byte
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return "", ERR
	}
	dhPrk := dhKey.GetPrivateKey()
	//dhPrk := Node.Keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	if mcOne.IsGroup {
		//获取群共享密钥
		sharekey, ERR = StaticProxyClientManager.groupParserManager.GetShareKey(mcOne.To.GetAddr())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
			return "", ERR
		}
	} else {
		addrRemote := mcOne.From
		if mcOne.FromIsSelf {
			addrRemote = mcOne.To
		}
		dhPuk, ERR := db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), addrRemote)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return "", ERR
		}
		//utils.Log.Info().Msgf("加密用公私密钥:%+v %+v", dhPrk, dhPuk)
		//生成共享密钥sharekey
		shareKey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, *dhPuk))
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return "", utils.NewErrorSysSelf(err)
		}
		sharekey = shareKey[:]
	}

	//dhPrk := Node.Keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	//dhPuk, ERR := db.ImProxyClient_FindUserDhPuk(Node.GetNetId(), remoteAddr)
	//if !ERR.CheckSuccess() {
	//	utils.Log.Error().Msgf("错误:%s", ERR.String())
	//	return "", ERR
	//}
	////utils.Log.Info().Msgf("加密用公私密钥:%+v %+v", dhPrk, dhPuk)
	////生成共享密钥sharekey
	//sharekey, err := dh.KeyExchange(dh.NewDHPair(dhPrk, *dhPuk))
	//if err != nil {
	//	utils.Log.Error().Msgf("错误:%s", err.Error())
	//	return "", utils.NewErrorSysSelf(err)
	//}
	fileName := filePath
	if filePath == "" {
		fileName = filepath.Join(config.DownloadFileDir, dirName, mcOne.FileName)
	}
	dir, fileName := filepath.Split(fileName)
	err := utils.CheckCreateDir(dir)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return "", utils.NewErrorSysSelf(err)
	}
	fileName = filepath.Join(dir, fileName)
	file, err := os.Create(fileName)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return "", utils.NewErrorSysSelf(err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)

	//utils.Log.Info().Msgf("计算协商密钥 私钥:%s 公钥:%s key:%s", hex.EncodeToString(dhPrk[:]), hex.EncodeToString((*dhPuk)[:]), hex.EncodeToString(sharekey[:]))
	for _, sendId := range mcOne.FileContent {
		//utils.Log.Info().Msgf("解密文件消息:%+v %+v", sendId, sharekey)
		var proxyItrOne imdatachain.DataChainProxyItr
		if mcOne.IsGroup {
			proxyItrOne, ERR = db.ImProxyClient_FindGroupDataChainByID(*Node.GetNetId(), mcOne.To.GetAddr(), sendId)
		} else {
			proxyItrOne, ERR = db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), *Node.GetNetId(), sendId)
		}
		if !ERR.CheckSuccess() {
			//utils.Log.Info().Msgf("保存记录")
			return "", ERR
		}
		//utils.Log.Info().Msgf("解密文件消息:%p", proxyItrOne)
		ERR = proxyItrOne.DecryptContent(sharekey[:])
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return "", ERR
		}
		clientItr := proxyItrOne.GetClientItr()
		fileDataChain := clientItr.(*imdatachain.DatachainFile)

		_, err := writer.Write(fileDataChain.Block)
		if err != nil && err != io.EOF {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return "", utils.NewErrorSysSelf(err)
		}
		err = writer.Flush()
		if err != nil && err != io.EOF {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return "", utils.NewErrorSysSelf(err)
		}

		//imgBasea64Str, ERR := config.ParseImgBase64InfoV1(fileDataChain.Block)
		//if !ERR.CheckSuccess() {
		//	utils.Log.Error().Msgf("错误:%s", ERR.String())
		//	return ERR
		//}
		//content := []byte(imgBasea64Str)
		//mcOne.Content = append(mcOne.Content, content...)
	}
	return fileName, utils.NewErrorSuccess()
}

/*
给好友发送消息
*/
func SendFriendMsgAPI_old(addr nodeStore.AddressNet, content string, t uint64, pullAndPushID uint64, quoteID string) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")
	//检查引用id是否合法
	var quotoIDBS []byte
	var err error
	if quoteID != "" {
		quotoIDBS, err = hex.DecodeString(quoteID)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	}
	//id := ulid.Make()
	//utils.Log.Info().Msgf("给好友发送消息：%s", content)
	messageOne := &model.MessageContent{
		Type:          t,                             //消息类型
		FromIsSelf:    true,                          //是否自己发出的
		From:          *Node.GetNetId(),              //发送者
		To:            addr,                          //接收者
		Content:       []byte(content),               //消息内容
		Time:          time.Now().Unix(),             //时间
		SendID:        ulid.Make().Bytes(),           //
		QuoteID:       quotoIDBS,                     //
		State:         config.MSG_GUI_state_not_send, //
		PullAndPushID: pullAndPushID,                 //
	}
	//utils.Log.Info().Msgf("发送消息：%+v", messageOne)
	//保存未发送状态的消息
	messageOne, ERR := db.AddMessageHistory_old(messageOne)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("发送消息:%s", ERR.String())
		return ERR
	}
	go sendMsgSync(messageOne)
	return utils.NewErrorSuccess()
}

/*
给好友发送base64编码图片
*/
func SendImageBase64(addr nodeStore.AddressNet, imgBase64 string) utils.ERROR {

	mimeType, imgBinary, ERR := config.ParseImgBase64(imgBase64)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送文本消息 错误:%s", ERR.String())
		return ERR
	}

	sendFileInfo := imdatachain.SendFileInfo{
		Addr:        addr,              //
		FileNameAbs: "",                //真实文件路径
		SendTime:    time.Now().Unix(), //发送时间
		MimeType:    mimeType,          //文件类型
		//FileSize    uint64 //文件总大小
		//BlockSize   uint64 //块大小
		//Hash        []byte //文件hash
		//BlockTotal  uint64 //文件块总数
		//BlockIndex  uint64 //文件块编号，从0开始，连续增长的整数
		Block:   imgBinary, //文件块内容
		IsGroup: false,
	}

	fileSize := utils.Byte(len(imgBinary))
	//utils.Log.Info().Msgf("发送base64图片")

	fileHash := utils.Hash_SHA3_256(imgBinary)
	fileName := hex.EncodeToString(fileHash)
	fileTotal := utils.Byte(fileSize) / config.DataChainBlockContentSize
	if fileSize%config.DataChainBlockContentSize != 0 {
		fileTotal++
	}

	sendFileInfo.Hash = fileHash
	sendFileInfo.Name = fileName
	sendFileInfo.FileSize = uint64(fileSize)
	sendFileInfo.BlockTotal = uint64(fileTotal)
	sendFileInfo.BlockSize = uint64(config.DataChainBlockContentSize)

	ERR = db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), &sendFileInfo)
	if ERR.CheckFail() {
		return ERR
	}

	StaticProxyClientManager.ParserClient.SendFile()

	return utils.NewErrorSuccess()
}

/*
给好友发送文件
*/
func SendFile(addr nodeStore.AddressNet, filePath string) utils.ERROR {
	//utils.Log.Info().Msgf("判断是否是图片")

	//判断是否是图片
	ok, mimeType, fileinfo, err := utils.FileContentTypeIsImage(filePath)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}

	utils.Log.Info().Str("文件类型", mimeType).Send()

	sendFileInfo := imdatachain.SendFileInfo{
		Addr:        addr,              //
		FileNameAbs: filePath,          //真实文件路径
		SendTime:    time.Now().Unix(), //发送时间
		MimeType:    mimeType,          //文件类型
		//FileSize    uint64 //文件总大小
		//BlockSize   uint64 //块大小
		//Hash        []byte //文件hash
		//BlockTotal  uint64 //文件块总数
		//BlockIndex  uint64 //文件块编号，从0开始，连续增长的整数
		//Block       []byte //文件块内容
		IsGroup: false,
	}

	utils.Log.Info().Msgf("判断是否是图片:%t %d %d", ok, fileinfo.Size(), config.FILE_image_size_max)
	//是图片，并且在规定范围内，可以直接显示成图片

	file, err := os.Open(filePath)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	defer file.Close()
	//计算文件hash
	hash_sha3 := sha3.New256()
	_, err = io.Copy(hash_sha3, file)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	fileHash := hash_sha3.Sum(nil)

	utils.Log.Info().Msgf("发送的文件hash:%s", hex.EncodeToString(fileHash))
	fileinfo, err = file.Stat()
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//fileHash := utils.Hash_SHA3_256(imgBinary)
	_, fileName := filepath.Split(filePath)
	//fileName := file.Name()
	fileSize := uint64(fileinfo.Size())
	fileTotal := imdatachain.BuildBlockTotal(fileSize, uint64(config.DataChainBlockContentSize))
	sendFileInfo.Hash = fileHash
	sendFileInfo.Name = fileName
	sendFileInfo.FileSize = fileSize
	sendFileInfo.BlockTotal = fileTotal
	sendFileInfo.BlockSize = uint64(config.DataChainBlockContentSize)
	utils.Log.Info().Msgf("文件拆分个数:%d %d %d", fileSize, config.DataChainBlockContentSize, fileTotal)
	utils.Log.Info().Msgf("文件拆分个数:%+v", sendFileInfo)

	ERR := db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), &sendFileInfo)
	if ERR.CheckFail() {
		return ERR
	}
	StaticProxyClientManager.ParserClient.SendFile()
	utils.Log.Info().Msgf("发送文件完成")
	return utils.NewErrorSuccess()
}

/*
给好友发送文件
*/
func SendVoice(addr nodeStore.AddressNet, filePath, mimeType, voiceCoding string, voiceSecond int64) utils.ERROR {
	//utils.Log.Info().Msgf("判断是否是图片")
	var bs []byte
	var fileName = ""
	if voiceCoding == "" {
		//判断是否是图片
		ok, mType, fileinfo, err := utils.FileContentTypeIsImage(filePath)
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		utils.Log.Info().Str("文件类型", mType).Send()

		utils.Log.Info().Msgf("判断是否是图片:%t %d %d", ok, fileinfo.Size(), config.FILE_image_size_max)
		//是图片，并且在规定范围内，可以直接显示成图片

		fileName = fileinfo.Name()
		mimeType = mType

		//检查大小是否超限
		if fileinfo.Size() >= int64(config.DataChainBlockContentSize) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_size_max_over, strconv.Itoa(int(config.DataChainBlockContentSize)))
		}
		bs, err = os.ReadFile(filePath)
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
	} else {
		fileName = ulid.Make().String()
	}
	voice := imdatachain.NewDatachainVoice(*Node.GetNetId(), addr, fileName, mimeType, voiceSecond, bs, voiceCoding)
	//保存并解析数据链记录
	ERR := StaticProxyClientManager.ParserClient.SaveDataChain(voice.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送文本消息 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
群发送base64编码图片
*/
func GroupSendImageBase64(groupId []byte, imgBase64 string) utils.ERROR {
	//utils.Log.Info().Msgf("发送base64图片")
	//查询群是否存在
	groupInfo, ERR := db.FindGroupInfoAllList(*Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	//查询是否是本群成员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if userInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	mimeType, imgBinary, ERR := config.ParseImgBase64(imgBase64)
	if ERR.CheckFail() {
		return ERR
	}

	sendFileInfo := imdatachain.SendFileInfo{
		Addr:        *nodeStore.NewAddressNet(groupId), //
		FileNameAbs: "",                                //真实文件路径
		SendTime:    time.Now().Unix(),                 //发送时间
		MimeType:    mimeType,                          //文件类型
		//FileSize    uint64 //文件总大小
		//BlockSize   uint64 //块大小
		//Hash        []byte //文件hash
		//BlockTotal  uint64 //文件块总数
		//BlockIndex  uint64 //文件块编号，从0开始，连续增长的整数
		Block:   imgBinary, //文件块内容
		IsGroup: true,
	}

	fileSize := utils.Byte(len(imgBinary))
	//utils.Log.Info().Msgf("发送base64图片")

	fileHash := utils.Hash_SHA3_256(imgBinary)
	fileName := hex.EncodeToString(fileHash)
	fileTotal := utils.Byte(fileSize) / config.DataChainBlockContentSize
	if fileSize%config.DataChainBlockContentSize != 0 {
		fileTotal++
	}

	sendFileInfo.Hash = fileHash
	sendFileInfo.Name = fileName
	sendFileInfo.FileSize = uint64(fileSize)
	sendFileInfo.BlockTotal = uint64(fileTotal)
	sendFileInfo.BlockSize = uint64(config.DataChainBlockContentSize)

	ERR = db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), &sendFileInfo)
	if ERR.CheckFail() {
		return ERR
	}

	groupParser, ERR := StaticProxyClientManager.groupParserManager.GetGroupParse(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	//发送给群
	groupParser.SendFile()

	return utils.NewErrorSuccess()
}

/*
群发送文件
*/
func GroupSendFile(groupId []byte, filePath string) utils.ERROR {
	//查询群是否存在
	groupInfo, ERR := db.FindGroupInfoAllList(*Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	//查询是否是本群成员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if userInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	//utils.Log.Info().Msgf("判断是否是图片")
	//判断是否是图片
	ok, mimeType, fileinfo, err := utils.FileContentTypeIsImage(filePath)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}

	utils.Log.Info().Str("文件类型", mimeType).Send()

	sendFileInfo := imdatachain.SendFileInfo{
		Addr:        *nodeStore.NewAddressNet(groupId), //
		FileNameAbs: filePath,                          //真实文件路径
		SendTime:    time.Now().Unix(),                 //发送时间
		MimeType:    mimeType,                          //文件类型
		//FileSize    uint64 //文件总大小
		//BlockSize   uint64 //块大小
		//Hash        []byte //文件hash
		//BlockTotal  uint64 //文件块总数
		//BlockIndex  uint64 //文件块编号，从0开始，连续增长的整数
		//Block       []byte //文件块内容
		IsGroup: true,
	}

	utils.Log.Info().Msgf("判断是否是图片:%t %d %d", ok, fileinfo.Size(), config.FILE_image_size_max)
	//是图片，并且在规定范围内，可以直接显示成图片
	//if ok {
	//	if fileinfo.Size() <= config.FILE_image_size_max {
	//		sendFileInfo.Type = config.FILE_type_image_base64
	//		//imgByte, err := os.ReadFile(filePath)
	//		//if err != nil {
	//		//	utils.Log.Error().Msgf("读取图片错误:%s", err.Error())
	//		//	return utils.NewErrorSysSelf(err)
	//		//}
	//		//imgBase64 := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(imgByte)
	//		////utils.Log.Info().Msgf("图片base64:%s", imgBase64)
	//		//return GroupSendImageBase64(groupId, imgBase64)
	//	} else {
	//		sendFileInfo.Type = config.FILE_type_image_binary
	//	}
	//}
	//当作文件发送
	//utils.Log.Info().Msgf("发送普通文件")

	file, err := os.Open(filePath)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	defer file.Close()
	//计算文件hash
	hash_sha3 := sha3.New256()
	_, err = io.Copy(hash_sha3, file)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	fileHash := hash_sha3.Sum(nil)

	//utils.Log.Info().Msgf("发送的文件hash:%s", hex.EncodeToString(fileHash))
	fileinfo, err = file.Stat()
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//fileHash := utils.Hash_SHA3_256(imgBinary)
	_, fileName := filepath.Split(filePath)
	//fileName := file.Name()
	fileSize := uint64(fileinfo.Size())
	fileTotal := imdatachain.BuildBlockTotal(fileSize, uint64(config.DataChainBlockContentSize))
	sendFileInfo.Hash = fileHash
	sendFileInfo.Name = fileName
	sendFileInfo.FileSize = fileSize
	sendFileInfo.BlockTotal = fileTotal
	sendFileInfo.BlockSize = uint64(config.DataChainBlockContentSize)
	//utils.Log.Info().Msgf("文件拆分个数:%d %d %d", fileSize, config.DataChainBlockContentSize, fileTotal)
	//utils.Log.Info().Msgf("文件拆分个数:%+v", sendFileInfo)

	ERR = db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), &sendFileInfo)
	if ERR.CheckFail() {
		return ERR
	}

	groupParser, ERR := StaticProxyClientManager.groupParserManager.GetGroupParse(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	//发送给群
	groupParser.SendFile()

	utils.Log.Info().Msgf("发送文件完成")
	return utils.NewErrorSuccess()
}

/*
构建出完整的文件到指定文件夹
*/
func JoinBuildFile(addr nodeStore.AddressNet, sendId []byte, filePath string) (string, utils.ERROR) {
	mc, ERR := db.FindMessageHistoryBySendIdV2(*Node.GetNetId(), addr, sendId)
	if ERR.CheckFail() {
		return "", ERR
	}
	if mc.TransProgress != 100 {
		return "", utils.NewErrorSuccess()
	}
	return JoinFile(mc, filePath)
}

/*
给好友发送文本消息
*/
func SendTextAPI(addr nodeStore.AddressNet, content string, quoteID string) utils.ERROR {
	utils.Log.Info().Msgf("发送文本消息")
	//检查引用id是否合法
	var quotoIDBS []byte
	var err error
	if quoteID != "" {
		quotoIDBS, err = hex.DecodeString(quoteID)
		if err != nil {
			ERR := utils.NewErrorSysSelf(err)
			utils.Log.Error().Msgf("发送文本消息 错误:%s", ERR.String())
			return ERR
		}
	}

	//先重发之前的消息，有则重发，无则不用发
	//StaticProxyClientManager.RetrySendDataChain(Node.GetNetId(), addr)

	sendText := imdatachain.NewDataChainSendText(*Node.GetNetId(), addr, content, quotoIDBS)
	//保存并解析数据链记录
	ERR := StaticProxyClientManager.ParserClient.SendText(sendText.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送文本消息 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
再次给好友发送消息
*/
func SendFriendMsgAgainAPIV2(fromAddr, toAddr nodeStore.AddressNet, sendID []byte) utils.ERROR {
	StaticProxyClientManager.RetrySendDataChain(fromAddr, toAddr)
	return utils.NewErrorSuccess()
}

/*
再次给好友发送消息
*/
func SendFriendMsgAgainAPI(fromAddr, toAddr nodeStore.AddressNet, sendID []byte) utils.ERROR {
	messageOne := &model.MessageContent{
		FromIsSelf: true,                          //是否自己发出的
		From:       fromAddr,                      //发送者
		To:         toAddr,                        //接收者
		SendID:     sendID,                        //
		State:      config.MSG_GUI_state_not_send, //
	}
	//修改为未发送状态
	ERR := db.UpdateSendMessageState(messageOne)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("发送消息:%s", ERR.String())
		return ERR
	}
	messageOne, ERR = db.FindMessageContentSend(fromAddr, toAddr, sendID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("重发消息:%+v", messageOne)

	go sendMsgSync(messageOne)
	return utils.NewErrorSuccess()
}

func sendMsgSync(messageOne *model.MessageContent) {
	//先往前端推送一次未发送消息
	mi := messageOne.ConverVO()
	mi.Subscription = config.SUBSCRIPTION_type_msg
	mi.State = config.MSG_GUI_state_not_send
	subscription.AddSubscriptionMsg(mi)
	//utils.Log.Info().Msgf("发送消息")

	mi = messageOne.ConverVO()
	mi.Subscription = config.SUBSCRIPTION_type_msg
	//发送P2P消息
	ERR := sendMsg(messageOne)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送消息错误:%s", ERR.String())
		//再往前端推送一次发送失败状态
		mi.State = config.MSG_GUI_state_fail
		messageOne.State = config.MSG_GUI_state_fail
	} else {
		//utils.Log.Info().Msgf("发送消息")
		mi.State = config.MSG_GUI_state_success
		messageOne.State = config.MSG_GUI_state_success
	}
	//utils.Log.Info().Msgf("发送消息")
	ERR = db.UpdateSendMessageState(messageOne)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改消息状态错误:%s", ERR.String())
		//再往前端推送一次发送失败状态
		mi.State = config.MSG_GUI_state_fail
	}
	subscription.AddSubscriptionMsg(mi)
}

/*
从好友那里查询共享文件列表
*/
func GetShareboxList(addr nodeStore.AddressNet, dirPath string) (*model.FileList, utils.ERROR) {
	//utils.Log.Info().Msgf("从好友那里查询共享文件列表:%s", dirPath)
	dirPath = strings.Trim(dirPath, " ") //去掉前后空格
	if dirPath != "" {
		dirPath = filepath.Clean(dirPath)     //统一格式化为"\file1\file2"
		dirPath = strings.Trim(dirPath, "\\") //去掉前后的"\"符号
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, []byte(dirPath))
	bs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_file_getShareboxList, &addr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("从好友那里查询共享文件列表 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}

	//utils.Log.Info().Msgf("从好友那里查询共享文件列表:%s", dirPath)
	fileList, err := model.ParseFilelist(&nr.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysRemote(err.Error())
	}
	if fileList == nil {
		return nil, utils.NewErrorSuccess()
	}
	return fileList, utils.NewErrorSuccess()
}

/*
从好友那里查询博客类别列表
*/
func GetUserCircleClassNames(addr nodeStore.AddressNet) ([]model.ClassCount, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	bs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_circle_getClassNames, &addr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	// utils.Log.Info().Msgf("从好友那里查询博客圈子列表:%s", dirPath)
	classCounts, err := model.ParseClassNames(&nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return classCounts, utils.NewErrorSuccess()
}

/*
客户端向代理节点获取一个租用时间的订单
@serverAddr    uint64     //存储提供商地址
@useTime       uint64     //空间使用时间 单位：1天
*/
func GetProxyOrders(serverAddr nodeStore.AddressNet, spaceTotal, useTime uint64) (*model.OrderForm, utils.ERROR) {
	order := &model.OrderForm{
		UserAddr:   *Node.GetNetId(), //这里填别人的地址，相当于赠送给别人
		SpaceTotal: spaceTotal,
		UseTime:    useTime,
	}
	bs, err := order.Proto()
	//utils.Log.Info().Msgf("本次购买空间:%d %d %+v", spaceTotal, order.SpaceTotal, bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_getorders, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	order, err = model.ParseOrderForm(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("返回订单:%+v", order)
	//if order.TotalPrice > 0 {
	//	ERR := StaticProxyClientManager.AddOrdersUnpaid(order)
	//	return order, ERR
	//}
	//免支付订单
	ERR = StaticProxyClientManager.AddOrders(order)
	return order, ERR
}

/*
获取续费订单
@serverAddr    uint64     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func GetProxyRenewalOrders(preNumber []byte, serverAddr nodeStore.AddressNet, useTime uint64) (*model.OrderForm, utils.ERROR) {
	order := &model.OrderForm{
		PreNumber: preNumber,
		UseTime:   useTime,
	}
	bs, err := order.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_getRenewalOrders, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}

	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	order, err = model.ParseOrderForm(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if order.TotalPrice > 0 {
		ERR := db.StorageClient_SaveOrderFormNotPay(order)
		return order, ERR
	}
	ERR = db.StorageClient_SaveOrderFormInUse(order)
	return order, ERR
}

/*
先查询本地数据库中的节点信息，本地没有则发送消息网络中搜索
*/
func FindAndSearchUserProxy(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if userinfo != nil {
		return userinfo, utils.NewErrorSuccess()
	}
	userinfo, ERR = SearchUserProxy(addr)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
广播并搜索一个节点的信息及代理节点
*/
func SearchUserProxy(addr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	//utils.Log.Info().Msgf("发送广播搜索用户信息:%s", addr.B58String())
	if bytes.Equal(Node.GetNetId().GetAddr(), addr.GetAddr()) {
		userinfo, ERR := db.GetSelfInfo(*Node.GetNetId())
		if ERR.CheckFail() {
			return nil, ERR
		}
		return userinfo, utils.NewErrorSuccess()
	}
	//判断要搜索的节点刚好是自己在代理
	spacesTotal, _, _ := StaticProxyServerManager.AuthManager.QueryUserSpaces(addr)
	if spacesTotal > 0 {
		userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), addr)
		if ERR.CheckFail() {
			return nil, ERR
		}
		//是自己的代理节点，则添加到代理列表中后返回给查询用户
		userinfo.Proxy.Store(utils.Bytes2string(Node.GetNetId().GetAddr()), Node.GetNetId())
		ERR = db.ImProxyClient_SaveUserinfo(*Node.GetNetId(), userinfo)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		return userinfo, utils.NewErrorSuccess()
	}

	//bs := []byte(content)
	np := utils.NewNetParams(config.NET_protocol_version_v1, addr.GetAddr())
	bs, err := np.Proto()
	if err != nil {
		//return nil, utils.NewErrorSysSelf(err)
		utils.Log.Error().Msgf("格式化广播消息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	//flood.RegisterRequest(utils.Bytes2string(addr))
	//defer flood.RemoveRequest(utils.Bytes2string(addr))
	//utils.Log.Info().Msgf("发送广播消息")
	bs, ERR := Node.SendMulticastMsgWaitRequest(config.MSGID_IM_PROXY_multicast_search, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("发送广播搜索用户信息 错误:%s", ERR.String())
		return nil, ERR
	}
	if bs == nil {
		utils.Log.Info().Msgf("发送广播搜索用户信息 返回参数错误:%s", addr.B58String())
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}

	//utils.Log.Info().Msgf("返回个人信息")
	np, err = utils.ParseNetParams(*bs)
	if err != nil {
		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	userinfo, err := model.ParseUserInfo(&np.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析用户信息 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}

	//解析公钥信息
	keys, ERR := utilsleveldb.LeveldbParseKeyMore(userinfo.GroupDHPuk)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
		return nil, utils.NewErrorSysSelf(err)
	}
	//解析公钥版本号
	versionBs, ERR := keys[0].BaseKey()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	version := utils.BytesToUint64ByBigEndian(versionBs)
	if version != config.DHPUK_version_1 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_IM_dh_version_unknown, "")
	}
	//解析公钥内容
	dhPuk, ERR := keys[1].BaseKey()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	userinfo.GroupDHPuk = dhPuk

	//utils.Log.Info().Msgf("返回个人信息")
	//flood.ResponseItr(utils.Bytes2string(userinfo.Addr), userinfo)
	//合并代理节点并保存
	//MulticastSearchProxy_recvLock.Lock()
	//defer MulticastSearchProxy_recvLock.Unlock()
	userinfoDB, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), userinfo.Addr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("查询用户信息 错误:%s", err.Error())
		return nil, ERR
	}

	//utils.Log.Info().Msgf("返回个人信息")
	userinfo.Proxy.Range(func(k, v interface{}) bool {
		userinfoDB.Proxy.Store(k, v)
		return true
	})
	//utils.Log.Info().Msgf("返回个人信息:%+v", userinfo)
	ERR = db.ImProxyClient_SaveUserinfo(*Node.GetNetId(), userinfo)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("保存用户信息 错误:%s", ERR.String())
		return nil, ERR
	}
	//个人信息有变化
	if userinfoDB == nil || userinfo.Nickname != userinfoDB.Nickname {
		//推送给前端，刷新好友列表
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
		subscription.AddSubscriptionMsg(&msgInfo)
	}

	utils.Log.Info().Msgf("发送广播搜索用户信息:%s", addr.B58String())
	return userinfo, utils.NewErrorSuccess()
}

/*
获取代理节点同步数据链高度
*/
func GetProxyDataChainIndex(serverAddr nodeStore.AddressNet) (*big.Int, utils.ERROR) {
	utils.Log.Info().Msgf("获取远端已经同步的index：%s", serverAddr.B58String())
	np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	bs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_getDataChainIndex, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取代理节点同步数据链高度 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	if nr.Data == nil || len(nr.Data) == 0 {
		return big.NewInt(0), utils.NewErrorSuccess()
	}
	//返回成功
	index := new(big.Int).SetBytes(nr.Data)
	return index, utils.NewErrorSuccess()
}

/*
上传数据链
*/
func UploadDataChain(serverAddr nodeStore.AddressNet, itrs []imdatachain.DataChainProxyItr) utils.ERROR {
	byteList := make([][]byte, 0, len(itrs))
	for _, one := range itrs {
		bs, err := one.Proto()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		byteList = append(byteList, *bs)
	}
	bs, err := model.BytesProto(byteList, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, bs)
	nbs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_upload_datachain, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传数据链 错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
下载数据链，一部分是同步未引入数据链的消息，一部分是多端原因引起的已经引入数据链的消息
@serverAddr    nodeStore.AddressNet    代理节点地址
@index         []byte                  本地索引最新高度
@return    []imdatachain.DataChainProxyItr    未引入数据链的新消息
@return    []imdatachain.DataChainProxyItr    已经引入数据链的消息
@return    utils.ERROR                     错误
*/
func DownloadDataChain(serverAddr nodeStore.AddressNet, index []byte) ([]imdatachain.DataChainProxyItr, []imdatachain.DataChainProxyItr, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, index)
	nbs, err := np.Proto()
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_download_datachain, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传数据链 错误:%s", ERR.String())
		return nil, nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nil, nr.ConvertERROR()
	}
	//返回成功
	list, list2, err := model.ParseBytes(nr.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(list))
	for _, one := range list {
		itr, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		itrs = append(itrs, itr)
	}
	itrs2 := make([]imdatachain.DataChainProxyItr, 0, len(list))
	for _, one := range list2 {
		itr, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		itrs2 = append(itrs2, itr)
	}
	return itrs, itrs2, utils.NewErrorSuccess()
}

/*
同步数据链，服务器从客户端同步数据链来保存
@serverAddr    nodeStore.AddressNet           客户端节点地址
@index         []byte                         本地索引最新高度
@return    []imdatachain.DataChainProxyItr    数据链的消息
@return    utils.ERROR                        错误
*/
func SyncDataChain(serverAddr nodeStore.AddressNet, index []byte) ([]imdatachain.DataChainProxyItr, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, index)
	nbs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_sync_datachain, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传数据链 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	list, _, err := model.ParseBytes(nr.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(list))
	for _, one := range list {
		itr, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		itrs = append(itrs, itr)
	}
	return itrs, utils.NewErrorSuccess()
}

/*
发送一条消息给对方节点本人或代理节点
*/
func SendDatachainMsgToClientOrProxy(addrTo nodeStore.AddressNet, itr imdatachain.DataChainProxyItr) utils.ERROR {
	utils.Log.Info().Msgf("发送私聊消息:%s", addrTo.B58String())
	ERR := utils.NewErrorSuccess()
	//utils.Log.Info().Msgf("开始发送消息")
	if !itr.Forward() {
		ERR = utils.NewErrorSuccess()
		return ERR
	}
	//要发送的节点就是自己
	if bytes.Equal(addrTo.GetAddr(), Node.GetNetId().GetAddr()) {
		return StaticProxyServerManager.CheckSaveDataChainNolink(itr)
	}

	bs, err := itr.Proto()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	nbs, err := np.Proto()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	//toAddr := itr.GetAddrTo()
	//utils.Log.Info().Msgf("开始发送消息:%s", toAddr.B58String())
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_send_datachain, &addrTo, nbs, time.Second*20)
	if ERR.CheckFail() {
		//这里的p2p网络有个问题，当对方下线再上线会清除加密缓存，而自己未下线时使用的旧加密key，对方无法解密。故第一次发送失败重发一次。
		utils.Log.Info().Msgf("发送消息 错误:%s", ERR.String())
		resultBS, ERR = Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_send_datachain, &addrTo, nbs, time.Second*20)
		if ERR.CheckFail() {
			utils.Log.Info().Msgf("发送消息 错误:%s", ERR.String())
			return ERR
		}
	}
	//utils.Log.Info().Msgf("消息有返回")
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		ERR = utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
		return ERR
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Error().Msgf("解析远端参数 错误:%s", err.Error())
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	if !nr.CheckSuccess() {
		ERR = nr.ConvertERROR()
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
		}
		return ERR
	}
	//返回成功
	return ERR
}

/*
发送群消息给代理节点
*/
func SendGroupDataChain(serverAddr nodeStore.AddressNet, proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	if bytes.Equal(serverAddr.GetAddr(), Node.GetNetId().GetAddr()) {
		//_, _, ERR := GroupAuthMember(config.DBKEY_improxy_user_group_list_knit, proxyItr.GetBase().GroupID, proxyItr.GetAddrFrom())
		//if !ERR.CheckSuccess() {
		//	utils.Log.Info().Msgf("群权限验证 失败:%s", ERR.String())
		//	return ERR
		//}
		ERR := StaticProxyServerManager.groupKnitManager.KnitDataChain(proxyItr)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("群权限验证 失败:%s", ERR.String())
			return ERR
		}
		return utils.NewErrorSuccess()
	}

	bs, err := proxyItr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	utils.Log.Info().Str("发送群消息", serverAddr.B58String()).Int("包大小", len(*bs)).Send()
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_group_send_datachain, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("发送群消息 错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
发送群消息给代理节点
*/
func MulticastGroupDataChain(serverAddr nodeStore.AddressNet, proxyItrs []imdatachain.DataChainProxyItr) utils.ERROR {
	utils.Log.Info().Msgf("发送群聊消息:%s", serverAddr.B58String())
	var ERR utils.ERROR
	if bytes.Equal(serverAddr.GetAddr(), Node.GetNetId().GetAddr()) {
		//_, _, ERR := GroupAuthMember(config.DBKEY_improxy_group_list_knit, proxyItrs[0].GetBase().GroupID, proxyItrs[0].GetAddrFrom())
		//if !ERR.CheckSuccess() {
		//	return ERR
		//}
		for _, itrOne := range proxyItrs {
			ERR = StaticProxyServerManager.gsmm.ParseGroupDataChain(itrOne)
			ERR = StaticProxyClientManager.groupParserManager.SaveGroupDataChain(itrOne)
		}
		return utils.NewErrorSuccess()
	}

	byteList := make([][]byte, 0, len(proxyItrs))
	for _, one := range proxyItrs {
		bs, err := one.Proto()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		byteList = append(byteList, *bs)
	}
	bs, err := model.BytesProto(byteList, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, bs)
	nbs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("发送群聊消息:%s", serverAddr.B58String())
	_, ERR = Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_group_multicast_datachain, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("广播群消息 错误:%s", ERR.String())
		return ERR
	}
	//返回成功
	ERR = utils.NewErrorSuccess()
	return ERR
}

/*
下载群数据链消息
*/
func DownloadGroupDataChain(addr nodeStore.AddressNet, groupId, index []byte) ([]imdatachain.DataChainProxyItr, utils.ERROR) {
	params := [][]byte{groupId, index}
	bs, err := model.BytesProto(params, nil)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, bs)
	nbs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_group_download_datachain, &addr, nbs, time.Minute)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("下载数据链 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	list, _, err := model.ParseBytes(nr.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(list))
	for _, one := range list {
		itr, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		itrs = append(itrs, itr)
	}
	return itrs, utils.NewErrorSuccess()
}

/*
下载群数据链消息
*/
func GetGroupDataChainStartIndex(addr nodeStore.AddressNet, groupId []byte) (imdatachain.DataChainProxyItr, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, groupId)
	nbs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_group_download_datachain_start, &addr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("查询数据链起始位置 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	proxyItr, ERR := imdatachain.ParseDataChain(nr.Data)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return proxyItr, utils.NewErrorSuccess()
}

/*
从对方节点获取自己的sendindex
*/
func GetSendIndexForMyself(addr nodeStore.AddressNet) ([]byte, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	nbs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_IM_PROXY_get_sendindex, &addr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("查询数据链起始位置 错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	return nr.Data, utils.NewErrorSuccess()
}
