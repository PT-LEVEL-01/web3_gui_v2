package light

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain/rpc"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

func RegisterMinerMsg() {
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETADDRESSTX, GetAddressTx)                                         //获取地址交易
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETALLTX, GetAllTx)                                                 //获取所有交易
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETMINERBLOCK, GetMinerBlock)                                       //获取挖矿交易
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL, GetLightNodeDetail)                             //获取轻节点详情
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL, GetWitnessNodeDetail)                         //获取见证者节点详情
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL, GetCommunityNodeDetail)                     //获取社区节点详情
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCOMMUNITYLISTFORMINER, GetCommunityListForMiner)                 //获取社区节点列表
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE, GetWitnessesListWithPage)                 //获得候选见证人和见证人列表 带翻页
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER, GetWitnessesInfoForMiner)                 //获得某个候选见证人和见证人信息
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSESLISTV0WITHPAGE, GetWitnessesListV0WithPage)             //获得见证人列表 带翻页
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSESBACKUPLISTV0WITHPAGE, GetWitnessesBackUpListV0WithPage) //获得候选见证人列表 带翻页
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETDEPOSITNUMFORALL, GetDepositNumForAll)                           //获取总的质押量
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETABIINPUT, GetAbiInput)                                           //获取批量交易input
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETTXGAS, GetTxGas)                                                 //根据交易类型获取交易gas 默认普通交易
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCONTRACTEVENT, GetContractEvent)                                 //根据区块高度和交易hash获取奖励合约事件 默认普通交易
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_ADDERC20, AddErc20)                                                 //收藏代币
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_DELERC20, DelErc20)                                                 //移除代币
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETERC20, GetErc20)                                                 //获取代币
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETERC20VALUE, GetErc20SumBalance)                                  //获取代币
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SEARCHERC20, SearchErc20)                                           //搜索代币
	//Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE, GetErc20SumBalance)                             //获取多代币地址的总余额
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_TRANSFERERC20, TransferErc20)         //代币转账
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETTOKENBALANCE, GetTokenBalance)     //获取代币余额
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETTOKENBALANCES, GetTokenBalances)   //获取多地址代币余额
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETREWARDCONTRACT, GetRewardContract) //获取多地址代币余额
}

/*
获取地址交易
*/
func GetAddressTx(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETADDRESSTX_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETADDRESSTX_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETADDRESSTX_REV, pkg(model.TypeWrong, "address"))
		return
	}

	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))

	start := (pageInt - 1) * pageSizeInt
	//addrkey := []byte(config.Address_history_tx + "_" + strings.ToLower(addr.(string)))
	addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(addr.(string)))...)

	index := rpc.GetAddrTxIndex(addr.(string))
	//balanceHistory := mining.NewBalanceHistory()
	data := make([]map[string]interface{}, 0)
	for i := 0; i < pageSizeInt; i++ {
		indexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(indexBs, index-uint64(start)-uint64(i))

		txid, err := db.LevelTempDB.HGet(addrkey, indexBs)
		if err != nil {
			continue
		}
		if len(txid) != 0 {
			txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(txid)
			if txItr == nil {
				//customTx, err := balanceHistory.GetCustomTx(txid)
				//if err != nil {
				//	engine.Log.Warn("rpc request tx history record not found: %s", hex.EncodeToString(txid))
				//	continue
				//}
				//data = append(data, convertRewardHistoryItemToTxInfo(customTx))
				//continue
			}
			var blockHashStr string
			if blockHash != nil {
				blockHashStr = hex.EncodeToString(*blockHash)
			}
			//txinfo转map，然后重写某些字段
			item := rpc.JsonMethod(txItr)
			tx, _, _ := mining.FindTx(txid)
			txClass, vins, vouts := rpc.DealTxInfo(tx, addr.(string), blockHeight)
			if txClass > 0 {
				item["type"] = txClass
				if vins != nil {
					item["vin"] = vins
				}
				if vouts != nil {
					item["vout"] = vouts
				}
			}
			outMap := make(map[string]interface{})
			outMap["txinfo"] = item
			outMap["upchaincode"] = code
			outMap["blockheight"] = blockHeight
			outMap["blockhash"] = blockHashStr
			outMap["timestamp"] = timestamp
			outMap["iscustomtx"] = false
			//outMap["balance_info"] = "" //余额信息
			data = append(data, outMap)
		}

	}
	data1 := map[string]interface{}{}
	data1["count"] = index
	data1["data"] = data
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETADDRESSTX_REV, pkg(model.Success, data1))
}

