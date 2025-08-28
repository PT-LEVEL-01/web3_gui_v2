package chain_plus

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"github.com/go-xorm/xorm"
	jsoniter "github.com/json-iterator/go"
	"sort"
	"strings"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining"
	"web3_gui/chain/sqlite3_db"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/chain_boot/model"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	keystconfig "web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetAddrTxIndex(addr string) uint64 {
	// 获取最大索引
	chain := mining.GetLongChain()
	if chain != nil {
		return chain.GetAccountHistoryBalanceCount(addr)
	}
	return 0
}

/*
获取地址交易
*/
func TransactionRecord(addr string, page, pageSize int) (uint64, []map[string]interface{}) {
	pageSizeInt := pageSize
	if pageSizeInt > chainbootconfig.PageSizeLimit {
		pageSizeInt = chainbootconfig.PageSizeLimit
	}
	if pageSizeInt == 0 {
		pageSizeInt = chainbootconfig.PageTotalDefault
	}
	startIndex := uint64(page * pageSizeInt)
	endIndex := uint64((page + 1) * pageSizeInt)

	txids := make([][]byte, 0)
	maxIndex := GetAddrTxIndex(addr)
	count := maxIndex

	if maxIndex > startIndex {
		startIndex = maxIndex - startIndex
	} else {
		startIndex = 0
	}

	if maxIndex > endIndex {
		endIndex = maxIndex - endIndex
	} else {
		endIndex = 0
	}

	for j := startIndex; j > endIndex; j-- {
		addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(addr))...)
		indexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(indexBs, j)
		txid, err := db.LevelTempDB.HGet(addrkey, indexBs)
		if err != nil {
			continue
		}

		txids = append(txids, txid)
	}

	outMap := sync.Map{}
	wg := sync.WaitGroup{}
	for _, txid := range txids {
		wg.Add(1)
		go func(txid []byte) {
			defer wg.Done()
			if data := parseOneTx(txid); data != nil {
				outMap.Store(utils.Bytes2string(txid), data)
			}
		}(txid)
	}
	wg.Wait()

	items := make([]map[string]interface{}, 0)
	for _, txid := range txids {
		if item, ok := outMap.Load(utils.Bytes2string(txid)); ok {
			items = append(items, item.(map[string]interface{}))
		}
	}

	//out := make(map[string]interface{})
	//out["count"] = count
	//out["data"] = items
	//res, err = model.Tojson(out)
	return count, items
}

// 解析交易
func parseOneTx(txid []byte) map[string]interface{} {
	txItr, code, blockHash := mining.FindTx(txid)
	if txItr == nil {
		return nil
	}

	if blockHash == nil {
		return nil
	}

	//获取区块信息
	bh, err := mining.LoadBlockHeadByHash(blockHash)
	if err != nil {
		return nil
	}

	var blockHashStr string
	if blockHash != nil {
		blockHashStr = hex.EncodeToString(*blockHash)
	}

	item := JsonMethod(txItr.GetVOJSON())
	item["blockhash"] = blockHashStr
	item["timestamp"] = bh.Time
	out := make(map[string]interface{})
	out["txinfo"] = item
	out["timestamp"] = bh.Time
	out["blockhash"] = blockHashStr
	out["blockheight"] = bh.Height
	out["upchaincode"] = code
	return out
}

func JsonMethod(content interface{}) map[string]interface{} {
	var name map[string]interface{}
	if mashalContent, err := json.Marshal(content); err != nil {
		//engine.Log.Error("2364行错误%s", err.Error())
	} else {
		d := json.NewDecoder(bytes.NewReader(mashalContent))
		d.UseNumber()
		if err := d.Decode(&name); err != nil {
			//engine.Log.Error("2369行错误%s", err.Error())
		} else {
			for k, v := range name {
				name[k] = v
			}
		}
	}
	return name
}

