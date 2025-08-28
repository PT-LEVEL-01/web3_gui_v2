package server_api

import (
	"cmp"
	"encoding/hex"
	"slices"
	"web3_gui/config"
	"web3_gui/im/im"
	"web3_gui/im/im/imdatachain"
	"web3_gui/keystore/v2/base58"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
创建一个群
@nickname     string    群名称
@proxyAddr    string    代理节点地址，没有代理传空字符串
@shoutUp      bool      是否禁言
*/
func (a *SdkApi) ImProxyClient_CreateGroup(nickname, proxyAddr string, shoutUp bool) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	ERR := im.CreateGroup(nickname, proxyAddr, shoutUp)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
获取自己创建的群列表
*/
func (a *SdkApi) ImProxyClient_GetCreateGroupList() map[string]interface{} {
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	groups, ERR := im.GetCreateGroupList()
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	groupsVO := make([]imdatachain.DataChainCreateGroupVO, 0, len(*groups))
	for _, one := range *groups {
		vo := one.ConverVO()
		groupsVO = append(groupsVO, *vo)
	}
	slices.SortFunc(groupsVO, func(a, b imdatachain.DataChainCreateGroupVO) int {
		return cmp.Compare(a.CreateTime, b.CreateTime)
	})
	//ERR := utils.NewErrorSuccess()
	resultMap["list"] = groupsVO
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
获取群成员
@groupId    string    群id
*/
func (a *SdkApi) ImProxyClient_GetGroupMembers(groupId string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("获取群成员:%s", groupId)
	groupBs := base58.Decode(groupId)
	resultMap := make(map[string]interface{})
	groupInfo, userlist, ERR := im.GetGroupMembers(groupBs)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("获取群成员列表失败:%s", ERR.String())
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	//return userlist
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = userlist
	resultMap["groupinfo"] = groupInfo
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
修改一个群
@groupId      string    群id
@proxyAddr    string    代理节点地址，没有代理传空字符串
@nickname     string    群名称
@shoutUp      bool      是否禁言
@force        bool      是否强制托管，当群代理节点不工作的时候，强制自己托管
*/
func (a *SdkApi) ImProxyClient_UpdateGroup(groupId, proxyAddr, nickname string, shoutUp, force bool) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	groupBs := base58.Decode(groupId)
	//addr := nodeStore.AddressFromB58String(proxyAddr)
	ERR := im.UpdateGroup(groupBs, proxyAddr, nickname, shoutUp, force)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
群邀请新人
@groupId      string        群id
@addrStr      []string      邀请的成员地址
*/
func (a *SdkApi) ImProxyClient_GroupInvitation(groupId string, addrStr []string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	groupBs := base58.Decode(groupId)
	for _, one := range addrStr {
		addr := nodeStore.AddressFromB58String(one)
		ERR = im.GroupInvitationFriend(groupBs, addr)
		if !ERR.CheckSuccess() {
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
接受群邀请
@groupId      string      群id
@addrStr      string      群管理员地址
*/
func (a *SdkApi) ImProxyClient_GroupAccept(tokenStr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("接受群邀请")
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if tokenStr == "" {
		//return errors.New("addr is \"\"")
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "tokenStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	token, err := hex.DecodeString(tokenStr)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "tokenStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = im.GroupAccept(token)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
申请加入群聊
@nickname    string    群名称
@shoutUp     bool      是否禁言
*/
func (a *SdkApi) ImProxyClient_GroupApply(groupId string, addrStr string) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})

	//orders := im.StaticProxyClientManager.GetOrdersList()
	////orders := storage.StClient.GetOrdersList()
	//if err != nil {
	//	ERR = utils.NewErrorSysSelf(err)
	//	resultMap["code"] = ERR.Code
	//	resultMap["error"] = ERR.Msg
	//	return resultMap
	//}
	//proxyVOs := make([]*model.ImProxyVO, 0, len(orders))
	//for _, one := range orders {
	//	proxyVOs = append(proxyVOs, one.ConverVO())
	//}
	////ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
同意添加人入群聊
@groupIdStr    string      群id
@addrStr       string      添加的新成员地址
*/
func (a *SdkApi) ImProxyClient_GroupAddMember(tokenStr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if tokenStr == "" {
		//return errors.New("addr is \"\"")
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "tokenStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	token, err := hex.DecodeString(tokenStr)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "tokenStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = im.AgreeGroupAddMember(token)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR = utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
踢人出群
@groupIdStr    string      群id
@addrStr       string      添加的新成员地址
*/
func (a *SdkApi) ImProxyClient_GroupDelMember(groupId string, addrStr []string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if groupId == "" {
		//return errors.New("addr is \"\"")
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "groupId")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	groupIdBs := base58.Decode(groupId)
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range addrStr {
		addr := nodeStore.AddressFromB58String(one)
		addrs = append(addrs, addr)
	}
	ERR = im.GroupDelMember(groupIdBs, addrs)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR = utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
解散一个群
@groupId    string      群id
*/
func (a *SdkApi) ImProxyClient_DissolveGroup(groupId string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	groupBs := base58.Decode(groupId)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	ERR := im.DissolveGroup(groupBs)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
退出一个群
@groupId    string      群id
*/
func (a *SdkApi) ImProxyClient_QuitGroup(groupId string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	groupBs := base58.Decode(groupId)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	ERR := im.GroupMemberQuit(groupBs)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
发送群消息
@groupId    string      群id
@content    string      消息内容
@quoteID    string      引用内容id
*/
func (a *SdkApi) ImProxyClient_GroupSendText(groupId string, content, quoteID string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	groupBs := base58.Decode(groupId)
	ERR = im.SendGroupText(groupBs, content, quoteID)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//ERR := utils.NewErrorSuccess()
	//resultMap["list"] = proxyVOs
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
群发送文件
@groupIdStr     string    群地址
@filePath       string    文件路径
*/
func (a *SdkApi) ImProxyClient_GroupSendFiles(groupIdStr, filePath string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("发送文件:%s", filePath)
	resultMap := make(map[string]interface{})
	if groupIdStr == "" || filePath == "" {
		//return errors.New("参数不能为 \"\"")
		msg := "groupIdStr"
		if filePath == "" {
			msg = "filePath"
		}
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(groupIdStr)
	//_, fileName := filepath.Split(filePath)
	//先发送消息
	ERR := im.GroupSendFile(addr.GetAddr(), filePath)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["list"] = list
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
群发送图片
@groupIdStr     string    群地址
@imgBase64      string    图片base64编码
*/
func (a *SdkApi) ImProxyClient_GroupSendImage(groupIdStr, imgBase64 string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	resultMap := make(map[string]interface{})
	if imgBase64 == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "imgBase64")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if groupIdStr == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "groupIdStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(groupIdStr)
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	ERR := im.GroupSendImageBase64(addr.GetAddr(), imgBase64)
	//ERR := utils.NewErrorSysSelf(err)
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap

	//给前端发送一个通知
	//msgInfo := im.NewMessageInfo(config.SUBSCRIPTION_type_msg, config.MSG_type_image, true, config.NetAddr, addrStr,
	//	imgBase64, utils.FormatTimeToSecond(time.Now()), 0, nil, 0)
	//im.AddSubscriptionMsg(msgInfo)
	//return nil
}
