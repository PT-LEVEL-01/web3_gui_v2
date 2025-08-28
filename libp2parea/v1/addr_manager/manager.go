package addr_manager

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
	gconfig "web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var (

//超级节点地址列表文件地址
// Path_SuperPeerAddress = filepath.Join(gconfig.Path_configDir, "nodeEntry.json")

//超级节点地址最大数量
// Sys_config_entryCount = 1000
//本地保存的超级节点地址列表
// Sys_superNodeEntry = new(sync.Map) //make(map[string]string, Sys_config_entryCount)
//清理本地保存的超级节点地址间隔时间
// Sys_cleanAddressTicker = time.Minute * 1
//需要关闭定时清理超级节点地址列表程序时，向它发送一个信号
// Sys_StopCleanSuperPeerEntry = make(chan bool)

// 保存不同渠道获得超级节点地址方法
// loadAddrFuncs = make([]func() []string, 0)
// startLoadChan    = make(chan bool, 1)     //当本机没有可用的超级节点地址，这里会收到一个信号
// SubscribesChan   = make(chan string, 100) //保存超级节点的ip地址，当有可用的超级节点地址，这里会收到一个信号
// reconnectionChan = make(chan bool, 1)     //断线重连信号
)

/*
地址管理
*/
type AddrManager struct {
	discoverEntryPath  string                  //超级节点地址列表文件地址
	p2pMaxConnPath     string                  //p2p最大连接配置文件
	Sys_superNodeEntry *sync.Map               //本地保存的超级节点地址列表
	loadAddrFuncs      []func([]byte) []string //保存不同渠道获得超级节点地址方法
	levelDB            *utilsleveldb.LevelDB   //
}

func NewAddrManager() *AddrManager {
	am := AddrManager{
		discoverEntryPath:  filepath.Join(gconfig.Path_configDir, "nodeEntry.json"),
		p2pMaxConnPath:     filepath.Join(gconfig.Path_configDir, "config.json"),
		Sys_superNodeEntry: new(sync.Map),
		loadAddrFuncs:      make([]func([]byte) []string, 0),
		// levelDB:            levelDB,
	}
	am.RegisterFunc()
	return &am
}

/*
注册加载地址方法
*/
func (this *AddrManager) RegisterFunc() {
	this.registerFunc(this.loadByDB)
	this.registerFunc(this.LoadSuperPeerEntry)

}

/*
设置发现节点地址列表文件路径
*/
func (this *AddrManager) SetDiscoverEntryPath(dePath string) {
	this.discoverEntryPath = dePath
}

/*
设置数据库
*/
func (this *AddrManager) SetLevelDB(levelDB *utilsleveldb.LevelDB) {
	this.levelDB = levelDB
}

/*
从所有渠道加载超级节点地址列表
@return    []string    ip地址列表
@return    []string    域名列表
*/
func (this *AddrManager) LoadAddrForAll(areaName []byte) ([]string, []string) {
	addrAll := make([]string, 0)
	dnsAll := make([]string, 0)
	//加载设置的地址
	this.Sys_superNodeEntry.Range(func(k, v interface{}) bool {
		addrOne, ok := k.(string)
		if !ok {
			return false
		}

		key, ok := v.([]byte)
		if !ok {
			return false
		}
		if !bytes.Equal(areaName, key) {
			return false
		}

		utils.Log.Info().Msgf("======> LoadAddrForAll %s", addrOne)
		addrAll = append(addrAll, addrOne)
		return true
	})

	//	//加载本地文件
	//	//官网获取
	//	//私网获取
	//	//局域网组播获取
	//	LoadByMulticast()

	for _, one := range this.loadAddrFuncs {
		addrOne := one(areaName)
		for _, two := range addrOne {
			if IsDNS(two) {
				dnsAll = append(dnsAll, two)
			} else {
				addrAll = append(addrAll, two)
			}
		}
		// addrAll = append(addrAll, addrOne...)
	}

	return addrAll, dnsAll
}

/*
添加一个获得超级节点地址方法
*/
func (this *AddrManager) registerFunc(newFunc func([]byte) []string) {
	this.loadAddrFuncs = append(this.loadAddrFuncs, newFunc)
}

/*
添加一个地址
*/
func (this *AddrManager) AddSuperPeerAddr(areaName []byte, addr string) {
	this.Sys_superNodeEntry.Store(addr, areaName)
}

/*
保存超级节点地址列表到本地配置文件
@path  保存到本地的磁盘路径
*/
func (this *AddrManager) saveSuperPeerEntry(path string) {
	fileBytes, _ := json.Marshal(this.Sys_superNodeEntry)
	file, _ := os.Create(path)
	file.Write(fileBytes)
	file.Close()
}

/*
解析DNS
*/
func AnalysisDNS(dnsName string, timeout time.Duration) (addr string) {
	// utils.Log.Info().Msgf("dns:%s", dnsName)
	//检查是否是DNS
	if !IsDNS(dnsName) {
		return dnsName
	}

	//设置1秒钟超时
	conn, err := net.DialTimeout("tcp", dnsName, timeout)
	if err != nil {
		// utils.Log.Info().Msgf("dns 222:%s", addr)
		return
	}
	addr = conn.RemoteAddr().String()
	// utils.Log.Info().Msgf("dns 333:%s", addr)
	conn.Close()
	return
}

const dnsRegexp_old = `[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z]{0,62})\.?`

// dns判定是否正确的正则表达式
const dnsRegexp = `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`

/*
正则判断地址是否是域名
*/
func IsDNS(dnsName string) bool {
	isOk, _ := regexp.MatchString(dnsRegexp, dnsName)
	return isOk
}
