package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"math/big"
	"math/rand"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/cross"
	"web3_gui/chain/evm/precompiled/faucet"
	"web3_gui/chain/mining"
	"web3_gui/chain/sqlite3_db"
	utils2 "web3_gui/chain/utils"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

type HistoryItemVO struct {
	GenerateId string   //
	IsIn       bool     //资金转入转出方向，true=转入;false=转出;
	Type       uint64   //交易类型
	InAddr     []string //输入地址
	OutAddr    []string //输出地址
	Value      uint64   //交易金额
	Txid       string   //交易id
	Height     uint64   //区块高度
	Payload    string   //
}

/*
	历史记录
*/

func GetTransactionHistoty(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.GetLongChain() == nil {
		res, err = model.Tojson(model.Success)
		return
	}
	//如果是见证人，需要时间间隔控制
	if ok, _, _, _, _ := mining.GetWitnessStatus(); ok {
		utils.SetTimeToken(config.TIMETOKEN_GetTransactionHistoty, time.Second*5)
		if allow := utils.GetTimeToken(config.TIMETOKEN_GetTransactionHistoty, false); !allow {
			res, err = model.Tojson(model.Success)
			return
		}
	}

	id := ""
	idItr, ok := rj.Get("id")
	if ok {
		if !rj.VerifyType("id", "string") {
			res, err = model.Errcode(model.TypeWrong, "id")
			return
		}
		id = idItr.(string)
	}
	var startId *big.Int
	if id != "" {
		var ok bool
		startId, ok = new(big.Int).SetString(id, 10)
		if !ok {
			res, err = model.Errcode(model.TypeWrong, "id")
			return
		}
	}

	total := 0
	totalItr, ok := rj.Get("total")
	if ok {
		total = int(totalItr.(float64))
		// fmt.Println("total", total)
	}
	hivos := make([]HistoryItemVO, 0)

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(hivos)
		return
	}
	his := chain.GetHistoryBalance(startId, total)

	for _, one := range his {
		hivo := HistoryItemVO{
			GenerateId: one.GenerateId.String(),      //
			IsIn:       one.IsIn,                     //资金转入转出方向，true=转入;false=转出;
			Type:       one.Type,                     //交易类型
			InAddr:     make([]string, 0),            //输入地址
			OutAddr:    make([]string, 0),            //输出地址
			Value:      one.Value,                    //交易金额
			Txid:       hex.EncodeToString(one.Txid), //交易id
			Height:     one.Height,                   //区块高度
			Payload:    string(one.Payload),          //
		}

		for _, two := range one.InAddr {
			hivo.InAddr = append(hivo.InAddr, two.B58String())
		}

		for _, two := range one.OutAddr {
			hivo.OutAddr = append(hivo.OutAddr, two.B58String())
		}

		hivos = append(hivos, hivo)
	}

	res, err = model.Tojson(hivos)
	return
}

/*
	查询数据库中key对应的value
*/
// func MergeTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	switchItr, ok := rj.Get("switch")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "switch")
// 		return
// 	}
// 	isOpen := switchItr.(bool)

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "gas")
// 		return
// 	}
// 	gas := toUint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	if pwd != config.Wallet_keystore_default_pwd {
// 		res, err = model.Tojson(model.FailPwd)
// 		return
// 	}

// 	var unifieaddr *crypto.AddressCoin
// 	unifieaddrItr, ok := rj.Get("unifieaddr") //归集地址
// 	if ok {
// 		addrStr := unifieaddrItr.(string)
// 		if addrStr != "" {
// 			addrMul := crypto.AddressFromB58String(addrStr)
// 			if !crypto.ValidAddr(config.AddrPre, addrMul) {
// 				res, err = model.Errcode(ContentIncorrectFormat, "unifieaddr")
// 				// engine.Log.Info("11111111111111111111")
// 				return
// 			}
// 			unifieaddr = &addrMul
// 		}
// 	}

// 	totalMax := toUint64(0)
// 	totalMaxItr, ok := rj.Get("totalmax")
// 	if ok {
// 		totalMax = toUint64(totalMaxItr.(float64))
// 	}

// 	if isOpen {
// 		mining.SwitchOpenMergeTx(pwd, gas, unifieaddr, totalMax)
// 	} else {
// 		mining.SwitchCloseMergeTx(pwd)
// 	}

// 	res, err = model.Tojson(model.Success)
// 	return
// }

/*
查询节点总数
*/
//func GetNodeTotal(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	outMap := make(map[string]interface{})
//	//获取存储超级节点地址
//	nameinfo := name.FindName(config.Name_store)
//	if nameinfo == nil {
//		//域名不存在
//		engine.Log.Debug("Domain name does not exist")
//		outMap["total_addr"] = 0  //全网节点总数
//		outMap["total_space"] = 0 //全网存储空间总数
//		res, err = model.Tojson(outMap)
//		return
//	}
//	// nets := client.GetCloudPeer()
//	// if nets == nil {
//	// 	return
//	// }
//	//判断自己是否在超级节点地址里
//	have := false
//	for _, one := range nameinfo.NetIds {
//		if bytes.Equal(config.Area.GetNetId(), one) {
//			have = true
//			break
//		}
//	}
//	//没有在列表里
//	if !have {
//		engine.Log.Debug("You are not in the super node address")
//		outMap["total_addr"] = config.GetSpaceTotalAddr() //全网节点总数
//		outMap["total_space"] = config.GetSpaceTotal()    //全网存储空间总数
//		res, err = model.Tojson(outMap)
//		return
//	}
//	// peers := server.CountStorePeers()
//	// total := len(*peers)
//	outMap["total_addr"], outMap["total_space"] = server.CountStoreTotal()
//	// = config.GetSpaceTotalAddr() //全网节点总数
//	// outMap["total_space"] = config.GetSpaceTotal()    //全网存储空间总数
//	res, err = model.Tojson(outMap)
//	return
//}

/*
查询地址的nonce
*/
func GetNonce(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //地址
	if !ok {
		// engine.Log.Info("11111111111111111111")
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}
	if addr == nil {
		res, err = model.Errcode(model.NoField, "address")
		return
	}

	nonceInt, e := mining.GetAddrNonce(addr)
	if e != nil {
		res, err = model.Errcode(SystemError, e.Error())
		return
	}
	out := make(map[string]interface{})
	out["nonce"] = nonceInt.Uint64()
	res, err = model.Tojson(out)
	return
}

/*
查询地址的缓存Nonce
*/
func getPendingNonce(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //地址
	if !ok {
		// engine.Log.Info("11111111111111111111")
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}
	if addr == nil {
		res, err = model.Errcode(model.NoField, "address")
		return
	}

	nonceInt, e := mining.GetAddrNonce(addr)
	if e != nil {
		res, err = model.Errcode(SystemError, e.Error())
		return
	}
	cacheNonceInt := mining.GetLongChain().GetTransactionManager().GetUnpackedTransaction().FindAddrNonce(addr)
	out := make(map[string]interface{})
	if nonceInt.Cmp(&cacheNonceInt) >= 0 {
		out["nonce"] = nonceInt.Uint64()
	} else {
		out["nonce"] = cacheNonceInt.Uint64()
	}
	res, err = model.Tojson(out)
	return
}

/*
查询多个地址的nonce
*/
func GetNonceMore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address_list") //地址
	if !ok {
		// engine.Log.Info("11111111111111111111")
		res, err = model.Errcode(model.NoField, "address_list")
		return
	}
	addrStrs, ok := addrItr.([]interface{})

	if !ok || addrStrs == nil || len(addrStrs) == 0 {
		res, err = model.Errcode(model.NoField, "address_list")
		return
	}
	addrNonceMap := mining.GetAddrNonceMore(addrStrs)
	res, err = model.Tojson(addrNonceMap)
	return
}

/*
见证人信息
*/
type WitnessInfo struct {
	IsCandidate bool   //是否是候选见证人
	IsBackup    bool   //是否是备用见证人
	IsKickOut   bool   //没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
	Addr        string //见证人地址
	Payload     string //
	Value       uint64 //见证人押金
}

/*
查询自己是否是见证人
*/
func GetWitnessInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	winfo := WitnessInfo{}

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(winfo)
		return
	}
	var witnessAddr crypto.AddressCoin
	winfo.IsCandidate, winfo.IsBackup, winfo.IsKickOut, witnessAddr, winfo.Value = mining.GetWitnessStatus()
	winfo.Addr = witnessAddr.B58String()

	addr := config.Area.Keystore.GetCoinbase()
	winfo.Payload = mining.FindWitnessName(addr.Addr)

	res, err = model.Tojson(winfo)
	return
}

/*
见证人
*/
type WitnessVO struct {
	Addr            string //见证人地址
	Payload         string //
	Score           uint64 //押金
	Vote            uint64 //投票者押金
	CreateBlockTime int64  //预计出块时间
}

/*
见证人列表
*/
type WitnessList struct {
	Addr           string  `json:"addr"`             //见证人地址
	Payload        string  `json:"payload"`          //
	Score          uint64  `json:"score"`            //押金
	Vote           uint64  `json:"vote"`             //投票者押金
	AddBlockCount  uint64  `json:"add_block_count"`  // 出块数量
	AddBlockReward uint64  `json:"add_block_reward"` // 出块奖励
	Ratio          float64 `json:"ratio"`            // 奖励比例
}

type WitnessListSort []WitnessList

func (e WitnessListSort) Len() int {
	return len(e)
}

func (e WitnessListSort) Less(i, j int) bool {
	return e[i].Vote > e[j].Vote
}

func (e WitnessListSort) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

/*
社区节点列表
*/
type CommunityList struct {
	Addr           string  `json:"addr"`             //社区节点地址
	WitnessAddr    string  `json:"witness_addr"`     //见证人地址
	Payload        string  `json:"payload"`          //
	Score          uint64  `json:"score"`            //押金
	Vote           uint64  `json:"vote"`             //投票者押金
	AddBlockCount  uint64  `json:"add_block_count"`  // 出块数量
	AddBlockReward uint64  `json:"add_block_reward"` // 出块奖励
	RewardRatio    float64 `json:"reward_ratio"`     // 奖励比例
	DisRatio       float64 `json:"dis_ratio"`        // 分配比例
}

/*
获取候选见证人列表
*/
func GetCandidateList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	chain := mining.GetLongChain()
	if chain == nil {
		return model.Errcode(SystemError, "get chain failed")
	}

	wits := chain.WitnessBackup.GetAllWitness()

	wvos := []WitnessList{}
	// total := len(wits)

	for _, v := range wits {
		addBlockCount, addBlockReward := GetAddressAddBlockReward(v.Addr.B58String())
		wvo := WitnessList{
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
	sort.Sort(WitnessListSort(wvos))

	// currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()
	// wbg := currWitness.WitnessBigGroup

	// allWiness := append(wbg.Witnesses, wbg.WitnessBackup...)
	// addrs := make([]common.Address, len(allWiness))
	// for k, one := range allWiness {
	// 	addrs[k] = common.Address(evmutils.AddressCoinToAddress(*one.Addr))
	// }

	// from := config.Area.Keystore.GetCoinbase().Addr
	// _, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(from, addrs)
	// if err != nil {
	// 	return model.Errcode(SystemError, "address")
	// }

	// wvos := make([]WitnessVO, 0)
	// for k, one := range allWiness {
	// 	wvo := WitnessVO{
	// 		Addr:            one.Addr.B58String(),              //见证人地址
	// 		Payload:         mining.FindWitnessName(*one.Addr), //
	// 		Score:           one.Score,                         //押金
	// 		Vote:            votes[k].Uint64(),                 //      voteValue,            //投票票数
	// 		CreateBlockTime: one.CreateBlockTime,               //预计出块时间
	// 	}
	// 	wvos = append(wvos, wvo)

	// }
	return model.Tojson(wvos)
}

/*
获取候选见证人列表
*/
func GetWitnessesList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	wbg := mining.GetWitnessListSort()

	balanceMgr := mining.GetLongChain().GetBalance()

	wvos := make([]WitnessVO, 0)
	for _, one := range wbg.Witnesses {

		name := mining.FindWitnessName(*one.Addr)
		// engine.Log.Info("查询到的见证人名称:%s", name)
		//vote := precompiled.GetWitnessVote(*one.Addr)
		vote := balanceMgr.GetWitnessVote(one.Addr)

		wvo := WitnessVO{
			Addr:            one.Addr.B58String(), //见证人地址
			Payload:         name,                 //
			Score:           one.Score,            //押金
			Vote:            vote,                 //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,  //预计出块时间
		}
		wvos = append(wvos, wvo)
	}

	res, err = model.Tojson(wvos)
	return
}

/*
获得全网候选见证人和见证人列表
*/
func GetWitnessesListWithPage(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}
	start := (pageInt - 1) * pageSizeInt

	//currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()
	//
	//wbg := currWitness.WitnessBigGroup
	//witnessOrigin := append(wbg.Witnesses, wbg.WitnessBackup...)

	//wvos, total, err := witnessesListWithRange(witnessOrigin, start, start+pageSizeInt)

	witnessOrigin := mining.GetLongChain().WitnessBackup.GetAllWitness()
	wvos, total, err := witnessesListWithRangeV1(witnessOrigin, start, start+pageSizeInt)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "list witness failed")
		return
	}

	data := map[string]interface{}{}
	data["count"] = total
	data["data"] = wvos

	return model.Tojson(data)
}

/*
获取出块见证人列表
*/
func GetWitnessesListV0WithPage(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()

	wbg := currWitness.WitnessBigGroup
	witnessOrigin := append(wbg.Witnesses)

	wvos, total, err := witnessesListWithRange(witnessOrigin, 0, config.Witness_backup_max)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "list witness failed")
		return
	}

	data := map[string]interface{}{}
	data["count"] = total
	data["data"] = wvos

	return model.Tojson(data)
}

/*
获取候选见证人列表
*/
func GetWitnessesBackUpListV0WithPage(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}
	start := (pageInt - 1) * pageSizeInt

	currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()

	wbg := currWitness.WitnessBigGroup
	witnessOrigin := append(wbg.WitnessBackup)

	wvos, total, err := witnessesListWithRange(witnessOrigin, start, start+pageSizeInt)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "list witness failed")
		return
	}

	data := map[string]interface{}{}
	data["count"] = total
	data["data"] = wvos

	return model.Tojson(data)
}

/*
获得全网/出块/候选 见证人
*/
func witnessesListWithRange(wits []*mining.Witness, start, end int) (res []WitnessList, total int, err error) {
	// 因为共识原因，见证人取消质押，仍然在见证人组中
	// 这里通过见证人押金判断一下
	witness := []*mining.Witness{}
	//vmAddrs := make([]common.Address, 0)
	for _, one := range wits {
		if mining.GetDepositWitnessAddr(one.Addr) > 0 {
			witness = append(witness, one)
			//vmAddrs = append(vmAddrs, common.Address(evmutils.AddressCoinToAddress(*one.Addr)))
		}
	}

	// 填充数据

	balanceMgr := mining.GetLongChain().GetBalance()
	wvos := []WitnessList{}
	//rates, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(config.Area.Keystore.GetCoinbase().Addr, vmAddrs)
	//if err != nil {
	//	return nil, 0, err
	//}

	total = len(witness)

	for _, one := range witness {
		addBlockCount, addBlockReward := GetAddressAddBlockReward(one.Addr.B58String())
		wvo := WitnessList{
			Addr:    one.Addr.B58String(),              //见证人地址
			Payload: mining.FindWitnessName(*one.Addr), //名字
			Score:   one.Score,                         //质押量
			//Vote:           votes[i].Uint64(),                 // 总票数
			Vote:           balanceMgr.GetWitnessVote(one.Addr), // 总票数
			AddBlockCount:  addBlockCount,
			AddBlockReward: addBlockReward,
			//Ratio:          float64(rates[i]),
			Ratio: float64(balanceMgr.GetDepositRate(one.Addr)),
		}
		wvos = append(wvos, wvo)
	}

	// 按投票数排序
	sort.Sort(WitnessListSort(wvos))

	// 分页
	outs := []WitnessList{}
	for i, one := range wvos {
		if i >= start && i < end {
			outs = append(outs, one)
		}
	}

	return outs, total, nil
}

