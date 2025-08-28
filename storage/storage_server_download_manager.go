package storage

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type DownloadServerManager struct {
	lock               *sync.RWMutex                  //
	DBC                *DBCollections                 //数据库管理
	downloadStep       map[string]*DownloadServerStep //传输任务列表。key:string=用户地址+文件ID联合做key;value:*DownloadServerStep=;
	parallelTotal      chan bool                      //并行传输最大数量
	finishSignal       chan []byte                    //传输完成信号
	chunksRefCountLock *sync.RWMutex                  //
	chunksRefCount     map[string]uint64              //key:string=块文件名称;value:uint64=引用次数;
}

func NewDownloadServerManager(dbc *DBCollections) *DownloadServerManager {
	parallelTotal := 1000
	dm := DownloadServerManager{
		lock:               new(sync.RWMutex),
		DBC:                dbc,
		downloadStep:       make(map[string]*DownloadServerStep),
		parallelTotal:      make(chan bool, parallelTotal),
		finishSignal:       make(chan []byte, 1),
		chunksRefCountLock: new(sync.RWMutex),
		chunksRefCount:     make(map[string]uint64),
	}
	for range parallelTotal {
		dm.parallelTotal <- false
	}
	go dm.LoopClean()
	return &dm
}

func (this *DownloadServerManager) LoopClean() {
	for one := range this.finishSignal {
		this.lock.Lock()
		delete(this.downloadStep, utils.Bytes2string(one))
		this.lock.Unlock()
	}
}

/*
添加一个下载任务
*/
func (this *DownloadServerManager) AddDownloadTask(uAddr nodeStore.AddressNet, fileIndex *model.FileIndex) {
	key := append(uAddr.GetAddr(), fileIndex.Hash...)
	ds := NewDownloadServerStep(uAddr, fileIndex, this.DBC, this.finishSignal, this)
	ds.Start(this.parallelTotal)
	this.lock.Lock()
	defer this.lock.Unlock()
	this.downloadStep[utils.Bytes2string(key)] = ds
}

/*
把分片从数据库中取出来
*/
func (this *DownloadServerManager) TakeChunkFromDB(chunks [][]byte) utils.ERROR {
	this.chunksRefCountLock.Lock()
	defer this.chunksRefCountLock.Unlock()
	err := utils.CheckCreateDir(config.FileName_dir_filechunk_name)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	for _, one := range chunks {
		count, ok := this.chunksRefCount[utils.Bytes2string(one)]
		if ok {
			count++
			this.chunksRefCount[utils.Bytes2string(one)] = count
			continue
		}
		this.chunksRefCount[utils.Bytes2string(one)] = 1

		//从数据库中取出来，保存到本地磁盘
		chunkName := hex.EncodeToString(one)
		bs, ERR := this.DBC.FindChunk(one)
		if !ERR.CheckSuccess() {
			return ERR
		}
		file, err := os.OpenFile(filepath.Join(config.FileName_dir_filechunk_name, chunkName), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			file.Close()
			return utils.NewErrorSysSelf(err)
		}
		_, err = file.Write(*bs)
		if err != nil {
			file.Close()
			return utils.NewErrorSysSelf(err)
		}
		file.Close()
	}
	return utils.NewErrorSuccess()
}

/*
把分片从数据库中取出来
*/
func (this *DownloadServerManager) CleanChunks(chunks [][]byte) error {
	this.chunksRefCountLock.Lock()
	defer this.chunksRefCountLock.Unlock()
	for _, one := range chunks {
		count, _ := this.chunksRefCount[utils.Bytes2string(one)]
		if count > 0 {
			count--
		}
		if count > 0 {
			this.chunksRefCount[utils.Bytes2string(one)] = count
			continue
		}
		os.Remove(filepath.Join(config.FileName_dir_filechunk_name, hex.EncodeToString(one)))
	}
	return nil
}

/*
文件下载服务器端处理步骤
*/
type DownloadServerStep struct {
	id              []byte                 //任务ID
	userAddr        nodeStore.AddressNet   //
	fileIndex       *model.FileIndex       //
	DBC             *DBCollections         //数据库管理
	finishSignal    chan []byte            //
	downloadServerM *DownloadServerManager //
}

func NewDownloadServerStep(uAddr nodeStore.AddressNet, fileIndex *model.FileIndex, dbc *DBCollections,
	finishSignal chan []byte, downloadServerM *DownloadServerManager) *DownloadServerStep {
	uss := DownloadServerStep{
		id:              append(uAddr.GetAddr(), fileIndex.Hash...),
		userAddr:        uAddr,
		fileIndex:       fileIndex,
		DBC:             dbc,
		finishSignal:    finishSignal,
		downloadServerM: downloadServerM,
	}
	return &uss
}

/*
开始按步骤处理
*/
func (this *DownloadServerStep) Start(parallelTotal chan bool) {
	go func() {
		//utils.Log.Info().Msgf("把分片从数据库中取出来")
		//把分片从数据库中取出来
		err := this.TakeChunkFromDB()
		if err != nil {
			utils.Log.Error().Msgf("把分片从数据库中取出来 失败:%s", err.Error())
			return
		}
		//utils.Log.Info().Msgf("等待文件传输完成")
		//向对方传输块文件完成
		err = TransferFileChunk(this.fileIndex, this.userAddr, parallelTotal, config.FileName_dir_filechunk_name)
		if err != nil {
			return
		}
		//utils.Log.Info().Msgf("传输文件完成，开始清理本地块文件")
		this.CleanChunks()
		//utils.Log.Info().Msgf("清理块文件完成")
		this.finishSignal <- this.id
	}()
}

/*
把分片从数据库中取出来
*/
func (this *DownloadServerStep) TakeChunkFromDB() error {
	this.downloadServerM.TakeChunkFromDB(this.fileIndex.Chunks)
	return nil
}

/*
清理文件夹中的块文件
*/
func (this *DownloadServerStep) CleanChunks() error {
	this.downloadServerM.CleanChunks(this.fileIndex.Chunks)
	return nil
}
