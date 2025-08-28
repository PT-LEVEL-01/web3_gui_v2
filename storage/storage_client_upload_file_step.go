package storage

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	chainconfig "web3_gui/chain/config"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/db"
	"web3_gui/im/model"
	libp2pconfig "web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
负责一个文件的上传管理
1.切片。
2.发送索引文件。
3.上传文件块。
4.等服务器保存成功。
*/
type UploadFileStep struct {
	serverAddr    nodeStore.AddressNet //
	fileIndex     *model.FileIndex     //
	saveDBSignal  chan bool            //
	uploadManager *UploadManager       //
}

func NewUploadFileStep(sAddr nodeStore.AddressNet, fileIndex *model.FileIndex) *UploadFileStep {
	uf := UploadFileStep{
		serverAddr:   sAddr,
		fileIndex:    fileIndex,
		saveDBSignal: make(chan bool, 1),
	}
	return &uf
}

func (this *UploadFileStep) Start(parallelTotal chan bool) {
	go func() {
		defer func() {
			this.CleanChunks()
			this.uploadManager.delFile(this.fileIndex.ID)
		}()
		//先切片
		ERR := this.diced()
		if !ERR.CheckSuccess() {
			this.fileIndex.Error = ERR.String()
			return
		}
		//utils.Log.Info().Msgf("切片完成")
		//发送索引文件
		ERR = this.SendUploadFileIndex()
		if !ERR.CheckSuccess() {
			//不需要上传文件
			if ERR.Code == config.ERROR_CODE_storage_No_need_upload_files {
				//utils.Log.Info().Msgf("文件重复，不需要上传文件")
			} else {
				//utils.Log.Info().Msgf("发送文件索引错误:%s", ERR.String())
				this.fileIndex.Error = ERR.String()
			}
			return
		}
		//utils.Log.Info().Msgf("发送索引文件完成:%s", hex.EncodeToString(this.fileIndex.ID))
		//再上传块文件
		err := TransferFileChunk(this.fileIndex, this.serverAddr, parallelTotal, config.FileName_dir_filechunk_name)
		if err != nil {
			//this.CleanChunks()
			//utils.Log.Info().Msgf("上传块文件错误:%s", err.Error())
			this.fileIndex.Error = err.Error()
			return
		}
		//utils.Log.Info().Msgf("等服务器保存到数据库:%s", hex.EncodeToString(this.fileIndex.ID))
		//等服务器保存到数据库
		err = this.saveDB()
		if err != nil {
			//this.CleanChunks()
			//utils.Log.Info().Msgf("上传块文件错误:%s", err.Error())
			this.fileIndex.Error = err.Error()
			return
		}
		//this.CleanChunks()
		//utils.Log.Info().Msgf("上传文件完成:%s", hex.EncodeToString(this.fileIndex.ID))
		//this.uploadManager.delFile(this.fileIndex.ID)
	}()
}

