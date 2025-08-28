package config

import (
	"time"
	"web3_gui/utils/utilsleveldb"
)

var (
	Path_leveldb = "messagecache" //leveldb保存广播消息缓存

	DBKEY_addr_list           = utilsleveldb.RegDbKeyExistPanicByUint64(2001) //保存ip地址和端口的字符串列表
	DBKEY_version             = utilsleveldb.RegDbKeyExistPanicByUint64(2002) //数据版本号
	DBKEY_broadcast_msg_add   = utilsleveldb.RegDbKeyExistPanicByUint64(2003) //存放广播消息自增值的key
	DBKEY_broadcast_msg_clear = utilsleveldb.RegDbKeyExistPanicByUint64(2004) //存放广播消息最后清理值的key
	DBKEY_broadcast_msg       = utilsleveldb.RegDbKeyExistPanicByUint64(2005) //存放广播消息的key

	CurBroadcastAddValue   uint64 = 1
	CurBroadcastClearValue uint64 = 1
)

const (
	MAX_CLEAN_DATA_LENGTH           = 10000            // 单次最大清理数据量
	CLEAN_DATA_TIME                 = time.Minute * 30 // 每30分钟清理一次数据
	CLEAN_DATA_INTERVAL_TIME        = time.Second      // 删除的数据比较大时，清理中间的间隔时间
	CUR_VERSION              uint64 = 1                // 当前数据库版本号
)

/*
设置leveldb数据库路径
*/
func SetLevelDBPath(leveldbPath string) {
	Path_leveldb = leveldbPath
}