/*
获取候选见证人信息
*/
func GetWitnessesInfoForMiner(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)

	addr := crypto.AddressFromB58String(addrStr)
	score := toUint64(0)
	witness := mining.GetLongChain().WitnessChain.FindWitnessByAddr(addr)
	if witness == nil {
		return model.Errcode(model.Nomarl, "not found witness")
	}
	score = witness.Score

	addBlockCount, addBlockReward := GetAddressAddBlockReward(addrStr)

	//rates, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(config.Area.Keystore.GetCoinbase().Addr,
	//	[]common.Address{common.Address(evmutils.AddressCoinToAddress(addr))})
	//if err != nil {
	//	return model.Errcode(SystemError, "address")
	//}
	balanceMgr := mining.GetLongChain().GetBalance()
	voteNum := balanceMgr.GetWitnessVote(&addr)
	ratio := balanceMgr.GetDepositRate(&addr)

	wvo := WitnessList{
		Addr:           addrStr,                      //见证人地址
		Payload:        mining.FindWitnessName(addr), //名字
		Score:          score,                        //质押量
		Vote:           voteNum,                      // 总票数
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		Ratio:          float64(ratio),
	}
	res, err = model.Tojson(wvo)
	return
}

/*
获取社区节点列表
*/
func GetCommunityList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vss := mining.GetCommunityListSort()
	res, err = model.Tojson(vss)
	return
}

/*
获取社区节点列表
*/
func GetCommunityListForMiner(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	//community := mining.GetCommunityRewardListSortNew()
	//sort.Slice(community, func(i, j int) bool {
	//	if community[i].Vote > community[j].Vote {
	//		return true
	//	}
	//	return false
	//})
	balanceMgr := mining.GetLongChain().GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()

	communityAddrs := []crypto.AddressCoin{}
	wmc.Range(func(key, value any) bool {
		communityAddrs = append(communityAddrs, value.([]crypto.AddressCoin)...)
		return true
	})

	comms := make([]CommunityList, 0)
	for _, commAddr := range communityAddrs {
		if comminfo := balanceMgr.GetDepositCommunity(&commAddr); comminfo != nil {
			ratio := balanceMgr.GetDepositRate(&comminfo.SelfAddr)
			comms = append(comms, CommunityList{
				Addr:        comminfo.SelfAddr.B58String(),
				WitnessAddr: comminfo.WitnessAddr.B58String(),
				Payload:     comminfo.Name,
				Score:       comminfo.Value,
				Vote:        balanceMgr.GetCommunityVote(&commAddr),
				RewardRatio: float64(ratio),
				DisRatio:    float64(ratio),
			})
		}
	}

	sort.Slice(comms, func(i, j int) bool {
		if comms[i].Vote > comms[j].Vote {
			return true
		}
		return false
	})

	count := len(comms)

	//data := make([]CommunityList, 0)
	//for i, _ := range comms {
	//	if i >= start && len(comms) < pageSizeInt {
	//		data = append(data, comms[i])
	//	}
	//}

	data := map[string]interface{}{}
	data["count"] = count
	start, end, ok := helppager(count, pageInt, pageSizeInt)
	if ok {
		out := make([]CommunityList, end-start)
		copy(out, comms[start:end])
		data["data"] = out
	} else {
		data["data"] = []interface{}{}
	}

	res, err = model.Tojson(data)
	return
}

/*
	获取轻节点列表
*/
//func GetCommunityDetail(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	page, ok := rj.Get("page")
//	if !ok {
//		page = float64(1)
//	}
//	pageInt := int(page.(float64))
//
//	pageSize, ok := rj.Get("page_size")
//	if !ok {
//		pageSize = float64(10)
//	}
//	pageSizeInt := int(pageSize.(float64))
//
//	start := (pageInt - 1) * pageSizeInt
//
//	var addr *crypto.AddressCoin
//	addrItr, ok := rj.Get("address") //押金冻结的地址
//	if !ok {
//		// engine.Log.Info("11111111111111111111")
//		res, err = model.Errcode(model.NoField, "address")
//		return
//	}
//	addrStr := addrItr.(string)
//	if addrStr != "" {
//		addrMul := crypto.AddressFromB58String(addrStr)
//		addr = &addrMul
//	}
//
//	if addrStr != "" {
//		dst := crypto.AddressFromB58String(addrStr)
//		if !crypto.ValidAddr(config.AddrPre, dst) {
//			res, err = model.Errcode(ContentIncorrectFormat, "address")
//			// engine.Log.Info("11111111111111111111")
//			return
//		}
//	}
//
//	if mining.GetAddrState(*addr) != 2 {
//		//本地址不是社区节点
//		res, err = model.Errcode(RuleField, "address")
//		return
//	}
//
//	chain := mining.GetLongChain()
//	if chain == nil {
//		res, err = model.Errcode(model.Nomarl, "The chain end is not synchronized")
//		return
//	}
//	currentHeight := chain.GetCurrentBlock()
//
//	//查询未分配的奖励
//	ns, notSend, err := mining.FindNotSendReward(addr)
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//		return
//	}
//	if ns != nil && notSend != nil && len(*notSend) > 0 {
//		//查询代上链的奖励是否已经上链
//		*notSend, _ = checkTxUpchain(*notSend, currentHeight)
//		// engine.Log.Info("查询到的奖励:%d", ns.Reward)
//		//有未分配完的奖励
//		community := ns.Reward / 10
//		light := ns.Reward - community
//		lightNum := ns.LightNum
//		if lightNum > 0 {
//			lightNum--
//		}
//		rewardTotal := mining.RewardTotal{
//			CommunityReward: community,                           //社区节点奖励
//			LightReward:     light,                               //轻节点奖励
//			StartHeight:     ns.StartHeight,                      //
//			Height:          ns.EndHeight,                        //最新区块高度
//			IsGrant:         false,                               //是否可以分发奖励，24小时候才可以分发奖励
//			AllLight:        lightNum,                            //所有轻节点数量
//			RewardLight:     ns.LightNum - toUint64(len(*notSend)), //已经奖励的轻节点数量
//			IsNew:           false,
//		}
//		// engine.Log.Info("11111111111111111111")
//		// engine.Log.Info("此次未分配完的奖励:%d %d Reward:%d light:%d community:%d allReward:%d", ns.StartHeight, ns.EndHeight, ns.Reward, light, community, light+community)
//		res, err = model.Tojson(rewardTotal)
//		return
//	}
//	if ns == nil {
//		//需要加载以前的奖励快照
//		findCommunityStartHeightByAddr(*addr)
//		res, err = model.Errcode(model.Nomarl, "load reward history")
//		// engine.Log.Info("11111111111111111111")
//		return
//	}
//	// engine.Log.Info("11111111111111111111")
//	startHeight := ns.EndHeight + 1
//
//	// engine.Log.Info("333333333333333333")
//	//奖励都分配完了，查询新的奖励
//	rt, _, err := mining.GetRewardCount(addr, startHeight, 0)
//	if err != nil {
//		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
//			res, err = model.Errcode(RewardCountSync, err.Error())
//			return
//		}
//		// engine.Log.Info("444444444444444")
//		res, err = model.Errcode(model.Nomarl, err.Error())
//		return
//	}
//	rt.IsNew = true
//
//	lightNum := rt.AllLight
//	if lightNum > 0 {
//		lightNum--
//	}
//	rewardTotal := mining.RewardTotal{
//		CommunityReward: rt.CommunityReward, //社区节点奖励
//		LightReward:     rt.LightReward,     //轻节点奖励
//		StartHeight:     rt.StartHeight,     //
//		Height:          rt.Height,          //最新区块高度
//		IsGrant:         rt.IsGrant,         //是否可以分发奖励，24小时候才可以分发奖励
//		AllLight:        lightNum,           //所有轻节点数量
//		RewardLight:     rt.RewardLight,     //已经奖励的轻节点数量
//		IsNew:           rt.IsNew,
//	}
//	res, err = model.Tojson(rewardTotal)
//	return
//}

/*
投票详情
*/
type VoteInfoVO struct {
	Txid        string //交易id
	WitnessAddr string //见证人地址
	Value       uint64 //投票数量
	Height      uint64 //区块高度
	AddrSelf    string //自己投票的地址
	Payload     string //
}

type Vinfos struct {
	infos []VoteInfoVO
}

func (this *Vinfos) GetInfos() []VoteInfoVO {
	return this.infos
}

func (this *Vinfos) Len() int {
	return len(this.infos)
}

func (this *Vinfos) Less(i, j int) bool {
	if this.infos[i].Height < this.infos[j].Height {
		return false
	} else {
		return true
	}
}

func (this *Vinfos) Swap(i, j int) {
	this.infos[i], this.infos[j] = this.infos[j], this.infos[i]
}

func NewVinfos(infos []VoteInfoVO) Vinfos {
	return Vinfos{
		infos: infos,
	}
}

/*
获得自己给哪些见证人投过票的列表
@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func GetVoteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	var items []*mining.DepositInfo
	switch voteType {
	case mining.VOTE_TYPE_community:
		// engine.Log.Info("11111111111")
		items = mining.GetDepositCommunityList()
	case mining.VOTE_TYPE_light:
		// engine.Log.Info("11111111111")
		items = mining.GetDepositLightList()
	case mining.VOTE_TYPE_vote:
		// engine.Log.Info("11111111111")
		items = mining.GetDepositVoteList()
	}
	vinfos := Vinfos{
		infos: make([]VoteInfoVO, 0, len(items)),
	}
	// engine.Log.Info("查询押金结果:%d", len(items))
	for _, item := range items {
		name := mining.FindWitnessName(item.WitnessAddr)
		viVO := VoteInfoVO{
			// TokenId:        hex.EncodeToString(ti.TokenId), //
			WitnessAddr: item.WitnessAddr.B58String(), //见证人地址
			Value:       item.Value,                   //投票数量
			// Height:      item.Height,           //区块高度
			AddrSelf: item.SelfAddr.B58String(), //自己投票的地址
			Payload:  name,                      //
		}
		vinfos.infos = append(vinfos.infos, viVO)
	}

	// balances := mining.GetVoteList()

	// for _, one := range balances {
	// 	// fmt.Println("查看", one)
	// 	one.Txs.Range(func(k, v interface{}) bool {
	// 		ti := v.(*mining.TxItem)
	// 		if ti.VoteType != voteType {
	// 			return true
	// 		}
	// 		viVO := VoteInfoVO{
	// 			TokenId:        hex.EncodeToString(ti.TokenId), //
	// 			WitnessAddr: one.Addr.B58String(),        //见证人地址
	// 			Value:       ti.Value,                    //投票数量
	// 			Height:      ti.Height,                   //区块高度
	// 			AddrSelf:    ti.Addr.B58String(),         //自己投票的地址
	// 		}
	// 		viVO.Payload = mining.FindWitnessName(*ti.Addr)

	// 		vinfos.infos = append(vinfos.infos, viVO)
	// 		return true
	// 	})
	// }

	// fmt.Println("排序前查看投票", vinfos)
	// sort.Sort(&vinfos)
	sort.Stable(&vinfos)
	// fmt.Println("排序后查看投票", vinfos)
	res, err = model.Tojson(vinfos.infos)
	return
}

// /*
// 	获得自己给哪些社区节点投过票的列表
// */
// func GetCommunityVoteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	balances := mining.GetVoteList()
// 	vinfos := Vinfos{
// 		infos: make([]VoteInfoVO, 0),
// 	}
// 	for _, one := range balances {
// 		// fmt.Println("查看", one)
// 		one.Txs.Range(func(k, v interface{}) bool {
// 			ti := v.(*mining.TxItem)
// 			viVO := VoteInfoVO{
// 				TokenId:        hex.EncodeToString(ti.TokenId), //
// 				WitnessAddr: one.Addr.B58String(),        //见证人地址
// 				Value:       ti.Value,                    //投票数量
// 				Height:      ti.Height,                   //区块高度
// 				AddrSelf:    "",                          //自己投票的地址
// 			}
// 			vinfos.infos = append(vinfos.infos, viVO)
// 			return true
// 		})
// 	}

// 	// fmt.Println("排序前查看投票", vinfos)
// 	sort.Sort(&vinfos)
// 	// fmt.Println("排序后查看投票", vinfos)
// 	res, err = model.Tojson(vinfos.infos)
// 	return
// }

/*
查询一笔交易是否成功
@return    int    1=未确认；2=成功；3=失败；
*/
func FindTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	txItr, ok := rj.Get("txid")
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txidStr := txItr.(string)
	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(txid)
	var blockHashStr string
	if blockHash != nil {
		blockHashStr = hex.EncodeToString(*blockHash)
	}

	tx, err := mining.LoadTxBase(txid)
	if err != nil {
		res, err = model.Errcode(model.NotExist, "txid")
		return
	}
	item := JsonMethod(txItr)
	item["bloom"] = hex.EncodeToString(tx.GetBloom())
	out := make(map[string]interface{})
	out["txinfo"] = item
	out["upchaincode"] = code
	out["blockheight"] = blockHeight
	out["blockhash"] = blockHashStr
	out["timestamp"] = timestamp

	res, err = model.Tojson(out)
	return
}

/*
查询批量查询交易是否成功
@return    int    1=未确认；2=成功；3=失败；
*/
func FindTxs(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	txidsItr, ok := rj.Get("txids")
	if !ok {
		res, err = model.Errcode(model.NoField, "txids")
		return
	}

	bs, err := json.Marshal(txidsItr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txids")
		return
	}
	txids := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&txids)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txids")
		return
	}

	if len(txids) <= 0 {
		res, err = model.Errcode(model.NoField, "txids")
		return
	}

	outMap := sync.Map{}
	wg := sync.WaitGroup{}
	var oneerr error
	for _, txidStr := range txids {
		wg.Add(1)
		go func(txidStr string) {
			defer wg.Done()
			txid, err := hex.DecodeString(txidStr)
			if err != nil {
				if oneerr == nil {
					oneerr = errors.Errorf("交易ID格式错误:%s", txidStr)
				}
				return
			}
			txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(txid)
			var blockHashStr string
			if blockHash != nil {
				blockHashStr = hex.EncodeToString(*blockHash)
			}
			tx, err := mining.LoadTxBase(txid)
			if err != nil {
				if oneerr == nil {
					oneerr = errors.Errorf("交易不存在:%s", txidStr)
				}
				return
			}
			item := JsonMethod(txItr)
			item["bloom"] = hex.EncodeToString(tx.GetBloom())
			out := make(map[string]interface{})
			out["txinfo"] = item
			out["upchaincode"] = code
			out["blockheight"] = blockHeight
			out["blockhash"] = blockHashStr
			out["timestamp"] = timestamp
			outMap.Store(txidStr, out)
		}(txidStr)
	}
	wg.Wait()
	if oneerr != nil {
		res, err = model.Errcode(model.Nomarl, oneerr.Error())
		return
	}

	items := []interface{}{}
	for _, txidStr := range txids {
		if item, ok := outMap.Load(txidStr); ok {
			items = append(items, item)
		}
	}
	res, err = model.Tojson(items)
	return
}

/*
查询数据库中key对应的value
*/
func GetValueForKey(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	keyItr, ok := rj.Get("key")
	if !ok {
		res, err = model.Errcode(model.NoField, "key")
		return
	}
	keyBs, e := hex.DecodeString(keyItr.(string))
	if e != nil {
		res, err = model.Errcode(model.TypeWrong, "key")
		return
	}
	value, e := db.LevelDB.Find(keyBs)
	if e != nil {
		res, err = model.Errcode(SystemError, e.Error())
		return
	}
	res, err = model.Tojson(value)
	return
}

type BlockHeadVO struct {
	Hash              string   //区块头hash
	Height            uint64   //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64   //矿工组高度
	GroupHeightGrowth uint64   //组高度增长量。默认0为自动计算增长量（兼容之前的区块）,最少增量为1
	Previousblockhash string   //上一个区块头hash
	Nextblockhash     string   //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
	NTx               uint64   //交易数量
	MerkleRoot        string   //交易默克尔树根hash
	Tx                []string //本区块包含的交易id
	Time              int64    //出块时间，unix时间戳
	Witness           string   //此块见证人地址
	Sign              string   //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	ExtSign           []string //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	Reward            uint64   //出块奖励
	Gas               uint64   //燃料费
	Destroy           uint64   //销毁
	Bloom             string   //bloom过滤器
}