/*
切片一个文件
*/
func (this *UploadFileStep) diced() utils.ERROR {
	fileIndex := this.fileIndex
	//计算源文件hash
	hashBs, err := utils.FileSHA3_256(fileIndex.AbsPath)
	if err != nil {
		utils.Log.Error().Msgf("计算文件hash错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}

	_, prk, ERR := Node.Keystore.GetNetAddrKeyPair(chainconfig.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("获取私钥 ERR", ERR.String()).Send()
		return ERR
	}
	//utils.Log.Info().Msgf("加密文件密码私钥:%+v", prk)

	//加密文件密码用
	key, iv := splitKeyAndIv(prk)
	//把原文件hash加密
	cipherHashBs, err := utils.AesCTR_Encrypt(key, iv, hashBs)
	if err != nil {
		utils.Log.Error().Msgf("获取私钥 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}

	//加密文件
	key, iv = splitKeyAndIv(hashBs)
	//切片文件
	newFileIndex, ERR := Diced(key, iv, fileIndex.AbsPath, config.FileName_dir_filechunk_name)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("切片文件错误:%s", ERR.String())
		return ERR
	}
	//utils.Log.Info().Msgf("文件ID:%+v", newFileIndex.ID)
	fileIndex.Hash = newFileIndex.Hash                         //文件加密后的hash值
	fileIndex.UserAddr = newFileIndex.UserAddr                 //用户地址
	fileIndex.Pwds = [][]byte{cipherHashBs}                    //
	fileIndex.Version = newFileIndex.Version                   //版本号
	fileIndex.Name = newFileIndex.Name                         //文件名称
	fileIndex.FileSize = newFileIndex.FileSize                 //文件总大小
	fileIndex.ChunkCount = newFileIndex.ChunkCount             //分片总量
	fileIndex.ChunkOneSize = newFileIndex.ChunkOneSize         //每一个分片大小
	fileIndex.Chunks = newFileIndex.Chunks                     //每一个分片ID
	fileIndex.PullIDs = newFileIndex.PullIDs                   //
	fileIndex.ChunkOffsetIndex = newFileIndex.ChunkOffsetIndex //
	fileIndex.PermissionType = newFileIndex.PermissionType     //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
	fileIndex.EncryptionType = newFileIndex.EncryptionType     //加密类型
	fileIndex.Time = newFileIndex.Time                         //创建时间
	fileIndex.Lock = new(sync.RWMutex)                         //
	this.fileIndex = fileIndex
	atomic.CompareAndSwapUint32(&fileIndex.Status, config.UploadFile_status_diced, config.UploadFile_status_upload)
	return db.StorageClient_SaveFilePwd(fileIndex.Hash, hashBs)
}

/*
上传文件索引
*/
func (this *UploadFileStep) SendUploadFileIndex() utils.ERROR {
	//循环发送，直到成功为止
	bf := utils.NewBackoffTimer(config.MSG_timeout_interval...)
	for {
		//utils.Log.Info().Msgf("发送上传文件索引:%+v", this.fileIndex)
		ERR := SendUploadFileIndex(this.serverAddr, this.fileIndex)
		if ERR.CheckSuccess() {
			break
		}
		if ERR.Code == libp2pconfig.ERROR_code_wait_msg_timeout {
			bf.Wait()
		}
		this.fileIndex.Error = ERR.String()
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
将文件保存到数据库
*/
func (this *UploadFileStep) saveDB() error {
	<-this.saveDBSignal
	return nil
}

/*
设置保存到数据库完成
*/
func (this *UploadFileStep) SetSaveDBFinish() {
	select {
	case this.saveDBSignal <- false:
	default:
	}
}

/*
清理文件夹中的块文件
*/
func (this *UploadFileStep) CleanChunks() error {
	for _, one := range this.fileIndex.Chunks {
		os.Remove(filepath.Join(config.FileName_dir_filechunk_name, hex.EncodeToString(one)))
	}
	return nil
}

/*
清理文件夹中的块文件
*/
func (this *UploadFileStep) ConverUploadTask() *file_transfer.FileTransferTaskVO {
	ftt := file_transfer.FileTransferTaskVO{
		ClassID: 0, //文件传输单元
		//SupplierID    string //提供者ID
		//Name: this.fileIndex.Name[0], //文件名称
		//TaskType      int    //任务类型。1=从共享文件夹下载;2=主动发送文件;
		LocalFilePath: this.fileIndex.AbsPath, //本地文件路径
		//CreateTime:    this.fileIndex.Time[0],  //创建时间
		OverTime: 0,                       //超时时间
		FileSize: this.fileIndex.FileSize, //文件总大小
		//OffsetIndex   uint64 //已经下载的大小
	}
	for _, one := range this.fileIndex.ChunkOffsetIndex {
		ftt.OffsetIndex += one
	}
	if this.fileIndex.Time != nil && len(this.fileIndex.Time) >= 1 {
		ftt.CreateTime = this.fileIndex.Time[0]
	}
	if this.fileIndex.Name != nil && len(this.fileIndex.Name) >= 1 {
		ftt.Name = this.fileIndex.Name[0]
	}
	return &ftt
}
