package utilsleveldb

import (
	"web3_gui/utils"
)

/*
保存数据到列表
@return    []byte    记录的index索引
*/
func (this *Cache) List_Save(db *LevelDB, dbkey *LeveldbKey, value *[]byte) ([]byte, utils.ERROR) {
	indexBs, ERR := db.GetIndex(dbkey)
	if ERR.CheckFail() {
		return nil, ERR
	}
	indexKey, ERR := BuildLeveldbKey(indexBs)
	if ERR.CheckFail() {
		return nil, ERR
	}
	this.Set_Save(dbkey, indexKey, value)
	return indexBs, utils.NewErrorSuccess()
}

/*
保存数据到列表
@return    []byte    记录的index索引
*/
func (this *Cache) List_SaveOrUpdate(dbkey *LeveldbKey, index, value *[]byte) utils.ERROR {
	indexKey, ERR := BuildLeveldbKey(*index)
	if ERR.CheckFail() {
		return ERR
	}
	this.Set_Save(dbkey, indexKey, value)
	return utils.NewErrorSuccess()
}

/*
通过index查询列表
@return    *[]byte    查询到的结果
@return    bool       是否在删除列表中
@return    error      错误
*/
func (this *Cache) List_FindByIndex(dbkey *LeveldbKey, index *[]byte) (*[]byte, bool, utils.ERROR) {
	indexKey, ERR := BuildLeveldbKey(*index)
	if ERR.CheckFail() {
		return nil, false, ERR
	}
	value, ok := this.Set_Find(dbkey, indexKey)
	return value, ok, utils.NewErrorSuccess()
}

/*
保存数据到列表
*/
func (this *Cache) List_Remove(dbkey *LeveldbKey, index *[]byte) utils.ERROR {
	indexKey, ERR := BuildLeveldbKey(*index)
	if ERR.CheckFail() {
		return ERR
	}
	this.Set_Remove(dbkey, indexKey)
	return utils.NewErrorSuccess()
}
