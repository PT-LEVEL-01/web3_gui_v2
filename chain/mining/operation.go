package mining

import (
	"bytes"
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"strconv"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain/sqlite3_db"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/utils"
)

/*
获取账户所有地址的余额
@return    uint64    可用余额
@return    uint64    冻结余额
@return    uint64    锁仓余额
*/
func FindBalanceValue() (uint64, uint64, uint64) {
	chain := forks.GetLongChain()
	if chain == nil {
		return 0, 0, 0
	}
	var notspend, frozen, lock uint64
	for _, one := range Area.Keystore.GetAddrAll() {
		n, f, l := GetBalanceForAddrSelf(one.Addr)
		notspend += n
		frozen += f
		lock += l
	}
	return notspend, frozen, lock
}

/*
通过地址获取余额
@return    uint64    可用余额
@return    uint64    锁定余额
*/
func GetBalanceForAddrSelf(addr crypto.AddressCoin) (uint64, uint64, uint64) {
	chain := forks.GetLongChain()
	if chain == nil {
		return 0, 0, 0
	}

	_, notspend := db.GetNotSpendBalance(&addr)
	//查询锁定余额
	frozenValue := GetAddrFrozenValue(&addr)
	//判断是否有社区节点锁定
	//communityFrozen := GetCommunityVoteRewardFrozen(&addr)
	//获取未提现的奖励
	//communityFrozen := GetFrozenReward(addr)
	lockValue, lockVoteReward := chain.Balance.FindLockTotalByAddr(&addr)

	// n:15892962410 f:0 l:16892962410 cacheLockNotspend:0 cacheLockVoteReward:0
	// engine.Log.Info("n:%d f:%d l:%d cacheLockNotspend:%d cacheLockVoteReward:%d", notspend, frozenValue, communityFrozen, lockValue, lockVoteReward)

	//frozenValue += communityFrozen
	//notspend -= communityFrozen
	notspend -= lockValue
	frozenValue -= lockVoteReward
	lockValue += lockVoteReward

	//查询代币冻结
	fba := GetLongChain().GetTransactionManager().unpackedTransaction.getTokenFrozenBalance(nil, addr)
	baLockup := GetTokenLockedBalance(nil, addr)

	notspend -= (fba + baLockup.Uint64())
	frozenValue += fba
	lockValue += baLockup.Uint64()

	// n:18446744072709551616 f:16892962410 l:16892962410
	// engine.Log.Info("n:%d f:%d l:%d", notspend, frozenValue, communityFrozen)

	return notspend, frozenValue, lockValue
}

/*
通过TokenId和地址获取Token余额和锁定金额
@return    uint64    可用余额
@return    uint64    锁定余额
*/
func GetTokenNotSpendAndLockedBalance(tokenId []byte, addr crypto.AddressCoin) (uint64, uint64, uint64) {
	frozenValue := GetLongChain().GetTransactionManager().unpackedTransaction.getTokenFrozenBalance(tokenId, addr)
	lockedValue := GetTokenLockedBalance(tokenId, addr)
	notSpend := getTokenNotSpendBalance(&tokenId, &addr)
	notSpendValue := uint64(0)
	if notSpend.Uint64() < (frozenValue + lockedValue.Uint64()) {
		notSpendValue = 0
	} else {
		notSpendValue = notSpend.Uint64() - frozenValue - lockedValue.Uint64()
	}
	//utils.Log.Info().Msgf("tokenid:%s addr:%s value1:%d value2:%d value3:%d", hex.EncodeToString(tokenId),
	//	addr.B58String(), notSpendValue, frozenValue, lockedValue.Uint64())
	return notSpendValue, frozenValue, lockedValue.Uint64()
}

/*
通过地址获取余额，查询全网其他节点地址的余额
@return    uint64    可用余额
@return    uint64    锁定余额
*/
func GetNotspendByAddrOther(chain *Chain, addr crypto.AddressCoin) (uint64, uint64, uint64) {
	//chain := forks.GetLongChain()
	if chain == nil {
		return 0, 0, 0
	}
	_, notspend := db.GetNotSpendBalance(&addr)
	//查询锁定余额
	frozenValue := GetAddrFrozenValue(&addr)
	//判断是否有社区节点锁定
	communityFrozen := GetCommunityVoteRewardFrozen(&addr)
	// engine.Log.Info("n:%d f:%d l:%d", notspend, frozenValue, communityFrozen)
	// lockValue := chain.Balance.FindLockTotalByAddr(&addr)
	frozenValue += communityFrozen
	notspend -= communityFrozen
	// notspend -= lockValue
	// engine.Log.Info("n:%d f:%d l:%d", notspend, frozenValue, communityFrozen)
	return notspend, frozenValue, 0
}

/*
	获取账户所有地址的余额
	@return    uint64    可用余额
	@return    uint64    冻结余额
	@return    uint64    锁仓余额
*/
// func GetBalances() (uint64, uint64, uint64) {
// 	count := uint64(0)
// 	countf := uint64(0)
// 	countLockup := uint64(0)

// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		return 0, 0, 0
// 	}

// 	txitems, bfs, itemsLockup := chain.GetBalance().FindBalanceAll()
// 	for _, one := range txitems {
// 		count = count + one.Value
// 	}
// 	//统计冻结的余额
// 	for _, one := range bfs {
// 		countf = countf + one.Value
// 	}
// 	//统计锁仓的余额
// 	for _, one := range itemsLockup {
// 		countLockup = countLockup + one.Value
// 	}

// 	// engine.Log.Info("打印各种余额 %d %d", count, countf)

// 	return count, countf, countLockup
// }

/*
	获取所有item
*/
// func GetBalanceAllItems() ([]*TxItem, []*TxItem, []*TxItem) {
// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		return nil, nil, nil
// 	}
// 	return chain.GetBalance().FindBalanceAll()
// }

/*
	获取所有地址余额明细
*/
// func GetBalanceAllAddrs() (map[string]uint64, map[string]uint64, map[string]uint64) {
// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		// basMap := make(map[string]uint64)      //可用余额
// 		// fbasMap := make(map[string]uint64)     //冻结的余额
// 		// baLockupMap := make(map[string]uint64) //锁仓的余额
// 		return make(map[string]uint64), make(map[string]uint64), make(map[string]uint64)
// 	}
// 	return chain.GetBalance().FindBalanceAll()
// }

/*
	获取区块是否同步完成
*/
// func GetSyncFinish() bool {
// 	//判断是否同步完成
// 	if forks.GetHighestBlock() <= 0 {
// 		//区块未同步完成，不能挖矿
// 		// engine.Log.Info("区块最新高度 %d", forks.GetHighestBlock())
// 		return false
// 	}
// 	chain := forks.GetLongChain()
// 	if forks.GetHighestBlock() > chain.GetPulledStates() {
// 		//区块未同步完成，不能挖矿
// 		// engine.Log.Info("区块最新高度 %d 区块同步高度 %d", forks.GetHighestBlock(), chain.GetPulledStates())
// 		return false
// 	}
// 	return true
// }

// var oldCountTotal = uint64(0)
// var countTotal = uint64(0)
// var onece sync.Once

// func Onece() {
// 	onece.Do(func() {
// 		for {
// 			time.Sleep(time.Second)
// 			newTotal := atomic.LoadUint64(&countTotal)
// 			fmt.Println("====================\n每秒钟处理交易笔数", newTotal-oldCountTotal)
// 			atomic.StoreUint64(&oldCountTotal, countTotal)
// 		}
// 	})
// }

var createTxLock sync.RWMutex

