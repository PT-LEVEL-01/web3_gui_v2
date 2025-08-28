package db

import (
	"web3_gui/cross_chain/usdt/config"
	"web3_gui/cross_chain/usdt/models"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

func SaveTxInfo(txInfos []*models.TxInfo) utils.ERROR {
	if len(txInfos) == 0 {
		return utils.NewErrorSuccess()
	}
	txOne := txInfos[0]
	coinTypeDBKey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint32ToBytesByBigEndian(txOne.ChainCoinType))
	if ERR.CheckFail() {
		return ERR
	}
	cache := utilsleveldb.NewCache()
	blockHeightBs := utils.Uint64ToBytesByBigEndian(txOne.BlockHeight)
	//保存区块高度
	cache.Set_Save(DBKEY_scanBlockHeight, coinTypeDBKey, &blockHeightBs)
	//保存交易记录
	for _, one := range txInfos {
		bs, err := one.Proto()
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return utils.NewErrorSysSelf(err)
		}
		//保存交易详情
		cache.List_Save(config.Leveldb, DBKEY_tx_info, bs)
	}
	err := config.Leveldb.Cache_CommitCache(cache)
	return utils.NewErrorSysSelf(err)
}
