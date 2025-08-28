package utils

import (
	"errors"
	"log"
	"sync"

	"github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

// Strategy 策略数据库
var quant = &DBLedis{onceConnLevelDB: sync.Once{}, DB: nil}

var Quant *ledis.DB

// 初始化ledis数据库
func InitLedis(dbpath string) {
	var err error
	Quant, err = quant.InitDB(dbpath)

	if err != nil {
		log.Fatalln("Quant init err", err.Error())
	}

}

// DBConn leveldb 连接
type DBLedis struct {
	onceConnLevelDB sync.Once

	DB *ledis.DB
}

// InitDB 初始化数据库
func (db *DBLedis) InitDB(name string) (curDB *ledis.DB, err error) {

	db.onceConnLevelDB.Do(func() {
		cfg := config.NewConfigDefault()
		cfg.DataDir = name
		cfg.Databases = 1024

		//os.RemoveAll(cfg.DataDir)

		tdb, err := ledis.Open(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
		db.DB, err = tdb.Select(0)
		curDB = db.DB
		if err != nil {
			return
		}
	})
	return
}

type LedisDB struct {
	path string
	db   *ledis.DB
	once sync.Once
}

func CreateLedisDB(path string) (*LedisDB, error) {
	// golog.InitLogger("logs/log.txt", 0, true)
	// golog.Info("start %s", "log")
	// golog.Infof("start %s\n", "log")
	lldb := LedisDB{
		path: path,
		// db   *leveldb.DB
		once: sync.Once{},
	}
	err := lldb.InitDB()
	if err != nil {
		return nil, err
	}
	return &lldb, nil
}

// 链接leveldb
func (this *LedisDB) InitDB() (err error) {
	this.once.Do(func() {
		// utils.Log.Info().Msgf("这个方法执行了多少次")
		//没有db目录会自动创建
		// this.db, err = leveldb.OpenFile(this.path, nil)
		// //	defer db.Close()
		// if err != nil {
		// 	return
		// }
		// // cleanDB()
		// return

		cfg := config.NewConfigDefault()
		cfg.DataDir = this.path
		cfg.Databases = 1024

		//os.RemoveAll(cfg.DataDir)
		var tdb *ledis.Ledis
		tdb, err = ledis.Open(cfg)
		if err != nil {
			return
		}
		this.db, err = tdb.Select(0)
		// curDB = db.DB
		if err != nil {
			return
		}

	})
	return
}

/*
获取数据库连接
*/
func (this *LedisDB) GetDB() *ledis.DB {
	return this.db
}

/*
保存
*/
func (this *LedisDB) Save(id []byte, bs *[]byte) error {

	// utils.Log.Debug().Msgf("保存到leveldb %s %s", hex.EncodeToString(id), string(*bs))

	//levedb保存相同的key，原来的key保存的数据不会删除，因此保存之前先删除原来的数据
	// this.db.Del(id)
	// this.db.ZRem()
	_, err := this.db.Del(id)
	if err != nil {
		// utils.Log.Error().Msgf("Delete error while saving leveldb", err)
		return err
	}
	if bs == nil {
		err = this.db.Set(id, nil)
		// err = this.db.Put(id, nil, nil)
	} else {
		err = this.db.Set(id, *bs)
		// err = this.db.Put(id, *bs, nil)
	}
	// if err != nil {
	// 	// utils.Log.Error().Msgf("Leveldb save error", err)
	// }
	return err
}

/*
查找
*/
func (this *LedisDB) Find(txId []byte) (*[]byte, error) {
	// this.db.Get()
	value, err := this.db.Get(txId)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

/*
删除
*/
func (this *LedisDB) Remove(id []byte) error {
	_, err := this.db.Del(id)
	return err
}

/*
	初始化数据库的时候，清空一些数据
*/
// func (this *LedisDB) cleanDB(name string) {
// 	this.db.z
// 	// _, err := this.Tags([]byte(config.Name))
// 	_, err := this.Tags([]byte(name))
// 	if err == nil {
// 		// for _, one := range keys {
// 		// 	utils.Log.Info().Msgf("开始删除域名 %s", hex.EncodeToString(one))
// 		// 	err = Remove(one)
// 		// 	if err != nil {
// 		// 		utils.Log.Info().Msgf("删除错误 %s", err.Error())
// 		// 	}
// 		// }
// 		// for _, one := range keys {
// 		// 	value, _ := Find(one)
// 		// 	if value != nil {
// 		// 		utils.Log.Info().Msgf("查找域名 %s", hex.EncodeToString(one))

// 		// 	}

// 		// }
// 	}
// 	// utils.Log.Info().Msgf("删除域名 end")

// 	// db.
// }

// 根据Tags遍历
// func (this *LedisDB) Tags(tag []byte) ([][]byte, error) {
// 	// keys := make([][]byte, 0)
// 	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
// 	iter := this.db.NewIterator(nil, nil)
// 	for iter.Next() {
// 		if bytes.HasPrefix(iter.Key(), tag) {
// 			// utils.Log.Info().Msgf("匹配的 %s", iter.Key())
// 			// keys = append(keys, iter.Key())
// 			this.db.Delete(iter.Key(), nil)
// 		}
// 	}
// 	iter.Release()
// 	err := iter.Error()
// 	return nil, err
// }

/*
	打印所有key
*/
// func (this *LedisDB) PrintAll() ([][]byte, error) {
// 	// keys := make([][]byte, 0)
// 	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
// 	iter := this.db.NewIterator(nil, nil)
// 	for iter.Next() {
// 		// utils.Log.Info().Msgf("%s", hex.EncodeToString(iter.Key()))
// 		// fmt.Println(hex.EncodeToString(iter.Key()))
// 		fmt.Println("key", hex.EncodeToString(iter.Key()), "value", hex.EncodeToString(iter.Value()))
// 	}
// 	iter.Release()
// 	err := iter.Error()
// 	return nil, err
// }

/*
	查询指定前缀的key
*/
// func (this *LedisDB) FindPrefixKeyAll(tag []byte) ([][]byte, [][]byte, error) {
// 	keys := make([][]byte, 0)
// 	values := make([][]byte, 0)
// 	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
// 	iter := this.db.NewIterator(nil, nil)
// 	for iter.Next() {
// 		if bytes.HasPrefix(iter.Key(), tag) {
// 			// utils.Log.Info().Msgf("匹配的 %s", iter.Key())
// 			// utils.Log.Info().Msgf("匹配的 %s", iter.Value())
// 			keys = append(keys, iter.Key())
// 			// db.Delete(iter.Key(), nil)
// 			value, err := this.db.Get(iter.Key(), nil)
// 			if err != nil {
// 				return nil, nil, err
// 			}
// 			values = append(values, value)
// 			// utils.Log.Info().Msgf("查询的结果 %s", value)
// 		}
// 	}
// 	iter.Release()
// 	err := iter.Error()
// 	return keys, values, err
// }

/*
检查是否是空数据库
*/
func (this *LedisDB) CheckNullDB(key []byte) (bool, error) {
	n, err := this.db.Exists(key)
	if err != nil {
		return false, err
	}
	if n <= 0 {
		return true, nil
	}
	return false, nil
	// // _, err := this.Find(config.Key_block_start)
	// _, err := this.Find(key)
	// if err != nil {
	// 	if err == ledis.ErrNotFound {
	// 		//认为这是一个空数据库
	// 		return true, nil
	// 	}
	// 	return false, err
	// }
	// return false, nil
}

/*
检查key是否存在
@return    bool    true:存在;false:不存在;
*/
func (this *LedisDB) CheckHashExist(hash []byte) (bool, error) {
	n, err := this.db.Exists(hash)
	if err != nil {
		return false, err
	}
	if n <= 0 {
		return false, nil
	}
	return true, nil

	// fmt.Println(hex.EncodeToString(hash))
	// _, err := this.Find(hash)
	// if err != nil {
	// 	if err == leveldb.ErrNotFound {
	// 		// fmt.Println("db 没找到")
	// 		// utils.Log.Debug().Msgf("db 没找到 %s", hex.EncodeToString(hash))
	// 		return false
	// 	}
	// 	// fmt.Println("db 错误")
	// 	// utils.Log.Debug().Msgf("db 错误 %s", hex.EncodeToString(hash))
	// 	return true
	// }
	// // fmt.Println("db 找到了")
	// // utils.Log.Debug().Msgf("db 找到了 %s", hex.EncodeToString(hash))
	// return true
}

/*
添加一个数据到zset
*/
func (this *LedisDB) AddZSet(name, member *[]byte, orderId int64) error {
	sp := ledis.ScorePair{
		Score:  orderId,
		Member: *member,
	}
	n, err := this.db.ZAdd(*name, sp)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("add failt")
	}
	return nil
}

/*
添加一个数据到zset，id为自增长
*/
func (this *LedisDB) AddZSetAutoincrId(name *[]byte, member []byte, duration int64) error {
	n, err := this.db.ZCard(*name)
	if err != nil {
		return err
	}
	sp := ledis.ScorePair{
		Score:  n + 1,
		Member: member,
	}
	n, err = this.db.ZAdd(*name, sp)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("add failt")
	}
	this.db.Expire(*name, duration) //设置过期时间
	return nil
}

/*
获取一个member的下标，如果不存在，则返回-1
*/
func (this *LedisDB) GetZSetByMember(name *[]byte, member []byte) (int64, error) {
	return this.db.ZRank(*name, member)
}

/*
添加一个数据到zset
*/
func (this *LedisDB) GetZSetPage(name *[]byte, min, max int64, count int) (*[]ledis.ScorePair, error) {
	//从大到小开始查询
	sps, err := this.db.ZRangeByScoreGeneric(*name, min, max, 0, count, false)
	if err != nil {
		return nil, err
	}
	return &sps, err
}

/*
添加一个数据到zset
*/
func (this *LedisDB) GetZSetAll(name *[]byte) (*[]ledis.ScorePair, error) {
	n, err := this.db.ZCard(*name)
	if err != nil {
		return nil, err
	}
	//
	sps, err := this.db.ZRangeByScoreGeneric(*name, 0, n, 0, int(n), true)
	if err != nil {
		return nil, err
	}
	return &sps, err
}

/*
删除zset中一定范围score的数据
*/
func (this *LedisDB) DelZSet(name *[]byte, min, max int64) error {
	_, err := this.db.ZRemRangeByScore(*name, min, max)
	return err
}

/*
删除zset中多个members
*/
func (this *LedisDB) DelZSetByMembers(name *[]byte, members ...[]byte) error {
	_, err := this.db.ZRem(*name, members...)
	return err
}

/*
添加一个数据到zset
*/
func (this *LedisDB) DelZSetAll(name *[]byte) error {
	_, err := this.db.ZClear(*name)
	return err
}

/*
 * 查询根据 start stop 查询返回zset
 */
func (ldb *LedisDB) GetRangeZSet(name *[]byte, start, stop int) (*[]ledis.ScorePair, error) {
	rel, err := ldb.db.ZRange(*name, start, stop)
	return &rel, err
}

/*
 * 设置key 过期时间 单位秒
 */
func (ldb *LedisDB) SetExpire(key *[]byte, duration int64) (int64, error) {
	return ldb.db.Expire(*key, duration)
}

/**
 * 添加一个List
 */
func (ldb *LedisDB) Rpush(key, val []byte) (int64, error) {
	return ldb.db.RPush(key, val)
}

/**
 * 获取List
 */
func (ldb *LedisDB) LRange(key []byte, start, end int32) ([][]byte, error) {
	return ldb.db.LRange(key, start, end)
}

/*
得到一个数据到zset，数量
*/
func (this *LedisDB) ZSetCard(name *[]byte) (n int64, err error) {
	n, err = this.db.ZCard(*name)

	if err != nil {
		return 0, nil
	}
	return
}