func convertRewardHistoryItemToTxInfo(hi *mining.HistoryItem) map[string]interface{} {
	outMap := make(map[string]interface{})

	vin := []*mining.VinVO{}
	for _, v := range hi.InAddr {
		vin = append(vin, &mining.VinVO{
			Addr: v.B58String(),
		})
	}

	vout := []*mining.VoutVO{}
	for _, v := range hi.OutAddr {
		vout = append(vout, &mining.VoutVO{
			Address: v.B58String(),
			Value:   hi.Value,
		})
	}

	txHash := hex.EncodeToString(hi.Txid)
	txInfo := mining.TxBaseVO{
		Hash:      txHash,                           //本交易hash，不参与区块hash，只用来保存
		Type:      hi.Type,                          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易                //输入交易数量
		Vin:       vin,                              //交易输入                //输出交易数量
		Vout:      vout,                             //交易输出
		Payload:   string(hi.Payload),               //备注信息
		BlockHash: hex.EncodeToString(hi.BlockHash), //本交易属于的区块hash，不参与区块hash，只用来保存
		Reward:    hi.Value,
	}
	item := rpc.JsonMethod(txInfo)
	item["action"] = "call"
	outMap["txinfo"] = item
	outMap["upchaincode"] = 2
	outMap["blockheight"] = hi.Height
	outMap["blockhash"] = hex.EncodeToString(hi.BlockHash)
	outMap["timestamp"] = hi.Timestamp
	outMap["iscustomtx"] = true

	return outMap
}

/*
获取所有交易
*/
func GetAllTx(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETALLTX_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))

	start := (pageInt - 1) * pageSizeInt

	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETALLTX_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	his := chain.GetHistoryBalanceDesc(uint64(start), pageSizeInt)
	data := make([]map[string]interface{}, 0)
	//balanceHistory := mining.NewBalanceHistory()
	for _, h := range his {
		if h.Txid != nil {
			txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(h.Txid)
			if txItr == nil {
				//customTx, err := balanceHistory.GetCustomTx(h.Txid)
				//if err != nil {
				//	continue
				//}
				//data = append(data, convertRewardHistoryItemToTxInfo(customTx))
				//continue
			}
			var blockHashStr string
			if blockHash != nil {
				blockHashStr = hex.EncodeToString(*blockHash)
			}
			item := rpc.JsonMethod(txItr)
			tx, _, _ := mining.FindTx(h.Txid)
			vouts := tx.GetVout()
			outs := *vouts
			var addr string
			if bytes.Equal(outs[0].Address, precompiled.RewardContract) || bytes.Equal(outs[0].Address, ens.GetRegisterAddr()) {
				txVin := *tx.GetVin()
				addr = txVin[0].GetPukToAddr().B58String()
				txClass, vins, vout := rpc.DealTxInfo(tx, addr, blockHeight)
				if txClass > 0 {
					item["type"] = txClass
					if vins != nil {
						item["vin"] = vins
					}
					if vout != nil {
						item["vout"] = vout
					}
				}
			}
			outMap := make(map[string]interface{})
			outMap["txinfo"] = item
			outMap["upchaincode"] = code
			outMap["blockheight"] = blockHeight
			outMap["blockhash"] = blockHashStr
			outMap["timestamp"] = timestamp
			outMap["iscustomtx"] = false
			data = append(data, outMap)
		}

	}

	data1 := map[string]interface{}{}

	data1["count"] = chain.GetHistoryBalanceCount() //去除创世块部署奖励合约那笔交易
	data1["data"] = data

	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETALLTX_REV, pkg(model.Success, data1))
	return
}

/*
获取挖矿交易
*/
func GetMinerBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMINERBLOCK_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMINERBLOCK_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMINERBLOCK_REV, pkg(model.TypeWrong, "address"))
		return
	}

	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))

	start := (pageInt - 1) * pageSizeInt
	//addrkey := []byte(config.Miner_history_tx + "_" + strings.ToLower(addr.(string)))
	addrkey := append(config.Miner_history_tx, []byte(strings.ToLower(addr.(string)))...)

	index := rpc.GetMinerBlockIndex(addr.(string))

	data := make([]map[string]interface{}, 0)
	for i := 0; i < pageSizeInt; i++ {
		indexBs := make([]byte, 8)
		binary.LittleEndian.PutUint64(indexBs, index-uint64(start)-uint64(i))
		blockHash, err := db.LevelTempDB.HGet(addrkey, indexBs)
		if err != nil {
			continue
		}
		if len(blockHash) != 0 {
			head, err := mining.LoadBlockHeadVOByHash(&blockHash)
			if err != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMINERBLOCK_REV, pkg(rpc.SystemError, err.Error()))
				return
			}
			gas := uint64(0)
			for _, txItr := range head.Txs {
				gas += txItr.GetGas()
			}
			reward := gas
			outMap := make(map[string]interface{})
			outMap["blockhash"] = hex.EncodeToString(blockHash)
			outMap["blockheight"] = head.BH.Height
			outMap["timestamp"] = head.BH.Time
			outMap["tx_count"] = head.BH.NTx
			outMap["previous_hash"] = hex.EncodeToString(head.BH.Previousblockhash)
			outMap["block_reward"] = reward
			outMap["destroy"] = 0
			data = append(data, outMap)
		}

	}
	data1 := map[string]interface{}{}
	data1["count"] = index
	data1["data"] = data
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETMINERBLOCK_REV, pkg(model.Success, data1))
}

