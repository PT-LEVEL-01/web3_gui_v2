package config

import (
	"time"
	"web3_gui/utils"
)

var (
	Path_leveldb = "messagecache" //leveldb保存广播消息缓存

	DBKEY_peers_entry         = []byte{1}                            //key以“1_”前缀的含义：保存节点地址列表 格式：22_[地址]
	DBKEY_version             = utils.Uint64ToBytesByBigEndian(2000) //数据版本号
	DBKEY_broadcast_msg_add   = utils.Uint64ToBytesByBigEndian(3000) //存放广播消息自增值的key
	DBKEY_broadcast_msg_clear = utils.Uint64ToBytesByBigEndian(3001) //存放广播消息最后清理值的key
	DBKEY_broadcast_msg       = utils.Uint64ToBytesByBigEndian(3002) //存放广播消息的key

	CurBroadcastAddValue   uint64 = 1
	CurBroadcastClearValue uint64 = 1
)

const (
	MAX_CLEAN_DATA_LENGTH           = 10000            // 单次最大清理数据量
	CLEAN_DATA_TIME                 = time.Minute * 30 // 每30分钟清理一次数据
	CLEAN_DATA_INTERVAL_TIME        = time.Second      // 删除的数据比较大时，清理中间的间隔时间
	CUR_VERSION              uint64 = 1                // 当前数据库版本号
)

// /*
// 	构建未导入区块的交易key标记
// 	将未导入的区块中的交易使用此key保存到数据库中作为标记，如果已经导入过区块，则删除此标记
// 	用作验证已存在的交易hash
// */
// func BuildTxNotImport(txid []byte) []byte {
// 	return append([]byte(DB_PRE_Tx_Not_Import), txid...)
// }
//
//func init() {
//	ok := utilsleveldb.RegisterDbKey(DBKEY_peers_entry)
//	if !ok {
//		panic("dbkey exist")
//	}
//	ok = utilsleveldb.RegisterDbKey(DBKEY_version)
//	if !ok {
//		panic("dbkey exist")
//	}
//	ok = utilsleveldb.RegisterDbKey(DBKEY_broadcast_msg_add)
//	if !ok {
//		panic("dbkey exist")
//	}
//	ok = utilsleveldb.RegisterDbKey(DBKEY_broadcast_msg_clear)
//	if !ok {
//		panic("dbkey exist")
//	}
//	ok = utilsleveldb.RegisterDbKey(DBKEY_broadcast_msg)
//	if !ok {
//		panic("dbkey exist")
//	}
//}

/*
设置leveldb数据库路径
*/
func SetLevelDBPath(leveldbPath string) {
	Path_leveldb = leveldbPath
}
