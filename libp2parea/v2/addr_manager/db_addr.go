/*
私有节点地址
*/
package addr_manager

import (
	ma "github.com/multiformats/go-multiaddr"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
通过本地数据库获得超级节点地址
*/
func (this *AddrManager) loadByDB() []ma.Multiaddr {
	if this.levelDB == nil {
		return nil
	}
	items, err := this.levelDB.FindMapAllToList(*config.DBKEY_addr_list)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return nil
	}
	addrs := make([]ma.Multiaddr, 0, len(items))
	for _, item := range items {
		a, err := ma.NewMultiaddrBytes(item.Value)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			continue
		}
		addrs = append(addrs, a)
	}
	return addrs
}

/*
节点地址保存到数据库
*/
func (this *AddrManager) SavePeerEntryToDB(peers []ma.Multiaddr) {
	if this.levelDB == nil {
		return
	}
	for _, peer := range peers {
		bs := peer.Bytes()
		dbKey, ERR := utilsleveldb.BuildLeveldbKey(bs)
		if ERR.CheckFail() {
			utils.Log.Error().Str("保存地址错误", ERR.String()).Send()
			continue
		}
		ERR = this.levelDB.SaveMap(*config.DBKEY_addr_list, *dbKey, bs, nil)
		if ERR.CheckFail() {
			utils.Log.Error().Str("保存地址错误", ERR.String()).Send()
			continue
		}
	}
}