/*
获取轻节点详情
*/
func GetLightNodeDetail(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	//如果不是轻节点
	list := precompiled.GetLightList([]crypto.AddressCoin{addrCoin})
	if len(list) != 1 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(rpc.NotLightNode, "address"))
		return
	}

	lightDetail := list[0]

	cAddr := ""
	cName := ""
	cVote := new(big.Int)
	reward := precompiled.GetLightNodeReward(addrCoin)
	if lightDetail.C != evmutils.ZeroAddress {
		c := evmutils.AddressToAddressCoin(lightDetail.C.Bytes())
		cAddr = c.B58String()
		cs := precompiled.GetCommunityList([]crypto.AddressCoin{c})
		if len(cs) == 1 {
			cName = cs[0].Name
			cVote = cs[0].Vote
		}
	}

	outMap := make(map[string]interface{})
	outMap["light_addr"] = addr
	outMap["community_addr"] = cAddr
	outMap["community_name"] = cName
	outMap["contract"] = precompiled.RewardContract.B58String()
	outMap["deposit"] = new(big.Int).Sub(lightDetail.Score, lightDetail.Vote).Uint64()
	outMap["vote"] = lightDetail.Vote.Uint64()
	outMap["start_height"] = lightDetail.BlockHeight.Uint64()
	outMap["reward"] = reward
	//outMap["frozen_reward"] = precompiled.GetMyLightFrozenReward(addrCoin)

	var frozenReward uint64
	if cAddr != "" {
		cAddress := crypto.AddressFromB58String(cAddr)
		cRewardPool := precompiled.GetCommunityRewardPool(cAddress)
		if cRewardPool.Cmp(big.NewInt(0)) > 0 {
			cRate, err := precompiled.GetRewardRatio(cAddress)
			if err != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(rpc.SystemError, "address"))
				return
			}
			lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
			v := new(big.Int).Mul(new(big.Int).SetUint64(mining.GetLongChain().GetBalance().GetDepositVote(&addrCoin).Value), new(big.Int).SetInt64(1e8))
			ratio := new(big.Int).Quo(v, cVote)
			lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
			frozenReward = lBigValue.Uint64()
		}
	}

	outMap["frozen_reward"] = frozenReward
	outMap["name"] = lightDetail.Name
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETLIGHTNODEDETAIL_REV, pkg(model.Success, outMap))
}

/*
获取见证者节点详情
*/
func GetWitnessNodeDetail(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(model.TypeWrong, "address"))
		return
	}
	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	// 获得质押量
	depositAmount := mining.GetDepositWitnessAddr(&addrCoin)
	if depositAmount <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(rpc.NotWitnessNode, "address"))
		return
	}

	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))
	start := (pageInt - 1) * pageSizeInt

	addBlockCount, addBlockReward := rpc.GetAddressAddBlockReward(addr.(string))

	detail, ok := precompiled.GetWitnessDetail(addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(rpc.SystemError, "address"))
		return
	}
	witnessName := mining.FindWitnessName(addrCoin)
	data := rpc.WitnessInfoV0{
		Deposit:        depositAmount,
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		RewardRatio:    float64(detail.Rate),
		TotalReward:    detail.Reward.Uint64(),
		FrozenReward:   detail.RemainReward.Uint64(),
		Name:           witnessName,
		CommunityNode:  []rpc.CommunityNode{},
		DestroyNum:     0,
	}
	// 总票数
	voteTotal := uint64(0)
	data.CommunityCount = uint64(len(detail.Communitys))
	i := 0
	cs := make([]rpc.CommunityNode, 0)
	sort.Sort(precompiled.CommunityRewardSort(detail.Communitys))
	// 获取社区节点详情
	for _, c := range detail.Communitys {
		voteTotal += c.Vote.Uint64()
		if i >= start && len(cs) < pageSizeInt {
			coin := evmutils.AddressToAddressCoin(c.Addr[:])
			cs = append(cs, rpc.CommunityNode{
				Name:        c.Name,
				Addr:        coin.B58String(),
				Deposit:     c.Score.Uint64(),
				Reward:      c.Reward.Uint64(),
				LightNum:    c.LightCount.Uint64(),
				VoteNum:     c.Vote.Uint64(),
				RewardRatio: float64(c.Rate),
			})
		}
		i++
	}
	data.CommunityNode = cs
	data.Vote = voteTotal
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSNODEDETAIL_REV, pkg(model.Success, data))
}

