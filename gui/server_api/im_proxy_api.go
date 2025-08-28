package server_api

import (
	"encoding/hex"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
查询本代理节点信息
*/
func (a *SdkApi) IMProxyServer_GetProxyInfo() map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	info, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if info == nil {
		info = model.CreateStorageServerInfo()
		ERR := db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), info)
		if !ERR.CheckSuccess() {
			//ERR := utils.NewErrorSysSelf(err)
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	sinfoVO := info.ConverVO()
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
	//ERR := utils.NewErrorSuccess()
	resultMap["info"] = info
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
打开或关闭代理服务器
*/
func (a *SdkApi) IMProxyServer_SetOpen(isOpen bool) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if sinfo == nil {
		sinfo = model.CreateStorageServerInfo()
	}
	sinfo.IsOpen = isOpen

	ERR = db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if isOpen {
		im.SetImProxyOpen()
		//storage.SetStorageServerOpen()
	} else {
		im.SetImProxyClose()
		//storage.SetStorageServerClose()
	}
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
添加存储目录
*/
func (a *SdkApi) IMProxyServer_AddDirectory(dir string) map[string]interface{} {
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

	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
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
	ERR = im.StaticProxyServerManager.DBC.AddDBone(dir)
	if !ERR.CheckSuccess() {
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	sinfo.Directory = append(sinfo.Directory, dir)
	ERR = db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), sinfo)
	if !ERR.CheckSuccess() {
		im.StaticProxyServerManager.DBC.DelDBone(dir)
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
func (a *SdkApi) IMProxyServer_DelDirectory(dir string) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
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
	ERR = db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), sinfo)
	if !ERR.CheckSuccess() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	err := im.StaticProxyServerManager.DBC.DelDBone(dir)
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
func (a *SdkApi) IMProxyServer_GetStorageServiceList() map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	storageServerList, ERR := im.StaticProxyClientManager.GetOrderServerList()
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	storageServerListVO := make([]*model.StorageServerInfoVO, 0)
	for _, one := range storageServerList {
		vo := one.ConverVO()
		selling, sold, _ := im.StaticProxyClientManager.QueryUserSpacesLocal(one.Addr)
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
获取离线服务节点列表
@price     uint64    月租
@open      bool      是否开启服务
*/
func (a *SdkApi) IMProxyClient_GetProxyList() map[string]interface{} {
	resultMap := make(map[string]interface{})
	imServers, ERR := db.ImProxyClient_GetProxyList(*im.Node.GetNetId())
	if ERR.CheckFail() {
		//resultMap["code"] = rpcmodel.Nomarl
		//resultMap["error"] = err.Error()
		//return resultMap
		//ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	imServerVO := make([]*model.StorageServerInfoVO, 0)
	for _, one := range imServers {
		imServerVO = append(imServerVO, one.ConverVO())
	}
	sort.SliceStable(imServerVO, func(i, j int) bool {
		return imServerVO[i].Count < imServerVO[j].Count
	})
	//resultMap["list"] = imServerVO
	//resultMap["code"] = rpcmodel.Success
	//return resultMap
	ERR = utils.NewErrorSuccess()
	resultMap["list"] = imServerVO
	resultMap["code"] = ERR.Code
	return resultMap
}

var ImProxy_server_SetSelling_Lock = new(sync.Mutex)

/*
设置本节点正在销售的存储空间
@selling    int64    要添加或减少的存储空间
*/
func (a *SdkApi) IMProxyServer_SetSelling(selling float64) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	Storage_server_SetSelling_Lock.Lock()
	defer Storage_server_SetSelling_Lock.Unlock()
	resultMap := make(map[string]interface{})
	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
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
	ERR = db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), sinfo)
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
设置本节点离线服务
设置本节点存储空间单价和空间限制
@priceUnit     uint64   //单价 单位：1G
@selling       uint64   //售卖总容量 单位：1G
@userFreelimit uint64   //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
@userCanTotal  uint64   //每个用户可以购买的空间总量 单位：1G
@useTimeMax    uint64   //每个订单租用时间最大值 单位：天
@renewalTime   uint64   //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
*/
func (a *SdkApi) IMProxyServer_SetProxyInfo(nickName string, priceUnit, selling, userFreelimit, userCanTotal, useTimeMax,
	renewalTime uint64) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	ImProxy_server_SetSelling_Lock.Lock()
	defer ImProxy_server_SetSelling_Lock.Unlock()
	resultMap := make(map[string]interface{})
	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*im.Node.GetNetId(), true)
	if ERR.CheckFail() {
		//ERR = utils.NewErrorSysSelf(err)
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
	ERR = db.ImProxyServer_SetProxyInfoSelf(*im.Node.GetNetId(), sinfo)
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
获取代理节点租用订单
@serverAddr    uint64     //代理节点地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func (a *SdkApi) ImProxyClient_GetOrder(serverAddr string, spaceTotal, useTime uint64) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	resultMap := make(map[string]interface{})
	orders, ERR := im.GetProxyOrders(sAddr, spaceTotal, useTime)
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
@serverAddr    uint64     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func (a *SdkApi) ImProxyClient_GetRenewalOrders(preNumber, serverAddr string, useTime uint64) map[string]interface{} {
	ERR := utils.NewErrorSuccess()
	sAddr := nodeStore.AddressFromB58String(serverAddr)
	resultMap := make(map[string]interface{})
	number, err := hex.DecodeString(preNumber)
	if err != nil {
		resultMap["code"] = config.ERROR_CODE_params_format
		resultMap["error"] = "preNumber"
		return resultMap
	}
	orders, ERR := im.GetProxyRenewalOrders(number, sAddr, useTime)
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
获取订单列表
*/
func (a *SdkApi) ImProxyClient_GetOrderList(startNumber16 string, limit uint64) map[string]interface{} {
	startNumber := []byte{}
	if startNumber16 != "" {
		var err error
		startNumber, err = hex.DecodeString(startNumber16)
		if err != nil {
			ERR := utils.NewErrorSysSelf(err)
			resultMap := make(map[string]interface{})
			resultMap["code"] = ERR.Code
			resultMap["error"] = ERR.Msg
			return resultMap
		}
	}
	if limit > 100 {
		limit = 100
	}
	if limit == 0 {
		limit = 100
	}
	orders := im.StaticProxyClientManager.GetOrdersListRange(startNumber, limit)
	ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	resultMap["list"] = orders
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}

/*
设置订单等待上链
*/
func (a *SdkApi) ImProxyClient_SetOrderWaitOnChain(orderId16 string, lockHeight uint64) map[string]interface{} {
	orderId, err := hex.DecodeString(orderId16)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		resultMap := make(map[string]interface{})
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	ERR := im.StaticProxyClientManager.SetOrderWaitOnChain(orderId, lockHeight)
	//ERR := utils.NewErrorSuccess()
	resultMap := make(map[string]interface{})
	//resultMap["list"] = orders
	resultMap["code"] = ERR.Code
	resultMap["error"] = ERR.Msg
	return resultMap
}
