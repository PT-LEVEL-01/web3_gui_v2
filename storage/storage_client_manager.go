package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type StorageClient struct {
	lock        *sync.RWMutex           //
	AuthManager map[string]*AuthManager //权限管理。key:string=服务器地址;value:*AuthManager=权限管理;
	//UnpaidOrders []*model.OrderForm        //未支付订单
	//Orders       []*model.OrderForm        //已支付订单
	//RenewalOrder []*model.OrderForm        //续费订单
	UnpaidOrders map[string]*model.OrderForm //未支付订单.key:string=订单编号;value:*model.OrderForm=订单;
	Orders       map[string]*model.OrderForm //已支付订单
	RenewalOrder map[string]*model.OrderForm //续费订单
	Upload       map[string]*UploadManager   //上传文件任务管理。key:string=服务器地址;value:*UploadManager=上传文件管理;
	download     *ClientDownload             //下载的任务管理
}

func CreateStorageClient() (*StorageClient, utils.ERROR) {
	ss := StorageClient{
		lock:        new(sync.RWMutex),             //
		AuthManager: make(map[string]*AuthManager), //
		//UnpaidOrders: make([]*model.OrderForm, 0),     //未支付订单
		//Orders:       make([]*model.OrderForm, 0),     //已支付订单
		//RenewalOrder: make([]*model.OrderForm, 0),     //续费订单
		UnpaidOrders: make(map[string]*model.OrderForm), //未支付订单
		Orders:       make(map[string]*model.OrderForm), //已支付订单
		RenewalOrder: make(map[string]*model.OrderForm), //续费订单
		Upload:       make(map[string]*UploadManager),   //
		download:     NewClientDownload(),               //
	}
	//加载客户端订单
	ERR := ss.LoadStorageClientOriders()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//从数据库加载上传文件列表
	err := ss.loadUploadFileList()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return &ss, utils.NewErrorSuccess()
}

/*
获取订单中的云存储服务器列表
*/
func (this *StorageClient) GetOrderServerList() ([]model.StorageServerInfo, error) {
	addrs := make([]nodeStore.AddressNet, 0)
	this.lock.RLock()
	for keyStr, _ := range this.AuthManager {
		addrs = append(addrs, *nodeStore.NewAddressNet([]byte(keyStr)))
	}
	this.lock.RUnlock()
	list, err := db.StorageClient_GetStorageServerListByAddrs(addrs...)
	if err != nil {
		return nil, err
	}
	return list, nil
}

/*
本地查询用户在一个提供商的容量
*/
func (this *StorageClient) QueryUserSpacesLocal(addr nodeStore.AddressNet) (utils.Byte, utils.Byte, utils.Byte) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if value, ok := this.AuthManager[utils.Bytes2string(addr.GetAddr())]; ok {
		return value.QueryUserSpaces(addr)
	}
	return 0, 0, 0
}

/*
获取活跃的订单列表
*/
func (this *StorageClient) GetOrdersList() []*model.OrderFormVO {
	this.lock.RLock()
	defer this.lock.RUnlock()
	orders := make([]*model.OrderFormVO, 0, len(this.Orders))
	for _, one := range this.Orders {
		orders = append(orders, one.ConverVO())
	}
	return orders
}

/*
添加未支付订单
*/
func (this *StorageClient) AddOrdersUnpaid(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.UnpaidOrders[utils.Bytes2string(order.Number)] = order
	//this.UnpaidOrders = append(this.UnpaidOrders, order)
	return utils.NewErrorSuccess()
}

