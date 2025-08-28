package utils

import (
	"time"

	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
保存到 Map中有Map 集合中
@dbkey    []byte    数据库ID
@keyOut   []byte    外层map索引
@keyIn    []byte    内层map索引
@value    []byte    内层map索引对应的值
*/
func (this *LevelDB) SaveMapInMap(dbkey, keyOut, keyIn, value []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if keyOut, err = LeveldbBuildKey(keyOut); err != nil {
		return err
	}
	if keyIn, err = LeveldbBuildKey(keyIn); err != nil {
		return err
	}
	if err = checkValueSize(value); err != nil {
		return err
	}

	// keyOutLenBs := Uint16ToBytesByBigEndian(uint16(len(keyOut)))
	// keyInLenBs := Uint16ToBytesByBigEndian(uint16(len(keyIn)))
	// tempKey := append(dbkey, DataType_Map_Len...)
	// tempKey = append(tempKey, keyOutLenBs...)
	tempKey := append(dbkey, keyOut...)
	// tempKey = append(tempKey, keyInLenBs...)
	tempKey = append(tempKey, keyIn...)
	return this.db.Put(tempKey, value, nil)
}

/*
查询 Map中有Map 集合中内层key对应的值
@dbkey    []byte    数据库ID
@keyOut   []byte    外层map索引
@keyIn    []byte    内层map索引
@return   []byte    内层map索引对应的值
*/
func (this *LevelDB) FindMapInMapByKeyIn(dbkey, keyOut, keyIn []byte) ([]byte, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	if keyOut, err = LeveldbBuildKey(keyOut); err != nil {
		return nil, err
	}
	if keyIn, err = LeveldbBuildKey(keyIn); err != nil {
		return nil, err
	}
	// keyOutLenBs := Uint16ToBytesByBigEndian(uint16(len(keyOut)))
	// keyInLenBs := Uint16ToBytesByBigEndian(uint16(len(keyIn)))
	// tempKey := append(dbkey, DataType_Map_Len...)
	// tempKey = append(tempKey, keyOutLenBs...)
	tempKey := append(dbkey, keyOut...)
	// tempKey = append(tempKey, keyInLenBs...)
	tempKey = append(tempKey, keyIn...)
	value, err := this.db.Get(tempKey, nil)
	return value, err
}

/*
查询 Map中有Map 集合中内层key对应的值
@dbkey    []byte    数据库ID
@keyOut   []byte    外层map索引
@keyIn    []byte    内层map索引
@return   []byte    内层map索引对应的值
*/
func (this *LevelDB) RemoveMapInMapByKeyIn(dbkey, keyOut, keyIn []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if keyOut, err = LeveldbBuildKey(keyOut); err != nil {
		return err
	}
	if keyIn, err = LeveldbBuildKey(keyIn); err != nil {
		return err
	}
	// keyOutLenBs := Uint16ToBytesByBigEndian(uint16(len(keyOut)))
	// keyInLenBs := Uint16ToBytesByBigEndian(uint16(len(keyIn)))
	// tempKey := append(dbkey, DataType_Map_Len...)
	// tempKey = append(tempKey, keyOutLenBs...)
	tempKey := append(dbkey, keyOut...)
	// tempKey = append(tempKey, keyInLenBs...)
	tempKey = append(tempKey, keyIn...)
	err = this.db.Delete(tempKey, nil)
	return err
}

/*
间隔删除 Map中有Map 集合中外层key对应的值
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@dbkey       []byte           数据库ID
@keyOut      []byte           外层map索引
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapInMapByKeyOutInterval(dbkey, keyOut []byte, num uint64, interval time.Duration) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if keyOut, err = LeveldbBuildKey(keyOut); err != nil {
		return err
	}
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}
	// keyOutLenBs := Uint16ToBytesByBigEndian(uint16(len(keyOut)))
	// tempKey := append(dbkey, DataType_Map_Len...)
	// tempKey = append(tempKey, keyOutLenBs...)
	tempKey := append(dbkey, keyOut...)
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
