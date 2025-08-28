package server_api

import (
	"encoding/hex"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/gui/tray"
	"web3_gui/im/db"
	"web3_gui/im/im"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/libp2parea/v2/cake/transfer_manager"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
检查IM模块信息
*/
func (a *SdkApi) IM_PrintLog(itr interface{}) {
	utils.Log.Info().Msgf("%+v", itr)
}

/*
检查IM模块信息
*/
func (a *SdkApi) IM_GetInfo() bool {
	// utils.Log.Info().Msgf("%s", config.NetAddr)
	return true
}

var isStart = false
var isStartLock = new(sync.Mutex)

/*
启动IM模块
*/
func (a *SdkApi) IM_StartIm(passwd string, init bool) map[string]interface{} {
	isStartLock.Lock()
	defer isStartLock.Unlock()
	resultMap := make(map[string]interface{})
	//utils.Log.Info().Msgf("111111111111111")
	if isStart {
		ERR := utils.NewErrorSuccess()
		//resultMap["info"] = model.ConverUserInfoVO(userInfo)
		resultMap["code"] = ERR.Code
		return resultMap
	}
	InitConfig()

	utils.Log.Info().Msgf("key文件路径:%s", config.KeystoreFileAbsPath)
	ERR := im.StartUP(passwd, init, a.Ctx)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("启动报错:%s", ERR.String())
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		return resultMap
	}
	//utils.Log.Info().Msgf("3333333333333333")
	isStart = true
	ERR = utils.NewErrorSuccess()
	//resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	return resultMap

}

/*
等待IM模块完成
*/
func (a *SdkApi) IM_StartIm_wait(passwd string, init bool) string {
	im.Node.WaitAutonomyFinish()
	return config.NetAddr
}