/*
添加已支付订单
当支付费用为0时，直接添加到已支付账单列表
*/
func (this *StorageClient) AddOrders(order *model.OrderForm) utils.ERROR {
	//utils.Log.Info()
	this.lock.Lock()
	defer this.lock.Unlock()
	ERR := db.StorageClient_SaveOrderFormInUse(order)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//添加到已经支付订单列表
	this.Orders[utils.Bytes2string(order.Number)] = order
	//this.Orders = append(this.Orders, order)
	auth, ok := this.AuthManager[utils.Bytes2string(order.ServerAddr.GetAddr())]
	if !ok {
		auth = CreateAuthManager(utils.Bytes2string(order.ServerAddr.GetAddr()))
		this.AuthManager[utils.Bytes2string(order.ServerAddr.GetAddr())] = auth
		this.Upload[utils.Bytes2string(order.ServerAddr.GetAddr())] = NewUploadManager()
	}
	auth.AddPurchaseSpace(order.ServerAddr, utils.Byte(order.SpaceTotal)*utils.GB)
	return utils.NewErrorSuccess()
}

/*
移动未支付订单到已支付订单中
*/
func (this *StorageClient) MoveOrders(order *model.OrderForm) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	return nil
}

/*
添加已支付续费订单
当支付费用为0时，直接添加到已支付账单列表
*/
func (this *StorageClient) AddRenewalOrders(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	ERR := db.StorageClient_SaveOrderFormInUse(order)
	if !ERR.CheckSuccess() {
		return ERR
	}
	this.RenewalOrder[utils.Bytes2string(order.Number)] = order
	//this.RenewalOrder = append(this.RenewalOrder, order)
	auth, ok := this.AuthManager[utils.Bytes2string(order.ServerAddr.GetAddr())]
	if !ok {
		auth = CreateAuthManager(utils.Bytes2string(order.ServerAddr.GetAddr()))
	}
	//找到续费关联订单
	var oldOrder *model.OrderForm
	for _, one := range this.Orders {
		if bytes.Equal(one.ServerAddr.GetAddr(), order.ServerAddr.GetAddr()) && bytes.Equal(one.Number, order.PreNumber) {
			oldOrder = one
			break
		}
	}
	if oldOrder != nil {
		//添加到已经支付的续费订单列表
		this.RenewalOrder[utils.Bytes2string(order.Number)] = order
		//this.RenewalOrder = append(this.RenewalOrder, order)
	} else {
		//添加到已经支付的订单列表
		this.Orders[utils.Bytes2string(order.Number)] = order
		//this.Orders = append(this.Orders, order)
		////判断空间是否减少
		//count := uint64(0)
		//if order.SpaceTotal < oldOrder.SpaceTotal {
		//	count = oldOrder.SpaceTotal - order.SpaceTotal
		//}
		auth.AddPurchaseSpace(order.ServerAddr, utils.Byte(order.SpaceTotal)*utils.GB)
	}
	return utils.NewErrorSuccess()
}

/*
移动未支付的续费订单到已支付续费订单中
*/
func (this *StorageClient) MoveRenewalOrders(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()

	ERR := db.StorageClient_SaveOrderFormInUse(order)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//找到并删除未支付订单
	delete(this.UnpaidOrders, utils.Bytes2string(order.Number))
	//for i, one := range this.UnpaidOrders {
	//	if bytes.Equal(one.Number, order.Number) && bytes.Equal(one.ServerAddr, order.ServerAddr) {
	//		temp := this.UnpaidOrders[:i]
	//		temp = append(temp, this.UnpaidOrders[:i+1]...)
	//		this.UnpaidOrders = temp
	//		break
	//	}
	//}
	//添加到已经支付订单列表
	this.Orders[utils.Bytes2string(order.Number)] = order
	//this.Orders = append(this.Orders, order)
	auth, ok := this.AuthManager[utils.Bytes2string(order.ServerAddr.GetAddr())]
	if ok {
		auth.AddPurchaseSpace(order.ServerAddr, utils.Byte(order.SpaceTotal)*utils.GB)
	}
	return utils.NewErrorSuccess()
}