/*
获取社区节点详情
*/
func GetCommunityNodeDetail(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))
	start := (pageInt - 1) * pageSizeInt

	communityList := precompiled.GetCommunityList([]crypto.AddressCoin{addrCoin})
	if len(communityList) == 1 && communityList[0].Wit == evmutils.ZeroAddress {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(rpc.NotCommunityNode, "address"))
		return
	}

	if len(communityList) != 1 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(rpc.SystemError, "address"))
		return
	}
	community := communityList[0]

	//reward := precompiled.GetNodeReward(addrCoin)
	reward := precompiled.GetCommunityNodeReward(addrCoin)
	ratio, err := precompiled.GetRewardRatio(addrCoin)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(rpc.SystemError, "address"))
		return
	}
	light := precompiled.GetLightListByC(addrCoin)
	witnessAddr := evmutils.AddressToAddressCoin(community.Wit[:])
	name := mining.FindWitnessName(witnessAddr)
	//frozenReward := precompiled.GetMyFrozenReward(addrCoin)
	//frozenReward := precompiled.GetMyCommunityFrozenReward(addrCoin)
	var frozenReward uint64
	cRewardPool := precompiled.GetCommunityRewardPool(addrCoin)
	if cRewardPool.Cmp(big.NewInt(0)) > 0 {
		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(ratio))), new(big.Int).SetInt64(100))
		cBigValue := new(big.Int).Sub(cRewardPool, lTotal)
		frozenReward = cBigValue.Uint64()
	}
	data := rpc.CommunityInfoV0{
		Deposit:      community.Score.Uint64(),
		Vote:         community.Vote.Uint64(),
		LightCount:   uint64(len(light)),
		RewardRatio:  float64(ratio),
		Reward:       reward,
		WitnessName:  name,
		WitnessAddr:  witnessAddr.B58String(),
		StartHeight:  community.BlockHeight.Uint64(),
		FrozenReward: frozenReward,
		Contract:     precompiled.RewardContract.B58String(),
		Name:         community.Name,
		LightNode:    []rpc.LightNode{},
	}
	i := 0

	totalVote := uint64(0)
	for _, l := range light {
		totalVote += l.Vote.Uint64()
	}

	sort.Slice(light, func(i, j int) bool {
		if light[i].Vote.Cmp(light[j].Vote) == 1 {
			return true
		}
		return false
	})
	//MMSH7zVgj3JY8WaPCXxM1bQgWAWwgyu8VQ1V4
	for _, l := range light {
		if i >= start && len(data.LightNode) < pageSizeInt {
			lightAddr := evmutils.AddressToAddressCoin(l.Addr[:])
			reward := precompiled.GetNodeReward(lightAddr)
			ratio := uint64(0)
			if totalVote != 0 {
				ratio = l.Vote.Uint64() * 100 * 1e3 / totalVote
			}

			ln := rpc.LightNode{
				Addr:        lightAddr.B58String(),
				Reward:      reward,
				RewardRatio: float64(ratio),
				VoteNum:     l.Vote.Uint64(),
				Name:        l.Name,
				Deposit:     new(big.Int).Sub(l.Score, l.Vote).Uint64(),
				//LastVotedHeight: mining.GetLightLastVotedHeight(lightAddr),
			}
			data.LightNode = append(data.LightNode, ln)
		}
		i++
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYNODEDETAIL_REV, pkg(model.Success, data))
}

/*
获取社区节点列表
*/
func GetCommunityListForMiner(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYLISTFORMINER_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10)
	}
	pageSizeInt := int(pageSize.(float64))

	start := (pageInt - 1) * pageSizeInt

	community := mining.GetCommunityRewardListSortNew()
	sort.Slice(community, func(i, j int) bool {
		if community[i].Vote > community[j].Vote {
			return true
		}
		return false
	})

	count := len(community)

	wvos := make([]rpc.CommunityList, 0)
	i := 0
	for _, one := range community {
		if i >= start && len(wvos) < pageSizeInt {
			//addBlockCount, _ := rpc.GetAddressAddBlockReward(one.Addr)

			wvo := rpc.CommunityList{
				Addr:        one.Addr, //社区节点地址
				WitnessAddr: one.Witness,
				Payload:     one.Payload, //名字
				Score:       one.Score,   //质押量
				Vote:        one.Vote,    // 总票数
				//AddBlockCount:  addBlockCount,
				//AddBlockReward: one.Reward,
				DisRatio:    float64(one.Rate),
				RewardRatio: float64(one.Rate),
			}
			wvos = append(wvos, wvo)
		}
		i++

	}
	data1 := map[string]interface{}{}
	data1["count"] = count
	data1["data"] = wvos
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYLISTFORMINER_REV, pkg(model.Success, data1))
}

