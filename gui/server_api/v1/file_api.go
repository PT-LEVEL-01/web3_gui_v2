package server_api

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/wailsapp/wails/v3/pkg/application"
	"path/filepath"
	"web3_gui/config"
	"web3_gui/libp2parea/v1/cake/transfer_manager"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
获取共享目录列表
*/
func (a *SdkApi) File_GetShareboxList() map[string]interface{} {
	resultMap := make(map[string]interface{})
	list, err := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirs()
	if err != nil && err.Error() != leveldb.ErrNotFound.Error() {
		utils.Log.Info().Msgf("获取共享目录列表错误:%s", err.Error())
		//return nil
		ERR := utils.NewErrorSysSelf(err)
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
打开目录对话框
*/
func (a *SdkApi) File_OpenDirectoryDialog() map[string]interface{} {
	resultMap := make(map[string]interface{})
	result, err := application.OpenFileDialog().CanChooseDirectories(true).PromptForSingleSelection()
	if err != nil {
		utils.Log.Error().Msgf("打开文件夹选择框错误:%s", err.Error())
		//return ""
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//if result == "" {
	//	application.InfoDialog().SetMessage(result).Show()
	//}

	//result, err := application.OpenDirectoryDialog(a.Ctx, application.Options{
	//	DefaultDirectory: "",      //对话框打开时显示的目录
	//	DefaultFilename:  "",      //默认文件名
	//	Title:            "选择文件夹", //对话框的标题
	//	// Filters:                    "", //文件过滤器列表
	//	ShowHiddenFiles:            true,  //显示系统隐藏的文件
	//	CanCreateDirectories:       false, //允许用户创建目录
	//	ResolvesAliases:            false, //如果为 true，则返回文件而不是别名
	//	TreatPackagesAsDirectories: false, //允许导航到包
	//})
	//if err != nil {
	//	utils.Log.Error().Msgf("打开文件夹选择框错误:%s", err.Error())
	//	//return ""
	//	ERR := utils.NewErrorSysSelf(err)
	//	resultMap["code"] = ERR.Code
	//	resultMap["error"] = ERR.Msg
	//	return resultMap
	//}
	//return result
	ERR := utils.NewErrorSuccess()
	resultMap["path"] = result
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
打开多文件选择对话框
*/
func (a *SdkApi) File_OpenMultipleFilesDialog() map[string]interface{} {
	resultMap := make(map[string]interface{})
	results, err := application.OpenFileDialog().
		CanChooseFiles(true).
		CanCreateDirectories(true).
		ShowHiddenFiles(true).
		PromptForMultipleSelection()

	//results, err := application.OpenMultipleFilesDialog(a.Ctx, application.OpenDialogOptions{
	//	DefaultDirectory: "",     //对话框打开时显示的目录
	//	DefaultFilename:  "",     //默认文件名
	//	Title:            "选择文件", //对话框的标题
	//	// Filters:                    "", //文件过滤器列表
	//	ShowHiddenFiles:            true,  //显示系统隐藏的文件
	//	CanCreateDirectories:       false, //允许用户创建目录
	//	ResolvesAliases:            false, //如果为 true，则返回文件而不是别名
	//	TreatPackagesAsDirectories: false, //允许导航到包
	//})
	if err != nil {
		utils.Log.Error().Msgf("打开文件夹选择框错误:%s", err.Error())
		//return []string{}
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return results
	ERR := utils.NewErrorSuccess()
	resultMap["paths"] = results
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
添加共享目录
*/
func (a *SdkApi) File_AddSharebox(dirPath string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	err := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirsAdd(dirPath)
	if err != nil {
		//return model.Success
		ERR := utils.NewErrorSysSelf(err)
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
	err := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirsDel(dirPath)
	if err != nil {
		//return model.Success
		ERR := utils.NewErrorSysSelf(err)
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
下载文件
*/
func (a *SdkApi) File_download(addrStr, filePath string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	_, fileName := filepath.Split(filePath)
	addr := nodeStore.AddressFromB58String(addrStr)
	err := transfer_manager.TransferMangerStatic.NewPullTask(filePath, filepath.Join(config.DownloadFileDir, fileName), addr)
	if err != nil {
		//return model.Success
		ERR := utils.NewErrorSysSelf(err)
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
打开文件多选框
*/
func (a *SdkApi) OpenFileDialog() map[string]interface{} {
	return a.File_OpenMultipleFilesDialog()
	//resultMap := make(map[string]interface{})
	//
	//results, err := application.OpenMultipleFilesDialog(a.Ctx, application.OpenDialogOptions{
	//	DefaultDirectory: "",         //对话框打开时显示的目录
	//	DefaultFilename:  "",         //默认文件名
	//	Title:            "选择要发送的文件", //对话框的标题
	//	// Filters:                    "", //文件过滤器列表
	//	ShowHiddenFiles:            true,  //显示系统隐藏的文件
	//	CanCreateDirectories:       false, //允许用户创建目录
	//	ResolvesAliases:            false, //如果为 true，则返回文件而不是别名
	//	TreatPackagesAsDirectories: false, //允许导航到包
	//})
	//if err != nil {
	//	utils.Log.Error().Msgf("打开文件选择框错误:%s", err.Error())
	//	//return []string{}
	//	ERR := utils.NewErrorSysSelf(err)
	//	resultMap["code"] = ERR.Code
	//	resultMap["error"] = ERR.Msg
	//	return resultMap
	//}
	////return result
	//ERR := utils.NewErrorSuccess()
	//resultMap["paths"] = results
	//resultMap["code"] = ERR.Code
	//return resultMap
}

/*
查看文件下载列表
*/
func (a *SdkApi) GetFileDownloadList() map[string]interface{} {
	resultMap := make(map[string]interface{})
	fileinfos := make([]*transfer_manager.FileinfoVO, 0)
	list := transfer_manager.TransferMangerStatic.PullTaskList()
	for _, one := range list {
		fileOne := transfer_manager.ConverPullTaskVO(one)
		fileinfos = append(fileinfos, &fileOne)
	}
	//return fileinfos
	ERR := utils.NewErrorSuccess()
	resultMap["list"] = fileinfos
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
暂停一个下载任务
*/
func (a *SdkApi) File_StopDownload(pullTaskID uint64) map[string]interface{} {
	resultMap := make(map[string]interface{})
	err := transfer_manager.TransferMangerStatic.PullTaskStop(pullTaskID)
	if err != nil {
		//return model.Nomarl
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return model.Success
	ERR := utils.NewErrorSuccess()
	//resultMap["list"] = fileinfos
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
继续一个下载任务
*/
func (a *SdkApi) File_StartDownload(pullTaskID uint64) map[string]interface{} {
	resultMap := make(map[string]interface{})
	err := transfer_manager.TransferMangerStatic.PullTaskStart(pullTaskID, "")
	if err != nil {
		//return model.Nomarl
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return model.Success
	ERR := utils.NewErrorSuccess()
	//resultMap["list"] = fileinfos
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
删除一个下载任务
*/
func (a *SdkApi) File_DelDownload(pullTaskID uint64) map[string]interface{} {
	resultMap := make(map[string]interface{})
	err := transfer_manager.TransferMangerStatic.PullTaskDel(pullTaskID)
	if err != nil {
		//return model.Nomarl
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//return model.Success
	ERR := utils.NewErrorSuccess()
	//resultMap["list"] = fileinfos
	resultMap["code"] = ERR.Code
	return resultMap
}