/*
上传文件
*/
func (this *StorageClient) UploadFile(serverAddr nodeStore.AddressNet, dirID []byte, files ...string) utils.ERROR {
	//检查这些文件大小，并询问存储空间是否足够
	sizeTotal := int64(0)
	for _, filePathOne := range files {
		fileInfo, err := os.Stat(filePathOne)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		sizeTotal += fileInfo.Size()
	}
	if sizeTotal == 0 {
		return utils.NewErrorSuccess()
	}
	spacesTotal, spacesUse, ERR := GetFreeSpace(serverAddr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	utils.Log.Info().Msgf("空间总量:%d 已使用空间:%d 剩余空间:%d", spacesTotal, spacesUse, spacesTotal-spacesUse)
	if spacesTotal < spacesUse+uint64(sizeTotal) {
		return utils.NewErrorBus(config.ERROR_CODE_storage_Insufficient_user_space, "")
	}

	have := false
	this.lock.RLock()
	upload, ok := this.Upload[utils.Bytes2string(serverAddr.GetAddr())]
	if ok {
		for _, one := range files {
			upload.AddFile(serverAddr, dirID, one)
		}
		have = true
	}
	this.lock.RUnlock()
	if have {
		return utils.NewErrorSuccess()
	}
	//当没有这个对象时，创建一个新的
	this.lock.Lock()
	upload, ok = this.Upload[utils.Bytes2string(serverAddr.GetAddr())]
	if !ok {
		upload = NewUploadManager()
		this.Upload[utils.Bytes2string(serverAddr.GetAddr())] = upload
	}
	for _, one := range files {
		upload.AddFile(serverAddr, dirID, one)
	}
	have = true
	this.lock.Unlock()
	return utils.NewErrorSuccess()
}

/*
文件传输完成
*/
//func (this *StorageClient) UpdateFileUploadFinish(serverAddr nodeStore.AddressNet, fileID []byte) error {
//	this.lock.RUnlock()
//	defer this.lock.RUnlock()
//	um, ok := this.Upload[utils.Bytes2string(serverAddr)]
//	if !ok {
//		return nil
//	}
//	um.UpdateFileStatusUploadFinish(fileID)
//	return nil
//}

/*
文件保存数据库完成
*/
func (this *StorageClient) UpdateFileSavedbFinish(serverAddr nodeStore.AddressNet, fileID []byte) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	um, ok := this.Upload[utils.Bytes2string(serverAddr.GetAddr())]
	if !ok {
		return nil
	}
	um.UpdateFileStatusSaveDBFinish(fileID)
	return nil
}

/*
加载客户端订单
*/
func (this *StorageClient) LoadStorageClientOriders() utils.ERROR {
	ordersInUse, err := db.StorageClient_LoadOrderFormInUse()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	ordersNotPay, err := db.StorageClient_LoadOrderFormNotPay()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	for _, one := range ordersInUse {
		if one.PreNumber == nil || len(one.PreNumber) == 0 {
			ERR := this.AddOrders(&one)
			if !ERR.CheckSuccess() {
				return ERR
			}
		} else {
			ERR := this.AddRenewalOrders(&one)
			if !ERR.CheckSuccess() {
				return ERR
			}
		}
	}
	for _, one := range ordersNotPay {
		ERR := this.AddOrdersUnpaid(&one)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	return utils.NewErrorSuccess()
}

/*
从数据库加载上传文件列表
*/
func (this *StorageClient) loadUploadFileList() error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return nil
}

/*
创建新文件夹
*/
func (this *StorageClient) CreateNewDir(serverAddr nodeStore.AddressNet, parentDirID []byte, newDirName string) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return SendCreateNewDir(serverAddr, parentDirID, newDirName)
}

/*
删除多个文件和文件夹
*/
func (this *StorageClient) DelDirAndFiles(serverAddr nodeStore.AddressNet, dirIDs, fileIDs [][]byte) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return SendDelDirAndFile(serverAddr, dirIDs, fileIDs)
}