/*
检查交易
*/
func CheckTxBase64(txBase64StdStr string, checkBalance bool) (mining.TxItr, utils.ERROR) {
	txjsonBs, err := base64.StdEncoding.DecodeString(txBase64StdStr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	txItr, err := mining.ParseTxBaseProto(0, &txjsonBs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}

	vins := *txItr.GetVin()
	if len(vins) == 0 {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_Tx_Vin_fail, "")
	}
	utils.Log.Info().Bool("是否检查gas", checkBalance).Send()
	if checkBalance {
		//验证锁定高度
		if err = txItr.CheckLockHeight(mining.GetHighestBlock()); err != nil {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_Tx_LockHeight_TooLow, "")
		}
		//检查Nonce
		tm := mining.GetLongChain().GetTransactionManager()
		_, err = tm.CheckNonce(txItr)
		if err != nil {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_Tx_Nonce_fail, "")
		}
		// 支付类检查是否免gas
		if err := mining.CheckTxPayFreeGas(txItr); err != nil {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "")
		}
		spend := txItr.GetSpend()
		src := vins[0].GetPukToAddr()
		if txItr.Class() == config.Wallet_tx_type_address_transfer {
			tx := txItr.(*mining.TxAddressTransfer)
			gas := txItr.GetGas()
			total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(src, gas)
			if total < gas {
				//资金不够
				return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
			}
			vouts := *txItr.GetVout()
			amount := vouts[0].Value
			payTotal, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&tx.PayAddress, amount)
			if amount == 0 {
				amount = payTotal
			}
			if payTotal < amount {
				//资金不够
				return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
			}
		} else {
			total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(src, spend)
			if total < spend {
				//资金不够
				return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
			}
		}
		// engine.Log.Info("rpc transaction received %s", hex.EncodeToString(*txItr.GetHash()))
	}
	if err = txItr.CheckSign(); err != nil {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_Tx_Sign_fail, "")
	}
	//验证domain
	if !txItr.CheckDomain() {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_Tx_Domain_fail, "")
	}
	return txItr, utils.NewErrorSuccess()
}

/*
获得全网/出块/候选 见证人
*/
func WitnessesListWithRangeV1(wits []*mining.BackupWitness, start, end int) ([]model.WitnessInfoList, int, utils.ERROR) {
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		return nil, 0, ERR
	}
	wvos := []model.WitnessInfoList{}
	total := len(wits)
	for _, v := range wits {
		addBlockCount, addBlockReward := chain.Balance.GetAddBlockNum(v.Addr.B58String())
		wvo := model.WitnessInfoList{
			Addr:           v.Addr.B58String(),                   //见证人地址
			Payload:        mining.FindWitnessName(*v.Addr),      //名字
			Score:          mining.GetDepositWitnessAddr(v.Addr), //质押量
			Vote:           chain.Balance.GetWitnessVote(v.Addr), // 总票数
			AddBlockCount:  addBlockCount,
			AddBlockReward: addBlockReward,
			Ratio:          float64(chain.Balance.GetDepositRate(v.Addr)),
		}
		wvos = append(wvos, wvo)
	}
	// 按投票数排序
	sort.Sort(model.WitnessListSort(wvos))
	// 分页
	outs := []model.WitnessInfoList{}
	for i, one := range wvos {
		if i >= start && i < end {
			outs = append(outs, one)
		}
	}
	return outs, total, utils.NewErrorSuccess()
}

