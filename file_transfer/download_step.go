package file_transfer

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
	"web3_gui/config"
	"web3_gui/keystore/v2/base58"
	"web3_gui/libp2parea/v2"
	libp2pconfig "web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type DownloadStep struct {
	area                    *libp2parea.Node            //
	serverAddr              nodeStore.AddressNet        //
	FileIndex               *FileIndex                  //
	localPath               string                      //文件下载完后，保存到本地的路径
	startAndStopSignal      chan bool                   //开始下载和暂停的信号
	saveDBSignal            chan bool                   //
	OffsetIndex             uint64                      //块起始位置偏移量
	chunkSizeCalculator     *engine.ChunkSizeCalculator //
	fileChunk               *FileChunk                  //
	pullFinishSignalMonitor chan *DownloadStep          //
	SupplierID              []byte                      //提供者ID
	PullID                  []byte                      //文件下载者ID
	downloadFile            DownloadFile                //
	transfer                *Transfer                   //
}

func NewDownloadStep(area *libp2parea.Node, sAddr nodeStore.AddressNet, fileIndex *FileIndex, localPath string,
	pullFinishSignalMonitor chan *DownloadStep, downloadFile DownloadFile, transfer *Transfer) *DownloadStep {
	uf := DownloadStep{
		area:                    area,
		serverAddr:              sAddr,
		FileIndex:               fileIndex,
		localPath:               localPath,
		startAndStopSignal:      make(chan bool, 1),
		saveDBSignal:            make(chan bool, 1),
		chunkSizeCalculator:     engine.NewChunkSizeCalculator(),
		pullFinishSignalMonitor: pullFinishSignalMonitor,
		SupplierID:              fileIndex.SupplierID,
		PullID:                  fileIndex.PullID,
		downloadFile:            downloadFile,
		transfer:                transfer,
	}
	return &uf
}

/*
开始下载
*/
func (this *DownloadStep) Start() {
	go func() {
		this.startAndStopSignal <- true
		this.downloadStep()
	}()
}

/*
开始等待同意下载
*/
func (this *DownloadStep) StartWait() {
	go func() {
		this.downloadStep()
	}()
}

/*
开始下载
*/
func (this *DownloadStep) downloadStep() utils.ERROR {
	for {
		start := <-this.startAndStopSignal
		if start {
			break
		}
	}
	this.fileChunk = &FileChunk{
		ClassID:     this.FileIndex.ClassID,
		SupplierID:  this.FileIndex.SupplierID,
		PullID:      this.FileIndex.PullID,
		OffsetIndex: 0,
		ChunkSize:   0,
		Data:        nil,
	}
	ERR := this.createFile()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("下载文件错误:%s", ERR.String())
		return ERR
	}
	this.loopDownload()
	ERR = this.SendTransferFinish()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("下载文件错误:%s", ERR.String())
		return ERR
	}
	//utils.Log.Info().Msgf("下载文件完成 111:%v", this.PullID)
	this.pullFinishSignalMonitor <- this
	//utils.Log.Info().Msgf("下载文件完成 222:%v", this.PullID)
	return utils.NewErrorSuccess()
}

