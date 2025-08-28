package mining

import (
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
)

/*
同步区块并刷新本地数据库
*/
func LightWrapperSyncBlockFlashDB(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	bhvo, err := LightWrapperFindBlockHeadNeighbor(bhash, peerBlockInfo)
	if err != nil {
		return nil, err
	}
	if bhvo == nil {
		return nil, config.ERROR_chain_sysn_block_fail
	}
	bhvo.BH.BuildBlockHash()
	// bs, err := bhvo.BH.Json()
	bs, err := bhvo.BH.Proto()
	if err != nil {
		return nil, err
	}
	// if bhvo.BH.Nextblockhash == nil {
	// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
	// }
	bhashkey := config.BuildBlockHead(*bhash)
	db.LevelDB.Save(bhashkey, bs)
	for _, one := range bhvo.Txs {
		// for j, _ := range *one.GetVout() {
		// 	(*bhvo.Txs[i].GetVout())[j].Txid = nil
		// }
		// one.SetBlockHash(*bhash)
		// bs, err := one.Json()
		bs, err := one.Proto()
		if err != nil {
			return nil, err
		}
		txhashkey := config.BuildBlockTx(*one.GetHash())
		db.LevelDB.Save(txhashkey, bs)
	}
	return bhvo, nil
}

/*
从邻居节点数据库中查询区块头
*/
func LightWrapperFindBlockHeadNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	var bhvo *BlockHeadVO
	var bs *[]byte
	var err error
	var newBhvo *BlockHeadVO

	peers := peerBlockInfo.Sort()
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := libp2parea.SortNetAddrForSpeed(addrs)

	for i, _ := range logicNodesInfo {
		key := logicNodesInfo[i].AddrNet
		bs, err = getBlockHeadVO(key, bhash)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		engine.Log.Info("Send to :%s", key.B58String())
		if bs == nil {
			engine.Log.Info("Send query message to node from:%s bs is nil", key.B58String())
			continue
		}
		// engine.Log.Info("bs长度:%d", len(*bs))
		newBhvo, err = ParseBlockHeadVOProto(bs)
		// newBhvo, err = ParseBlockHeadVO(bs)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		// engine.Log.Info("sync block info:%+v", newBhvo)
		bhvo = newBhvo
		//检查本区块是否有nextHash
		if newBhvo.BH.Nextblockhash != nil && len(newBhvo.BH.Nextblockhash) > 0 {
			// engine.Log.Info("this block next block hash not nil")
			return newBhvo, err
		}
		engine.Log.Info("this block next block hash is nil")
		//为空也返回
		// return newBhvo, nil
	}
	//如果从，未超时的节点同步到区块，则不继续同步了
	if bhvo != nil {
		return bhvo, nil
	}

	return bhvo, err
}