/*
打开合并交易功能
*/
func SendToAddress(srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, domain string, domainType uint64) (*Tx_Pay, error) {
	unlock := createTxMutex.Lock(srcAddress.B58String())

	//验证未上链交易数量
	addrStr := utils.Bytes2string(*srcAddress)
	tsrItr, ok := forks.GetLongChain().TransactionManager.unpackedTransaction.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		if tsr.TrsLen() >= config.Wallet_addr_tx_count_max {
			unlock()
			return nil, errors.New("unpackedTransaction too more")
		}
	}

	txpay, err := CreateTxPay(srcAddress, address, amount, gas, frozenHeight, pwd, comment, domain, domainType)
	unlock()
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}

	//txpay.BuildHash()

	// engine.Log.Info("create tx finish!")
	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		GetLongChain().Balance.DelLockTx(txpay)
		return nil, errors.Wrap(err, "add tx fail!")
	}

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
铸造一个NFT
*/
func BuildNFT(srcAddress *crypto.AddressCoin, gas uint64, pwd, comment string,
	owner *crypto.AddressCoin, name, symbol, resources string) (*Tx_nft, error) {

	txpay, err := CreateTxNFT(srcAddress, gas, pwd, comment, owner, name, symbol, resources)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	// engine.Log.Info("create tx finish!")
	forks.GetLongChain().TransactionManager.AddTx(txpay)

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
转移一个NFT
*/
func TransferNFT(gas uint64, pwd, comment string, nftID []byte, owner *crypto.AddressCoin) (*Tx_nft, error) {

	txpay, err := TransferTxNFT(gas, pwd, comment, nftID, owner)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	// engine.Log.Info("create tx finish!")
	forks.GetLongChain().TransactionManager.AddTx(txpay)

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
	给一个地址转账
*/
// func MergeTx(items []*TxItem, address *crypto.AddressCoin, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
// 	// return nil, errors.New("hahahaha")
// 	// engine.Log.Info("开始合并交易")
// 	txpay, err := MergeTxPay(items, address, gas, frozenHeight, pwd, comment)
// 	if err != nil {
// 		// fmt.Println("创建交易失败", err)
// 		return nil, err
// 	}
// 	txpay.BuildHash()
// 	// engine.Log.Info("create tx finish!")
// 	forks.GetLongChain().transactionManager.AddTx(txpay)

// 	MulticastTx(txpay)
// 	// engine.Log.Info("multicast tx finish!")
// 	// utils.PprofMem()
// 	return txpay, nil
// }

/*
给多个收款地址转账
*/
func SendToMoreAddress(addr *crypto.AddressCoin, address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	txpay, err := CreateTxsPay(addr, address, gas, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()

	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		GetLongChain().Balance.DelLockTx(txpay)
		return nil, errors.Wrap(err, "add tx fail!")
	}
	MulticastTx(txpay)

	return txpay, nil
}

/*
给多个地址转账,带Payload签名
*/
func SendToMoreAddressByPayload(addr *crypto.AddressCoin, address []PayNumber, gas uint64, pwd string, cs *CommunitySign, startHeight, endHeight uint64) (*Tx_Vote_Reward, error) {
	txpay, err := CreateTxVoteReward(addr, address, gas, pwd, startHeight, endHeight)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()

	forks.GetLongChain().TransactionManager.AddTx(txpay)
	MulticastTx(txpay)

	//	unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	return txpay, nil
}

/*
从邻居节点查询起始区块hash
*/
func FindStartBlockForNeighbor() *ChainInfo {
	startBlockHashMap := make(map[string]struct{})
	for _, key := range Area.NodeManager.GetLogicNodes() {
		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getStartBlockHead, &key, nil, config.Mining_block_time*time.Second)
		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getStartBlockHead, &key, nil, config.Mining_sync_timeout)
		if err != nil {
			engine.Log.Info("FindStartBlockForNeighbor %s error:%s", key.B58String(), err.Error())
			continue
		}
		if bs == nil || len(*bs) == 0 {
			continue
		}
		//utils.Log.Info().Int("返回", len(*bs)).Send()

		chainInfo, err := ParseChainInfo(bs)
		if err != nil {
			return nil
		}

		//检查创始块hash是否相等
		if len(config.CheckStartBlockHash) > 0 && !bytes.Equal(chainInfo.StartBlockHash, config.CheckStartBlockHash) {
			startBlockHashMap[hex.EncodeToString(chainInfo.StartBlockHash)] = struct{}{}
			continue
		}

		return chainInfo

		// message, _ := message_center.SendNeighborMsg(config.MSGID_getStartBlockHead, &key, nil)

		// // bs := flood.WaitRequest(mc.CLASS_getBlockHead, hex.EncodeToString(message.Body.Hash), 0)
		// bs, _ := flood.WaitRequest(message_center.CLASS_getBlockHead, utils.Bytes2string(message.Body.Hash), 0)
		// // fmt.Println("有消息返回了啊")
		// if bs == nil {
		// 	// fmt.Println("从邻居节点查询起始区块hash 发送共享文件消息失败，可能超时")
		// 	continue
		// }
		// chainInfo, err := ParseChainInfo(bs)
		// // chainInfo := new(ChainInfo)
		// // // var jso = jsoniter.ConfigCompatibleWithStandardLibrary
		// // // err := json.Unmarshal(*bs, chainInfo)
		// // decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		// // decoder.UseNumber()
		// // err := decoder.Decode(chainInfo)
		// if err != nil {
		// 	return nil
		// }

		// return chainInfo
		// }
	}
	if len(startBlockHashMap) > 0 {
		h := hex.EncodeToString(config.CheckStartBlockHash)
		for k, _ := range startBlockHashMap {
			engine.Log.Info("创始区块:%s 与本节点创始区块:%s 不匹配！", k, h)
		}
	}

	return nil
}

/*
从邻居节点查询区块头和区块中的交易
*/
func FindBlockForNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	var bhvo *BlockHeadVO
	var bs *[]byte
	var err error
	var newBhvo *BlockHeadVO
	//	engine.Log.Info("节点数量111:%d", len(peerBlockInfo.Peers))

	//优先从高度高的节点同步，有的节点不是区块链节点，返回高度是0，会影响同步
	for _, one := range peerBlockInfo.Sort() {
		bs, err = getBlockHeadVO(*one.Addr, bhash)
		if err != nil {
			utils.Log.Info().Msgf("Send query message to node from:%s error:%s", one.Addr.B58String(), err.Error())
			continue
		}
		//engine.Log.Info("Send to :%s", key.B58String())
		if bs == nil || len(*bs) == 0 {
			utils.Log.Info().Msgf("Send query message to node from:%s bs is nil", one.Addr.B58String())
			continue
		}
		//engine.Log.Info("bs长度:%d", len(*bs))
		utils.Log.Info().Int("bs长度", len(*bs)).Send()
		newBhvo, err = ParseBlockHeadVOProto(bs)
		// newBhvo, err = ParseBlockHeadVO(bs)
		if err != nil {
			utils.Log.Info().Msgf("Send query message to node from:%s error:%s", one.Addr.B58String(), err.Error())
			continue
		}
		if newBhvo.BH.Nextblockhash != nil && len(newBhvo.BH.Nextblockhash) > 0 {
			// engine.Log.Info("this block next block hash not nil")
			return newBhvo, err
		}
		utils.Log.Info().Msgf("this block next block hash is nil")
	}

	//根据超时时间排序
	// logicNodes := nodeStore.GetLogicNodes()
	// logicNodes = append(logicNodes, nodeStore.GetNodesClient()...)
	//logicNodesInfo := libp2parea.SortNetAddrForSpeed(logicNodes)
	//工作模式，false=询问未超时节点;true=询问超时节点;
	//如果不切换工作模式，会导致节点同步永远落后几个区块高度

	// mode := false
	// for i := 0; i < 2; i++ {
	// engine.Log.Info("节点数量:%d", len(peerBlockInfo.Peers))
	// TAG:
	//peers := peerBlockInfo.Sort()
	// engine.Log.Info("节点数量:%d", len(peers))
	//addrs := make([]nodeStore.AddressNet, 0)
	//for _, one := range peers {
	//	addrs = append(addrs, *one.Addr)
	//}
	//logicNodesInfo := libp2parea.SortNetAddrForSpeed(addrs)
	//engine.Log.Info("节点数量:%d", len(logicNodesInfo))
	logicNodes := Area.NodeManager.GetLogicNodes()
	logicNodes = append(logicNodes, Area.NodeManager.GetNodesClient()...)
	for i, _ := range logicNodes {
		//询问未超时节点工作模式下：遇到超时的节点，则退出
		// if !mode && one.Speed >= int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		//询问超时节点工作模式下：遇到未超时的节点，则退出
		// if mode && one.Speed < int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		//key := &logicNodesInfo[i].AddrNet
		key := &logicNodes[i]
		// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
		// engine.Log.Info("Send query message to node %s %s", key.B58String(), hex.EncodeToString(*bhash))
		bs, err = getBlockHeadVO(*key, bhash)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		//engine.Log.Info("Send to :%s", key.B58String())
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
	//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
	//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
	// if !mode {
	// 	engine.Log.Info("switch sync timeout nodes mode")
	// 	mode = true
	// 	// goto TAG
	// 	continue
	// }
	// }
	// if bs != nil {
	// 	engine.Log.Info("this block nextblock nil %s", string(*bs))
	// }
	return bhvo, err
}

/*
从邻居节点数据库中查询区块头
*/
func FindBlockHeadNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHead, error) {
	var bh *BlockHead
	var bs *[]byte
	var err error
	var newBh *BlockHead

	peers := peerBlockInfo.Sort()
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := libp2parea.SortNetAddrForSpeed(addrs)

	for i, _ := range logicNodesInfo {
		addrOne := &logicNodesInfo[i].AddrNet
		// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
		engine.Log.Info("Send query message to node %s", addrOne.B58String())
		bhashkey := config.BuildBlockHead(*bhash)
		bs, err = getKeyValue(*addrOne, &bhashkey)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", addrOne.B58String(), err.Error())
			continue
		}
		if bs == nil {
			engine.Log.Info("Send query message to node from:%s bs is nil", addrOne.B58String())
			continue
		}
		newBh, err = ParseBlockHeadProto(bs)

		// newBhvo, err = ParseBlockHeadVOProto(bs)
		// newBhvo, err = ParseBlockHeadVO(bs)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", addrOne.B58String(), err.Error())
			continue
		}
		bh = newBh
		//检查本区块是否有nextHash
		if newBh.Nextblockhash != nil && len(newBh.Nextblockhash) > 0 {
			engine.Log.Info("this block next block hash not nil")
			return newBh, err
		}
		engine.Log.Info("this block next block hash is nil")
		//为空也返回
		// return newBhvo, nil
	}
	//如果从，未超时的节点同步到区块，则不继续同步了
	if bh != nil {
		return bh, nil
	}
	//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
	//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
	// if !mode {
	// 	engine.Log.Info("switch sync timeout nodes mode")
	// 	mode = true
	// 	// goto TAG
	// 	continue
	// }
	// }
	// if bs != nil {
	// 	engine.Log.Info("this block nextblock nil %s", string(*bs))
	// }
	return bh, err
}

/*
查询邻居节点已经导入的最高区块
*/
func FindLastBlockForNeighbor(peerBlockInfo *PeerBlockInfoDESC) (*BlockHead, error) {
	var err error
	var bh *BlockHead

	peers := peerBlockInfo.Sort()
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := libp2parea.SortNetAddrForSpeed(addrs)

	for i, _ := range logicNodesInfo {
		key := &logicNodesInfo[i].AddrNet

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockLastCurrent, key, nil, config.Mining_block_time*time.Second)
		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockLastCurrent, key, nil, config.Mining_sync_timeout)
		if err != nil {
			engine.Log.Info("FindLastBlockForNeighbor error:%s", err.Error())
			continue
		}
		// message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockLastCurrent, key, nil)
		// bs, _ := flood.WaitRequest(message_center.CLASS_getBlockLastCurrent, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
		if bs == nil {
			continue
		}
		bh, err = ParseBlockHeadProto(bs)
		if err != nil {
			continue
		}
		return bh, nil
	}
	return nil, err
}

// func FindBlockForNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
// 	var bhvo *BlockHeadVO
// 	var bs *[]byte
// 	var err error
// 	var newBhvo *BlockHeadVO

