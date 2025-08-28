package im

import (
	"bytes"
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"slices"
	"sync"
	"time"
	chainConfig "web3_gui/chain/config"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/crypto"
	"web3_gui/utils/utilsleveldb"
)

/*
创建一个群
@nickname    string    群昵称
@shoutUp     bool      是否闭言
*/
func CreateGroup(nickname, proxyAddr string, shoutUp bool) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")
	var addr nodeStore.AddressNet
	if proxyAddr == "" {
		addr = *Node.GetNetId()
	} else {
		addr = nodeStore.AddressFromB58String(proxyAddr)
	}

	//获取签名用私钥
	puk, prk, ERR := Node.Keystore.GetNetAddrKeyPair(chainConfig.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}

	//获取DH密钥对
	dhPuk := dhKey.GetPublicKey()

	//生成一个随机数，作为群聊天加解密密码
	randKey, err := crypto.Rand32Byte()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//使用群成员公钥，和自己的私钥协程密钥加密群通信密码

	var memberPuk [32]byte
	copy(memberPuk[:], dhPuk[:])
	dhPrk := dhKey.PrivateKey
	//utils.Log.Info().Msgf("打印公钥:%+v 私钥:%+v", one.GroupDHPuk, prk)
	//生成共享密钥sharekey
	sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, memberPuk))
	if err != nil {
		utils.Log.Info().Msgf("生成共享密钥 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	utils.Log.Info().Msgf("加密共享密钥参数:%+v %+v", sharekey, randKey)
	cipherText, err := utils.AesCTR_Encrypt(sharekey[:], nil, randKey[:])
	if err != nil {
		utils.Log.Info().Msgf("用协商密钥加密 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//one.GroupShareKey = cipherText

	createGroup := imdatachain.NewDataChainCreateGroup(*Node.GetNetId(), addr, nickname, shoutUp, cipherText, puk, dhPuk[:])

	signText := append(createGroup.GroupID, utils.Int64ToBytesByBigEndian(createGroup.CreateTime)...)
	//签名
	sign := ed25519.Sign(prk, signText)
	createGroup.MembersSign = append(createGroup.MembersSign, sign)

	//utils.Log.Info().Msgf("发送文本消息")
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(createGroup)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
获取自己创建的群列表
*/
func GetCreateGroupList() (*[]imdatachain.DataChainCreateGroup, utils.ERROR) {
	//utils.Log.Info().Msgf("发送消息")
	groupList, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("获取群列表 错误:%s", ERR.String())
		return nil, ERR
	}
	return groupList, ERR
}

/*
获取一个群的基本信息
*/
func GetGroupInfo(groupId []byte) (*imdatachain.DataChainCreateGroup, utils.ERROR) {
	//查询群是否存在
	groupInfo, ERR := db.FindGroupInfoAllList(*Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return nil, ERR
	}
	if groupInfo == nil {
		return nil, utils.NewErrorSuccess()
	}
	return groupInfo, ERR
}

/*
获取群信息和成员
@return    *model.UserInfoVO        群信息
@return    *model.UserinfoListVO    群成员列表
@return    utils.ERROR              错误
*/
func GetGroupMembers(groupId []byte) (*model.UserInfoVO, *model.UserinfoListVO, utils.ERROR) {
	//utils.Log.Info().Msgf("发送消息")
	members, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("获取群列表 错误:%s", ERR.String())
		return nil, nil, ERR
	}
	createGroup, ERR := GetGroupInfo(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("获取群列表 错误:%s", ERR.String())
		return nil, nil, ERR
	}
	userinfoList := model.NewUserList()
	for _, one := range *members {
		if bytes.Equal(one.Addr.GetAddr(), Node.GetNetId().GetAddr()) {
			user, ERR := db.GetSelfInfo(*Node.GetNetId())
			if !ERR.CheckSuccess() {
				return nil, nil, ERR
			}
			one.Nickname = user.Nickname
			one.HeadNum = user.HeadNum
		}

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
	groupInfo := model.UserInfo{
		Addr:      *nodeStore.NewAddressNet(createGroup.GroupID),
		AddrAdmin: createGroup.AddrFrom,
		Nickname:  createGroup.Nickname,
		//RemarksName: "",
		HeadNum: 0,
		Status:  0,
		Time:    0,
		IsGroup: true,
		GroupId: createGroup.GroupID,
		Proxy:   new(sync.Map),
		ShoutUp: createGroup.ShoutUp,
	}
	groupInfo.Proxy.Store(utils.Bytes2string(createGroup.ProxyMajor.GetAddr()), createGroup.ProxyMajor)
	groupInfoVO := model.ConverUserInfoVO(&groupInfo)
	return groupInfoVO, model.ConverUserListVO(userinfoList), ERR
}

/*
修改一个群
@nickname    string    群昵称
@shoutUp     bool      是否闭言
@force       bool      是否强制托管
*/
func UpdateGroup(groupId []byte, proxyAddr string, nickname string, shoutUp, force bool) utils.ERROR {
	var proxyMajor nodeStore.AddressNet
	if proxyAddr == "" {
		proxyMajor = *Node.GetNetId()
	} else {
		proxyMajor = nodeStore.AddressFromB58String(proxyAddr)
	}
	groupinfo, userinfo, ERR := GroupAuthMember(*config.DBKEY_improxy_group_list_create, groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	if !userinfo.Admin {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("查询群sendIndex:%+v", sendIndex)
	updateGroup := imdatachain.NewDataChainUpdateGroup(*Node.GetNetId(), proxyMajor, nickname, groupinfo.CreateTime, shoutUp, groupId)
	updateGroup.SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	//代理节点拒绝服务的情况下，强制托管不用给代理节点发送消息
	if !force {
		ERR = SendGroupDataChain(groupinfo.ProxyMajor, updateGroup)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
			return ERR
		}
	}

	//当代理节点拒绝服务的时候，把代理节点设置为自己，则强制自己作为群构建者。
	//但是要避免代理节点在工作的时候，强制设置为自己会掉消息。那么建议先设置闭言再改构建者。
	if bytes.Equal(proxyMajor.GetAddr(), Node.GetNetId().GetAddr()) {
		StaticProxyClientManager.TrusteeshipKnitGroup(updateGroup)
	}
	return utils.NewErrorSuccess()
}

/*
解散一个群
@nickname    string    群昵称
@shoutUp     bool      是否闭言
*/
func DissolveGroup(groupId []byte) utils.ERROR {
	utils.Log.Info().Msgf("解散群聊")
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
	//是不是管理员
	if !userInfo.Admin {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//自己创建的群，就可以解散
	dissolveGroup := imdatachain.NewProxyGroupDissolve(*Node.GetNetId(), groupId)

	//获取sendIndex
	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送群消息 错误:%s", ERR.String())
		return ERR
	}
	//proxyItr := quitGroup.GetProxyItr()
	dissolveGroup.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	//quitGroup.EncryptContent(shareKey)

	//utils.Log.Info().Msgf("构造的群消息结构:%d %+v", quitGroup.GetCmd(), quitGroup)
	//发送给群
	ERR = SendGroupDataChain(groupInfo.ProxyMajor, dissolveGroup)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
邀请好友加入群
@nickname    string    群昵称
@shoutUp     bool      是否闭言
*/
func GroupInvitationFriend(groupId []byte, addr nodeStore.AddressNet) utils.ERROR {
	//查询群是否存在
	groupInfo, ERR := db.FindGroupInfoAllList(*Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	userSelf, ERR := db.GetSelfInfo(*Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}

	//查询群成员
	users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	var addrAdmin nodeStore.AddressNet
	for _, one := range *users {
		utils.Log.Info().Msgf("查询出来的群成员:%+v", one)
		if one.Admin {
			addrAdmin = one.Addr
			break
		}
	}

	//utils.Log.Info().Msgf("发送消息")
	createGroup := imdatachain.NewDatachainGroupInvitation(*Node.GetNetId(), addr, addrAdmin, groupId,
		userSelf.Nickname, groupInfo.Nickname)
	utils.Log.Info().Msgf("邀请好友入群:%+v", createGroup)
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(createGroup.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
好友同意加入群聊
@groupId            []byte                    群ID
@addrGroupAdmin     nodeStore.AddressNet      群管理员地址
*/
func GroupAccept(token []byte) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")

	userSelf, ERR := db.GetSelfInfo(*Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}

	//通过token查询群id和管理员地址
	proxyItr, ERR := db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), *Node.GetNetId(), token)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = DecryptContent(proxyItr)
	if ERR.CheckFail() {
		return ERR
	}
	groupInvitation := proxyItr.GetClientItr().(*imdatachain.DatachainGroupInvitation)
	groupId := groupInvitation.GroupID
	addrGroupAdmin := groupInvitation.AdminAddr

	//获取签名用私钥
	puk, prk, ERR := Node.Keystore.GetNetAddrKeyPair(chainConfig.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//
	acceptTime := time.Now().Unix()
	signText := append(groupId, utils.Int64ToBytesByBigEndian(acceptTime)...)
	//签名
	sign := ed25519.Sign(prk, signText)
	//获取DH密钥对
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//获取DH密钥对
	dhPuk := dhKey.GetPublicKey()
	//dhPuk := Node.Keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	acceptGroup := imdatachain.NewDatachainGroupAccept(*Node.GetNetId(), addrGroupAdmin, groupId, acceptTime, sign, puk,
		dhPuk[:], token, userSelf.Nickname, groupInvitation.GroupNickname)
	//utils.Log.Info().Msgf("好友同意加入群聊")
	utils.Log.Info().Msgf("好友同意加入群聊:%+v", acceptGroup)
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(acceptGroup.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
同意群添加成员，只有群管理员可以添加
@groupId      []byte                  群ID
@addrNew      nodeStore.AddressNet    添加的新成员地址
*/
func AgreeGroupAddMember(token []byte) utils.ERROR {
	utils.Log.Info().Msgf("管理员添加成员")

	//通过token查询群id和新加入成员地址
	proxyItr, ERR := db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), *Node.GetNetId(), token)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = DecryptContent(proxyItr)
	if ERR.CheckFail() {
		return ERR
	}
	groupAccept := proxyItr.GetClientItr().(*imdatachain.DatachainGroupAccept)
	groupId := groupAccept.GroupID
	addrNew := groupAccept.AddrFrom

	//查询群是否存在
	groupInfo, ERR := db.ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//查询是否是本群管理员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if userInfo == nil || !userInfo.Admin {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//获取新成员的公钥、签名等信息
	newMember, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index, *Node.GetNetId(), token)
	//newMember, ERR := db.FindUserListByGroup(config.DBKEY_apply_remote_userlist, Node.GetNetId(), addrNew, groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("查询新成员签名信息 错误:%s", ERR.String())
		return ERR
	}

	//查询是否在列表中
	//userInfo, ERR = db.FindUserListByAddr(config.DBKEY_apply_remote_userlist, addrNew)
	//if !ERR.CheckSuccess() {
	//	utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
	//	return ERR
	//}
	//查询群成员
	users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	members := *users
	members = append(members, *newMember)

	//utils.Log.Info().Msgf("------群成员:%+v", members)

	//获取自己的私钥
	//dhKey := Node.Keystore.GetDHKeyPair()
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//获取DH密钥对
	//dhPuk := dhKey.GetPublicKey()

	//utils.Log.Info().Msgf("打印自己的公钥:%s 私钥:%s", hex.EncodeToString(dhKey.KeyPair.PublicKey[:]),
	//hex.EncodeToString(dhKey.KeyPair.PrivateKey[:]))

	//生成一个随机数，作为群聊天加解密密码
	randKey, err := crypto.Rand32Byte()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//utils.Log.Info().Msgf("群随机数密码原文:%s", hex.EncodeToString(randKey[:]))

	//使用群成员公钥，和自己的私钥协程密钥加密群通信密码
	for i, one := range members {
		var memberPuk [32]byte
		copy(memberPuk[:], one.GroupDHPuk)
		prk := dhKey.PrivateKey
		//utils.Log.Info().Msgf("打印公钥:%+v 私钥:%+v", one.GroupDHPuk, prk)
		//生成共享密钥sharekey
		sharekey, err := keystore.KeyExchange(keystore.NewDHPair(prk, memberPuk))
		if err != nil {
			utils.Log.Info().Msgf("生成共享密钥 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}

		cipherText, err := utils.AesCTR_Encrypt(sharekey[:], nil, randKey[:])
		if err != nil {
			utils.Log.Info().Msgf("用协商密钥加密 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		members[i].GroupShareKey = cipherText
		//one.GroupShareKey = cipherText

		//utils.Log.Info().Msgf("管理员与:%s 的协商密钥是:%s 加密后的群密码:%s", one.Addr.B58String(),
		//hex.EncodeToString(sharekey[:]), hex.EncodeToString(cipherText))
	}
	//utils.Log.Info().Msgf("群成员共享密钥:%+v", members)
	groupMember := imdatachain.NewDatachainGroupMember(*Node.GetNetId(), groupInfo.ProxyMajor, groupInfo.Nickname,
		groupInfo.CreateTime, groupInfo.ShoutUp, groupId, members)
	//utils.Log.Info().Msgf("群成员变动:%+v", groupMember)
	//保存并解析数据链记录

	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	groupMember.SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))

	//utils.Log.Info().Msgf("群成员变动:%d %+v", groupMember.GetCmd(), groupMember)

	for _, one := range members {
		utils.Log.Info().Msgf("最后打印一遍加密后的shareKey:%s", hex.EncodeToString(one.GroupShareKey))
	}
	//发送给群
	ERR = SendGroupDataChain(groupInfo.ProxyMajor, groupMember)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}

	clientAddMember := imdatachain.NewDatachainGroupAddMember(*Node.GetNetId(), addrNew, groupInfo.GroupID, token)

	utils.Log.Info().Msgf("同意添加成员:%+v", clientAddMember)
	//再传给自己的链
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(clientAddMember.GetProxyItr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("同意添加用户入群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
同意群添加成员，只有群管理员可以添加
@groupId      []byte                  群ID
@addrNew      nodeStore.AddressNet    添加的新成员地址
*/
func GroupDelMember(groupId []byte, addrs []nodeStore.AddressNet) utils.ERROR {
	utils.Log.Info().Msgf("管理员删除成员")

	//查询群是否存在
	groupInfo, ERR := db.ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//查询是否是本群管理员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if userInfo == nil || !userInfo.Admin {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//管理员不能删除自己
	for _, one := range addrs {
		if bytes.Equal(Node.GetNetId().GetAddr(), one.GetAddr()) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_del_admin, "")
		}
	}

	//查询旧有群成员
	usersOld, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	//删除群成员
	membersNew := make([]model.UserInfo, 0, len(*usersOld)-len(addrs))
	for _, one := range *usersOld {
		for _, addrNew := range addrs {
			//匹配上的用户删除
			if bytes.Equal(one.Addr.GetAddr(), addrNew.GetAddr()) {
				continue
			}
			membersNew = append(membersNew, one)
		}
	}

	//utils.Log.Info().Msgf("------群成员:%+v", members)

	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//获取DH密钥对
	//dhPuk := dhKey.GetPublicKey()
	//获取自己的私钥
	//dhKey := Node.Keystore.GetDHKeyPair()

	//utils.Log.Info().Msgf("打印自己的公钥:%s 私钥:%s", hex.EncodeToString(dhKey.KeyPair.PublicKey[:]), hex.EncodeToString(dhKey.KeyPair.PrivateKey[:]))

	//生成一个随机数，作为群聊天加解密密码
	randKey, err := crypto.Rand32Byte()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//utils.Log.Info().Msgf("群随机数密码原文:%s", hex.EncodeToString(randKey[:]))

	//使用群成员公钥，和自己的私钥协程密钥加密群通信密码
	for i, one := range membersNew {
		var memberPuk [32]byte
		copy(memberPuk[:], one.GroupDHPuk)
		prk := dhKey.PrivateKey
		//utils.Log.Info().Msgf("打印公钥:%+v 私钥:%+v", one.GroupDHPuk, prk)
		//生成共享密钥sharekey
		sharekey, err := keystore.KeyExchange(keystore.NewDHPair(prk, memberPuk))
		if err != nil {
			utils.Log.Info().Msgf("生成共享密钥 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}

		cipherText, err := utils.AesCTR_Encrypt(sharekey[:], nil, randKey[:])
		if err != nil {
			utils.Log.Info().Msgf("用协商密钥加密 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		membersNew[i].GroupShareKey = cipherText
		//one.GroupShareKey = cipherText

		//utils.Log.Info().Msgf("管理员与:%s 的协商密钥是:%s 加密后的群密码:%s", one.Addr.B58String(),
		//hex.EncodeToString(sharekey[:]), hex.EncodeToString(cipherText))
	}
	//utils.Log.Info().Msgf("群成员共享密钥:%+v", members)
	groupMember := imdatachain.NewDatachainGroupMember(*Node.GetNetId(), groupInfo.ProxyMajor, groupInfo.Nickname,
		groupInfo.CreateTime, groupInfo.ShoutUp, groupId, membersNew)
	//utils.Log.Info().Msgf("群成员变动:%+v", groupMember)
	//保存并解析数据链记录

	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	groupMember.SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))

	//utils.Log.Info().Msgf("群成员变动:%d %+v", groupMember.GetCmd(), groupMember)

	for _, one := range membersNew {
		utils.Log.Info().Msgf("最后打印一遍加密后的shareKey:%s", hex.EncodeToString(one.GroupShareKey))
	}
	//发送给群
	ERR = SendGroupDataChain(groupInfo.ProxyMajor, groupMember)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}

	return ERR
}

/*
群删除成员，只有群管理员可以移除
@groupId      []byte                    群ID
@addrAdmin    []nodeStore.AddressNet    删除的成员地址
*/
func GroupRemoveMember(groupId []byte, addrs []nodeStore.AddressNet) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")
	//查询群是否存在
	groupInfo, ERR := db.ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if groupInfo == nil {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}
	//查询是否是本群管理员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	if userInfo == nil || !userInfo.Admin {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
	}

	//查询群成员
	users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	members := *users

	//删除成员
	members = slices.DeleteFunc(members, func(userOne model.UserInfo) bool {
		for _, one := range addrs {
			if bytes.Equal(one.GetAddr(), userOne.Addr.GetAddr()) {
				return true
			}
		}
		return false
	})
	//获取自己的私钥
	dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//获取DH密钥对
	//dhPuk := dhKey.GetPublicKey()
	//获取自己的私钥
	//dhKey := Node.Keystore.GetDHKeyPair()
	//生成一个随机数，作为群聊天加解密密码
	randKey, err := crypto.Rand32Byte()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//使用群成员公钥，和自己的私钥协程密钥加密群通信密码
	for _, one := range members {
		var memberPuk [32]byte
		copy(memberPuk[:], one.GroupSignPuk)
		//生成共享密钥sharekey
		sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhKey.PrivateKey, memberPuk))
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		cipherText, err := utils.AesCTR_Encrypt(sharekey[:], nil, randKey[:])
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		one.GroupShareKey = cipherText
	}
	createGroup := imdatachain.NewDatachainGroupMember(*Node.GetNetId(), groupInfo.ProxyMajor, groupInfo.Nickname,
		groupInfo.CreateTime, groupInfo.ShoutUp, groupId, members)
	//utils.Log.Info().Msgf("发送文本消息")
	//保存并解析数据链记录
	ERR = StaticProxyClientManager.ParserClient.SaveDataChain(createGroup)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
成员退出群聊
@groupId      []byte                  群ID
*/
func GroupMemberQuit(groupId []byte) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")

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
	quitGroup := imdatachain.NewDataChainProxyGroupMemberQuit(*Node.GetNetId(), groupId)

	//
	//groupText := imdatachain.NewDataChainGroupSendText(Node.GetNetId(), groupId, []byte(content), quotoIDBS)

	//获取sendIndex
	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送群消息 错误:%s", ERR.String())
		return ERR
	}
	//proxyItr := quitGroup.GetProxyItr()
	quitGroup.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	//quitGroup.EncryptContent(shareKey)

	//utils.Log.Info().Msgf("构造的群消息结构:%d %+v", quitGroup.GetCmd(), quitGroup)
	//发送给群
	ERR = SendGroupDataChain(groupInfo.ProxyMajor, quitGroup)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

/*
发送群消息
@groupId      []byte                  群ID
@addr         nodeStore.AddressNet    退出的成员地址
*/
func SendGroupText(groupId []byte, content, quoteID string) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")
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

	//获取群共享密钥
	shareKey, ERR := StaticProxyClientManager.groupParserManager.GetShareKey(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
		return ERR
	}

	//
	groupText := imdatachain.NewDataChainGroupSendText(*Node.GetNetId(), groupId, []byte(content), quotoIDBS)

	//获取sendIndex
	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送群消息 错误:%s", ERR.String())
		return ERR
	}
	proxyItr := groupText.GetProxyItr()
	proxyItr.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	ERR = proxyItr.EncryptContent(shareKey)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}

	utils.Log.Info().Hex("发送一条群消息", groupId).Send()

	groupParser, ERR := StaticProxyClientManager.groupParserManager.GetGroupParse(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Hex("发送一条群消息", groupId).Send()
	//发送给群
	ERR = groupParser.SendText(proxyItr)
	utils.Log.Info().Hex("发送一条群消息", groupId).Send()

	//发送给群
	//ERR = SendGroupDataChain(groupInfo.ProxyMajor, proxyItr)
	//if !ERR.CheckSuccess() {
	//	utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
	//	return ERR
	//}
	return ERR
}

