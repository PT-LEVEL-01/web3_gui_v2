package storage

import (
	"encoding/hex"
	"github.com/oklog/ulid/v2"
	"os"
	"path/filepath"
	"sync"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
上传文件的多任务管理
*/
type UploadManager struct {
	files         *sync.Map //保存上传任务 key:文件hash=;value:*UploadFileStep=;
	tag           chan bool //
	parallelTotal chan bool //并行上传最大数量
}

func NewUploadManager() *UploadManager {
	parallelTotal := 1 //并行上传最大数量
	um := &UploadManager{
		files:         new(sync.Map),                  //
		tag:           make(chan bool, 1),             //
		parallelTotal: make(chan bool, parallelTotal), //
	}
	um.tag <- false
	for range parallelTotal {
		um.parallelTotal <- false
	}
	//go um.Start()
	return um
}

/*
获取上传文件列表
*/
func (this *UploadManager) GetUploadFileList() []*file_transfer.FileTransferTaskVO {
	files := make([]*file_transfer.FileTransferTaskVO, 0)
	this.files.Range(func(k, v interface{}) bool {
		task := v.(*UploadFileStep)
		t := file_transfer.ManagerStatic.GetClass(config.FILE_TRANSFER_CLASS_storage)
		if t == nil {
			return true
		}
		for i, sid := range task.fileIndex.SupplierIDs {
			if len(sid) == 0 {
				continue
			}
			ftt := t.FindTransferTask(sid)
			if ftt == nil {
				continue
			}
			task.fileIndex.ChunkOffsetIndex[i] = ftt.OffsetIndex
		}
		files = append(files, task.ConverUploadTask())
		return true
	})
	return files
}

/*
添加一个上传文件
@dirID    []byte    所属文件夹ID
*/
func (this *UploadManager) AddFile(sAddr nodeStore.AddressNet, dirID []byte, filePath string) error {
	fileIndex := model.FileIndex{
		ID: ulid.Make().Bytes(),
		//Name:    []string{fileinfo.Name()},
		Status:  config.UploadFile_status_start,
		AbsPath: filePath,
		DirID:   [][]byte{dirID},
	}
	utils.Log.Info().Msgf("开始上传一个文件:%s", filePath)
	//fileIndex.Hash = hashBs
	ufs := NewUploadFileStep(sAddr, &fileIndex)
	ufs.uploadManager = this
	this.files.Store(utils.Bytes2string(fileIndex.ID), ufs)
	ufs.Start(this.parallelTotal)
	return nil
}

/*
上传文件完成，从列表中删除
*/
func (this *UploadManager) delFile(id []byte) {
	utils.Log.Info().Msgf("从列表中删除文件:%s", hex.EncodeToString(id))
	this.files.Delete(utils.Bytes2string(id))
}

/*
修改一个文件切片完成
*/
//func (this *UploadManager) UpdateFileStatusDicedFinish(fileID []byte) bool {
//	this.lock.RLock()
//	defer this.lock.RUnlock()
//	for _, one := range this.files {
//		if bytes.Equal(one.ID, fileID) {
//			return atomic.CompareAndSwapUint32(&one.Status, UploadFile_status_diced, UploadFile_status_upload)
//		}
//	}
//	return false
//}

/*
修改一个文件上传完成
*/
//func (this *UploadManager) UpdateFileStatusUploadFinish(fileID []byte) bool {
//	this.lock.RLock()
//	defer this.lock.RUnlock()
//	for _, one := range this.files {
//		if bytes.Equal(one.fileIndex.ID, fileID) {
//			return atomic.CompareAndSwapUint32(&one.Status, UploadFile_status_upload, UploadFile_status_saveDB)
//		}
//	}
//	return false
//}

/*
修改一个文件在服务器保存数据库完成
*/
func (this *UploadManager) UpdateFileStatusSaveDBFinish(fileID []byte) bool {
	utils.Log.Info().Msgf("设置文件已经保存到数据库:%s", hex.EncodeToString(fileID))
	v, ok := this.files.Load(utils.Bytes2string(fileID))
	if !ok {
		return false
	}
	task := v.(*UploadFileStep)
	task.SetSaveDBFinish()

	//删除相关切片文件
	for _, one := range task.fileIndex.Chunks {
		os.Remove(filepath.Join(config.FileName_dir_filechunk_name, hex.EncodeToString(one)))
	}
	return true
}

/*
接收新上传信号和完成后删除信号
*/
//func (this *UploadManager) Start() {
//	for range this.tag {
//		//检查第一个文件是否上传完成
//		del := false
//		wait := false
//		this.lock.Lock()
//		if len(this.files) == 0 {
//			wait = true
//		} else {
//			ok := atomic.CompareAndSwapUint32(&this.files[0].Status, UploadFile_status_finish, UploadFile_status_finish_over)
//			if ok {
//				temp := make([]*model.FileIndex, len(this.files)-1)
//				if len(temp) != 0 {
//					copy(temp, this.files[1:])
//				}
//				this.files = temp
//				del = true
//			}
//		}
//		this.lock.Unlock()
//		if wait {
//			continue
//		}
//		if !del {
//			continue
//		}
//		//有删除后继续新的文件切片
//		this.lock.RLock()
//		size := uint64(0)
//		//统计已经切片的文件大小
//		for _, one := range this.files {
//			status := atomic.LoadUint32(&one.Status)
//			switch status {
//			case UploadFile_status_start:
//				continue
//			case UploadFile_status_diced:
//				size += one.FileSize
//				continue
//			case UploadFile_status_upload:
//				size += one.FileSize
//				continue
//			case UploadFile_status_saveDB:
//				size += one.FileSize
//				continue
//			case UploadFile_status_finish:
//				continue
//			case UploadFile_status_finish_over:
//				continue
//			case UploadFile_status_stop:
//				continue
//			default:
//				continue
//			}
//		}
//		if size >= UploadFile_Diced_total {
//			continue
//		}
//		//不满设置大小继续切片
//		for _, one := range this.files {
//			ok := atomic.CompareAndSwapUint32(&one.Status, UploadFile_status_start, UploadFile_status_diced)
//			if !ok {
//				continue
//			}
//			size += one.FileSize
//			go this.diced(one)
//			if size >= UploadFile_Diced_total {
//				break
//			}
//		}
//		this.lock.RUnlock()
//
//	}
//}

/*
切片一个文件
*/
//func (this *UploadManager) diced(fileIndex *model.FileIndex) {
//	newFileIndex, err := Diced(fileIndex.AbsPath)
//	if err != nil {
//		utils.Log.Error().Msgf("切片文件错误:%s", err)
//		return
//	}
//	fileIndex.ID = newFileIndex.ID                         //文件加密后的hash值
//	fileIndex.UserAddr = newFileIndex.UserAddr             //用户地址
//	fileIndex.Version = newFileIndex.Version               //版本号
//	fileIndex.Name = newFileIndex.Name                     //文件名称
//	fileIndex.FileSize = newFileIndex.FileSize             //文件总大小
//	fileIndex.ChunkCount = newFileIndex.ChunkCount         //分片总量
//	fileIndex.ChunkOneSize = newFileIndex.ChunkOneSize     //每一个分片大小
//	fileIndex.Chunks = newFileIndex.Chunks                 //每一个分片ID
//	fileIndex.PermissionType = newFileIndex.PermissionType //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
//	fileIndex.EncryptionType = newFileIndex.EncryptionType //加密类型
//	fileIndex.Time = newFileIndex.Time                     //创建时间
//	atomic.CompareAndSwapUint32(&fileIndex.Status, UploadFile_status_diced, UploadFile_status_upload)
//}

/*
从数据库加载上传文件列表
*/
func LoadUploadList() {

}