// 	//根据超时时间排序
// 	logicNodes := nodeStore.GetLogicNodes()
// 	logicNodes = append(logicNodes, nodeStore.GetNodesClient()...)
// 	logicNodesInfo := libp2parea.SortNetAddrForSpeed(logicNodes)
// 	//工作模式，false=询问未超时节点;true=询问超时节点;
// 	//如果不切换工作模式，会导致节点同步永远落后几个区块高度
// 	mode := false
// 	for i := 0; i < 2; i++ {

// 		// TAG:
// 		for _, one := range logicNodesInfo {
// 			//询问未超时节点工作模式下：遇到超时的节点，则退出
// 			if !mode && one.Speed >= int64(time.Second*config.Wallet_sync_block_timeout) {
// 				continue
// 			}
// 			//询问超时节点工作模式下：遇到未超时的节点，则退出
// 			if mode && one.Speed < int64(time.Second*config.Wallet_sync_block_timeout) {
// 				continue
// 			}
// 			key := one.AddrNet
// 			// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
// 			engine.Log.Info("Send query message to node %s", key.B58String())
// 			bs, err = getValueForNeighbor(key, bhash)
// 			if err != nil {
// 				engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
// 				continue
// 			}
// 			if bs == nil {
// 				engine.Log.Info("Send query message to node from:%s bs is nil", key.B58String())
// 				continue
// 			}
// 			newBhvo, err = ParseBlockHeadVOProto(bs)
// 			// newBhvo, err = ParseBlockHeadVO(bs)
// 			if err != nil {
// 				engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
// 				continue
// 			}
// 			bhvo = newBhvo
// 			//检查本区块是否有nextHash
// 			if newBhvo.BH.Nextblockhash != nil && len(newBhvo.BH.Nextblockhash) > 0 {
// 				// engine.Log.Info("this block nextblock not nil")
// 				return newBhvo, err
// 			}
// 			//为空也返回
// 			// return newBhvo, nil
// 		}
// 		//如果从，未超时的节点同步到区块，则不继续同步了
// 		if bhvo != nil {
// 			return bhvo, nil
// 		}
// 		//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
// 		//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
// 		if !mode {
// 			engine.Log.Info("switch sync timeout nodes mode")
// 			mode = true
// 			// goto TAG
// 			continue
// 		}
// 	}
// 	if bs != nil {
// 		engine.Log.Info("this block nextblock nil %s", string(*bs))
// 	}
// 	return bhvo, err
// }

// func FindBlockForNeighbor(bhash *[]byte) *BlockHeadVO {
// 	bhvo := new(BlockHeadVO)
// 	bs := getValueForNeighbor(bhash)
// 	if bs == nil {
// 		engine.Log.Info("Error synchronizing chunk from neighbor node")
// 		return nil
// 	}
// 	bhvo, err := ParseBlockHeadVO(bs)
// 	if err != nil {
// 		return nil
// 	}
// 	return bhvo
// }

/*
查询邻居节点数据库，key：value查询
bhash:不带前缀的区块哈希
*/
func getBlockHeadVO(key nodeStore.AddressNet, bhash *[]byte) (*[]byte, error) {

	start := config.TimeNow()

	//pl time
	//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockHeadVO, &key, bhash, config.Mining_block_time*time.Second)
	//engine.Log.Error("---------------------bhashbhashbhashbhashbhashbhashbhash %s", hex.EncodeToString(*bhash))
	bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockHeadVO, &key, bhash, config.Mining_sync_timeout)

	if err != nil {
		engine.Log.Info("getBlockHeadVO1 error:%s", err.Error())
		return nil, err
	}
	utils.Log.Info().Int("返回", len(*bs)).Send()
	// message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockHeadVO, &key, bhash)
	// // engine.Log.Info("44444444444 %s", key.B58String())
	// // bs := flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
	// bs, _ := flood.WaitRequest(message_center.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
	if bs == nil {
		endTime := config.TimeNow()
		// engine.Log.Info("5555555555555555 %s", key.B58String())
		//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
		engine.Log.Error("Receive %s message timeout %s", key.B58String(), config.TimeNow().Sub(start))
		//有可能是对方没有查询到区块，返回空，则判定它超时
		if (endTime.Unix() - start.Unix()) < config.Wallet_sync_block_timeout {
			libp2parea.AddNodeAddrSpeed(key, time.Second*(config.Wallet_sync_block_timeout+1))
		} else {
			//TODO 应该取平均数
			//保存上一次同步超时时间
			// config.NetSpeedMap.Store(utils.Bytes2string(key), config.TimeNow().Sub(start))
			libp2parea.AddNodeAddrSpeed(key, config.TimeNow().Sub(start))
		}
		// err = errors.New("Failed to send shared file message, may timeout")

		return nil, config.ERROR_chain_sync_block_timeout
	}
	libp2parea.AddNodeAddrSpeed(key, config.TimeNow().Sub(start))
	// engine.Log.Info("Receive message %s", key.B58String())
	return bs, nil
}

/*
查询邻居节点数据库，key：value查询
*/
func getKeyValue(addrOne nodeStore.AddressNet, key *[]byte) (*[]byte, error) {

	start := config.TimeNow()

	//pl time
	//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getDBKey_one, &addrOne, key, config.Mining_block_time*time.Second)
	bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getDBKey_one, &addrOne, key, config.Mining_sync_timeout)

	// message, _ := message_center.SendNeighborMsg(config.MSGID_getDBKey_one, &addrOne, key)
	// // engine.Log.Info("44444444444 %s", addrOne.B58String())
	// // bs := flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
	// bs, err := flood.WaitRequest(config.CLASS_getKeyValue, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
	if err != nil || bs == nil {
		endTime := config.TimeNow()
		// engine.Log.Info("5555555555555555 %s", addrOne.B58String())
		//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
		engine.Log.Error("Receive %s message timeout %s", addrOne.B58String(), config.TimeNow().Sub(start))
		//有可能是对方没有查询到区块，返回空，则判定它超时
		if (endTime.Unix() - start.Unix()) < config.Wallet_sync_block_timeout {
			libp2parea.AddNodeAddrSpeed(addrOne, time.Second*(config.Wallet_sync_block_timeout+1))
		} else {
			//TODO 应该取平均数
			//保存上一次同步超时时间
			// config.NetSpeedMap.Store(utils.Bytes2string(addrOne), config.TimeNow().Sub(start))
			libp2parea.AddNodeAddrSpeed(addrOne, config.TimeNow().Sub(start))
		}
		// err = errors.New("Failed to send shared file message, may timeout")

		return nil, config.ERROR_chain_sync_block_timeout
	}
	libp2parea.AddNodeAddrSpeed(addrOne, config.TimeNow().Sub(start))
	// engine.Log.Info("Receive message %s", addrOne.B58String())
	return bs, nil
}

// func getValueForNeighbor(bhash *[]byte) *[]byte {
// 	// fmt.Println("1查询区块或交易", hex.EncodeToString(*bhash))
// 	var bs *[]byte
// 	var err error
// 	for {
// 		logicNodes := nodeStore.GetLogicNodes()
// 		logicNodes = OrderNodeAddr(logicNodes)
// 		for _, key := range logicNodes {
// 			engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
// 			engine.Log.Info("Send query message to node %s", key.B58String())

// 			message, _ := message_center.SendNeighborMsg(config.MSGID_getTransaction, key, bhash)
// 			// engine.Log.Info("44444444444 %s", key.B58String())
// 			bs = flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
// 			if bs == nil {
// 				// engine.Log.Info("5555555555555555 %s", key.B58String())
// 				//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
// 				engine.Log.Info("Receive message timeout %s", key.B58String())
// 				err = errors.New("Failed to send shared file message, may timeout")
// 				continue
// 			}
// 			engine.Log.Info("Receive message %s", key.B58String())
// 			// engine.Log.Info("66666666666666 %s", key.B58String())
// 			err = nil
// 			break
// 		}
// 		// engine.Log.Info("7777777777777777777")
// 		if err == nil {
// 			// engine.Log.Info("888888888888888")
// 			break
// 		}
// 		// engine.Log.Info("99999999999999999999")
// 	}
// 	if err != nil {
// 		// engine.Log.Info("10101010101001010101")
// 		engine.Log.Warn("Failed to query block transaction", hex.EncodeToString(*bhash))
// 	}
// 	// engine.Log.Info("11 11 11 11 11")
// 	// if bs == nil {
// 	// 	engine.Log.Info("查询区块或交易结果 %s", bs)
// 	// } else {
// 	// 	engine.Log.Info("查询区块或交易结果 %s", string(*bs))
// 	// }
// 	return bs
// }

