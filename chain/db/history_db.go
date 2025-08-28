/*
转账历史纪录的增删改查
key 格式
history_[in/out]_[自己的地址]_[对方的地址]
*/
package db

import "web3_gui/chain/config"

/*
保存一个历史交易纪录
*/
func SaveHistory(id []byte, bs *[]byte) error {
	txhashkey := config.BuildBlockTx(id)
	return LevelTempDB.Save(txhashkey, bs)
	// return db.Put(id, *bs, nil)
}