/*
获取自己的信息
*/
func (a *SdkApi) IM_GetSelfInfo() map[string]interface{} {
	resultMap := make(map[string]interface{})
	userInfo, ERR := db.GetSelfInfo(*im.Node.GetNetId())
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return model.ConverUserInfoVO(userInfo)
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
设置自己的信息
@nickname      string    //昵称
@headNum       uint64    //头像编号，外部预设编号及含义。
@traySwitch    bool      //PC端是否打开系统托盘
*/
func (a *SdkApi) IM_SetSelfInfo(nickname string, headNum uint64, traySwitch bool) map[string]interface{} {
	resultMap := make(map[string]interface{})
	userInfo, ERR := db.GetSelfInfo(*im.Node.GetNetId())
	if !ERR.CheckSuccess() {
		//return err
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if nickname != "" {
		userInfo.Nickname = nickname
	}
	if headNum != 0 {
		userInfo.HeadNum = headNum
	}
	userInfo.Tray = traySwitch
	if traySwitch {
		tray.OpenSystemTray()
	} else {
		tray.CloseSystemTray()
	}
	//utils.Log.Info().Msgf("修改个人信息:%+v", userInfo)
	//return db.UpdateSelfInfo(userInfo)
	ERR = db.UpdateSelfInfo(*im.Node.GetNetId(), userInfo)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("修改个人信息错误:%s", ERR.String())
		//return err
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	subscription.CacheSetUserInfo(userInfo)
	//im.SetUserSelf_cache(userInfo)
	//utils.Log.Info().Msgf("修改个人信息成功")
	//newUserInfo, _ := db.GetSelfInfo()
	//utils.Log.Info().Msgf("查询个人信息:%+v", newUserInfo)
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
发送文本消息
*/
func (a *SdkApi) IM_SendMsg_old(content, toAddr, quoteID string) map[string]interface{} {
	//utils.Log.Info().Msgf("发送消息:%s", content)
	resultMap := make(map[string]interface{})
	if content == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "content")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
		//return nil
	}
	if toAddr != "" {
		addr := nodeStore.AddressFromB58String(toAddr)
		//utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
		ERR := im.SendFriendMsgAPI_old(addr, content, config.MSG_type_text, 0, quoteID)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Msgf("发送广播消息")
	//没有目标节点，则是广播消息
	//广播消息
	im.SendMulticastMessage(content)
	//return nil
	ERR := utils.NewErrorSuccess()
	//resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
发送文本消息
@content    string    发送内容
@toAddr     string    好友地址
@quoteID    string    引用内容id
*/
func (a *SdkApi) IM_SendMsg(content, toAddr, quoteID string) map[string]interface{} {
	utils.Log.Info().Msgf("发送文本消息长度:%d", len(content))
	resultMap := make(map[string]interface{})
	if content == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "content")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
		//return nil
	}
	if toAddr != "" {
		addr := nodeStore.AddressFromB58String(toAddr)
		//utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
		ERR := im.SendTextAPI(addr, content, quoteID)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Msgf("发送广播消息")
	//没有目标节点，则是广播消息
	//广播消息
	im.SendMulticastMessage(content)
	//return nil
	ERR := utils.NewErrorSuccess()
	//resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
重发文本消息
@fromAddr    string    发送者
@toAddr      string    接收者
@sendID      string    消息发送id
*/
func (a *SdkApi) IM_SendMsgAgain(fromAddr, toAddr, sendID string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Msgf("发送消息:%s", content)
	resultMap := make(map[string]interface{})
	msg := ""
	if sendID == "" {
		msg = "sendID"
	}
	if fromAddr == "" {
		msg = "fromAddr"
	}
	if toAddr == "" {
		msg = "toAddr"
	}
	if msg != "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	fAddr := nodeStore.AddressFromB58String(fromAddr)
	tAddr := nodeStore.AddressFromB58String(toAddr)
	sendIDBs, err := hex.DecodeString(sendID)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "sendID")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	ERR := im.SendFriendMsgAgainAPIV2(fAddr, tAddr, sendIDBs)
	//utils.Log.Info().Msgf("发送广播消息")
	//没有目标节点，则是广播消息
	//广播消息
	//im.SendMulticastMessage(content)
	//return nil
	//ERR := utils.NewErrorSuccess()
	//resultMap["info"] = model.ConverUserInfoVO(userInfo)
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
给一个好友发送文件
@addrStr     string    好友地址
@filePath    string    文件路径
*/
func (a *SdkApi) SendFiles_old(addrStr, filePath string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	resultMap := make(map[string]interface{})
	if addrStr == "" || filePath == "" {
		//return errors.New("参数不能为 \"\"")
		msg := "addrStr"
		if filePath == "" {
			msg = "filePath"
		}
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	utils.Log.Info().Msgf("发送文件:%s", filePath)
	addr := nodeStore.AddressFromB58String(addrStr)
	//先发送文件
	tm, ERR := transfer_manager.TransferMangerStatic.NewPushTask(filePath, addr)
	if ERR.CheckFail() {
		//return err
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	utils.Log.Info().Msgf("发送文件ID:%+v", tm)

	_, fileName := filepath.Split(filePath)
	//先发送消息
	ERR = im.SendFriendMsgAPI_old(addr, fileName, config.MSG_type_file_old, tm.PushTaskID, "")
	if !ERR.CheckSuccess() {
		//return err
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	utils.Log.Info().Msgf("tm:%+v", tm)
	//return err
	ERR = utils.NewErrorSuccess()
	//resultMap["list"] = list
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
给一个好友发送文件
@addrStr     string    好友地址
@filePath    string    文件路径
*/
func (a *SdkApi) SendFiles(addrStr, filePath string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("发送文件:%s", filePath)
	resultMap := make(map[string]interface{})
	if addrStr == "" || filePath == "" {
		//return errors.New("参数不能为 \"\"")
		msg := "addrStr"
		if filePath == "" {
			msg = "filePath"
		}
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	//_, fileName := filepath.Split(filePath)
	//先发送消息
	ERR := im.SendFile(addr, filePath)
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
给一个好友发送图片
*/
func (a *SdkApi) SendImage(addrStr, imgBase64 string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	resultMap := make(map[string]interface{})
	if imgBase64 == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "imgBase64")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if addrStr == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	ERR := im.SendImageBase64(addr, imgBase64)
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

/*
构建出完整的文件到指定文件夹
*/
func (a *SdkApi) JoinBuildFile(addrStr, sendIdStr, filePath string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	resultMap := make(map[string]interface{})
	if sendIdStr == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "sendIdStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if addrStr == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)

	sendId, err := hex.DecodeString(sendIdStr)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "sendIdStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Msgf("构建文件到指定目录:%s", runtime.GOOS)
	filePath, ERR := im.JoinBuildFile(addr, sendId, filePath)
	//ERR := utils.NewErrorSysSelf(err)
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg

	//成功则打开文件夹
	if ERR.CheckSuccess() && runtime.GOOS == "windows" {
		dir, _ := filepath.Split(filePath)
		//cmd := "explorer.exe " + dir
		//utils.Log.Info().Msgf("windows系统打开文件夹:%s", cmd)
		err = exec.Command("cmd", "/c", "explorer", dir).Start()
		if err != nil {
			utils.Log.Info().Msgf("打开文件错误:%s %s", err.Error(), dir)
		}
	}

	return resultMap

	//给前端发送一个通知
	//msgInfo := im.NewMessageInfo(config.SUBSCRIPTION_type_msg, config.MSG_type_image, true, config.NetAddr, addrStr,
	//	imgBase64, utils.FormatTimeToSecond(time.Now()), 0, nil, 0)
	//im.AddSubscriptionMsg(msgInfo)
	//return nil
}

/*
给一个好友发送语音消息
@addrStr     string    * 好友地址
@mimeType    string      文件类型
@voiceCoding    string      语音编码字符串
@voiceSecond    int64       语音秒数
*/
func (a *SdkApi) SendVoiceBase64(addrStr, mimeType, voiceCoding string, voiceSecond int64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Msgf("发送语音消息:%s", filePath)
	resultMap := make(map[string]interface{})
	if addrStr == "" {
		msg := "addrStr"
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	if mimeType == "" || voiceCoding == "" {
		//return errors.New("参数不能为 \"\"")
		msg := "mimeType"
		if voiceCoding == "" {
			msg = "voiceCoding"
		}
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	//_, fileName := filepath.Split(filePath)
	//发送语音文件消息
	ERR := im.SendVoice(addr, "", mimeType, voiceCoding, voiceSecond)
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
给一个好友发送语音消息
传文件路径，文件类型和语音编码传空
@addrStr     string    * 好友地址
@filePath    string      文件路径
@mimeType    string      文件类型
@voiceCoding    string      语音编码字符串
@voiceSecond    int64     * 语音秒数
*/
func (a *SdkApi) SendVoiceFile(addrStr, filePath string, voiceSecond int64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("发送语音消息:%s", filePath)
	resultMap := make(map[string]interface{})
	if addrStr == "" {
		msg := "addrStr"
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	if filePath == "" {
		//return errors.New("参数不能为 \"\"")
		msg := "filePath"
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, msg)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	//_, fileName := filepath.Split(filePath)
	//发送语音文件消息
	ERR := im.SendVoice(addr, filePath, "", "", voiceSecond)
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
接收消息
*/
func (a *SdkApi) IM_GetMsg() map[string]interface{} {
	resultMap := make(map[string]interface{})
	itr := subscription.GetSubscriptionMsg(time.Hour)
	//return
	ERR := utils.NewErrorSuccess()
	resultMap["info"] = itr
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询好友基本信息
@addrStr    string    好友地址
*/
func (a *SdkApi) IM_GetFriendInfo(addrStr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	resultMap := make(map[string]interface{})
	if addrStr == "" {
		//return nil
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	userinfo, ERR := db.FindUserListByAddr(*im.Node.GetNetId(), addr)
	//userinfo, ERR := im.FindUserInFriendList(addr)
	if !ERR.CheckSuccess() {
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	have := false
	if userinfo != nil {
		have = true
		//如果列表中没有详细信息，则查询一次数据库
		if userinfo.Nickname == "" {
			ui, ERR := db.ImProxyClient_FindUserinfo(*im.Node.GetNetId(), addr)
			if ERR.CheckSuccess() && ui != nil {
				userinfo.Nickname = ui.Nickname
				userinfo.HeadNum = ui.HeadNum
				userinfo.GroupDHPuk = ui.GroupDHPuk
			}
		}
	}
	go func(have bool) {
		//网络获取好友信息
		// userinfoVO := a.IM_SearchFriendInfo(addrStr)
		userInfo, ERR := im.GetFriendInfoAPI(addr)
		if !ERR.CheckSuccess() {
			return
		}
		//对比新的好友信息和本地保存的是否一样
		change := false
		if userinfo == nil {
			change = true
		} else {
			if userInfo.Nickname != userinfo.Nickname {
				change = true
			}
			if userInfo.HeadNum != userinfo.HeadNum {
				change = true
			}
			if userInfo.GroupDHPuk == nil || len(userInfo.GroupDHPuk) == 0 {
				userInfo.GroupDHPuk = userinfo.GroupDHPuk
			}
		}
		if !change {
			return
		}
		//修改好友信息
		ERR = db.ImProxyClient_SaveUserinfo(*im.Node.GetNetId(), userInfo)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("修改好友信息 错误:%s", ERR.String())
			return
		}
		//ERR = db.SaveUserList(config.DBKEY_friend_userlist, userInfo, nil)
		//db.UpdateUserList(*userInfo, config.DBKEY_friend_userlist)
		//if !ERR.CheckSuccess() {
		//  return
		//}
		userInfoVO := model.ConverUserInfoVO(userInfo)
		m := make(map[string]interface{})
		m["Addr"] = userInfoVO.Addr
		m["Nickname"] = userInfoVO.Nickname
		m["HeadNum"] = userInfoVO.HeadNum
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_update_userinfo, Data: m}
		subscription.AddSubscriptionMsg(&msgInfo)
	}(have)
	if have {
		//return model.ConverUserInfoVO(userinfo)
		ERR := utils.NewErrorSuccess()
		resultMap["info"] = model.ConverUserInfoVO(userinfo)
		resultMap["have"] = true
		resultMap["code"] = ERR.Code
		return resultMap
	}
	//return nil
	ERR = utils.NewErrorSuccess()
	//resultMap["info"] = itr
	resultMap["have"] = false
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
搜索好友基本信息
@addrStr    string    好友地址
*/
func (a *SdkApi) IM_SearchFriendInfo(addrStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	//utils.Log.Info().Msgf("搜索好友地址:%s", addrStr)
	if addrStr == "" {
		//return nil
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	userInfo, ERR := im.GetFriendInfoAPI(addr)
	if !ERR.CheckSuccess() {
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	userInfoVO := model.ConverUserInfoVO(userInfo)
	//return userInfoVO
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = userInfoVO
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
添加好友
@addrStr    string    好友地址
*/
func (a *SdkApi) IM_AddFriend(addrStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	addr := nodeStore.AddressFromB58String(addrStr)
	ERR := im.AddFriendAPI(addr)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	subscription.AddSubscriptionMsg(&msgInfo)
	//return e
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = msgInfo
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
删除好友
@addrStr    string    好友地址
*/
func (a *SdkApi) IM_DelFriend(addrStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	addr := nodeStore.AddressFromB58String(addrStr)
	ERR := im.DelFriendAPI(addr)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	//给前端发送一个通知
	//im.AddSubscriptionMsg(&msgInfo)
	//return e
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = msgInfo
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
获取好友申请列表
*/
func (a *SdkApi) IM_GetNewFriend() map[string]interface{} {
	resultMap := make(map[string]interface{})
	userList, ERR := im.GetNewFriendAPI()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("获取好友申请列表失败:%s", ERR.String())
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return userList
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = userList
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
同意好友申请
@tokenStr    string    申请记录令牌
*/
func (a *SdkApi) IM_AgreeApplyFriend(tokenStr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Msgf("同意好友申请")
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
	//utils.Log.Info().Msgf("同意好友申请")
	ERR := im.AgreeApplyFriendAPI(token)
	if !ERR.CheckSuccess() {
		//return err
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Msgf("同意好友申请")
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	subscription.AddSubscriptionMsg(&msgInfo)
	//utils.Log.Info().Msgf("同意好友申请")
	//return nil
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = msgInfo
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
好友列表
*/
func (a *SdkApi) IM_GetFriendList() map[string]interface{} {
	resultMap := make(map[string]interface{})
	userlist, ERR := im.GetFriendListAPI()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("获取好友列表失败:%s", ERR.String())
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return userlist
	ERR = utils.NewErrorSuccess()
	resultMap["info"] = userlist
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
设置好友昵称备注
*/
func (a *SdkApi) IM_SetFriendRemarksname(addrStr, remarksname string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	if addrStr == "" {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	addr := nodeStore.AddressFromB58String(addrStr)
	ERR := im.SetFriendRemarksName(addr, remarksname)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("失败:%s", ERR.String())
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["info"] = userList
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询消息历史记录
@startIndex    string    查询消息范围起始index，首次查询传空字符串
@count         uint64    查询消息条数
@remoteAddr    string    查询的好友地址
*/
func (a *SdkApi) GetChatHistory(startIndex string, count uint64, remoteAddr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	utils.Log.Info().Msgf("查询消息历史记录:%s %d %s", startIndex, count, remoteAddr)
	resultMap := make(map[string]interface{})
	var startIndexBs []byte
	if startIndex != "" {
		var err error
		startIndexBs, err = hex.DecodeString(startIndex)
		if err != nil {
			ERR := utils.NewErrorSysSelf(err)
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	if remoteAddr == "" {
		utils.Log.Info().Msgf("查询消息历史记录失败:addr is \"\"")
		//return nil
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "remoteAddr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(remoteAddr)
	msgList, ERR := im.GetChatHistoryAPI(startIndexBs, count, addr)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("查询消息历史记录失败:%s", ERR.String())
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	resultMap["list"] = msgList
	resultMap["code"] = ERR.Code
	return resultMap

}

/*
删除所有聊天消息历史记录
*/
func (a *SdkApi) IM_RemoveChatHistoryAll(remoteAddr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	//if remoteAddr == "" {
	//	utils.Log.Info().Msgf("删除所有聊天消息历史记录 失败:addr is \"\"")
	//	//return nil
	//	ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "remoteAddr")
	//	resultMap["code"] = ERR.Code
	//	resultMap["error"] = ERR.Msg
	//	return resultMap
	//}
	//addr := nodeStore.AddressFromB58String(remoteAddr)
	//msgList, err := im.GetChatHistoryAPI(startIndex, count, addr)
	//if err != nil {
	//	utils.Log.Info().Msgf("查询消息历史记录失败:%s", err.Error())
	//	//return nil
	//	ERR := utils.NewErrorSysSelf(err)
	//	resultMap["code"] = ERR.Code
	//	resultMap["error"] = ERR.Msg
	//	return resultMap
	//}
	//list := model.ConverMessageContentListVO(msgList)
	ERR := utils.NewErrorSuccess()
	//resultMap["list"] = list
	resultMap["code"] = ERR.Code
	return resultMap

}

/*
通知取消闪烁
*/
func (a *SdkApi) IM_NoticeCancelFlicker(key string) map[string]interface{} {
	//utils.Log.Info().Msgf("通知取消闪烁")
	resultMap := make(map[string]interface{})
	tray.SendNotice(false, key)
	ERR := utils.NewErrorSuccess()
	resultMap["code"] = ERR.Code
	return resultMap
}