/*
下载多个文件和文件夹
*/
func (this *StorageClient) DownloadDirAndFiles(serverAddr nodeStore.AddressNet, dirIndexs []*model.DirectoryIndex,
	fileIndexs []*model.FileIndex, localPath string) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()

	//utils.Log.Info().Msgf("要下载的文件夹:%+v", dirIndexs)
	//去远端迭代查询文件夹中的子文件
	dirIDs := make([][]byte, 0)
	for _, one := range dirIndexs {
		dirIDs = append(dirIDs, one.ID)
	}
	dirIndexs, ERR := ParseFilesWithin(serverAddr, dirIDs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	utils.Log.Info().Msgf("目录:%+v", dirIndexs)
	//解析为下载任务
	downloads := make([]DownloadStep, 0)
	for _, one := range fileIndexs {
		//过滤密码
		for i, two := range one.UserAddr {
			if bytes.Equal(two, Node.GetNetId().GetAddr()) {
				one.UserAddr = [][]byte{two}
				one.Pwds = [][]byte{one.Pwds[i]}
				break
			}
		}
		downloadOne := NewDownloadStep(serverAddr, one, localPath)
		downloads = append(downloads, *downloadOne)
	}
	for {
		tempDirs := make([]*model.DirectoryIndex, 0)
		for _, one := range dirIndexs {
			utils.Log.Info().Msgf("本文件夹名称:%s %p", one.Name, one.ParentDir)
			tempDirPath := one.Name
			//构建文件存放路径
			for tempDir := one.ParentDir; tempDir != nil; tempDir = tempDir.ParentDir {
				tempDirPath = filepath.Join(tempDir.Name, tempDirPath)
				utils.Log.Info().Msgf("loop本文件夹名称:%s %s", tempDir.Name, tempDirPath)
			}
			tempDirPath = filepath.Join(localPath, tempDirPath)
			//utils.Log.Info().Msgf("文件夹路径:%s", tempDirPath)
			for _, fileOne := range one.Files {
				//过滤密码
				for i, two := range fileOne.UserAddr {
					if bytes.Equal(two, Node.GetNetId().GetAddr()) {
						fileOne.UserAddr = [][]byte{two}
						fileOne.Pwds = [][]byte{fileOne.Pwds[i]}
						break
					}
				}
				downloadOne := NewDownloadStep(serverAddr, fileOne, tempDirPath)
				utils.Log.Info().Msgf("构建下载任务:%+v", downloadOne)
				downloads = append(downloads, *downloadOne)
			}
			tempDirs = append(tempDirs, one.Dirs...)
		}
		if len(tempDirs) == 0 {
			break
		}
		dirIndexs = tempDirs
	}
	//保存记录到数据库
	ids, ERR := StorageClient_SaveDownloadTask(downloads)
	if ERR.CheckFail() {
		utils.Log.Error().Str("保存下载记录错误 ERR", ERR.String()).Send()
		return ERR
	}
	for i, one := range downloads {
		one.dbId = ids[i]
	}
	this.download.AddDownloadTask(downloads)
	return utils.NewErrorSuccess()
}

/*
获取正在下载列表
*/
func (this *StorageClient) GetDownloadList() []*file_transfer.DownloadStepVO {
	return this.download.GetDownloadList()
}

/*
获取正在上传列表
*/
func (this *StorageClient) GetUploadList() []*file_transfer.FileTransferTaskVO {
	fts := make([]*file_transfer.FileTransferTaskVO, 0)
	for _, one := range this.Upload {
		fts = append(fts, one.GetUploadFileList()...)
	}
	return fts
}

/*
获取下载完成列表
*/
func (this *StorageClient) GetDownloadFinishList(startIndex []byte, limit uint64) ([]*file_transfer.DownloadStepVO, utils.ERROR) {
	return this.download.GetDownloadFinishList(startIndex, limit)
}

/*
获取上传完成列表
*/
func (this *StorageClient) GetUploadFinishList() []*file_transfer.FileTransferTaskVO {
	fts := make([]*file_transfer.FileTransferTaskVO, 0)
	for _, one := range this.Upload {
		fts = append(fts, one.GetUploadFileList()...)
	}
	return fts
}

/*
修改文件或文件夹名称
*/
func (this *StorageClient) UpdateDirAndFileName(serverAddr nodeStore.AddressNet, dirID, fileID []byte, newName string) utils.ERROR {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return SendUpdateDirAndFile(serverAddr, dirID, fileID, newName)
}
