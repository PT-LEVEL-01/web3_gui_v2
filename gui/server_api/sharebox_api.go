package server_api

import (
	"encoding/hex"
	chainconfig "web3_gui/chain/config"
	"web3_gui/config"
	"web3_gui/im/im"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/cake/transfer_manager"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/storage"
	"web3_gui/utils"
)

/*
获取共享目录列表
*/
func (a *SdkApi) File_GetShareboxList() map[string]interface{} {
	resultMap := make(map[string]interface{})
	list, ERR := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirs()
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取共享目录列表错误:%s", ERR.String())
		//return nil
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//return list
		ERR := utils.NewErrorSuccess()
		resultMap["list"] = list
		resultMap["code"] = ERR.Code
		return resultMap
	}
}

/*
查询好友的共享文件
@addrStr    string    好友地址
@dirPath    string    目录路径
*/
func (a *SdkApi) IM_GetShareboxList(addrStr, dirPath string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	//utils.Log.Info().Msgf("开始查询好友共享文件列表:%s %s", addrStr, dirPath)
	if addrStr == "" {
		utils.Log.Info().Msgf("开始查询好友共享文件列表")
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "addrStr")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	addr := nodeStore.AddressFromB58String(addrStr)
	// utils.Log.Info().Msgf("给指定节点发送消息:%s %s", addr.B58String(), content)
	fileList, ERR := im.GetShareboxList(addr, dirPath)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("开始查询好友共享文件列表:%s", ERR.String())
		//return nil
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Msgf("开始查询好友共享文件列表:%v", fileList)
	//return fileList
	ERR = utils.NewErrorSuccess()
	resultMap["list"] = fileList
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
添加共享目录
*/
func (a *SdkApi) File_AddSharebox(dirPath string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	ERR := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirsAdd(dirPath)
	if ERR.CheckFail() {
		//return model.Success
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//return model.Success
		ERR := utils.NewErrorSuccess()
		//resultMap["paths"] = results
		resultMap["code"] = ERR.Code
		return resultMap
	}
}

/*
删除共享目录
*/
func (a *SdkApi) File_DelSharebox(dirPath string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	ERR := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirsDel(dirPath)
	if ERR.CheckFail() {
		//return model.Success
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	} else {
		//return model.Success
		ERR := utils.NewErrorSuccess()
		//resultMap["paths"] = results
		resultMap["code"] = ERR.Code
		return resultMap
	}
}

/*
获取多个文件、文件夹中的所有文件的详细信息
*/
func (a *SdkApi) Sharebox_GetFileInfo(dirPaths, filePaths []string) map[string]interface{} {
	//utils.Log.Info().Interface("文件夹", dirPaths).Interface("文件列表", filePaths).Send()
	resultMap := make(map[string]interface{})
	fhp, ERR := storage.FileHashM.AddProcess(append(dirPaths, filePaths...))
	if ERR.CheckFail() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	filePriceVOs := make([]model.FilePriceVO, 0, len(fhp.FilePrices))
	for _, one := range fhp.FilePrices {
		filePriceVOs = append(filePriceVOs, *one.ConverVO())
	}
	ERR = utils.NewErrorSuccess()
	resultMap["pid"] = hex.EncodeToString(fhp.Id)
	resultMap["list"] = filePriceVOs
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
获取异步计算的文件详细信息
*/
func (a *SdkApi) Sharebox_GetFileHash(pid string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	pidBs, err := hex.DecodeString(pid)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "pid")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//utils.Log.Info().Interface("文件夹", dirPaths).Interface("文件列表", filePaths).Send()
	fhp := storage.FileHashM.FindProcess(pidBs)
	if fhp == nil {
		ERR := utils.NewErrorSuccess()
		resultMap["list"] = nil
		resultMap["code"] = ERR.Code
		return resultMap
	}
	filePriceVOs := make([]model.FilePriceVO, 0, len(fhp.FilePrices))
	for _, one := range fhp.FilePrices {
		filePriceVOs = append(filePriceVOs, *one.ConverVO())
	}
	ERR := utils.NewErrorSuccess()
	resultMap["list"] = filePriceVOs
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
设置一个文件的价格
*/
func (a *SdkApi) Sharebox_SetFilePrice(fileHash string, price uint64) map[string]interface{} {
	resultMap := make(map[string]interface{})
	bs, err := hex.DecodeString(fileHash)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "fileHash")
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR := storage.SavePrice(bs, price)
	//resultMap["list"] = filePrices
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
获取一个文件的订单
@remoteAddr    string    远程节点地址
@fileHash      string    文件hash
*/
func (a *SdkApi) Sharebox_GetFileOrder(remoteAddr, fileHash string, price uint64) map[string]interface{} {
	resultMap := make(map[string]interface{})
	params := make(map[string]interface{})
	params["addr"] = remoteAddr
	params["fileHash16"] = fileHash
	params["price"] = price
	reusltMap, ERR := config.Node.RunRpcMethod(config.RPC_Method_Sharebox_GetFileOrder, params)
	if ERR.CheckFail() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	return reusltMap
}

/*
支付一个订单
@remoteAddr    string    远程节点地址
@fileHash      string    文件hash
*/
func (a *SdkApi) PayOrder(serverAddr, orderId16 string, amount uint64, pwd string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	params := make(map[string]interface{})
	params["serverAddr"] = serverAddr
	params["orderId16"] = orderId16
	params["amount"] = amount
	params["pwd"] = pwd
	params["gas"] = chainconfig.Wallet_tx_gas_min
	//utils.Log.Info().Interface("参数", params).Send()
	reusltMap, ERR := config.Node.RunRpcMethod(config.RPC_Method_chain_PayOrder, params)
	if ERR.CheckFail() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	return reusltMap
}
