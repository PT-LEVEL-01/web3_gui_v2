package storage

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/utils"
)

const (
	StorageServerMultcastInterval = time.Second * 10 //time.Minute * 20 //广播时间间隔
)

// 本节点是否打开存储提供者广播
var StorageServerMultcastIsOpen *atomic.Bool

// 保存在线次数 key:string=网络地址;value:uint64=在线次数;
//var StorageServerListCount = new(sync.Map)

func InitStorageServerSetup() error {
	StorageServerMultcastIsOpen = new(atomic.Bool)
	sinfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		utils.Log.Error().Msgf("加载离线服务节点列表错误:%s", err.Error())
		return err
	}
	if sinfo == nil {
		StorageServerMultcastIsOpen.Store(false)
	} else {
		StorageServerMultcastIsOpen.Store(sinfo.IsOpen)
	}

	//for _, one := range servers {
	//	StorageServerListCount.Store(utils.Bytes2string(one.Addr), one.Count)
	//}
	return nil
}

func SetStorageServerOpen() {
	StorageServerMultcastIsOpen.Store(true)
}

func SetStorageServerClose() {
	StorageServerMultcastIsOpen.Store(false)
}

/*
启动定时广播自己是否提供IM数据离线存储服务器
*/
func InitStorageServerMultcast() {
	go LoopMultcastStorageServerInfo()
}

/*
定时广播本节点是否提供IM数据离线存储服务器
*/
func LoopMultcastStorageServerInfo() {
	ticker := time.NewTicker(StorageServerMultcastInterval)
	for range ticker.C {
		if !StorageServerMultcastIsOpen.Load() {
			continue
		}
		//执行一次广播
		StorageServerMultcastOnce()
	}
}

/*
广播一次本节点信息
*/
func StorageServerMultcastOnce() {
	//utils.Log.Info().Msgf("广播一次本节点存储服务器信息")
	storageServerInfo, err := db.StorageServer_GetServerInfo()
	if err != nil {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", err.Error())
		return
	}
	storageServerInfo.Directory = nil
	bs, err := storageServerInfo.Proto()
	if err != nil {
		return
	}

	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return
	}

	Node.SendMulticastMsg(config.MSGID_STORAGE_multicast_nodeinfo_recv, bs)
	storageServerInfo.Addr = *Node.GetNetId()
	AddStorageServerList(storageServerInfo)
}

var AddStorageServerListLock = new(sync.Mutex)

func AddStorageServerList(sinfo *model.StorageServerInfo) {
	AddStorageServerListLock.Lock()
	defer AddStorageServerListLock.Unlock()
	sinfos, err := db.StorageClient_GetStorageServerListByAddrs(sinfo.Addr)
	if err != nil {
		utils.Log.Error().Msgf("查询存储提供者信息错误:%s", err.Error())
		return
	}

	count := uint64(0)
	if sinfos != nil && len(sinfos) > 0 {
		count = sinfos[0].Count
	}
	if count != uint64(math.MaxUint64) {
		count++
	}
	sinfo.Count = count
	db.StorageClient_AddStorageServer(sinfo)
}