/*
获得候选见证人和见证人列表 带翻页
*/
func GetWitnessesListWithPage(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	withPage, err := rpc.GetWitnessesListWithPage(rj, nil, nil)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(withPage, &m)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE_REV, pkg(model.Success, m["result"]))
}

/*
获得某个候选见证人和见证人信息
*/
func GetWitnessesInfoForMiner(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER_REV, pkg(model.NoField, "address"))
		return
	}
	addrStr := addrItr.(string)

	addr := crypto.AddressFromB58String(addrStr)
	score := uint64(0)
	witness := mining.GetLongChain().WitnessChain.FindWitnessByAddr(addr)
	if witness == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER_REV, pkg(model.Nomarl, "not found witness"))
		return
	}
	score = witness.Score

	addBlockCount, addBlockReward := rpc.GetAddressAddBlockReward(addrStr)

	rates, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(config.Area.Keystore.GetCoinbase().Addr,
		[]common.Address{common.Address(evmutils.AddressCoinToAddress(addr))})
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER_REV, pkg(rpc.SystemError, "address"))
		return
	}

	wvo := rpc.WitnessList{
		Addr:           addrStr,                      //见证人地址
		Payload:        mining.FindWitnessName(addr), //名字
		Score:          score,                        //质押量
		Vote:           (votes[0]).Uint64(),          // 总票数
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		Ratio:          float64(rates[0]),
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESINFOFORMINER_REV, pkg(model.Success, wvo))
}

/*
获得见证人列表 带翻页
*/
func GetWitnessesListV0WithPage(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTV0WITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	page, err := rpc.GetWitnessesListV0WithPage(rj, nil, nil)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTV0WITHPAGE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(page, &m)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTWITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLISTV0WITHPAGE_REV, pkg(model.Success, m["result"]))
}

/*
获得候选见证人列表 带翻页
*/
func GetWitnessesBackUpListV0WithPage(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESBACKUPLISTV0WITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	page, err := rpc.GetWitnessesBackUpListV0WithPage(rj, nil, nil)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESBACKUPLISTV0WITHPAGE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(page, &m)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESBACKUPLISTV0WITHPAGE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESBACKUPLISTV0WITHPAGE_REV, pkg(model.Success, m["result"]))
}

/*
获取总的质押量
*/
func GetDepositNumForAll(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	lightTotal := precompiled.GetLightTotal(nil)
	depositAmount := lightTotal * config.Mining_light_min
	community := precompiled.GetAllCommunity(nil)
	for _, c := range community {
		depositAmount += c.Score.Uint64()
	}

	wbg := mining.GetBackupWitnessList()
	for _, w := range wbg {
		depositAmount += w.Score
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETDEPOSITNUMFORALL_REV, pkg(model.Success, depositAmount))
}

