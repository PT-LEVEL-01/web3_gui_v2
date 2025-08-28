package storage

import (
	"bytes"
	"cmp"
	"encoding/hex"
	"github.com/oklog/ulid/v2"
	"slices"
	"sync"
	"time"
	chainConfig "web3_gui/chain/config"
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type StorageServer struct {
	lock            *sync.RWMutex                   //
	AuthManager     *AuthManager                    //权限管理
	DBC             *DBCollections                  //数据库管理
	UnpaidOrders    map[string]*model.OrderForm     //未支付订单.key:string=用户地址;value:*model.OrderForm=订单;
	Orders          map[string]*model.OrderForm     //已支付订单
	RenewalOrder    map[string]*model.OrderForm     //续费订单
	UploadServer    map[string]*UploadServerManager //用户上传文件管理.key:string=用户地址;value:*UploadServerManager=上传管理;
	SellingLock     uint64                          //
	downloadManager *DownloadServerManager          //
	syncCount       *chain_plus.ChainSyncCount      //区块链同步统计程序
}

func CreateStorageServer() (*StorageServer, utils.ERROR) {
	dbc := CreateDBCollections()
	storageServerInfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if storageServerInfo != nil {
		ERR := dbc.AddDBone(storageServerInfo.Directory...)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
	}

	//
	auth := CreateAuthManager("")
	dm := NewDownloadServerManager(dbc)
	ss := StorageServer{
		lock:            new(sync.RWMutex), //
		AuthManager:     auth,
		DBC:             dbc,
		UnpaidOrders:    make(map[string]*model.OrderForm), //未支付订单
		Orders:          make(map[string]*model.OrderForm), //已支付订单
		RenewalOrder:    make(map[string]*model.OrderForm), //续费订单
		downloadManager: dm,                                //
	}
	err = ss.LoadOrdersAndInit()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := ss.LoadUserSpacesUse()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	go ss.LoopClean()
	return &ss, utils.NewErrorSuccess()
}

/*
查询一个用户的未支付订单
*/
func (this *StorageServer) FindUnpaidOrders(userAddr nodeStore.AddressNet) *model.OrderForm {
	//有未支付的订单，先支付
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(userAddr.GetAddr())]
	if !ok {
		return nil
	}
	return unpaidOrders
}

/*
创建订单
*/
func (this *StorageServer) CreateOrders(userAddr nodeStore.AddressNet, spaceTotal utils.Byte, useTime uint64) (*model.OrderForm, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()
	//查询一个用户的未支付订单
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(userAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorBus(config.ERROR_CODE_order_not_pay, "")
	}
	storageServerInfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}

	//用户空闲空间限制
	spaceTotalPay, _, remain := this.AuthManager.QueryUserSpaces(userAddr)
	if uint64(spaceTotal+remain/utils.GB) > storageServerInfo.UserFreelimit {
		remainGB := remain / utils.GB
		if remainGB%utils.GB > 0 {
			remainGB += 1
		}
		spaceTotal = utils.Byte(storageServerInfo.UserFreelimit) - remain/utils.GB
		//utils.Log.Info().Msgf("购买空间数量:%d %d %d %d", spaceTotal, spaceTotal, remain, storageServerInfo.UserFreelimit)
	}
	if spaceTotal == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_over_free_space_limit, "")
	}
	//检查用户可以购买的空间总量
	if (spaceTotalPay/utils.GB)+spaceTotal > utils.Byte(storageServerInfo.UserCanTotal) {
		spaceTotal = utils.Byte(storageServerInfo.UserCanTotal) - spaceTotalPay
		//utils.Log.Info().Msgf("购买空间数量:%d %d %d %d", spaceTotal, spaceTotalPay, spaceTotal, storageServerInfo.UserCanTotal)
	}
	if spaceTotal == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_over_pay_space_limit, "")
	}
	//不能超过最大时间
	if useTime > storageServerInfo.UseTimeMax {
		useTime = storageServerInfo.UseTimeMax
	}
	total := storageServerInfo.PriceUnit * uint64(spaceTotal) * useTime
	//number, err := GetGenID()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	addrSelf := Node.GetNetId()
	orders := model.OrderForm{
		Number:     ulid.Make().Bytes(), //订单编号，全局自增长ID
		UserAddr:   userAddr,            //消费者地址
		ServerAddr: *addrSelf,           //服务器地址
		SpaceTotal: uint64(spaceTotal),  //购买空间数量 单位：1G
		UseTime:    uint64(useTime),     //空间使用时间 单位：1天
		TotalPrice: uint64(total),       //订单总金额
		//ChainTx    []byte               //区块链上的交易
		//TxHash     []byte               //已经上链的交易hash
		CreateTime: time.Now().Unix(), //订单创建时间
		//TimeOut    int64                //订单过期时间
	}
	this.Orders[utils.Bytes2string(orders.Number)] = &orders
	if orders.TotalPrice > 0 {
		ERR := db.StorageServer_SaveOrderFormNotPay(&orders)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		this.SellingLock += orders.SpaceTotal
		return &orders, utils.NewErrorSuccess()
	}

	//免支付订单
	//过期时间是服务期
	orders.PayLockBlockHeight = (uint64(useTime) * 24 * 60 * 60) / uint64(chainConfig.Mining_block_time/time.Second)
	//orders.TimeOut = time.Now().Unix() + (int64(useTime) * 24 * 60 * 60)
	ERR := db.StorageServer_SaveOrderFormInUse(&orders)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	this.AuthManager.AddPurchaseSpace(orders.UserAddr, utils.Byte(orders.SpaceTotal)*utils.GB)
	InitUserTopDirIndex(orders.UserAddr)
	return &orders, utils.NewErrorSuccess()
}

