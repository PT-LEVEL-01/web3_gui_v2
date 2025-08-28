package storage

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/cake/transfer_manager"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type UploadServerManager struct {
	lock  *sync.RWMutex                //
	files map[string]*UploadServerStep //
	tag   chan bool                    //
	DBC   *DBCollections               //数据库管理
}

func NewUploadServerManager(dbc *DBCollections) *UploadServerManager {
	um := &UploadServerManager{
		lock:  new(sync.RWMutex),                  //
		files: make(map[string]*UploadServerStep), //
		tag:   make(chan bool, 1),                 //
		DBC:   dbc,
	}
	um.tag <- false
	//go um.Start()
	return um
}

/*
添加一个文件索引
*/
//func (this *UploadServerManager) AddFileIndex(fileIndex *model.FileIndex) {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	fileIndex.Status = UploadFile_server_status_start
//	this.files[utils.Bytes2string(fileIndex.ID)] = fileIndex
//}

/*
添加一个文件索引
*/
func (this *UploadServerManager) StartTransfer(uAddr nodeStore.AddressNet, fileIndex model.FileIndex) {
	//utils.Log.Info().Msgf("添加一个文件索引")
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.files[utils.Bytes2string(fileIndex.Hash)]
	if ok {
		return
	}
	uss := NewUploadServerStep(uAddr, &fileIndex, this.DBC)
	uss.Start()
}

/*
暂停下载
*/
func (this *UploadServerManager) StopUpload(fileID ...[]byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	//for _, one := range fileID {
	//	fileIndex, ok := this.files[utils.Bytes2string(one)]
	//	if !ok {
	//		continue
	//	}
	//	ok = atomic.CompareAndSwapUint32(&fileIndex.Status, UploadFile_server_status_start, UploadFile_server_status_transfer)
	//	if !ok {
	//		continue
	//	}
	//	select {
	//	case this.tag <- false:
	//	default:
	//	}
	//}
}

/*
上传文件服务器处理步骤
*/
type UploadServerStep struct {
	uAddr     nodeStore.AddressNet //
	fileIndex *model.FileIndex     //
	DBC       *DBCollections       //数据库管理
}

func NewUploadServerStep(uAddr nodeStore.AddressNet, fileIndex *model.FileIndex, dbc *DBCollections) *UploadServerStep {
	uss := UploadServerStep{
		uAddr:     uAddr,
		fileIndex: fileIndex,
		DBC:       dbc,
	}
	return &uss
}

/*
开始按步骤处理
*/
func (this *UploadServerStep) Start() {
	signalID, signalChan := file_transfer.ManagerStatic.GetListeningPullFinishSignal()
	//signalID, signalChan := transfer_manager.TransferMangerStatic.GetListeningPullFinishSignal()
	go func() {
		go func() {
			this.CleanChunks()
		}()
		utils.Log.Info().Msgf("等待新文件传输")
		//等待传输完成
		ERR := WaitDownloadFinish(signalID, signalChan, *this.fileIndex)
		if !ERR.CheckSuccess() {
			return
		}
		utils.Log.Info().Msgf("等待新文件保存到数据库")
		//保存到数据库
		ERR = this.saveDB()
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("保存到数据库失败:%s", ERR.String())
			return
		}
		ERR.Msg = string(this.fileIndex.ID)
		utils.Log.Info().Msgf("通知客户端已经保存到数据库:%s", hex.EncodeToString(this.fileIndex.ID))
		//通知客户端已经保存到数据库
		ERR = this.NotifyFinish(ERR)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("通知客户端已经保存到数据库:%s", ERR.String())
			return
		}
		utils.Log.Info().Msgf("上传文件完成")
	}()
}