/*
获取批量交易input
*/
func GetAbiInput(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	t := "function"
	method, ok := rj.Get("method")
	if !ok {
		method = ""
		t = "constructor"
	}

	t1 := rpc.T{
		Constant:        true,
		Name:            method.(string),
		Inputs:          []rpc.Input{},
		Outputs:         []rpc.Input{},
		Payable:         false,
		StateMutability: "view",
		Type:            t,
	}

	fieldStr, ok := rj.Get("field_type")
	if !ok {
		fieldStr = ""
	}
	if !rj.VerifyType("field_type", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(model.TypeWrong, "address"))
		return
	}
	fields := make([]string, 0)
	if len(fieldStr.(string)) != 0 {
		fields = strings.Split(fieldStr.(string), "-")
		for _, field := range fields {
			t1.Inputs = append(t1.Inputs, rpc.Input{
				Name: "",
				Type: field,
			})
		}
	}
	initData := []rpc.T{t1}

	initBody, _ := json.Marshal(initData)

	ab, err := abi.JSON(bytes.NewReader(initBody))
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, "abi json err"))
		return
	}

	paramStr, ok := rj.Get("params")
	if !ok {
		paramStr = ""
	}
	if !rj.VerifyType("params", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(model.TypeWrong, "params"))
		return
	}
	params := strings.Split(paramStr.(string), "-")
	args := make([]interface{}, 0)

	if len(paramStr.(string)) != 0 {
		if len(params) != len(fields) {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, "field_type length must same as params"))
			return
		}

		for k, param := range params {
			switch fields[k] {
			case "address":
				args = append(args, common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(param))))
			case "string":
				args = append(args, param)
			case "bytes":
				bytesParam, err := hex.DecodeString(param)
				if err != nil {
					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(model.Nomarl, err.Error()))
					return
				}
				args = append(args, bytesParam)
			case "uint256":
				intParam, err := decimal.NewFromString(param)
				if err != nil {
					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(model.Nomarl, err.Error()))
					return
				}
				args = append(args, intParam.BigInt())
			case "uint8":
				p, err := strconv.Atoi(param)
				if err != nil {
					_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, err.Error()))
					return
				}
				args = append(args, uint8(p))
			case "address[]":
				address := make([]common.Address, 0)
				adds := strings.Split(param, ",")
				for _, add := range adds {
					address = append(address, common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(add))))
				}
				args = append(args, address)
			case "string[]":
				strs := strings.Split(param, ",")

				args = append(args, strs)
			case "uint256[]":
				ints := make([]*big.Int, 0)
				intsS := strings.Split(param, ",")
				for _, i := range intsS {
					i1, err := strconv.Atoi(i)
					if err != nil {
						_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, err.Error()))
						return
					}
					ints = append(ints, big.NewInt(int64(i1)))
				}
				args = append(args, ints)
			case "bool":
				if param == "true" {
					args = append(args, true)
				} else {
					args = append(args, false)
				}
			default:
				engine.Log.Error("invalid type", fields[k])
			}
		}
	}

	packB, err := ab.Pack(method.(string), args...)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(rpc.SystemError, "abi Pack err"))
		return
	}

	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETABIINPUT_REV, pkg(model.Success, packB))
}

/*
根据交易类型获取交易gas 默认普通交易
*/
func GetTxGas(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTXGAS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	txType, ok := rj.Get("type")
	if !ok {
		txType = float64(config.Wallet_tx_type_pay)
	}

	txTypeInt := uint64(txType.(float64))

	txGasSum := uint64(0)
	txCount := uint64(0)

	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTXGAS_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	hs := chain.GetHistoryBalanceDesc(0, 500)
	for _, h := range hs {
		if len(h.Txid) != 0 {
			txItr, _, _ := mining.FindTxJsonV1(h.Txid)
			if txItr == nil {
				continue
			}

			if txItr.Class() == txTypeInt {
				txCount++
				txGasSum += txItr.GetGas()
			}
			if txCount >= 5 {
				break
			}
		}
	}

	normal := uint64(0)
	if txCount != 0 {
		normal = txGasSum / txCount
	}

	if normal < config.Wallet_tx_gas_min {
		normal = config.Wallet_tx_gas_min
	}

	low := normal / 125 * 100
	if low < config.Wallet_tx_gas_min {
		low = config.Wallet_tx_gas_min
	}

	data := make(map[string]interface{})
	data["low"] = low
	data["normal"] = normal
	data["fast"] = normal * 125 / 100
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTXGAS_REV, pkg(model.Success, data))
}

/*
根据区块高度和交易hash获取奖励合约事件
*/
func GetContractEvent(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	list := &go_protos.ContractEventInfoList{}
	height, ok := rj.Get("height")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.NoField, "height"))
		return
	}

	if !rj.VerifyType("height", "float64") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.TypeWrong, "height"))
		return
	}

	hash, ok := rj.Get("hash")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.NoField, "hash"))
		return
	}

	if !rj.VerifyType("hash", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.TypeWrong, "hash"))
		return
	}

	hByte, err := hex.DecodeString(hash.(string))
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.TypeWrong, "hash"))
		return
	}
	intHeight := evmutils.New(int64(height.(float64)))
	key := append([]byte(config.DBKEY_BLOCK_EVENT), intHeight.Bytes()...)
	body, err := db.LevelDB.GetDB().HGet(key, hByte)
	if err != nil {
		engine.Log.Error("获取保存合约事件日志失败%s", err.Error())
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.Success, list))
		return
	}
	if len(body) == 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.Success, list))
		return
	}

	err = proto.Unmarshal(body, list)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.Success, list))
		return
	}

	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTEVENT_REV, pkg(model.Success, list))
}