/*
创建续费订单
@preNumber     []byte    续费的订单ID
@spaceTotal    uint64    要续费的空间数量，只能缩减，不能增加
@useTime       []byte    续费的时间
*/
func (this *StorageServer) CreateRenewalOrders(preNumber []byte, spaceTotal, useTime uint64) (*model.OrderForm, utils.ERROR) {
	this.lock.Lock()
	defer this.lock.Unlock()
	oldOrder, ok := this.Orders[utils.Bytes2string(preNumber)]
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_not_exist, "")
	}
	//查询一个用户的未支付订单
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(oldOrder.UserAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorSuccess()
	}
	storageServerInfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	//服务到期关闭
	if storageServerInfo.RenewalTime == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_Service_expiration_and_closure, "")
	}

	//判断订单到期
	pullHeight := this.syncCount.GetPullHeight()
	if oldOrder.PayLockBlockHeight <= pullHeight {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_overtime, "")
	}
	//未到续费时间
	if oldOrder.PayLockBlockHeight > pullHeight && (oldOrder.PayLockBlockHeight-pullHeight) >
		uint64(storageServerInfo.RenewalTime*24*60*60/uint64(chainConfig.Mining_block_time/time.Second)) {
		//未到续费时间
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_not_overtime, "") // config.ERROR_storage_orders_not_overtime
	}

	//now := time.Now().Unix()
	//if oldOrder.TimeOut <= now {
	//	//订单到期，不能续费
	//
	//	return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_overtime, "")
	//}
	//if oldOrder.TimeOut > now && (oldOrder.TimeOut-now) > int64(storageServerInfo.RenewalTime)*24*60*60 {
	//	//订单未到续费时间
	//	return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_not_overtime, "") // config.ERROR_storage_orders_not_overtime
	//}
	//不能超过最大时间
	if useTime > storageServerInfo.UseTimeMax {
		useTime = storageServerInfo.UseTimeMax
	}
	//续费空间只能减少，不能增加
	if spaceTotal > oldOrder.SpaceTotal {
		//return nil, config.ERROR_storage_user_renewal_space_total_max
		spaceTotal = oldOrder.SpaceTotal
	}

	total := storageServerInfo.PriceUnit * spaceTotal * useTime
	//newNumber, err := GetGenID()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	addrSelf := Node.GetNetId()
	orders := model.OrderForm{
		Number:     ulid.Make().Bytes(), //订单编号，全局自增长ID
		PreNumber:  preNumber,           //
		UserAddr:   oldOrder.UserAddr,   //消费者地址
		ServerAddr: *addrSelf,           //服务器地址
		SpaceTotal: spaceTotal,          //购买空间数量 单位：1G
		UseTime:    useTime,             //空间使用时间 单位：1天
		TotalPrice: total,               //订单总金额
		//ChainTx    []byte               //区块链上的交易
		//TxHash     []byte               //已经上链的交易hash
		CreateTime: time.Now().Unix(), //订单创建时间
		//TimeOut    int64                //订单过期时间
	}
	this.Orders[utils.Bytes2string(orders.Number)] = &orders
	if orders.TotalPrice > 0 {
		ERR := db.StorageServer_SaveOrderFormNotPay(&orders)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		return &orders, utils.NewErrorSysSelf(err)
	}
	//免支付订单
	//过期时间是服务期
	orders.PayLockBlockHeight = (uint64(useTime) * 24 * 60 * 60) / uint64(chainConfig.Mining_block_time/time.Second)
	//orders.TimeOut = time.Now().Unix() + (int64(useTime) * 3 * 60)
	ERR := db.StorageServer_SaveOrderFormInUse(&orders)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	this.AuthManager.AddPurchaseSpace(orders.UserAddr, utils.Byte(orders.SpaceTotal)*utils.GB)
	InitUserTopDirIndex(orders.UserAddr)
	return &orders, utils.NewErrorSysSelf(err)
}