/*
从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlockForNeighbor(height uint64, peerBlockInfo *PeerBlockInfoDESC) ([]*BlockHeadVO, error) {
	engine.Log.Info("Synchronize unacknowledged chunks from neighbor nodes")

	heightBs := utils.Uint64ToBytes(height)

	var bs *[]byte
	var err error
	// for i := 0; i < 10; i++ {
	// engine.Log.Info("Synchronize unacknowledged total: %d", i)
	// logicNodes := nodeStore.GetLogicNodes()
	logicNodes := peerBlockInfo.Sort()
	for j, _ := range logicNodes {
		engine.Log.Info("Synchronize unacknowledged from:%s height:%d", logicNodes[j].Addr.B58String(), height)

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getUnconfirmedBlock, logicNodes[j].Addr, &heightBs, config.Mining_block_time*time.Second)
		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getUnconfirmedBlock, logicNodes[j].Addr, &heightBs, config.Mining_sync_timeout)
		if err != nil {
			//消息未发送成功
			continue
		}
		// message, err := message_center.SendNeighborMsg(config.MSGID_getUnconfirmedBlock, logicNodes[j].Addr, &heightBs)
		// if err != nil {
		// 	//消息未发送成功
		// 	continue
		// }

		// // bs = flood.WaitRequest(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), 0)
		// bs, _ = flood.WaitRequest(message_center.CLASS_getUnconfirmedBlock, utils.Bytes2string(message.Body.Hash), 0)
		if bs == nil {
			engine.Log.Info("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
			err = errors.New("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
			continue
		} else {
			err = nil
		}
		engine.Log.Info("Synchronize unacknowledged ok")
		break
	}
	// if err == nil {
	// engine.Log.Info("Synchronize unacknowledged exist")
	// break
	// }
	// }
	// engine.Log.Info("获取的未确认区块 bs", string(*bs))
	if bs == nil {
		engine.Log.Warn("Get unacknowledged block BS error")
		return nil, err
	}

	rbsp := new(go_protos.RepeatedBytes)
	err = proto.Unmarshal(*bs, rbsp)
	if err != nil {
		engine.Log.Warn("Get unacknowledged block BS error:%s", err.Error())
		return nil, err
	}

	blockHeadVOs := make([]*BlockHeadVO, 0)
	for _, one := range rbsp.Bss {
		bhvo, err := ParseBlockHeadVOProto(&one)
		if err != nil {
			engine.Log.Warn("Get unacknowledged block BS error:%s", err.Error())
			return nil, err
		}
		blockHeadVOs = append(blockHeadVOs, bhvo)
	}

	// temp := make([]interface{}, 0)
	// // err = json.Unmarshal(*bs, &temp)
	// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	// decoder.UseNumber()
	// err = decoder.Decode(&temp)

	// blockHeadVOs := make([]*BlockHeadVO, 0)
	// for _, one := range temp {
	// 	bs, err := json.Marshal(one)
	// 	blockVOone, err := ParseBlockHeadVO(&bs)
	// 	if err != nil {
	// 		engine.Log.Warn("Get unacknowledged block BS error:%s", err.Error())
	// 		continue
	// 	}
	// 	blockHeadVOs = append(blockHeadVOs, blockVOone)
	// }

	engine.Log.Info("Get unacknowledged block Success")
	return blockHeadVOs, nil

}

/*
缴纳押金，成为备用见证人
*/
func DepositIn(amount, gas uint64, pwd, payload string, rate uint16) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().Balance.DepositIn(amount, gas, pwd, payload, rate)
	if err != nil {
		// fmt.Println("缴纳押金失败", err)
	}
	// fmt.Println("缴纳押金完成")
	return err
}

/*
退还押金
@addr    string    可选（默认退回到原地址）。押金赎回到的账户地址
@amount  uint64    可选（默认退还全部押金）。押金金额
*/
func DepositOut(addr string, amount, gas uint64, pwd string) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().Balance.DepositOut(addr, amount, gas, pwd)

	return err
}

/*
给见证人投票
*/
func VoteIn(t, rate uint16, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string, gasPrice uint64) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().Balance.VoteIn(t, rate, witnessAddr, addr, amount, gas, pwd, payload)
	//err := forks.GetLongChain().Balance.VoteInContract(t, rate, witnessAddr, addr, amount, gas, pwd, payload, gasPrice)
	return err

}

// 查询见证人是否已存在（共轻节点使用）
func SearchWitnessIsExist(voter crypto.AddressCoin) bool {
	return forks.GetLongChain().Balance.witnessBackup.haveWitness(&voter)
}

/*
退还见证人投票押金
*/
func VoteOut(voteType uint16, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string, gasPrice uint64) error {
	//缴纳备用见证人押金交易
	return forks.GetLongChain().Balance.VoteOut(voteType, addr, amount, gas, pwd, payload)
	//return forks.GetLongChain().Balance.VoteOutContract(voteType, addr, amount, gas, pwd, gasPrice)

}

// /*
// 	给社区节点投票
// */
// func VoteInLight(witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, pwd string) error {
// 	//缴纳备用见证人押金交易
// 	err := forks.GetLongChain().balance.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		// fmt.Println("缴纳押金失败", err)
// 	}
// 	// fmt.Println("缴纳押金完成")
// 	return err
// }

// /*
// 	退还给社区节点投票
// */
// func VoteOutLight(witnessAddr *crypto.AddressCoin, txid, addr string, amount, gas uint64, pwd string) error {
// 	//缴纳备用见证人押金交易
// 	return forks.GetLongChain().balance.VoteOut(witnessAddr, txid, addr, amount, gas, pwd)

// }

/*
获取见证人状态
@return    bool    是否是候选见证人
@return    bool    是否是备用见证人
@return    bool    是否是没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
@return    crypto.AddressCoin    见证人地址
*/
func GetWitnessStatus() (IsCandidate bool, IsBackup bool, IsKickOut bool, Addr crypto.AddressCoin, value uint64) {
	addrInfo := Area.Keystore.GetCoinbase()
	Addr = addrInfo.Addr // .GetAddrStr()
	IsCandidate = forks.GetLongChain().WitnessBackup.FindWitness(Addr)
	IsBackup = forks.GetLongChain().WitnessChain.FindWitness(addrInfo.Addr)
	IsKickOut = forks.GetLongChain().WitnessBackup.FindWitnessInBlackList(Addr)
	txItem := forks.GetLongChain().Balance.GetDepositIn()
	if txItem == nil {
		value = 0
	} else {
		value = uint64(txItem.Value)
	}
	return
}

/*
根据地址获取见证人状态
@return    bool    是否是候选见证人
@return    bool    是否是备用见证人
@return    bool    是否是没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
@return    crypto.AddressCoin    见证人地址
*/
func GetWitnessStatusByAddr(addr *keystore.AddressInfo) (IsCandidate bool, IsBackup bool, IsKickOut bool, Addr crypto.AddressCoin, value uint64) {
	Addr = addr.Addr // .GetAddrStr()
	IsCandidate = forks.GetLongChain().WitnessBackup.FindWitness(Addr)
	IsBackup = forks.GetLongChain().WitnessChain.FindWitness(Addr)
	IsKickOut = forks.GetLongChain().WitnessBackup.FindWitnessInBlackList(Addr)
	txItem := forks.GetLongChain().Balance.GetDepositIn()
	if txItem == nil {
		value = 0
	} else {
		value = uint64(txItem.Value)
	}
	return
}

/*
获取候选见证人列表
*/
func GetWitnessListSort() *WitnessBigGroup {
	return forks.GetLongChain().WitnessBackup.GetWitnessListSort()
}

/*
*
获取候选见证人列表
*/
func GetBackupWitnessList() []*BackupWitness {
	return forks.GetLongChain().WitnessBackup.witnesses
}

/*
获取社区节点列表
*/
func GetCommunityListSort() []*VoteScoreVO {
	return forks.GetLongChain().WitnessBackup.GetCommunityListSort()
}

/*
	获得自己给哪些见证人投过票的列表
*/
// func GetVoteList() []*Balance {
// 	for _, one := range keystore.GetAddrAll() {
// 		one.Addr
// 		GetLongChain().Balance.GetDepositVote()
// 	}
// 	return forks.GetLongChain().Balance.GetVoteList()
// }

/*
获取社区节点押金列表
*/
func GetDepositCommunityList() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	for _, one := range Area.Keystore.GetAddrAll() {
		txItem := GetLongChain().Balance.GetDepositCommunity(&one.Addr)
		if txItem != nil {
			items = append(items, txItem)
		}
	}
	return items
}

/*
获取轻节点押金列表
*/
func GetDepositLightList() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	for _, one := range Area.Keystore.GetAddrAll() {
		txItem := GetLongChain().Balance.GetDepositLight(&one.Addr)
		if txItem != nil {
			items = append(items, txItem)
		}
	}
	return items
}

/*
获取轻节点投票押金列表
*/
func GetDepositVoteList() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	for _, one := range Area.Keystore.GetAddrAll() {
		txItem := GetLongChain().Balance.GetDepositVote(&one.Addr)
		if txItem != nil {
			items = append(items, txItem)
		}
	}
	return items
}

/*
查询一个交易是否上链成功，以及交易详情
@return    TxItr    交易详情
@return    uint64    1=未上链;2=成功上链;3=上链失败;
*/
func FindTx(txid []byte) (TxItr, uint64, *[]byte) {
	txItr, err := LoadTxBase(txid)
	if err != nil {
		return nil, 1, nil
	}
	bs, err := db.GetTxToBlockHash(&txid)
	if err != nil {
		return txItr, 1, nil
	}

	if bs != nil && len(*bs) > 0 {
		return txItr, 2, bs
	}
	lockheight := txItr.GetLockHeight()
	//查询当前确认的区块高度
	height := GetLongChain().GetCurrentBlock()
	if height > lockheight {
		//超过了锁定高度还没有上链，则失败了
		return txItr, 3, nil
	}
	return txItr, 1, nil
}

/*
查询一个交易是否上链成功，以及交易详情
@return    TxItr    交易详情
@return    uint64    1=未上链;2=成功上链;3=上链失败;
*/
func FindTxNFT(txid []byte) (TxItr, uint64) {
	if ParseTxClass(txid) != config.Wallet_tx_type_nft {
		return nil, 1
	}
	txItr, err := LoadTxBase(txid)
	if err != nil {
		return nil, 1
	}

	owner, err := FindNFTOwner(txid)
	if err != nil {
		return nil, 1
	}

	nftTx := txItr.(*Tx_nft)
	nftTx.NFT_Owner = *owner

	_, err = db.GetTxToBlockHash(&txid)
	if err != nil {
		lockheight := nftTx.GetLockHeight()
		//查询当前确认的区块高度
		height := GetLongChain().GetCurrentBlock()
		if height > lockheight {
			//超过了锁定高度还没有上链，则失败了
			return nftTx, 3
		}
		return nftTx, 1
	}
	return nftTx, 2
}

func FindTxJsonVo(txid []byte) (interface{}, uint64, *[]byte, uint64, int64) {
	txItr, code, blockHash := FindTx(txid)
	var itr interface{}
	if txItr != nil {
		itr = txItr.GetVOJSON()
	}

	if blockHash == nil {
		return itr, code, nil, 0, 0
	}

	//获取区块信息
	bh, err := LoadBlockHeadByHash(blockHash)
	if err != nil {
		return itr, code, blockHash, 0, 0
	}
	return itr, code, blockHash, bh.Height, bh.Time
}

