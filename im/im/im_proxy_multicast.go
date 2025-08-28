package im

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
	imProxyMultcastInterval = time.Second * 10 //time.Minute * 20 //广播时间间隔
)

// 本节点是否打开离线服务
var imProxyOfflineDataIsOpen *atomic.Bool = new(atomic.Bool)

// 保存离线服务器在线次数 key:string=网络地址;value:uint64=在线次数;
//var imServerOfflineCount = new(sync.Map)

func InitImProxySetup() utils.ERROR {
	//imProxyOfflineDataIsOpen = new(atomic.Bool)
	sinfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*Node.GetNetId(), true)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("加载离线服务节点列表错误:%s", ERR.String())
		return ERR
	}
	//if sinfo == nil {
	//	StorageServerMultcastIsOpen.Store(false)
	//} else {
	//	StorageServerMultcastIsOpen.Store(sinfo.IsOpen)
	//}
	if sinfo == nil {
		SetImProxyClose()
		return utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("本节点单价:%d", priceUnit)
	if sinfo.IsOpen {
		SetImProxyOpen()
	} else {
		SetImProxyClose()
	}
	return utils.NewErrorSuccess()
}

func SetImProxyOpen() {
	imProxyOfflineDataIsOpen.Store(true)
}

func SetImProxyClose() {
	imProxyOfflineDataIsOpen.Store(false)
}

/*
启动定时广播自己是否提供IM数据离线存储服务器
*/
func InitImProxyMultcast() {
	go LoopMultcastImProxyInfo()
}

/*
定时广播本节点是否提供IM数据离线存储服务器
*/
func LoopMultcastImProxyInfo() {
	ticker := time.NewTicker(imProxyMultcastInterval)
	for range ticker.C {
		if !imProxyOfflineDataIsOpen.Load() {
			continue
		}
		//执行一次广播
		ImProxyMultcastOnce()
	}
}

/*
广播一次本节点信息
*/
func ImProxyMultcastOnce() utils.ERROR {
	serverInfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*Node.GetNetId(), true)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("广播离线服务节点时查询本节点信息错误:%s", ERR.String())
		return ERR
	}
	bs, err := serverInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		//return nil, utils.NewErrorSysSelf(err)
		utils.Log.Error().Msgf("格式化广播消息错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	ERR = Node.SendMulticastMsg(config.MSGID_IM_PROXY_multicast_nodeinfo_recv, bs)
	return ERR
}

var AddImProxyLock = new(sync.Mutex)

/*
添加一个代理服务器信息
*/
func AddProxy(info *model.StorageServerInfo) utils.ERROR {
	AddImProxyLock.Lock()
	defer AddImProxyLock.Unlock()
	sinfos, ERR := db.ImProxyClient_GetProxyListByAddrs(*Node.GetNetId(), info.Addr)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("查询存储提供者信息错误:%s", ERR.String())
		return ERR
	}
	count := uint64(0)
	if sinfos != nil && len(sinfos) > 0 {
		count = sinfos[0].Count
	}
	if count != uint64(math.MaxUint64) {
		count++
	}
	info.Count = count
	return db.ImProxyClient_AddImProxy(*Node.GetNetId(), info)
}