/*
定时清理过期的订单
包括过期未支付的订单、存储时间过期未续费的订单
*/
func (this *StorageServer) LoopClean() {
	var err error
	var ERR utils.ERROR
	for range time.NewTicker(config.Clean_Interval).C {
		//utils.Log.Info().Msgf("定期检查服务器订单")
		//now := time.Now().Unix()
		//定期检查未支付订单

		//定期检查已支付订单服务期是否超时
		this.lock.Lock()
		for orderKey, one := range this.Orders {
			//未超时
			pullHeight := uint64(0)
			if one.PayLockBlockHeight > pullHeight {
				continue
			}
			//if one.TimeOut > now {
			//	continue
			//}
			//超过服务时间，查看是否有未支付订单，有则等一等
			have := false
			for _, uOone := range this.UnpaidOrders {
				if bytes.Equal(uOone.PreNumber, one.Number) {
					have = true
					break
				}
			}
			if have {
				continue
			}

			var renewalOrder *model.OrderForm
			//检查是否有续费订单
			for _, orderOne := range this.RenewalOrder {
				if bytes.Equal(orderOne.PreNumber, one.Number) {
					renewalOrder = orderOne
					break
				}
			}

			utils.Log.Info().Msgf("删除服务器过期订单:%+v", one)
			ERR = db.StorageServer_MoveOrderToInUseTimeout(one, renewalOrder)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("删除超期订单失败:%s", err.Error())
				continue
			}
			//删除过期订单
			delete(this.Orders, orderKey)
			//没有续费，直接删除
			if renewalOrder == nil {
				this.AuthManager.SubPurchaseSpace(one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
				continue
			}
			//删除续费订单
			delete(this.RenewalOrder, utils.Bytes2string(renewalOrder.Number))
			//把续费订单放入正在服务的订单列表
			this.Orders[utils.Bytes2string(renewalOrder.Number)] = renewalOrder
			//有续费，判断空间是否减少
			if renewalOrder.SpaceTotal < one.SpaceTotal {
				count := one.SpaceTotal - renewalOrder.SpaceTotal
				this.AuthManager.SubPurchaseSpace(one.UserAddr, utils.Byte(count)*utils.GB)
			}
		}
		this.lock.Unlock()
		this.cleanUserFiles()
	}
}

/*
清理用户上传的文件，以达到空间要求
按照上传时间顺序，清理最新上传的文件
*/
func (this *StorageServer) cleanUserFiles() utils.ERROR {
	//utils.Log.Info().Msgf("开始清理使用空间不够的用户文件")
	//查询用户购买了多少空间
	userList := this.AuthManager.QueryUserList()
	for _, one := range userList {
		spacesTotal, spacesUse, _ := this.AuthManager.QueryUserSpaces(one)
		//utils.Log.Info().Msgf("检查这一用户空间:%s %d", one.B58String(), spacesTotal)
		if spacesTotal == 0 {
			utils.Log.Info().Msgf("清除所有数据:%s", one.B58String())
			//删除全部文件
			_, delFileIndexs, ERR := db.StorageServer_DelUserDirAndFileAll(one)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("删除用户全部文件错误:%s", ERR.String())
				continue
			}
			//删除数据
			for _, fileOne := range delFileIndexs {
				for _, one := range fileOne.Chunks {
					utils.Log.Info().Msgf("删除块数据:%s", hex.EncodeToString(one))
					this.DBC.DelChunk(one)
				}
			}
			//删除用户
			this.AuthManager.DelUser(one)
		} else if spacesTotal < spacesUse {
			utils.Log.Info().Msgf("清除部分文件:%s", one.B58String())
			//清理部分文件
			dir, ERR := db.StorageServer_GetUserTopDirIndex(one)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户的顶层目录错误:%s", ERR.String())
				continue
			}
			//迭代查询文件夹中的文件
			_, files, _, _, ERR := db.StorageServer_GetDirIndexRecursion([][]byte{dir.ID})
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户的顶层目录错误:%s", ERR.String())
				continue
			}
			//按照上传时间排序，优先删除最新上传的文件
			newFiles := make([]*model.FileIndex, 0, len(files))
			for _, fileOne := range files {
				fileOne.FilterUser(one)
				newFiles = append(newFiles, fileOne)
			}
			slices.SortFunc(newFiles, func(a, b *model.FileIndex) int {
				return cmp.Compare(a.Time[0], b.Time[0])
			})
			//计算释放的空间
			sizeTotal := uint64(0)
			//要删除的文件
			delFileIDs := make([][]byte, 0)
			for _, fileOne := range newFiles {
				//utils.Log.Info().Msgf("")
				sizeTotal += fileOne.FileSize
				delFileIDs = append(delFileIDs, fileOne.ID)
				if spacesTotal >= spacesUse-utils.Byte(sizeTotal) {
					break
				}
			}
			//删除用户文件
			_, delFileIndexs, ERR := db.StorageServer_DelDirAndFileIndex(one, nil, delFileIDs)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("删除用户全部文件错误:%s", ERR.String())
				continue
			}
			//删除数据
			for _, fileOne := range delFileIndexs {
				for _, one := range fileOne.Chunks {
					this.DBC.DelChunk(one)
				}
			}
			this.AuthManager.SubUseSpace(one, utils.Byte(sizeTotal))
		}
	}
	return utils.NewErrorSuccess()
}

