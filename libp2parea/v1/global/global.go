package global

import (
	"sync"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type Global struct {
	engine      *engine.Engine        // 通用消息engine
	StartEngine bool                  // 是否启动过engine
	levelDB     *utilsleveldb.LevelDB // leveldb
	Area        map[string]*AreaInfo  // 保存所有区域key:string=;value:*Area=;
	areaLock    *sync.Mutex           // 区域名锁
	Addr        string                // 监听IP地址
	Port        uint16                // 监听端口
}

/*
 * NewGlobal 创建域名管理器
 *
 * @param	key		keystore	keystore
 * @param	pwd		string		密码
 * @return	global	*Global		域名管理器
 */
func NewGlobal(key keystore.KeystoreInterface, pwd string) *Global {
	_, puk, err := key.GetNetAddr(pwd)
	if err != nil && err.Error() == keystore.ERROR_address_empty.Error() {
		_, puk, err = key.CreateNetAddr(pwd, pwd)
	}
	if err != nil {
		// utils.Log.Info().Msgf("GetNetAddr error:%s", err.Error())
		return nil
	}
	addrNet := nodeStore.BuildAddr(puk)

	sessionEngine := engine.NewEngine(addrNet.B58String())

	return &Global{
		Area:     make(map[string]*AreaInfo),
		areaLock: new(sync.Mutex),
		engine:   sessionEngine,
	}
}

/*
 * 获取监听的地址和端口
 */
func (this *Global) GetTcpHost() (string, uint16) {
	return this.Addr, this.Port
}

/*
 * 获取消息引擎
 */
func (this *Global) GetEngine() *engine.Engine {
	return this.engine
}

/*
 * 获取leveldb信息
 *
 * @return levelDB	leveldb指针
 */
func (this *Global) GetLevelDB() *utilsleveldb.LevelDB {
	return this.levelDB
}

/*
 * 设置leveldb信息
 *
 * @param levelDB 	leveldb指针
 */
func (this *Global) SetLevelDB(levelDB *utilsleveldb.LevelDB) {
	this.levelDB = levelDB
}

type AreaInfo struct {
	AreaName string //域名称
	TcpAddr  string //TCP地址
	TcpPort  uint32 //TCP端口
}

/*
 * CheckOrAddAreaInfo 检查地址信息是否存在，不存在则添加相关记录
 *
 * @param	areaName	string	域名名称
 * @param	ipAddr		string	ip地址
 * @param	port		string	监听端口
 * @return	success		bool	是否添加成功
 */
func (this *Global) CheckOrAddAreaInfo(areaName string) bool {
	// 加锁，防止多协程操作出现异常
	this.areaLock.Lock()
	defer this.areaLock.Unlock()

	// 检查之前是否存在相同域名的记录信息
	_, exit := this.Area[areaName]
	if exit {
		// 存在，则不允许再添加，每个global针对同一种域名只能添加一个
		return true
	}

	// 添加域名记录信息
	this.Area[areaName] = &AreaInfo{
		AreaName: areaName,
		TcpAddr:  "",
		TcpPort:  0,
	}

	// 之前不存在，已添加记录信息
	return false
}

/*
 * RemoveAreaInfo 移除域名对应的记录信息
 *
 * @param	areaName	string	要移除的域名
 */
func (this *Global) RemoveAreaInfo(areaName []byte) {
	this.areaLock.Lock()
	defer this.areaLock.Unlock()

	// 删除区域名称对应的记录信息
	strAreaName := utils.Bytes2string(areaName)
	delete(this.Area, strAreaName)

	// // 不存在域名信息时, 将关闭leveldb、清空的连接信息
	// if len(this.Area) == 0 {
	// 	// 关闭leveldb
	// 	if this.levelDB != nil {
	// 		this.levelDB.Close()
	// 		this.levelDB = nil
	// 	}
	// }
}
