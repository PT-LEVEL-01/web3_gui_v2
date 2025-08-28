package db

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/libp2parea/adapter/engine"
)

var swapPoolKey = config.DBKEY_swap_tx_pool

// 保存swap发单
func SaveSwapTxProm(tx *go_protos.TxSwap) error {
	field := tx.TxBase.Hash

	bs, err := tx.Marshal()
	if err != nil {
		engine.Log.Error("序列化swap发单失败：%s %s", hex.EncodeToString(tx.TxBase.Hash), err.Error())
		return err
	}

	_, err = LevelTempDB.HSet(swapPoolKey, field, bs)
	if err != nil {
		engine.Log.Error("保存swap发单失败：%s %s", hex.EncodeToString(tx.TxBase.Hash), err.Error())
		return err
	}

	return LevelTempDB.CommitPrefix(nil, LevelTempDB.HEncodeHashKey(swapPoolKey, field))
}

// 加载swap发单池
func LoadSwapTxPromPool() []*go_protos.TxSwap {
	values, err := LevelTempDB.HGetAll(swapPoolKey)
	if err != nil {
		return nil
	}

	list := make([]*go_protos.TxSwap, 0)

	for _, v := range values {
		tx := new(go_protos.TxSwap)
		err := proto.Unmarshal(v.Value, tx)
		if err != nil {
			continue
		}

		list = append(list, tx)
	}

	return list
}

// 移除挂单
func RemoveSwapTxProm(hash []byte) error {
	_, err := LevelTempDB.HDel(swapPoolKey, hash)
	return err
}
