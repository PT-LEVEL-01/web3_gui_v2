package utils

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	// DBKeySize              = 8                  //dbkey为uint64大端序列化的[]byte，避免查询时与短前缀混合，规定固定长度为8
	MaxKeySize           = 65535              //max key size
	MaxValueSize     int = 1024 * 1024 * 1024 //max value size
	DBRemoveNum          = 100                //分批次删除map集合中的数据，每批次默认数量
	DBRemoveInterval     = time.Second        //分批次删除map集合中的数据，每批次默认间隔时间
	// DBType_map_ley_len     = 2                  //Map结构key长度需要2字节保存
)

var (
	// errDBKeySize         = errors.New("dbkey length not equal to 8")
	// errIndexSize         = errors.New("index length not equal to 8")
	// errDBKeyNotEqualZero = errors.New("invalid dbkey byte not equal 0")
	errKeySize = errors.New("invalid key size")
	// errKeyNotEqualZero = errors.New("invalid key byte not equal 0")
	// errKeyNotEqualFull = errors.New("invalid key byte not equal 255")
	errValueSize = errors.New("invalid value size")

	// IndexZero = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	// IndexFull = []byte{255, 255, 255, 255, 255, 255, 255, 255}

	DataType_Index      = []byte{1} //保存自增长ID
	DataType_Data_Map   = []byte{2} //Map结构数据保存
	DataType_Data_List  = []byte{3} //List结构数据保存
	DataType_Map_In_Map = []byte{4} //map in the map结构
)

type leveldbKey []byte

func NewLeveldbKey(key []byte) (leveldbKey, error) {
	var err error
	key, err = LeveldbBuildKey(key)
	if err != nil {
		return nil, err
	}
	return leveldbKey(key), nil
}

/*
leveldb中，所有的key都带一个key长度前缀，用2字节保存。
2字节容量是65535，因此key长度不能大于65535。
避免不同长度的key前缀查询的时候混合。
*/
func LeveldbBuildKey(key []byte) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}
	keyLenBs := Uint16ToBytesByBigEndian(uint16(len(key)))
	return append(keyLenBs, key...), nil
}

/*
leveldb中，所有的key都带一个key长度前缀，用2字节保存。
2字节容量是65535，因此key长度不能大于65535。
避免不同长度的key前缀查询的时候混合。
@return    uint16    长度
@return    []byte    解析出来的key
*/
func LeveldbParseKey(key []byte) (uint16, []byte, error) {
	if len(key) < 2 {
		return 0, nil, errKeySize
	}
	length := BytesToUint16ByBigEndian(key[:2])
	if len(key) < int(length+2) {
		return 0, nil, errKeySize
	}
	return length, key[2 : 2+length], nil
}

/*
多个长度+[]byte拼接，通过长度解析出来多个数据。不能有不符合规范的数据，有则报错
*/
func LeveldbParseKeyMore(key []byte) ([][]byte, error) {
	keys := make([][]byte, 0)
	index := 0
	length := len(key)
	for {
		l, k, err := LeveldbParseKey(key[index:])
		if err != nil {
			return keys, err
		}
		keys = append(keys, k)
		index += int(l)
		if index == length {
			break
		}
	}
	return keys, nil
}

/*
检查DBKey长度
*/
// func checkBDKeySize(dbkey []byte) error {
// 	if len(dbkey) != DBKeySize {
// 		return errDBKeySize
// 	}
// 	return nil
// }

/*
检查Index长度
*/
// func checkIndexSize(index []byte) error {
// 	if len(index) != DBKeySize {
// 		return errIndexSize
// 	}
// 	return nil
// }

/*
检查Key长度
*/
func checkKeySize(key []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	}
	return nil
}

/*
检查Value长度
*/
func checkValueSize(value []byte) error {
	if len(value) > MaxValueSize {
		return errValueSize
	}
	return nil
}

type LevelDB struct {
	path string
	db   *leveldb.DB
	once sync.Once
}

func CreateLevelDB(path string) (*LevelDB, error) {
	lldb := LevelDB{
		path: path,
		once: sync.Once{},
	}
	err := lldb.InitDB()
	if err != nil {
		return nil, err
	}
	return &lldb, nil
}

// 链接leveldb
func (this *LevelDB) InitDB() (err error) {
	this.once.Do(func() {
		//没有db目录会自动创建
		this.db, err = leveldb.OpenFile(this.path, nil)
		if err != nil {
			return
		}
		return
	})
	return
}