/*
群权限验证
@groupId    []byte                  群ID
@addr       nodeStore.AddressNet    地址
*/
func GroupAuthMember(key utilsleveldb.LeveldbKey, groupId []byte, addr nodeStore.AddressNet) (*imdatachain.DataChainCreateGroup,
	*model.UserInfo, utils.ERROR) {
	//检查群是否存在
	groupInfo, ERR := db.ImProxyClient_FindGroupInfo(key, *Node.GetNetId(), groupId)
	if !ERR.CheckSuccess() {
		return nil, nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
	}
	if groupInfo == nil {
		utils.Log.Info().Msgf("群列表中未找到")
		//检查群是否已经解散
		ok, ERR := db.ImProxyClient_FindDissolveGroupExist(*Node.GetNetId(), groupId)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		if ok {
			//群已经解散
			return nil, nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_dissolve, "")
		}
		utils.Log.Info().Msgf("群不存在")
		//群不存在
		return nil, nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
	}
	//查询是否本群管理员
	userInfo, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupId, addr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return nil, nil, ERR
	}
	if userInfo == nil {
		//不是群成员
		return nil, nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}
	return groupInfo, userInfo, utils.NewErrorSuccess()
}

/*
发送群文件
@groupId      []byte                  群ID
@addr         nodeStore.AddressNet    退出的成员地址
*/
func SendGroupFile(groupId []byte, content, quoteID string) utils.ERROR {
	//utils.Log.Info().Msgf("发送消息")
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

	//获取群共享密钥
	shareKey, ERR := StaticProxyClientManager.groupParserManager.GetShareKey(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
		return ERR
	}

	//
	groupText := imdatachain.NewDataChainGroupSendText(*Node.GetNetId(), groupId, []byte(content), quotoIDBS)

	//获取sendIndex
	sendIndex, ERR := StaticProxyClientManager.groupParserManager.GetSendIndex(groupId)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("发送群消息 错误:%s", ERR.String())
		return ERR
	}
	proxyItr := groupText.GetProxyItr()
	proxyItr.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	ERR = proxyItr.EncryptContent(shareKey)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}

	utils.Log.Info().Msgf("构造的群消息结构:%d %+v", proxyItr.GetCmd(), proxyItr)
	//发送给群
	ERR = SendGroupDataChain(groupInfo.ProxyMajor, proxyItr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
		return ERR
	}
	return ERR
}

var searchUserInfoIntervalLock = new(sync.RWMutex)
var searchUserInfoInterval = make(map[string]int64)

/*
间隔1小时查询一次用户的代理节点
*/
func SearchUserProxyInterval(addr nodeStore.AddressNet) {
	searchUserInfoIntervalLock.Lock()
	defer searchUserInfoIntervalLock.Unlock()
	unix := time.Now().Unix()
	a := addr.GetAddr()
	key := utils.Bytes2string(a.Bytes())
	before, ok := searchUserInfoInterval[key]
	if ok {
		//有记录，则对比上次与当前时间间隔
		if unix-before < 60*60 {
			return
		}
	}
	searchUserInfoInterval[key] = unix
	go SearchUserProxy(addr)
}
