package server_api

import (
	"bytes"
	"cmp"
	"encoding/hex"
	"math"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/keystore/v1/base58"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/storage"
	"web3_gui/utils"
)

/*
获取存储服务器状态信息
*/
func (a *SdkApi) Storage_server_GetStatus() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
		ERR = db.StorageServer_SetServerInfo(sinfo)
		if !ERR.CheckSuccess() {
			//ERR = utils.NewErrorSysSelf(err)
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	sinfoVO := sinfo.ConverVO()
	//检查剩余容量
	sinfoVO.DirectoryFreeSize = make([]uint64, len(sinfoVO.Directory))
	for i, one := range sinfoVO.Directory {
		_, _, freeSize, err := utils.GetDiskFreeSpace(one)
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
		sinfoVO.DirectoryFreeSize[i] = freeSize / 1024 / 1024 / 1024
	}
	resultMap["storageServerInfo"] = sinfoVO
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
打开或关闭云存储服务器
*/
func (a *SdkApi) Storage_server_SetOpen(isOpen bool) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	sinfo.IsOpen = isOpen

	ERR = db.StorageServer_SetServerInfo(sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if isOpen {
		storage.SetStorageServerOpen()
	} else {
		storage.SetStorageServerClose()
	}
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
添加存储目录
*/
func (a *SdkApi) Storage_server_AddDirectory(dir string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	dir = filepath.Join(dir, config.STORAGE_server_dbpath_name)
	ok, err := utils.PathExists(dir)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if !ok {
		err := utils.Mkdir(dir)
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}

	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	//对比是否重复
	for _, one := range sinfo.Directory {
		if one == dir {
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	ERR = storage.StServer.DBC.AddDBone(dir)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	sinfo.Directory = append(sinfo.Directory, dir)
	ERR = db.StorageServer_SetServerInfo(sinfo)
	if !ERR.CheckSuccess() {
		storage.StServer.DBC.DelDBone(dir)
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
删除存储目录
*/
func (a *SdkApi) Storage_server_DelDirectory(dir string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	//对比并删除
	for i, one := range sinfo.Directory {
		if one == dir {
			temp := sinfo.Directory[:i]
			temp = append(temp, sinfo.Directory[i+1:]...)
			sinfo.Directory = temp
			break
		}
	}
	ERR = db.StorageServer_SetServerInfo(sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	err = storage.StServer.DBC.DelDBone(dir)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
获取自己订单中云存储服务器列表
*/
func (a *SdkApi) Storage_client_GetStorageServiceList() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	storageServerList, err := storage.StClient.GetOrderServerList()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	storageServerListVO := make([]*model.StorageServerInfoVO, 0)
	for _, one := range storageServerList {
		vo := one.ConverVO()
		selling, sold, _ := storage.StClient.QueryUserSpacesLocal(one.Addr)
		vo.Selling = uint64(selling)
		vo.Sold = uint64(sold)
		//查询已经使用的容量
		storageServerListVO = append(storageServerListVO, vo)
	}
	//排序
	sort.SliceStable(storageServerListVO, func(i, j int) bool {
		return storageServerListVO[i].Count < storageServerListVO[j].Count
	})
	resultMap["list"] = storageServerListVO
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
获取搜索到的云存储服务器列表
*/
func (a *SdkApi) Storage_client_GetSearchStorageServiceList() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	storageServerList, err := db.StorageClient_GetStorageServerList()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}

	storageServerListVO := make([]*model.StorageServerInfoVO, 0)
	for _, one := range storageServerList {
		storageServerListVO = append(storageServerListVO, one.ConverVO())
	}

	sort.SliceStable(storageServerListVO, func(i, j int) bool {
		return storageServerListVO[i].Count < storageServerListVO[j].Count
	})

	resultMap["list"] = storageServerListVO
	resultMap["code"] = ERR.Code
	return resultMap
}

var Storage_server_SetSelling_Lock = new(sync.Mutex)

/*
设置本节点正在销售的存储空间
@selling    int64    要添加或减少的存储空间
*/
func (a *SdkApi) Storage_server_SetSelling(selling float64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	Storage_server_SetSelling_Lock.Lock()
	defer Storage_server_SetSelling_Lock.Unlock()
	resultMap := make(map[string]interface{})
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	//判断整数
	if math.Mod(selling, 1) != 0 {
		resultMap["code"] = config.ERROR_CODE_params_format // rpcmodel.TypeWrong
		resultMap["error"] = "selling"
		return resultMap
	}
	sinfo.Selling = sinfo.Selling + uint64(selling)
	ERR = db.StorageServer_SetServerInfo(sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
设置本节点存储空间单价和空间限制
@priceUnit     uint64   //单价 单位：1G
@selling       uint64   //售卖总容量 单位：1G
@userFreelimit uint64   //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
@userCanTotal  uint64   //每个用户可以购买的空间总量 单位：1G
@useTimeMax    uint64   //每个订单租用时间最大值 单位：天
@renewalTime   uint64   //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
*/
func (a *SdkApi) Storage_server_SetPriceUnit(nickName string, priceUnit, selling, userFreelimit, userCanTotal, useTimeMax,
	renewalTime uint64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	Storage_server_SetSelling_Lock.Lock()
	defer Storage_server_SetSelling_Lock.Unlock()
	resultMap := make(map[string]interface{})
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	sinfo.Nickname = nickName
	sinfo.PriceUnit = priceUnit
	sinfo.Selling = selling
	sinfo.UserFreelimit = userFreelimit
	sinfo.UserCanTotal = userCanTotal
	sinfo.UseTimeMax = useTimeMax
	sinfo.RenewalTime = renewalTime
	ERR = db.StorageServer_SetServerInfo(sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
获取订单列表
*/
func (a *SdkApi) Storage_Client_GetOrderList() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	orders := storage.StClient.GetOrdersList()
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	resultMap["list"] = orders
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
获取订单
@serverAddr    uint64     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func (a *SdkApi) Storage_Client_GetOrders(serverAddr string, spaceTotal, useTime uint64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	resultMap := make(map[string]interface{})
	orders, ERR := storage.ClientGetOrders(sAddr, spaceTotal, useTime)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ordersVO := orders.ConverVO()
	resultMap["orders"] = ordersVO
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
订单续费
@preNumber     string     //前一个订单的id
@serverAddr    string     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func (a *SdkApi) Storage_Client_GetRenewalOrders(preNumber, serverAddr string, spaceTotal, useTime uint64) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	resultMap := make(map[string]interface{})
	number, err := hex.DecodeString(preNumber)
	if err != nil {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "preNumber"
		return resultMap
	}
	orders, ERR := storage.ClientGetRenewalOrders(number, sAddr, spaceTotal, useTime)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	ordersVO := orders.ConverVO()
	resultMap["orders"] = ordersVO
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
客户端获取文件列表
@serverAddr    string    服务器节点地址
@dirIDStr      string    文件夹id，顶层文件夹传空字符串
*/
func (a *SdkApi) Storage_Client_GetFileList(serverAddr string, dirIDStr string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	var dirID []byte
	if dirIDStr != "" {
		dirID = base58.Decode(dirIDStr)
	}
	dIndex, ERR := storage.GetFileList(sAddr, dirID)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	a.dirIndex = dIndex
	ERR = utils.NewErrorSuccess()
	resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
客户端上传文件
@serverAddr    string      服务器地址
@dirID         string      远端文件夹id
@filesPath     []string    需上传的文件路径
*/
func (a *SdkApi) Storage_Client_UploadFiles(serverAddr string, dirID string, filesPath []string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if dirID == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "dirID"
		return resultMap
	}
	dID := base58.Decode(dirID)
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	ERR = storage.StClient.UploadFile(sAddr, dID, filesPath...)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.String()
		return resultMap
	}
	//resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
创建新文件夹
@serverAddr        string    服务器地址
@parentDirIDStr    string    上级文件夹id
@dirName           string    文件夹名称
*/
func (a *SdkApi) Storage_Client_CreateDir(serverAddr string, parentDirIDStr, dirName string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Str("创建文件夹", parentDirIDStr).Send()
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if parentDirIDStr == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "parentDirID"
		return resultMap
	}
	parentDirID := base58.Decode(parentDirIDStr)
	if dirName == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "dirID"
		return resultMap
	}
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	ERR = storage.StClient.CreateNewDir(sAddr, parentDirID, dirName)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
删除多个文件和文件夹
@serverAddr          string      服务器地址
@dirAndFilesIDStr    []string    文件或文件夹id
*/
func (a *SdkApi) Storage_Client_DelDirAndFile(serverAddr string, dirAndFilesIDStr []string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if serverAddr == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "serverAddr"
		return resultMap
	}
	if len(dirAndFilesIDStr) == 0 {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "dirAndFilesIDStr"
		return resultMap
	}
	//utils.Log.Info().Msgf("要删除的文件和文件夹:%+v", dirAndFilesIDStr)
	dirs := make([][]byte, 0)
	files := make([][]byte, 0)
	have := false
	for _, one := range dirAndFilesIDStr {
		have = false
		idOne := base58.Decode(one)
		for _, dirOne := range a.dirIndex.Dirs {
			if bytes.Equal(dirOne.ID, idOne) {
				dirs = append(dirs, dirOne.ID)
				have = true
				break
			}
		}
		if have {
			continue
		}
		for _, dirOne := range a.dirIndex.Files {
			if bytes.Equal(dirOne.Hash, idOne) {
				files = append(files, dirOne.Hash)
				have = true
				break
			}
		}
		if have {
			continue
		}
		//未找到这个id
		resultMap["code"] = config.ERROR_CODE_Not_present
		resultMap["error"] = one
		return resultMap
	}
	//utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirs, files)
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	ERR = storage.StClient.DelDirAndFiles(sAddr, dirs, files)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
下载文件和文件夹
@serverAddr          string      服务器地址
@dirAndFilesIDStr    []string    下载的文件和文件夹列表
@localPath           string      保存到本地的文件路径
*/
func (a *SdkApi) Storage_Client_download(serverAddr string, dirAndFilesIDStr []string, localPath string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if serverAddr == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "serverAddr"
		return resultMap
	}
	if len(dirAndFilesIDStr) == 0 {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "dirAndFilesIDStr"
		return resultMap
	}

	dirs := make([]*model.DirectoryIndex, 0)
	files := make([]*model.FileIndex, 0)
	have := false
	for _, one := range dirAndFilesIDStr {
		have = false
		idOne := base58.Decode(one)
		for _, dirOne := range a.dirIndex.Dirs {
			if bytes.Equal(dirOne.ID, idOne) {
				dirs = append(dirs, dirOne)
				have = true
				break
			}
		}
		if have {
			continue
		}
		for _, dirOne := range a.dirIndex.Files {
			if bytes.Equal(dirOne.Hash, idOne) {
				files = append(files, dirOne)
				have = true
				break
			}
		}
		if have {
			continue
		}
		//未找到这个id
		resultMap["code"] = config.ERROR_CODE_Not_present
		resultMap["error"] = one
		return resultMap
	}
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	ERR = storage.StClient.DownloadDirAndFiles(sAddr, dirs, files, localPath)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
下载列表
*/
func (a *SdkApi) Storage_Client_downloadList() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	//list := file_transfer.ManagerStatic.GetDownloadList()
	list := storage.StClient.GetDownloadList()
	slices.SortFunc(list, func(a, b *file_transfer.DownloadStepVO) int {
		if a.CreateTime == b.CreateTime {
			return cmp.Compare(a.Name, b.Name)
		} else {
			return cmp.Compare(a.CreateTime, b.CreateTime)
		}
	})
	resultMap["list"] = list
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
下载完成列表
@startIndexStr    string    首页查询传空字符串；分页查询使用上页最后一条数据的dbid
@limit            int       本次查询需要返回的数量
*/
func (a *SdkApi) Storage_Client_downloadFinishList(startIndexStr string, limit int) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	startIndex := base58.Decode(startIndexStr)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	//list := file_transfer.ManagerStatic.GetDownloadList()
	list, ERR := storage.StClient.GetDownloadFinishList(startIndex, uint64(limit))
	if ERR.CheckFail() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	//slices.SortFunc(list, func(a, b *file_transfer.DownloadStepVO) int {
	//	if a.CreateTime == b.CreateTime {
	//		return cmp.Compare(a.Name, b.Name)
	//	} else {
	//		return cmp.Compare(a.CreateTime, b.CreateTime)
	//	}
	//})
	resultMap["list"] = list
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
上传列表
*/
func (a *SdkApi) Storage_Client_uploadList() map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Msgf("获取上传列表")
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	//list := file_transfer.ManagerStatic.GetUploadList()
	list := storage.StClient.GetUploadFinishList()
	slices.SortFunc(list, func(a, b *file_transfer.FileTransferTaskVO) int {
		if a.CreateTime == b.CreateTime {
			return cmp.Compare(a.Name, b.Name)
		} else {
			return cmp.Compare(a.CreateTime, b.CreateTime)
		}
	})
	//utils.Log.Info().Msgf("上传列表 %+v", list)
	resultMap["list"] = list
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
修改文件或文件夹名称
@serverAddr          string    服务器地址
@dirAndFileIDStr     string    文件或文件夹id
@newName             string    新名称
*/
func (a *SdkApi) Storage_Client_UpdateDirAndFileName(serverAddr string, dirAndFileIDStr, newName string) map[string]interface{} {
	defer utils.PrintPanicStack(nil)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	if serverAddr == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "serverAddr"
		return resultMap
	}
	if dirAndFileIDStr == "" {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "dirAndFileIDStr"
		return resultMap
	}
	//utils.Log.Info().Msgf("要删除的文件和文件夹:%+v", dirAndFilesIDStr)
	var dirID []byte  // := make([][]byte, 0)
	var fileID []byte //files := make([][]byte, 0)
	have := false
	idOne := base58.Decode(dirAndFileIDStr)
	for _, dirOne := range a.dirIndex.Dirs {
		if bytes.Equal(dirOne.ID, idOne) {
			dirID = idOne
			have = true
			break
		}
	}
	//不在文件夹中，就在文件中
	if !have {
		for _, dirOne := range a.dirIndex.Files {
			if bytes.Equal(dirOne.Hash, idOne) {
				fileID = idOne
				have = true
				break
			}
		}
	}
	//都不存在
	if !have {
		resultMap["code"] = config.ERROR_CODE_Not_present
		resultMap["error"] = "dirAndFileIDStr"
		return resultMap
	}

	//utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirs, files)
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	ERR = storage.StClient.UpdateDirAndFileName(sAddr, dirID, fileID, newName)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR = utils.NewErrorSuccess()
	//resultMap["dir"] = dIndex.ConverVO()
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}
