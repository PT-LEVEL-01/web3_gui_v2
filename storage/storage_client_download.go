package storage

import (
	"context"
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"os"
	"path/filepath"
	"sync"
	"time"
	chainconfig "web3_gui/chain/config"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/keystore/v2/base58"
	libp2pconfig "web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type ClientDownload struct {
	lock          *sync.RWMutex            //
	downloadList  map[string]*DownloadStep //下载文件列表
	downloadTotal chan bool                //同时下载总量
	finishSignal  chan []byte              //传输完成信号
}

func NewClientDownload() *ClientDownload {
	cd := ClientDownload{
		lock:          new(sync.RWMutex),
		downloadList:  make(map[string]*DownloadStep),
		downloadTotal: make(chan bool, 5),
		finishSignal:  make(chan []byte, 1),
	}
	go cd.LoopClean()
	return &cd
}

func (this *ClientDownload) LoopClean() {
	for one := range this.finishSignal {
		this.lock.Lock()
		delete(this.downloadList, utils.Bytes2string(one))
		this.lock.Unlock()
	}
}

/*
添加下载文件任务
*/
func (this *ClientDownload) AddDownloadTask(downloadList []DownloadStep) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range downloadList {
		one.StartDownload(this.finishSignal)
		id := append(one.serverAddr.GetAddr(), one.fileIndex.Hash...)
		this.downloadList[utils.Bytes2string(id)] = &one
	}
}

/*
获取下载列表
*/
func (this *ClientDownload) GetDownloadList() []*file_transfer.DownloadStepVO {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ds := make([]*file_transfer.DownloadStepVO, 0)
	for _, v := range this.downloadList {
		df := file_transfer.DownloadStepVO{
			//PullID     string //文件下载者ID
			Name:       v.fileIndex.Name[0],  //
			FileSize:   v.fileIndex.FileSize, //文件总大小
			PullSize:   0,                    //已经下载的大小
			Speed:      0,                    //下载速度。单位：B/s
			LocalPath:  v.fileIndex.AbsPath,  //文件下载完后，保存到本地的路径
			Status:     v.status,             //状态
			CreateTime: v.fileIndex.Time[0],  //创建时间
		}
		t := file_transfer.ManagerStatic.GetClass(config.FILE_TRANSFER_CLASS_storage)
		if t != nil {
			//计算下载进度
			v.fileIndex.Lock.RLock()
			for _, pullID := range v.fileIndex.PullIDs {
				ds := t.FindDownload(pullID)
				if ds == nil {
					continue
				}
				df.PullSize += ds.OffsetIndex
			}
			v.fileIndex.Lock.RUnlock()
		}
		ds = append(ds, &df)
	}
	//slices.SortFunc(ds, func(a, b *file_transfer.DownloadStepVO) int {
	//	return cmp.Compare(a.CreateTime, b.CreateTime)
	//})
	return ds
}

/*
获取下载完成列表
@startIndex []byte, limit uint64
*/
func (this *ClientDownload) GetDownloadFinishList(startIndex []byte, limit uint64) ([]*file_transfer.DownloadStepVO, utils.ERROR) {
	dss, ERR := StorageClient_LoadDownloadTaskFinish(startIndex, limit)
	if ERR.CheckFail() {
		return nil, ERR
	}
	ds := make([]*file_transfer.DownloadStepVO, 0)
	for _, v := range dss {
		df := file_transfer.DownloadStepVO{
			DBID:       string(base58.Encode(v.dbId)),
			PullID:     "",                   //文件下载者ID
			Name:       v.fileIndex.Name[0],  //
			FileSize:   v.fileIndex.FileSize, //文件总大小
			PullSize:   0,                    //已经下载的大小
			Speed:      0,                    //下载速度。单位：B/s
			LocalPath:  v.localPath,          //文件下载完后，保存到本地的路径
			Status:     v.status,             //状态
			CreateTime: v.createTime.Unix(),  //创建时间
			FinishTime: v.finishTime.Unix(),  //
		}
		ds = append(ds, &df)
	}
	return ds, utils.NewErrorSuccess()
}

/*
一个下载文件任务
*/
type DownloadStep struct {
	dbId       []byte               //数据库id
	serverAddr nodeStore.AddressNet //服务器地址
	fileIndex  *model.FileIndex     //文件索引
	localPath  string               //保存到本地的路径
	status     int                  //状态
	createTime time.Time            //创建时间
	finishTime time.Time            //下载完成时间
}

func NewDownloadStep(sAddr nodeStore.AddressNet, fileIndex *model.FileIndex, localPath string) *DownloadStep {
	ds := DownloadStep{
		serverAddr: sAddr,     //
		fileIndex:  fileIndex, //
		localPath:  localPath, //保存到本地的路径
	}
	return &ds
}

