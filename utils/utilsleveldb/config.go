package utilsleveldb

import (
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"
	"web3_gui/utils"
)

const (
	MaxKeySize           = 65535              //max key size
	MaxValueSize     int = 1024 * 1024 * 1024 //max value size
	DBRemoveNum          = 100                //分批次删除map集合中的数据，每批次默认数量
	DBRemoveInterval     = time.Second        //分批次删除map集合中的数据，每批次默认间隔时间
)

var (
	// errDBKeySize         = errors.New("dbkey length not equal to 8")
	// errIndexSize         = errors.New("index length not equal to 8")
	// errDBKeyNotEqualZero = errors.New("invalid dbkey byte not equal 0")
	//errKeySize = errors.New("invalid key size")
	// errKeyNotEqualZero = errors.New("invalid key byte not equal 0")
	// errKeyNotEqualFull = errors.New("invalid key byte not equal 255")
	//errValueSize = errors.New("invalid value size")

	// IndexZero = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	// IndexFull = []byte{255, 255, 255, 255, 255, 255, 255, 255}

	dataType_Index            = []byte{1}  //保存自增长ID
	dataType_Data_Map_start   = []byte{2}  //Map结构数据保存
	dataType_Data_Map         = []byte{3}  //Map结构数据保存
	dataType_Data_Map_end     = []byte{4}  //Map结构数据保存
	dataType_Data_List_start  = []byte{5}  //List结构数据保存
	dataType_Data_List        = []byte{6}  //List结构数据保存
	dataType_Data_List_end    = []byte{7}  //List结构数据保存
	dataType_Map_In_Map_start = []byte{8}  //map in the map结构
	dataType_Map_In_Map       = []byte{9}  //map in the map结构
	dataType_Map_In_Map_end   = []byte{10} //map in the map结构

	DataType_Index_bs            []byte //
	DataType_Data_Map_start_bs   []byte //
	DataType_Data_Map_bs         []byte //
	DataType_Data_Map_end_bs     []byte //
	DataType_Data_List_start_bs  []byte //
	DataType_Data_List_bs        []byte //
	DataType_Data_List_end_bs    []byte //
	DataType_Map_In_Map_start_bs []byte //
	DataType_Map_In_Map_bs       []byte //
	DataType_Map_In_Map_end_bs   []byte //

	ERROR_CODE_key_size   = utils.RegErrCodeExistPanic(1001, "数据库key长度错误")         //
	ERROR_CODE_value_size = utils.RegErrCodeExistPanic(1002, "数据库value长度错误")       //
	ERROR_CODE_not_found  = utils.RegErrCodeExistPanic(1003, "数据库中查询的key未找到value") //
)

func Init() {
	var ERR utils.ERROR
	DataType_Index_bs, ERR = LeveldbBuildKey(dataType_Index)
	if !ERR.CheckSuccess() {
		panic("错误：" + ERR.String())
	}
	DataType_Data_Map_start_bs, ERR = LeveldbBuildKey(dataType_Data_Map_start)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Data_Map_bs, ERR = LeveldbBuildKey(dataType_Data_Map)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Data_Map_end_bs, ERR = LeveldbBuildKey(dataType_Data_Map_end)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Data_List_start_bs, ERR = LeveldbBuildKey(dataType_Data_List_start)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Data_List_bs, ERR = LeveldbBuildKey(dataType_Data_List)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Data_List_end_bs, ERR = LeveldbBuildKey(dataType_Data_List_end)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Map_In_Map_start_bs, ERR = LeveldbBuildKey(dataType_Map_In_Map_start)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Map_In_Map_bs, ERR = LeveldbBuildKey(dataType_Map_In_Map)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	DataType_Map_In_Map_end_bs, ERR = LeveldbBuildKey(dataType_Map_In_Map_end)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
}

/*
数据库中带的基本信息
*/
type LeveldbInfo struct {
	Version    int   //版本
	CreateTime int64 //创建时间
}

type LeveldbKey struct {
	key []byte
}

