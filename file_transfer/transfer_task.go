package file_transfer

import (
	"github.com/oklog/ulid/v2"
	"os"
	"time"
	"web3_gui/keystore/v2/base58"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
文件传输任务
*/
type FileTransferTask struct {
	ClassID       uint64     //文件传输单元
	SupplierID    []byte     //提供者ID
	PullID        []byte     //文件下载者ID
	TaskType      int        //任务类型。1=从共享文件夹下载;2=主动发送文件;
	LocalFilePath string     //本地文件路径
	CreateTime    int64      //创建时间
	OverTime      int64      //超时时间
	FileName      string     //文件名称
	FileSize      uint64     //文件总大小
	OffsetIndex   uint64     //已经下载的大小
	uploadFile    UploadFile //
}

/*
创建一个文件传输任务
*/
func NewFileTransferTask(db *utilsleveldb.LevelDB, classID uint64, localPath string, uploadFile UploadFile,
	pullID []byte) (*FileTransferTask, utils.ERROR) {
	fileSize := uint64(0)
	fileName := ""
	if uploadFile == nil {
		fileInfo, err := os.Stat(localPath)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		fileSize = uint64(fileInfo.Size())
		fileName = fileInfo.Name()
	} else {
		fileName, fileSize = uploadFile.GetFileSize()
	}
	now := time.Now()
	ft := FileTransferTask{
		ClassID:    classID,
		SupplierID: ulid.Make().Bytes(), //提供者ID
		PullID:     pullID,              //文件下载者ID
		//TaskType      :1,    //任务类型。1=从共享文件夹下载;2=主动发送文件;
		LocalFilePath: localPath,                                              //本地文件路径
		CreateTime:    now.Unix(),                                             //创建时间
		OverTime:      now.Unix() + int64(Transfer_task_overtime/time.Second), //超时时间
		FileSize:      fileSize,                                               //
		FileName:      fileName,                                               //
		uploadFile:    uploadFile,                                             //
	}
	return &ft, utils.NewErrorSuccess()
}

type FileTransferTaskVO struct {
	ClassID       uint64 //文件传输单元
	SupplierID    string //提供者ID
	Name          string //文件名称
	TaskType      int    //任务类型。1=从共享文件夹下载;2=主动发送文件;
	LocalFilePath string //本地文件路径
	CreateTime    int64  //创建时间
	OverTime      int64  //超时时间
	FileSize      uint64 //文件总大小
	OffsetIndex   uint64 //已经下载的大小
}

/*
转换上传列表
*/
func (this *FileTransferTask) ConverVO() *FileTransferTaskVO {
	ds := FileTransferTaskVO{
		ClassID:       this.ClassID,                           //文件传输单元
		SupplierID:    string(base58.Encode(this.SupplierID)), //提供者ID
		Name:          this.FileName,                          //
		TaskType:      this.TaskType,                          //任务类型。1=从共享文件夹下载;2=主动发送文件;
		LocalFilePath: this.LocalFilePath,                     //本地文件路径
		CreateTime:    this.CreateTime,                        //创建时间
		OverTime:      this.OverTime,                          //超时时间
		FileSize:      this.FileSize,                          //文件总大小
		OffsetIndex:   this.OffsetIndex,                       //已经下载的大小
	}
	return &ds
}