/*
开始下载文件
*/
func (this *DownloadStep) StartDownload(finishSignal chan []byte) {
	signalID, signalChan := file_transfer.ManagerStatic.GetListeningPullFinishSignal()

	//signalID, signalChan := transfer_manager.TransferMangerStatic.GetListeningPullFinishSignal()
	go func() {
		utils.Log.Info().Msgf("开始下载文件:%s 1111111", this.fileIndex.Name)
		//发送请求下载文件消息
		ERR := this.SendApplyDownload()
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("发送请求下载文件消息 错误：%s", ERR.String())
			return
		}
		utils.Log.Info().Msgf("开始下载文件:%s 2222222", this.fileIndex.Name)
		//等待传输完成
		ERR = WaitDownloadFinish(signalID, signalChan, *this.fileIndex)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("等待传输完成 错误：%s", ERR.String())
			return
		}

		utils.Log.Info().Msgf("开始下载文件:%s 33333333", this.fileIndex.Name)
		//解码合并文件
		ERR = this.DecodeMergeFiles()
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("解码合并文件 错误：%s", ERR.String())
			return
		}

		utils.Log.Info().Msgf("开始下载文件:%s 44444444", this.fileIndex.Name)
		this.CleanChunks()
		utils.Log.Info().Msgf("开始下载文件:%s 555555555", this.fileIndex.Name)
		//utils.Log.Info().Msgf("清理块文件完成")

		//下载任务从正在下载列表移动到已完成列表
		ERR = StorageClient_MoveDownloadTaskFinish(this)
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("记录保存到数据库 错误：%s", ERR.String())
			return
		}
		finishSignal <- append(this.serverAddr.GetAddr(), this.fileIndex.Hash...)
		utils.Log.Info().Msgf("下载文件完成")
	}()
}

/*
发送请求下载文件消息
*/
func (this *DownloadStep) SendApplyDownload() utils.ERROR {
	backoff := utils.NewBackoffTimerChan(file_transfer.NET_overtime_retry...)
	for {
		ERR := SendApplyDownload(this.serverAddr, *this.fileIndex)
		if ERR.CheckSuccess() {
			return ERR
		}
		if ERR.Code == libp2pconfig.ERROR_code_wait_msg_timeout {
			backoff.Wait(context.Background())
			continue
		}
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
解码合并文件
*/
func (this *DownloadStep) DecodeMergeFiles() utils.ERROR {
	_, prk, ERR := Node.Keystore.GetNetAddrKeyPair(chainconfig.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取私钥 错误:%s", ERR.String())
		return ERR
	}
	//解密文件密码用
	key, iv := splitKeyAndIv(prk)
	//把原文件hash解密
	hashBs, err := utils.AesCTR_Encrypt(key, iv, this.fileIndex.Pwds[0])
	if err != nil {
		utils.Log.Error().Msgf("获取私钥 错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//pwd, err := db.StorageClient_FindFilePwd(this.fileIndex.ID)
	//if err != nil {
	//	return err
	//}
	key, iv = splitKeyAndIv(hashBs)
	filePath := filepath.Join(this.localPath, this.fileIndex.Name[0])
	//err = MergeFile(key, iv, this.fileIndex, transfer_manager.Recfilepath, filePath)
	err = MergeFile(key, iv, this.fileIndex, file_transfer.FILEPATH_Recfilepath, filePath)
	return utils.NewErrorSysSelf(err)
}

/*
清理文件夹中的块文件
*/
func (this *DownloadStep) CleanChunks() error {
	for _, one := range this.fileIndex.Chunks {
		os.Remove(filepath.Join(file_transfer.FILEPATH_Recfilepath, hex.EncodeToString(one)))
		//os.Remove(filepath.Join(transfer_manager.Recfilepath, hex.EncodeToString(one)))
	}
	return nil
}

func (this *DownloadStep) Proto() (*[]byte, error) {
	sfi := this.fileIndex.Conver()
	sds := go_protos.StorageDownloadStep{
		DBID:       this.dbId,
		ServerAddr: this.serverAddr.GetAddr(),
		FileIndex:  sfi,
		LocalPath:  this.localPath,
		Status:     uint32(this.status),
		CreateTime: this.createTime.Unix(),
		FinishTime: this.finishTime.Unix(),
	}
	bs, err := sds.Marshal()
	return &bs, err
}

func ParseDownloadStep(bs []byte) (*DownloadStep, error) {
	sds := new(go_protos.StorageDownloadStep)
	err := proto.Unmarshal(bs, sds)
	if err != nil {
		return nil, err
	}
	fi, err := model.ConverFileIndex(sds.FileIndex)
	if err != nil {
		return nil, err
	}
	ds := DownloadStep{
		dbId:       sds.DBID,
		serverAddr: *nodeStore.NewAddressNet(sds.ServerAddr),
		fileIndex:  fi,
		localPath:  sds.LocalPath,
		status:     int(sds.Status),
		createTime: time.Unix(sds.CreateTime, 0),
		finishTime: time.Unix(sds.FinishTime, 0),
	}
	return &ds, nil
}
