package flood

import (
	"sync"
	"web3_gui/utils"
)

var floodKeyRegisterMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

/*
注册一个ID，并判断是否有重复
@return    bool    是否成功
*/
func RegisterFloodKey(id []byte) bool {
	_, ok := floodKeyRegisterMap.LoadOrStore(utils.Bytes2string(id), nil)
	if ok {
		utils.Log.Error().Msgf("重复注册的数据库ID:%d", id)
	}
	return !ok
}