/*
保存
*/
func (this *LevelDB) Save(id []byte, bs *[]byte) error {
	var err error
	if id, err = LeveldbBuildKey(id); err != nil {
		return err
	}
	if err = checkValueSize(*bs); err != nil {
		return err
	}
	// utils.Log.Debug().Msgf("保存到leveldb %s %s", hex.EncodeToString(id), string(*bs))
	//levedb保存相同的key，原来的key保存的数据不会删除，因此保存之前先删除原来的数据
	err = this.db.Delete(id, nil)
	if err != nil {
		// utils.Log.Error().Msgf("Delete error while saving leveldb", err)
		return err
	}
	if bs == nil {
		err = this.db.Put(id, nil, nil)
	} else {
		err = this.db.Put(id, *bs, nil)
	}
	return err
}

/*
查找
*/
func (this *LevelDB) Find(txId []byte) (*[]byte, error) {
	var err error
	if txId, err = LeveldbBuildKey(txId); err != nil {
		return nil, err
	}
	value, err := this.db.Get(txId, nil)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

/*
删除
*/
func (this *LevelDB) Remove(id []byte) error {
	var err error
	if id, err = LeveldbBuildKey(id); err != nil {
		return err
	}
	return this.db.Delete(id, nil)
}

/*
初始化数据库的时候，清空一些数据
*/
func (this *LevelDB) cleanDB(name string) {
	_, err := this.Tags([]byte(name))
	if err == nil {
		// for _, one := range keys {
		// 	utils.Log.Info().Msgf("开始删除域名 %s", hex.EncodeToString(one))
		// 	err = Remove(one)
		// 	if err != nil {
		// 		utils.Log.Info().Msgf("删除错误 %s", err.Error())
		// 	}
		// }
		// for _, one := range keys {
		// 	value, _ := Find(one)
		// 	if value != nil {
		// 		utils.Log.Info().Msgf("查找域名 %s", hex.EncodeToString(one))
		// 	}
		// }
	}
}

// 根据Tags遍历
func (this *LevelDB) Tags(tag []byte) ([][]byte, error) {
	var err error
	if tag, err = LeveldbBuildKey(tag); err != nil {
		return nil, err
	}
	// keys := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), tag) {
			// utils.Log.Info().Msgf("匹配的 %s", iter.Key())
			// keys = append(keys, iter.Key())
			this.db.Delete(iter.Key(), nil)
		}
	}
	iter.Release()
	err = iter.Error()
	return nil, err
}

/*
打印所有key
*/
func (this *LevelDB) PrintAll() ([][]byte, error) {
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		// fmt.Println("key", hex.EncodeToString(iter.Key()), "value", hex.EncodeToString(iter.Value()))
		fmt.Println("key", iter.Key(), "value", iter.Value())
	}
	iter.Release()
	err := iter.Error()
	return nil, err
}

/*
查询指定前缀的key
*/
func (this *LevelDB) FindPrefixKeyAll(tag []byte) ([][]byte, [][]byte, error) {
	var err error
	if tag, err = LeveldbBuildKey(tag); err != nil {
		return nil, nil, err
	}
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := this.db.NewIterator(util.BytesPrefix(tag), nil)
	for iter.Next() {
		// utils.Log.Info().Msgf("匹配的 %s", iter.Key())
		// utils.Log.Info().Msgf("匹配的 %s", iter.Value())
		keys = append(keys, iter.Key())
		// db.Delete(iter.Key(), nil)
		value, err := this.db.Get(iter.Key(), nil)
		if err != nil {
			return nil, nil, err
		}
		values = append(values, value)
		// utils.Log.Info().Msgf("查询的结果 %s", value)
	}
	iter.Release()
	err = iter.Error()
	return keys, values, err
}

/*
检查是否是空数据库
*/
func (this *LevelDB) CheckNullDB(key []byte) (bool, error) {
	var err error
	if key, err = LeveldbBuildKey(key); err != nil {
		return false, err
	}
	// _, err := this.Find(config.Key_block_start)
	_, err = this.Find(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			//认为这是一个空数据库
			return true, nil
		}
		return false, err
	}
	return false, nil
}

/*
检查key是否存在
@return    bool    true:存在;false:不存在;
*/
func (this *LevelDB) CheckHashExist(hash []byte) bool {
	var err error
	// if hash, err = LeveldbBuildKey(hash); err != nil {
	// 	return false
	// }
	// fmt.Println(hex.EncodeToString(hash))
	_, err = this.Find(hash)
	if err != nil {
		if err == leveldb.ErrNotFound {
			// fmt.Println("db 没找到")
			// utils.Log.Debug().Msgf("db 没找到 %s", hex.EncodeToString(hash))
			return false
		}
		// fmt.Println("db 错误")
		// utils.Log.Debug().Msgf("db 错误 %s", hex.EncodeToString(hash))
		return true
	}
	// fmt.Println("db 找到了")
	// utils.Log.Debug().Msgf("db 找到了 %s", hex.EncodeToString(hash))
	return true
}

/*
获取数据库连接
*/
func (this *LevelDB) GetDB() *leveldb.DB {
	return this.db
}

/*
关闭leveldb连接
*/
func (this *LevelDB) Close() {
	this.db.Close()
}

