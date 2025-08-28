package db

import (
	"math/big"

	"web3_gui/chain/config"
	"web3_gui/utils"
)

/*
保存网络中区块最新高度
*/
func SaveHighstBlock(number uint64) error {
	bs := utils.Uint64ToBytes(number)
	return LevelDB.Save([]byte(config.Block_Highest), &bs)
}

/*
查询网络中区块最新高度
*/
func GetHighstBlock() uint64 {
	bs, err := LevelDB.Find([]byte(config.Block_Highest))
	if err != nil {
		return 0
	}
	return utils.BytesToUint64(*bs)
}

/*
保存一笔交易所属的区块hash
*/
func SaveTxToBlockHash(txid, blockhash *[]byte) error {
	key := config.BuildTxToBlockHash(*txid)
	// engine.Log.Info("SaveTxToBlockHash :%s", hex.EncodeToString(*txid))
	return LevelTempDB.Save(key, blockhash)
}

/*
获取一笔交易所属的区块hash
*/
func GetTxToBlockHash(txid *[]byte) (*[]byte, error) {
	// engine.Log.Info("GetTxToBlockHash :%s", hex.EncodeToString(*txid))
	key := config.BuildTxToBlockHash(*txid)
	value, err := LevelTempDB.Find(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func IncContractCount(value *big.Int) error {
	count := GetContractCount()
	count.Add(count, value)
	return SetContractCount(count)
}
func DecContractCount(count *big.Int) error {
	count.Neg(count)
	return IncContractCount(count)
}

// 设置有效的合约数量
func SetContractCount(count *big.Int) error {
	key := []byte(config.EFFECTIVE_contract_count)
	v := count.Bytes()
	return LevelDB.Save(key, &v)
}

// 获取有效的合约数量
func GetContractCount() *big.Int {
	key := []byte(config.EFFECTIVE_contract_count)

	bs, err := LevelDB.Find(key)
	if err != nil {
		return big.NewInt(0)
	}
	return big.NewInt(0).SetBytes(*bs)
}