/*
通过区块高度查询一个区块详细信息
*/
func FindBlock(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	heightItr, ok := rj.Get("height")
	var height uint64
	if !ok || heightItr.(float64) == -1 {
		height = mining.GetLongChain().GetCurrentBlock()
	} else {
		height = toUint64(heightItr.(float64))
	}
	bh := mining.LoadBlockHeadByHeight(height)
	// bh := mining.FindBlockHead(height)
	if bh == nil {
		res, err = model.Errcode(model.NotExist)
		return
	}

	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	//reward := toUint64(config.BLOCK_TOTAL_REWARD)
	gas := toUint64(0)
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
		}
	}

	reward += gas

	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	//查询区块的bloom
	bloom := ""
	if bloomBs, err := mining.GetBlockHeadBloom(bh.Height); err == nil {
		bloom = hex.EncodeToString(bloomBs)
	}

	bhvo := BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		//Gas:               reward - toUint64(config.BLOCK_TOTAL_REWARD),
		Gas:     gas,
		Destroy: 0,
		Bloom:   bloom,
	}
	res, err = model.Tojson(bhvo)
	return

}

/*
通过区块最新高度
*/
func GetBlockHeight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	heightItr, ok := rj.Get("height")
	var height uint64
	if ok {
		switch heightItr.(float64) {
		case -1:
			height = mining.GetLongChain().GetCurrentBlock()
		case -2:
			height = mining.GetLongChain().HighestBlock
		case -3:
			height = mining.GetLongChain().PulledStates
		}
	} else {
		height = mining.GetLongChain().GetCurrentBlock()
	}
	res, err = model.Tojson(height)
	return
}

/*
通过区块Hash查询一个区块详细信息
*/
func FindBlockByHash(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hashItr, ok := rj.Get("hash")
	if !ok {
		res, err = model.Errcode(model.NoField, "hash")
		return
	}

	hash, err := hex.DecodeString(hashItr.(string))
	if err != nil {
		res, err = model.Errcode(ContentIncorrectFormat)
		return
	}
	bh, err := mining.LoadBlockHeadByHash(&hash)
	if bh == nil {
		res, err = model.Errcode(model.NotExist)
		return
	}

	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	//reward := toUint64(config.BLOCK_TOTAL_REWARD)
	gas := toUint64(0)
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
		}
	}

	reward += gas

	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	//查询区块的bloom
	bloom := ""
	if bloomBs, err := mining.GetBlockHeadBloom(bh.Height); err == nil {
		bloom = hex.EncodeToString(bloomBs)
	}

	bhvo := BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		//Gas:               reward - toUint64(config.BLOCK_TOTAL_REWARD),
		Gas:     gas,
		Destroy: 0,
		Bloom:   bloom,
	}

	res, err = model.Tojson(bhvo)
	return

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
			engine.Log.Error("this addr not community:%s error:%s", addr.B58String(), err.Error())
			return
		}
		var sn *sqlite3_db.SnapshotReward
		//判断数据库是否有快照记录
		sn, _, err = mining.FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			engine.Log.Error("querying database Error %s", err.Error())
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
				engine.Log.Error("findCommunityStartHeightByAddr load blockhead error:%s", err.Error())
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
				engine.Log.Info(err.Error())
			}
		}
		// engine.Log.Info("findCommunityStartHeightByAddr Community finish count block!")
	}, nil)

}

/*
查询社区奖励
*/
func GetCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	engine.Log.Info("11111111111111111111")
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if !ok {
		// engine.Log.Info("11111111111111111111")
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}

	if addrStr != "" {
		dst := crypto.AddressFromB58String(addrStr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(ContentIncorrectFormat, "address")
			// engine.Log.Info("11111111111111111111")
			return
		}
	}
	// engine.Log.Info("11111111111111111111")
	// engine.Log.Info("地址 "+addrStr+" d%", mining.GetAddrState(*addr))

	if mining.GetAddrState(*addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(RuleField, "address")
		// engine.Log.Info("11111111111111111111")
		return
	}

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Errcode(ChainNotSync, "The chain end is not synchronized")
		// engine.Log.Info("11111111111111111111")
		return
	}
	currentHeight := chain.GetCurrentBlock()

	// engine.Log.Info("11111111111111111111")
	// engine.Log.Info("11111111111111111111")
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(addr)
	if err != nil {
		// engine.Log.Info("222222222222222222222")
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("11111111111111111111")
		return
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
		// engine.Log.Info("此次未分配完的奖励:%d %d Reward:%d light:%d community:%d allReward:%d", ns.StartHeight, ns.EndHeight, ns.Reward, light, community, light+community)
		res, err = model.Tojson(rewardTotal)
		return
	}
	if ns == nil {
		//需要加载以前的奖励快照
		FindCommunityStartHeightByAddr(*addr)
		res, err = model.Errcode(model.Nomarl, "load reward history")
		// engine.Log.Info("11111111111111111111")
		return
	}
	// engine.Log.Info("11111111111111111111")
	startHeight := ns.EndHeight + 1

	// engine.Log.Info("333333333333333333")
	//奖励都分配完了，查询新的奖励
	rt, _, err := mining.GetRewardCount(addr, startHeight, 0)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			res, err = model.Errcode(RewardCountSync, err.Error())
			return
		}
		// engine.Log.Info("444444444444444")
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
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
	res, err = model.Tojson(rewardTotal)
	// engine.Log.Info("11111111111111111111")
	return
}

/*
给轻节点分发奖励
旧版，废弃
*/
func SendCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// engine.Log.Info("22222222222222")
	// var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //社区节点地址
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	addrStr := addrItr.(string)
	if addrStr == "" {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	addr := crypto.AddressFromB58String(addrStr)
	if !crypto.ValidAddr(config.AddrPre, addr) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		// engine.Log.Info("22222222222222")
		return
	}

	if mining.GetAddrState(addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(RuleField, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	//查询社区节点公钥
	puk, ok := config.Area.Keystore.GetPukByAddr(addr)
	if !ok {
		res, err = model.Errcode(PubKeyNotExists, config.ERROR_public_key_not_exist.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		// engine.Log.Info("22222222222222")
		return
	}
	gas := toUint64(gasItr.(float64))
	// engine.Log.Info("22222222222222")
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		// engine.Log.Info("22222222222222")
		return
	}
	pwd := pwdItr.(string)

	startHeightItr, ok := rj.Get("startheight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startheight")
		// engine.Log.Info("22222222222222")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endheightItr, ok := rj.Get("endheight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endheight")
		// engine.Log.Info("22222222222222")
		return
	}
	endheight := toUint64(endheightItr.(float64))
	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Errcode(ChainNotSync, "The chain end is not synchronized")
		// engine.Log.Info("22222222222222")
		return
	}
	currentHeight := chain.GetCurrentBlock()
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(&addr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询待上链的奖励是否已经上链
		*notSend, ok = CheckTxUpchain(*notSend, currentHeight)
		if ok {
			//有未上链的奖励
			res, err = model.Errcode(RewardNotLinked, "There are rewards that are not linked")
			// engine.Log.Info("22222222222222")
			return
		}
		cs := mining.NewCommunitySign(puk, ns.StartHeight, ns.EndHeight)
		//有未分配完的奖励，继续分配
		err = mining.DistributionReward(&addr, notSend, gas, pwd, cs, ns.StartHeight, ns.EndHeight, currentHeight)
		if err != nil {
			if err.Error() == config.ERROR_password_fail.Error() {
				// engine.Log.Info("创建转账交易错误 222222222222")
				res, err = model.Errcode(model.FailPwd)
				return
			}
			if err.Error() == config.ERROR_not_enough.Error() {
				res, err = model.Errcode(NotEnough)
				return
			}
			res, err = model.Errcode(model.Nomarl, err.Error())
			// engine.Log.Info("22222222222222")
			return
		}
		// engine.Log.Info("此次分配奖励:%d", ns.StartHeight, ns.EndHeight, ns.LightNum)
		res, err = model.Tojson(model.Success)
		// engine.Log.Info("22222222222222")
		return
	}
	//检查奖励开始高度，避免重复奖励
	if ns != nil && startHeight <= ns.EndHeight {
		res, err = model.Errcode(RepeatReward, "Repeat reward")
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	//奖励都分配完了，查询新的奖励
	rt, notSend, err := mining.GetRewardCount(&addr, startHeight, endheight)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			res, err = model.Errcode(RewardCountSync, err.Error())
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	if !rt.IsGrant {
		//now := config.TimeNow().Unix()
		//pl time
		now := config.TimeNow().UnixNano()
		blockNum := uint64(config.Mining_community_reward_time/config.Mining_block_time) - (rt.Height - rt.StartHeight)
		wait := blockNum * uint64(config.Mining_block_time.Nanoseconds())
		futuer := time.Unix(0, now+int64(wait))
		// engine.Log.Info("22222222222222")
		res, err = model.Errcode(DistributeRewardTooEarly, "Please distribute the reward after "+futuer.Format("2006-01-02 15:04:05"))
		return
	}
	//创建快照
	err = mining.CreateRewardCount(addr, rt, *notSend)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	//创建快照成功后，删除缓存
	mining.CleanRewardCountProcessMap(&addr)

	// engine.Log.Info("22222222222222")
	// cs := mining.NewCommunitySign(puk, rt.StartHeight, rt.Height)
	// err = mining.DistributionReward(notSend, gas, pwd, cs, currentHeight)
	// if err != nil {
	// 	res, err = model.Errcode(model.Nomarl, err.Error())
	// 	return
	// }
	res, err = model.Tojson(model.Success)
	// engine.Log.Info("22222222222222")
	return
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
			engine.Log.Error(err.Error())
		}
	}
	//奖励修改为已经上链
	if len(resultUpchain) > 0 {
		var err error
		for _, one := range resultUpchain {
			err = one.UpdateDistribution(one.Id, one.Reward)
			if err != nil {
				engine.Log.Error(err.Error())
			}
		}
	}
	return resultReward, haveNotUpchain
}

/*
	检查是否有交易已经生成，但是未上链的奖励
*/
// func checkHaveNotUpchain(notSend []sqlite3_db.Reward)bool{

// }

