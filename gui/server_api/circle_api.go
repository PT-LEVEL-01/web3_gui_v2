package server_api

import (
	"encoding/hex"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
查询类别
*/
func (a *SdkApi) Circle_FindNewsClass() map[string]interface{} {
	resultMap := make(map[string]interface{})
	classNames, ERR := db.GetClass(*config.DBKEY_circle_news_release)
	if ERR.CheckFail() {
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["Class"] = classNames
		ERR := utils.NewErrorSuccess()
		resultMap["Class"] = classNames
		resultMap["code"] = ERR.Code
		return resultMap
	}
	//return m
}

/*
保存类别
*/
func (a *SdkApi) Circle_SaveNewsClass(className string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	classNames, ERR := db.GetClass(*config.DBKEY_circle_news_release)
	if ERR.CheckFail() {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		//return m
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	for _, one := range classNames {
		if one == className {
			//m["ErrorCode"] = model.Exist
			//return m
			//ERR := utils.NewErrorSysSelf(err)
			resultMap["code"] = config.ERROR_CODE_CIRCLE_classname_exist
			resultMap["error"] = ""
			return resultMap
		}
	}
	err := db.AddClass(*config.DBKEY_circle_news_release, className)
	if err != nil {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
	}
	ERR = utils.NewErrorSuccess()
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
保存到草稿列表
*/
func (a *SdkApi) Circle_SaveNewsDraft(className, title, content string, indexBsStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	var err error
	var index []byte
	if indexBsStr == "" {
		index, err = im.AddNewsToDraft(className, title, content)
	} else {
		index, err = hex.DecodeString(indexBsStr)
		if err != nil {
			ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "indexBsStr")
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
		index, err = db.UpdateNews(*config.DBKEY_circle_news_draft, index, className, title, content)
	}
	if err != nil {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["Index"] = index
	}
	//return m
	ERR := utils.NewErrorSuccess()
	resultMap["Index"] = hex.EncodeToString(index)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
保存到发布列表
*/
func (a *SdkApi) Circle_SaveNewsRelease(className, title, content string, indexBsStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	var err error
	var index []byte
	if indexBsStr == "" {
		index, err = im.AddNewsToRelease(className, title, content)
	} else {
		index, err = hex.DecodeString(indexBsStr)
		if err != nil {
			ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "indexBsStr")
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
		index, err = db.UpdateNews(*config.DBKEY_circle_news_release, index, className, title, content)
	}
	if err != nil {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["Index"] = index
	}
	//return m
	ERR := utils.NewErrorSuccess()
	resultMap["Index"] = hex.EncodeToString(index)
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询草稿箱中的新闻
*/
func (a *SdkApi) Circle_FindNewsDraft(className string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	news, ERR := im.FindNewsDraft(className)
	if !ERR.CheckSuccess() {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["News"] = news
	}
	//return m
	ERR = utils.NewErrorSuccess()
	resultMap["News"] = news
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询我的发布中的新闻
*/
func (a *SdkApi) Circle_FindNewsRelease(className string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	news, ERR := im.FindNewsRelease(className)
	if !ERR.CheckSuccess() {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["News"] = news
	}
	//return m
	//ERR := utils.NewErrorSuccess()
	resultMap["News"] = news
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询好友博客圈子列表
*/
func (a *SdkApi) Circle_FindClassNames(addrStr string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	utils.Log.Info().Msgf("查询好友博客圈子列表:%s", addrStr)
	if addrStr == "" {
		utils.Log.Info().Msgf("查询好友博客圈子列表")
		//return nil
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	//m := make(map[string]interface{})
	fileList, ERR := im.GetUserCircleClassNames(addr)
	if !ERR.CheckSuccess() {
		//m["ErrorCode"] = model.Nomarl
		//m["Message"] = err.Error()
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//m["ErrorCode"] = model.Success
		//m["ClassNames"] = fileList
	}
	//return m
	ERR = utils.NewErrorSuccess()
	resultMap["ClassNames"] = fileList
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询广播中收集的博客
*/
func (a *SdkApi) Circle_FindClassNamesMulticast() map[string]interface{} {
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	resultMap := make(map[string]interface{})
	fileList := im.GetMultcastClassCount()
	//m["ErrorCode"] = model.Success
	//m["ClassNames"] = fileList
	//return m
	ERR := utils.NewErrorSuccess()
	resultMap["ClassNames"] = fileList
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
查询广播中的博客，指定的一个分类博客列表
*/
func (a *SdkApi) Circle_FindClassNamesMulticastNewsList(className string) map[string]interface{} {
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	resultMap := make(map[string]interface{})
	fileList := im.GetMultcastClassNewsList(className)
	//m["ErrorCode"] = model.Success
	//m["News"] = fileList
	//return m
	ERR := utils.NewErrorSuccess()
	resultMap["ClassNames"] = fileList
	resultMap["code"] = ERR.Code
	return resultMap
}