/*
保存到Map
*/
func (this *LevelDB) SaveMap(dbkey, key, value []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	if err = checkValueSize(value); err != nil {
		return err
	}
	tempKey := append(dbkey, DataType_Data_Map...)
	return this.db.Put(append(tempKey, key...), value, nil)
}

/*
查询Map
*/
func (this *LevelDB) FindMap(dbkey, key []byte) ([]byte, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return nil, err
	}

	tempKey := append(dbkey, DataType_Data_Map...)
	return this.db.Get(append(tempKey, key...), nil)
}

type KVPair struct {
	isAddOrDel bool
	key        []byte
	value      []byte
}

func NewKVPair(isAddOrDel bool, key, value []byte) (*KVPair, error) {
	var err error
	if key, err = LeveldbBuildKey(key); err != nil {
		return nil, err
	}
	if err = checkValueSize(value); err != nil {
		return nil, err
	}
	kvp := KVPair{
		isAddOrDel: isAddOrDel,
		key:        key,
		value:      value,
	}
	return &kvp, nil
}

/*
保存到Map，可以一次保存多个
*/
func (this *LevelDB) SaveMapMore(dbkey []byte, kvs ...KVPair) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}

	tempKey := append(dbkey, DataType_Data_Map...)
	batch := new(leveldb.Batch)
	for _, one := range kvs {
		if one.isAddOrDel {
			batch.Put(append(tempKey, one.key...), one.value)
		} else {
			batch.Delete(append(tempKey, one.key...))
		}
	}
	return this.db.Write(batch, nil)
}

/*
查询Map中所有的key和value值
*/
func (this *LevelDB) FindMapAllToList(dbkey []byte) ([]DBItem, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	lists := make([]DBItem, 0)
	tempKey := append(dbkey, DataType_Data_Map...)
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		key := make([]byte, len(iter.Key())-(dbkeyLen))
		copy(key, iter.Key()[dbkeyLen:])
		_, key, err = LeveldbParseKey(key)
		if err != nil {
			return nil, err
		}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		item := DBItem{
			Key:   key,
			Value: value,
		}
		lists = append(lists, item)
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}

/*
集合中的项目
*/
type DBItem struct {
	Index []byte
	Key   []byte
	Value []byte
}

/*
删除Map中的一个key
*/
func (this *LevelDB) RemoveMapByKey(dbkey, key []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	tempKey := append(dbkey, DataType_Data_Map...)
	return this.db.Delete(append(tempKey, key...), nil)
}

/*
删除Map
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapByDbKey(dbkey []byte, num uint64, interval time.Duration) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}
	tempKey := append(dbkey, DataType_Data_Map...)
	ticker := time.NewTicker(time.Nanosecond)
	defer ticker.Stop()
	total := uint64(0)
	for range ticker.C {
		total = 0
		iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
		for iter.Next() {
			err = this.db.Delete(iter.Key(), nil)
			if err != nil {
				return err
			}
			total++
			if total >= num {
				break
			}
		}
		iter.Release()
		err = iter.Error()
		if err != nil {
			return err
		}
		if total == 0 {
			break
		}
		ticker.Reset(interval)
	}
	return nil
}

/*
保存到List集合中
*/
func (this *LevelDB) SaveList(dbkey, value []byte) (uint64, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return 0, err
	}
	if err = checkValueSize(value); err != nil {
		return 0, err
	}
	tr, err := this.db.OpenTransaction()
	if err != nil {
		return 0, err
	}
	//查询这个dbkey的最大index是多少
	indexDBKEY := append(dbkey, DataType_Index...)
	indexBs, err := tr.Get(indexDBKEY, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			indexBs = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			tr.Discard()
			return 0, err
		}
	}
	index := BytesToUint64ByBigEndian(indexBs)
	index++
	indexBs = Uint64ToBytesByBigEndian(index)
	//保存最新的index
	err = tr.Put(indexDBKEY, indexBs, nil)
	if err != nil {
		tr.Discard()
		return 0, err
	}
	//保存记录
	tempKey := append(dbkey, DataType_Data_List...)
	err = tr.Put(append(tempKey, indexBs...), value, nil)
	if err != nil {
		tr.Discard()
		return 0, err
	}
	err = tr.Commit()
	if err != nil {
		tr.Discard()
		return 0, err
	}
	return index, nil
}

/*
查询List集合中的所有项目
*/
func (this *LevelDB) FindList(dbkey []byte) ([]DBItem, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	lists := make([]DBItem, 0)
	tempKey := append(dbkey, DataType_Data_List...)
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		indexBs := make([]byte, len(iter.Key())-(dbkeyLen+len(DataType_Data_List)))
		copy(indexBs, iter.Key()[dbkeyLen+len(DataType_Data_List):])
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		lists = append(lists, DBItem{
			Index: indexBs,
			Value: value,
		})
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}