/*
加载数据库订单并且初始化对象
*/
func (this *StorageServer) LoadOrdersAndInit() error {
	orders, err := db.StorageServer_GetStorageServerInUseOrdersIDs()
	if err != nil {
		return err
	}
	for _, one := range orders {
		this.AuthManager.AddPurchaseSpace(one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
	}
	return nil
}

/*
加载数据库中所有用户已经使用的存储空间
*/
func (this *StorageServer) LoadUserSpacesUse() utils.ERROR {
	m, ERR := db.StorageServer_GetAllUserSpaceUseSize()
	if !ERR.CheckSuccess() {
		return ERR
	}
	for k, v := range m {
		addr := nodeStore.NewAddressNet([]byte(k)) //nodeStore.AddressNet([]byte(k))
		this.AuthManager.AddUseSpace(*addr, utils.Byte(v))
	}
	return utils.NewErrorSuccess()
}

/*
初始化用户顶层目录
*/
func InitUserTopDirIndex(userAddr nodeStore.AddressNet) utils.ERROR {
	dirIndex, ERR := db.StorageServer_GetUserTopDirIndex(userAddr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if dirIndex != nil {
		return utils.NewErrorSuccess()
	}
	//id, err := GetGenID()
	//if err != nil {
	//	return err
	//}
	dirIndex = &model.DirectoryIndex{
		ID: ulid.Make().Bytes(), //文件夹唯一ID采用全局自增长ID
		//Name    string           //文件夹名称
		//Dirs    []DirectoryIndex //文件夹中包含文件夹
		//Files   []FileIndex      //文件列表
		//DirsID  [][]byte         //包含文件夹的ID
		//FilesID [][]byte         //包含文件的ID
		UAddr: userAddr, //
	}
	return db.StorageServer_SaveUserTopDirIndex(userAddr, *dirIndex)
}

/*
添加上传文件索引
*/
func (this *StorageServer) UploadAddFileIndex(uAddr nodeStore.AddressNet, fileIndex model.FileIndex) utils.ERROR {
	//uAddr := fileIndex.UserAddr[0]
	//utils.Log.Info().Msgf("上传的文件索引:%+v", fileIndex)
	this.lock.Lock()
	defer this.lock.Unlock()

	_, useSpacess, _ := this.AuthManager.QueryUserSpaces(uAddr)
	utils.Log.Info().Msgf("本次使用空间:%d 总共使用空间:%d", fileIndex.FileSize, useSpacess)
	//查询这个用户的传输管理器是否存在
	us, ok := this.UploadServer[utils.Bytes2string(fileIndex.UserAddr[0])]
	if !ok {
		us = NewUploadServerManager(this.DBC)
	}
	ERR := utils.NewErrorSuccess()
	//查询索引是否存在
	fileIndexs, ERR := db.StorageServer_GetFileIndexMore(fileIndex.Hash)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//文件已经存在，则把用户添加到已经存在的列表中
	if len(fileIndexs) == 1 {
		utils.Log.Info().Msgf("文件已经存在:%+v %+v", fileIndexs[0], uAddr)
		if fileIndexs[0].CheckHaveUser(uAddr) {
			utils.Log.Info().Msgf("用户已经拥有此文件")
			//用户已经拥有此文件
			return utils.NewErrorBus(config.ERROR_CODE_exist, "")
		}
		//把用户添加到列表中
		fileIndexs[0].UserAddr = append(fileIndexs[0].UserAddr, uAddr.GetAddr())
		//把加密后的文件密码保存到列表中
		fileIndexs[0].Pwds = append(fileIndexs[0].Pwds, fileIndex.Pwds[0])
		fileIndexs[0].Name = append(fileIndexs[0].Name, fileIndex.Name[0])
		fileIndexs[0].DirID = append(fileIndexs[0].DirID, fileIndex.DirID[0])
		fileIndexs[0].PermissionType = append(fileIndexs[0].PermissionType, fileIndex.PermissionType[0])
		fileIndex = *fileIndexs[0]
		ERR = utils.NewErrorBus(config.ERROR_CODE_storage_No_need_upload_files, "")
	}
	//权限验证，空间大小验证
	if !this.AuthManager.AddUseSpace(uAddr, utils.Byte(fileIndex.FileSize)) {
		return utils.NewErrorBus(config.ERROR_CODE_storage_Insufficient_user_space, "")
	}
	//处理文件重复的情况
	if ERR.Code == config.ERROR_CODE_storage_No_need_upload_files {
		utils.Log.Info().Msgf("文件重复:%+v", fileIndex)
		ERR := db.StorageServer_SaveFileIndexToDir(uAddr, fileIndex.DirID[len(fileIndex.DirID)-1], fileIndex)
		//文件保存到对应文件夹
		if !ERR.CheckSuccess() {
			return ERR
		}
		dir, ERR := db.StorageServer_GetUserTopDirIndex(uAddr)
		utils.Log.Info().Msgf("用户顶层目录:%+v", dir)
		return ERR
	}

	//保存到数据库
	ERR = db.StorageServer_SaveFileIndex(uAddr, fileIndex)
	if !ERR.CheckSuccess() {
		return ERR
	}
	us.StartTransfer(uAddr, fileIndex)
	return utils.NewErrorSuccess()
}

/*
暂停上传文件
*/
func (this *StorageServer) UploadStop(uAddr nodeStore.AddressNet, fileID []byte) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	us, ok := this.UploadServer[utils.Bytes2string(uAddr.GetAddr())]
	if !ok {
		return nil
	}
	us.StopUpload(fileID)
	return nil
}

/*
暂停后恢复
*/
func (this *StorageServer) UploadReset(uAddr nodeStore.AddressNet, fileID []byte) error {
	return nil
}

/*
删除文件和文件夹
*/
func (this *StorageServer) DelDirAndFile(uAddr nodeStore.AddressNet, dirIDs, fileIDs [][]byte) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	utils.Log.Info().Msgf("删除文件和文件夹列表:%+v %+v", dirIDs, fileIDs)

	updateFileIndexs, delFileIndexs, ERR := db.StorageServer_DelDirAndFileIndex(uAddr, dirIDs, fileIDs)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//删除数据
	for _, fileOne := range delFileIndexs {
		for _, one := range fileOne.Chunks {
			this.DBC.DelChunk(one)
		}
	}

	//计算释放的空间
	sizeTotal := uint64(0)
	for _, one := range append(updateFileIndexs, delFileIndexs...) {
		sizeTotal += one.FileSize
	}
	_, useSpacess, _ := this.AuthManager.QueryUserSpaces(uAddr)
	utils.Log.Info().Msgf("减少使用空间:%d 总共使用空间:%d", sizeTotal, useSpacess)
	this.AuthManager.SubUseSpace(uAddr, utils.Byte(sizeTotal))
	return ERR
}

/*
创建目录
*/
func (this *StorageServer) CreateDir(dirIndex model.DirectoryIndex) utils.ERROR {
	//id, err := GetGenID()
	//if err != nil {
	//	return err
	//}
	dirIndex.ID = ulid.Make().Bytes()
	return db.StorageServer_SaveUserDirIndex(dirIndex)
}

/*
下载文件
*/
func (this *StorageServer) DownloadFile(uAddr nodeStore.AddressNet, fileIndex *model.FileIndex) error {
	this.downloadManager.AddDownloadTask(uAddr, fileIndex)
	return nil
}

/*
修改文件名称
*/
func (this *StorageServer) UpdateFileName(uAddr nodeStore.AddressNet, fileId []byte, newName string) utils.ERROR {
	return db.StorageServer_UpdateFileName(uAddr, fileId, newName)
}

/*
修改文件夹名称
*/
func (this *StorageServer) UpdateDirName(uAddr nodeStore.AddressNet, dirId []byte, newName string) utils.ERROR {
	return db.StorageServer_UpdateDirName(uAddr, dirId, newName)
}