/*
将组装好并签名的交易上链
*/
func PushTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	txJsonItr, ok := rj.Get("tx")
	if !ok {
		res, err = model.Errcode(model.NoField, "tx")
		return
	}

	txjson := txJsonItr.(string)
	if txjson == "" {
		res, err = model.Errcode(model.NoField, "empty tx")
		return
	}

	txjsonBs, e := base64.StdEncoding.DecodeString(txjson)
	if e != nil {
		engine.Log.Info("DecodeString fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	// mining.ParseTxBaseProto()

	// engine.Log.Info("txjson:%s", string(txjsonBs))
	txItr, e := mining.ParseTxBaseProto(0, &txjsonBs)
	// txItr, err := mining.ParseTxBase(0, &txjsonBs)
	if e != nil {
		engine.Log.Info("ParseTxBaseProto fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	//验证锁定高度
	if e = txItr.CheckLockHeight(mining.GetHighestBlock()); e != nil {
		return model.Errcode(model.Nomarl, e.Error())
	}

	// 支付类检查是否免gas
	if err := mining.CheckTxPayFreeGas(txItr); err != nil {
		return model.Errcode(GasTooLittle, "gas")
	}

	//检查Nonce
	tm := mining.GetLongChain().GetTransactionManager()
	_, err = tm.CheckNonce(txItr)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	spend := txItr.GetSpend()
	vins := *txItr.GetVin()
	if len(vins) == 0 {
		res, err = model.Errcode(model.NoField, "empty vin")
		return
	}
	src := vins[0].GetPukToAddr()

	if txItr.Class() == config.Wallet_tx_type_address_transfer {
		tx := txItr.(*mining.TxAddressTransfer)
		gas := txItr.GetGas()
		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(src, gas)
		if total < gas {
			//资金不够
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
		vouts := *txItr.GetVout()
		amount := vouts[0].Value
		payTotal, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&tx.PayAddress, amount)
		if amount == 0 {
			amount = payTotal
		}
		if payTotal < amount {
			//资金不够
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
	} else {
		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(src, spend)
		if total < spend {
			//资金不够
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
	}

	// engine.Log.Info("rpc transaction received %s", hex.EncodeToString(*txItr.GetHash()))
	if e := txItr.CheckSign(); e != nil {
		engine.Log.Info("transaction check fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	//验证domain
	if !txItr.CheckDomain() {
		return model.Errcode(model.Nomarl, "domain name resolution failed")
	}

	e = mining.AddTx(txItr)
	if e != nil {
		engine.Log.Info("AddTx fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	res, err = model.Tojson(model.Success)
	return
}

/*
合约预执行
*/
func PreContractTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	txJsonItr, ok := rj.Get("tx")
	if !ok {
		res, err = model.Errcode(model.NoField, "tx")
		return
	}

	txjson := txJsonItr.(string)

	txjsonBs, e := base64.StdEncoding.DecodeString(txjson)
	if e != nil {
		engine.Log.Info("DecodeString fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	txItr, e := mining.ParseTxBaseProto(0, &txjsonBs)
	if e != nil {
		engine.Log.Info("ParseTxBaseProto fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	if txItr.GetGas() < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}

	tx := txItr.(*mining.Tx_Contract)

	if len(tx.Payload) < 4 {
		return []byte{}, errors.New(e.Error() + "调用合约参数错误")
	}

	fee := tx.GetSpend()
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(tx.Vin[0].GetPukToAddr(), fee)
	if total < fee {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	//验证domain
	if !txItr.CheckDomain() {
		return model.Errcode(model.Nomarl, "domain name resolution failed")
	}

	//验证合约
	gasUsed, err := tx.PreExecNew()
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(gasUsed)
	return
} /*

push
*/
func PushContractTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	txJsonItr, ok := rj.Get("tx")
	if !ok {
		res, err = model.Errcode(model.NoField, "tx")
		return
	}

	txjson := txJsonItr.(string)

	txjsonBs, e := base64.StdEncoding.DecodeString(txjson)
	if e != nil {
		engine.Log.Info("DecodeString fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	// mining.ParseTxBaseProto()

	// engine.Log.Info("txjson:%s", string(txjsonBs))
	txItr, e := mining.ParseTxBaseProto(0, &txjsonBs)
	// txItr, err := mining.ParseTxBase(0, &txjsonBs)
	if e != nil {
		engine.Log.Info("ParseTxBaseProto fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

	//验证锁定高度
	if e = txItr.CheckLockHeight(mining.GetHighestBlock()); e != nil {
		return model.Errcode(model.Nomarl, e.Error())
	}

	if txItr.GetGas() < config.Wallet_tx_gas_min {
		return model.Errcode(GasTooLittle, "gas")
	}

	//检查Nonce
	tm := mining.GetLongChain().GetTransactionManager()
	_, err = tm.CheckNonce(txItr)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	// engine.Log.Info("rpc transaction received %s", hex.EncodeToString(*txItr.GetHash()))
	if e := txItr.CheckSign(); e != nil {
		engine.Log.Info("transaction check fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, "签名错误"+e.Error())
		return
	}

	tx := txItr.(*mining.Tx_Contract)

	if len(tx.Payload) < 4 {
		res, err = model.Errcode(model.Nomarl, "调用合约参数错误")
		return
	}

	fee := tx.GetSpend()
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(tx.Vin[0].GetPukToAddr(), fee)
	if total < fee {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	//验证domain
	if !txItr.CheckDomain() {
		return model.Errcode(model.Nomarl, "domain name resolution failed")
	}

	//验证合约
	gasUsed, err := tx.PreExecNew()
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	tx.GasUsed = gasUsed

	tx.BuildHash()

	e = mining.AddTx(txItr)
	if e != nil {
		engine.Log.Error("AddTx fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	res, err = model.Tojson(model.Success)
	return
}

/*
批量将组装好并签名的交易上链
*/
func PushTxs(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	txs, ok := rj.Get("txs")
	if !ok {
		res, err = model.Errcode(model.NoField, "txs")
		return
	}

	bs, err := json.Marshal(txs)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txs")
		return
	}

	pays := make([]mining.Tx_Pay, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&pays)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txs")
		return
	}

	for k, v := range pays {
		if e := v.CheckSign(); e != nil {
			engine.Log.Info("transaction check fail:%s", e.Error())
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}

		if e := mining.AddTx(&pays[k]); e != nil {
			engine.Log.Info("AddTx fail:%s", e.Error())
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}
	}
	res, err = model.Tojson(model.Success)
	return
}

/*
通过一定范围的区块高度查询多个区块详细信息
*/
func FindBlockRange(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := toUint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
		//Txs []mining.TxItr `json:"txs"` //交易明细
	}
	//待返回的区块
	start := config.TimeNow()
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {

		bhvo := BlockHeadOut{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		//bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			//txItr, e := mining.LoadTxBase(one)
			//if e != nil {
			//	res, err = model.Errcode(model.Nomarl, e.Error())
			//	return
			//}

			txItrJson, code, txItr := mining.FindTxJsonVoV1(one)
			if txItr == nil {
				res, err = model.Errcode(model.Nomarl, "交易不存在")
				return
			}
			item := JsonMethod(txItrJson)
			txClass, vins, vouts := DealTxInfo(txItr, "", bh.Height)
			if txClass > 0 {
				item["type"] = txClass
				if vins != nil {
					item["vin"] = vins
				}
				if vouts != nil {
					item["vout"] = vouts
					item["vout_total"] = len(vouts.([]mining.VoutVO))
				}

				//if txClass == config.Wallet_tx_type_reward_W {
				//	item["hash"] = item["hash"].(string) + "-1"
				//}
			}
			//合约交易是20
			var (
				gasUsed  uint64
				gasLimit uint64
				gasPrice uint64
			)

			if txItr.Class() == config.Wallet_tx_type_contract {
				tx := txItr.(*mining.Tx_Contract)
				gasUsed = tx.GetGasLimit()
				gasLimit = config.EVM_GAS_MAX
				gasPrice = tx.GasPrice
			} else {
				gasUsed = txItr.GetGas()
				gasLimit = gasUsed
				gasPrice = config.DEFAULT_GAS_PRICE
			}
			item["free"] = txItr.GetGas()
			item["gas_used"] = gasUsed
			item["gas_limit"] = gasLimit
			item["gas_price"] = gasPrice
			item["upchaincode"] = code
			//如果是27，
			if txClass != config.Wallet_tx_type_reward_W {
				bhvo.Txs = append(bhvo.Txs, item)
			} else if txClass > 0 {
				if len(item["vout"].([]mining.VoutVO)) > 0 {
					bhvo.Txs = append(bhvo.Txs, item)
				}
			}
			//
			if txItr.Class() == config.Wallet_tx_type_mining {
				newiTem := make(map[string]interface{})
				for k, v := range item {
					newiTem[k] = v
				}
				returnVouts := []mining.VoutVO{}
				vouts := txItr.GetVout()
				outs := *vouts
				if bytes.Equal(outs[0].Address, precompiled.RewardContract) {
					returnVouts = append(returnVouts, mining.VoutVO{Address: precompiled.RewardContract.B58String(), Value: outs[0].Value})
					newiTem["type"] = config.Wallet_tx_type_mining
					newiTem["vout"] = returnVouts
					newiTem["vout_total"] = 1
					newiTem["vin"] = []mining.VinVO{}
					newiTem["vin_total"] = 0
					newiTem["hash"] = common.Bytes2Hex(*txItr.GetHash()) + "-1"
					newTx := append([]map[string]interface{}{newiTem}, bhvo.Txs...)
					bhvo.Txs = newTx
				}

			}

		}

		bhvos = append(bhvos, bhvo)
	}
	end := config.TimeNow().Sub(start)
	res, err = model.Tojson(bhvos)
	engine.Log.Info("2272行%s,%s", end, config.TimeNow().Sub(start))
	return

}

/*
通过一定范围的区块高度查询多个区块详细信息
*/
func FindBlockRangeProto(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := toUint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}

	//待返回的区块
	bhvos := make([]*[]byte, 0, endHeight-startHeight+1)

	for i := startHeight; i <= endHeight; i++ {

		bhvo := mining.BlockHeadVO{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItr, e := mining.LoadTxBase(one)
			// txItr, e := mining.FindTxBase(one)
			if e != nil {
				res, err = model.Errcode(model.Nomarl, e.Error())
				return
			}
			bhvo.Txs = append(bhvo.Txs, txItr)
		}
		bs, e := bhvo.Proto()
		if e != nil {
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}
		bhvos = append(bhvos, bs)

		// bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return

}

/*
查询区块处理交易时间
*/
func GetBlockTime(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	start, ok := rj.Get("start")
	if !ok {
		res, err = model.Errcode(model.NoField, "start")
		return
	}

	end, ok := rj.Get("end")
	if !ok {
		res, err = model.Errcode(model.NoField, "end")
		return
	}

	ty, ok := rj.Get("type")
	if !ok {
		res, err = model.Errcode(model.NoField, "type")
		return
	}

	startH := toUint64(start.(float64))
	if startH == 0 {
		startH = 1
	}
	endH := toUint64(end.(float64))
	if endH == 0 {
		endH = mining.GetHighestBlock()
	}

	type Blocktime struct {
		Height uint64
		Time   string
	}
	var max time.Duration
	var avg time.Duration
	type SaveBlockTime struct {
		Start  uint64
		End    uint64
		Max    string
		Avg    string
		Bigs   []Blocktime
		Blocks []Blocktime
	}
	bigs := make([]Blocktime, 0)
	blocks := make([]Blocktime, 0)
	var total time.Duration
	for i := startH; i <= endH; i++ {
		v := mining.SaveBlockTime[i]
		vt, err := time.ParseDuration(v)
		if err != nil {
			continue
		}

		if vt > max {
			max = vt
		}
		total += vt
		if int(ty.(float64)) != 0 {
			blocks = append(blocks, Blocktime{
				i, vt.String(),
			})
		}

		if vt.Seconds() > 1 {
			bigs = append(bigs, Blocktime{
				i, vt.String(),
			})
		}
	}
	avgt := total.Nanoseconds() / int64(endH-startH)
	avg = time.Duration(avgt)
	resultbs := SaveBlockTime{
		startH, endH,
		max.String(), avg.String(), bigs, blocks,
	}
	res, err = model.Tojson(resultbs)
	return
}

func GetContractCount(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	count := db.GetContractCount()
	data := make(map[string]interface{})
	data["count"] = count.String()
	res, err = model.Tojson(data)
	return
}

/*
获取社区节点列表通过合约获取
*/
func GetCommunityListNew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	balanceMgr := mining.GetLongChain().GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()

	communityAddrs := []crypto.AddressCoin{}
	wmc.Range(func(key, value any) bool {
		communityAddrs = append(communityAddrs, value.([]crypto.AddressCoin)...)
		return true
	})

	comms := make([]CommunityList, 0)
	for _, commAddr := range communityAddrs {
		if comminfo := balanceMgr.GetDepositCommunity(&commAddr); comminfo != nil {
			ratio := balanceMgr.GetDepositRate(&comminfo.SelfAddr)
			comms = append(comms, CommunityList{
				Addr:        comminfo.SelfAddr.B58String(),
				WitnessAddr: comminfo.WitnessAddr.B58String(),
				Payload:     comminfo.Name,
				Score:       config.Mining_vote,
				Vote:        balanceMgr.GetCommunityVote(&commAddr),
				RewardRatio: float64(ratio),
				DisRatio:    float64(ratio),
			})
		}
	}

	sort.Slice(comms, func(i, j int) bool {
		if comms[i].Vote > comms[j].Vote {
			return true
		}
		return false
	})

	// //vss := mining.GetCommunityListSortNew()
	// balanceMgr := mining.GetLongChain().GetBalance()
	// wmc := balanceMgr.GetWitnessMapCommunitys()

	// communityAddrs := []crypto.AddressCoin{}
	// wmc.Range(func(key, value any) bool {
	// 	communityAddrs = append(communityAddrs, value.([]crypto.AddressCoin)...)
	// 	return true
	// })

	// comms := make([]CommunityList, 0)
	// for _, commAddr := range communityAddrs {
	// 	if comminfo := balanceMgr.GetDepositCommunity(&commAddr); comminfo != nil {
	// 		commVote := balanceMgr.GetCommunityVote(&commAddr)
	// 		ratio := balanceMgr.GetDepositRate(&comminfo.SelfAddr)
	// 		comms = append(comms, CommunityList{
	// 			Addr:        comminfo.SelfAddr.B58String(),
	// 			WitnessAddr: comminfo.WitnessAddr.B58String(),
	// 			Payload:     comminfo.Name,
	// 			Score:       config.Mining_vote,
	// 			Vote:        commVote,
	// 			RewardRatio: float64(ratio),
	// 			DisRatio:    float64(ratio),
	// 		})
	// 	}
	// }

	// sort.Slice(comms, func(i, j int) bool {
	// 	if comms[i].Vote > comms[j].Vote {
	// 		return true
	// 	}
	// 	return false
	// })

	count := len(comms)

	out := map[string]interface{}{}
	out["count"] = count
	out["data"] = comms

	res, err = model.Tojson(out)
	return
}

/*
获取社区节点列表通过合约获取
*/
func GetAllCommunityAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	type communityAddrInfo struct {
		Name    string
		Address string
	}
	balanceMgr := mining.GetLongChain().GetBalance()
	dc := balanceMgr.GetDepositCommunityMap()

	communityAddrInfos := []*communityAddrInfo{}
	dc.Range(func(key, value any) bool {
		item := value.(*mining.DepositInfo)
		communityAddrInfos = append(communityAddrInfos, &communityAddrInfo{
			Name:    item.Name,
			Address: item.SelfAddr.B58String(),
		})
		return true
	})

	res, err = model.Tojson(communityAddrInfos)
	return
}

func GetLightList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	//vss := mining.GetLightListSortNew(pageInt, pageSizeInt)

	//lightTotal := precompiled.GetLightTotal(config.Area.Keystore.GetCoinbase().Addr)
	balanceMgr := mining.GetLongChain().GetBalance()
	lightItems := balanceMgr.GetAllLights()

	lights := make([]mining.VoteScoreOut, 0)
	for _, lightitem := range lightItems {
		if lightinfo := balanceMgr.GetDepositVote(&lightitem.SelfAddr); lightinfo != nil {
			lights = append(lights, mining.VoteScoreOut{
				Witness: lightinfo.WitnessAddr.B58String(),
				Addr:    lightinfo.SelfAddr.B58String(),
				Payload: lightinfo.Name,
				Name:    lightinfo.Name,
				Score:   lightitem.Value,
				Vote:    lightinfo.Value,
			})
		}
	}

	sort.Slice(lights, func(i, j int) bool {
		if lights[i].Vote > lights[j].Vote {
			return true
		}
		return false
	})

	count := len(lights)
	data := map[string]interface{}{}
	data["count"] = count
	start, end, ok := helppager(count, pageInt, pageSizeInt)
	if ok {
		out := make([]mining.VoteScoreOut, end-start)
		copy(out, lights[start:end])
		data["data"] = out
	} else {
		data["data"] = []interface{}{}
	}
	res, err = model.Tojson(data)
	return
}

func JsonMethod(content interface{}) map[string]interface{} {
	var name map[string]interface{}
	if mashalContent, err := json.Marshal(content); err != nil {
		engine.Log.Error("2364行错误%s", err.Error())
	} else {
		d := json.NewDecoder(bytes.NewReader(mashalContent))
		d.UseNumber()
		if err := d.Decode(&name); err != nil {
			engine.Log.Error("2369行错误%s", err.Error())
		} else {
			for k, v := range name {
				name[k] = v
			}
		}
	}
	return name

}
func GetNodeNum(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	wbg := mining.GetWitnessListSort()
	count := len(wbg.Witnesses)
	count1 := len(wbg.WitnessBackup)

	//commNum, count3 := precompiled.GetRoleTotal(config.Area.Keystore.GetCoinbase().Addr)
	commNum := int(0)
	balanceMgr := mining.GetLongChain().GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()
	lightitems := balanceMgr.GetAllLights()
	wmc.Range(func(key, value any) bool {
		if commAddrItrs, ok := wmc.Load(key); ok {
			commAddrs := commAddrItrs.([]crypto.AddressCoin)
			commNum += len(commAddrs)
		}
		return true
	})

	data := make(map[string]interface{})
	data["wit_num"] = count
	data["back_wit_num"] = count1
	data["community_num"] = commNum
	data["light_num"] = len(lightitems)
	return model.Tojson(data)
}

// 获取收藏的代币
func GetErc20(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	key := []byte(config.DBKEY_ERC20_COIN)
	type Coin struct {
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
		Address string `json:"address"`
	}
	list, err := db.LevelDB.HGetAll(key)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
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
	res, err = model.Tojson(coinList)
	return
}

// 获取收藏的代币（余额）
func GetErc20Value(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrP, ok := rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrP)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	address := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&address)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	if len(address) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	addressMap := make(map[string]struct{}, len(address))
	for _, v := range address {
		addressMap[v] = struct{}{}
	}

	key := []byte(config.DBKEY_ERC20_COIN)
	list, err := db.LevelDB.HGetAll(key)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	addrList := config.Area.Keystore.GetAddr()

	coinList := []Erc20Info{}
	for _, v := range list {
		if _, ok := addressMap[string(v.Field)]; !ok {
			continue
		}

		contractAddr := string(v.Field)
		value := precompiled.GetMultiAddrBalance(crypto.AddressFromB58String(contractAddr), addrList)

		info := db.GetErc20Info(contractAddr)
		decimal := info.Decimals

		coinList = append(coinList, Erc20Info{
			Address:  info.Address,
			Name:     info.Name,
			Symbol:   info.Symbol,
			Decimals: info.Decimals,
			Value:    precompiled.ValueToString(value, decimal),
		})
	}
	return model.Tojson(coinList)
}

type Erc20Info struct {
	Address     string   `json:"address"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimals    uint8    `json:"decimals"`
	Value       string   `json:"value"`
	TotalSupply *big.Int `json:"-"`
	Sort        int      `json:"-"`
}

func GetErc20SumBalance(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contractaddressesI, ok := rj.Get("contractaddresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddresses")
		return
	}

	bs, err := json.Marshal(contractaddressesI)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "contractaddresses")
		return
	}
	contractaddresses := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&contractaddresses)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "contractaddresses")
		return
	}

	if len(contractaddresses) <= 0 {
		res, err = model.Errcode(model.NoField, "contractaddresses")
		return
	}

	addrList := config.Area.Keystore.GetAddr()

	list := make([]Erc20Info, len(contractaddresses))
	for k, c := range contractaddresses {
		value := precompiled.GetMultiAddrBalance(crypto.AddressFromB58String(c), addrList)

		info := db.GetErc20Info(c)
		decimal := info.Decimals

		list[k] = Erc20Info{
			Address:  info.Address,
			Name:     info.Name,
			Symbol:   info.Symbol,
			Decimals: info.Decimals,
			Value:    precompiled.ValueToString(value, decimal),
		}
	}

	return model.Tojson(list)
}

type Token struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
}

// 收藏币种
func AddErc20(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	tokenP, ok := rj.Get("tokens")
	if !ok {
		res, err = model.Errcode(model.NoField, "tokens")
		return
	}

	bs, err := json.Marshal(tokenP)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "tokens")
		return
	}
	tokens := make([]Token, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&tokens)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "tokens")
		return
	}

	if len(tokens) <= 0 {
		res, err = model.Errcode(model.NoField, "tokens")
		return
	}

	for _, v := range tokens {
		//保存到key即可
		key := []byte(config.DBKEY_ERC20_COIN)
		db.LevelDB.HSet(key, []byte(v.Address), precompiled.EncodeToken([]byte(v.Name), []byte(v.Symbol)))
	}

	res, err = model.Tojson("success")
	return
	////
	//name, ok := rj.Get("name")
	//if !ok {
	//	res, err = model.Errcode(model.NoField, "name")
	//	return
	//}
	//address, ok := rj.Get("address")
	//if !ok {
	//	res, err = model.Errcode(model.NoField, "address")
	//	return
	//}
	////保存到key即可
	//key := []byte(config.DBKEY_ERC20_COIN)
	//_, err = db.LevelDB.HSet(key, []byte(address.(string)), []byte(name.(string)))
	//if err != nil {
	//	res, err = model.Errcode(model.Nomarl, err.Error())
	//	return
	//}
	//res, err = model.Tojson("success")
	//return
}

// 收藏币种
func DelErc20(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrP, ok := rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrP)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	address := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&address)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	if len(address) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	//保存到key即可
	key := []byte(config.DBKEY_ERC20_COIN)
	for _, v := range address {
		db.LevelDB.HDel(key, []byte(v))
	}
	return model.Tojson("success")
}

// 搜索币种
func SearchErc20(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	keywordI, ok := rj.Get("keyword")

	if !ok || keywordI.(string) == "" {
		erc20list := db.GetAllErc20Info()
		return model.Tojson(erc20list)
	}

	keyword := keywordI.(string)

	chars := ".@$!%*#_~?^&\\-"
	//只能包含字母或数字 ,特殊字符
	if matched, _ := regexp.MatchString(fmt.Sprintf("^(?i)[a-z0-9%s]+$", chars), keyword); !matched {
		return model.Errcode(model.TypeWrong, "keyword")
	}

	list := make([]db.Erc20Info, 0)
	if isAddressCoin(keyword) {
		//判断地址前缀是否正确
		if len(base58.Decode(keyword[len(keyword)-1:])) > 0 && crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(keyword)) {
			erc := db.GetErc20Info(keyword)
			list = append(list, erc)
			return model.Tojson(list)
		}
	}

	erc20list := db.GetAllErc20Info()
	keyword = strEscape(keyword, chars)
	reg := regexp.MustCompile(strings.ToLower(keyword))
	for _, v := range erc20list {
		ind := reg.FindIndex([]byte(strings.ToLower(v.Name)))
		if ind != nil {
			v.Sort = ind[0]
			list = append(list, v)
		}
	}

	sort.Sort(db.Erc20Sort(list))

	return model.Tojson(list)
}

func isAddressCoin(str string) bool {
	if str == "" {
		return false
	}
	lastStr := str[len(str)-1:]
	lastByte := base58.Decode(lastStr)
	if len(lastByte) <= 0 {
		return false
	}
	preLen := int(lastByte[0])
	if preLen > len(str)-1 {
		return false
	}

	return true
}

// 字符串转义
func strEscape(word string, chars string) string {
	var rs []rune
	m := make(map[rune]struct{})
	for _, c := range chars {
		m[c] = struct{}{}
	}
	for _, r := range word {
		if _, ok := m[r]; ok {
			rs = append(rs, '\\', r)
		} else {
			rs = append(rs, r)
		}
	}

	return string(rs)
}

// 代币转账
func TransferErc20(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	toItr, ok := rj.Get("toaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "toaddress")
		return
	}
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}

	decimal := precompiled.GetDecimals(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)
	amount := precompiled.StringToValue(amountItr.(string), decimal)
	if amount == nil || amount.Cmp(big.NewInt(0)) == 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	payload := precompiled.BuildErc20TransferBigInput(toItr.(string), amount)
	comment := common.Bytes2Hex(payload)
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

func GetTokenBalance(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	balance := precompiled.GetBigBalance(src, src.B58String(), addr)

	decimal := precompiled.GetDecimals(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)

	res, err = model.Tojson(precompiled.ValueToString(balance, decimal))

	return res, err
}

func GetTokenBalances(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addressesP, ok := rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addressesP)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	addresses := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	err = decoder.Decode(&addresses)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	addrs := make([]crypto.AddressCoin, len(addresses))
	for k, v := range addresses {
		addr := crypto.AddressFromB58String(v)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, addr) {
			res, err = model.Errcode(ContentIncorrectFormat, "addresses")
			return
		}
		//_, ok := config.Area.Keystore.FindAddress(addr)
		//if !ok {
		//	res, err = model.Errcode(ContentIncorrectFormat, "addresses")
		//	return
		//}
		addrs[k] = addr
	}

	addrItr, ok := rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}

	type RTokenBalance struct {
		Address string `json:"address"`
		Value   string `json:"value"`
	}
	result := make([]RTokenBalance, len(addresses))

	balances := precompiled.GetMultiBigBalance(contractAddr, addrs)
	if len(balances) == 0 {
		return model.Tojson(result)
	}

	tokenbalances := make([]TokenBalance, len(addresses))
	for k, v := range addrs {
		balance := balances[k]
		tokenbalances[k] = TokenBalance{v, balance}
	}
	sort.Sort(TokenBalanceSort(tokenbalances))

	decimal := precompiled.GetDecimals(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)
	for k, v := range tokenbalances {
		result[k] = RTokenBalance{v.Address.B58String(), precompiled.ValueToString(v.Value, decimal)}
	}

	return model.Tojson(result)
}

type TokenBalance struct {
	Address crypto.AddressCoin
	Value   *big.Int
}
type TokenBalanceSort []TokenBalance

func (this TokenBalanceSort) Len() int {
	return len(this)
}

func (this TokenBalanceSort) Less(i, j int) bool {
	return this[i].Value.Cmp(this[j].Value) > 0
}

func (this TokenBalanceSort) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func DelayTCoin(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if config.AddrPre != "TEST" {
		res, err = model.Errcode(model.Nomarl, "the rpc interface only call in test chain")
		return
	}

	src := config.Area.Keystore.GetCoinbase().Addr
	//部署合约时，value为0
	amount := toUint64(0)
	gas := uint64(2 * config.Wallet_tx_gas_min)

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := faucet.FAUCET_BIN
	source := ""
	sourceStr, ok := rj.Get("source")
	if ok && rj.VerifyType("source", "string") {
		source = sourceStr.(string)
	}
	defaultClass := config.FAUCET_CONTRACT
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, nil, amount, gas, 1, pwd, comment, source, uint64(defaultClass), gasPrice, "")

	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	result["contract_address"] = txpay.Vout[0].Address.B58String()
	res, err = model.Tojson(result)

	return res, err
}

// 领取测试币
func ReceiveTCOin(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if config.AddrPre != "TEST" {
		res, err = model.Errcode(NotTestChain, "the rpc interface only call in test chain")
		return
	}
	src := config.Area.Keystore.GetCoinbase().Addr

	gas := uint64(config.Wallet_tx_gas_min)
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)
	dst := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	pwd := config.Wallet_keystore_default_pwd
	comment := "发送测试币"

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
	temp = new(big.Int).Div(temp, big.NewInt(1024))
	if gas < temp.Uint64() {
		res, err = model.Errcode(GasTooLittle, "gas")
		return
	}
	amount := toUint64(config.FAUCET_COIN)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas)
	if total < amount+gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	t1 := mining.GetFaucetTime(addr)
	if !config.TimeNow().After(time.Unix(t1+24*60*60, 0)) {
		res, err = model.Errcode(TestCoinLimit, "lock time has not expired. Please try again later")
		return
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}

	txpay, err := mining.SendToAddress(&src, &dst, amount, gas, frozenHeight, pwd, comment, domain, domainType)
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	err = mining.SetFaucetTime(addr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 充值水龙头
func ReChargeTCOin(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if config.AddrPre != "TEST" {
		res, err = model.Errcode(model.Nomarl, "the rpc interface only call in test chain")
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
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}

	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	comment := faucet.BuildReChargeTCoinInput()
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

/*
通过一定范围的区块高度查询多个区块详细信息新版本
*/
func FindBlockRangeV1(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := toUint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
	}

	//待返回的区块
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		bhvo := BlockHeadOut{}
		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		txMap := sync.Map{}
		var wg sync.WaitGroup
		for _, one := range bh.Tx {
			wg.Add(1)
			go func(txid []byte) {
				defer wg.Done()
				data := parseOneTx(txid)
				txMap.Store(utils.Bytes2string(txid), data)
			}(one)
		}

		wg.Wait()

		for _, one := range bh.Tx {
			if v, ok := txMap.Load(utils.Bytes2string(one)); ok {
				tx := v.(map[string]interface{})
				bhvo.Txs = append(bhvo.Txs, tx)
			}
		}
		bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return
}

/*
通过一定范围的区块高度查询多个区块详细信息新版本
*/
func FindBlockRangeV1Old(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := toUint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
		//Txs []mining.TxItr `json:"txs"` //交易明细
	}
	//待返回的区块
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {

		bhvo := BlockHeadOut{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		txMap := sync.Map{}
		txCustomMap := sync.Map{}
		var wg sync.WaitGroup

		for _, one := range bh.Tx {
			wg.Add(1)
			go func(hash []byte) {
				defer wg.Done()
				txItrJson, code, txItr := mining.FindTxJsonVoV1(hash)
				if txItr == nil {
					return
				}
				item := JsonMethod(txItrJson)
				item["block_height"] = bh.Height
				item["timestamp"] = bh.Time
				item["blockhash"] = hex.EncodeToString(bh.Hash)

				txClass, vins, vouts, isDomainRealTx := DealTxInfoV3(txItr, "", bh.Height)
				if isDomainRealTx {
					//合约交易是20
					gasUsed := txItr.GetGas()
					gasLimit := config.EVM_GAS_MAX
					gasPrice := config.DEFAULT_GAS_PRICE

					item["free"] = txItr.GetGas()
					item["gas_used"] = gasUsed
					item["gas_limit"] = gasLimit
					item["gas_price"] = gasPrice
					item["upchaincode"] = code
					item["is_custom"] = false

					txMap.Store(hex.EncodeToString(hash), item)
				} else {
					if txClass > 0 {
						item["type"] = txClass
						if vins != nil {
							item["vin"] = vins
						}
						if vouts != nil {
							item["vout"] = vouts
							item["vout_total"] = len(vouts.([]mining.VoutVO))
						}
					}
					//合约交易是20
					gasUsed := txItr.GetGas()
					gasLimit := config.EVM_GAS_MAX
					gasPrice := config.DEFAULT_GAS_PRICE

					item["free"] = txItr.GetGas()
					item["gas_used"] = gasUsed
					item["gas_limit"] = gasLimit
					item["gas_price"] = gasPrice
					item["upchaincode"] = code
					item["gas"] = txItr.GetGas()
					if txClass == config.Wallet_tx_type_contract {
						item["is_custom"] = false
					} else {
						item["is_custom"] = true
					}

					txMap.Store(hex.EncodeToString(hash), item)
				}

				// 同时添加域名真实交易和域名伪造交易
				// 避免hash冲突
				if txClass == config.Wallet_tx_type_domain_register ||
					txClass == config.Wallet_tx_type_domain_renew ||
					txClass == config.Wallet_tx_type_domain_withdraw ||
					txClass == config.Wallet_tx_type_subdomain_register ||
					txClass == config.Wallet_tx_type_subdomain_renew ||
					txClass == config.Wallet_tx_type_subdomain_withdraw {
					customTxs := []map[string]interface{}{}
					item2 := make(map[string]interface{})
					for k, v := range item {
						item2[k] = v
					}

					// 特殊处理官方提现，根域名提现
					if txClass == config.Wallet_tx_type_domain_withdraw || txClass == config.Wallet_tx_type_subdomain_withdraw {
						for i := range vins.([]mining.VinVO) {
							item2 := make(map[string]interface{})
							for k, v := range item {
								item2[k] = v
							}
							returnVins := []mining.VinVO{}
							returnVouts := []mining.VoutVO{}
							returnVins = append(returnVins, vins.([]mining.VinVO)[i])
							returnVouts = append(returnVouts, vouts.([]mining.VoutVO)[i])
							item2["vin"] = returnVins
							item2["vout"] = returnVouts
							item2["vout_total"] = len(returnVouts)

							customHash := fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
							item2["hash"] = customHash
							item2["type"] = txClass
							gasUsed := txItr.GetGas()
							gasLimit := config.EVM_GAS_MAX
							gasPrice := config.DEFAULT_GAS_PRICE

							item2["free"] = txItr.GetGas()
							item2["gas_used"] = gasUsed
							item2["gas_limit"] = gasLimit
							item2["gas_price"] = gasPrice
							item2["upchaincode"] = code
							item2["gas"] = txItr.GetGas()
							item2["block_height"] = bh.Height
							item2["is_custom"] = true
							customTxs = append(customTxs, item2)
							txCustomMap.Store(hex.EncodeToString(hash), customTxs)
						}
					} else {
						if vins != nil {
							item2["vin"] = vins
						}
						if vouts != nil {
							item2["vout"] = vouts
							item2["vout_total"] = len(vouts.([]mining.VoutVO))
						}

						customHash := fmt.Sprintf("%s%d", hex.EncodeToString(hash), txClass)
						item2["hash"] = customHash
						item2["type"] = txClass
						gasUsed := txItr.GetGas()
						gasLimit := config.EVM_GAS_MAX
						gasPrice := config.DEFAULT_GAS_PRICE

						item2["free"] = txItr.GetGas()
						item2["gas_used"] = gasUsed
						item2["gas_limit"] = gasLimit
						item2["gas_price"] = gasPrice
						item2["upchaincode"] = code
						item2["gas"] = txItr.GetGas()
						item2["block_height"] = bh.Height
						item2["is_custom"] = true
						customTxs = append(customTxs, item2)
						txCustomMap.Store(hex.EncodeToString(hash), customTxs)
					}
				}

				// 伪造交易
				if txClass == config.Wallet_tx_type_vote_in ||
					txClass == config.Wallet_tx_type_vote_out ||
					txClass == config.Wallet_tx_type_community_out ||
					txClass == config.Wallet_tx_type_mining ||
					txClass == config.Wallet_tx_type_community_distribute {
					customTxs := []map[string]interface{}{}
					rewardLogs := precompiled.GetRewardHistoryLog(bh.Height, hash)
					for i, v := range rewardLogs {
						if v.Utype < 1 || v.Utype > 3 {
							continue
						}

						returnVins := []*mining.VinVO{}
						returnVins = append(returnVins, &mining.VinVO{
							Addr: precompiled.RewardContract.B58String(),
						})
						returnVouts := []*mining.VoutVO{}
						returnVouts = append(returnVouts, &mining.VoutVO{
							Address: v.Into,
							Value:   v.Reward.Uint64(),
						})

						txType := toUint64(0)
						switch v.Utype {
						case 1:
							txType = config.Wallet_tx_type_reward_W
						case 2:
							txType = config.Wallet_tx_type_reward_C
						case 3:
							txType = config.Wallet_tx_type_reward_L
						}
						rand.Seed(int64(txType + txClass + uint64(i)))
						txInfo := mining.TxBaseVO{
							Hash:       fmt.Sprintf("%v%d", hex.EncodeToString(hash), rand.Intn(10000)), //本交易hash，不参与区块hash，只用来保存
							Type:       txType,                                                          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易                //输入交易数量
							Vin_total:  uint64(len(returnVins)),
							Vin:        returnVins, //交易输入
							Vout_total: uint64(len(returnVouts)),
							Vout:       returnVouts,                            //交易输出
							Payload:    hex.EncodeToString(txItr.GetPayload()), //备注信息
							BlockHash:  hex.EncodeToString(bh.Hash),            //本交易属于的区块hash，不参与区块hash，只用来保存
							Timestamp:  uint64(bh.Time),
						}
						txitem := make(map[string]interface{})
						b, _ := json.Marshal(txInfo)
						json.Unmarshal(b, &txitem)
						// 补充字段
						txitem["gas"] = txItr.GetGas()
						txitem["free"] = 0
						txitem["gas_limit"] = config.EVM_GAS_MAX
						txitem["gas_price"] = config.DEFAULT_GAS_PRICE
						txitem["gas_used"] = 0
						txitem["upchaincode"] = 2
						txitem["block_height"] = bh.Height
						txitem["temputpye"] = v.Utype
						txitem["is_custom"] = true
						customTxs = append(customTxs, txitem)
					}
					if len(customTxs) > 0 {
						txCustomMap.Store(hex.EncodeToString(hash), customTxs)
					}
				}
			}(one)

		}

		wg.Wait()

		for _, one := range bh.Tx {
			txhash := hex.EncodeToString(one)
			if v, ok := txMap.Load(txhash); ok {
				value := v.(map[string]interface{})
				bhvo.Txs = append(bhvo.Txs, value)
			}

			if v, ok := txCustomMap.Load(txhash); ok {
				if txs, ok := v.([]map[string]interface{}); ok {
					bhvo.Txs = append(bhvo.Txs, txs...)
				}
			}
		}
		bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return
}

/*
获取一定区块高度范围内的区块新版本,列表链上真实交易，浏览器专用接口
*/
func FindBlockRangeV2(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := toUint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
		//Txs []mining.TxItr `json:"txs"` //交易明细
	}
	//待返回的区块
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {

		bhvo := BlockHeadOut{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}

		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		txMap := sync.Map{}
		//txCustomMap := sync.Map{}
		var wg sync.WaitGroup

		for _, one := range bh.Tx {
			wg.Add(1)
			go func(hash []byte) {
				defer wg.Done()
				txItrJson, code, txItr := mining.FindTxJsonVoV1(hash)
				if txItr == nil {
					return
				}
				item := JsonMethod(txItrJson)
				item["block_height"] = bh.Height
				item["timestamp"] = bh.Time
				item["blockhash"] = hex.EncodeToString(bh.Hash)

				gasUsed := txItr.GetGas()
				gasLimit := config.EVM_GAS_MAX
				gasPrice := config.DEFAULT_GAS_PRICE

				item["free"] = txItr.GetGas()
				item["gas_used"] = gasUsed
				item["gas_limit"] = gasLimit
				item["gas_price"] = gasPrice
				item["upchaincode"] = code

				txMap.Store(hex.EncodeToString(hash), item)
			}(one)
		}

		wg.Wait()

		for _, one := range bh.Tx {
			txhash := hex.EncodeToString(one)
			if v, ok := txMap.Load(txhash); ok {
				value := v.(map[string]interface{})
				bhvo.Txs = append(bhvo.Txs, value)
			}
		}
		bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return
}

type BlockHeadOut struct {
	BH  *mining.BlockHead        `json:"bh"`  //区块
	Txs []map[string]interface{} `json:"txs"` //交易明细
}

type queryTx struct {
	isAll    bool
	onlyErc  bool     //仅erc交易
	addrs    sync.Map //vin过滤
	ercAddrs sync.Map //erc过滤
}

// 解析真实交易和自定义交易,仅限config.Address_history_tx的记录
//func parseTxAndCustomTx(hash []byte, height uint64, qtx queryTx) []map[string]interface{} {
//	resTxs := make([]map[string]interface{}, 0)
//	txItr, code, _ := mining.FindTx(hash)
//	if txItr == nil {
//		return nil
//	}
//
//	txItrJson := txItr.GetVOJSON()
//	gasUsed := txItr.GetGas()
//	gasLimit := config.EVM_GAS_MAX
//	gasPrice := config.DEFAULT_GAS_PRICE
//
//	vouts := txItr.GetVout()
//	vout := *vouts
//	vins := txItr.GetVin()
//	vin := *vins
//	vinAddrStr := ""
//	if len(vin) > 0 {
//		vinAddrStr = vin[0].GetPukToAddr().B58String()
//	}
//	voutAddrStr := ""
//	voutValue := toUint64(0)
//	if len(vout) > 0 {
//		voutAddrStr = vout[0].GetAddrStr()
//		voutValue = vout[0].Value
//	}
//
//	payload := txItr.GetPayload()
//	// 截断payload
//	txClass, toAddr, value := precompiled.UnpackCustomPayload(payload)
//	if txClass == 0 {
//		txClass = txItr.Class()
//	}
//
//	// 截断payload
//	payload64Str := cutPayload(payload)
//
//	// 解析erc20,erc721,erc1155交易
//	// 备注erc20真实交易过滤掉
//	erc20Info := db.GetErc20Info(voutAddrStr)
//	isErc20 := false
//	_, ercQuery := qtx.ercAddrs.Load(voutAddrStr)
//	if erc20Info.Address != "" && ercQuery {
//		customItem := JsonMethod(txItrJson)
//		gasUsed := txItr.GetGas()
//		gasLimit := config.EVM_GAS_MAX
//		gasPrice := config.DEFAULT_GAS_PRICE
//		customItem["free"] = gasUsed
//		customItem["gas_used"] = gasUsed
//		customItem["gas_limit"] = gasLimit
//		customItem["gas_price"] = gasPrice
//		customItem["upchaincode"] = code
//		if value > 0 {
//			itemVout := []mining.VoutVO{mining.VoutVO{
//				Address: toAddr,
//				Value:   voutValue,
//			},
//			}
//			customItem["vout"] = itemVout
//		}
//		//当主链币大于0,则认为是主链币交易
//		if voutValue == 0 {
//			customItem["type"] = txClass
//			customItem["tokenType"] = 1
//			customItem["erc20_address"] = erc20Info.Address
//			customItem["erc20_name"] = erc20Info.Name
//			customItem["erc20_symbol"] = erc20Info.Symbol
//			customItem["erc20_decimals"] = erc20Info.Decimals
//			customItem["erc20_total_supply"] = erc20Info.TotalSupply
//			//代币
//			ercDec := int32(0)
//			if erc20Info.Decimals > 0 {
//				ercDec = -int32(erc20Info.Decimals)
//			}
//			customItem["voutTotalStr"] = decimal.NewFromFloatWithExponent(float64(value)/1e8, ercDec).String()
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			isErc20 = true
//		}
//	}
//
//	erc721Info := db.GetErc721Info(voutAddrStr)
//	if erc721Info.Address != "" && ercQuery {
//		customItem := JsonMethod(txItrJson)
//		gasUsed := txItr.GetGas()
//		gasLimit := config.EVM_GAS_MAX
//		gasPrice := config.DEFAULT_GAS_PRICE
//		customItem["free"] = gasUsed
//		customItem["gas_used"] = gasUsed
//		customItem["gas_limit"] = gasLimit
//		customItem["gas_price"] = gasPrice
//		customItem["upchaincode"] = code
//		customItem["tokenType"] = 2
//		customItem["payload"] = payload64Str
//		resTxs = append(resTxs, customItem)
//		isErc20 = true
//	}
//
//	erc1155Info := db.GetErc1155Info(voutAddrStr)
//	if erc1155Info.Address != "" && ercQuery {
//		customItem := JsonMethod(txItrJson)
//		gasUsed := txItr.GetGas()
//		gasLimit := config.EVM_GAS_MAX
//		gasPrice := config.DEFAULT_GAS_PRICE
//		customItem["free"] = gasUsed
//		customItem["gas_used"] = gasUsed
//		customItem["gas_limit"] = gasLimit
//		customItem["gas_price"] = gasPrice
//		customItem["upchaincode"] = code
//		customItem["tokenType"] = 3
//		customItem["payload"] = payload64Str
//		resTxs = append(resTxs, customItem)
//		isErc20 = true
//	}
//
//	if qtx.onlyErc || isErc20 {
//		return resTxs
//	}
//
//	//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
//	if bytes.Equal(vout[0].Address, precompiled.RewardContract) && txItr.Class() == config.Wallet_tx_type_mining {
//		switch txClass {
//		case config.Wallet_tx_type_mining:
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				gasUsed := txItr.GetGas()
//				gasLimit := config.EVM_GAS_MAX
//				gasPrice := config.DEFAULT_GAS_PRICE
//
//				itemVin := []mining.VinVO{
//					mining.VinVO{
//						Addr: "",
//					},
//				}
//				customItem["vin"] = itemVin
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetRewardHistoryLog(height, hash)
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				if v.Utype != 1 {
//					continue
//				}
//
//				_, vinOk := qtx.addrs.Load(precompiled.RewardContract.B58String())
//				_, voutOk := qtx.addrs.Load(v.Into)
//				if !(vinOk || voutOk) && !qtx.isAll {
//					continue
//				}
//
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Reward.Uint64(),
//				},
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//
//				txType := config.Wallet_tx_type_reward_W
//
//				customItem["type"] = txType
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//		}
//		return resTxs
//	}
//
//	//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
//	if bytes.Equal(vout[0].Address, precompiled.RewardContract) && txItr.Class() == config.Wallet_tx_type_contract {
//		switch txClass {
//		case config.Wallet_tx_type_community_in:
//			//社区节点质押,只需要变更类型
//			customItem := JsonMethod(txItrJson)
//			customItem["type"] = txClass
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_community_out:
//			//社区节点取消质押
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: vinAddrStr,
//					Value:   config.Mining_vote,
//				},
//				}
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//				customItem["type"] = txClass
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetRewardHistoryLog(height, hash)
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				if v.Utype < 2 || v.Utype > 3 {
//					continue
//				}
//
//				_, vinOk := qtx.addrs.Load(precompiled.RewardContract.B58String())
//				_, voutOk := qtx.addrs.Load(v.Into)
//				if !(vinOk || voutOk) && !qtx.isAll {
//					continue
//				}
//
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Reward.Uint64(),
//				},
//				}
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//
//				txType := toUint64(0)
//				switch v.Utype {
//				case 2:
//					txType = config.Wallet_tx_type_reward_C
//				case 3:
//					txType = config.Wallet_tx_type_reward_L
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["type"] = txType
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			return resTxs
//		case config.Wallet_tx_type_vote_in:
//			//轻节点投票
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				customItem["type"] = txClass
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetRewardHistoryLog(height, hash)
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				if v.Utype < 2 || v.Utype > 3 {
//					continue
//				}
//
//				_, vinOk := qtx.addrs.Load(precompiled.RewardContract.B58String())
//				_, voutOk := qtx.addrs.Load(v.Into)
//				if !(vinOk || voutOk) && !qtx.isAll {
//					continue
//				}
//
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Reward.Uint64(),
//				},
//				}
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//
//				txType := toUint64(0)
//				switch v.Utype {
//				case 1:
//					txType = config.Wallet_tx_type_reward_W
//				case 2:
//					txType = config.Wallet_tx_type_reward_C
//				case 3:
//					txType = config.Wallet_tx_type_reward_L
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["type"] = txType
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			return resTxs
//		case config.Wallet_tx_type_vote_out:
//			//轻节点取消投票
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: vinAddrStr,
//					Value:   value,
//				},
//				}
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//				customItem["type"] = txClass
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetRewardHistoryLog(height, hash)
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				if v.Utype < 2 || v.Utype > 3 {
//					continue
//				}
//
//				_, vinOk := qtx.addrs.Load(precompiled.RewardContract.B58String())
//				_, voutOk := qtx.addrs.Load(v.Into)
//				if !(vinOk || voutOk) && !qtx.isAll {
//					continue
//				}
//
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Reward.Uint64(),
//				},
//				}
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//
//				txType := toUint64(0)
//				switch v.Utype {
//				case 1:
//					txType = config.Wallet_tx_type_reward_W
//				case 2:
//					txType = config.Wallet_tx_type_reward_C
//				case 3:
//					txType = config.Wallet_tx_type_reward_L
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["type"] = txType
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			return resTxs
//		case config.Wallet_tx_type_light_in:
//			//轻节点质押
//			customItem := JsonMethod(txItrJson)
//			customItem["type"] = txClass
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_light_out:
//			//轻节点取消质押
//			customItem := JsonMethod(txItrJson)
//			itemVin := []mining.VinVO{mining.VinVO{
//				Addr: precompiled.RewardContract.B58String(),
//			},
//			}
//			itemVout := []mining.VoutVO{mining.VoutVO{
//				Address: vinAddrStr,
//				Value:   config.Mining_light_min,
//			},
//			}
//			customItem["vin"] = itemVin
//			customItem["vout"] = itemVout
//			customItem["type"] = txClass
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_community_distribute: //社区分账
//			//真实交易
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: vinAddrStr,
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: precompiled.RewardContract.B58String(),
//					Value:   value,
//				},
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//				customItem["type"] = txClass
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetRewardHistoryLog(height, hash)
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				if v.Utype != 2 && v.Utype != 3 {
//					continue
//				}
//
//				_, vinOk := qtx.addrs.Load(precompiled.RewardContract.B58String())
//				_, voutOk := qtx.addrs.Load(v.Into)
//				if !(vinOk || voutOk) && !qtx.isAll {
//					continue
//				}
//
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Reward.Uint64(),
//				},
//				}
//
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//
//				txType := toUint64(0)
//				switch v.Utype {
//				case 1:
//					txType = config.Wallet_tx_type_reward_W
//				case 2:
//					txType = config.Wallet_tx_type_reward_C
//				case 3:
//					txType = config.Wallet_tx_type_reward_L
//				}
//
//				customItem["type"] = txType
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//			return resTxs
//		}
//	}
//
//	// 3.根域名注册合约交易
//	if bytes.Equal(vout[0].Address, ens.GetRegisterAddr()) {
//		if txItr.Class() == config.Wallet_tx_type_contract {
//			switch txClass {
//			case config.Wallet_tx_type_domain_register: // 域名注册
//				customItem := JsonMethod(txItrJson)
//				customItem["type"] = txClass
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//				return resTxs
//			case config.Wallet_tx_type_domain_renew: // 域名续费
//				customItem := JsonMethod(txItrJson)
//				customItem["type"] = txClass
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//				return resTxs
//			case config.Wallet_tx_type_domain_withdraw: // 域名提现
//				//真实交易
//				_, vinOk := qtx.addrs.Load(vinAddrStr)
//				_, voutOk := qtx.addrs.Load(voutAddrStr)
//				if vinOk || voutOk || qtx.isAll {
//					customItem := JsonMethod(txItrJson)
//					itemVin := []mining.VinVO{mining.VinVO{
//						Addr: precompiled.DomainRegisterContract.B58String(),
//					},
//					}
//					itemVout := []mining.VoutVO{mining.VoutVO{
//						Address: vinAddrStr,
//						Value:   0,
//					},
//					}
//					customItem["free"] = gasUsed
//					customItem["gas_used"] = gasUsed
//					customItem["gas_limit"] = gasLimit
//					customItem["gas_price"] = gasPrice
//					customItem["upchaincode"] = code
//					customItem["vin"] = itemVin
//					customItem["vout"] = itemVout
//					customItem["type"] = txClass
//					customItem["tokenType"] = 4
//					customItem["payload"] = payload64Str
//					resTxs = append(resTxs, customItem)
//				}
//
//				rewardLogs := precompiled.GetEnsHistoryLog(height, *txItr.GetHash())
//				for i, v := range rewardLogs {
//					_, vinOk := qtx.addrs.Load(precompiled.DomainRegisterContract.B58String())
//					_, voutOk := qtx.addrs.Load(v.Into)
//					if !(vinOk || voutOk) && !qtx.isAll {
//						continue
//					}
//
//					customItem := JsonMethod(txItrJson)
//					itemVin := []mining.VinVO{mining.VinVO{
//						Addr: precompiled.RewardContract.B58String(),
//					},
//					}
//					itemVout := []mining.VoutVO{mining.VoutVO{
//						Address: v.Into,
//						Value:   v.Amount.Uint64(),
//					},
//					}
//					customItem["free"] = gasUsed
//					customItem["gas_used"] = gasUsed
//					customItem["gas_limit"] = gasLimit
//					customItem["gas_price"] = gasPrice
//					customItem["upchaincode"] = code
//					customItem["vin"] = itemVin
//					customItem["vout"] = itemVout
//					customItem["type"] = txClass
//					customItem["tokenType"] = 4
//					customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//					customItem["is_custom"] = true
//					customItem["payload"] = payload64Str
//					resTxs = append(resTxs, customItem)
//				}
//				return resTxs
//			}
//		}
//	}
//
//	//跨链转账
//	if txItr.Class() == config.Wallet_tx_type_contract {
//		switch txClass {
//		case config.Wallet_tx_type_l1_l2_transfer: // l1到l2转账
//			customItem := JsonMethod(txItrJson)
//			customItem["type"] = txClass
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_l2_l1_withdraw: // l1到l2转账
//			customItem := JsonMethod(txItrJson)
//			vout[0].Value = value
//			customItem["vout"] = vout
//			customItem["type"] = txClass
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		}
//	}
//
//	//子域名合约
//	if txItr.Class() == config.Wallet_tx_type_contract {
//		switch txClass {
//		case config.Wallet_tx_type_domain_register: // 子域名注册
//			customItem := JsonMethod(txItrJson)
//			customItem["type"] = config.Wallet_tx_type_subdomain_register
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_domain_renew: // 子域名续费
//			customItem := JsonMethod(txItrJson)
//			customItem["type"] = config.Wallet_tx_type_subdomain_renew
//			customItem["free"] = gasUsed
//			customItem["gas_used"] = gasUsed
//			customItem["gas_limit"] = gasLimit
//			customItem["gas_price"] = gasPrice
//			customItem["upchaincode"] = code
//			customItem["tokenType"] = 4
//			customItem["payload"] = payload64Str
//			resTxs = append(resTxs, customItem)
//			return resTxs
//		case config.Wallet_tx_type_domain_withdraw: // 子域名提现
//			//真实交易
//			_, vinOk := qtx.addrs.Load(vinAddrStr)
//			_, voutOk := qtx.addrs.Load(voutAddrStr)
//			if vinOk || voutOk || qtx.isAll {
//				customItem := JsonMethod(txItrJson)
//				itemVin := []mining.VinVO{mining.VinVO{
//					Addr: precompiled.RewardContract.B58String(),
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: vinAddrStr,
//					Value:   0,
//				},
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//				customItem["type"] = config.Wallet_tx_type_subdomain_withdraw
//				customItem["tokenType"] = 4
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//
//			rewardLogs := precompiled.GetEnsHistoryLog(height, *txItr.GetHash())
//			for i, v := range rewardLogs {
//				customItem := JsonMethod(txItrJson)
//				itemVin := []mining.VinVO{mining.VinVO{
//					//Addr: precompiled.DomainRegisterContract.B58String(),
//					Addr: vinAddrStr,
//				},
//				}
//				itemVout := []mining.VoutVO{mining.VoutVO{
//					Address: v.Into,
//					Value:   v.Amount.Uint64(),
//				},
//				}
//				customItem["free"] = gasUsed
//				customItem["gas_used"] = gasUsed
//				customItem["gas_limit"] = gasLimit
//				customItem["gas_price"] = gasPrice
//				customItem["upchaincode"] = code
//				customItem["vin"] = itemVin
//				customItem["vout"] = itemVout
//				customItem["type"] = txClass
//				customItem["tokenType"] = 4
//				customItem["hash"] = fmt.Sprintf("%s%d", hex.EncodeToString(hash), i)
//				customItem["is_custom"] = true
//				customItem["payload"] = payload64Str
//				resTxs = append(resTxs, customItem)
//			}
//			return resTxs
//		}
//	}
//
//	item := JsonMethod(txItrJson)
//	item["free"] = gasUsed
//	item["gas_used"] = gasUsed
//	item["gas_limit"] = gasLimit
//	item["gas_price"] = gasPrice
//	item["upchaincode"] = code
//	item["tokenType"] = 4
//	item["payload"] = payload64Str
//
//	// 其它的交易
//	resTxs = append(resTxs, item)
//
//	return resTxs
//}

/*
构建离线交易
*/
func CreateOfflineTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	addrItr, ok := rj.Get("srcaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "srcaddress")
		return
	}
	srcaddr := addrItr.(string)
	//src := crypto.AddressFromB58String(srcaddr)
	////判断地址前缀是否正确
	//if !crypto.ValidAddr(config.AddrPre, src) {
	//	res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
	//	return
	//}

	addrItr, ok = rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)
	//dst := crypto.AddressFromB58String(addr)
	//if !crypto.ValidAddr(config.AddrPre, dst) {
	//	res, err = model.Errcode(model.ContentIncorrectFormat, "address")
	//	return
	//}

	// 获取密码
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取amount
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(5010, "amount")
		return
	}

	// 获取nonce
	nonceItr, ok := rj.Get("nonce")
	if !ok {
		res, err = model.Errcode(model.NoField, "nonce")
		return
	}
	nonce := toUint64(nonceItr.(float64))
	if nonce < 0 {
		res, err = model.Errcode(5010, "nonce")
		return
	}

	// 获取currentHeight
	currentHeightItr, ok := rj.Get("currentHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "currentHeight")
		return
	}
	currentHeight := toUint64(currentHeightItr.(float64))
	if currentHeight <= 0 {
		res, err = model.Errcode(5010, "currentHeight")
		return
	}

	// 获取gas
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	// 获取冻结高度
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// 获取comment
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}

	// 获取keyStorePath 钱包路径
	keyStorePath := ""
	keyStorePathItr, ok := rj.Get("key_store_path")
	if ok {
		keyStorePath = keyStorePathItr.(string)
	}

	tx, hash, err := mining.CreateOfflineTx(keyStorePath, srcaddr, addr, pwd, comment, amount, gas, frozenHeight, nonce, currentHeight, domain, domainType)
	//paybs, _ := txpay.Proto()

	data := make(map[string]interface{})
	data["tx"] = tx
	data["hash"] = hash

	// 测试

	//
	//txItr, _ := mining.ParseTxBaseProto(0, paybs)
	////fmt.Println(fmt.Sprintf("%+v", txItr))
	//
	//e := mining.AddTx(txItr)
	//if e != nil {
	//	engine.Log.Info("AddTx fail:%s", e.Error())
	//	res, err = model.Errcode(model.Nomarl, e.Error())
	//	return
	//}

	//txjsonBs, e := base64.StdEncoding.DecodeString(data["txpay"].(string))
	//if e != nil {
	//	engine.Log.Info("DecodeString fail:%s", e.Error())
	//	res, err = model.Errcode(model.Nomarl, e.Error())
	//	return
	//}
	//txItr, _ := mining.ParseTxBaseProto(0, &txjsonBs)
	//engine.Log.Error("77777777777777777777777777777777777777777777")
	//fmt.Println(fmt.Sprintf("%+v", txItr))

	res, err = model.Tojson(data)

	return
}

/*
构建离线合约交易
*/
func CreateOfflineContractTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	addrItr, ok := rj.Get("srcaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "srcaddress")
		return
	}
	if !mining.TypeConversion("string", addrItr) {
		return []byte{}, errors.New("addrItr 类型错误")
	}
	srcaddr := addrItr.(string)

	//判断地址前缀是否正确
	//if srcaddr != "" && !crypto.ValidAddr("iCom", crypto.AddressFromB58String(srcaddr)) {
	//	res, err = model.Errcode(model.TypeWrong, "srcaddress 地址错误")
	//	return
	//}

	match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", srcaddr)
	if srcaddr != "" && !match {
		res, err = model.Errcode(model.TypeWrong, "srcaddress 地址错误")
		return
	}

	addrItr, ok = rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !mining.TypeConversion("string", addrItr) {
		res, err = model.Errcode(model.TypeWrong, "address 地址错误")
		return
	}
	addr := addrItr.(string)
	//判断地址前缀是否正确
	//if addr != "" && !crypto.ValidAddr("iCom", crypto.AddressFromB58String(addr)) {
	//	res, err = model.Errcode(model.TypeWrong, "address 地址错误")
	//	return
	//}

	match, _ = regexp.MatchString("^[a-zA-Z0-9]{20,60}$", addr)
	if addr != "" && !match {
		res, err = model.Errcode(model.TypeWrong, "address 地址错误")
		return
	}

	// 获取密码
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	if !mining.TypeConversion("string", pwdItr) {
		res, err = model.Errcode(model.TypeWrong, "pwd 地址错误")
		return
	}
	pwd := pwdItr.(string)

	// 获取amount
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	if !mining.TypeConversion("float64", amountItr) {
		res, err = model.Errcode(model.TypeWrong, "amount 地址错误")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(5010, "amount")
		return
	}

	// 获取gas
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	if !mining.TypeConversion("float64", gasItr) {
		res, err = model.Errcode(model.TypeWrong, "gas 地址错误")
		return
	}
	gas := toUint64(gasItr.(float64))

	// 获取gasPrice
	gasPriceItr, ok := rj.Get("gas_price")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas_price")
		return
	}
	if !mining.TypeConversion("float64", gasPriceItr) {
		res, err = model.Errcode(model.TypeWrong, "gas_price 地址错误")
		return
	}
	gasPrice := toUint64(gasPriceItr.(float64))

	// 获取冻结高度
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")

	if !mining.TypeConversion("float64", frozenHeightItr) {
		res, err = model.Errcode(model.TypeWrong, "frozen_height 地址错误")
		return
	}

	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// 获取nonce
	nonceItr, ok := rj.Get("nonce")
	if !ok {
		res, err = model.Errcode(model.NoField, "nonce")
		return
	}
	if !mining.TypeConversion("float64", nonceItr) {
		res, err = model.Errcode(model.TypeWrong, "nonce 地址错误")
		return
	}
	nonce := toUint64(nonceItr.(float64))
	if nonce < 0 {
		res, err = model.Errcode(5010, "nonce")
		return
	}

	// 获取currentHeight
	currentHeightItr, ok := rj.Get("currentHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "currentHeight")
		return
	}
	if !mining.TypeConversion("float64", currentHeightItr) {
		res, err = model.Errcode(model.TypeWrong, "currentHeight 地址错误")
		return
	}
	currentHeight := toUint64(currentHeightItr.(float64))
	if currentHeight <= 0 {
		res, err = model.Errcode(5010, "currentHeight")
		return
	}

	// 获取comment
	comment := ""
	commentItr, ok := rj.Get("comment")

	if !mining.TypeConversion("string", commentItr) {
		res, err = model.Errcode(model.TypeWrong, "comment 地址错误")
		return
	}

	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	if len(comment) < 8 {
		res, err = model.Errcode(model.TypeWrong, "comment 地址错误")
		return
	}

	// 获取abi
	abi := ""
	abiItr, ok := rj.Get("abi")

	if !mining.TypeConversion("string", abiItr) {
		res, err = model.Errcode(model.TypeWrong, "abi 错误")
		return
	}

	if ok && rj.VerifyType("comment", "string") {
		abi = abiItr.(string)
	}

	// 获取source
	source := ""
	sourceItr, ok := rj.Get("source")

	if !mining.TypeConversion("string", sourceItr) {
		res, err = model.Errcode(model.TypeWrong, "source 错误")
		return
	}

	if ok && rj.VerifyType("comment", "string") {
		source = sourceItr.(string)
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")

	if ok && rj.VerifyType("domain", "string") {
		if !mining.TypeConversion("string", domainItr) {
			res, err = model.Errcode(model.TypeWrong, "domain 地址错误")
			return
		}
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")

	if ok {
		if !mining.TypeConversion("float64", domainTypeItr) {
			res, err = model.Errcode(model.TypeWrong, "domain_type 地址错误")
			return
		}
		domainType = toUint64(domainTypeItr.(float64))
	}

	// 获取keyStorePath 钱包路径
	keyStorePath := ""
	keyStorePathItr, ok := rj.Get("key_store_path")

	if ok {
		if !mining.TypeConversion("string", keyStorePathItr) {
			res, err = model.Errcode(model.TypeWrong, "key_store_path 地址错误")
			return
		}
		keyStorePath = keyStorePathItr.(string)
	}

	tx, hash, addressContract, err := mining.CreateOfflineContractTx(keyStorePath, srcaddr, addr, pwd, comment, amount, gas, frozenHeight, gasPrice, nonce, currentHeight, domain, domainType, abi, source)
	//paybs, _ := txpay.Proto()
	if err != nil {
		engine.Log.Error("build tx fail :%s", err.Error())
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	data := make(map[string]interface{})
	data["tx"] = tx
	data["hash"] = hash
	data["address"] = addressContract
	// 测试

	//
	//txItr, _ := mining.ParseTxBaseProto(0, paybs)
	////fmt.Println(fmt.Sprintf("%+v", txItr))
	//
	//e := mining.AddTx(txItr)
	//if e != nil {
	//	engine.Log.Info("AddTx fail:%s", e.Error())
	//	res, err = model.Errcode(model.Nomarl, e.Error())
	//	return
	//}

	//txjsonBs, e := base64.StdEncoding.DecodeString(data["txpay"].(string))
	//if e != nil {
	//	engine.Log.Info("DecodeString fail:%s", e.Error())
	//	res, err = model.Errcode(model.Nomarl, e.Error())
	//	return
	//}
	//txItr, _ := mining.ParseTxBaseProto(0, &txjsonBs)
	//engine.Log.Error("77777777777777777777777777777777777777777777")
	//fmt.Println(fmt.Sprintf("%+v", txItr))

	res, err = model.Tojson(data)

	return
}

/*
获取comment
*/
func GetComment(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	// 获取tag
	tagItr, ok := rj.Get("tag")
	if !ok {
		res, err = model.Errcode(model.NoField, "tag")
		return
	}
	tag := tagItr.(string)

	// 获取jsonData
	jsonDataItr, ok := rj.Get("jsonData")
	if !ok {
		res, err = model.Errcode(model.NoField, "jsonData")
		return
	}
	jsonDataItrs := jsonDataItr.(string)

	b := []byte(jsonDataItrs)

	jsonData := make(map[string]interface{})
	//engine.Log.Error(jsonDataItrs)
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "json类型错误")
		return
	}

	//jsonData := jsonDataT.(map[string]interface{})

	comment := ""

	comment, err = mining.GetComment(tag, jsonData)

	//bb, _ := json.Marshal(comment)
	//engine.Log.Error("000000000000000000000000000000")
	//fmt.Println(bb)

	if err != nil {
		res, err = model.Errcode(model.TypeWrong, err.Error())
		return
	}

	data := make(map[string]interface{})
	data["comment"] = comment
	//
	res, err = model.Tojson(data)

	return
}

/*
MultDeal
*/
func MultDeal(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	// 获取tag
	tagItr, ok := rj.Get("tag")
	if !ok {
		res, err = model.Errcode(model.NoField, "tag")
		return
	}
	tag := tagItr.(string)

	// 获取jsonData
	jsonDataItr, ok := rj.Get("jsonData")
	if !ok {
		res, err = model.Errcode(model.NoField, "jsonData")
		return
	}
	jsonDataItrs := jsonDataItr.(string)

	b := []byte(jsonDataItrs)

	jsonData := make(map[string]interface{})

	err = json.Unmarshal(b, &jsonData)
	if err != nil {

		fmt.Println("Umarshal failed:", err)
		return
	}

	addrItr, ok := rj.Get("srcaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "srcaddress")
		return
	}
	srcaddr := addrItr.(string)
	//src := crypto.AddressFromB58String(srcaddr)
	////判断地址前缀是否正确
	//if !crypto.ValidAddr(config.AddrPre, src) {
	//	res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
	//	return
	//}

	addrItr, ok = rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)
	//dst := crypto.AddressFromB58String(addr)
	//if !crypto.ValidAddr(config.AddrPre, dst) {
	//	res, err = model.Errcode(model.ContentIncorrectFormat, "address")
	//	return
	//}

	// 获取密码
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取amount
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(5010, "amount")
		return
	}

	// 获取gas
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	// 获取gasPrice
	gasPriceItr, ok := rj.Get("gas_price")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas_price")
		return
	}
	gasPrice := toUint64(gasPriceItr.(float64))

	// 获取冻结高度
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// 获取nonce
	nonceItr, ok := rj.Get("nonce")
	if !ok {
		res, err = model.Errcode(model.NoField, "nonce")
		return
	}
	nonce := toUint64(nonceItr.(float64))
	if nonce < 0 {
		res, err = model.Errcode(5010, "nonce")
		return
	}

	// 获取currentHeight
	currentHeightItr, ok := rj.Get("currentHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "currentHeight")
		return
	}
	currentHeight := toUint64(currentHeightItr.(float64))
	if currentHeight <= 0 {
		res, err = model.Errcode(5010, "currentHeight")
		return
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}

	// 获取keyStorePath 钱包路径
	keyStorePath := ""
	keyStorePathItr, ok := rj.Get("key_store_path")
	if ok {
		keyStorePath = keyStorePathItr.(string)
	}

	hash, err := mining.MultDeal(tag, jsonData, keyStorePath, srcaddr, addr, pwd, amount, gas, frozenHeight, gasPrice, nonce, currentHeight, domain, domainType)

	if err != nil {
		engine.Log.Info("AddTx fail:%s", err.Error())
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	data := make(map[string]interface{})
	data["hash"] = hash

	res, err = model.Tojson(data)

	return
}

/*
构建离线交易V1
*/
func CreateOfflineTxV1(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// 获取密码
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// 获取nonce
	nonceItr, ok := rj.Get("nonce")
	if !ok {
		res, err = model.Errcode(model.NoField, "nonce")
		return
	}
	nonce := toUint64(nonceItr.(float64))
	if nonce < 0 {
		res, err = model.Errcode(5010, "nonce")
		return
	}

	// 获取currentHeight
	currentHeightItr, ok := rj.Get("currentHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "currentHeight")
		return
	}
	currentHeight := toUint64(currentHeightItr.(float64))
	if currentHeight <= 0 {
		res, err = model.Errcode(5010, "currentHeight")
		return
	}

	// 获取冻结高度
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := toUint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = toUint64(domainTypeItr.(float64))
	}

	// 获取keyStorePath 钱包路径
	keyStorePath := ""
	keyStorePathItr, ok := rj.Get("key_store_path")
	if ok {
		keyStorePath = keyStorePathItr.(string)
	}

	// 获取tag
	tagItr, ok := rj.Get("tag")
	if !ok {
		res, err = model.Errcode(model.NoField, "tag")
		return
	}
	tag := tagItr.(string)

	// 获取jsonData
	jsonDataItr, ok := rj.Get("jsonData")
	if !ok {
		res, err = model.Errcode(model.NoField, "jsonData")
		return
	}
	jsonDataItrs := jsonDataItr.(string)

	result := mining.BuildOfflineTx(keyStorePath, pwd, nonce, currentHeight, frozenHeight, domainType, domain, tag, jsonDataItrs)
	dataInfo := utils2.ParseDataInfo(result)
	if dataInfo.Code != 200 {
		return nil, errors.New(dataInfo.Data.(string))
	}
	info := dataInfo.Data.(map[string]interface{})
	return model.Tojson(info)
}

/*
获得全网/出块/候选 见证人
*/
func witnessesListWithRangeV1(wits []*mining.BackupWitness, start, end int) (res []WitnessList, total int, err error) {

	chain := mining.GetLongChain()
	if chain == nil {
		return nil, 0, errors.New("get chain failed")
	}

	wvos := []WitnessList{}
	total = len(wits)

	for _, v := range wits {
		addBlockCount, addBlockReward := GetAddressAddBlockReward(v.Addr.B58String())
		wvo := WitnessList{
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
	sort.Sort(WitnessListSort(wvos))

	// 分页
	outs := []WitnessList{}
	for i, one := range wvos {
		if i >= start && i < end {
			outs = append(outs, one)
		}
	}

	return outs, total, nil
}

// 根据区块区块哈希和交易索引，查询区块中的指定交易信息
func FindTxByBlockHashAndIndex(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hashItr, ok := rj.Get("hash")
	if !ok {
		res, err = model.Errcode(model.NoField, "hash")
		return
	}

	indexItr, ok := rj.Get("index")
	if !ok {
		res, err = model.Errcode(model.NoField, "index")
		return
	}

	hash, err := hex.DecodeString(hashItr.(string))
	if err != nil {
		res, err = model.Errcode(ContentIncorrectFormat)
		return
	}
	bh, err := mining.LoadBlockHeadByHash(&hash)
	if bh == nil {
		res, err = model.Errcode(model.NotExist, "hash")
		return
	}

	txInd := toUint64(indexItr.(float64))
	if txInd >= bh.NTx {
		res, err = model.Errcode(model.Nomarl, "index out of range")
		return
	}

	txid := bh.Tx[txInd]
	txidStr := hex.EncodeToString(txid)

	isType1 := false
	if txidStr[len(txidStr)-2:] == "-1" {
		isType1 = true
		txidStr = txidStr[:len(txidStr)-2]
	}

	outMap := make(map[string]interface{})

	txItr, code, _ := mining.FindTx(txid)
	tx, err := mining.LoadTxBase(txid)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "tx is wrong")
		return
	}
	item := JsonMethod(txItr.GetVOJSON())
	txClass, vins, vouts := DealTxInfoV2(tx)
	if txClass > 0 {
		item["type"] = txClass
		if vins != nil {
			item["vin"] = vins
		}
		if vouts != nil {
			item["vout"] = vouts
			item["vout_total"] = len(vouts.([]mining.VoutVO))
		}
	}
	if isType1 {
		returnVouts := []mining.VoutVO{}
		vouts := tx.GetVout()
		outs := *vouts
		returnVouts = append(returnVouts, mining.VoutVO{Address: precompiled.RewardContract.B58String(), Value: outs[0].Value})
		item["type"] = config.Wallet_tx_type_mining
		item["vout"] = returnVouts
		item["vout_total"] = 1
		item["vin"] = []mining.VinVO{}
		item["vin_total"] = 0
		item["hash"] = common.Bytes2Hex(*tx.GetHash()) + "-1"
	}
	outMap["txinfo"] = item
	outMap["upchaincode"] = code
	outMap["blockheight"] = bh.Height
	outMap["blockhash"] = hex.EncodeToString(bh.Hash)
	outMap["timestamp"] = bh.Time
	res, err = model.Tojson(outMap)
	return
}

// 根据交易id，查询交易字节信息
func FindTxBs(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	txItr, ok := rj.Get("txid")
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txidStr := txItr.(string)
	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	tx, err := mining.LoadTxBase(txid)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	txbs, err := tx.Proto()
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "tx proto fail")
		return
	}
	outMap := make(map[string]interface{})
	outMap["type"] = tx.Class()
	outMap["tx"] = base64.StdEncoding.EncodeToString(*txbs)
	res, err = model.Tojson(outMap)
	return
}

/*
提供给中心化服务的主链内存状态信息
*/
func GetChainState(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}
	addrStr := addrItr.(string)

	addr := crypto.AddressFromB58String(addrStr)
	ok = crypto.ValidAddr(config.AddrPre, addr)
	if !ok {
		return model.Errcode(ContentIncorrectFormat, "address")
	}

	typeItr, ok := rj.Get("type")
	if !ok {
		return model.Errcode(model.NoField, "type")
	}
	typeInt := toUint64(typeItr.(float64))
	if typeInt != 3 && mining.GetAddrState(addr) != int(typeInt) {
		return model.Errcode(RuleField, "address type")
	}

	item := map[string]interface{}{
		"allReward":    toUint64(0),
		"remainReward": toUint64(0),
		"allBlock":     toUint64(0),
		"ratio":        float64(0),
	}

	switch typeInt {
	case 1: // 见证人
		addBlockCount, addBlockReward := GetAddressAddBlockReward(addr.B58String())
		ratio := float64(mining.GetLongChain().Balance.GetDepositRate(&addr))
		item["allReward"] = addBlockReward
		item["allBlock"] = addBlockCount
		item["ratio"] = float64(ratio)
		//case 2: // 社区
		//	var remainReward uint64
		//	allReward := toUint64(0)
		//	if val, err := db.LevelTempDB.Get(config.BuildCommunityAllReward(addr)); err == nil {
		//		allReward = utils.BytesToUint64(val)
		//	}
		//	cRewardPool := big.NewInt(0)
		//	if val, err := db.LevelTempDB.Get(config.BuildCommunityRewardPool(addr)); err == nil {
		//		cRewardPool.SetUint64(utils.BytesToUint64(val))
		//	}

		//	//allReward := precompiled.GetNodeReward(addr)
		//	ratio := mining.GetLongChain().Balance.GetDepositRate(&addr)
		//	if cRewardPool.Cmp(big.NewInt(0)) > 0 {
		//		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(ratio))), new(big.Int).SetInt64(100))
		//		cBigValue := new(big.Int).Sub(cRewardPool, lTotal)
		//		remainReward = cBigValue.Uint64()
		//	}

		//	item["ratio"] = float64(ratio)
		//	item["allReward"] = allReward
		//	item["remainReward"] = remainReward
		//case 3: // 轻节点
		//	cAddr := crypto.AddressCoin{}
		//	value := toUint64(0)
		//	allReward := toUint64(0)
		//	if depositInfo := mining.GetLongChain().Balance.GetDepositVote(&addr); depositInfo != nil {
		//		cAddr = depositInfo.WitnessAddr
		//		value = depositInfo.Value
		//	}
		//	if val, err := db.LevelTempDB.Get(config.BuildLightAllReward(addr)); err == nil {
		//		allReward = utils.BytesToUint64(val)
		//	}

		//	cRewardPool := big.NewInt(0)
		//	if val, err := db.LevelTempDB.Get(config.BuildCommunityRewardPool(cAddr)); err == nil {
		//		cRewardPool.SetUint64(utils.BytesToUint64(val))
		//	}
		//	cVoteInt := mining.GetLongChain().Balance.GetCommunityVote(&cAddr)
		//	cVote := big.NewInt(int64(cVoteInt))
		//	cRate := mining.GetLongChain().Balance.GetDepositRate(&cAddr)
		//	var remainReward uint64
		//	if cRewardPool.Cmp(big.NewInt(0)) > 0 && cVote.Cmp(big.NewInt(0)) > 0 && value != 0 {
		//		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
		//		v := new(big.Int).Mul(new(big.Int).SetUint64(value), new(big.Int).SetInt64(1e8))
		//		ratio := new(big.Int).Quo(v, cVote)
		//		lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
		//		remainReward = lBigValue.Uint64()
		//	}
		//	item["allReward"] = allReward
		//	item["remainReward"] = remainReward
	}

	return model.Tojson(item)
}

/*
查询见证人与社区奖励池
*/
func GetRewardPool(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}
	addrStr := addrItr.(string)

	addr := crypto.AddressFromB58String(addrStr)
	ok = crypto.ValidAddr(config.AddrPre, addr)
	if !ok {
		return model.Errcode(ContentIncorrectFormat, "address")
	}

	addrType := mining.GetAddrState(addr)

	balanceMgr := mining.GetLongChain().Balance
	rewardPoolValue := new(big.Int)
	switch addrType {
	case 1: // 见证人
		witRewardPools := balanceMgr.GetWitnessRewardPool()
		if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(addr)); ok {
			rewardPoolValue, _ = balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(addr, witRewardPool.(*big.Int))
		}
	case 2: // 社区
		rewardPoolValue, _ = balanceMgr.CalculateCommunityRewardAndLightReward(addr)
	case 3: // 轻节点
		if lightvoteinfo := balanceMgr.GetDepositVote(&addr); lightvoteinfo != nil {
			_, lightrewards := balanceMgr.CalculateCommunityRewardAndLightReward(lightvoteinfo.WitnessAddr)
			if v, ok := lightrewards[utils.Bytes2string(lightvoteinfo.SelfAddr)]; ok {
				rewardPoolValue = v
			}
		}
	}

	item := map[string]interface{}{
		"address":    addr.B58String(),
		"type":       addrType,
		"pool_value": rewardPoolValue.Uint64(),
	}

	return model.Tojson(item)
}

/*
查询节点的当前累计的见证人+社区+轻节点投票奖励池
*/
func GetBalanceVote(addrs ...crypto.AddressCoin) uint64 {
	wlist := []crypto.AddressCoin{} //见证人地址集
	clist := []crypto.AddressCoin{} //社区地址集
	llist := []crypto.AddressCoin{} //轻节点地址集
	if len(addrs) != 0 {
		for _, one := range addrs {
			addrType := mining.GetAddrState(one)
			switch addrType {
			case 1: // 见证人
				wlist = append(wlist, one)
			case 2: // 社区
				clist = append(clist, one)
			case 3: // 轻节点
				llist = append(llist, one)
			}
		}
	} else {
		for _, one := range config.Area.Keystore.GetAddrAll() {
			addrType := mining.GetAddrState(one.Addr)
			switch addrType {
			case 1: // 见证人
				wlist = append(wlist, one.Addr)
			case 2: // 社区
				clist = append(clist, one.Addr)
			case 3: // 轻节点
				llist = append(llist, one.Addr)
			}
		}
	}

	eg := errgroup.Group{}
	eg.SetLimit(runtime.NumCPU())
	wPoolValue := atomic.Uint64{} //见证人奖励池
	cPoolValue := atomic.Uint64{} //社区奖励池
	lPoolValue := atomic.Uint64{} //轻节点奖励池

	balanceMgr := mining.GetLongChain().Balance

	for _, one := range wlist {
		addr := one
		eg.Go(func() error {
			witRewardPools := balanceMgr.GetWitnessRewardPool()
			if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(addr)); ok {
				value, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(addr, witRewardPool.(*big.Int))
				wPoolValue.Add(value.Uint64())
			}
			return nil
		})
	}

	for _, one := range clist {
		addr := one
		eg.Go(func() error {
			value, _ := balanceMgr.CalculateCommunityRewardAndLightReward(addr)
			cPoolValue.Add(value.Uint64())
			return nil
		})
	}

	for _, one := range llist {
		addr := one
		eg.Go(func() error {
			if lightvoteinfo := balanceMgr.GetDepositVote(&addr); lightvoteinfo != nil {
				_, lightrewards := balanceMgr.CalculateCommunityRewardAndLightReward(lightvoteinfo.WitnessAddr)
				if value, ok := lightrewards[utils.Bytes2string(lightvoteinfo.SelfAddr)]; ok {
					lPoolValue.Add(value.Uint64())
				}
			}
			return nil
		})
	}

	eg.Wait()

	return wPoolValue.Load() + cPoolValue.Load() + lPoolValue.Load()
}

// 节点账户下的代币余额
func GetAccountsErc20Value(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(10000)
	}
	pageSizeInt := int(pageSize.(float64))
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	addrItr, ok := rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)
	contractAddr := crypto.AddressFromB58String(addr)

	type AccountToken struct {
		Address string `json:"address"`
		Value   string `json:"value"`
	}

	list := config.Area.Keystore.GetAddr()
	total := len(config.Area.Keystore.GetAddr())
	start := (pageInt - 1) * pageSizeInt
	end := start + pageSizeInt

	vos := make([]AccountToken, 0)
	if start > total {
		return model.Tojson(vos)
	}
	if end > total {
		end = total
	}

	addrs := make([]crypto.AddressCoin, end-start)
	for i, val := range list[start:end] {
		addrs[i] = val.Addr
	}
	balances := precompiled.GetMultiBigBalance(contractAddr, addrs)
	if len(balances) == 0 {
		return model.Tojson(vos)
	}

	decimal := precompiled.GetDecimals(config.Area.Keystore.GetCoinbase().Addr.B58String(), addr)
	for k := range addrs {
		vos = append(vos, AccountToken{addrs[k].B58String(), precompiled.ValueToString(balances[k], decimal)})
	}
	return model.Tojson(vos)
}

// 合约缓存设置
func SetContractAddrCache(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	setAddrsStr, errcode := getArrayStrParams(rj, "setAddresses")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "setAddresses")
		return
	}
	delAddrsStr, errcode := getArrayStrParams(rj, "delAddresses")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "delAddresses")
		return
	}
	setAddrs := []crypto.AddressCoin{}
	delAddrs := []crypto.AddressCoin{}
	for _, addrStr := range setAddrsStr {
		addr := crypto.AddressFromB58String(addrStr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, addr) {
			res, err = model.Errcode(ContentIncorrectFormat, "setAddress")
			return
		}
		setAddrs = append(setAddrs, addr)
	}

	for _, addrStr := range delAddrsStr {
		addr := crypto.AddressFromB58String(addrStr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, addr) {
			res, err = model.Errcode(ContentIncorrectFormat, "setAddress")
			return
		}
		delAddrs = append(delAddrs, addr)
	}

	data := []string{}
	addrs := mining.GetLongChain().GetBalance().SetCacheContract(setAddrs, delAddrs)
	for _, v := range addrs {
		data = append(data, v.B58String())
	}

	res, err = model.Tojson(data)
	return
}

// 跨链转账
func CrossChainTransfer(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	comment := cross.BuildDepositEth()

	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 跨链提现
func CrossChainWithdraw(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	txItr, ok := rj.Get("txHash")
	if !ok {
		res, err = model.Errcode(model.NoField, "txHash")
		return
	}
	txHashBs, err := hex.DecodeString(txItr.(string))
	if err != nil {
		return model.Errcode(model.NoField, "txHash")

	}
	txHash := common.BytesToHash(txHashBs)

	l2BlockHashItr, ok := rj.Get("l2BlockHash")
	if !ok {
		res, err = model.Errcode(model.NoField, "l2BlockHash")
		return
	}
	l2BlockHashBs, err := hex.DecodeString(l2BlockHashItr.(string))
	if err != nil {
		return model.Errcode(model.NoField, "l2BlockHash")

	}
	l2BlockHash := common.BytesToHash(l2BlockHashBs)

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := new(big.Int).SetUint64(toUint64(amountItr.(float64)))

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	comment := cross.BuildWithdraw(txHash, l2BlockHash, amount)

	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 截断payload,有点用户合约太大了,造成rpc传输失败
func cutPayload(payload []byte) string {
	if len(payload) > 64 {
		return common.Bytes2Hex(payload[:64])
	}
	return common.Bytes2Hex(payload)
}

// 帮助分页
func helppager(totallen, page, pageSize int) (int, int, bool) {
	if page <= 0 || pageSize <= 0 {
		return 0, 0, false
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start > totallen {
		return 0, 0, false
	}

	if end > totallen {
		end = totallen
	}

	return start, end, true
}