// 用于从区块中查询
func FindTxJsonVoV1(txid []byte) (interface{}, uint64, TxItr) {
	txItr, code, _ := FindTx(txid)
	var itr interface{}
	if txItr != nil {
		itr = txItr.GetVOJSON()
	}
	return itr, code, txItr
}
func getMineTx(hByte []byte, height uint64) ([]precompiled.LogRewardHistoryV0, error) {
	intHeight := evmutils.New(int64(height))
	key := append([]byte(config.DBKEY_BLOCK_EVENT), intHeight.Bytes()...)
	body, err := db.LevelDB.GetDB().HGet(key, hByte)
	if err != nil || len(body) == 0 {
		return nil, err
	}
	list := &go_protos.ContractEventInfoList{}
	err = proto.Unmarshal(body, list)
	if err != nil {
		return nil, err
	}

	historys := make([]precompiled.LogRewardHistoryV0, 0, len(list.ContractEvents))
	for _, l := range list.ContractEvents {
		v0 := precompiled.LogRewardHistoryV0{}
		log := precompiled.UnPackLogRewardHistoryLog(l)
		address := evmutils.AddressToAddressCoin(log.Into[:])
		v0.Into = address.B58String()
		if log.From == evmutils.ZeroAddress {
			v0.From = ""
		} else {
			from := evmutils.AddressToAddressCoin(log.From[:])
			v0.From = from.B58String()
		}
		v0.Reward = log.Reward
		historys = append(historys, v0)
	}
	return historys, err
}

func FindTxJsonV1(txid []byte) (TxItr, uint64, *[]byte) {
	return FindTx(txid)
}

/*
查询地址角色状态
@return    int    1=见证人;2=社区节点;3=轻节点;4=什么也不是;
*/
func GetAddrStateOld(addr crypto.AddressCoin) int {
	witnessBackup := forks.GetLongChain().WitnessBackup
	////是否是轻节点
	//_, isLight := witnessBackup.haveLight(&addr)
	//if isLight {
	//	return 3
	//}
	////是否是社区节点
	//_, isCommunity := witnessBackup.haveCommunityList(&addr)
	//if isCommunity {
	//	return 2
	//}
	state := precompiled.GetAddrState(config.Area.Keystore.GetAddr()[0].Addr, addr)
	if state != 4 {
		return int(state)
	}
	//是否是见证人
	isWitness := witnessBackup.haveWitness(&addr)
	if isWitness {
		return 1
	}
	return 4
}

/*
查询地址角色状态
@return    int    1=见证人;2=社区节点;3=轻节点;4=什么也不是;
*/
func GetAddrState(addr crypto.AddressCoin) int {
	chain := forks.GetLongChain()
	if chain == nil {
		return 0
	}
	witnessBackup := chain.WitnessBackup

	//是否是见证人
	if witnessBackup.haveWitness(&addr) {
		return 1
	}

	//是否社区节点
	if GetDepositCommunityAddr(&addr) > 0 {
		return 2
	}

	//是否轻节点
	if GetDepositLightAddr(&addr) > 0 {
		return 3
	}

	return 4
}

/*
添加一个自定义交易
验证交易并广播
*/
func AddTx(txItr TxItr) error {
	if txItr == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	txItr.BuildHash()

	err := forks.GetLongChain().TransactionManager.AddTx(txItr)
	if err != nil {
		//等待上链,请稍后重试.
		return errors.Wrap(err, "Waiting for the chain, please try again later")
	}
	MulticastTx(txItr)
	return nil
}

/*
创建一个转款交易
@params height 当前区块高度 item 交易item  pubs 地址公钥对 address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxPayM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, address *crypto.AddressCoin, amount, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_Pay, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]*Vin, 0)
	total := uint64(0)
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			// Nonce:
			// Txid: item.Txid,      //UTXO 前一个交易的id
			// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk: puk, //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

		total = total + uint64(item.Value)
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: *address, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, &vout)
	}
	var pay *Tx_Pay
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_pay, //交易类型
		Vin_total:  uint64(len(vins)),         //输入交易数量
		Vin:        vins,                      //交易输入
		Vout_total: uint64(len(vouts)),        //输出交易数量
		Vout:       vouts,                     //交易输出
		Gas:        gas,                       //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
		//		CreateTime: config.TimeNow().Unix(),         //创建时间
		Comment: []byte{},
	}
	pay = &Tx_Pay{
		TxBase: base,
	}
	//给输出签名，防篡改
	for i, _ := range pay.Vin {
		sign := pay.GetWaitSign(uint64(i))
		if sign == nil {
			//预签名时数据错误
			return nil, errors.New("Data error while pre signing")
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}
	//pay.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return pay, nil
}

/*
创建多个转款交易
@params height 当前区块高度 item 交易item  pubs 地址公钥对 address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxsPayM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, address []PayNumber, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_Pay, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]*Vin, 0)
	total := uint64(0)
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			// Txid: item.Txid,      //UTXO 前一个交易的id
			// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk: puk, //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

		total = total + uint64(item.Value)
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	//构建交易输出
	vouts := make([]*Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
			Address: one.Address, //钱包地址
		}
		vouts = append(vouts, &vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}
	var pay *Tx_Pay
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_pay, //交易类型
		Vin_total:  uint64(len(vins)),         //输入交易数量
		Vin:        vins,                      //交易输入
		Vout_total: uint64(len(vouts)),        //输出交易数量
		Vout:       vouts,                     //交易输出
		Gas:        gas,                       //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
		//		CreateTime: config.TimeNow().Unix(),         //创建时间
		Comment: []byte{},
	}
	pay = &Tx_Pay{
		TxBase: base,
	}

	//给输出签名，防篡改
	for i, _ := range pay.Vin {
		sign := pay.GetWaitSign(uint64(i))
		if sign == nil {
			//预签名时数据错误
			return nil, errors.New("Data error while pre signing")
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}
	//pay.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return pay, nil
}

/*
创建一个投票交易
@params height 当前区块高度 item 交易item  pubs 地址公钥对 voteType 投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；  witnessAddr 接受者地址 addr 投票者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxVoteInM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, voteType uint16, witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_vote_in, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	if voteType == 1 && amount != config.Mining_vote {
		//交押金数量最少需要
		return nil, errors.New("Minimum deposit required" + strconv.FormatUint(config.Mining_vote, 10))
	}
	if voteType == 3 && amount != config.Mining_light_min {
		//交押金数量最少需要
		return nil, errors.New("Minimum deposit required" + strconv.FormatUint(config.Mining_light_min, 10))
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]*Vin, 0)
	total := uint64(0)
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			// Txid: item.Txid,      //UTXO 前一个交易的id
			// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk: puk, //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

		total = total + uint64(item.Value)
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		// fmt.Println("自己地址数量", len(keystore.GetAddr()))
		//为空则转给自己
		dstAddr = returnaddr //keystore.GetAddr()[0]
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
		dstAddr = crypto.AddressFromB58String(addr)
	}
	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, &vout)
	}
	var txin *Tx_vote_in
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_in, //交易类型
		Vin_total:  uint64(len(vins)),             //输入交易数量
		Vin:        vins,                          //交易输入
		Vout_total: uint64(len(vouts)),            //输出交易数量
		Vout:       vouts,                         //交易输出
		Gas:        gas,                           //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
		//		CreateTime: config.TimeNow().Unix(),         //创建时间
		Comment: []byte{},
	}

	// voteAddr := NewVoteAddressByAddr(witnessAddr)

	txin = &Tx_vote_in{
		TxBase:   base,
		Vote:     witnessAddr,
		VoteType: voteType,
	}

	//给输出签名，防篡改
	for i, _ := range txin.Vin {
		sign := txin.GetWaitSign(uint64(i))
		if sign == nil {
			return nil, config.ERROR_get_sign_data_fail
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		// fmt.Printf("sign前:puk:%x signdst:%x", md5.Sum(one.Puk), md5.Sum(*sign))
		txin.Vin[i].Sign = *sign
	}
	//txin.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return txin, nil
}

/*
创建一个投票押金退还交易
退还按交易为单位，交易的押金全退
@param height 区块高度 voteitems 投票item items 余额 pubs 地址公钥对 witness 见证人 addr 投票地址
*/
func CreateTxVoteOutM(height uint64, voteitems, items []*TxItem, pubs map[string]ed25519.PublicKey, witness *crypto.AddressCoin, addr string, amount, gas uint64, returnaddr crypto.AddressCoin) (*Tx_vote_out, error) {
	//查找余额
	vins := make([]*Vin, 0)
	total := uint64(0)
	//TODO 此处item为投票
	for _, item := range voteitems {
		//TODO txid对应的vout addr. 即上一个输出的out addr
		voutaddr := *item.Addr
		puk, ok := pubs[voutaddr.B58String()]
		if !ok {
			continue
		}

		vin := Vin{
			// Txid: item.Txid,      //UTXO 前一个交易的id
			// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk: puk, //公钥
			//			Sign: *sign,         //签名
		}
		vins = append(vins, &vin)

		total = total + uint64(item.Value)
		if total >= amount+gas {
			break
		}
	}

	// fmt.Println("==============3")
	//资金不够
	//TODO 此处items为余额
	//var returnaddr crypto.AddressCoin //找零退回地址
	if total < amount+gas {
		for _, item := range items {
			// if k == 0 {
			// 	returnaddr = *item.Addr
			// }
			addrstr := *item.Addr
			puk, ok := pubs[addrstr.B58String()]
			if !ok {
				continue
			}

			vin := Vin{
				// Txid: item.Txid,      //UTXO 前一个交易的id
				// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk: puk, //公钥
				//						Sign: *sign,           //签名
			}
			vins = append(vins, &vin)

			total = total + uint64(item.Value)
			if total >= amount+gas {
				break
			}
		}
	}
	// fmt.Println("==============4")
	//余额不够给手续费
	if total < (amount + gas) {
		// fmt.Println("押金不够")
		//押金不够
		return nil, config.ERROR_not_enough
	}
	// fmt.Println("==============5")

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		//为空则转给自己
		dstAddr = returnaddr
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
		dstAddr = crypto.AddressFromB58String(addr)
	}
	// fmt.Println("==============6")

	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   total - gas, //输出金额 = 实际金额 * 100000000
		Address: dstAddr,     //钱包地址
	}
	vouts = append(vouts, &vout)

	//	crateTime := config.TimeNow().Unix()

	var txout *Tx_vote_out
	//
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_out, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  uint64(len(vins)),              //输入交易数量
		Vin:        vins,                           //交易输入
		Vout_total: uint64(len(vouts)),             //输出交易数量
		Vout:       vouts,                          //
		Gas:        gas,                            //交易手续费
		LockHeight: height + 100,                   //锁定高度
		//		CreateTime: crateTime,                      //创建时间
		Comment: []byte{},
	}
	txout = &Tx_vote_out{
		TxBase: base,
	}
	// fmt.Println("==============7")

	//给输出签名，防篡改
	for i, _ := range txout.Vin {
		sign := txout.GetWaitSign(uint64(i))
		if sign == nil {
			return nil, config.ERROR_get_sign_data_fail
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		txout.Vin[i].Sign = *sign
	}
	//txout.BuildHash()
	return txout, nil
}