func NewLeveldbKey(key []byte) *LeveldbKey {
	lkey := LeveldbKey{key: key}
	return &lkey
}

/*
一个不会出错的方法
报错会panic
项目初始化时候使用
*/
func BuildDbKeyByUinta64(key uint64) (*LeveldbKey, utils.ERROR) {
	return BuildDbKeyByUinta64_old(key)
	var keyBs []byte
	if key > math.MaxUint16 {
		if key > math.MaxUint32 {
			keyBs = utils.Uint64ToBytesByBigEndian(key)
		} else {
			keyBs = utils.Uint32ToBytesByBigEndian(uint32(key))
		}
	} else {
		if key > math.MaxUint8 {
			keyBs = utils.Uint16ToBytesByBigEndian(uint16(key))
		} else {
			keyBs = []byte{uint8(key)}
		}
	}
	return BuildLeveldbKey(keyBs)
}

/*
一个不会出错的方法
报错会panic
项目初始化时候使用
*/
func BuildDbKeyByUinta64_old(key uint64) (*LeveldbKey, utils.ERROR) {
	keyBs := utils.Uint64ToBytesByBigEndian(key)
	return BuildLeveldbKey(keyBs)
}

/*
联合key
*/
func NewLeveldbKeyJoin(keys ...[]byte) (*LeveldbKey, utils.ERROR) {
	bs := make([]byte, 0)
	for _, one := range keys {
		keyOne, ERR := LeveldbBuildKey(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		bs = append(bs, keyOne...)
	}
	key := LeveldbKey{bs}
	return &key, utils.NewErrorSuccess()
}

/*
把多个dbkey合并为1个
*/
func JoinDbKey(dbkeys ...LeveldbKey) *LeveldbKey {
	if len(dbkeys) == 0 {
		return nil
	}
	if len(dbkeys) == 1 {
		return &dbkeys[0]
	}
	dbkeysLength := 0
	for _, one := range dbkeys {
		dbkeysLength += len(one.Byte())
	}
	newKeybs := make([]byte, dbkeysLength)
	index := 0
	for _, one := range dbkeys {
		copy(newKeybs[index:], one.Byte())
		index += len(one.Byte())
		//newKeybs = append(newKeybs, one.key...)
	}
	key := LeveldbKey{newKeybs}
	return &key
}

/*
获取去掉长度前缀的原始key
*/
func (this *LeveldbKey) BaseKey() ([]byte, utils.ERROR) {
	_, key, ERR := LeveldbParseKey(this.key)
	return key, ERR
}

/*
获取加上长度前缀的字节序
*/
func (this *LeveldbKey) Byte() []byte {
	return this.key
}

/*
获取加上长度前缀的字节序
*/
func (this *LeveldbKey) JoinKey(key *LeveldbKey) *LeveldbKey {
	bs := make([]byte, len(this.key)+len(key.Byte()))
	copy(bs, this.key)
	copy(bs[len(this.key):], key.key)
	newKey := LeveldbKey{bs}
	return &newKey
}

/*
获取加上长度前缀的字节序
*/
func (this *LeveldbKey) JoinByte(key *[]byte) *LeveldbKey {
	bs := make([]byte, len(this.key)+len(*key))
	copy(bs, this.key)
	copy(bs[len(this.key):], *key)
	newKey := LeveldbKey{bs}
	return &newKey
}

/*
leveldb中，所有的key都带一个key长度前缀，用2字节保存。
2字节容量是65535，因此key长度不能大于65535。
避免不同长度的key前缀查询的时候混合。
*/
func LeveldbBuildKey(key []byte) ([]byte, utils.ERROR) {
	ERR := checkKeySize(key)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	keyLenBs := utils.Uint16ToBytesByBigEndian(uint16(len(key)))
	return append(keyLenBs, key...), utils.NewErrorSuccess()
}

/*
构建一个key，给[]byte加上前缀
*/
func BuildLeveldbKey(key []byte) (*LeveldbKey, utils.ERROR) {
	key, ERR := LeveldbBuildKey(key)
	if ERR.CheckFail() {
		return nil, ERR
	}
	lkey := LeveldbKey{key: key}
	return &lkey, utils.NewErrorSuccess()
}

/*
创建一个数据库key
*/
func NewLeveldbKey_old(key []byte) (*LeveldbKey, utils.ERROR) {
	return BuildLeveldbKey(key)
}

/*
leveldb中，所有的key都带一个key长度前缀，用2字节保存。
2字节容量是65535，因此key长度不能大于65535。
避免不同长度的key前缀查询的时候混合。
@return    uint16    长度
@return    []byte    解析出来的key
*/
func LeveldbParseKey(key []byte) (uint16, []byte, utils.ERROR) {
	if len(key) < 2 {
		return 0, nil, utils.NewErrorBus(ERROR_CODE_key_size, "parsed key is less than 2")
	}
	length := utils.BytesToUint16ByBigEndian(key[:2])
	if len(key) < int(length+2) {
		return 0, nil, utils.NewErrorBus(ERROR_CODE_key_size, "parsed key data is incomplete")
	}
	return length, key[2 : 2+length], utils.NewErrorSuccess()
}

/*
多个长度+[]byte拼接，通过长度解析出来多个数据。不能有不符合规范的数据，有则报错
*/
func LeveldbParseKeyMore(key []byte) ([]LeveldbKey, utils.ERROR) {
	if len(key) < 2 {
		return nil, utils.NewErrorBus(ERROR_CODE_key_size, "parsed key is less than 2")
	}
	keys := make([]LeveldbKey, 0)
	index := 0
	length := len(key)
	for {
		l := utils.BytesToUint16ByBigEndian(key[index : index+2])
		end := index + 2 + int(l)
		if len(key) < end {
			return nil, utils.NewErrorBus(ERROR_CODE_key_size, "parsed key data is incomplete")
		}
		keyOne := LeveldbKey{key[index:end]}
		keys = append(keys, keyOne)
		index = end
		if index == length {
			break
		}
	}
	return keys, utils.NewErrorSuccess()
}

/*
检查Key长度
*/
func checkKeySize(key []byte) utils.ERROR {
	if len(key) == 0 {
		return utils.NewErrorBus(ERROR_CODE_key_size, "key size is 0")
	}
	if len(key) > MaxKeySize {
		return utils.NewErrorBus(ERROR_CODE_key_size, "key size exceeds "+strconv.Itoa(MaxKeySize))
	}
	return utils.NewErrorSuccess()
}

/*
检查Value长度
*/
func checkValueSize(value []byte) utils.ERROR {
	if len(value) > MaxValueSize {
		return utils.NewErrorBus(ERROR_CODE_value_size, "value size exceeds "+strconv.Itoa(MaxValueSize))
	}
	return utils.NewErrorSuccess()
}

/*
一次数据库操作
*/
type KVPair struct {
	IsAddOrDel bool   //操作类型，true=保存或修改;false=删除;
	Key        []byte //
	Value      []byte //
}

/*
创建一次数据库操作
@isAddOrDel    bool    true=保存或修改;false=删除;
*/
func NewKVPair(isAddOrDel bool, key, value []byte) (*KVPair, utils.ERROR) {
	ERR := checkValueSize(value)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	kvp := KVPair{
		IsAddOrDel: isAddOrDel,
		Key:        key,
		Value:      value,
	}
	return &kvp, utils.NewErrorSuccess()
}

/*
一次数据库操作
*/
type KVMore struct {
	Key   LeveldbKey //
	Value []byte     //
}

/*
集合中的项目
*/
type DBItem struct {
	Index []byte
	Key   LeveldbKey
	Value []byte
}

type IndexLock struct {
	Index *big.Int
	lock  *sync.Mutex
}

func NewIndexLock(index *big.Int) *IndexLock {
	indexLock := IndexLock{
		Index: index,
		lock:  new(sync.Mutex),
	}
	return &indexLock
}
