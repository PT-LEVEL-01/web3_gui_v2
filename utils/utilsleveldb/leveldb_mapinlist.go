package utilsleveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"time"
	"web3_gui/utils"
)

/*
保存到MapInList集合中
@isTransaction    bool      是否是事务操作
@return    []byte    index索引字节
@return    error     错误
*/
func (this *LevelDB) SaveMapInList(dbkey, key LeveldbKey, value []byte, batch *leveldb.Batch) ([]byte, utils.ERROR) {
	keyOne := JoinDbKey(dbkey, key)
	return this.SaveList(*keyOne, value, batch)
}

/*
保存多条记录到MapInList集合中
@isTransaction    bool      是否是事务操作
@return           []byte    索引字节
@return           error     错误
*/
func (this *LevelDB) SaveMapInListMore(dbkey, key LeveldbKey, value [][]byte, batch *leveldb.Batch) ([][]byte, utils.ERROR) {
	keyOne := JoinDbKey(dbkey, key)
	return this.SaveListMore(*keyOne, value, batch)
}

/*
修改MapInList集合中的一个key的value值
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) SaveOrUpdateMapInListByIndex(dbkey, key LeveldbKey, index, value []byte, batch *leveldb.Batch) utils.ERROR {
	keyOne := JoinDbKey(dbkey, key)
	return this.SaveOrUpdateListByIndex(*keyOne, index, value, batch)
}

/*
删除MapInList集合中的一个key
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) RemoveMapInListByIndex(dbkey, key LeveldbKey, index []byte, batch *leveldb.Batch) utils.ERROR {
	keyOne := JoinDbKey(dbkey, key)
	return this.RemoveListByIndex(*keyOne, index, batch)
}

/*
删除MapInList集合中的一个key
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) RemoveMapInListMore(isTransaction bool, dbkey, key LeveldbKey, index ...[]byte) utils.ERROR {
	keyOne := JoinDbKey(dbkey, key)
	return this.RemoveListMore(isTransaction, *keyOne, index...)
}

/*
间隔删除MapInList集合中的一个key
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapInListByKeyInterval(dbkey, key LeveldbKey, num uint64, interval time.Duration) error {
	keyOne := JoinDbKey(dbkey, key)
	return this.RemoveListInterval(*keyOne, num, interval)
}

/*
删除MapInList集合中的所有记录
*/
func (this *LevelDB) RemoveMapInListAll(dbkey, key LeveldbKey, batch *leveldb.Batch) error {
	keyOne := JoinDbKey(dbkey, key)
	return this.RemoveListAll(*keyOne, batch)
}

/*
查找MapInList集合中的一个key的value值
*/
func (this *LevelDB) FindMapInListByIndex(dbkey, key LeveldbKey, index []byte) (*DBItem, utils.ERROR) {
	keyOne := JoinDbKey(dbkey, key)
	return this.FindListByIndex(*keyOne, index)
}

/*
查询MapInList集合中所有key
*/
func (this *LevelDB) FindMapInListAllByKey(dbkey, key LeveldbKey) ([]DBItem, error) {
	keyOne := JoinDbKey(dbkey, key)
	return this.FindListAll(*keyOne)
}

/*
查询MapInList集合中一个key下的列表记录总数
@return    uint64    记录总条数
@return    []byte    开始index
@return    []byte    结束index
@return    error     记录总条数
*/
func (this *LevelDB) FindMapInListTotal(dbkey, key LeveldbKey) (uint64, []byte, []byte, utils.ERROR) {
	keyOne := JoinDbKey(dbkey, key)
	return this.FindListTotal(*keyOne)
}

/*
查询MapInList集合中一个范围的记录
不包含startIndex
@order    bool    查询顺序。true=从前向后查询;false=从后向前查询;
*/
func (this *LevelDB) FindMapInListRangeByKeyIn(dbkey, key LeveldbKey, startIndex []byte, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	keyOne := JoinDbKey(dbkey, key)
	return this.FindListRange(*keyOne, startIndex, limit, order)
}

/*
查询MapInList集合中一个范围的记录
不包含startIndex
@order    bool    查询顺序。true=从前向后查询;false=从后向前查询;
*/
func (this *LevelDB) FindMapInListRangeByKeyOut(dbkey LeveldbKey, startIndex []byte, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	//keyOne := JoinDbKey(dbkey, key)
	return this.FindListRange(dbkey, startIndex, limit, order)
}

/*
查询MapInList集合中一个范围的记录
不包含startIndex
*/
func (this *LevelDB) FindMapInListAll(dbkey LeveldbKey) ([]DBItem, utils.ERROR) {
	return this.FindMapInListRangeByKeyOut(dbkey, nil, 0, true)
}
