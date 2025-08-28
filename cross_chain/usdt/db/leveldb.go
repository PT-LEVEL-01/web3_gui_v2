package db

import (
	usdtconfig "web3_gui/cross_chain/usdt/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var (
	DBKEY_scanBlockHeight = utilsleveldb.RegDbKeyExistPanicByUint64(101) //不同链的统计高度
	DBKEY_wallet_address  = utilsleveldb.RegDbKeyExistPanicByUint64(102) //钱包地址
	DBKEY_tx_info         = utilsleveldb.RegDbKeyExistPanicByUint64(103) //交易信息详情
)

func ConnLevelDB() (*utilsleveldb.LevelDB, utils.ERROR) {
	ldb, err := utilsleveldb.CreateLevelDB(usdtconfig.PATH_db)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	usdtconfig.Leveldb = ldb
	return ldb, utils.NewErrorSuccess()
}

/*
查询一个链的统计高度
*/
func GetScanBlockHeight(coinType uint32) (uint64, utils.ERROR) {
	coinTypeDBKey, ERR := utilsleveldb.BuildLeveldbKey(utils.Uint32ToBytesByBigEndian(coinType))
	if ERR.CheckFail() {
		return 0, ERR
	}
	bs, err := usdtconfig.Leveldb.FindMap(*DBKEY_scanBlockHeight, *coinTypeDBKey)
	if err != nil {
		return 0, utils.NewErrorSysSelf(err)
	}
	height := utils.BytesToUint64ByBigEndian(bs)
	return height, utils.NewErrorSuccess()
}

/*
保存一个链的统计高度
*/
func SaveScanBlockHeight(coinType uint32, height uint64) {
	//usdtconfig.Leveldb.
}

/*
修改交易归集状态
*/
func UpdateSweepTxInfo() {

}