/*
创建一个空文件
*/
func (this *DownloadStep) createFile() utils.ERROR {
	//utils.Log.Info().Msgf("创建文件:%s", this.localPath)
	dir, _ := filepath.Split(this.localPath)
	err := utils.CheckCreateDir(dir)
	if err != nil {
		utils.Log.Info().Msgf("创建文件夹错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//先创建一个文件
	f, err := os.Create(this.localPath)
	if err != nil {
		utils.Log.Info().Msgf("创建文件错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	defer f.Close()
	//设置文件大小，先占用磁盘空间
	if err := f.Truncate(int64(this.FileIndex.FileSize)); err != nil {
		utils.Log.Info().Msgf("创建文件错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
循环下载
*/
func (this *DownloadStep) loopDownload() utils.ERROR {
	var spend time.Duration
	var err error
	var ERR utils.ERROR
	start := time.Now()

	backoff := utils.NewBackoffTimerChan(NET_overtime_retry...)
	for {
		//utils.Log.Info().Msgf("本次下载，上次下载花费时间:%s", spend)
		chunkSize := this.chunkSizeCalculator.GetChunkSize(spend, true)
		//chunkSize := ChunkSizeCalculators(chunkSize, spend)
		start = time.Now()
		ERR = this.getRemoteChunk(chunkSize, NET_timeout)
		spend = time.Now().Sub(start)
		//utils.Log.Info().Msgf("本次下载花费时间:%s", spend)
		if !ERR.CheckSuccess() {
			if ERR.Code == libp2pconfig.ERROR_code_wait_msg_timeout {
				utils.Log.Info().Msgf("name:%s 请求超时:%s %s", this.FileIndex.Name, ERR.String(), spend)
				backoff.Wait(context.Background())
				continue
			}
			utils.Log.Info().Msgf("name:%s 获取块文件错误:%s", this.FileIndex.Name, ERR.String())
			return ERR
		}
		backoff.Reset()
		err = writeFileChunk(this.localPath, int64(this.fileChunk.OffsetIndex)-int64(len(this.fileChunk.Data)), this.fileChunk.Data, this.downloadFile)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Msgf("name:%s 本次块文件大小:%d 花费时间:%s 获取的数据大小:%d", this.FileIndex.Name, chunkSize, spend, len(this.fileChunk.Data))
		this.fileChunk.Data = nil
		//utils.Log.Info().Msgf("name:%s fileChunk:%+v", this.FileIndex.Name, this.fileChunk)
		//utils.Log.Info().Msgf("name:%s fileIndex:%+v", this.FileIndex.Name, this.FileIndex)
		//下载完成
		if this.fileChunk.OffsetIndex == this.FileIndex.FileSize {
			break
		}
	}
	return utils.NewErrorSuccess()
}

/*
从远端获取文件块
*/
func (this *DownloadStep) getRemoteChunk(chunkSize utils.Byte, timeout time.Duration) utils.ERROR {
	//utils.Log.Info().Uint64("从远端获取文件块 chunkSize", uint64(chunkSize)).Interface("timeout", timeout).Send()
	atomic.StoreUint64(&this.fileChunk.OffsetIndex, this.OffsetIndex)
	atomic.StoreUint64(&this.fileChunk.ChunkSize, uint64(chunkSize))
	this.fileChunk.Data = nil
	bs, err := this.fileChunk.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := this.area.SendP2pMsgHEWaitRequest(NET_CODE_get_file_chunk, &this.serverAddr, bs, timeout)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("从远端获取文件块 错误:%s", ERR.String())
		return utils.NewErrorSysSelf(err)
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		ERR := nr.ConvertERROR()
		utils.Log.Info().Msgf("获取文件块错误:%s", ERR.String())
		return nr.ConvertERROR()
	}
	//utils.Log.Info().Msgf("返回文件块 成功")
	//返回成功
	this.fileChunk.Data = nr.Data
	offset := this.fileChunk.OffsetIndex + uint64(len(nr.Data))
	atomic.StoreUint64(&this.fileChunk.OffsetIndex, offset)
	atomic.StoreUint64(&this.OffsetIndex, this.fileChunk.OffsetIndex)
	return utils.NewErrorSuccess()
}

/*
发送传输完成
*/
func (this *DownloadStep) SendTransferFinish() utils.ERROR {
	atomic.StoreUint64(&this.fileChunk.OffsetIndex, 0)
	//this.fileChunk.OffsetIndex = 0
	atomic.StoreUint64(&this.fileChunk.ChunkSize, 0)
	//this.fileChunk.ChunkSize = 0
	this.fileChunk.Data = nil
	bs, err := this.fileChunk.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	var resultBS *[]byte
	backoff := utils.NewBackoffTimerChan(NET_overtime_retry...)
	var ERR utils.ERROR
	for {
		resultBS, ERR = this.area.SendP2pMsgHEWaitRequest(NET_CODE_notice_transfer_finish, &this.serverAddr, bs, time.Second*10)
		if ERR.CheckSuccess() {
			break
		}
		if ERR.Code == libp2pconfig.ERROR_code_wait_msg_timeout {
			backoff.Wait(context.Background())
			continue
		}
		utils.Log.Info().Msgf("发送传输完成消息 错误:%s", ERR.String())
		return ERR
	}

	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
下载列表
*/
type DownloadStepVO struct {
	DBID       string //
	PullID     string //文件下载者ID
	Name       string //文件名称
	FileSize   uint64 //文件总大小
	PullSize   uint64 //已经下载的大小
	Speed      uint64 //下载速度。单位：B/s
	LocalPath  string //文件下载完后，保存到本地的路径
	Status     int    //状态
	CreateTime int64  //创建时间
	FinishTime int64  //完成时间
}

/*
转换下载列表
*/
func (this *DownloadStep) ConverVO() *DownloadStepVO {
	ds := DownloadStepVO{
		PullID:     string(base58.Encode(this.PullID)),
		Name:       this.FileIndex.Name,
		FileSize:   atomic.LoadUint64(&this.FileIndex.FileSize),
		PullSize:   atomic.LoadUint64(&this.fileChunk.OffsetIndex),
		Speed:      0,
		LocalPath:  this.localPath,
		Status:     0,
		CreateTime: this.FileIndex.Time,
	}
	return &ds
}
