package file_transfer

import (
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var genIDLock = new(sync.Mutex)
var genID *big.Int
var genIDOnece = new(sync.Once)

/*
获取自增长ID
*/
func GetGenID(db *utilsleveldb.LevelDB) ([]byte, utils.ERROR) {
	var err error
	genIDOnece.Do(func() {
		//初始化自增长ID到内存
		var number []byte
		number, err = getGenID(db)
		if err != nil {
			return
		}
		genID = new(big.Int).SetBytes(number)
	})
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	genIDLock.Lock()
	defer genIDLock.Unlock()
	temp := new(big.Int).Add(genID, big.NewInt(1))
	ERR := saveGenID(db, temp.Bytes())
	if err != nil {
		return nil, ERR
	}
	genID = temp
	return temp.Bytes(), utils.NewErrorSuccess()
}

/*
获取数据库中全局自增长ID
*/
func getGenID(db *utilsleveldb.LevelDB) ([]byte, error) {
	dbItem, err := db.Find(*DBKEY_GenID)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return []byte{0}, nil
		}
		return nil, err
	}
	return dbItem.Value, nil
}

/*
保存数据库中全局自增长ID
*/
func saveGenID(db *utilsleveldb.LevelDB, number []byte) utils.ERROR {
	return db.Save(*DBKEY_GenID, &number)
}

/*
保存白名单列表
*/
func saveWhiteList(db *utilsleveldb.LevelDB, classID uint64, list []*nodeStore.AddressNet) utils.ERROR {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return ERR
	}
	newList := make([][]byte, 0, len(list))
	for _, one := range list {
		newList = append(newList, one.GetAddr())
	}
	bs, err := model.BytesProto(newList, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return db.SaveMap(*DBKEY_white_list, *dbkey, bs, nil)
}

/*
获取白名单列表
*/
func getWhiteList(db *utilsleveldb.LevelDB, classID uint64) ([]*nodeStore.AddressNet, utils.ERROR) {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := db.FindMap(*DBKEY_white_list, *dbkey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	list, _, err := model.ParseBytes(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	addrs := make([]*nodeStore.AddressNet, 0, len(list))
	for _, one := range list {
		addrOne := nodeStore.NewAddressNet(one) // nodeStore.AddressNet(one)
		addrs = append(addrs, addrOne)
	}
	return addrs, utils.NewErrorSuccess()
}

/*
保存共享文件夹名单
*/
func saveShareDirs(db *utilsleveldb.LevelDB, classID uint64, dirs []string) utils.ERROR {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return ERR
	}
	newList := make([][]byte, 0, len(dirs))
	for _, one := range dirs {
		newList = append(newList, []byte(one))
	}
	bs, err := model.BytesProto(newList, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return db.SaveMap(*DBKEY_share_dir, *dbkey, bs, nil)
}

/*
获取共享文件夹名单
*/
func getShareDirs(db *utilsleveldb.LevelDB, classID uint64) ([]string, utils.ERROR) {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := db.FindMap(*DBKEY_share_dir, *dbkey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	list, _, err := model.ParseBytes(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	dirs := make([]string, 0, len(list))
	for _, one := range list {
		dirs = append(dirs, string(one))
	}
	return dirs, utils.NewErrorSuccess()
}

/*
保存是否自动接收文件
*/
func saveAutoReceive(db *utilsleveldb.LevelDB, classID uint64, auto uint32) utils.ERROR {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs := utils.Uint32ToBytes(auto)
	return db.SaveMap(*DBKEY_auto_Receive, *dbkey, bs, nil)
}

/*
获取是否自动接收文件
*/
func getAutoReceive(db *utilsleveldb.LevelDB, classID uint64) (uint32, utils.ERROR) {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytes(classID))
	if !ERR.CheckSuccess() {
		return 0, ERR
	}
	bs, err := db.FindMap(*DBKEY_auto_Receive, *dbkey)
	if err != nil {
		return 0, utils.NewErrorSysSelf(err)
	}
	auto := utils.BytesToUint32(bs)
	return auto, utils.NewErrorSuccess()
}