// 收藏币种
func AddErc20(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	tokenP, ok := rj.Get("tokens")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(model.NoField, "tokens"))
		return
	}

	bs, err := json.Marshal(tokenP)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(model.TypeWrong, "tokens"))
		return
	}
	tokens := make([]rpc.Token, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&tokens)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(model.TypeWrong, "tokens"))
		return
	}

	if len(tokens) <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(model.NoField, "tokens"))
		return
	}

	for _, v := range tokens {
		//保存到key即可
		key := []byte(config.DBKEY_ERC20_COIN)
		db.LevelDB.HSet(key, []byte(v.Address), precompiled.EncodeToken([]byte(v.Name), []byte(v.Symbol)))
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_ADDERC20_REV, pkg(model.Success, "success"))
}

// 移除代币
func DelErc20(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrP, ok := rj.Get("addresses")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(model.NoField, "addresses"))
		return
	}

	bs, err := json.Marshal(addrP)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(model.TypeWrong, "addresses"))
		return
	}
	address := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&address)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(model.TypeWrong, "addresses"))
		return
	}

	if len(address) <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(model.NoField, "addresses"))
		return
	}

	//保存到key即可
	key := []byte(config.DBKEY_ERC20_COIN)
	for _, v := range address {
		db.LevelDB.HDel(key, []byte(v))
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_DELERC20_REV, pkg(model.Success, "success"))
}

// 获取代币
func GetErc20(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	key, ok := rj.Get("key")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20_REV, pkg(model.NoField, "key"))
		return
	}
	k := []byte(key.(string))
	type Coin struct {
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
		Address string `json:"address"`
	}
	list, err := db.LevelDB.HGetAll(k)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20_REV, pkg(model.Nomarl, err.Error()))
		return
	}

	coinList := []Coin{}
	for _, v := range list {
		n, s, err := precompiled.DecodeToken(v.Value)
		if err != nil {
			continue
		}
		coinList = append(coinList, Coin{Name: string(n), Symbol: string(s), Address: string(v.Field)})
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20_REV, pkg(model.Success, coinList))
}

// 获取代币
func GetErc20SumBalance(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrP, ok := rj.Get("contractaddresses")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(model.NoField, "contractaddresses"))
		return
	}

	bs, err := json.Marshal(addrP)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(model.TypeWrong, "contractaddresses"))
		return
	}
	address := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&address)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(model.TypeWrong, "contractaddresses"))
		return
	}

	if len(address) <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(model.NoField, "contractaddresses"))
		return
	}

	addressMap := make(map[string]struct{}, len(address))
	for _, v := range address {
		addressMap[v] = struct{}{}
	}

	addres, _ := rj.Get("addr")
	sli := addres.([]interface{})
	addrList := make([]*keystore.AddressInfo, len(sli))
	for i, i2 := range sli {
		v := MapChangeAddressInfo(i2)
		addrList[i] = &v
	}
	coinList := []rpc.Erc20Info{}
	for _, v := range address {
		contractAddr := v
		value := precompiled.GetMultiAddrBalance(crypto.AddressFromB58String(contractAddr), addrList)

		info := db.GetErc20Info(contractAddr)
		decimal := info.Decimals

		coinList = append(coinList, rpc.Erc20Info{
			Address:  info.Address,
			Name:     info.Name,
			Symbol:   info.Symbol,
			Decimals: info.Decimals,
			Value:    precompiled.ValueToString(value, decimal),
		})
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20VALUE_REV, pkg(model.Success, coinList))
}

// 搜索代币
func SearchErc20(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SEARCHERC20_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	keywordI, ok := rj.Get("keyword")

	if !ok || keywordI.(string) == "" {
		erc20list := db.GetAllErc20Info()
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SEARCHERC20_REV, pkg(model.Success, erc20list))
		return
	}

	keyword := keywordI.(string)

	//只能包含字母或数字
	if matched, _ := regexp.MatchString("^(?i)[a-z0-9]+$", keyword); !matched {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SEARCHERC20_REV, pkg(model.TypeWrong, "keyword"))
		return
	}

	list := make([]db.Erc20Info, 0)
	//判断地址前缀是否正确
	if len(base58.Decode(keyword[len(keyword)-1:])) > 0 && crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(keyword)) {
		erc := db.GetErc20Info(keyword)
		list = append(list, erc)
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SEARCHERC20_REV, pkg(model.Success, list))
		return
	}
	erc20list := db.GetAllErc20Info()
	reg := regexp.MustCompile(strings.ToLower(keyword))
	for _, v := range erc20list {
		ind := reg.FindIndex([]byte(strings.ToLower(v.Name)))
		if ind != nil {
			v.Sort = ind[0]
			list = append(list, v)
		}
	}
	sort.Sort(db.Erc20Sort(list))
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SEARCHERC20_REV, pkg(model.Success, list))
}