type BlockVotesVO struct {
	EndHeight uint64
	Group     []GroupVO
}

type GroupVO struct {
	StartHeight    uint64
	EndHeight      uint64
	CommunityVotes []VoteScoreRewadrVO
}

type VoteScoreRewadrVO struct {
	VoteScore              //
	LightVotes []VoteScore //轻节点投票列表
	Reward     uint64      //这个见证人获得的奖励
}

/*
查询历史轻节点投票
*/
func FindLightVote(startHeight, endHeight uint64) (*BlockVotesVO, error) {
	// engine.Log.Info("FindLightVote 111111 %d %d", startHeight, endHeight)

	bvVO := &BlockVotesVO{
		EndHeight: endHeight,
		Group:     make([]GroupVO, 0),
	}

	//查找上一个已经确认的组
	var preBlock *Block
	preGroup := forks.LongChain.WitnessChain.WitnessGroup
	for {
		// engine.Log.Info("FindLightVote 111111")
		preGroup = preGroup.PreGroup
		ok, preGroupBlock := preGroup.CheckBlockGroup(nil)
		if ok {
			preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
			break
		}
	}
	// engine.Log.Info("FindLightVote 111111")
	//找到查询的起始区块
	for {
		// engine.Log.Info("FindLightVote 111111 height:%d", preBlock.Height)
		if preBlock.Height > startHeight {
			if preBlock.PreBlock == nil {
				break
			}
			preBlock = preBlock.PreBlock
		}
		if preBlock.Height < startHeight {
			if preBlock.NextBlock == nil {
				break
			}
			preBlock = preBlock.NextBlock
		}
		if preBlock.Height == startHeight {
			break
		}
	}
	// engine.Log.Info("FindLightVote 111111")
	//找到这个见证人组的第一个见证人
	for {
		// engine.Log.Info("FindLightVote 111111")
		if preBlock.PreBlock == nil {
			//已经是创始区块了
			break
		}
		temp := preBlock.Witness.WitnessBigGroup
		if preBlock.PreBlock.Witness.WitnessBigGroup == temp {
			preBlock = preBlock.PreBlock
		} else {
			break
		}
	}
	// engine.Log.Info("FindLightVote 111111")
	//找到从start高度开始，到最新高度的所有见证人组的首块
	isFind := false
	for ; preBlock != nil && preBlock.NextBlock != nil && preBlock.Height <= endHeight; preBlock = preBlock.NextBlock {
		// engine.Log.Info("FindLightVote 111111 height:%d", preBlock.Height)
		// if preBlock.NextBlock == nil {
		// 	engine.Log.Info("FindLightVote 111111")
		// 	break
		// }
		if isFind && preBlock.NextBlock.Witness.WitnessBigGroup == preBlock.Witness.WitnessBigGroup {
			// preBlock = preBlock.NextBlock
			// engine.Log.Info("FindLightVote 111111")
			continue
		} else {
			isFind = false
		}
		_, txs, err := preBlock.NextBlock.LoadTxs()
		if err != nil {
			// engine.Log.Info("FindLightVote 111111")
			return nil, err
		}
		if txs == nil || len(*txs) <= 0 {
			// preBlock = preBlock.NextBlock
			// if preBlock.Height > endHeight {
			// 	engine.Log.Info("FindLightVote 111111")
			// 	break
			// }
			continue
		}
		//查找这组见证人中的奖励
		reward, ok := (*txs)[0].(*Tx_reward)
		if !ok {
			continue
		}
		isFind = true

		groupVO := new(GroupVO)
		groupVO.CommunityVotes = make([]VoteScoreRewadrVO, 0)
		// groupVO.StartHeight = preBlock.Height
		groupVO.EndHeight = preBlock.Height
		//组装下一个见证人组的投票
		for _, one := range preBlock.Witness.WitnessBigGroup.Witnesses {
			// engine.Log.Info("FindLightVote 111111")
			m := make(map[string]*[]VoteScore)
			for i, two := range one.Votes {
				vo := VoteScore{
					Witness: one.Votes[i].Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
					Addr:    one.Votes[i].Addr,    //投票人地址
					Scores:  one.Votes[i].Scores,  //押金
					Vote:    one.Votes[i].Vote,    //获得票数
				}
				v, ok := m[utils.Bytes2string(*two.Witness)]
				if ok {

				} else {
					temp := make([]VoteScore, 0)
					v = &temp
				}
				*v = append(*v, vo)
				m[utils.Bytes2string(*two.Witness)] = v
			}

			for _, two := range one.CommunityVotes {
				// two.
				// engine.Log.Info("FindLightVote 111111")
				vo := VoteScore{
					Witness: two.Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
					Addr:    two.Addr,    //投票人地址
					Scores:  two.Scores,  //押金
					Vote:    two.Vote,    //获得票数
				}
				vsVOone := VoteScoreRewadrVO{
					VoteScore:  vo,
					LightVotes: make([]VoteScore, 0),
					Reward:     0,
				}
				//查找奖励
				for _, one := range *reward.GetVout() {
					if bytes.Equal(one.Address, *vsVOone.VoteScore.Addr) {
						vsVOone.Reward = one.Value
						break
					}
				}

				v, ok := m[utils.Bytes2string(*two.Addr)]
				if ok {
					vsVOone.LightVotes = *v
				}

				// for i, one := range one.Votes {
				// 	vs := VoteScore{
				// 		Witness: one.Votes[i].Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
				// 		Addr:    one.Votes[i].Addr,    //投票人地址
				// 		Scores:  one.Votes[i].Scores,  //押金
				// 		Vote:    one.Votes[i].Vote,    //获得票数
				// 	}
				// 	vsVOone.LightVotes = append(vsVOone.LightVotes, vs)
				// }
				groupVO.CommunityVotes = append(groupVO.CommunityVotes, vsVOone)
			}
		}
		bvVO.Group = append(bvVO.Group, *groupVO)
		// blocks = append(blocks, preBlock)

		// preBlock = preBlock.NextBlock
		// if preBlock.Height > endHeight {
		// 	engine.Log.Info("FindLightVote 111111")
		// 	break
		// }
	}
	// engine.Log.Info("FindLightVote 111111")
	return bvVO, nil

}

/*
	通过区块高度，查询一个区块头信息
*/
// func FindBlockHead(height uint64) *BlockHead {
// 	bhash, err := db.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
// 	if err != nil {
// 		return nil
// 	}
// 	bs, err := db.Find(*bhash)
// 	if err != nil {
// 		return nil
// 	}
// 	bh, err := ParseBlockHead(bs)
// 	if err != nil {
// 		return nil
// 	}
// 	return bh
// }

/*
	从数据库加载一整个区块，包括区块中的所有交易
*/
// func LoadBlockHeadVOByHash(hash []byte) (*BlockHeadVO, error) {
// 	bhvo := new(BlockHeadVO)
// 	bhvo.Txs = make([]TxItr, 0)
// 	//通过区块hash查找区块头
// 	bs, err := db.Find(hash)
// 	if err != nil {
// 		return nil, err
// 	} else {
// 		bh, err := ParseBlockHead(bs)
// 		if err != nil {
// 			return nil, err
// 		}
// 		bhvo.BH = bh
// 		for _, one := range bh.Tx {
// 			txOne, err := FindTxBase(one)
// 			if err != nil {
// 				return nil, err
// 			}
// 			bhvo.Txs = append(bhvo.Txs, txOne)
// 		}
// 	}
// 	return bhvo, nil
// }

type RewardTotal struct {
	CommunityReward uint64 //社区节点奖励
	LightReward     uint64 //轻节点奖励
	StartHeight     uint64 //统计的开始区块高度
	Height          uint64 //最新区块高度
	IsGrant         bool   //是否可以分发奖励，24小时后才可以分发奖励
	AllLight        uint64 //所有轻节点数量
	RewardLight     uint64 //已经奖励的轻节点数量
	IsNew           bool   //是否是新的奖励，新的奖励需要创建快照
}

/*
奖励结果明细
*/
type RewardTotalDetail struct {
	startHeight uint64
	endHeight   uint64
	RT          *RewardTotal
	RL          *[]sqlite3_db.RewardLight
}

var rewardCountProcessMapLock = new(sync.Mutex)
var rewardCountProcessMap = make(map[string]*RewardTotalDetail) //new(sync.Map) //保存统计社区奖励结果。key:string=社区地址;value:*RewardTotalDetail=奖励结果明细;
type KeyMutex struct {
	mutexMap sync.Map
}

func (k *KeyMutex) Lock(key interface{}) func() {
	mutexValue, _ := k.mutexMap.LoadOrStore(key, &sync.Mutex{})
	mutex := mutexValue.(*sync.Mutex)
	mutex.Lock()
	return func() {
		mutex.Unlock()
	}
}