/*
获取一个社区的累计未分配奖励
*/
func GetCommunityRewardPool(addr coin_address.AddressCoin) (*mining.RewardTotal, utils.ERROR) {
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		return nil, ERR
	}
	currentHeight := chain.GetCurrentBlock()

	addrC := crypto.AddressCoin(addr)
	if mining.GetAddrState(addrC) != 2 {
		//此地址不是社区节点
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_address_not_community, "")
		return nil, ERR
	}

	// engine.Log.Info("11111111111111111111")
	// engine.Log.Info("11111111111111111111")
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(&addrC)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询代上链的奖励是否已经上链
		*notSend, _ = CheckTxUpchain(*notSend, currentHeight)
		// engine.Log.Info("查询到的奖励:%d", ns.Reward)
		//有未分配完的奖励
		community := ns.Reward / 10
		light := ns.Reward - community
		lightNum := ns.LightNum
		if lightNum > 0 {
			lightNum--
		}
		rewardTotal := mining.RewardTotal{
			CommunityReward: community,                           //社区节点奖励
			LightReward:     light,                               //轻节点奖励
			StartHeight:     ns.StartHeight,                      //
			Height:          ns.EndHeight,                        //最新区块高度
			IsGrant:         false,                               //是否可以分发奖励，24小时候才可以分发奖励
			AllLight:        lightNum,                            //所有轻节点数量
			RewardLight:     ns.LightNum - uint64(len(*notSend)), //已经奖励的轻节点数量
			IsNew:           false,
		}
		// engine.Log.Info("11111111111111111111")
		// engine.Log.Info("此次未分配完的奖励:%d %d Reward:%d light:%d community:%d allReward:%d", ns.StartHeight,
		//ns.EndHeight, ns.Reward, light, community, light+community)
		//res, err = model.Tojson(rewardTotal)
		return &rewardTotal, utils.NewErrorSuccess()
	}
	if ns == nil {
		//需要加载以前的奖励快照
		FindCommunityStartHeightByAddr(addrC)
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_community_reward_load_async, "")
	}
	// engine.Log.Info("11111111111111111111")
	startHeight := ns.EndHeight + 1

	// engine.Log.Info("333333333333333333")
	//奖励都分配完了，查询新的奖励
	rt, _, err := mining.GetRewardCount(&addrC, startHeight, 0)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			//res, err = model.Errcode(RewardCountSync, err.Error())
			//return
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_community_reward_load_async, "")
		}
		// engine.Log.Info("444444444444444")
		//res, err = model.Errcode(model.Nomarl, err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	rt.IsNew = true

	lightNum := rt.AllLight
	if lightNum > 0 {
		lightNum--
	}
	rewardTotal := mining.RewardTotal{
		CommunityReward: rt.CommunityReward, //社区节点奖励
		LightReward:     rt.LightReward,     //轻节点奖励
		StartHeight:     rt.StartHeight,     //
		Height:          rt.Height,          //最新区块高度
		IsGrant:         rt.IsGrant,         //是否可以分发奖励，24小时候才可以分发奖励
		AllLight:        lightNum,           //所有轻节点数量
		RewardLight:     rt.RewardLight,     //已经奖励的轻节点数量
		IsNew:           rt.IsNew,
	}
	return &rewardTotal, utils.NewErrorSuccess()
}