// 获取多代币地址的总余额
//func GetErc20SumBalance(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	rj := new(model.RpcJson)
//	err := json.Unmarshal(*message.Body.Content, &rj)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(rpc.SystemError, err.Error()))
//		return
//	}
//	contractaddressesI, ok := rj.Get("contractaddresses")
//	if !ok {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(model.NoField, "contractaddresses"))
//		return
//	}
//
//	bs, err := json.Marshal(contractaddressesI)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(model.TypeWrong, "contractaddresses"))
//		return
//	}
//	contractaddresses := make([]string, 0)
//	decoder := json.NewDecoder(bytes.NewBuffer(bs))
//	decoder.UseNumber()
//	err = decoder.Decode(&contractaddresses)
//	if err != nil {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(model.TypeWrong, "contractaddresses"))
//		return
//	}
//
//	if len(contractaddresses) <= 0 {
//		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(model.NoField, "contractaddresses"))
//		return
//	}
//
//	addres, _ := rj.Get("addr")
//	sli := addres.([]interface{})
//	addrList := make([]*keystore.AddressInfo, len(sli))
//	for i, i2 := range sli {
//		v := MapChangeAddressInfo(i2)
//		addrList[i] = &v
//	}
//
//	list := make([]rpc.Erc20Info, len(contractaddresses))
//	for k, c := range contractaddresses {
//		value := precompiled.GetMultiAddrBalance(crypto.AddressFromB58String(c), addrList)
//
//		info := db.GetErc20Info(c)
//		decimal := info.Decimals
//
//		list[k] = rpc.Erc20Info{
//			Address:  info.Address,
//			Name:     info.Name,
//			Symbol:   info.Symbol,
//			Decimals: info.Decimals,
//			Value:    precompiled.ValueToString(value, decimal),
//		}
//	}
//	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETERC20SUMBALANCE_REV, pkg(model.Success, list))
//}

// 代币转账
func TransferErc20(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_TRANSFERERC20_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	coinBaseAddr, _ := rj.Get("coinBaseAddr")
	addr, _ := rj.Get("addr")
	amountItr, _ := rj.Get("amountItr")
	toItr, _ := rj.Get("toItr")

	decimal := precompiled.GetDecimals(coinBaseAddr.(string), addr.(string))
	amount := precompiled.StringToValue(amountItr.(string), decimal)
	if amount == nil || amount.Cmp(big.NewInt(0)) == 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_TRANSFERERC20_REV, pkg(rpc.AmountIsZero, "amount"))
		return
	}
	payload := precompiled.BuildErc20TransferBigInput(toItr.(string), amount)
	comment := common.Bytes2Hex(payload)

	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_TRANSFERERC20_REV, pkg(model.Success, comment))
}

// 获取代币余额
func GetTokenBalance(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(model.NoField, "contractaddress"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
		return
	}
	balance := precompiled.GetBigBalance(src, src.B58String(), addr)
	decimal := precompiled.GetDecimals(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCE_REV, pkg(model.Success, precompiled.ValueToString(balance, decimal)))

}

// 获取代币余额
func GetTokenBalances(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addressesP, ok := rj.Get("addresses")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.NoField, "addresses"))
		return
	}

	bs, err := json.Marshal(addressesP)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.TypeWrong, "addresses"))
		return
	}
	addresses := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	err = decoder.Decode(&addresses)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.TypeWrong, "addresses"))
		return
	}

	addrs := make([]crypto.AddressCoin, len(addresses))
	for k, v := range addresses {
		addr := crypto.AddressFromB58String(v)
		addrs[k] = addr
	}

	addrItr, ok := rj.Get("contractaddress")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.NoField, "contractaddress"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
		return
	}

	tokenbalances := make([]rpc.TokenBalance, len(addresses))
	for k, v := range addrs {
		balance := precompiled.GetBigBalance(v, v.B58String(), addr)
		tokenbalances[k] = rpc.TokenBalance{v, balance}
	}

	sort.Sort(rpc.TokenBalanceSort(tokenbalances))

	type RTokenBalance struct {
		Address string `json:"address"`
		Value   string `json:"value"`
	}
	coinBaseAddr, ok := rj.Get("coinBaseAddr")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.NoField, "coinBaseAddr"))
		return
	}
	decimal := precompiled.GetDecimals(coinBaseAddr.(string), addr)
	result := make([]RTokenBalance, len(addresses))
	for k, v := range tokenbalances {
		result[k] = RTokenBalance{v.Address.B58String(), precompiled.ValueToString(v.Value, decimal)}
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETTOKENBALANCES_REV, pkg(model.Success, result))
}

// 获取代币余额
func GetRewardContract(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDCONTRACT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDCONTRACT_REV, pkg(model.Success, precompiled.RewardContract.B58String()))
}
