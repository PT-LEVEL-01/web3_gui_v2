/*
私有节点地址
*/
package addr_manager

import (
	"encoding/json"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

// func init() {
// 	registerFunc(loadByDB)
// }

/*
通过本地数据库获得超级节点地址
*/
func (this *AddrManager) loadByDB(areaName []byte) []string {
	key := append(config.DBKEY_peers_entry, areaName...)
	var bs *[]byte
	newKey, ERR := utilsleveldb.BuildLeveldbKey(key)
	if !ERR.CheckSuccess() {
		return nil
	}
	if newbs, err := this.levelDB.Find(*newKey); err == nil && newbs != nil && newbs.Value != nil {
		//utils.Log.Info().Msgf("===========加载新key的levelDB+++++++++++")
		bs = &newbs.Value
	} else {
		newEntryKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_peers_entry)
		if !ERR.CheckSuccess() {
			return nil
		}
		it, err := this.levelDB.Find(*newEntryKey)
		//utils.Log.Info().Msgf("===========加载旧方式的levelDB+++++++++++")
		if err != nil || it == nil || it.Value == nil {
			// utils.Log.Info().Msgf("find peer entry to db error:%s", err.Error())
			return nil
		}
		bs = &it.Value
	}
	// utils.Log.Info().Msgf("保存地址到数据库:%s", string(*bs))
	ips := make([]string, 0)
	err := json.Unmarshal(*bs, &ips)
	if err != nil {
		utils.Log.Info().Msgf("unmarshal error:%s", err.Error())
		return nil
	}
	return ips
}

/*
节点地址保存到数据库
*/
func (this *AddrManager) SavePeerEntryToDB(peers []string, areaName []byte) {
	bs, err := json.Marshal(peers)
	if err != nil {
		utils.Log.Info().Msgf("save peer entry to db error:%s", err.Error())
		return
	}
	// utils.Log.Info().Msgf("保存地址到数据库:%s", string(bs))
	key := append(config.DBKEY_peers_entry, areaName...)
	newKey, ERR := utilsleveldb.BuildLeveldbKey(key)
	if !ERR.CheckSuccess() {
		return
	}
	ERR = this.levelDB.Save(*newKey, &bs)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("save peer entry to db error:%s", ERR.String())
		return
	}
}