/*
把分片保存到数据库
*/
func (this *UploadServerStep) saveDB() utils.ERROR {
	//块文件保存到数据库
	for _, one := range this.fileIndex.Chunks {
		bs, err := os.ReadFile(filepath.Join(file_transfer.FILEPATH_Recfilepath, hex.EncodeToString(one)))
		//bs, err := os.ReadFile(filepath.Join(transfer_manager.Recfilepath, hex.EncodeToString(one)))
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		ERR := this.DBC.SaveChunk(one, bs)
		//utils.Log.Info().Msgf("保存到数据库:%s %s", hex.EncodeToString(one), ERR.String())
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//utils.Log.Info().Msgf("开始移动文件到对应文件夹:%+v", this.fileIndex)
	//文件保存到对应文件夹
	ERR := db.StorageServer_MoveFileIndex(this.uAddr, nil, this.fileIndex.DirID[0], this.fileIndex.Hash)
	return ERR
}

/*
通知客户端保存到数据库完成
*/
func (this *UploadServerStep) NotifyFinish(err utils.ERROR) utils.ERROR {
	//netResult := model.NewNetResult(config.ERROR_CODE_success, "", this.fileIndex.ID)
	//if err != nil {
	//	netResult = model.NewNetResult(config.ERROR_CODE_system_error_remote, err.Error(), nil)
	//}
	//bs, err := netResult.Marshal()
	//if err != nil {
	//	utils.Log.Error().Msgf("序列化报错:%s", err.Error())
	//	return err
	//}
	//bs = []byte(err.Error())
	//循环发送，直到成功为止
	bf := utils.NewBackoffTimer(config.MSG_timeout_interval...)
	for {
		ERR := SendUploadFileSaveDB(this.uAddr, err)
		if ERR.CheckSuccess() {
			break
		}
		bf.Wait()
	}
	return utils.NewErrorSuccess()
}

/*
清理文件夹中的块文件
*/
func (this *UploadServerStep) CleanChunks() error {
	for _, one := range this.fileIndex.Chunks {
		os.Remove(filepath.Join(transfer_manager.Recfilepath, hex.EncodeToString(one)))
	}
	return nil
}

/*
向对方传输块文件
*/
func TransferFileChunk(fileIndex *model.FileIndex, remoteAddr nodeStore.AddressNet, parallelTotal chan bool, dirPath string) error {
	<-parallelTotal
	defer func() {
		parallelTotal <- false
	}()
	fails := make([]failChunks, 0, len(fileIndex.Chunks))
	for i, one := range fileIndex.Chunks {
		fails = append(fails, failChunks{
			pullID: fileIndex.PullIDs[i],
			chunk:  one,
			index:  i,
		})
	}
	fileIndex.SupplierIDs = make([][]byte, len(fileIndex.Chunks))

	signalID, signal := file_transfer.ManagerStatic.GetListeningPushFinishSignal()

	//signalId, signal := transfer_manager.TransferMangerStatic.GetListeningPushFinishSignal()
	//当发送失败，则间隔时间再次发送
	pts := make(map[string]*[]byte)
	timer := utils.NewBackoffTimer(60)
	for {
		newFails := make([]failChunks, 0)
		for _, one := range fails {
			pathOne := filepath.Join(dirPath, hex.EncodeToString(one.chunk))
			pathOne, err := filepath.Abs(pathOne)
			if err != nil {
				return err
			}
			//utils.Log.Info().Msgf("开始上传块文件:%s", pathOne)
			supplierID, ERR := file_transfer.ManagerStatic.SendFile(config.FILE_TRANSFER_CLASS_storage, remoteAddr, pathOne, one.pullID)
			//ptOne, err := transfer_manager.TransferMangerStatic.NewPushTask(pathOne, remoteAddr)
			if !ERR.CheckSuccess() {
				//utils.Log.Info().Msgf("上传块文件错误:%s", ERR.String())
				newFails = append(newFails, one)
				continue
			}
			pts[utils.Bytes2string(supplierID)] = nil
			fileIndex.SupplierIDs[one.index] = supplierID
		}
		//utils.Log.Info().Msgf("失败多少个:%d", len(newFails))
		if len(newFails) == 0 {
			break
		}
		fails = newFails
		timer.Wait()
	}
	//等待上传完成
	for one := range signal {
		//utils.Log.Info().Msgf("发送成功信号:%+v", one)
		key := utils.Bytes2string(one.SupplierID)
		if _, ok := pts[key]; ok {
			delete(pts, key)
		}
		if len(pts) == 0 {
			break
		}
	}

	file_transfer.ManagerStatic.ReturnListeningPushFinishSignal(signalID)

	//transfer_manager.TransferMangerStatic.ReturnListeningPushFinishSignal(signalId)
	return nil
}

type failChunks struct {
	pullID []byte //下载者的id
	chunk  []byte //文件块hash
	index  int    //数组索引
}

/*
等待传输完成
*/
func WaitDownloadFinish(signalID string, signalChan chan *file_transfer.DownloadStep, fileIndex model.FileIndex) utils.ERROR {
	//utils.Log.Info().Msgf("等待文件传输完成:%+v", fileIndex)
	total := make(map[string]string)
	for _, one := range fileIndex.Chunks {
		//utils.Log.Info().Msgf("文件分片:%s", hex.EncodeToString(one))
		total[hex.EncodeToString(one)] = ""
	}
	//utils.Log.Info().Msgf("等待传输完成")
	for one := range signalChan {
		fileName := one.FileIndex.Name
		//fileName := one.FileInfo.Name
		fileName = strings.Split(fileName, "_")[0]
		//utils.Log.Info().Msgf("收到文件分片传输完成:%s", fileName)
		if _, ok := total[fileName]; ok {
			//utils.Log.Info().Msgf("删除一个任务:%s", fileName)
			delete(total, fileName)
		}
		//utils.Log.Info().Msgf("任务是否归零:%d", total)
		if len(total) == 0 {
			break
		}
	}
	//utils.Log.Info().Msgf("等待文件传输完成")
	file_transfer.ManagerStatic.ReturnListeningPullFinishSignal(signalID)
	//transfer_manager.TransferMangerStatic.ReturnListeningPullFinishSignal(signalID)
	return utils.NewErrorSuccess()
}

/*
向对方传输块文件
*/
//func TransferFileChunk(fileIndex *model.FileIndex, remoteAddr nodeStore.AddressNet, parallelTotal chan bool, dirPath string) error {
//	<-parallelTotal
//	defer func() {
//		parallelTotal <- false
//	}()
//	pts := make(map[uint64]*transfer_manager.PushTask)
//	fails := make([][]byte, 0)
//	for _, one := range fileIndex.Chunks {
//		fails = append(fails, one)
//	}
//
//	signalId, signal := transfer_manager.TransferMangerStatic.GetListeningPushFinishSignal()
//	//当发送失败，则间隔时间再次发送
//	timer := utils.NewBackoffTimer(60)
//	for {
//		newFails := make([][]byte, 0)
//		for _, one := range fails {
//			pathOne := filepath.Join(dirPath, hex.EncodeToString(one))
//			pathOne, err := filepath.Abs(pathOne)
//			if err != nil {
//				return err
//			}
//			//utils.Log.Info().Msgf("开始上传块文件:%s %v", filepath.Join(FileName_dir_filechunk_name, hex.EncodeToString(one)), this.serverAddr)
//			ptOne, err := transfer_manager.TransferMangerStatic.NewPushTask(pathOne, remoteAddr)
//			if err != nil {
//				//utils.Log.Info().Msgf("开始上传块文件:%s", err.Error())
//				newFails = append(newFails, one)
//				continue
//			}
//			pts[ptOne.PushTaskID] = nil
//		}
//		if len(newFails) == 0 {
//			break
//		}
//		fails = newFails
//		timer.Wait()
//	}
//	//等待上传完成
//	for one := range signal {
//		if _, ok := pts[one.PushTaskID]; ok {
//			delete(pts, one.PushTaskID)
//		}
//		if len(pts) == 0 {
//			break
//		}
//	}
//	transfer_manager.TransferMangerStatic.ReturnListeningPushFinishSignal(signalId)
//	return nil
//}

/*
等待传输完成
*/
//func WaitDownloadFinish(signalID string, signalChan chan *transfer_manager.PullTask, fileIndex model.FileIndex) utils.ERROR {
//	//utils.Log.Info().Msgf("等待文件传输完成:%+v", fileIndex)
//	total := make(map[string]string)
//	for _, one := range fileIndex.Chunks {
//		//utils.Log.Info().Msgf("文件分片:%s", hex.EncodeToString(one))
//		total[hex.EncodeToString(one)] = ""
//	}
//	//utils.Log.Info().Msgf("等待传输完成")
//	for one := range signalChan {
//		fileName := one.FileInfo.Name
//		fileName = strings.Split(fileName, "_")[0]
//		//utils.Log.Info().Msgf("收到文件分片传输完成:%s", fileName)
//		if _, ok := total[fileName]; ok {
//			//utils.Log.Info().Msgf("删除一个任务:%s", fileName)
//			delete(total, fileName)
//		}
//		//utils.Log.Info().Msgf("任务是否归零:%d", total)
//		if len(total) == 0 {
//			break
//		}
//	}
//	//utils.Log.Info().Msgf("等待文件传输完成")
//	transfer_manager.TransferMangerStatic.ReturnListeningPullFinishSignal(signalID)
//	return utils.NewErrorSuccess()
//}
