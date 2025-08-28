package addr_manager

import (
	ma "github.com/multiformats/go-multiaddr"
	"sync"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
地址管理
*/
type AddrManager struct {
	lock           *sync.RWMutex           //
	superNodeEntry map[string]ma.Multiaddr //本地保存的超级节点地址列表
	loadAddrFuncs  []func() []ma.Multiaddr //保存不同渠道获得超级节点地址方法
	levelDB        *utilsleveldb.LevelDB   //
}

func NewAddrManager() *AddrManager {
	am := AddrManager{
		lock:           new(sync.RWMutex),
		superNodeEntry: make(map[string]ma.Multiaddr),
		loadAddrFuncs:  make([]func() []ma.Multiaddr, 0),
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
	//this.registerFunc(this.LoadSuperPeerEntry)
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
func (this *AddrManager) LoadAddrForAll() ([]ma.Multiaddr, []ma.Multiaddr) {
	addrAll := make([]ma.Multiaddr, 0)
	dnsAll := make([]ma.Multiaddr, 0)
	//加载设置的地址
	this.lock.RLock()
	for _, addr := range this.superNodeEntry {
		//utils.Log.Info().Str("ip", key).Send()
		info, ERR := engine.CheckAddr(addr)
		if ERR.CheckFail() {
			continue
		}
		if info.IsDNS {
			dnsAll = append(dnsAll, addr)
		} else {
			addrAll = append(addrAll, addr)
		}
	}
	this.lock.RUnlock()

	//	//加载本地文件
	//	//官网获取
	//	//私网获取
	//	//局域网组播获取
	//	LoadByMulticast()

	for _, one := range this.loadAddrFuncs {
		addrOne := one()
		for _, two := range addrOne {
			info, ERR := engine.CheckAddr(two)
			if ERR.CheckFail() {
				continue
			}
			if info.IsDNS {
				dnsAll = append(dnsAll, two)
			} else {
				addrAll = append(addrAll, two)
			}
		}
	}
	return addrAll, dnsAll
}

/*
添加一个获得超级节点地址方法
*/
func (this *AddrManager) registerFunc(newFunc func() []ma.Multiaddr) {
	this.loadAddrFuncs = append(this.loadAddrFuncs, newFunc)
}

/*
添加一个地址
*/
func (this *AddrManager) AddSuperPeerAddr(addr ma.Multiaddr) utils.ERROR {
	_, ERR := engine.CheckAddr(addr)
	if ERR.CheckFail() {
		return ERR
	}
	this.lock.Lock()
	this.superNodeEntry[addr.String()] = addr
	this.lock.Unlock()
	return ERR
}

///*
//解析DNS
//*/
//func AnalysisDNS(dnsName string, timeout time.Duration) (string, error) {
//	// utils.Log.Info().Msgf("dns:%s", dnsName)
//	//检查是否是DNS
//	if !IsDNS(dnsName) {
//		return dnsName
//	}
//
//	//设置1秒钟超时
//	conn, err := net.DialTimeout("tcp", dnsName, timeout)
//	if err != nil {
//		// utils.Log.Info().Msgf("dns 222:%s", addr)
//		return
//	}
//	addr = conn.RemoteAddr().String()
//	// utils.Log.Info().Msgf("dns 333:%s", addr)
//	conn.Close()
//	return
//}