/*
检查交易是否上链，未上链，超过上链高度，则取消上链。
已上链的则修改数据库
*/
func CheckTxUpchain(notSend []sqlite3_db.RewardLight, currentHeight uint64) ([]sqlite3_db.RewardLight, bool) {
	// engine.Log.Info("333333333333")
	txidUpchain := make(map[string]int)                     //保存已经上链的交易
	txidNotUpchain := make(map[string]int)                  //保存未上链的交易
	resultUpchain := make([]sqlite3_db.RewardLight, 0)      //保存需要修改为上链的奖励记录
	resultUnLockHeight := make([]sqlite3_db.RewardLight, 0) //保存需要回滚的奖励记录
	resultReward := make([]sqlite3_db.RewardLight, 0)       //返回的未上链的结果
	haveNotUpchain := false                                 //保存是否存在未上链的奖励
	for i, _ := range notSend {
		one := notSend[i]
		if one.Txid == nil {
			resultReward = append(resultReward, one)
			continue
		}
		//查询交易是否上链
		//先查询缓存
		_, ok := txidUpchain[utils.Bytes2string(one.Txid)]
		if ok {
			//上链了
			resultUpchain = append(resultUpchain, one)
			continue
		}
		_, ok = txidNotUpchain[utils.Bytes2string(one.Txid)]
		if ok {
			//未上链，判断是否超过冻结高度
			if one.LockHeight < currentHeight {
				//回滚
				resultUnLockHeight = append(resultUnLockHeight, one)
			} else {
				//等待
				resultReward = append(resultReward, one)
				haveNotUpchain = true
			}
			continue
		}

		//缓存没有，只有去数据库查询
		txItr, err := mining.LoadTxBase(one.Txid)
		blockhash, berr := db.GetTxToBlockHash(&one.Txid)
		// if berr != nil || blockhash == nil {
		// 	return config.ERROR_tx_format_fail
		// }
		// txItr, err := mining.FindTxBase(one.TokenId)
		if err != nil || txItr == nil || berr != nil || blockhash == nil {
			// if err != nil || txItr == nil || txItr.GetBlockHash() == nil {
			txidNotUpchain[utils.Bytes2string(one.Txid)] = 0
			//未上链，判断是否超过冻结高度
			if one.LockHeight < currentHeight {
				//回滚
				resultUnLockHeight = append(resultUnLockHeight, one)
			} else {
				//等待
				resultReward = append(resultReward, one)
				haveNotUpchain = true
			}
		} else {
			//上链了，修改状态
			txidUpchain[utils.Bytes2string(one.Txid)] = 0
			resultUpchain = append(resultUpchain, one)
		}
	}
	//奖励回滚
	if len(resultUnLockHeight) > 0 {
		ids := make([]uint64, 0)
		for _, one := range resultUnLockHeight {
			ids = append(ids, one.Id)
		}
		err := new(sqlite3_db.RewardLight).RemoveTxid(ids)
		if err != nil {
			//engine.Log.Error(err.Error())
			utils.Log.Error().Err(err).Send()
		}
	}
	//奖励修改为已经上链
	if len(resultUpchain) > 0 {
		var err error
		for _, one := range resultUpchain {
			err = one.UpdateDistribution(one.Id, one.Reward)
			if err != nil {
				//engine.Log.Error(err.Error())
				utils.Log.Error().Err(err).Send()
			}
		}
	}
	return resultReward, haveNotUpchain
}

var findCommunityStartHeightByAddrOnceLock = new(sync.Mutex)
var findCommunityStartHeightByAddrOnce = make(map[string]bool) // new(sync.Once)

