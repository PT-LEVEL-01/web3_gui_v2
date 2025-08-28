package mining

import (
	"github.com/gogo/protobuf/proto"
	"math/big"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/utils"
)

var (
	// 见证人/社区/轻节点奖励事件签名
	Event_LogRewardHistory = ""
	// 社区和轻节点奖励池
	Event_CommunityRewardPool = ""
)

func init() {
	// 预先生成好合约事件签名
	abiObj, _ := abi.JSON(strings.NewReader(precompiled.REWARD_RATE_ABI))
	Event_LogRewardHistory = common.Bytes2Hex(abiObj.Events["LogRewardHistory"].ID.Bytes())
	Event_CommunityRewardPool = common.Bytes2Hex(abiObj.Events["LogCommunityRewardPool"].ID.Bytes())
}

// 统计合约事件
func (this *BalanceManager) countContractEvents(bhvo *BlockHeadVO) {
	for _, txItr := range bhvo.Txs {
		if txItr.Class() == config.Wallet_tx_type_pay {
			// 排除一些不产生合约事件的交易类型
			continue
		}

		txhash := txItr.GetHash()
		key := append([]byte(config.DBKEY_BLOCK_EVENT), big.NewInt(int64(bhvo.BH.Height)).Bytes()...)
		listBytes, err := db.LevelDB.GetDB().HGet(key, *txhash)
		if err != nil {
			continue
		}
		list := go_protos.ContractEventInfoList{}
		err = proto.Unmarshal(listBytes, &list)
		if err != nil {
			continue
		}

		for _, v := range list.GetContractEvents() {
			// 累计奖励
			if Event_LogRewardHistory == v.Topic {
				historyLog := precompiled.UnPackLogRewardHistoryLog(v)
				switch historyLog.Utype {
				case 1: //见证人
				case 2: //社区
					addr := evmutils.AddressToAddressCoin(historyLog.Into[:])
					value := historyLog.Reward.Uint64()
					val, err := db.LevelTempDB.Get(config.BuildCommunityAllReward(addr))
					if err == nil {
						value += utils.BytesToUint64(val)
					}
					bs := utils.Uint64ToBytes(value)
					db.LevelTempDB.Save(config.BuildCommunityAllReward(addr), &bs)
				case 3: //轻节点
					addr := evmutils.AddressToAddressCoin(historyLog.Into[:])
					value := historyLog.Reward.Uint64()
					val, err := db.LevelTempDB.Get(config.BuildLightAllReward(addr))
					if err == nil {
						value += utils.BytesToUint64(val)
					}
					bs := utils.Uint64ToBytes(value)
					db.LevelTempDB.Save(config.BuildLightAllReward(addr), &bs)
				}
			}

			// 社区奖励池
			if Event_CommunityRewardPool == v.Topic {
				cRewardPool := precompiled.UnPackLogCommunityRewardPool(v)
				addr := evmutils.AddressToAddressCoin(cRewardPool.Community[:])
				bs := utils.Uint64ToBytes(cRewardPool.Reward.Uint64())
				db.LevelTempDB.Save(config.BuildCommunityRewardPool(addr), &bs)
			}
		}
	}
}