var createTxMutex = KeyMutex{}

/*
奖励统计
*/
func GetRewardCount(addr *crypto.AddressCoin, startHeight, endHeight uint64) (*RewardTotal, *[]sqlite3_db.RewardLight, error) {
	// engine.Log.Info("2222222222222222")
	currentHeight := forks.LongChain.GetCurrentBlock()
	if endHeight <= 0 || endHeight > currentHeight {
		endHeight = currentHeight
	}

	var rt *RewardTotal
	var rl *[]sqlite3_db.RewardLight
	var err error
	//先查询是否有缓存结果
	have := false
	rewardCountProcessMapLock.Lock()
	rtd, ok := rewardCountProcessMap[utils.Bytes2string(*addr)]
	if ok {
		have = true
		if rtd != nil {
			//有缓存，对比结束区块高度是否相差太远，24小时产生86400个块，也就是24小时之后才可以重新创建新的缓存
			if endHeight > rtd.endHeight && endHeight-rtd.endHeight > uint64(config.Mining_community_reward_time.Seconds()) {
				have = false
			} else {
				rt = rtd.RT
				rl = rtd.RL
			}
		}
	}
	// engine.Log.Info("2222222222222222")
	if !have {
		err = config.ERROR_get_reward_count_sync
		//没有缓存和程序，则启动一个统计程序
		rewardCountProcessMap[utils.Bytes2string(*addr)] = nil
		utils.Go(func() {
			rt, rl, err := RewardCountProcess(addr, startHeight, endHeight)
			if err != nil {
				engine.Log.Info("RewardCountProcess error:%s", err.Error())
				return
			}
			rtd := RewardTotalDetail{
				startHeight: startHeight,
				endHeight:   endHeight,
				RT:          rt,
				RL:          rl,
			}
			rewardCountProcessMapLock.Lock()
			rewardCountProcessMap[utils.Bytes2string(*addr)] = &rtd
			rewardCountProcessMapLock.Unlock()
		}, nil)
	} else {
		// err = config.ERROR_get_reward_count_sync
	}
	rewardCountProcessMapLock.Unlock()
	return rt, rl, err
}

/*
删除缓存
*/
func CleanRewardCountProcessMap(addr *crypto.AddressCoin) {
	rewardCountProcessMapLock.Lock()
	delete(rewardCountProcessMap, utils.Bytes2string(*addr))
	rewardCountProcessMapLock.Unlock()
}

/*
奖励统计明细，消耗时间长
*/
func RewardCountProcess(addr *crypto.AddressCoin, startHeight, endHeight uint64) (*RewardTotal, *[]sqlite3_db.RewardLight, error) {
	// engine.Log.Info("555555555555555555555")
	rewardDBs := make([]sqlite3_db.RewardLight, 0)

	//查询新的统计
	bvvo, err := FindLightVote(startHeight, endHeight)
	if err != nil {
		// engine.Log.Info("555555555555555555555")
		return nil, nil, err
	}
	// engine.Log.Info("555555555555555555555")
	// lightVout := make([]mining.Vout, 0)
	allReward := uint64(0)
	light := uint64(0)
	for _, one := range bvvo.Group {
		for _, voteOne := range one.CommunityVotes {
			//找到这个社区
			if bytes.Equal(*voteOne.Addr, *addr) {
				allReward += voteOne.Reward
				// engine.Log.Info("社区节点:%s 总奖励:%d %d %d", voteOne.Addr.B58String(), voteOne.Reward, one.StartHeight, one.EndHeight)

				for _, two := range voteOne.LightVotes {
					temp := new(big.Int).Mul(big.NewInt(int64(voteOne.Reward)), big.NewInt(int64(two.Vote)))
					value := new(big.Int).Div(temp, big.NewInt(int64(voteOne.Vote)))
					//10%给社区节点
					reward := value.Uint64()
					reward = reward - (reward / 10)
					light += reward
					r := sqlite3_db.RewardLight{
						Addr:         *two.Addr, //轻节点地址
						Reward:       reward,    //自己奖励多少
						Distribution: 0,         //已经分配的奖励
					}
					// engine.Log.Info("统计一个轻节点投票 %s %d", two.Addr.B58String(), reward)
					rewardDBs = append(rewardDBs, r)
				}
				break
			}
		}
	}
	community := allReward - light
	// engine.Log.Info("本次统计的奖励:%d %d allReward:%d community:%d light:%d", startHeight, endHeight, allReward, community, light)
	reward := RewardTotal{
		CommunityReward: community,      //社区节点奖励
		LightReward:     light,          //轻节点奖励
		StartHeight:     startHeight,    //
		Height:          bvvo.EndHeight, //最新区块高度
		//IsGrant:         (bvvo.EndHeight - startHeight) > (config.Mining_community_reward_time / config.Mining_block_time), //是否可以分发奖励，24小时后才可以分发奖励
		IsGrant:     (bvvo.EndHeight - startHeight) > uint64(config.Mining_community_reward_time.Nanoseconds()/config.Mining_block_time.Nanoseconds()), //是否可以分发奖励，24小时后才可以分发奖励 pl time
		AllLight:    0,                                                                                                                                 //所有轻节点数量
		RewardLight: 0,                                                                                                                                 //已经奖励的轻节点数量
	}

	//合并奖励，把相同轻节点地址的奖励合并了
	voutMap := make(map[string]*sqlite3_db.RewardLight)
	for i, _ := range rewardDBs {
		one := rewardDBs[i]
		// engine.Log.Info("统计一个轻节点投票 222 %s", one.Addr)
		if one.Reward == 0 {
			continue
		}
		v, ok := voutMap[utils.Bytes2string(one.Addr)]
		if ok {
			v.Reward = v.Reward + one.Reward
			continue
		}
		voutMap[utils.Bytes2string(one.Addr)] = &(rewardDBs)[i]
	}
	vouts := make([]sqlite3_db.RewardLight, 0)
	for _, v := range voutMap {
		// engine.Log.Info("统计一个轻节点投票 333 %s", v.Addr)
		vouts = append(vouts, *v)
	}

	//给社区节点的10%奖励
	communityReward := sqlite3_db.RewardLight{
		Addr:         *addr,     //轻节点地址
		Reward:       community, //自己奖励多少
		Distribution: 0,         //已经分配的奖励
	}
	vouts = append(vouts, communityReward)

	//统计所有轻节点的数量
	reward.AllLight = uint64(len(vouts))
	// engine.Log.Info("统计一个轻节点投票 333 %s", v.Addr)
	// engine.Log.Info("555555555555555555555")
	return &reward, &vouts, nil
}

/*
从数据库获取未分配完的奖励和统计
*/
func FindNotSendReward(addr *crypto.AddressCoin) (*sqlite3_db.SnapshotReward, *[]sqlite3_db.RewardLight, error) {
	// startHeight := uint64(0)
	//查询最新的快照
	s, err := new(sqlite3_db.SnapshotReward).Find(*addr)
	if err != nil {
		return nil, nil, err
	}
	if s == nil {
		return nil, nil, nil
	} else {
		rewardNotSend, err := new(sqlite3_db.RewardLight).FindNotSend(s.Id)
		if err != nil {
			return nil, nil, err
		}
		return s, rewardNotSend, nil

	}
}

/*
创建轻节点奖励快照
*/
func CreateRewardCount(addr crypto.AddressCoin, rt *RewardTotal, rs []sqlite3_db.RewardLight) error {
	ss := &sqlite3_db.SnapshotReward{
		Addr:        addr,                                //社区节点地址
		StartHeight: rt.StartHeight,                      //快照开始高度
		EndHeight:   rt.Height,                           //快照结束高度
		Reward:      rt.LightReward + rt.CommunityReward, //此快照的总共奖励
		LightNum:    uint64(len(rs)),                     //
	}
	// engine.Log.Info("创建快照:%d %d", ss.Reward, len(rs))
	// for _, one := range rs {
	// 	addr := crypto.AddressCoin(one.Addr)
	// 	engine.Log.Info("创建快照奖励明细:%s %d", addr.B58String(), one.Reward)
	// }

	err := new(sqlite3_db.SnapshotReward).Add(ss)
	if err != nil {
		return err
	}

	ss, err = new(sqlite3_db.SnapshotReward).Find(addr)
	if err != nil {
		return err
	}

	count := uint64(0)
	for _, one := range rs {
		count++
		one.Sort = count
		one.SnapshotId = ss.Id
		err := new(sqlite3_db.RewardLight).Add(&one)
		if err != nil {
			//TODO 事务回滚
			return err
		}
	}
	return nil
}

/*
分配奖励
@height    uint64    当前区块高度，方便计算180天线性释放时间
*/
func DistributionReward(addr *crypto.AddressCoin, notSend *[]sqlite3_db.RewardLight, gas uint64, pwd string, cs *CommunitySign, startHeight, endHeight, currentHeight uint64) error {
	// engine.Log.Info("DistributionReward %+v", notSend)
	if notSend == nil || len(*notSend) <= 0 {
		return nil
	}

	// notSend

	// max := len(*notSend)

	if len(*notSend) > config.Wallet_community_reward_max {
		temp := (*notSend)[:config.Wallet_community_reward_max]
		notSend = &temp
		// max = config.Wallet_community_reward_max
	}

	// var tx TxItr
	// var err error
	// for {
	//计算平摊的手续费
	value := new(big.Int).Div(big.NewInt(int64(gas)), big.NewInt(int64(len(*notSend)))).Uint64()

	payNum := make([]PayNumber, 0)
	for i := 0; i < len(*notSend); i++ {
		one := (*notSend)[i]
		addr := crypto.AddressCoin(one.Addr)
		//engine.Log.Info("奖励地址:%s %d %d %d", addr.B58String(), one.Reward, value, one.Reward-value)
		//pns := LinearRelease180DayForLight(addr, one.Reward-value, currentHeight)
		var rewardNew uint64
		if i == len(*notSend)-1 {
			yu := gas - value*uint64(len(*notSend))
			if yu < 0 {
				yu = 0
			}
			rewardNew = one.Reward - value - yu
		} else {
			rewardNew = one.Reward - value
		}

		pns := LinearRelease180DayForLight(addr, rewardNew, currentHeight)
		payNum = append(payNum, pns...)
	}
	tx, err := SendToMoreAddressByPayload(addr, payNum, gas, pwd, cs, startHeight, endHeight)
	if err != nil {
		// if err.Error() == config.ERROR_pay_vin_too_much.Error() {
		// if max <= 1 {
		// 	engine.Log.Error(err.Error())
		// 	return err
		// }
		// max = max / 2
		// // continue
		// }
		engine.Log.Error(err.Error(), "")
		return err
	} else {
		// break
	}
	// }
	//修改数据库，分配奖励修改为上链中
	for i := 0; i < len(*notSend); i++ {
		one := (*notSend)[i]
		one.Txid = *tx.GetHash()
		one.LockHeight = tx.GetLockHeight() // LockHeight
		err := one.UpdateTxid(one.Id)
		if err != nil {
			engine.Log.Error(err.Error(), "")
		}
	}
	return err
}