/*
查找一个地址成为社区节点的开始高度
*/
func FindCommunityStartHeightByAddr(addr crypto.AddressCoin) {
	// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
	ok := false
	findCommunityStartHeightByAddrOnceLock.Lock()
	_, ok = findCommunityStartHeightByAddrOnce[utils.Bytes2string(addr)]
	findCommunityStartHeightByAddrOnce[utils.Bytes2string(addr)] = false
	findCommunityStartHeightByAddrOnceLock.Unlock()
	if ok {
		// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
		return
	}
	// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
	utils.Go(func() {
		bhHash, err := db.LevelTempDB.Find(config.BuildCommunityAddrStartHeight(addr))
		if err != nil {
			//engine.Log.Error("this addr not community:%s error:%s", addr.B58String(), err.Error())
			utils.Log.Error().Err(err).Send()
			return
		}
		var sn *sqlite3_db.SnapshotReward
		//判断数据库是否有快照记录
		sn, _, err = mining.FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			//engine.Log.Error("querying database Error %s", err.Error())
			utils.Log.Error().Err(err).Send()
			return
		}
		//有记录，就不再恢复历史记录了
		if sn != nil {
			// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
			return
		}
		//建立多个快照，后面一次性保存
		snapshots := make([]sqlite3_db.SnapshotReward, 0)

		var bhvo *mining.BlockHeadVO
		var txItr mining.TxItr
		// var err error
		// var addrTx crypto.AddressCoin
		var ok bool
		// var cs *mining.CommunitySign
		var have bool
		for {
			if bhHash == nil || len(*bhHash) <= 0 {
				// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
				break
			}
			bhvo, err = mining.LoadBlockHeadVOByHash(bhHash)
			if err != nil {
				//engine.Log.Error("findCommunityStartHeightByAddr load blockhead error:%s", err.Error())
				utils.Log.Error().Err(err).Send()
				return
			}
			// engine.Log.Info("findCommunityStartHeightByAddr Community start count block height:%d", bhvo.BH.Height)
			bhHash = &bhvo.BH.Nextblockhash
			if len(snapshots) <= 0 {
				//创建一个空奖励快照
				snapshotsOne := sqlite3_db.SnapshotReward{
					Addr:        addr,           //社区节点地址
					StartHeight: bhvo.BH.Height, //快照开始高度
					EndHeight:   bhvo.BH.Height, //快照结束高度
					Reward:      0,              //此快照的总共奖励
					LightNum:    0,              //奖励的轻节点数量
				}
				snapshots = append(snapshots, snapshotsOne)
			}
			for _, txItr = range bhvo.Txs {
				//判断交易类型
				if txItr.Class() != config.Wallet_tx_type_voting_reward {
					continue
				}

				_, ok = config.Area.Keystore.FindPuk((*txItr.GetVin())[0].Puk)

				// //检查签名
				// addrTx, ok, cs = mining.CheckPayload(txItr)
				// if !ok {
				// 	//签名不正确
				// 	continue
				// }
				// //判断地址是否属于自己
				// _, ok = keystore.FindAddress(addrTx)
				if !ok {
					//签名者地址不属于自己
					continue
				}

				txVoteReward := txItr.(*mining.Tx_Vote_Reward)

				//判断有没有这个快照
				have = false
				for _, one := range snapshots {
					if bytes.Equal(addr, one.Addr) && one.StartHeight == txVoteReward.StartHeight && one.EndHeight == txVoteReward.EndHeight {
						have = true
						break
					}
				}
				if have {
					continue
				}
				//没有这个快照就创建
				snapshotsOne := sqlite3_db.SnapshotReward{
					Addr:        addr,                     //社区节点地址
					StartHeight: txVoteReward.StartHeight, //快照开始高度
					EndHeight:   txVoteReward.EndHeight,   //快照结束高度
					Reward:      0,                        //此快照的总共奖励
					LightNum:    0,                        //奖励的轻节点数量
				}

				snapshots = append(snapshots, snapshotsOne)
			}
		}
		// engine.Log.Info("findCommunityStartHeightByAddr Community end count block!")
		//开始保存快照
		for i, _ := range snapshots {
			one := snapshots[i]
			err = one.Add(&one)
			if err != nil {
				//engine.Log.Info(err.Error())
				utils.Log.Error().Err(err).Send()
			}
		}
		// engine.Log.Info("findCommunityStartHeightByAddr Community finish count block!")
	}, nil)
}

/*
社区手动分发奖励
社区向自己的所有的轻节点分发奖励，并清零社区奖励池
address 社区节点地址
*/
func CreateTxVoteRewardByCommunity(address coin_address.AddressCoin, gas uint64, pwd, comment string) (*mining.Tx_Vote_Reward, utils.ERROR) {
	addr := crypto.AddressCoin(address)
	role := mining.GetAddrState(addr)
	if role != 2 {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_address_not_community, "")
	}
	txpay, err := mining.CreateTxVoteRewardNew(&addr, gas, pwd, comment)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			return nil, utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == config.ERROR_not_enough.Error() {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return nil, utils.NewErrorSysSelf(err)
	}

	return txpay, utils.NewErrorSuccess()
}

/*
轻节点手动分发奖励
社区向自己的所有的轻节点分发奖励，并清零社区奖励池
address 轻节点地址
*/
func CreateTxVoteRewardByLight(address coin_address.AddressCoin, gas uint64, pwd, comment string) (*mining.Tx_Vote_Reward, utils.ERROR) {
	addr := crypto.AddressCoin(address)

	role := mining.GetAddrState(addr)
	if role != 3 {
		return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_address_not_light, "")
	}

	txpay, err := mining.CreateTxVoteRewardNew(&addr, gas, pwd, comment)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			return nil, utils.NewErrorBus(keystconfig.ERROR_code_password_fail, "")
		} else if err.Error() == config.ERROR_not_enough.Error() {
			return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_BalanceNotEnough, "")
		}
		return nil, utils.NewErrorSysSelf(err)
	}

	return txpay, utils.NewErrorSuccess()
}
