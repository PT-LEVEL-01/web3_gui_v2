package storage

import (
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb"
	"path/filepath"
	"sync"
	"web3_gui/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
数据库集合
*/
type DBCollections struct {
	lock   *sync.RWMutex //
	FullDB []*DBone      //已经存满的数据库
	FreeDB []*DBone      //有闲置空间的数据库
}

func CreateDBCollections() *DBCollections {
	dbc := new(DBCollections)
	dbc.lock = new(sync.RWMutex)
	dbc.FullDB = make([]*DBone, 0)
	dbc.FreeDB = make([]*DBone, 0)
	return dbc
}

/*
添加一个磁盘，为数据库提供存储空间
*/
func (this *DBCollections) AddDBone(paths ...string) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	var err error
	//判断是否有重复磁盘路径
	for i, one := range paths {
		abs := ""
		abs, err = filepath.Abs(one)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		paths[i] = abs
		for _, fullDBone := range this.FullDB {
			_, err = filepath.Rel(fullDBone.leveldb.GetPath(), abs)
			if err == nil {
				return utils.NewErrorBus(config.ERROR_CODE_storage_db_path_Homologous, "") // config.ERROR_db_path_Homologous
			}
		}
		for _, fullDBone := range this.FreeDB {
			_, err = filepath.Rel(fullDBone.leveldb.GetPath(), abs)
			if err == nil {
				return utils.NewErrorBus(config.ERROR_CODE_storage_db_path_Homologous, "")
			}
		}
		//
		for j, two := range paths {
			if i == j {
				continue
			}
			_, err = filepath.Rel(two, abs)
			if err == nil {
				return utils.NewErrorBus(config.ERROR_CODE_storage_db_path_Homologous, "")
			}
		}
	}
	dbs := make([]*DBone, 0, len(paths))
	for _, one := range paths {
		dbone, err := NewDBone(one)
		if err != nil {
			//回滚，将之前创建的数据库链接关闭
			for _, dbOne := range dbs {
				dbOne.Close()
			}
			return utils.NewErrorSysSelf(err)
		}
		dbs = append(dbs, dbone)
	}
	this.FreeDB = append(this.FreeDB, dbs...)
	return utils.NewErrorSuccess()
}

/*
减少一个磁盘，为数据库减少存储空间
*/
func (this *DBCollections) DelDBone(paths string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	for i, one := range this.FreeDB {
		if one.leveldb.GetPath() == paths {
			one.Close()
			temp := this.FreeDB[:i]
			temp = append(temp, this.FreeDB[i+1:]...)
			this.FreeDB = temp
			return nil
		}
	}
	for i, one := range this.FullDB {
		if one.leveldb.GetPath() == paths {
			one.Close()
			temp := this.FullDB[:i]
			temp = append(temp, this.FullDB[i+1:]...)
			this.FreeDB = temp
			return nil
		}
	}
	return nil
}

/*
保存文件块
@number    []byte    文件块的hash
@bs        []byte    块文件内容
*/
func (this *DBCollections) SaveChunk(number []byte, bs []byte) utils.ERROR {
	//numberStr := hex.EncodeToString(number)
	//utils.Log.Info().Msgf("保存块数据:%s %d", numberStr, len(bs))
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	for {
		//utils.Log.Info().Msgf("打印数据库数量:%d %d", len(this.FreeDB), len(this.FullDB))
		//有错误，则放到已满列表
		//this.lock.Lock()
		if len(this.FreeDB) == 0 {
			//this.lock.Unlock()
			return utils.NewErrorBus(config.ERROR_CODE_storage_db_full, "")
		}
		ERR = this.FreeDB[0].leveldb.SaveMap(*config.DBKEY_storage_server_file_chunk, *dbkey, bs, nil)
		if ERR.CheckSuccess() {
			return ERR
		}
		//放入已满列表
		this.FullDB = append(this.FullDB, this.FreeDB[0])
		this.FreeDB = this.FreeDB[1:]
		//this.lock.Unlock()
	}
	//utils.Log.Info().Msgf("保存块数据:%s %d", numberStr, len(bs))
	return utils.NewErrorSuccess()
}

/*
删除文件块
*/
func (this *DBCollections) DelChunk(number []byte) utils.ERROR {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	var err error
	this.lock.RLock()
	for _, one := range append(this.FreeDB, this.FullDB...) {
		err = one.leveldb.RemoveMapByKey(*config.DBKEY_storage_server_file_chunk, *dbkey, nil)
		if err != nil {
			break
		}
	}
	this.lock.RUnlock()
	return utils.NewErrorSysSelf(err)
}

func (this *DBCollections) FindChunk(number []byte) (*[]byte, utils.ERROR) {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(number)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	this.lock.RLock()
	defer this.lock.RUnlock()
	//utils.Log.Info().Msgf("查询文件块:%s", hex.EncodeToString(number))
	for _, one := range this.FullDB {
		bs, err := one.leveldb.FindMap(*config.DBKEY_storage_server_file_chunk, *dbkey)
		if err != nil {
			if err.Error() == leveldb.ErrNotFound.Error() {
				utils.Log.Info().Msgf("查询文件块:%s 未找到", hex.EncodeToString(number))
			}
			continue
		}
		if bs == nil {
			continue
		}
		return &bs, utils.NewErrorSuccess()
	}
	for _, one := range this.FreeDB {
		bs, err := one.leveldb.FindMap(*config.DBKEY_storage_server_file_chunk, *dbkey)
		if err != nil {
			if err.Error() == leveldb.ErrNotFound.Error() {
				utils.Log.Info().Msgf("查询文件块:%s 未找到", hex.EncodeToString(number))
			}
			continue
		}
		if bs == nil {
			continue
		}
		return &bs, utils.NewErrorSuccess()
	}
	return nil, utils.NewErrorSuccess()
}

/*
一个数据库连接
*/
type DBone struct {
	leveldb *utilsleveldb.LevelDB //
	Full    bool                  //本数据库所在磁盘已满
	Stope   bool                  //手动暂停存储数据
}

func NewDBone(dbpath string) (*DBone, error) {
	utils.Log.Info().Msgf("创建一个数据库连接 %s", dbpath)
	db, err := utilsleveldb.CreateLevelDB(dbpath)
	if err != nil {
		return nil, err
	}
	dbone := DBone{
		leveldb: db,
		Full:    false,
		Stope:   false,
	}
	utils.Log.Info().Msgf("创建成功")
	return &dbone, nil
}

func (this *DBone) Close() {
	this.leveldb.Close()
}
