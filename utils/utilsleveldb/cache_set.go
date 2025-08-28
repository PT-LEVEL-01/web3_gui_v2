package utilsleveldb

import "web3_gui/utils"

func (this *Cache) Set_Save(dbkey, key *LeveldbKey, value *[]byte) {
	newKey := dbkey.JoinKey(key)
	tempKey := newKey.Byte()
	//utils.Log.Info().Hex("cache保存的key", tempKey).Hex("value", *value).Send()
	this.Save(&tempKey, value)
}

func (this *Cache) Set_Find(dbkey, key *LeveldbKey) (*[]byte, bool) {
	newKey := dbkey.JoinKey(key)
	tempKey := newKey.Byte()
	//utils.Log.Info().Hex("cache保存的key", tempKey).Hex("value", *value).Send()
	bs, ok := this.Find(&tempKey)
	return bs, ok
}

func (this *Cache) Set_Remove(dbkey, key *LeveldbKey) {
	tempKey := dbkey.key
	tempKey = append(tempKey, key.key...)
	newKey := make([]byte, len(tempKey))
	copy(newKey, tempKey)
	this.Remove(&newKey)
}

func (this *Cache) Set_FindMore(dbkey *LeveldbKey, keys []*LeveldbKey, must bool) (map[string][]byte, utils.ERROR) {
	//拼接所有的key
	dbkeys := make([][]byte, 0, len(keys))
	for _, one := range keys {
		dbkeys = append(dbkeys, append(dbkey.Byte(), one.Byte()...))
	}
	values, ERR := this.Base_FindMore(dbkeys, must)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return values, ERR
}
