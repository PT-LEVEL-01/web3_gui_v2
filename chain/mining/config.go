package mining

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"web3_gui/chain/config"
	"web3_gui/chain/mining/name"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter"
	"web3_gui/utils"
)

const (
	path_blocks     = "blocks"
	path_chainstate = "chainstate"

	Unit uint64 = 1e8 //输出金额 = 实际金额 * 100000000

)

var Area *libp2parea.Area

func SetArea(area *libp2parea.Area) {
	Area = area
}

// var ModuleEnable = false //是否启用这个模块

/*
构建未花费的交易输出key
*/
func BuildKeyForUnspentTransaction(txid []byte, voutIndex uint64) []byte {
	txidStr := utils.Bytes2string(txid)
	return append(config.AlreadyUsed_tx, []byte(txidStr+"_"+strconv.Itoa(int(voutIndex)))...)
	//return []byte(config.AlreadyUsed_tx + txidStr + "_" + strconv.Itoa(int(voutIndex)))
}

/*
判断是否解锁
@return    bool    是否解锁: true=已经解锁;false=未解锁;
*/
func CheckFrozenHeightFree(frozenHeight uint64, freeHeight uint64, freeTime int64) bool {
	//根据冻结高度判断余额是否可用
	if frozenHeight > config.Wallet_frozen_time_min {
		//按时间锁仓
		if int64(frozenHeight) > freeTime {
			return false
		} else {
			return true
		}
	} else {
		//按块高度锁仓
		if frozenHeight > freeHeight {
			return false
		} else {
			return true
		}
	}
}

/*
判断自己是否是存储节点
*/
func CheckNameStore() bool {
	//获取存储超级节点地址
	nameinfo := name.FindName(config.Name_store)
	if nameinfo == nil {
		//域名不存在
		// engine.Log.Debug("Domain name does not exist")
		return false
	}
	//判断自己是否在超级节点地址里
	for _, one := range nameinfo.NetIds {
		if bytes.Equal(Area.GetNetId(), one) {
			return true
		}
	}
	return false
}

/*
自定义奖励合约交易hash
*/
func CustomBuildRewardHash(txhash []byte, addr crypto.AddressCoin, txtype uint64) []byte {
	buf := bytes.NewBuffer(txhash)
	buf.Write(addr)

	id := make([]byte, 8)
	binary.PutUvarint(id, txtype)

	return append(id, utils.Hash_SHA3_256(buf.Bytes())...)
}