// pl
func LinearRelease180DayForLight(addr crypto.AddressCoin, total uint64, height uint64) []PayNumber {
	// return LinearRelease180DayForLightOld(addr, total, height)
	pns := make([]PayNumber, 0)
	pnOne := PayNumber{
		Address: addr,  //转账地址
		Amount:  total, //转账金额
	}
	pns = append(pns, pnOne)

	return pns
}

/*
180天线性释放给轻节点
*/
func LinearRelease180DayForLightOld(addr crypto.AddressCoin, total uint64, height uint64) []PayNumber {

	//TODO 处理好不能被180整除的情况

	pns := make([]PayNumber, 0)
	//25%直接到账
	first25 := new(big.Int).Div(big.NewInt(int64(total)), big.NewInt(int64(4)))
	//剩下的75%
	surplus := new(big.Int).Sub(big.NewInt(int64(total)), first25)

	// engine.Log.Error("180天线性释放 %d %d %d", total, first25.Uint64(), surplus.Uint64())

	pnOne := PayNumber{
		Address: addr,             //转账地址
		Amount:  first25.Uint64(), //转账金额
		// FrozenHeight: height + uint64(i*intervalHeight), //冻结高度
	}

	pns = append(pns, pnOne)

	dayOne := new(big.Int).Div(surplus, big.NewInt(int64(18))).Uint64()

	// dayOne := new(big.Int).Div(big.NewInt(int64(total)), big.NewInt(int64(180))).Uint64()
	intervalHeight := 60 * 60 * 24 * 10 / 10

	totalUse := uint64(0)
	for i := 0; i < 18; i++ {
		pnOne := PayNumber{
			Address:      addr,                                  //转账地址
			Amount:       dayOne,                                //转账金额
			FrozenHeight: height + uint64((i+1)*intervalHeight), //冻结高度
		}
		pns = append(pns, pnOne)
		totalUse = totalUse + dayOne
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if totalUse < surplus.Uint64() {
		// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
		pns[len(pns)-1].Amount = pns[len(pns)-1].Amount + (surplus.Uint64() - totalUse)
	}
	return pns
}

/*
合约交易功能
*/
func ContractTx(srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment, source string,
	class uint64, gasPrice uint64, abi string) (*Tx_Contract, error) {

	//utils.Log.Info().Interface("srcAddress", srcAddress).Interface("address", address).Interface("amount", amount).
	//	Interface("gas", gas).Interface("frozenHeight", frozenHeight).Interface("pwd", pwd).
	//	Interface("comment", comment).Interface("source", source).Interface("class", class).
	//	Interface("gasPrice", class).Interface("abi", class).Send()

	// return nil, errors.New("hahahaha")
	unlock := createTxMutex.Lock(srcAddress.B58String())
	txpay, err := CreateTxContractNew(srcAddress, address, amount, gas, frozenHeight, pwd, comment, source, class, gasPrice, abi)
	unlock()
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	// engine.Log.Info("create tx finish!")

	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		GetLongChain().Balance.DelLockTx(txpay)
		return nil, errors.Wrap(err, "add tx fail!")
	}

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
获取社区节点列表
*/
func GetCommunityListSortNew() []*VoteScoreVO {
	res := []*VoteScoreVO{}
	list := precompiled.GetAllCommunity(config.Area.Keystore.GetCoinbase().Addr)
	for _, v := range list {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		witCoin := evmutils.AddressToAddressCoin(v.Wit.Bytes())
		res = append(res, &VoteScoreVO{
			Addr:    coin.B58String(),
			Witness: witCoin.B58String(),
			Payload: v.Name,
			Score:   v.Score.Uint64(),
			Vote:    v.Vote.Uint64(),
		})
	}
	return res
}

/*
获取社区节点列表
*/
func GetCommunityRewardListSortNew() []*VoteScoreV1 {
	res := []*VoteScoreV1{}
	list := precompiled.GetAllCommunityReward(config.Area.Keystore.GetCoinbase().Addr)
	for _, v := range list {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		witCoin := evmutils.AddressToAddressCoin(v.Wit.Bytes())
		res = append(res, &VoteScoreV1{
			Addr:    coin.B58String(),
			Witness: witCoin.B58String(),
			Payload: v.Name,
			Score:   v.Score.Uint64(),
			Vote:    v.Vote.Uint64(),
			Reward:  v.Reward.Uint64(),
			Rate:    v.Rate,
		})
	}
	return res
}

/*
获取社区节点押金列表从合约统计
*/
func GetDepositCommunityListNew() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	list := []crypto.AddressCoin{}
	for _, one := range Area.Keystore.GetAddrAll() {
		list = append(list, one.Addr)
	}
	clist := precompiled.GetCommunityList(list)
	for _, v := range clist {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		witCoin := evmutils.AddressToAddressCoin(v.Wit.Bytes())
		items = append(items, &DepositInfo{
			WitnessAddr: witCoin,
			SelfAddr:    coin,
			Value:       v.Score.Uint64(),
		})
	}
	return items
}

/*
获取轻节点押金列表从合约统计
*/
func GetDepositLightListNew() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	list := []crypto.AddressCoin{}
	for _, one := range Area.Keystore.GetAddrAll() {
		list = append(list, one.Addr)
	}
	clist := precompiled.GetLightList(list)
	for _, v := range clist {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		witCoin := evmutils.AddressToAddressCoin(v.C.Bytes())
		items = append(items, &DepositInfo{
			WitnessAddr: witCoin,
			SelfAddr:    coin,
			Value:       v.Score.Uint64(),
			Name:        v.Cname,
		})
	}
	return items
}

/*
获取轻节点投票押金列表通过合约
*/
func GetDepositVoteListNew() []*DepositInfo {
	items := make([]*DepositInfo, 0)
	list := []crypto.AddressCoin{}
	for _, one := range Area.Keystore.GetAddrAll() {
		list = append(list, one.Addr)
	}
	clist := precompiled.GetLightList(list)
	for _, v := range clist {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		witCoin := evmutils.AddressToAddressCoin(v.C.Bytes())
		items = append(items, &DepositInfo{
			WitnessAddr: witCoin,
			SelfAddr:    coin,
			Value:       v.Vote.Uint64(),
			Name:        v.Cname,
		})
	}
	return items
}
func GetFrozenReward(addr crypto.AddressCoin) uint64 {
	return precompiled.GetMyFrozenReward(addr)
}

type VoteScoreOut struct {
	Witness string //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
	Addr    string //投票人地址
	Payload string //
	Score   uint64 //押金
	Vote    uint64 //获得的投票
	Name    string //自己的名字
}

/*
获取轻节点列表-合约
*/
func GetLightListSortNew(page, pageSize int) []*VoteScoreOut {

	res := []*VoteScoreOut{}

	start := (page - 1) * pageSize
	end := start + pageSize
	//if start > len(list) {
	//	return res
	//}
	//if end > len(list) {
	//	end = len(list)
	//}
	list := precompiled.GetAllLight(config.Area.Keystore.GetCoinbase().Addr, big.NewInt(int64(start)), big.NewInt(int64(end)))
	for _, v := range list {
		coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		var witAddr string
		if v.C != evmutils.ZeroAddress {
			witCoin := evmutils.AddressToAddressCoin(v.C.Bytes())
			witAddr = witCoin.B58String()
		}

		res = append(res, &VoteScoreOut{
			Addr:    coin.B58String(),
			Witness: witAddr,
			Payload: v.Cname,
			Score:   v.Score.Uint64() - v.Vote.Uint64(),
			Vote:    v.Vote.Uint64(),
			Name:    v.Name,
		})
	}
	return res
}

func SaveRewardLog(contractEvents []*go_protos.ContractEvent) {
	//start := config.TimeNow()
	rewardHistories := make([]db2.KVPair, 0)
	for _, v := range contractEvents {
		eventInfo := &go_protos.ContractEventInfo{
			//BlockHeight:     int64(bhvo.BH.Height),
			Topic:           v.Topic,
			TxId:            v.TxId,
			ContractAddress: v.ContractAddress,
			EventData:       v.EventData,
		}
		l := precompiled.UnPackLogRewardHistoryLog(eventInfo)
		keyValue := append(l.Into[:], l.From.Bytes()...)
		key := config.BuildRewardHistory(keyValue)
		rewardByte, err := db.LevelDB.Get(key)
		if err != nil {
			engine.Log.Error("get reward byte error")
			continue
		}
		reward := uint64(0)
		if rewardByte != nil {
			reward = utils.BytesToUint64(rewardByte)
		}

		rewardHistories = append(rewardHistories, db2.KVPair{
			Key:   config.BuildRewardHistory(keyValue),
			Value: utils.Uint64ToBytes(l.Reward.Uint64() + reward),
		})
	}
	if len(rewardHistories) > 0 {
		db.LevelDB.MSet(rewardHistories...)
	}
	//engine.Log.Info("end reward log start", config.TimeNow().Sub(start).Milliseconds())
}
