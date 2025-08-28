package rpc

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/evm/vm/storage"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/snapshot"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	rpc "web3_gui/libp2parea/adapter/sdk/jsonrpc2"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func RegisterChainWitnessVoteRPC() {
	//----base----
	rpc.RegisterRPC("getblocktime", GetBlockTime)       //--查询区块处理交易时间
	rpc.RegisterRPC("getdepositnum", GetDepositNum)     //--获取节点质押数量
	rpc.RegisterRPC("getlightList", GetLightList)       //--获取所有轻节点列表
	rpc.RegisterRPC("getNodeNum", GetNodeNum)           //--获取节点数量
	rpc.RegisterRPC("getnonce", GetNonce)               //--获取某个地址的nonce
	rpc.RegisterRPC("getPendingNonce", getPendingNonce) //--获取某个地址的nonce
	rpc.RegisterRPC("getnoncemore", GetNonceMore)       //--获取多个地址的nonce
	rpc.RegisterRPC("getroleaddr", GetRoleAddr)         //查询一个地址角色
	rpc.RegisterRPC("getaccountother", GetAccountOther) //查询一个地址余额
	//----witness----
	rpc.RegisterRPC("getvoteaddr", GetVoteAddr) //查询一个地址给谁投票，票数

	//----token----
	rpc.RegisterRPC("getinfo", HandleGetInfo)                               //获取基本信息{"method","getinfo")"params":{"id":10}}
	rpc.RegisterRPC("getnodebalance", HandleGetNodeBalance)                 //获取节点总余额
	rpc.RegisterRPC("getnewaddress", HandleGetNewAddress)                   //创建新地址 {"method":"getnewaddress","params":{"password":"123456"}}
	rpc.RegisterRPC("listaccounts", HandleListAccounts)                     //帐号列表{"method":"listaccounts"}
	rpc.RegisterRPC("listmultaccounts", HandleListMultAccounts)             //多签帐号列表{"method":"listmultaccounts"}
	rpc.RegisterRPC("infoaccount", HandleInfoAccount)                       //账号详情{"method":"infoaccount"}
	rpc.RegisterRPC("getaccount", HandleGetAccount)                         //获取某一帐号余额{"method":"getaccount","params":{"address":"1AX9mfCRZkdEg5Ci3G5SLcyGgecj6GTzLo"}}
	rpc.RegisterRPC("validateaddress", HandleValidateAddress)               //验证地址{"method":"validateaddress","params":{"address":"12EUY1EVnLJe4Ejb1VaL9NbuDQbBEV"}}
	rpc.RegisterRPC("import", Import)                                       //导入钱包
	rpc.RegisterRPC("export", Export)                                       //导出钱包
	rpc.RegisterRPC("sendtoaddress", SendToAddress)                         //转账
	rpc.RegisterRPC("sendtoaddressmore", SendToAddressmore)                 //给多人转账
	rpc.RegisterRPC("depositin", DepositIn)                                 //缴纳押金，成为见证人
	rpc.RegisterRPC("depositout", DepositOut)                               //退还押金
	rpc.RegisterRPC("votein", VoteInNew)                                    //给见证人投票押金
	rpc.RegisterRPC("canwitness", CheckCanWitness)                          // 检查是否能成为见证者节点
	rpc.RegisterRPC("voteout", VoteOutNew)                                  //退还给见证人投票押金
	rpc.RegisterRPC("updatepwd", UpdatePwd)                                 //修改支付密码
	rpc.RegisterRPC("createkeystore", CreateKeystore)                       //创建密钥文件
	rpc.RegisterRPC("namesinreg", NameInReg)                                //域名注册
	rpc.RegisterRPC("namesintansfer", NameInTransfer)                       //域名修改
	rpc.RegisterRPC("namesinrenew", NameInRenew)                            //域名续费
	rpc.RegisterRPC("namesinupdate", NameInUpdate)                          //域名更新
	rpc.RegisterRPC("namesout", NameOut)                                    //域名注销，退还押金
	rpc.RegisterRPC("getnames", GetNames)                                   //获取自己注册的域名列表
	rpc.RegisterRPC("findname", FindName)                                   //查询域名
	rpc.RegisterRPC("gettransactionhistory", GetTransactionHistoty)         //获得转账交易历史记录
	rpc.RegisterRPC("getwitnessinfo", GetWitnessInfo)                       //查询见证人状态
	rpc.RegisterRPC("getcandidatelist", GetCandidateList)                   //获得候选见证人列表
	rpc.RegisterRPC("getwitnesslist", GetWitnessesList)                     //获得见证人列表
	rpc.RegisterRPC("getcommunitylist", GetCommunityListNew)                //获取社区节点列表
	rpc.RegisterRPC("getallcommunityaddr", GetAllCommunityAddr)             //获取社区节点地址列表
	rpc.RegisterRPC("getvotelist", GetVoteList)                             //获得自己投过票的列表
	rpc.RegisterRPC("findtx", FindTx)                                       //检查一笔交易是否成功上链
	rpc.RegisterRPC("findtxs", FindTxs)                                     //检查批量查询交易是否成功上链
	rpc.RegisterRPC("findtxbyblockhashandindex", FindTxByBlockHashAndIndex) //根据区块区块哈希和交易索引，查询区块中的指定交易信息
	rpc.RegisterRPC("findtxbs", FindTxBs)                                   //根据交易id，查询交易字节信息
	rpc.RegisterRPC("findblock", FindBlock)                                 //通过区块高度查询一个区块信息
	rpc.RegisterRPC("blockHeight", GetBlockHeight)                          //获取最新区块高度
	rpc.RegisterRPC("findblockbyhash", FindBlockByHash)                     //通过区块Hash查询一个区块信息
	rpc.RegisterRPC("getcommunityreward", GetCommunityReward)               //获取一个社区累计奖励
	//rpc.RegisterRPC("sendcommunityreward", SendCommunityReward)             //分发社区奖励
	//rpc.RegisterRPC("getrewardpool", GetRewardPool) //查询见证人与社区奖励池
	rpc.RegisterRPC("tokenpublish", TokenPublish)         //发布一个token
	rpc.RegisterRPC("tokenpay", TokenPay)                 //使用token支付
	rpc.RegisterRPC("tokenpaymore", TokenPayMore)         //使用token支付给多个地址
	rpc.RegisterRPC("tokeninfo", TokenInfo)               //查询token信息
	rpc.RegisterRPC("tokenlist", TokenList)               //查询token信息列表
	rpc.RegisterRPC("newtokenorder", NewTokenOrder)       //发布一个token订单
	rpc.RegisterRPC("newtokenorderv2", NewTokenOrderV2)   //发布一个token订单
	rpc.RegisterRPC("canceltokenorder", CancelTokenOrder) //取消token订单
	rpc.RegisterRPC("tokenorderinfo", TokenOrderInfo)     //Token订单信息
	rpc.RegisterRPC("tokenorderpool", TokenOrderPool)     //Token订单池
	rpc.RegisterRPC("pushtx", PushTx)                     //将组装好并签名的交易上链
	//rpc.RegisterRPC("pushtxs", PushTxs)                                     //批量将组装好并签名的交易上链
	rpc.RegisterRPC("pushContractTx", PushContractTx)                   //批量将组装好并签名的交易上链
	rpc.RegisterRPC("PreContractTx", PreContractTx)                     // 合约预执行校验
	rpc.RegisterRPC("PreCallContract", PreCallContract)                 //批量将组装好并签名的交易上链
	rpc.RegisterRPC("createOfflineTx", CreateOfflineTx)                 //构建普通离线交易
	rpc.RegisterRPC("createOfflineTxV1", CreateOfflineTxV1)             //构建普通离线交易
	rpc.RegisterRPC("createOfflineContractTx", CreateOfflineContractTx) //构建合约离线交易
	rpc.RegisterRPC("getComment", GetComment)                           //获取commentget
	rpc.RegisterRPC("multDeal", MultDeal)                               //获取commentget
	rpc.RegisterRPC("getblocksrange", FindBlockRange)                   //获取一定区块高度范围内的区块
	rpc.RegisterRPC("getblocksrangeproto", FindBlockRangeProto)         //获取一定区块高度范围内的区块，返回的是proto格式
	rpc.RegisterRPC("findvalue", GetValueForKey)                        //查询数据库中key对应的value
	rpc.RegisterRPC("multiaccounts", HandleMultiAccounts)               //查询多地址账户信息
	rpc.RegisterRPC("multibalance", MultiBalance)                       //查询多地址账户信息
	//RegisterRPC("getnodetotal",          GetNodeTotal)          //获取存储节点总数
	//----nft----
	rpc.RegisterRPC("buildnft", BuildNFT)           //铸造NFT
	rpc.RegisterRPC("transfernft", TransferNFT)     //转让NFT SendNFT
	rpc.RegisterRPC("getnftbyowner", GetNFTByOwner) //查询一个地址的所有NFT
	rpc.RegisterRPC("getnftall", GetNFTAll)         //查询属于本钱包的所有NFT
	rpc.RegisterRPC("getnftid", GetNFTID)           //通过nft_id查询一个NFT

	//多签
	rpc.RegisterRPC("getpublickey", GetPublicKey)                   //获取公钥
	rpc.RegisterRPC("createmultsign", CreateMultsignAddress)        //创建多签集合
	rpc.RegisterRPC("multsignsendtoaddress", MultsignSendToAddress) //多签转账
	rpc.RegisterRPC("getrequestmultsigns", ListRequestMultsigns)    //列表多签
	rpc.RegisterRPC("getrequestmultsign", GetRequestMultsign)       //查询多签
	rpc.RegisterRPC("signmultsign", SignMultsign)                   //多签签名
	rpc.RegisterRPC("multnameinreg", MultNameInReg)                 //多签注册域名
	rpc.RegisterRPC("multnameintransfer", MultNameInTransfer)       //多签转让域名
	rpc.RegisterRPC("multnameinrenew", MultNameInRenew)             //多签续费域名
	rpc.RegisterRPC("multnameinupdate", MultNameInUpdate)           //多签更新域名
	rpc.RegisterRPC("multnameout", MultNameOut)                     //多签注销域名

	// miner部分
	rpc.RegisterRPC("getaddresstx", GetAddressTx)                               //获取地址交易
	rpc.RegisterRPC("getaddresserc20tx", GetAddressErc20Tx)                     //获取ERC20地址交易
	rpc.RegisterRPC("getalltx", GetAllTxV2)                                     //获取所有交易
	rpc.RegisterRPC("getminerblock", GetMinerBlock)                             //获取挖矿交易
	rpc.RegisterRPC("getlightnodedetail", GetLightNodeDetail)                   //获取轻节点详情
	rpc.RegisterRPC("getwitnessnodedetail", GetWitnessNodeDetail)               //获取见证者节点详情
	rpc.RegisterRPC("getcommunitynodedetail", GetCommunityNodeDetail)           //获取社区节点详情
	rpc.RegisterRPC("getcommunitylistforminer", GetCommunityListForMiner)       //获取社区节点列表
	rpc.RegisterRPC("getwitnesslistforminer", GetWitnessesListWithPage)         //获得全网候选见证人和见证人列表 带翻页
	rpc.RegisterRPC("getwitnessinfoforminer", GetWitnessesInfoForMiner)         //获得某个候选见证人和见证人信息
	rpc.RegisterRPC("getwitnesslistv0", GetWitnessesListV0WithPage)             //获取出块见证人列表 带翻页
	rpc.RegisterRPC("getwitnessbackuplistv0", GetWitnessesBackUpListV0WithPage) //获得候选见证人列表 带翻页
	rpc.RegisterRPC("getdepositnumall", GetDepositNumForAll)                    //获取总的节点质押量
	rpc.RegisterRPC("getabiinput", GetAbiInput)                                 //获取批量交易input
	rpc.RegisterRPC("gettxgas", GetTxGas)                                       //根据交易类型获取交易gas 默认普通交易
	rpc.RegisterRPC("getcontractevent", GetContractEvent)                       //根据区块高度和交易hash获取奖励合约事件
	rpc.RegisterRPC("getinternaltx", GetInternalTx)
	rpc.RegisterRPC("getcontracteventfilter", GetContractEventFilter) //根据区块高度范围和合约地址，过滤并获取合约事件
	rpc.RegisterRPC("getexchangerate", GetExchangeRate)               //获取交易所汇率
	rpc.RegisterRPC("getusdrmbrate", GetUsdRmbRate)                   //获取人民币和美元汇率

	rpc.RegisterRPC("getRewardContract", GetRewardContract)

	//网络部分
	rpc.RegisterRPC("getnetworkinfo", NetworkInfo) //获取本节点网络信息

	rpc.RegisterRPC("stopservice", StopService) //关闭服务器
	// "test":        Test,

	rpc.RegisterRPC("createContract", CreateContractByTx)         //--创建合约
	rpc.RegisterRPC("callContract", CallContractByTx)             //--调用合约
	rpc.RegisterRPC("staticCallContract", CallContractStack)      //--本地模拟调用
	rpc.RegisterRPC("getContractInfo", GetContractInfo)           //--获取合约状态
	rpc.RegisterRPC("getErc20Info", GetErc20Info)                 //--获取ERC20合约信息
	rpc.RegisterRPC("getAddrFeature", GetAddrFeature)             //--获取地址特征(1=主链币地址，2=合约地址，3=代币地址),帮助前端判断某个地址特征
	rpc.RegisterRPC("getErc20SumBalance", GetErc20SumBalance)     //--获取多代币地址的总余额
	rpc.RegisterRPC("searchErc20", SearchErc20)                   //--搜索代币
	rpc.RegisterRPC("multibalanceErc20", GetTokenBalances)        //--获取多地址代币余额
	rpc.RegisterRPC("balanceErc20", GetTokenBalance)              //--获取代币余额
	rpc.RegisterRPC("getErc20Value", GetErc20Value)               //--获取收藏的代币(余额）
	rpc.RegisterRPC("accountsErc20Value", GetAccountsErc20Value)  //--节点账户下的代币余额
	rpc.RegisterRPC("setcontractaddrcache", SetContractAddrCache) //合约缓存设置

	rpc.RegisterRPC("addErc20", AddErc20)                     //收藏代币
	rpc.RegisterRPC("delErc20", DelErc20)                     //移除代币
	rpc.RegisterRPC("getErc20", GetErc20)                     //获取收藏的代币
	rpc.RegisterRPC("transferErc20", TransferErc20)           //代币转账
	rpc.RegisterRPC("crossChainTransfer", CrossChainTransfer) //跨链转账
	rpc.RegisterRPC("crossChainWithdraw", CrossChainWithdraw) //跨链提现

	rpc.RegisterRPC("getContractCount", GetContractCount)       //查询有效合约的数量
	rpc.RegisterRPC("getContractSource", GetContractSource)     //获取合约源代码压缩文件
	rpc.RegisterRPC("getContractSourceV2", GetContractSourceV2) //获取合约源代码
	rpc.RegisterRPC("getSpecialContract", GetSpecialContract)   //获取特定类型合约地址
	rpc.RegisterRPC("verifyContract", CheckContract)            //验证合约
	rpc.RegisterRPC("withDrawReward", WithDrawReward)           //提现奖励
	rpc.RegisterRPC("communitydistribute", CommunityDistribute) //社区向轻节点分账
	//rpc.RegisterRPC("getrewardHistory", GetRewardHistory)       //
	rpc.RegisterRPC("receiveToken", ReceiveTCOin) //领取测试币
	//"rechargeToken":         ReChargeTCOin,      //给水龙头充值
	//"delayToken":            DelayTCoin,         //部署测试水龙头
	//ens部分
	//"delayController":   DelayController,   //部署控制器
	//"addController":     AddController,     //添加控制器
	rpc.RegisterRPC("setDomainManger", SetDomainManger) //设置控制器合约为根域名管理员，同时设置解析器
	rpc.RegisterRPC("domainTransfer", DomainTransfer)   //转让
	rpc.RegisterRPC("domainWithDraw", DomainWithDraw)   //提现
	//"requestRegister":   RequestRegister,   //请求注册
	rpc.RegisterRPC("registerDomain", RegisterDomain)     //注册域名
	rpc.RegisterRPC("renewDomain", ReNewDomain)           //续费域名
	rpc.RegisterRPC("getblocksrangeV1", FindBlockRangeV1) //获取一定区块高度范围内的区块新版本
	rpc.RegisterRPC("getblocksrangeV2", FindBlockRangeV2) //获取一定区块高度范围内的区块新版本,列表链上真实交易，浏览器专用接口

	rpc.RegisterRPC("getchainstate", GetChainState) //提供给中心化服务的主链内存状态信息
	//swap
	rpc.RegisterRPC("launchSwap", LaunchSwap)         //发布一个token
	rpc.RegisterRPC("swapPool", SwapPool)             //发布一个token
	rpc.RegisterRPC("accomplishSwap", AccomplishSwap) //发布一个token
	//address
	rpc.RegisterRPC("addressBind", AddressBind)         //地址绑定
	rpc.RegisterRPC("addressTransfer", AddressTransfer) //绑定地址转账
	rpc.RegisterRPC("addressFrozen", AddressFrozen)     //地址资金冻结、解冻
	rpc.RegisterRPC("pushmulttx", PushMultTx)           //推送多签交易

	rpc.RegisterRPC("alladdrbalance", AllAddrBalance) //查询全网余额

	rpc.RegisterRPC("depositfreegas", DepositFreeGas)         //质押免gas费
	rpc.RegisterRPC("listdepositfreegas", ListDepositFreeGas) //列表质押免gas地址
}

const (
	// 限制请求单页数据最大值
	pageSizeLimit = int(10000)
)

/*
关闭服务
*/
func StopService(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	utils.StopService()
	res, err = model.Tojson("success")
	return
}

// 获取基本信息
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "balance": 0,
//	       "testnet": false,
//	       "blocks": 0
//	   }
//	}
func HandleGetInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	value, valuef, valuelockup := mining.FindBalanceValue()

	//tbs := mining.ListTokenInfos()
	//tbVOs := make([]*mining.TokenInfoV0, 0)
	//for _, one := range tbs {
	//	tbVO := mining.ToTokenInfoV0(one)
	//	tbVOs = append(tbVOs, tbVO)
	//}

	currentBlock := toUint64(0)
	startBlock := toUint64(0)
	heightBlock := toUint64(0)
	pulledStates := toUint64(0)
	startBlockTime := toUint64(0)
	syncBlockFinish := false

	chain := mining.GetLongChain()
	if chain != nil {
		currentBlock = chain.GetCurrentBlock()
		startBlock = chain.GetStartingBlock()
		heightBlock = mining.GetHighestBlock()
		pulledStates = chain.GetPulledStates()
		startBlockTime = chain.GetStartBlockTime()
		syncBlockFinish = chain.SyncBlockFinish
	}

	info := Getinfo{
		Netid:          []byte(config.AddrPre),   //
		TotalAmount:    config.Mining_coin_total, //
		Balance:        value,                    //
		BalanceFrozen:  valuef,                   //
		BalanceLockup:  valuelockup,              //
		BalanceVote:    GetBalanceVote(),
		Testnet:        true,                                           //
		Blocks:         currentBlock,                                   //
		Group:          0,                                              //
		StartingBlock:  startBlock,                                     //区块开始高度
		StartBlockTime: startBlockTime,                                 //
		HighestBlock:   heightBlock,                                    //所链接的节点的最高高度
		CurrentBlock:   currentBlock,                                   //已经同步到的区块高度
		PulledStates:   pulledStates,                                   //正在同步的区块高度
		SnapshotHeight: snapshot.Height(),                              //快照高度
		BlockTime:      uint64(config.Mining_block_time.Nanoseconds()), //出块时间 pl time
		LightNode:      config.Mining_light_min,                        //轻节点押金数量
		CommunityNode:  config.Mining_vote,                             //社区节点押金数量
		WitnessNode:    config.Mining_deposit,                          //见证人押金数量
		NameDepositMin: config.Mining_name_deposit_min,                 //
		AddrPre:        config.AddrPre,                                 //
		//TokenInfos:      tbVOs,                                          //
		SyncBlockFinish: syncBlockFinish,
		ContractAddress: precompiled.RewardContract.B58String(),
	}
	res, err := model.Tojson(info)
	return res, err
}

// 详情
type Getinfo struct {
	Netid           []byte                `json:"netid"`            //网络版本号
	TotalAmount     uint64                `json:"TotalAmount"`      //发行总量
	Balance         uint64                `json:"balance"`          //可用余额
	BalanceFrozen   uint64                `json:"BalanceFrozen"`    //冻结的余额
	BalanceLockup   uint64                `json:"BalanceLockup"`    //锁仓的余额
	BalanceVote     uint64                `json:"BalanceVote"`      //当前奖励总金额
	Testnet         bool                  `json:"testnet"`          //是否是测试网络
	Blocks          uint64                `json:"blocks"`           //已经同步到的区块高度
	Group           uint64                `json:"group"`            //区块组高度
	StartingBlock   uint64                `json:"StartingBlock"`    //区块开始高度
	StartBlockTime  uint64                `json:"StartBlockTime"`   //创始区块出块时间
	HighestBlock    uint64                `json:"HighestBlock"`     //所链接的节点的最高高度
	CurrentBlock    uint64                `json:"CurrentBlock"`     //已经同步到的区块高度
	PulledStates    uint64                `json:"PulledStates"`     //正在同步的区块高度
	SnapshotHeight  uint64                `json:"SnapshotHeight"`   //快照高度
	BlockTime       uint64                `json:"BlockTime"`        //出块时间
	LightNode       uint64                `json:"LightNode"`        //轻节点押金数量
	CommunityNode   uint64                `json:"CommunityNode"`    //社区节点押金数量
	WitnessNode     uint64                `json:"WitnessNode"`      //见证人押金数量
	NameDepositMin  uint64                `json:"NameDepositMin"`   //域名押金最少金额
	AddrPre         string                `json:"AddrPre"`          //地址前缀
	TokenInfos      []*mining.TokenInfoV0 `json:"TokenInfos"`       //
	SyncBlockFinish bool                  `json:"SyncBlockFinish"`  //同步区块是否完成
	ContractAddress string                `json:"contract_address"` //奖励合约地址
}

// 获取节点总余额
func HandleGetNodeBalance(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	value, valuef, valuelockup := mining.FindBalanceValue()

	type Getinfo struct {
		Balance       uint64 `json:"balance"`       //可用余额
		BalanceFrozen uint64 `json:"BalanceFrozen"` //冻结的余额
		BalanceLockup uint64 `json:"BalanceLockup"` //锁仓的余额
	}
	info := Getinfo{
		Balance:       value,       //
		BalanceFrozen: valuef,      //
		BalanceLockup: valuelockup, //
	}
	res, err := model.Tojson(info)
	return res, err
}

// 创建新地址
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "address": "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q"
//	   }
//	}
func HandleGetNewAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if !rj.VerifyType("password", "string") {
		res, err = model.Errcode(model.TypeWrong, "password")
		return
	}
	password, ok := rj.Get("password")
	if !ok {
		res, err = model.Errcode(model.NoField, "password")
		return
	}

	name, _, _ := getParamString(rj, "name")

	var addr crypto.AddressCoin
	if name == "" {
		addr, err = config.Area.Keystore.GetNewAddr(password.(string), password.(string))
	} else {
		addr, err = config.Area.Keystore.GetNewAddrByName(name, password.(string), password.(string))
	}
	if err != nil {
		if err.Error() == config.ERROR_wallet_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, _ = model.Errcode(SystemError)
		return
	}
	getnewadress := model.GetNewAddress{Address: addr.B58String()}
	res, err = model.Tojson(getnewadress)
	return
}

type AccountVO struct {
	Index               int      //排序
	Name                string   //名称
	AddrCoin            string   //收款地址
	MainAddrCoin        string   //主地址收款地址
	SubAddrCoins        []string //从地址收款地址
	Value               uint64   //可用余额
	ValueFrozen         uint64   //冻结余额
	ValueLockup         uint64   //
	BalanceVote         uint64   `json:"BalanceVote"` //当前奖励总金额
	AddressFrozenStatus bool     //地址绑定冻结状态
	Type                int      //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

type AccountV1 struct {
	Name                 string   //名称
	AddrCoin             string   //收款地址
	MainAddrCoin         string   //主地址收款地址
	SubAddrCoins         []string //从地址收款地址
	Value                uint64   //可用余额
	ValueFrozen          uint64   //冻结余额
	ValueLockup          uint64   //
	BalanceVote          uint64   `json:"BalanceVote"` //当前奖励总金额
	ValueFrozenWitness   uint64   //见证人节点冻结奖励
	ValueFrozenCommunity uint64   //社区节点冻结奖励
	ValueFrozenLight     uint64   //轻节点冻结奖励
	DepositIn            uint64   //质押数量
	AddressFrozenStatus  bool     //地址绑定冻结状态
	Type                 int      //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

// 地址列表
func HandleListAccounts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
	vos := make([]AccountVO, 0)
	list := config.Area.Keystore.GetAddr()
	total := len(config.Area.Keystore.GetAddr())
	start := (pageInt - 1) * pageSizeInt
	end := start + pageSizeInt
	if start > total {
		res, err = model.Tojson(vos)
		return res, err
	}
	if end > total {
		end = total
	}
	// basMap := make(map[string]uint64)      //可用余额
	// fbasMap := make(map[string]uint64)     //冻结的余额
	// baLockupMap := make(map[string]uint64) //锁仓的余额
	// if tokenidStr == "" {
	// 	basMap, fbasMap, baLockupMap = mining.GetBalanceAllAddrs()
	// 	// bas, fbas, baLockup := mining.GetBalanceAllItems()
	// 	// for _, one := range bas {
	// 	// 	value, _ := basMap[utils.Bytes2string(*one.Addr)]
	// 	// 	basMap[utils.Bytes2string(*one.Addr)] = value + one.Value
	// 	// }
	// 	// for _, one := range fbas {
	// 	// 	value, _ := fbasMap[utils.Bytes2string(*one.Addr)]
	// 	// 	fbasMap[utils.Bytes2string(*one.Addr)] = value + one.Value
	// 	// }
	// 	// for _, one := range baLockup {
	// 	// 	value, _ := baLockupMap[utils.Bytes2string(*one.Addr)]
	// 	// 	baLockupMap[utils.Bytes2string(*one.Addr)] = value + one.Value
	// 	// }
	// } else {
	// 	tokenidbs, e := hex.DecodeString(tokenidStr)
	// 	if e != nil {
	// 		res, err = model.Errcode(model.TypeWrong, "token_id")
	// 		return
	// 	}
	// 	basMap, fbasMap, baLockupMap = token.FindTokenBalanceForTxid(tokenidbs)
	// }

	for i, val := range list[start:end] {
		ba, fba, baLockup := mining.GetBalanceForAddrSelf(val.Addr)

		// ba, _ := basMap[utils.Bytes2string(val.Addr)]
		// fba, _ := fbasMap[utils.Bytes2string(val.Addr)]
		// baLockup, _ := baLockupMap[utils.Bytes2string(val.Addr)]
		addrType := mining.GetAddrState(val.Addr)
		//特殊处理因黑名单踢出的见证人
		//首地址
		if start == 0 && i == 0 && addrType == 4 {
			if depositVal := mining.GetDepositWitnessAddr(&val.Addr); depositVal > 0 {
				addrType = 1
			}
		}

		mainAddr := ""
		if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(val.Addr)); err == nil {
			a := crypto.AddressCoin(bs)
			mainAddr = a.B58String()
		}

		subAddrs := []string{}
		if pairs, err := db.LevelDB.HGetAll(val.Addr); err == nil {
			for _, pair := range pairs {
				a := crypto.AddressCoin(pair.Field)
				subAddrs = append(subAddrs, a.B58String())
			}
		}

		vo := AccountVO{
			Index:               i + start,
			Name:                GetNickName(val.Addr),
			AddrCoin:            val.GetAddrStr(),
			MainAddrCoin:        mainAddr,
			SubAddrCoins:        subAddrs,
			Type:                addrType,
			Value:               ba,       //可用余额
			ValueFrozen:         fba,      //冻结余额
			ValueLockup:         baLockup, //
			BalanceVote:         GetBalanceVote(val.Addr),
			AddressFrozenStatus: mining.CheckAddressFrozenStatus(val.Addr),
		}
		vos = append(vos, vo)
	}
	res, err = model.Tojson(vos)
	return res, err

	// vos := make([]AccountVO, 0)

	// // list := make(map[string]uint64)
	// addr := keystore.GetAddr()
	// for i, val := range addr {
	// 	// list[val.B58String()] = mining.GetBalanceForAddr(&val)
	// 	vo := AccountVO{
	// 		Index:    i,
	// 		AddrCoin: val.AddrStr,
	// 		// Value:    mining.GetBalanceForAddr(&val),
	// 		Type: mining.GetAddrState(val.Addr),
	// 	}
	// 	vo.Value, vo.ValueFrozen = mining.GetBalanceForAddr(val.AddrStr)
	// 	vos = append(vos, vo)
	// }
	// res, err = model.Tojson(vos)
	// return res, err
}

// 地址列表
func HandleListMultAccounts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vos := []*AccountVO{}
	kvmap := db.LevelDB.WrapLevelDBPrekeyRange(config.DBKEY_multsign_set)
	keys := []string{}
	for key, _ := range kvmap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value, ok := kvmap[key]
		if !ok {
			continue
		}
		pukSet := &go_protos.MultsignSet{}
		if err := pukSet.Unmarshal(value); err == nil {
			isSelf := false
			for _, puk := range pukSet.Puks {
				if _, ok := config.Area.Keystore.FindPuk(puk); ok {
					isSelf = true
					break
				}
			}

			if isSelf {
				multAddr := crypto.AddressCoin(pukSet.MultAddress)
				mainAddr := ""
				if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(multAddr)); err == nil {
					a := crypto.AddressCoin(bs)
					mainAddr = a.B58String()
				}

				subAddrs := []string{}
				if pairs, err := db.LevelDB.HGetAll(multAddr); err == nil {
					for _, pair := range pairs {
						a := crypto.AddressCoin(pair.Field)
						subAddrs = append(subAddrs, a.B58String())
					}
				}
				ba, fba, baLockup := mining.GetBalanceForAddrSelf(multAddr)
				vo := &AccountVO{
					Name:         string(pukSet.Name),
					AddrCoin:     multAddr.B58String(),
					MainAddrCoin: mainAddr,
					SubAddrCoins: subAddrs,
					Type:         4,
					Value:        ba,       //可用余额
					ValueFrozen:  fba,      //冻结余额
					ValueLockup:  baLockup, //
				}
				vos = append(vos, vo)
			}
		}
	}

	res, err = model.Tojson(vos)
	return res, err
}

// 地址详情
func HandleInfoAccount(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	tokenId := []byte{}
	tokenidItr, ok := rj.Get("token_id")
	if ok {
		if !rj.VerifyType("token_id", "string") {
			res, err = model.Errcode(model.TypeWrong, "token_id")
			return
		}
		tokenidStr := tokenidItr.(string)
		tokenId, err = hex.DecodeString(tokenidStr)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, "decode error")
			return
		}
	}

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
	ba, fba, baLockup := toUint64(0), toUint64(0), toUint64(0)
	if len(tokenId) == 0 {
		ba, fba, baLockup = mining.GetBalanceForAddrSelf(addr)
	} else {
		ba, fba, baLockup = mining.GetTokenNotSpendAndLockedBalance(tokenId, addr)
	}

	depositIn := toUint64(0)
	//社区地址
	ty := mining.GetAddrState(addr)
	//特殊处理因黑名单踢出的见证人
	//首地址
	if ty == 4 && bytes.Equal(config.Area.Keystore.GetCoinbase().Addr, addr) {
		if depositVal := mining.GetDepositWitnessAddr(&addr); depositVal > 0 {
			ty = 1
		}
	}

	if len(tokenId) != 0 {
		ty = 4
	}

	balanceMgr := mining.GetLongChain().GetBalance()

	var wValue, cValue, lValue uint64 = 0, 0, 0
	switch ty {
	case 1:
		depositIn = config.Mining_deposit
		witRewardPools := balanceMgr.GetWitnessRewardPool()
		if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(addr)); ok {
			wreward, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(addr, witRewardPool.(*big.Int))
			wValue = wreward.Uint64()
		}

	case 2:
		depositIn = config.Mining_vote
		cAddr := addr
		creward, _ := balanceMgr.CalculateCommunityRewardAndLightReward(cAddr)
		cValue = creward.Uint64()
		//if itemInfo := balanceMgr.GetDepositCommunity(&cAddr); itemInfo != nil {
		//	wAddr := itemInfo.WitnessAddr
		//	witRewardPools := balanceMgr.GetWitnessRewardPool()
		//	if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(wAddr)); ok {
		//		wreward, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(wAddr, witRewardPool.(*big.Int))
		//		wValue = wreward.Uint64()
		//	}

		//	creward, _ := balanceMgr.CalculateCommunityRewardAndLightReward(cAddr)
		//	cValue = creward.Uint64()
		//}

	case 3:
		depositIn = config.Mining_light_min
		lAddr := addr
		if itemInfo := balanceMgr.GetDepositVote(&lAddr); itemInfo != nil {
			cAddr := itemInfo.WitnessAddr
			_, lightRewards := balanceMgr.CalculateCommunityRewardAndLightReward(cAddr)
			if v, ok := lightRewards[utils.Bytes2string(lAddr)]; ok {
				lValue = v.Uint64()
			}
		}
		//if itemInfo := balanceMgr.GetDepositVote(&lAddr); itemInfo != nil {
		//	cAddr := itemInfo.WitnessAddr
		//	if itemInfo := balanceMgr.GetDepositCommunity(&cAddr); itemInfo != nil {
		//		wAddr := itemInfo.WitnessAddr
		//		witRewardPools := balanceMgr.GetWitnessRewardPool()
		//		if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(wAddr)); ok {
		//			wreward, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(wAddr, witRewardPool.(*big.Int))
		//			wValue = wreward.Uint64()
		//		}

		//		creward, lightRewards := balanceMgr.CalculateCommunityRewardAndLightReward(cAddr)
		//		cValue = creward.Uint64()
		//		if v, ok := lightRewards[utils.Bytes2string(lAddr)]; ok {
		//			lValue = v.Uint64()
		//		}
		//	}
		//}
		//light := precompiled.GetLightList([]crypto.AddressCoin{addr})
		//if len(light) > 0 && light[0].Score.Uint64() != 0 {
		//	if light[0].C != evmutils.ZeroAddress {
		//		cAddr = evmutils.AddressToAddressCoin(light[0].C.Bytes())
		//	}
		//}
	}

	//if cAddr != nil {
	//	cRewardPool := precompiled.GetCommunityRewardPool(cAddr)
	//	if cRewardPool.Cmp(big.NewInt(0)) > 0 {
	//		cRate, err := precompiled.GetRewardRatio(cAddr)
	//		if err != nil {
	//			return model.Errcode(SystemError, "address")
	//		}

	//		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
	//		cBigValue := new(big.Int).Sub(cRewardPool, lTotal)
	//		cValue = cBigValue.Uint64()
	//		if depositIn == config.Mining_light_min {
	//			community := precompiled.GetCommunityList([]crypto.AddressCoin{cAddr})
	//			if len(community) == 0 {
	//				return model.Errcode(SystemError, "address")
	//			}
	//			//当前轻节点身份，不需要查出社区身份时候的未体现余额，新需求取消质押会提现所有奖励池
	//			cValue = 0
	//			v := new(big.Int).Mul(new(big.Int).SetUint64(balanceMgr.GetDepositVote(&addr).Value), new(big.Int).SetInt64(1e8))
	//			ratio := new(big.Int).Quo(v, community[0].Vote)
	//			lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
	//			lValue = lBigValue.Uint64()
	//		}
	//	}
	//}

	mainAddr := ""
	if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(addr)); err == nil {
		a := crypto.AddressCoin(bs)
		mainAddr = a.B58String()
	}

	subAddrs := []string{}
	if pairs, err := db.LevelDB.HGetAll(addr); err == nil {
		for _, pair := range pairs {
			a := crypto.AddressCoin(pair.Field)
			subAddrs = append(subAddrs, a.B58String())
		}
	}

	vo := AccountV1{
		Name:                 GetNickName(addr),
		AddrCoin:             addrStr,
		MainAddrCoin:         mainAddr,
		SubAddrCoins:         subAddrs,
		Value:                ba,       //可用余额
		ValueFrozen:          fba,      //冻结余额
		ValueLockup:          baLockup, //
		BalanceVote:          wValue + cValue + lValue,
		ValueFrozenWitness:   wValue,
		ValueFrozenCommunity: cValue,
		ValueFrozenLight:     lValue,
		DepositIn:            depositIn,
		AddressFrozenStatus:  mining.CheckAddressFrozenStatus(addr),
		Type:                 ty,
	}

	return model.Tojson(vo)
}

// 获取某一帐号余额
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "Balance": 0
//	   }
//	}
func HandleGetAccount(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	value, valueFrozen, _ := mining.GetBalanceForAddrSelf(addrCoin)
	// fmt.Println(addr)
	getaccount := model.GetAccount{
		Balance:       value,
		BalanceFrozen: valueFrozen,
	}
	res, err = model.Tojson(getaccount)
	return
}

// 验证地址
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "IsVerify": true,
//	       "IsMine": false,
//	       "IsType": 1,
//	       "Version": 0,
//	       "ExpVersion": 0,
//	       }
//	   }
//	}
func HandleValidateAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			res, err = model.Tojson(false)
			return
		}
	}()
	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	res, err = model.Tojson(ok)
	return
}

/*
转账
为了提高性能，接口不对地址正确性验证，前缀正确性验证，请提前验证了再请求
*/
func SendToAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// fmt.Println("++++++++++++++++++++\n时间开始")
	// start := config.TimeNow()
	// return nil, errors.New("error")
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

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

	// var change crypto.AddressCoin
	// changeItr, ok := rj.Get("changeaddress")
	// if ok {
	// 	changeStr := changeItr.(string)
	// 	if changeStr != "" {
	// 		change = crypto.AddressFromB58String(changeStr)
	// 		//判断地址前缀是否正确
	// 		if !crypto.ValidAddr(config.AddrPre, change) {
	// 			res, err = model.Errcode(ContentIncorrectFormat, "changeaddress")
	// 			return
	// 		}
	// 	}
	// }

	// engine.Log.Info("交易接口解析地址消耗 111 %s", config.TimeNow().Sub(start))

	addrItr, ok = rj.Get("address")
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

	// engine.Log.Info("交易接口解析地址消耗 222 %s", config.TimeNow().Sub(start))

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// frozenHeight = toUint64(config.TimeNow().Unix() + (20 * 5))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	if e := mining.CheckTxPayFreeGasWithParams(config.Wallet_tx_type_pay, src, amount, gas, comment); e != nil {
		res, err = model.Errcode(GasTooLittle, "gas too little")
		return
	}

	//temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
	//temp = new(big.Int).Div(temp, big.NewInt(1024))
	//if gas < temp.Uint64() {
	//	res, err = model.Errcode(GasTooLittle, "gas")
	//	return
	//}

	// fmt.Println("转账到地址", addr, amount, pwd, comment)

	// dst, e := utils.FromB58String(addr)
	// if err != nil {
	// 	err = e
	// 	res, _ = model.Errcode(5003, "error")
	// 	return
	// }

	//查询余额是否足够
	// value, _ := mining.GetBalance()
	// if amount > value {
	// 	res, err = model.Errcode(BalanceNotEnough)
	// 	return
	// }

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas)
	if total < amount+gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
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

	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(src.B58String(), domain, dst.B58String(), new(big.Int).SetUint64(domainType)) {
			return model.Errcode(model.Nomarl, "domain name resolution failed")
		}
	}

	// engine.Log.Info("交易接口解析地址消耗 333 %s", config.TimeNow().Sub(start))

	// engine.Log.Info("创建转账交易错误 00000000000000")
	// startTime := config.TimeNow()
	// var txpay mining.TxItr = nil
	// err = errors.New("error")
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

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)
	// engine.Log.Info("交易接口共消耗 %s", config.TimeNow().Sub(start))
	// res, err = model.Tojson("success")
	return
}

// 临时查询接口
func GetPublicKey(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	address, res, err := getParamString(rj, "address")
	if err != nil {
		return res, err
	}

	addr := crypto.AddressFromB58String(address)
	puk, ok := config.Area.Keystore.GetPukByAddr(addr)
	if !ok {
		return model.Errcode(model.Nomarl, "not found public key")
	}

	data := make(map[string]interface{})
	data["puk"] = string(base58.Encode(puk))
	data["address"] = addr.B58String()
	res, err = model.Tojson(data)
	return
}

/*
创建多签
*/
func CreateMultsignAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

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

	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}

	comment, _, _ := getParamString(rj, "comment")
	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}

	pukStrs, errcode := getArrayStrParams(rj, "puks")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "puks")
		return
	}
	if len(pukStrs) < config.MultsignSet_addrs_puk_min || len(pukStrs) > config.MultsignSet_addrs_puk_max {
		res, err = model.Errcode(model.Nomarl, fmt.Sprintf("limit number %d~%d", config.MultsignSet_addrs_puk_min, config.MultsignSet_addrs_puk_max))
		return
	}
	puks := make([][]byte, len(pukStrs))
	for i, one := range pukStrs {
		b := base58.Decode(one)
		if len(b) != ed25519.PublicKeySize {
			res, err = model.Errcode(model.Nomarl, "invalid public key")
			return res, err
		}
		puks[i] = b
	}

	mtx, err := mining.BuildCreateMultsignAddrTx(src, gas, pwd, comment, puks...)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson(mtx.GetVOJSON())
	return
}

/*
多签转账
*/
func MultsignSendToAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	multAddrItr, ok := rj.Get("multaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "multaddress")
		return
	}
	multAddrStr := multAddrItr.(string)
	multAddr := crypto.AddressFromB58String(multAddrStr)
	if !crypto.ValidAddr(config.AddrPre, multAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "multaddress")
		return
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

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	// frozenHeight = toUint64(config.TimeNow().Unix() + (20 * 5))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	if e := mining.CheckTxPayFreeGasWithParams(config.Wallet_tx_type_multsign_pay, multAddr, amount, gas, comment); e != nil {
		res, err = model.Errcode(GasTooLittle, "gas too little")
		return
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&multAddr, amount+gas)
	if total < amount+gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
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

	//验证domain
	if domain != "" {
		if !ens.CheckDomainResolve(multAddr.B58String(), domain, dst.B58String(), new(big.Int).SetUint64(domainType)) {
			return model.Errcode(model.Nomarl, "domain name resolution failed")
		}
	}

	txpay, err := mining.BuildRequestMultsignPayTx(multAddr, dst, amount, gas, frozenHeight, pwd, comment)
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

	res, err = model.Tojson(txpay.GetVOJSON())
	// engine.Log.Info("交易接口共消耗 %s", config.TimeNow().Sub(start))
	// res, err = model.Tojson("success")
	return
}

/*
获取多签交易
*/
func ListRequestMultsigns(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	txType, res, err := getParamUint64(rj, "tx_type")
	if err != nil {
		return res, err
	}

	preId := make([]byte, 8)
	binary.PutUvarint(preId, txType)
	kvmap := db.LevelDB.WrapLevelDBPrekeyRange(append(config.DBKEY_multsign_request_tx, preId...))
	keys := []string{}
	for key, _ := range kvmap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := []interface{}{}
	for _, key := range keys {
		value, ok := kvmap[key]
		if !ok {
			continue
		}

		txItr, err := mining.ParseTxBaseProto(txType, &value)
		if err != nil {
			continue
		}

		mtx, ok := txItr.(mining.DefaultMultTx)
		if !ok {
			continue
		}

		//判断所有者
		for _, vin := range mtx.GetMultVins() {
			if _, ok := config.Area.Keystore.FindPuk(vin.Puk); ok {
				data = append(data, txItr.GetVOJSON())
				break
			}
		}
	}

	res, err = model.Tojson(data)
	return
}

/*
获取多签交易
*/
func GetRequestMultsign(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	txidStr, res, err := getParamString(rj, "txid")
	if err != nil {
		return res, err
	}
	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	key := config.BuildMultsignRequestTx(txid)
	txBs, err := db.LevelDB.Find(key)
	if err == nil && len(*txBs) != 0 {
		//聚合签名
		txItr, err := mining.ParseTxBaseProto(0, txBs)
		if err != nil {
			return model.Errcode(model.Nomarl, err.Error())
		}

		return model.Tojson(txItr.GetVOJSON())
	}

	return model.Errcode(model.NotExist, "not found")
}

/*
多签签名
*/
func SignMultsign(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	txidStr, res, err := getParamString(rj, "txid")
	if err != nil {
		return res, err
	}

	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		return model.Errcode(model.TypeWrong, "decode txid")
	}

	addressStr, res, err := getParamString(rj, "address")
	if err != nil {
		return res, err
	}

	address := crypto.AddressFromB58String(addressStr)

	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}

	//signStr, res, err := getParamString(rj, "sign")
	//if err != nil {
	//	return res, err
	//}

	//sign, err := hex.DecodeString(signStr)
	//if err != nil {
	//	return model.Errcode(model.TypeWrong, "decode sign")
	//}
	puk, ok := config.Area.Keystore.GetPukByAddr(address)
	if !ok {
		return model.Errcode(model.Nomarl, "public not found")
	}
	_, prik, err := config.Area.Keystore.GetKeyByPuk(puk, pwd)
	if err != nil {
		return model.Errcode(model.Nomarl, "invalid key")
	}

	txbs, err := db.LevelDB.Find(config.BuildMultsignRequestTx(txid))
	if err != nil || len(*txbs) == 0 {
		return model.Errcode(model.NotExist, "not exist")
	}

	txClass := mining.ParseTxClass(txid)
	txItr, err := mining.ParseTxBaseProto(txClass, txbs)

	//是否实现多签
	mtx, ok := txItr.(mining.DefaultMultTx)
	if err != nil {
		return model.Errcode(model.Nomarl, "not found tx")
	}
	found := false
	for i, vin := range mtx.GetMultVins() {
		if crypto.CheckPukAddr(config.AddrPre, vin.Puk, address) {
			bs := txItr.GetSign(&prik, uint64(i))
			mtx.GetMultVins()[i].Sign = *bs
			found = true
		}
	}

	if !found {
		return model.Errcode(model.Nomarl, "not found puk/address")
	}

	//处理多签请求
	if err := mtx.RequestMultTx(txItr); err != nil {
		engine.Log.Warn("Multsign Request: %s %s", hex.EncodeToString(*txItr.GetHash()), err.Error())
		return nil, err
	}

	res, err = model.Tojson(txItr.GetVOJSON())
	return
}

/*
多人转账
*/
type PayNumber struct {
	Address      string `json:"address"`       //转账地址
	Amount       uint64 `json:"amount"`        //转账金额
	FrozenHeight uint64 `json:"frozen_height"` //冻结高度
}

/*
给多人转账
*/
func SendToAddressmore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// fmt.Println("++++++++++++++++++++\n时间开始")
	// start := config.TimeNow()

	var src crypto.AddressCoin
	srcAddrStr := ""
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcAddrStr = addrItr.(string)
		if srcAddrStr != "" {
			src = crypto.AddressFromB58String(srcAddrStr)
			//判断地址前缀是否正确
			// if !crypto.ValidAddr(config.AddrPre, src) {
			// 	res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			//判断地址是否包含在keystone里面
			// if !keystore.FindAddress(src) {
			// 	res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrItr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	addrs := make([]PayNumber, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&addrs)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	//给多人转账，可是没有地址
	if len(addrs) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	amount := toUint64(0)

	addr := make([]mining.PayNumber, 0)
	for _, one := range addrs {
		dst := crypto.AddressFromB58String(one.Address)
		//验证地址前缀
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(ContentIncorrectFormat, "addresses")
			return
		}
		pnOne := mining.PayNumber{
			Address:      dst,              //转账地址
			Amount:       one.Amount,       //转账金额
			FrozenHeight: one.FrozenHeight, //
		}
		// pnOne.FrozenHeight = toUint64(config.TimeNow().Unix() + (20 * 5))
		addr = append(addr, pnOne)
		amount += one.Amount
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}

	if e := mining.CheckTxPayFreeGasWithParams(config.Wallet_tx_type_pay, src, amount, gas, comment); e != nil {
		res, err = model.Errcode(GasTooLittle, "gas too little")
		return
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount+gas > value {
		res, err = model.Errcode(BalanceNotEnough)
		// engine.Log.Warn("余额不足 111111 %d %d", amount+gas, value)
		return
	}

	// engine.Log.Info("创建转账交易错误 00000000000000")
	txpay, err := mining.SendToMoreAddress(&src, addr, gas, pwd, comment)
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

	//验证域名
	if !txpay.CheckDomain() {
		return model.Errcode(model.Nomarl, "domain name resolution failed")
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)
	return
}

// 缴纳押金，成为见证人
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//	       "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//	       "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//	   }
//	}
func DepositIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(GasTooLittle, "gas")
		return
	}
	//比例
	rateItr, ok := rj.Get("rate")
	if !ok {
		res, err = model.Errcode(model.NoField, "奖励的分配比例未设定")
		return
	}
	rate := uint16(rateItr.(float64))
	if rate > 100 {
		res, err = model.Errcode(DistributeRatioTooBig, "分配比例不能大于100")
		return
	}
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= config.Wallet_vote_start_height {
		res, err = model.Errcode(VoteNotOpen)
		return
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount > value {
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	err = mining.DepositIn(amount, gas, pwd, payload, rate)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	config.SubmitDepositin = true
	return
}

// 退还押金
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//	       "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//	       "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//	   }
//	}
func DepositOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr := ""
	addrItr, ok := rj.Get("address")
	if ok {
		addr = addrItr.(string)

	}

	if addr != "" {
		dst := crypto.AddressFromB58String(addr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(ContentIncorrectFormat, "address")
			return
		}
	}

	amount := toUint64(0)
	amountItr, ok := rj.Get("amount")
	if ok {
		amount = toUint64(amountItr.(float64))
		if amount < 0 {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	engine.Log.Info("address:%s amount:%d gas:%d", addr, amount, gas)

	err = mining.DepositOut(addr, amount, gas, pwd)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

// 缴纳押金，成为见证人
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//	       "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//	       "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//	   }
//	}
func voteIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	addr := ""
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr = addrItr.(string)
	if addr == "" {
		res, err = model.Errcode(model.NoField, "address")
		return
	}

	var witnessAddr crypto.AddressCoin
	witnessAddrItr, ok := rj.Get("witness")
	if ok {
		// res, err = model.Errcode(5002, "witness")
		// return
		witnessStr := witnessAddrItr.(string)

		witnessAddr = crypto.AddressFromB58String(witnessStr)

		if witnessStr != "" {
			dst := crypto.AddressFromB58String(witnessStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(ContentIncorrectFormat, "witness")
				return
			}
		}
	}
	var rate uint16
	switch voteType {
	case mining.VOTE_TYPE_community:
		rateItr, ok := rj.Get("rate")
		if !ok {
			res, err = model.Errcode(model.NoField, "rate")
			return
		}
		rate = uint16(rateItr.(float64))
		if rate > 100 {
			res, err = model.Errcode(DistributeRatioTooBig, "分配比例不能大于100")
			return
		}
	case mining.VOTE_TYPE_vote:
	case mining.VOTE_TYPE_light:
		witnessAddr = nil
	default:
		res, err = model.Errcode(ParamError, "votetype")
		return
	}
	var dst crypto.AddressCoin
	dst = crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
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
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= config.Wallet_vote_start_height {
		res, err = model.Errcode(VoteNotOpen)
		return
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount > value {
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	err = mining.VoteIn(voteType, rate, witnessAddr, dst, amount, gas, pwd, payload, gasPrice)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

// 退还押金
//
//	{
//	   "jsonrpc": "2.0",
//	   "code": 2000,
//	   "result": {
//	       "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//	       "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//	       "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//	   }
//	}
func voteOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// var txid []byte
	// txidItr, ok := rj.Get("txid")
	// if ok {
	// 	txidStr := txidItr.(string)
	// 	txid, _ = hex.DecodeString(txidStr)
	// }

	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	addrStr := ""
	var addr crypto.AddressCoin
	addrItr, ok := rj.Get("address")
	if ok {
		addrStr = addrItr.(string)
	}
	if addrStr != "" {
		addr = crypto.AddressFromB58String(addrStr)
		if !crypto.ValidAddr(config.AddrPre, addr) {
			res, err = model.Errcode(ContentIncorrectFormat, "address")
			return
		}
	}

	switch voteType {
	case mining.VOTE_TYPE_community:
	case mining.VOTE_TYPE_vote:
	case mining.VOTE_TYPE_light:
		//判断是否取消所有投票，未取消所有投票之前，不能取消轻节点
		di := mining.GetLongChain().GetBalance().GetDepositVote(&addr)
		if di != nil {
			res, err = model.Errcode(NotDepositOutLight)
			return
		}
	default:
		res, err = model.Errcode(ParamError, "votetype")
		return
	}

	//退押金，amount可以为0，取消投票，必须要带取消质押金额
	amountItr, ok := rj.Get("amount")
	var amount uint64
	if !ok && voteType == mining.VOTE_TYPE_vote {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	if ok {
		amount = toUint64(amountItr.(float64))
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
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	// witnessAddrItr, ok := rj.Get("witness")
	// if !ok {
	// 	res, err = model.Errcode(model.NoField, "witness")
	// 	return
	// }
	// witnessStr := witnessAddrItr.(string)
	// witnessAddr := crypto.AddressFromB58String(witnessStr)

	// var witnessAddr crypto.AddressCoin
	// witnessAddrItr, ok := rj.Get("witness")
	// if ok {
	// 	witnessStr := witnessAddrItr.(string)
	// 	witnessAddr = crypto.AddressFromB58String(witnessStr)
	// }

	err = mining.VoteOut(voteType, addr, amount, gas, pwd, payload, gasPrice)
	if err != nil {
		// engine.Log.Info("--------------- 取消投票错误" + err.Error())

		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		//余额不足
		if err.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
		//投票已经存在
		if err.Error() == config.ERROR_vote_exist.Error() {
			res, err = model.Errcode(VoteExist)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

// /*
// 	缴纳轻节点押金
// */
// func depositInLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(5002, "amount")
// 		return
// 	}
// 	amount := toUint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := toUint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	var witnessAddr crypto.AddressCoin
// 	// witnessAddrItr, ok := rj.Get("witness")
// 	// if ok {
// 	// 	// res, err = model.Errcode(5002, "witness")
// 	// 	// return
// 	// 	witnessStr := witnessAddrItr.(string)

// 	// 	witnessAddr = crypto.AddressFromB58String(witnessStr)
// 	// }

// 	//查询余额是否足够
// 	if amount > mining.GetBalance() {
// 		res, err = model.Errcode(model.Nomarl, "余额不足")
// 		return
// 	}

// 	err = mining.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点退还押金
// */
// func depositOutLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	txid := ""
// 	txidItr, ok := rj.Get("txid")
// 	if ok {
// 		txid = txidItr.(string)
// 	}

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := toUint64(amountItr.(float64))

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

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOut(&witnessAddr, txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点给社区节点投票
// */
// func voteInLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(5002, "amount")
// 		return
// 	}
// 	amount := toUint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := toUint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	var witnessAddr crypto.AddressCoin
// 	witnessAddrItr, ok := rj.Get("witness")
// 	if ok {
// 		// res, err = model.Errcode(5002, "witness")
// 		// return
// 		witnessStr := witnessAddrItr.(string)

// 		witnessAddr = crypto.AddressFromB58String(witnessStr)
// 	}

// 	//查询余额是否足够
// 	if amount > mining.GetBalance() {
// 		res, err = model.Errcode(model.Nomarl, "余额不足")
// 		return
// 	}

// 	err = mining.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点给社区节点投票 退还押金
// */
// func voteOutLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	txid := ""
// 	txidItr, ok := rj.Get("txid")
// 	if ok {
// 		txid = txidItr.(string)
// 	}

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := toUint64(amountItr.(float64))

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

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOut(&witnessAddr, txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// //退还押金
// //{
// //    "jsonrpc": "2.0",
// //    "code": 2000,
// //    "result": {
// //        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
// //        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
// //        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
// //    }
// //}
// func voteOutOne(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	txidItr, ok := rj.Get("txid")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "txid")
// 		return
// 	}
// 	txid := txidItr.(string)

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := toUint64(amountItr.(float64))

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

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOutOne(txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

/*
修改钱包支付密码
*/
func UpdatePwd(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	engine.SetLogPath(filepath.Join("./", "log11.txt"))

	oldpwdItr, ok := rj.Get("oldpwd")
	if !ok {
		res, err = model.Errcode(5002, "oldpwd")
		return
	}
	oldpwd := oldpwdItr.(string)

	pwdItr, ok := rj.Get("newpwd")
	if !ok {
		res, err = model.Errcode(5002, "newpwd")
		return
	}
	pwd := pwdItr.(string)

	//钱包地址
	ok, err = config.Area.Keystore.UpdatePwd(oldpwd, pwd)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if !ok {
		//密码错误
		res, err = model.Errcode(model.FailPwd, errors.New("password fail").Error())
		return
	}
	// 修改0地址密码
	//addr := config.Area.Keystore.GetCoinbase().Addr.B58String()

	//res, err = model.Errcode(model.FailPwd, errors.New(fmt.Sprintf("GetCoinbase addr:%s,oldpwd:%s,pwd:%s", addr, oldpwd, pwd)).Error())
	//return

	// 修改0地址密码
	addr := config.Area.Keystore.GetCoinbase().Addr.B58String()
	ok, err = config.Area.Keystore.UpdateAddrPwd(addr, oldpwd, pwd)
	if err != nil {
		config.Area.Keystore.UpdatePwd(pwd, oldpwd)
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if !ok {
		config.Area.Keystore.UpdatePwd(pwd, oldpwd)
		//密码错误
		res, err = model.Errcode(model.FailPwd, errors.New("password fail").Error())
		return
	}

	//网络地址
	config.Area.Keystore.UpdateNetAddrPwd(oldpwd, pwd)
	if err != nil {
		config.Area.Keystore.UpdatePwd(pwd, oldpwd)
		config.Area.Keystore.UpdateAddrPwd(addr, pwd, oldpwd)
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if !ok {
		config.Area.Keystore.UpdatePwd(pwd, oldpwd)
		config.Area.Keystore.UpdateAddrPwd(addr, pwd, oldpwd)
		//密码错误
		res, err = model.Errcode(model.FailPwd, errors.New("password fail").Error())
		return
	}

	config.Wallet_keystore_default_pwd = pwd

	res, err = model.Tojson("success")
	return
}

/*
创建钱包
*/
func CreateKeystore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	// randomItr, ok := rj.Get("random")
	// if !ok {
	// 	res, err = model.Errcode(model.NoField, "random")
	// 	return
	// }

	// randomItrs := randomItr.([]interface{})
	// buf := bytes.NewBuffer(nil)
	// for _, one := range randomItrs {
	// 	onePoint := uint16(one.(float64))
	// 	_, e := buf.Write(utils.Uint16ToBytes(onePoint))
	// 	if e != nil {
	// 		res, err = model.Errcode(model.Nomarl, e.Error())
	// 		return
	// 	}
	// }
	// if buf.Len() != 4000 {
	// 	//随机数长度不等于2000
	// 	res, err = model.Errcode(model.Nomarl, "Random number length not equal to 2000")
	// 	return
	// }

	// rand1 := buf.Bytes()[:2000]
	// rand2 := buf.Bytes()[2000:]

	// pwdItr, ok := rj.Get("pwd")
	// if !ok {
	// 	res, err = model.Errcode(model.NoField, "pwd")
	// 	return
	// }
	// pwd := pwdItr.(string)

	// err = config.Area.Keystore.CreateKeystoreRand(filepath.Join(config.Path_configDir, config.Core_keystore), config.AddrPre, buf.Bytes(), rand1, rand2, pwd)
	// if err != nil {
	// 	res, err = model.Errcode(model.Nomarl, err.Error())
	// 	return
	// }

	res, err = model.Tojson("success")
	return
}

/*
导出钱包
*/
func Export(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	pwdItr, ok := rj.Get("password")
	if !ok {
		res, err = model.Errcode(5002, "password")
		return
	}
	pwd := pwdItr.(string)
	if pwd == "" {
		res, err = model.Errcode(model.NoField, "钱包密码不能为空")
		return
	}
	str, err := config.Area.Keystore.ExportMnemonic(pwd)
	if err != nil {
		if err.Error() == config.ERROR_wallet_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, _ = model.Errcode(SystemError)
		return
	}
	res, err = model.Tojson(str)
	return
}

/*
导入钱包
*/
func Import(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	wordsItr, ok := rj.Get("words")
	if !ok {
		res, err = model.Errcode(5002, "words")
		return
	}
	words := wordsItr.(string)
	if words == "" {
		res, err = model.Errcode(model.NoField, "助记词不能为空")
		return
	}

	pwdItr, ok := rj.Get("password") //words, pwd, firstCoinAddressPassword, netAddressAndDHkeyPassword
	if !ok {
		res, err = model.Errcode(5002, "password")
		return
	}
	pwd := pwdItr.(string)
	if pwd == "" {
		res, err = model.Errcode(model.NoField, "钱包密码不能为空")
		return
	}

	firstCoinAddressPasswordItr, ok := rj.Get("firstCoinAddressPassword")
	if !ok {
		res, err = model.Errcode(5002, "firstCoinAddressPassword")
		return
	}
	firstCoinAddressPassword := firstCoinAddressPasswordItr.(string)
	if firstCoinAddressPassword == "" {
		res, err = model.Errcode(model.NoField, "首个钱包地址密码不能为空")
		return
	}

	netAddressAndDHkeyPasswordItr, ok := rj.Get("netAddressAndDHkeyPassword") //words, pwd, firstCoinAddressPassword, netAddressAndDHkeyPassword
	if !ok {
		res, err = model.Errcode(5002, "netAddressAndDHkeyPassword")
		return
	}
	netAddressAndDHkeyPassword := netAddressAndDHkeyPasswordItr.(string)
	if netAddressAndDHkeyPassword == "" {
		res, err = model.Errcode(model.NoField, "网络地址密码和DHkey密码不能为空")
		return
	}

	err = config.Area.Keystore.ImportMnemonic(words, pwd, firstCoinAddressPassword, netAddressAndDHkeyPassword)
	if err != nil {

		res, _ = model.Errcode(SystemError)
		return
	}
	res, err = model.Tojson("success")
	return
}

//测试用
// func Test1(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	seed := "8kweNdXL7EEzLaaCpQ32CtxCp22CpuGiaT4LzP9iqzNEobExm1NN5CVuEwpqbrRBdXHdo2ZNCcTBCRssHfuWJMhnoTLzjJ5X5LsjKZ14Wsfudzfnjc63AXMR1tsNuxb69PGw6VUXqktbLU4XmdxdDpVN428uWmFbWDDTrjiL3kFBZzYxykh6C4GzkWLawWGGMeYpPwotXbVNmam9GL2qgb8X13QK3wcnrW6LnA4ChZSDhAQnVM2gd25Y7cSmshLoNfn7ky77wjBsLEu4KtJcCcrUdgBojVU5foe2s5AL4kPh1oF8Wo8QmqAQfGKpzjgVGupATVDv"
// 	items := mining.GetIteam()
// 	addr := crypto.AddressFromB58String("SELFCQW3Xq8JCBfGjC9VNXsxsuMdQ91scVNNx5")
// 	txpay, err := mining.CreateTxPayPub("123456789", seed, items, &addr, 5, 1, 0, "ok")
// 	if err != nil {
// 		fmt.Println(err)
// 		res, err = model.Errcode(model.FailPwd, err.Error())
// 		return
// 	}
// 	err = mining.AddTx(txpay)
// 	fmt.Println("xxxxxxxxxxx", err)
// 	res, err = model.Tojson(txpay.ConversionVO())
// 	return
// }
// func Test2(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	seed := "8kweNdXL7EEzLaaCpQ32CtxCp22CpuGiaT4LzP9iqzNEobExm1NN5CVuEwpqbrRBdXHdo2ZNCcTBCRssHfuWJMhnoTLzjJ5X5LsjKZ14Wsfudzfnjc63AXMR1tsNuxb69PGw6VUXqktbLU4XmdxdDpVN428uWmFbWDDTrjiL3kFBZzYxykh6C4GzkWLawWGGMeYpPwotXbVNmam9GL2qgb8X13QK3wcnrW6LnA4ChZSDhAQnVM2gd25Y7cSmshLoNfn7ky77wjBsLEu4KtJcCcrUdgBojVU5foe2s5AL4kPh1oF8Wo8QmqAQfGKpzjgVGupATVDv"
// 	items := mining.GetIteam()
// 	pn := mining.PayNumber{}
// 	pn.Address = crypto.AddressFromB58String("SELFCQW3Xq8JCBfGjC9VNXsxsuMdQ91scVNNx5")
// 	pn.Amount = 3
// 	pn1 := mining.PayNumber{}
// 	pn1.Address = crypto.AddressFromB58String("SELFDH7HvBhXrwA3AjfP9WSs8iMjYNcCMoiL95")
// 	pn1.Amount = 3
// 	pns := []mining.PayNumber{pn, pn1}
// 	txpay, err := mining.CreateTxsPayPub("123456789", seed, items, pns, 1, "ok")
// 	if err != nil {
// 		fmt.Println(err)
// 		res, err = model.Errcode(model.FailPwd, err.Error())
// 		return
// 	}
// 	err = mining.AddTx(txpay)
// 	fmt.Println("xxxxxxxxxxx", err)
// 	res, err = model.Tojson(txpay.ConversionVO())
// 	return
// }

// func Test(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	txid := "0a0000000000000025dfdd4de9055749ba209012252fa0437f77904ee1eca5690403769ee08204ee"
// 	seed := "8kweNdXL7EEzLaaCpQ32CtxCp22CpuGiaT4LzP9iqzNEobExm1NN5CVuEwpqbrRBdXHdo2ZNCcTBCRssHfuWJMhnoTLzjJ5X5LsjKZ14Wsfudzfnjc63AXMR1tsNuxb69PGw6VUXqktbLU4XmdxdDpVN428uWmFbWDDTrjiL3kFBZzYxykh6C4GzkWLawWGGMeYpPwotXbVNmam9GL2qgb8X13QK3wcnrW6LnA4ChZSDhAQnVM2gd25Y7cSmshLoNfn7ky77wjBsLEu4KtJcCcrUdgBojVU5foe2s5AL4kPh1oF8Wo8QmqAQfGKpzjgVGupATVDv"
// 	items := mining.GetIteam()
// 	tokenitems := token.GetTokenTxItem()
// 	addr := crypto.AddressFromB58String("SELFCQW3Xq8JCBfGjC9VNXsxsuMdQ91scVNNx5")
// 	txpay, err := payment.CreateTokenPayPub("123456789", seed, txid, items, tokenitems, &addr, 5, 1, 0, "ok")
// 	if err != nil {
// 		fmt.Println(err)
// 		res, err = model.Errcode(model.FailPwd, err.Error())
// 		return
// 	}
// 	err = mining.AddTx(txpay)
// 	fmt.Println("xxxxxxxxxxx", err)
// 	res, err = model.Tojson(txpay)
// 	return
// }

/*
铸造NFT
*/
func BuildNFT(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
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

	ownerItr, ok := rj.Get("owner")
	if !ok {
		res, err = model.Errcode(model.NoField, "owner")
		return
	}
	ownerStr := ownerItr.(string)
	owner := crypto.AddressFromB58String(ownerStr)
	if !crypto.ValidAddr(config.AddrPre, owner) {
		res, err = model.Errcode(ContentIncorrectFormat, "owner")
		return
	}

	name := ""
	nameItr, ok := rj.Get("name")
	if ok && rj.VerifyType("name", "string") {
		name = nameItr.(string)
	}
	runeLength = len([]byte(name))
	if runeLength > 1024 || runeLength <= 0 {
		res, err = model.Errcode(CommentOverLengthMax, "name")
		return
	}

	symbol := ""
	symbolItr, ok := rj.Get("symbol")
	if ok && rj.VerifyType("symbol", "string") {
		symbol = symbolItr.(string)
	}
	runeLength = len([]byte(symbol))
	if runeLength > 1024 || runeLength <= 0 {
		res, err = model.Errcode(CommentOverLengthMax, "symbol")
		return
	}

	resources := ""
	resourcesItr, ok := rj.Get("resources")
	if ok && rj.VerifyType("resources", "string") {
		resources = resourcesItr.(string)
	}
	runeLength = len([]byte(resources))
	if runeLength > 1024 || runeLength <= 0 {
		res, err = model.Errcode(CommentOverLengthMax, "resources")
		return
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas)
	if total < gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	txpay, err := mining.BuildNFT(&src, gas, pwd, comment, &owner, name, symbol, resources)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result, err := utils.ChangeMap(txpay.GetVOJSON())
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)
	return
}

/*
转让NFT
*/
func TransferNFT(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
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

	nftidItr, ok := rj.Get("nftid")
	if !ok {
		res, err = model.Errcode(model.NoField, "nftid")
		return
	}
	nftidStr := nftidItr.(string)
	nftid, e := hex.DecodeString(nftidStr)
	if e != nil {
		res, err = model.Errcode(ContentIncorrectFormat, "nftid")
		return
	}

	ownerItr, ok := rj.Get("owner")
	if !ok {
		res, err = model.Errcode(model.NoField, "owner")
		return
	}
	ownerStr := ownerItr.(string)
	owner := crypto.AddressFromB58String(ownerStr)
	if !crypto.ValidAddr(config.AddrPre, owner) {
		res, err = model.Errcode(ContentIncorrectFormat, "owner")
		return
	}

	//查询这个NFT的所属者
	oldOwner, err := mining.FindNFTOwner(nftid)
	if err != nil {
		return nil, err
	}
	//检查这个nft是否属于自己
	_, ok = config.Area.Keystore.FindAddress(*oldOwner)
	if !ok {
		res, err = model.Errcode(NftNotOwner, config.ERROR_tx_nft_not_our_own.Error())
		return
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(oldOwner, gas)
	if total < gas {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	txpay, err := mining.TransferNFT(gas, pwd, comment, nftid, &owner)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {

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
	result, err := utils.ChangeMap(txpay.GetVOJSON())
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	res, err = model.Tojson(result)
	return
}

/*
查询属于这一个地址的所有NFT
*/
func GetNFTByOwner(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	nfts, err := mining.GetAddrNFT(&dst)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	nftVOs := make([]interface{}, 0, len(nfts))
	for _, one := range nfts {
		if one.NFT_ID == nil || len(one.NFT_ID) == 0 {
			one.NFT_ID = *one.GetHash()
		}
		//查询这个NFT的所属者
		owner, err := mining.FindNFTOwner(one.NFT_ID)
		if err != nil {
			return nil, err
		}
		one.NFT_Owner = *owner
		nftVOs = append(nftVOs, one.GetVOJSON())

	}
	res, err = model.Tojson(nftVOs)
	return
}

/*
查询属于本钱包地址的所有NFT
*/
func GetNFTAll(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	nftVOs := make([]interface{}, 0)
	for _, one := range config.Area.Keystore.GetAddrAll() {
		nfts, e := mining.GetAddrNFT(&one.Addr)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		for _, one := range nfts {
			if one.NFT_ID == nil || len(one.NFT_ID) == 0 {
				one.NFT_ID = *one.GetHash()
			}
			//查询这个NFT的所属者
			owner, err := mining.FindNFTOwner(one.NFT_ID)
			if err != nil {
				return nil, err
			}
			one.NFT_Owner = *owner
			nftVOs = append(nftVOs, one.GetVOJSON())
		}
	}
	res, err = model.Tojson(nftVOs)
	return
}

/*
查询一个NFT详细信息
*/
func GetNFTID(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// engine.Log.Info("FindTx RPC")
	idItr, ok := rj.Get("nftid")
	if !ok {
		res, err = model.Errcode(model.NoField, "nftid")
		return
	}
	txidStr := idItr.(string)

	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(ContentIncorrectFormat, "nftid")
		return
	}

	if mining.ParseTxClass(txid) != config.Wallet_tx_type_nft {
		res, err = model.Errcode(ContentIncorrectFormat, "nftid")
		return
	}

	outMap := make(map[string]interface{})
	txItr, code := mining.FindTxNFT(txid) //mining.FindTxJsonVo(txid)

	txItrInterface := txItr.GetVOJSON()
	outMap["txinfo"] = txItrInterface
	outMap["upchaincode"] = code
	res, err = model.Tojson(outMap)
	return
}

/*
获取一个地址的角色
1=见证人;2=社区节点;3=轻节点;4=什么也不是
*/
func GetRoleAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	out := make(map[string]interface{})
	out["type"] = mining.GetAddrState(addrCoin)

	////是轻节点
	//if mining.GetDepositLightAddr(&addrCoin) > 0 {
	//	out["type"] = 3
	//}
	////是社区节点
	//if mining.GetDepositCommunityAddr(&addrCoin) > 0 {
	//	out["type"] = 2
	//}
	////是见证人
	//if mining.GetDepositWitnessAddr(&addrCoin) > 0 {
	//	out["type"] = 1
	//}
	res, err = model.Tojson(out)
	return
}

/*
获取某一帐号余额
*/
func GetAccountOther(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	value, valueFrozen, _ := mining.GetNotspendByAddrOther(mining.GetLongChain(), addrCoin)
	// fmt.Println(addr)
	getaccount := model.GetAccount{
		Balance:       value,
		BalanceFrozen: valueFrozen,
	}
	res, err = model.Tojson(getaccount)
	return
}

/*
获取一个地址投票给谁，投了多少票
*/
func GetVoteAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrItr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	addrs := []string{}
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&addrs)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	if len(addrs) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	addrInfos := []*AddrInfo{}
	communityAddrs := []crypto.AddressCoin{}
	lightAddrs := []crypto.AddressCoin{}
	addrIndex := make(map[string]int)
	for i, addr := range addrs {
		addrCoin := crypto.AddressFromB58String(addr)
		ok = crypto.ValidAddr(config.AddrPre, addrCoin)
		if !ok {
			res, err = model.Errcode(ContentIncorrectFormat, "addresses")
			return
		}

		// 查余额
		value, valueFrozen, _ := mining.GetNotspendByAddrOther(mining.GetLongChain(), addrCoin)
		// 查询地址角色
		role := mining.GetAddrState(addrCoin)
		addrInfo := &AddrInfo{
			Balance:       value,
			BalanceFrozen: valueFrozen,
			Role:          role,
		}

		switch role {
		case 1: // 见证人
			if mining.GetDepositWitnessAddr(&addrCoin) > 0 {
				role = 1
				addrInfo.DepositIn = config.Mining_deposit
			}
		case 2: // 社区
			communityAddrs = append(communityAddrs, addrCoin)
		case 3: // 轻节点
			lightAddrs = append(lightAddrs, addrCoin)
		default: // 无
		}
		addrIndex[addrCoin.B58String()] = i
		addrInfos = append(addrInfos, addrInfo)
	}

	balanceMgr := mining.GetLongChain().GetBalance()
	//社区节点
	if len(communityAddrs) > 0 {
		for _, addr := range communityAddrs {
			itemInfo := balanceMgr.GetDepositCommunity(&addr)
			if itemInfo != nil {
				i := addrIndex[addr.B58String()]
				addrInfos[i].DepositIn = config.Mining_vote
				addrInfos[i].VoteAddr = itemInfo.WitnessAddr.B58String()
				addrInfos[i].VoteIn = itemInfo.Value
			}
		}
		//communitys := precompiled.GetCommunityList(communityAddrs)
		//for _, community := range communitys {
		//	if community.Wit != evmutils.ZeroAddress {
		//		addr := evmutils.AddressToAddressCoin(community.Addr.Bytes())
		//		i := addrIndex[addr.B58String()]
		//		addrInfos[i].DepositIn = config.Mining_vote
		//		voteAddr := evmutils.AddressToAddressCoin(community.Wit.Bytes())
		//		addrInfos[i].VoteAddr = voteAddr.B58String()
		//		addrInfos[i].VoteIn = community.Vote.Uint64()
		//	}
		//}
	}

	//轻节点
	if len(lightAddrs) > 0 {
		for _, addr := range lightAddrs {
			itemInfo := balanceMgr.GetDepositVote(&addr)
			if itemInfo != nil {
				i := addrIndex[addr.B58String()]
				addrInfos[i].DepositIn = config.Mining_light_min
				addrInfos[i].VoteAddr = itemInfo.WitnessAddr.B58String()
				addrInfos[i].VoteIn = itemInfo.Value
			}
		}
		//lights := precompiled.GetLightList(lightAddrs)
		//for _, light := range lights {
		//	if light.Score.Uint64() != 0 {
		//		addr := evmutils.AddressToAddressCoin(light.Addr.Bytes())
		//		i := addrIndex[addr.B58String()]
		//		addrInfos[i].DepositIn = config.Mining_light_min
		//		voteAddr := evmutils.AddressToAddressCoin(light.C.Bytes())
		//		addrInfos[i].VoteAddr = voteAddr.B58String()
		//		addrInfos[i].VoteIn = light.Vote.Uint64()
		//	}
		//}
	}

	res, err = model.Tojson(addrInfos)
	return
}

/*
地址余额，角色，押金，投票给谁，及投票数量
*/
type AddrInfo struct {
	Balance       uint64 `json:"balance"`   //可用余额
	BalanceFrozen uint64 `json:"balance_f"` //锁定余额
	Role          int    `json:"role"`      //角色
	DepositIn     uint64 `json:"depositin"` //押金
	VoteAddr      string `json:"voteaddr"`  //给哪个社区节点地址投票
	VoteIn        uint64 `json:"votein"`    //投票金额
	VoteNum       uint64 `json:"votenum"`   //获得的投票数量
}

/*
获取本节点总的质押数量
*/
func GetDepositNum(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrStrs := []string{}
	addrItr, ok := rj.Get("addresses")
	if ok {
		bs, err2 := json.Marshal(addrItr)
		if err2 != nil {
			res, err = model.Errcode(model.TypeWrong, "addresses")
			return
		}

		decoder := json.NewDecoder(bytes.NewBuffer(bs))
		decoder.UseNumber()
		err = decoder.Decode(&addrStrs)
		if err != nil {
			res, err = model.Errcode(model.TypeWrong, "addresses")
			return
		}
	}

	addrs := []crypto.AddressCoin{}
	for _, addr := range addrStrs {
		addrs = append(addrs, crypto.AddressFromB58String(addr))
	}

	if len(addrs) <= 0 {
		allAddrs := mining.Area.Keystore.GetAddrAll()
		for _, addr := range allAddrs {
			addrs = append(addrs, addr.Addr)
		}
	}

	// depositAmount := toUint64(0)
	witCount := toUint64(0)
	communityCount := toUint64(0)
	lightCount := toUint64(0)
	for _, addr := range addrs {
		role := mining.GetAddrState(addr)
		switch role {
		case 1: // 见证人
			witCount++
		case 2: // 社区
			communityCount++
		case 3: // 轻节点
			lightCount++
		}

		// depositAmount += precompiled.GetMyDeposit(addr.Addr)
	}

	// if config.ParseInitFlag() {
	// 	depositAmount += config.Mining_deposit
	// }

	info := map[string]interface{}{
		"deposit": witCount*config.Mining_deposit + communityCount*config.Mining_vote + lightCount*config.Mining_light_min,
		// "total":   depositAmount,
	}
	res, err = model.Tojson(info)
	return
}

/*
获取总的质押量
*/
func GetDepositNumForAll(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//lights := precompiled.GetAllLight(nil)
	//depositAmount := toUint64(0)
	//for _, l := range lights {
	//	depositAmount += l.Score.Uint64()
	//}

	depositAmount := toUint64(0)
	balanceMgr := mining.GetLongChain().GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()
	wmc.Range(func(key, value any) bool {
		depositAmount += config.Mining_deposit
		comms := value.([]crypto.AddressCoin)
		depositAmount += (uint64(len(comms)) * config.Mining_vote)
		return true
	})
	lightitems := balanceMgr.GetAllLights()
	depositAmount += (uint64(len(lightitems)) * config.Mining_light_min)

	//lightTotal := precompiled.GetLightTotal(nil)
	//depositAmount := lightTotal * config.Mining_light_min
	//community := precompiled.GetAllCommunity(nil)
	//for _, c := range community {
	//	depositAmount += c.Score.Uint64()
	//}

	//wbg := mining.GetBackupWitnessList()
	//for _, w := range wbg {
	//	depositAmount += w.Score
	//}

	res, err = model.Tojson(depositAmount)
	return
}

type T struct {
	Constant        bool    `json:"constant"`
	Inputs          []Input `json:"inputs"`
	Name            string  `json:"name"`
	Outputs         []Input `json:"outputs"`
	Payable         bool    `json:"payable"`
	StateMutability string  `json:"stateMutability"`
	Type            string  `json:"type"`
}
type Input struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func GetAbiInput(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	t := "function"
	method, ok := rj.Get("method")
	if !ok {
		method = ""
		t = "constructor"
	}

	t1 := T{
		Constant:        true,
		Name:            method.(string),
		Inputs:          []Input{},
		Outputs:         []Input{},
		Payable:         false,
		StateMutability: "view",
		Type:            t,
	}

	fieldStr, ok := rj.Get("field_type")
	if !ok {
		fieldStr = ""
	}
	if !rj.VerifyType("field_type", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}
	fields := make([]string, 0)
	if len(fieldStr.(string)) != 0 {
		fields = strings.Split(fieldStr.(string), "-")
		for _, field := range fields {
			t1.Inputs = append(t1.Inputs, Input{
				Name: "",
				Type: field,
			})
		}
	}
	initData := []T{t1}

	initBody, _ := json.Marshal(initData)

	abi, err := abi.JSON(bytes.NewReader(initBody))
	if err != nil {
		engine.Log.Error("abi json err", err)
		res, err = model.Errcode(SystemError, "abi json err")
		return
	}

	paramStr, ok := rj.Get("params")
	if !ok {
		paramStr = ""
	}
	if !rj.VerifyType("params", "string") {
		res, err = model.Errcode(model.TypeWrong, "params")
		return
	}
	params := strings.Split(paramStr.(string), "-")
	args := make([]interface{}, 0)

	if len(paramStr.(string)) != 0 {
		if len(params) != len(fields) {
			res, err = model.Errcode(SystemError, "field_type length must same as params")
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
					return res, err
				}
				args = append(args, bytesParam)
			case "uint256":
				intParam, err := decimal.NewFromString(param)
				if err != nil {
					return res, err
				}
				args = append(args, intParam.BigInt())
			case "uint8":
				p, err := strconv.Atoi(param)
				if err != nil {
					res, err = model.Errcode(SystemError, err.Error())
					return res, err
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
						res, err = model.Errcode(SystemError, err.Error())
						return res, err
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

	packB, err := abi.Pack(method.(string), args...)
	if err != nil {
		res, err = model.Errcode(SystemError, "abi Pack err")
		return
	}

	res, err = model.Tojson(common.Bytes2Hex(packB))
	return res, err
}

func GetTxGas(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	/*
		txType, ok := rj.Get("type")
		if !ok {
			txType = float64(config.Wallet_tx_type_pay)
		}

			txCount := toUint64(0)

			chain := mining.GetLongChain()
			if chain == nil {
				return model.Tojson(SystemError)
			}

			hs := chain.GetHistoryBalanceDesc(0, 500)
			for _, h := range hs {
				if len(h.TokenId) != 0 {
					txItr, _, _ := mining.FindTxJsonV1(h.TokenId)
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

			normal := toUint64(0)
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

			res, err = model.Tojson(data)

			return res, err
	*/

	txAveGas := mining.GetLongChain().Balance.GetTxAveGas()
	count := toUint64(0)
	totalGas := toUint64(0)
	for _, gas := range txAveGas.AllGas {
		if gas > 0 {
			count++
			totalGas += gas
		}
	}

	normal := toUint64(0)
	if count != 0 {
		normal = totalGas / count
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

	res, err = model.Tojson(data)

	return res, err
}

func GetContractEvent(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var customTxEvent = false
	if _, ok := rj.Get("custom_tx_event"); ok {
		customTxEvent = true
	}

	//if !config.EVM_Reward_Enable && customTxEvent {
	if customTxEvent {
		//内存模式的事件信息
		return GetCustomTxEvents(rj, w, r)
	}

	list := &go_protos.ContractEventInfoList{}

	hash, ok := rj.Get("hash")
	if !ok {
		res, err = model.Errcode(model.NoField, "hash")
		return
	}

	if !rj.VerifyType("hash", "string") {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	hByte, err := hex.DecodeString(hash.(string))
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	var intHeight []byte
	height, ok := rj.Get("height")
	if !ok {
		_, _, _, blockHeight, _ := mining.FindTxJsonVo(hByte)
		intHeight = evmutils.New(0).SetUint64(blockHeight).Bytes()
	} else {
		intHeight = evmutils.New(int64(height.(float64))).Bytes()
	}

	key := append([]byte(config.DBKEY_BLOCK_EVENT), intHeight...)
	body, err := db.LevelDB.GetDB().HGet(key, hByte)
	if err != nil {
		//engine.Log.Error("获取保存合约事件日志失败%s", err.Error())
		res, err = model.Tojson(list)
		return
	}
	if len(body) == 0 {
		res, err = model.Tojson(list)
		return
	}

	err = proto.Unmarshal(body, list)
	if err != nil {
		res, err = model.Tojson(list)
		return
	}

	//historys := make([]precompiled.LogRewardHistoryV0, 0, len(list.ContractEvents))
	//for _, l := range list.ContractEvents {
	//	v0 := precompiled.LogRewardHistoryV0{}
	//	log := precompiled.UnPackLogRewardHistoryLog(l)
	//	address := evmutils.AddressToAddressCoin(log.Into[:])
	//	v0.Into = address.B58String()
	//	if log.From == evmutils.ZeroAddress {
	//		v0.From = ""
	//	} else {
	//		from := evmutils.AddressToAddressCoin(log.From[:])
	//		v0.From = from.B58String()
	//	}
	//	v0.Reward = log.Reward
	//	historys = append(historys, v0)
	//}

	res, err = model.Tojson(list)
	return res, err
}

// 获取自定义交易事件
func GetCustomTxEvents(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	e := go_protos.CustomTxEvents{
		CustomTxEvents: []*go_protos.CustomTxEvent{},
		RewardPools:    []*go_protos.CustomTxEvent{},
	}
	data := make(map[string]interface{})
	data["event_data"] = e.CustomTxEvents
	data["pool_data"] = e.RewardPools

	hash, ok := rj.Get("hash")
	if !ok {
		res, err = model.Errcode(model.NoField, "hash")
		return
	}

	if !rj.VerifyType("hash", "string") {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	hByte, err := hex.DecodeString(hash.(string))
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	key := config.BuildCustomTxEvent(hByte)
	body, err := db.LevelTempDB.Find(key)
	if err != nil || len(*body) == 0 {
		res, err = model.Tojson(data)
		return
	}

	if err = e.Unmarshal(*body); err != nil {
		res, err = model.Tojson(data)
		return
	}
	data["event_data"] = e.CustomTxEvents
	data["pool_data"] = e.RewardPools
	res, err = model.Tojson(data)
	return res, err
}

// 解析和过滤自定义交易事件到Vout
func parseCustomTxEvents2VoutV0s(txid []byte, addrsm map[string]struct{}) ([]*mining.VoutVO, error) {
	es := go_protos.CustomTxEvents{}
	key := config.BuildCustomTxEvent(txid)
	body, err := db.LevelTempDB.Find(key)
	if err != nil || len(*body) == 0 {
		return nil, errors.New("empty")
	}

	if err = es.Unmarshal(*body); err != nil {
		return nil, err
	}

	out := []*mining.VoutVO{}
	for _, e := range es.CustomTxEvents {
		if len(addrsm) == 0 || addrsm == nil {
			out = append(out, &mining.VoutVO{
				Value:   e.Value,
				Address: e.Addr,
			})
		} else {
			if _, ok := addrsm[e.Addr]; ok {
				out = append(out, &mining.VoutVO{
					Value:   e.Value,
					Address: e.Addr,
				})
			}
		}
	}

	if len(out) == 0 {
		return nil, errors.New("empty")
	}

	return out, nil
}

func GetExchangeRate(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	res, err = model.Tojson(0.5)
	return res, err
}

// {"code":200,"data":{"usdCny":"6.974"},"msg":"数据源汇率"}
type usdRmbRate struct {
	Code int64 `json:"code"`
	Data struct {
		UsdCny string `json:"usdCny"`
	} `json:"data"`
}

func GetUsdRmbRate(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	uri := "http://8.212.2.156:8889/get_exchange_rate"
	resp, err := http.Get(uri)
	if err != nil {
		res, err = model.Errcode(SystemError, "")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res, err = model.Errcode(SystemError, "")
		return
	}
	data := usdRmbRate{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		res, err = model.Errcode(SystemError, "")
		return
	}
	res, err = model.Tojson(data.Data.UsdCny)
	return
}

type DepositNum struct {
	Light     uint64 `json:"light"`     // 轻节点质押数量
	Community uint64 `json:"community"` // 社区节点质押数量
	Witness   uint64 `json:"witness"`   // 见证人节点质押数量
}

/*
获取地址交易
*/
func GetAddressTx_Bak(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}
	start := (pageInt - 1) * pageSizeInt
	//addrkey := []byte(config.Address_history_tx + "_" + strings.ToLower(addr.(string)))
	addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(addr.(string)))...)

	index := GetAddrTxIndex(addr.(string))
	//balanceHistory := mining.NewBalanceHistory()

	data := make([]map[string]interface{}, 0)
	for i := 0; i < pageSizeInt; i++ {
		indexBs := make([]byte, 8)
		startindex := toUint64(0)
		if index > (uint64(start) + uint64(i)) {
			startindex = index - uint64(start) - uint64(i)
		}
		binary.LittleEndian.PutUint64(indexBs, startindex)

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
				continue
			}
			var blockHashStr string
			if blockHash != nil {
				blockHashStr = hex.EncodeToString(*blockHash)
			}
			//txinfo转map，然后重写某些字段
			item := JsonMethod(txItr)
			tx, _, _ := mining.FindTx(txid)
			txClass, vins, vouts := DealTxInfo(tx, addr.(string), blockHeight)
			if txClass > 0 {
				item["type"] = txClass
				if vins != nil {
					item["vin"] = vins
				}
				if vouts != nil {
					item["vout"] = vouts
				}
			}
			if tx.Class() == 1 {
				v1 := item["vin"].([]interface{})
				v2 := v1[0].(map[string]interface{})
				v2["addr"] = ""
				item["vin"] = v1
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

	res, err = model.Tojson(data1)
	return
}

/*
获取地址交易
*/
func GetAddressTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addr := addrItr.(string)

	//contractAddr := ""
	//if contractAddrItr, ok := rj.Get("contractaddress"); ok {
	//	if rj.VerifyType("contractaddress", "string") {
	//		contractAddr = contractAddrItr.(string)
	//	}
	//}

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
	startIndex := uint64((pageInt - 1) * pageSizeInt)
	endIndex := uint64(pageInt * pageSizeInt)

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

	out := make(map[string]interface{})
	out["count"] = count
	out["data"] = items
	res, err = model.Tojson(out)
	return
}

/*
获取地址交易
NOTE: 带细化交易的方法
*/
//func GetAddressTxV2_Bak(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	//地址
//	addresses, errcode := getArrayStrParams(rj, "addresses")
//	if errcode != 0 {
//		res, err = model.Errcode(errcode, "addresses")
//		return
//	}
//
//	//币种
//	contractaddresses, code := getArrayStrParams(rj, "contractaddresses")
//	if code != 0 {
//		res, err = model.Errcode(code, "contractaddresses")
//		return
//	}
//	onlyErc := false
//	if onlyErcIter, ok := rj.Get("onlyerc"); ok {
//		if v, ok := onlyErcIter.(bool); ok {
//			onlyErc = v
//		}
//	}
//
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
//
//	limitTxCountOneAddress := 300 //单个地址限制最新的300条记录左右
//	timeOut := int64(6000)        //ms
//
//	pageSizeInt := int(pageSize.(float64))
//	if pageSizeInt > limitTxCountOneAddress {
//		pageSizeInt = limitTxCountOneAddress
//	}
//	startIndex := (pageInt - 1) * pageSizeInt
//	endIndex := pageInt * pageSizeInt
//
//	ercAddrs := sync.Map{}
//	addrs := sync.Map{}
//	if len(addresses) == 0 {
//		addrCoins := config.Area.Keystore.GetAddrAll()
//		for _, addr := range addrCoins {
//			addresses = append(addresses, addr.GetAddrStr())
//			addrs.Store(addr.GetAddrStr(), true)
//		}
//	} else {
//		for _, addr := range addresses {
//			addrs.Store(addr, true)
//		}
//	}
//
//	for _, addr := range contractaddresses {
//		ercAddrs.Store(addr, true)
//	}
//
//	//构造查询条件
//	qtx := queryTx{
//		isAll:    false,
//		onlyErc:  onlyErc,
//		ercAddrs: ercAddrs,
//		addrs:    addrs,
//	}
//
//	repeatedTx := sync.Map{}
//	startAt := time.Now().UnixMilli()
//	items := make([]map[string]interface{}, 0)
//	for _, addr := range addresses {
//		txCountOneAddress := 0
//		maxIndex := GetAddrTxIndex(addr)
//		for j := maxIndex; j > 0; j-- {
//			if txCountOneAddress >= limitTxCountOneAddress || time.Now().UnixMilli()-startAt > timeOut {
//				break
//			}
//
//			addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(addr))...)
//			indexBs := make([]byte, 8)
//			binary.LittleEndian.PutUint64(indexBs, j)
//			txid, err := db.LevelTempDB.HGet(addrkey, indexBs)
//			if err != nil {
//				continue
//			}
//
//			txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(txid)
//			if txItr == nil {
//				continue
//			}
//
//			var blockHashStr string
//			if blockHash != nil {
//				blockHashStr = hex.EncodeToString(*blockHash)
//			}
//			if _, ok := repeatedTx.Load(utils.Bytes2string(txid)); ok {
//				continue
//			}
//			repeatedTx.Store(utils.Bytes2string(txid), true)
//			txs := parseTxAndCustomTx(txid, blockHeight, qtx)
//			for i, _ := range txs {
//				item := make(map[string]interface{})
//				item["upchaincode"] = code
//				item["blockheight"] = blockHeight
//				item["blockhash"] = blockHashStr
//				item["timestamp"] = timestamp
//				item["txinfo"] = txs[i]
//				items = append(items, item)
//				txCountOneAddress++
//			}
//		}
//	}
//
//	//交易的总数(包含细化的)
//	count := len(items)
//	//按时间戳倒序
//	sort.Sort(txsSort(items))
//
//	out := make(map[string]interface{})
//	if len(items) == 0 {
//		out["count"] = 0
//		out["data"] = make([]map[string]interface{}, 0)
//		res, err = model.Tojson(out)
//		return
//	}
//
//	if startIndex > len(items) {
//		out["count"] = count
//		out["data"] = make([]map[string]interface{}, 0)
//		res, err = model.Tojson(out)
//		return
//	}
//
//	if endIndex > len(items) {
//		endIndex = len(items)
//	}
//
//	data := make([]map[string]interface{}, endIndex-startIndex)
//	copy(data[:], items[startIndex:endIndex])
//
//	out["count"] = count
//	out["data"] = data
//	res, err = model.Tojson(out)
//	return
//}

type txsSort []map[string]interface{}

func (x txsSort) Len() int           { return len(x) }
func (x txsSort) Less(i, j int) bool { return x[i]["timestamp"].(int64) > x[j]["timestamp"].(int64) }
func (x txsSort) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

/*
获取地址ERC20交易
*/
func GetAddressErc20Tx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//地址
	addresses, errcode := getArrayStrParams(rj, "addresses")
	if errcode != 0 {
		res, err = model.Errcode(errcode, "addresses")
		return
	}

	//币种
	contractaddresses, code := getArrayStrParams(rj, "contractaddresses")
	if code != 0 {
		res, err = model.Errcode(code, "contractaddresses")
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	addrm := sync.Map{}
	for _, addr := range addresses {
		addrm.Store(addr, true)
	}
	startIndex := (pageInt - 1) * pageSizeInt
	endIndex := pageInt * pageSizeInt

	txids := make([][]byte, 0)
	for _, caddr := range contractaddresses {
		maxIndex := GetAddrTxIndex(caddr)
		for j := maxIndex; j > 0; j-- {
			addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(caddr))...)
			indexBs := make([]byte, 8)
			binary.LittleEndian.PutUint64(indexBs, j)
			txid, err := db.LevelTempDB.HGet(addrkey, indexBs)
			if err != nil {
				continue
			}

			txids = append(txids, txid)
		}
	}

	outMap := sync.Map{}
	wg := sync.WaitGroup{}
	for _, txid := range txids {
		wg.Add(1)
		go func(txid []byte) {
			defer wg.Done()
			data := parseOneTx(txid)

			//过滤地址
			vinAddress := data["vinaddress"].(string)
			if _, ok := addrm.Load(vinAddress); !ok {
				return
			}

			outMap.Store(utils.Bytes2string(txid), data)
		}(txid)
	}
	wg.Wait()

	items := make([]map[string]interface{}, 0)
	for _, txid := range txids {
		if item, ok := outMap.Load(utils.Bytes2string(txid)); ok {
			items = append(items, item.(map[string]interface{}))
		}
	}

	count := len(items)
	out := make(map[string]interface{})
	if len(items) == 0 {
		out["count"] = 0
		out["data"] = make([]map[string]interface{}, 0)
		res, err = model.Tojson(out)
		return
	}

	if startIndex > len(items) {
		out["count"] = count
		out["data"] = make([]map[string]interface{}, 0)
		res, err = model.Tojson(out)
		return
	}

	if endIndex > len(items) {
		endIndex = len(items)
	}

	data := make([]map[string]interface{}, endIndex-startIndex)
	copy(data[:], items[startIndex:endIndex])

	out["count"] = count
	out["data"] = data
	res, err = model.Tojson(out)
	return
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
	item := JsonMethod(txInfo)
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
获取挖矿交易
*/
func GetMinerBlock(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}
	start := (pageInt - 1) * pageSizeInt
	//addrkey := []byte(config.Miner_history_tx + "_" + strings.ToLower(addr.(string)))
	addrkey := append(config.Miner_history_tx, []byte("_"+strings.ToLower(addr.(string)))...)

	index := GetMinerBlockIndex(addr.(string))

	data := make([]map[string]interface{}, 0)
	for i := 0; i < pageSizeInt; i++ {
		indexBs := make([]byte, 8)
		startindex := toUint64(0)
		if index > (uint64(start) + uint64(i)) {
			startindex = index - uint64(start) - uint64(i)
		}
		binary.LittleEndian.PutUint64(indexBs, startindex)

		blockHash, err := db.LevelTempDB.HGet(addrkey, indexBs)
		if err != nil {
			continue
		}
		if len(blockHash) != 0 {
			//head, err := mining.LoadBlockHeadByHash(&blockHash)
			head, err := mining.LoadBlockHeadVOByHash(&blockHash)
			if err != nil {
				return nil, err
			}

			//_, reward := mining.GetLongChain().Balance.GetAddBlockNum(head.Witness.B58String())
			gas := toUint64(0)
			for _, txItr := range head.Txs {
				gas += txItr.GetGas()
			}
			//reward := config.BLOCK_TOTAL_REWARD + gas
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

	res, err = model.Tojson(data1)
	return
}

/*
获取所有交易
*/
func GetAllTxV2(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(SystemError)
		return
	}

	his := chain.GetHistoryBalanceDesc(uint64(start), pageSizeInt)
	items := make([]map[string]interface{}, 0)

	for _, h := range his {
		if h.Txid != nil {
			tx := parseOneTx(h.Txid)
			items = append(items, tx)
		}
	}

	data := map[string]interface{}{}

	data["count"] = chain.GetHistoryBalanceCount() //去除创世块部署奖励合约那笔交易
	data["data"] = items

	res, err = model.Tojson(data)
	return
}

/*
获取所有交易
*/
func GetAllTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(SystemError)
		return
	}

	his := chain.GetHistoryBalanceDesc(uint64(start), pageSizeInt)
	data := make([]map[string]interface{}, 0)

	//balanceHistory := mining.NewBalanceHistory()

	for _, h := range his {
		if h.Txid != nil {
			txItr, code, blockHash, blockHeight, timestamp := mining.FindTxJsonVo(h.Txid)
			if txItr == nil {
				//customTx, err := balanceHistory.GetCustomTx(h.TokenId)
				//if err != nil {
				//	continue
				//}
				//data = append(data, convertRewardHistoryItemToTxInfo(customTx))
				continue
			}

			var blockHashStr string
			if blockHash != nil {
				blockHashStr = hex.EncodeToString(*blockHash)
			}
			item := JsonMethod(txItr)
			tx, _, _ := mining.FindTx(h.Txid)
			vouts := tx.GetVout()
			outs := *vouts
			var addr string

			if bytes.Equal(outs[0].Address, precompiled.RewardContract) || bytes.Equal(outs[0].Address, ens.GetRegisterAddr()) {
				txVin := *tx.GetVin()

				addr = txVin[0].GetPukToAddr().B58String()
				//engine.Log.Info("2788行%s,%s", addr, h.OutAddr[0].B58String())
				//if h.OutAddr[0].B58String() != addr {
				//	addr = h.OutAddr[0].B58String()
				//}
				txClass, vins, vout := DealTxInfo(tx, addr, blockHeight)
				if txClass > 0 {
					item["type"] = txClass
					if vins != nil {
						item["vin"] = vins
					}
					if vout != nil {
						item["vout"] = vout
					}
				}

				if tx.Class() == 1 {
					v1 := item["vin"].([]interface{})
					v2 := v1[0].(map[string]interface{})
					v2["addr"] = ""
					item["vin"] = v1
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

	res, err = model.Tojson(data1)
	return
}

func GetAddrTxIndex(addr string) uint64 {
	// 获取最大索引
	chain := mining.GetLongChain()
	if chain != nil {
		return chain.GetAccountHistoryBalanceCount(addr)
	}
	return 0
}
func GetMinerBlockIndex(addr string) uint64 {
	return mining.GetLongChain().Balance.GetBlockIndex(addr)

	// 获取最大索引
	//indexKey := []byte(config.Miner_tx_count + "_" + strings.ToLower(addr))
	//addrTxCount, err := db.LevelTempDB.Get(indexKey)
	//if err != nil {
	//	return 0, err
	//}
	//if addrTxCount == nil {
	//	return 0, err
	//}
	//return binary.LittleEndian.Uint64(addrTxCount), nil

}

/*
获取轻节点详情
*/
func GetLightNodeDetail(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	//如果不是轻节点
	//list := precompiled.GetLightList([]crypto.AddressCoin{addrCoin})
	//if len(list) != 1 {
	//	res, err = model.Errcode(NotLightNode, "address")
	//	return
	//}

	//lightDetail := list[0]

	//cAddr := ""
	//cName := ""
	//cVote := new(big.Int)
	////reward := precompiled.GetNodeReward(addrCoin)
	//reward := precompiled.GetLightNodeReward(addrCoin)
	//if lightDetail.C != evmutils.ZeroAddress {
	//	c := evmutils.AddressToAddressCoin(lightDetail.C.Bytes())
	//	cAddr = c.B58String()
	//	//reward, _ = getNodeReward(addrCoin, c)
	//	cs := precompiled.GetCommunityList([]crypto.AddressCoin{c})
	//	if len(cs) == 1 {
	//		cName = cs[0].Name
	//		cVote = cs[0].Vote
	//	}
	//}

	balanceMgr := mining.GetLongChain().GetBalance()
	lightinfo := balanceMgr.GetDepositLight(&addrCoin)
	if lightinfo == nil {
		res, err = model.Errcode(NotLightNode, "address")
		return
	}

	data := make(map[string]interface{})
	data["community_addr"] = ""
	data["community_name"] = ""
	data["vote"] = toUint64(0)
	data["last_voted_height"] = toUint64(0)
	data["reward"] = toUint64(0)
	data["voteAddress"] = ""
	data["frozen_reward"] = toUint64(0)
	if lightvoteinfo := balanceMgr.GetDepositVote(&addrCoin); lightvoteinfo != nil {
		if commInfo := balanceMgr.GetDepositCommunity(&lightvoteinfo.WitnessAddr); commInfo != nil {
			data["community_name"] = commInfo.Name
		}
		data["community_addr"] = lightvoteinfo.WitnessAddr.B58String()
		lastVoteOpHeight := balanceMgr.GetLastVoteOp(lightvoteinfo.SelfAddr)
		reward := balanceMgr.GetAddrReward(lightvoteinfo.SelfAddr)
		data["vote"] = lightvoteinfo.Value
		data["last_voted_height"] = lastVoteOpHeight
		data["reward"] = reward.Uint64()

		_, lightrewards := balanceMgr.CalculateCommunityRewardAndLightReward(lightvoteinfo.WitnessAddr)
		if v, ok := lightrewards[utils.Bytes2string(lightvoteinfo.SelfAddr)]; ok {
			data["frozen_reward"] = v.Uint64()
		}

		data["voteAddress"] = lightvoteinfo.WitnessAddr.B58String()
	}

	data["start_height"] = balanceMgr.GetLastVoteOp(lightinfo.SelfAddr)
	data["light_addr"] = addr
	data["name"] = lightinfo.Name
	data["contract"] = ""
	//if config.EVM_Reward_Enable {
	//	data["contract"] = precompiled.RewardContract.B58String()
	//}
	data["deposit"] = lightinfo.Value
	//data["start_height"] =
	//data["frozen_reward"] = precompiled.GetMyFrozenReward(addrCoin)
	//data["frozen_reward"] = precompiled.GetMyLightFrozenReward(addrCoin)

	//if cAddr != "" {
	//	cAddress := crypto.AddressFromB58String(cAddr)
	//	cRewardPool := precompiled.GetCommunityRewardPool(cAddress)
	//	if cRewardPool.Cmp(big.NewInt(0)) > 0 {
	//		cRate, err := precompiled.GetRewardRatio(cAddress)
	//		if err != nil {
	//			return model.Errcode(SystemError, "address")
	//		}
	//		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
	//		v := new(big.Int).Mul(new(big.Int).SetUint64(mining.GetLongChain().GetBalance().GetDepositVote(&addrCoin).Value), new(big.Int).SetInt64(1e8))
	//		ratio := new(big.Int).Quo(v, cVote)
	//		lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
	//		frozenReward = lBigValue.Uint64()
	//	}
	//}

	res, err = model.Tojson(data)
	return
}

// 获取当前被投票人给地址的奖励
func getNodeReward(from, to crypto.AddressCoin) (uint64, error) {
	fromHex := common.Address(evmutils.AddressCoinToAddress(from))
	toHex := common.Address(evmutils.AddressCoinToAddress(to))

	keyValue := append(fromHex[:], toHex[:]...)
	key := config.BuildRewardHistory(keyValue)
	rewardByte, err := db.LevelDB.Get(key)
	if err != nil {
		return 0, err
	}
	return utils.BytesToUint64(rewardByte), nil
}

// 获取地址的出块总数和总的奖励
func GetAddressAddBlockReward(addr string) (count uint64, reward uint64) {
	//ethAddr := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(addr)))
	//nodeReward := precompiled.GetNodeReward(crypto.AddressFromB58String(addr))

	return mining.GetLongChain().Balance.GetAddBlockNum(addr)
}

/*
获取见证者节点详情
*/
func GetWitnessNodeDetail(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	witAddr := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, witAddr)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}

	// 获得质押量
	depositAmount := mining.GetDepositWitnessAddr(&witAddr)
	if depositAmount <= 0 {
		res, err = model.Errcode(NotWitnessNode, "address")
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	addBlockCount, addBlockReward := GetAddressAddBlockReward(addr.(string))

	//
	//detail, ok := precompiled.GetWitnessDetail(witAddr)
	//if !ok {
	//	return model.Errcode(SystemError, "address")
	//}
	//witnessName := mining.FindWitnessName(witAddr)
	//data := WitnessInfoV0{
	//	Deposit:        depositAmount,
	//	AddBlockCount:  addBlockCount,
	//	AddBlockReward: addBlockReward,
	//	RewardRatio:    float64(detail.Rate),
	//	TotalReward:    detail.Reward.Uint64(),
	//	FrozenReward:   detail.RemainReward.Uint64(),
	//	Name:           witnessName,
	//	CommunityNode:  []CommunityNode{},
	//	DestroyNum:     0,
	//}
	frozenReward := big.NewInt(0)
	balanceMgr := mining.GetLongChain().GetBalance()
	ratio := balanceMgr.GetDepositRate(&witAddr)
	totalReward := balanceMgr.GetAddrReward(witAddr)
	wrp := balanceMgr.GetWitnessRewardPool()
	if val, ok := wrp.Load(utils.Bytes2string(witAddr)); ok {
		rewardpool := val.(*big.Int)
		frozenReward, _ = balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(witAddr, rewardpool)
	}

	witnessName := mining.FindWitnessName(witAddr)
	data := WitnessInfoV0{
		Deposit:        depositAmount,
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		RewardRatio:    float64(ratio),
		TotalReward:    totalReward.Uint64(),
		FrozenReward:   frozenReward.Uint64(),
		Name:           witnessName,
		CommunityNode:  []CommunityNode{},
		DestroyNum:     0,
	}

	// 总票数
	//voteTotal := toUint64(0)
	//data.CommunityCount = toUint64(len(detail.Communitys))
	//i := 0
	//cs := make([]CommunityNode, 0)

	//sort.Sort(precompiled.CommunityRewardSort(detail.Communitys))
	//// 获取社区节点详情
	//for _, c := range detail.Communitys {
	//	voteTotal += c.Vote.Uint64()
	//	if i >= start && len(cs) < pageSizeInt {
	//		coin := evmutils.AddressToAddressCoin(c.Addr[:])
	//		cs = append(cs, CommunityNode{
	//			Name:        c.Name,
	//			Addr:        coin.B58String(),
	//			Deposit:     c.Score.Uint64(),
	//			Reward:      c.Reward.Uint64(),
	//			LightNum:    c.LightCount.Uint64(),
	//			VoteNum:     c.Vote.Uint64(),
	//			RewardRatio: float64(c.Rate),
	//		})
	//	}
	//	i++
	//}

	// 总票数
	voteTotal := balanceMgr.GetWitnessVote(&witAddr)

	wmc := balanceMgr.GetWitnessMapCommunitys()
	cml := balanceMgr.GetCommunityMapLights()
	if items, ok := wmc.Load(utils.Bytes2string(witAddr)); ok {
		communityAddrs := items.([]crypto.AddressCoin)
		data.CommunityCount = uint64(len(communityAddrs))

		cs := make([]CommunityNode, 0)
		// 获取社区节点详情
		for _, cAddr := range communityAddrs {
			reward := balanceMgr.GetAddrReward(cAddr)
			commInfo := balanceMgr.GetDepositCommunity(&cAddr)
			lightNum := toUint64(0)
			if lights, ok := cml.Load(utils.Bytes2string(cAddr)); ok {
				lightNum = uint64(len(lights.([]crypto.AddressCoin)))
			}
			rewardRatio := balanceMgr.GetDepositRate(&witAddr)
			cs = append(cs, CommunityNode{
				Name:        commInfo.Name,
				Addr:        commInfo.SelfAddr.B58String(),
				Deposit:     commInfo.Value,
				Reward:      reward.Uint64(),
				LightNum:    lightNum,
				VoteNum:     balanceMgr.GetCommunityVote(&cAddr),
				RewardRatio: float64(rewardRatio),
			})
		}

		sort.Sort(CommunityNodeSort(cs))
		if start, end, ok := helppager(len(cs), pageInt, pageSizeInt); ok {
			out := make([]CommunityNode, end-start)
			copy(out, cs[start:end])
			data.CommunityNode = out
		} else {
			data.CommunityNode = []CommunityNode{}
		}
	}

	data.Vote = voteTotal

	return model.Tojson(data)
}

/*
获取社区节点详情
*/
func GetCommunityNodeDetail(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
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
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}

	//communityList := precompiled.GetCommunityList([]crypto.AddressCoin{addrCoin})
	//if len(communityList) == 1 && communityList[0].Wit == evmutils.ZeroAddress {
	//	res, err = model.Errcode(NotCommunityNode, "address")
	//	return
	//}

	//if len(communityList) != 1 {
	//	res, err = model.Errcode(SystemError, "address")
	//	return
	//}
	//community := communityList[0]

	//reward := precompiled.GetNodeReward(addrCoin)
	//reward := precompiled.GetCommunityNodeReward(addrCoin)
	//ratio, err := precompiled.GetRewardRatio(addrCoin)
	//if err != nil {
	//	res, err = model.Errcode(SystemError, "address")
	//	return
	//}
	//light := precompiled.GetLightListByC(addrCoin)
	//witnessAddr := evmutils.AddressToAddressCoin(community.Wit[:])
	//name := mining.FindWitnessName(witnessAddr)
	////frozenReward := precompiled.GetMyFrozenReward(addrCoin)
	////frozenReward := precompiled.GetMyCommunityFrozenReward(addrCoin)

	//var frozenReward uint64
	//cRewardPool := precompiled.GetCommunityRewardPool(addrCoin)
	//if cRewardPool.Cmp(big.NewInt(0)) > 0 {
	//	lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(ratio))), new(big.Int).SetInt64(100))
	//	cBigValue := new(big.Int).Sub(cRewardPool, lTotal)
	//	frozenReward = cBigValue.Uint64()
	//}

	balanceMgr := mining.GetLongChain().GetBalance()
	commInfo := balanceMgr.GetDepositCommunity(&addrCoin)
	if commInfo == nil {
		res, err = model.Errcode(NotCommunityNode, "address")
		return
	}

	cml := balanceMgr.GetCommunityMapLights()
	lightCount := toUint64(0)
	lightAddrs := []crypto.AddressCoin{}
	if items, ok := cml.Load(utils.Bytes2string(addrCoin)); ok {
		lightAddrs = items.([]crypto.AddressCoin)
		lightCount = uint64(len(lightAddrs))
	}
	rewardRatio := balanceMgr.GetDepositRate(&addrCoin)
	reward := balanceMgr.GetAddrReward(addrCoin)

	witnessName := ""
	witnessAddr := ""
	if witInfo := balanceMgr.GetDepositWitness(&commInfo.WitnessAddr); witInfo != nil {
		witnessName = witInfo.Name
		witnessAddr = witInfo.SelfAddr.B58String()
	}

	commFrozenReward, _ := balanceMgr.CalculateCommunityRewardAndLightReward(addrCoin)
	contract := ""
	//if config.EVM_Reward_Enable {
	//	contract = precompiled.RewardContract.B58String()
	//}
	data := CommunityInfoV0{
		Deposit:      commInfo.Value,
		Vote:         balanceMgr.GetCommunityVote(&addrCoin),
		LightCount:   lightCount,
		RewardRatio:  float64(rewardRatio),
		Reward:       reward.Uint64(),
		WitnessName:  witnessName,
		WitnessAddr:  witnessAddr,
		StartHeight:  commInfo.Height,
		FrozenReward: commFrozenReward.Uint64(),
		Contract:     contract,
		Name:         commInfo.Name,
		LightNode:    []LightNode{},
	}

	cvote := balanceMgr.GetCommunityVote(&commInfo.SelfAddr)
	lightNodes := []LightNode{}
	for _, lAddr := range lightAddrs {
		reward := balanceMgr.GetAddrReward(lAddr)
		lightvoteinfo := balanceMgr.GetDepositVote(&lAddr)

		lightName := ""
		lightDeposit := toUint64(0)
		if lightinfo := balanceMgr.GetDepositLight(&lAddr); lightinfo != nil {
			lightName = lightinfo.Name
			lightDeposit = lightinfo.Value
		}
		//兼容老版本,老版本放大了100000倍
		ratio := decimal.NewFromInt(int64(lightvoteinfo.Value)).DivRound(decimal.NewFromInt(int64(cvote)), 3).Mul(decimal.NewFromInt(100000))
		ln := LightNode{
			Addr:            lAddr.B58String(),
			Reward:          reward.Uint64(),
			RewardRatio:     ratio.InexactFloat64(),
			VoteNum:         lightvoteinfo.Value,
			Name:            lightName,
			Deposit:         lightDeposit,
			LastVotedHeight: balanceMgr.GetLastVoteOp(lAddr),
		}
		lightNodes = append(lightNodes, ln)
	}
	sort.Sort(LightNodeSort(lightNodes))

	if start, end, ok := helppager(len(lightNodes), pageInt, pageSizeInt); ok {
		out := make([]LightNode, end-start)
		copy(out, lightNodes[start:end])
		data.LightNode = out
	} else {
		data.LightNode = []LightNode{}
	}

	res, err = model.Tojson(data)
	return
}

type WitnessInfoV0 struct {
	Vote           uint64          `json:"vote"`             //总票数
	Deposit        uint64          `json:"deposit"`          // 总的质押数
	AddBlockCount  uint64          `json:"add_block_count"`  // 出块数
	AddBlockReward uint64          `json:"add_block_reward"` // 出块奖励
	RewardRatio    float64         `json:"reward_ratio"`     // 奖励比例
	TotalReward    uint64          `json:"total_reward"`     // 累计获得奖励
	FrozenReward   uint64          `json:"frozen_reward"`    //未提现奖励
	CommunityCount uint64          `json:"community_count"`  // 社区数量
	Name           string          `json:"name"`             // 见证者节点名称
	CommunityNode  []CommunityNode `json:"community_node"`
	DestroyNum     uint64          `json:"destroy_num"` //销毁数量
}

type CommunityNode struct {
	Name        string  `json:"name"`         // 名字
	Addr        string  `json:"addr"`         // 地址
	Deposit     uint64  `json:"deposit"`      // 质押量
	Reward      uint64  `json:"reward"`       // 奖励
	LightNum    uint64  `json:"light_num"`    // 轻节点数量(社区人数)
	VoteNum     uint64  `json:"vote_num"`     // 轻节点投票总数
	RewardRatio float64 `json:"reward_ratio"` // 奖励比例
}

type CommunityNodeSort []CommunityNode

func (c CommunityNodeSort) Len() int {
	return len(c)
}
func (c CommunityNodeSort) Less(i, j int) bool {
	return c[i].VoteNum > c[j].VoteNum
}
func (c CommunityNodeSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type CommunityInfoV0 struct {
	Deposit      uint64      `json:"deposit"`       // 总的质押数
	Vote         uint64      `json:"vote"`          //总票数
	LightCount   uint64      `json:"light_count"`   // 投票成员
	RewardRatio  float64     `json:"reward_ratio"`  // 奖励比例
	Reward       uint64      `json:"reward"`        // 累计获得奖励
	WitnessName  string      `json:"witness_name"`  // 见证人名称
	WitnessAddr  string      `json:"witness_addr"`  // 见证人地址
	StartHeight  uint64      `json:"start_height"`  // 成为社区节点的高度
	FrozenReward uint64      `json:"frozen_reward"` // 未提现奖励
	Contract     string      `json:"contract"`      // 提现合约地址
	Name         string      `json:"name"`          // 社区节点名称
	LightNode    []LightNode `json:"light_node"`
}

type LightNode struct {
	Addr            string  `json:"addr"`              // 地址
	Reward          uint64  `json:"reward"`            // 奖励
	RewardRatio     float64 `json:"reward_ratio"`      // 奖励比例
	VoteNum         uint64  `json:"vote_num"`          // 轻节点投票总数
	Deposit         uint64  `json:"deposit"`           // 质押量
	Name            string  `json:"name"`              //轻节点名字
	LastVotedHeight uint64  `json:"last_voted_height"` //轻节点最后投票高度
}

type LightNodeSort []LightNode

func (c LightNodeSort) Len() int {
	return len(c)
}
func (c LightNodeSort) Less(i, j int) bool {
	return c[i].VoteNum > c[j].VoteNum
}
func (c LightNodeSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func GetRewardHistory(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := toUint64(startHeightItr.(float64))
	if startHeight <= 1 {
		res, err = model.Errcode(model.Nomarl, "起始高度至少大于1")
		return
	}
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
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(0)
	}
	pageSizeInt := int(pageSize.(float64))
	if pageSizeInt > pageSizeLimit {
		pageSizeInt = pageSizeLimit
	}
	rewardLogs := []precompiled.LogRewardHistoryV0{}
	//待返回的区块
	for i := startHeight; i <= endHeight; i++ {

		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}
		//bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))
		if len(bh.Tx) > 0 {
			rewardTx, e := mining.LoadTxBase(bh.Tx[0])
			if e != nil {
				res, err = model.Errcode(model.Nomarl, e.Error())
				return
			}
			if rewardTx.Class() == config.Wallet_tx_type_mining {
				logs := precompiled.GetRewardHistoryLog(i, *rewardTx.GetHash())
				rewardLogs = append(rewardLogs, logs...)
			}

		}

	}
	if pageSizeInt == 0 {
		pageSizeInt = len(rewardLogs)
	}
	data := make(map[string]interface{})
	start := (pageInt - 1) * pageSizeInt
	end := start + pageSizeInt
	if start > len(rewardLogs) {
		data["total"] = 0
		data["list"] = []precompiled.LogRewardHistoryV0{}
		res, err = model.Tojson(data)
		return
	}
	if end > len(rewardLogs) {
		end = len(rewardLogs)
	}
	newList := []precompiled.LogRewardHistoryV0{}
	for _, v := range rewardLogs[start:end] {
		if v.From == "" {

			addrCoin := crypto.AddressFromB58String(v.Into)
			witnessName := mining.FindWitnessName(addrCoin)
			v.Name = witnessName
		}
		newList = append(newList, v)
	}
	data["total"] = len(rewardLogs)
	data["list"] = newList
	res, err = model.Tojson(data)
	return
}

// 返回交易类型，vin,vout
// func DealTxInfo(tx mining.TxItr, addr string) (uint64, interface{}, interface{}) {
// //tx, _, _ := mining.FindTx(txid)
// vouts := tx.GetVout()
// out := *vouts

// //如果是和奖励合约交互的交易
// if bytes.Equal(out[0].Address, precompiled.RewardContract) {

// 	//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
// 	if tx.Class() == config.Wallet_tx_type_mining {
// 		rewardLogs := precompiled.GetRewardHistoryLog(tx.GetLockHeight(), *tx.GetHash())
// 		for _, v := range rewardLogs {
// 			if v.Into == addr {
// 				returnVins := []mining.VinVO{}
// 				returnVins = append(returnVins, mining.VinVO{
// 					Addr: precompiled.RewardContract.B58String(),
// 				})
// 				returnVouts := []mining.VoutVO{}
// 				returnVouts = append(returnVouts, mining.VoutVO{
// 					Address: addr,
// 					Value:   v.Reward.Uint64(),
// 				})
// 				return config.Wallet_tx_type_reward_W, returnVins, returnVouts
// 			}
// 		}
// 	}
// 	//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
// 	if tx.Class() == config.Wallet_tx_type_contract {
// 		payload := tx.GetPayload()
// 		//解析
// 		txClass, params := precompiled.UnpackPayload(payload)
// 		switch txClass {
// 		case config.Wallet_tx_type_community_in:
// 			//社区节点质押,只需要变更类型
// 			return config.Wallet_tx_type_community_in, nil, nil
// 		case config.Wallet_tx_type_community_out:
// 			//社区节点取消质押
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   config.Mining_vote,
// 			})
// 			return config.Wallet_tx_type_community_out, returnVins, returnVouts
// 		case config.Wallet_tx_type_vote_in:
// 			//轻节点投票
// 			return config.Wallet_tx_type_vote_in, nil, nil
// 		case config.Wallet_tx_type_vote_out:
// 			//轻节点取消投票
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_vote_out, returnVins, returnVouts
// 		case config.Wallet_tx_type_light_in:
// 			//轻节点质押
// 			return config.Wallet_tx_type_light_in, nil, nil
// 		case config.Wallet_tx_type_light_out:
// 			//轻节点取消质押
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   config.Mining_light_min,
// 			})
// 			return config.Wallet_tx_type_light_out, returnVins, returnVouts
// 		case config.Wallet_tx_type_reward_C:
// 			//提现奖励
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_reward_C, returnVins, returnVouts
// 		case config.Wallet_tx_type_reward_L:
// 			//提现奖励
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_reward_L, returnVins, returnVouts
// 		case config.Wallet_tx_type_reward_W:
// 			//提现奖励
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_reward_W, returnVins, returnVouts
// 		case config.Wallet_tx_type_domain_register: // 域名注册
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_domain_register, returnVins, returnVouts
// 		case config.Wallet_tx_type_domain_renew: // 域名续费
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_domain_renew, returnVins, returnVouts
// 		case config.Wallet_tx_type_domain_withdraw: // 域名体现
// 			returnVins := []mining.VinVO{}
// 			returnVins = append(returnVins, mining.VinVO{
// 				Addr: precompiled.RewardContract.B58String(),
// 			})
// 			returnVouts := []mining.VoutVO{}
// 			returnVouts = append(returnVouts, mining.VoutVO{
// 				Address: addr,
// 				Value:   params,
// 			})
// 			return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts
// 		case config.Wallet_tx_type_community_distribute: //社区分账
// 			rewardLogs := precompiled.GetRewardHistoryLog(tx.GetLockHeight(), *tx.GetHash())
// 			for _, v := range rewardLogs {
// 				if v.Into == addr {
// 					returnVins := []mining.VinVO{}
// 					returnVins = append(returnVins, mining.VinVO{
// 						Addr: precompiled.RewardContract.B58String(),
// 					})
// 					returnVouts := []mining.VoutVO{}
// 					returnVouts = append(returnVouts, mining.VoutVO{
// 						Address: addr,
// 						Value:   v.Reward.Uint64(),
// 					})

// 					if v.Utype == 2 { // 社区
// 						return config.Wallet_tx_type_reward_C, returnVins, returnVouts
// 					} else if v.Utype == 3 { // 轻节点
// 						return config.Wallet_tx_type_reward_L, returnVins, returnVouts
// 					}
// 				}
// 			}
// 		}
// 	}
// }
// 	return 0, nil, nil
// }

// 返回交易类型,vin,vout
// 通过实际高度解析历史交易
func DealTxInfo(tx mining.TxItr, addr string, blockheight uint64) (uint64, interface{}, interface{}) {
	if tx == nil {
		return 0, nil, nil
	}

	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins

	if addr == "" {
		if len(vin) > 0 {
			addr = vin[0].PukToAddr.B58String()
		}
	}

	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {
		payload := tx.GetPayload()
		//解析
		txClass, params := precompiled.UnpackPayload(payload)
		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		// if tx.Class() == config.Wallet_tx_type_mining {
		// 	rewardLogs := precompiled.GetRewardHistoryLog(blockheight, *tx.GetHash())
		// 	for _, v := range rewardLogs {
		// 		if v.Into == addr {
		// 			returnVins := []mining.VinVO{}
		// 			returnVins = append(returnVins, mining.VinVO{
		// 				Addr: precompiled.RewardContract.B58String(),
		// 			})
		// 			returnVouts := []mining.VoutVO{}
		// 			returnVouts = append(returnVouts, mining.VoutVO{
		// 				Address: addr,
		// 				Value:   v.Reward.Uint64(),
		// 			})
		// 			return config.Wallet_tx_type_reward_W, returnVins, returnVouts
		// 		}
		// 	}
		// }

		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts
			case config.Wallet_tx_type_community_distribute: //社区分账
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_community_distribute, returnVins, returnVouts
			}
		}

	}

	// 根域名注册合约交互的交易
	if bytes.Equal(out[0].Address, ens.GetRegisterAddr()) {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_domain_register: // 域名注册
				return config.Wallet_tx_type_domain_register, nil, nil
			case config.Wallet_tx_type_domain_renew: // 域名续费
				return config.Wallet_tx_type_domain_renew, nil, nil
			case config.Wallet_tx_type_domain_withdraw: // 域名提现
				rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
				for _, v := range rewardLogs {
					if v.Into == addr {
						returnVins := []mining.VinVO{}
						returnVins = append(returnVins, mining.VinVO{
							Addr: precompiled.RewardContract.B58String(),
						})
						returnVouts := []mining.VoutVO{}
						returnVouts = append(returnVouts, mining.VoutVO{
							Address: addr,
							Value:   v.Amount.Uint64(),
						})

						return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts
					}
				}
			}
		}
	}

	// 剩下的可能是子域名合约的交易
	if tx.Class() == config.Wallet_tx_type_contract {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		switch txClass {
		case config.Wallet_tx_type_domain_register: // 子域名注册
			return config.Wallet_tx_type_subdomain_register, nil, nil
		case config.Wallet_tx_type_domain_renew: // 子域名续费
			return config.Wallet_tx_type_subdomain_renew, nil, nil
		case config.Wallet_tx_type_domain_withdraw: // 子域名提现
			rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
			for _, v := range rewardLogs {
				if v.Into == addr {
					returnVins := []mining.VinVO{}
					returnVins = append(returnVins, mining.VinVO{
						Addr: precompiled.RewardContract.B58String(),
					})
					returnVouts := []mining.VoutVO{}
					returnVouts = append(returnVouts, mining.VoutVO{
						Address: addr,
						Value:   v.Amount.Uint64(),
					})

					return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts
				}
			}
		}
	}

	return 0, nil, nil
}

// 返回交易类型,vin,vout
// 通过实际高度解析历史交易
// NOTE: 特殊处理了域名提现
func DealTxInfoV3(tx mining.TxItr, addr string, blockheight uint64) (uint64, interface{}, interface{}, bool) {
	if tx == nil {
		return 0, nil, nil, false
	}

	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins

	if addr == "" {
		if len(vin) > 0 {
			addr = vin[0].PukToAddr.B58String()
		}
	}

	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {
		payload := tx.GetPayload()
		//解析
		txClass, params := precompiled.UnpackPayload(payload)
		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		if tx.Class() == config.Wallet_tx_type_mining {
			returnVins := []mining.VinVO{}
			returnVins = append(returnVins, mining.VinVO{
				// Addr: precompiled.RewardContract.B58String(),
				Addr: "",
			})
			return config.Wallet_tx_type_mining, returnVins, nil, false
		}

		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil, false
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil, false
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil, false
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts, false
			case config.Wallet_tx_type_community_distribute: //社区分账
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_community_distribute, returnVins, returnVouts, false
			}
		}

	}

	// 根域名注册合约交互的交易
	if bytes.Equal(out[0].Address, ens.GetRegisterAddr()) {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_domain_register: // 域名注册
				return config.Wallet_tx_type_domain_register, nil, nil, true
			case config.Wallet_tx_type_domain_renew: // 域名续费
				return config.Wallet_tx_type_domain_renew, nil, nil, true
			case config.Wallet_tx_type_domain_withdraw: // 域名提现
				rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
				//for _, v := range rewardLogs {
				//	if v.Into == addr {
				//		returnVins := []mining.VinVO{}
				//		returnVins = append(returnVins, mining.VinVO{
				//			Addr: precompiled.RewardContract.B58String(),
				//		})
				//		returnVouts := []mining.VoutVO{}
				//		returnVouts = append(returnVouts, mining.VoutVO{
				//			Address: addr,
				//			Value:   v.Amount.Uint64(),
				//		})

				//		return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts, true
				//	}
				//}
				returnVins := []mining.VinVO{}
				returnVouts := []mining.VoutVO{}
				for _, v := range rewardLogs {
					returnVins = append(returnVins, mining.VinVO{
						Addr: v.From,
					})
					returnVouts = append(returnVouts, mining.VoutVO{
						Address: v.Into,
						Value:   v.Amount.Uint64(),
					})
				}

				return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts, true
			}
		}
	}

	// 剩下的可能是子域名合约的交易
	if tx.Class() == config.Wallet_tx_type_contract {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		switch txClass {
		case config.Wallet_tx_type_domain_register: // 子域名注册
			return config.Wallet_tx_type_subdomain_register, nil, nil, true
		case config.Wallet_tx_type_domain_renew: // 子域名续费
			return config.Wallet_tx_type_subdomain_renew, nil, nil, true
		case config.Wallet_tx_type_domain_withdraw: // 子域名提现
			rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
			//for _, v := range rewardLogs {
			//	if v.Into == addr {
			//		returnVins := []mining.VinVO{}
			//		returnVins = append(returnVins, mining.VinVO{
			//			Addr: precompiled.RewardContract.B58String(),
			//		})
			//		returnVouts := []mining.VoutVO{}
			//		returnVouts = append(returnVouts, mining.VoutVO{
			//			Address: addr,
			//			Value:   v.Amount.Uint64(),
			//		})

			//		return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts, true
			//	}
			//}
			returnVins := []mining.VinVO{}
			returnVouts := []mining.VoutVO{}
			for _, v := range rewardLogs {
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   v.Amount.Uint64(),
				})
			}
			return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts, true
		}
	}

	return tx.Class(), nil, nil, false
}

// 返回交易类型，vin,vout
func DealTxInfoV2(tx mining.TxItr) (uint64, interface{}, interface{}) {
	//tx, _, _ := mining.FindTx(txid)
	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins
	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {

		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		//if tx.Class() == config.Wallet_tx_type_mining {
		//	returnVins := []mining.VinVO{}
		//	returnVins = append(returnVins, mining.VinVO{
		//		Addr: precompiled.RewardContract.B58String(),
		//	})
		//	returnVouts := []mining.VoutVO{}
		//	rewardLogs := precompiled.GetRewardHistoryLog(tx.GetLockHeight(), *tx.GetHash())
		//	for _, v := range rewardLogs {
		//		if v.From == "" {
		//			returnVouts = append(returnVouts, mining.VoutVO{
		//				Address: v.Into,
		//				Value:   v.Reward.Uint64(),
		//			})
		//		}
		//	}
		//	return config.Wallet_tx_type_reward_W, returnVins, returnVouts
		//}
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			payload := tx.GetPayload()
			//解析
			txClass, params := precompiled.UnpackPayload(payload)
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts
			}
		}
	}
	return 0, nil, nil
}

// 查询多地址账户信息
func HandleMultiAccounts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	tokenId := []byte{}
	tokenidItr, ok := rj.Get("token_id")
	if ok {
		if !rj.VerifyType("token_id", "string") {
			res, err = model.Errcode(model.TypeWrong, "token_id")
			return
		}
		tokenidStr := tokenidItr.(string)
		tokenId, err = hex.DecodeString(tokenidStr)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, "decode error")
			return
		}
	}

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

	if len(addresses) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	vos := make([]AccountVO, 0)
	for i, val := range addresses {
		addr := crypto.AddressFromB58String(val)
		var ba, fba, baLockup uint64
		ty := mining.GetAddrState(addr)
		if len(tokenId) == 0 {
			ba, fba, baLockup = mining.GetBalanceForAddrSelf(addr)
		} else {
			ba, fba, baLockup = mining.GetTokenNotSpendAndLockedBalance(tokenId, addr)
			ty = 4
		}

		mainAddr := ""
		if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(addr)); err == nil {
			a := crypto.AddressCoin(bs)
			mainAddr = a.B58String()
		}

		subAddrs := []string{}
		if pairs, err := db.LevelDB.HGetAll(addr); err == nil {
			for _, pair := range pairs {
				a := crypto.AddressCoin(pair.Field)
				subAddrs = append(subAddrs, a.B58String())
			}
		}

		vo := AccountVO{
			Index:               i,
			Name:                GetNickName(addr),
			AddrCoin:            val,
			MainAddrCoin:        mainAddr,
			SubAddrCoins:        subAddrs,
			Type:                ty,
			Value:               ba,       //可用余额
			ValueFrozen:         fba,      //冻结余额
			ValueLockup:         baLockup, //
			BalanceVote:         GetBalanceVote(addr),
			AddressFrozenStatus: mining.CheckAddressFrozenStatus(addr),
		}
		vos = append(vos, vo)
	}
	return model.Tojson(vos)
}

// 查询多地址账户余额
func MultiBalance(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	if len(addresses) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	type AccountVO struct {
		Index    int    //排序
		AddrCoin string //收款地址
		Total    uint64 //总余额（可用余额+锁定余额）
	}

	vos := make([]AccountVO, 0)
	for i, val := range addresses {
		addr := crypto.AddressFromB58String(val)
		var ba uint64
		_, ba = db.GetNotSpendBalance(&addr)

		vo := AccountVO{
			Index:    i,
			AddrCoin: val,
			Total:    ba,
		}
		vos = append(vos, vo)
	}
	return model.Tojson(vos)
}

/*
推送多签交易
*/
func PushMultTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	payload, res, err := getParamString(rj, "payload")
	if err != nil {
		return res, err
	}

	txjsonBs, e := base64.StdEncoding.DecodeString(payload)
	if e != nil {
		engine.Log.Info("DecodeString fail:%s", e.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}

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

	gas := txItr.GetGas()
	amount := uint64(0)
	for _, one := range *txItr.GetVout() {
		amount += one.Value
	}

	comment := string(txItr.GetComment())
	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		res, err = model.Errcode(CommentOverLengthMax, "comment")
		return
	}
	vin := (*txItr.GetVin())[0]
	multAddr := vin.GetPukToAddr()
	if e := mining.CheckTxPayFreeGasWithParams(config.Wallet_tx_type_multsign_pay, *multAddr, amount, gas, comment); e != nil {
		res, err = model.Errcode(GasTooLittle, "gas too little")
		return
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

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(src, spend)
	if total < spend {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	//验证domain
	if !txItr.CheckDomain() {
		return model.Errcode(model.Nomarl, "domain name resolution failed")
	}

	//处理多签请求
	if mtx, ok := txItr.(mining.DefaultMultTx); ok {
		if err := mtx.RequestMultTx(txItr); err != nil {
			engine.Log.Warn("Multsign Request: %s %s", hex.EncodeToString(*txItr.GetHash()), err.Error())
			return model.Errcode(model.Nomarl, err.Error())
		}
	}

	res, err = model.Tojson("success")
	return
}

// 根据区块高度范围和合约地址，过滤并获取合约事件
func GetContractEventFilter(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	fromItr, ok := rj.Get("from")
	if !ok {
		res, err = model.Errcode(model.NoField, "from")
		return
	}
	if !rj.VerifyType("from", "float64") {
		res, err = model.Errcode(model.TypeWrong, "from")
		return
	}

	var to uint64
	toItr, ok := rj.Get("to")
	if ok {
		if !rj.VerifyType("to", "float64") {
			res, err = model.Errcode(model.TypeWrong, "to")
			return
		}
		to = toUint64(toItr.(float64))
	} else {
		to = mining.GetLongChain().GetCurrentBlock()
	}

	addrItr, ok := rj.Get("contractAddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractAddress")
		return
	}
	contractAddr := crypto.AddressFromB58String(addrItr.(string))
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractAddress")
		return
	}

	eventIdItr, ok := rj.Get("eventId")
	if !ok {
		res, err = model.Errcode(model.NoField, "eventId")
		return
	}
	eventId := eventIdItr.(string)
	if strings.HasPrefix(eventId, "0x") || strings.HasPrefix(eventId, "0X") {
		eventId = eventId[2:]
	}
	eventIdBs := []byte(eventId)

	from := toUint64(fromItr.(float64))
	if to > mining.GetLongChain().GetCurrentBlock() {
		to = mining.GetLongChain().GetCurrentBlock()
	}
	if from > to {
		return model.Errcode(model.Nomarl, "to out of range")
	}

	result := make([]*go_protos.ContractEventInfo, 0)

	//通过区块bloom过滤事件，找到包含topic的区块高度
	heights := mining.FilterBlockHeadBloomTopic(from, to, eventIdBs)
	if len(heights) == 0 {
		return model.Tojson(result)
	}

	//记录区块高度key集合
	bhHeights := make([][]byte, 0)
	for _, i := range heights {
		//bhHeights = append(bhHeights, []byte(config.BlockHeight+strconv.Itoa(int(i))))
		bhHeights = append(bhHeights, append(config.BlockHeight, []byte(strconv.Itoa(int(i)))...))
	}

	//通过区块高度批量查询查询所有hash
	bhHashs, err := db.LevelDB.MGet(bhHeights...)
	if err != nil {
		return model.Errcode(model.Nomarl, "db not found data")
	}

	//区块hashkey添加进前缀
	for k := range bhHashs {
		bhHashs[k] = config.BuildBlockHead(bhHashs[k])
	}

	//获取区块头集合
	bhs, err := db.LevelDB.MGet(bhHashs...)
	if err != nil {
		return model.Errcode(model.Nomarl, "db not found data")
	}

	//记录多区块多个交易pair，用于hash批量查询
	kfs := make([]db2.KFPair, 0)
	//遍历查询到的所有区块hash
	for k := range bhs {
		if bhs[k] == nil || len(bhs[k]) == 0 {
			break
		}
		bh, err := mining.ParseBlockHeadProto(&bhs[k])
		if err != nil {
			break
		}
		if len(bh.Tx) == 0 {
			continue
		}

		txs := make([][]byte, 0)
		for i := range bh.Tx {
			if bh.Tx[i][0] != config.Wallet_tx_type_contract && bh.Tx[i][0] != config.Wallet_tx_type_mining {
				continue
			}
			//获取交易
			//tx, err := mining.LoadTxBase(bh.Tx[i])
			//if err != nil {
			//	continue
			//}
			//
			//if !mining.BytesToBloom(tx.GetBloom()).Check(eventIdBs) {
			//	continue
			//}

			//过滤交易bloom
			if !mining.BytesToBloom(mining.GetBloomByTx(bh.Tx[i])).Check(eventIdBs) {
				continue
			}

			//添加待查询交易到切片
			txs = append(txs, bh.Tx[i])
		}
		key := append([]byte(config.DBKEY_BLOCK_EVENT), new(big.Int).SetUint64(bh.Height).Bytes()...)
		kfs = append(kfs, db2.KFPair{key, txs})
	}

	//批量查询多区块多交易
	values, err := db.LevelDB.HKMget(kfs)
	if err != nil {
		return model.Errcode(model.Nomarl, "db error")
	}

	//解析并过滤事件
	for i := range values {
		for j := range values[i] {
			list := &go_protos.ContractEventInfoList{}
			err = proto.Unmarshal(values[i][j], list)
			if err != nil {
				continue
			}
			for _, v := range list.ContractEvents {
				//过滤合约名
				if v.ContractAddress != addrItr.(string) {
					continue
				}
				//匹配主题和事件ID
				if v.Topic == eventId {
					result = append(result, v)
				}
			}
		}
	}
	return model.Tojson(result)
}

// 获取rpc字符串数组参数
func getArrayStrParams(rj *model.RpcJson, name string) ([]string, int) {
	itemItrs, ok := rj.Get(name)
	if !ok {
		return nil, model.NoField
	}

	bs, err := json.Marshal(itemItrs)
	if err != nil {
		return nil, model.TypeWrong
	}
	addresses := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	err = decoder.Decode(&addresses)
	if err != nil {
		return nil, model.TypeWrong
	}

	return addresses, 0
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

// 解析一个真实交易,同时过滤地址
func parseTxAndCustomTx(txid []byte, addrs ...string) map[string]interface{} {
	//isCustomTx := false
	//realtxid := []byte{}
	txItr, code, blockHash := mining.FindTx(txid)
	if txItr == nil {
		////特殊处理自定义交易
		//realtxid = bytes.TrimPrefix(txid, []byte{config.Wallet_tx_type_voting_reward})
		//txItr, code, blockHash = mining.FindTx(realtxid)
		//if txItr == nil {
		//	return nil
		//}
		//isCustomTx = true
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

	addrm := make(map[string]struct{})
	for _, addr := range addrs {
		addrm[addr] = struct{}{}
	}

	var item map[string]interface{}

	//特殊处理矿工交易,vout含有所有见证人交易记录,去除非查询地址的记录
	if txItr.Class() == config.Wallet_tx_type_mining {
		if txReward, ok := txItr.(*mining.Tx_reward); ok {
			newVouts := []*mining.VoutVO{}
			for _, vout := range txReward.Vout {
				if _, ok := addrm[vout.Address.B58String()]; ok {
					newVouts = append(newVouts, &mining.VoutVO{
						Value:        vout.Value,
						Address:      vout.Address.B58String(),
						FrozenHeight: vout.FrozenHeight,
					})
				}
			}

			item = JsonMethod(txReward.GetVOJSON())
			if len(newVouts) > 0 {
				item["vout"] = newVouts
			}
			item["is_custom_tx"] = true
		}
	}

	//if isCustomTx {
	//	switch txItr.Class() {
	//	case config.Wallet_tx_type_vote_in:
	//		//轻节点投票,伪造一笔发奖励的交易
	//		voteIn := txItr.(*mining.Tx_vote_in)
	//		item = JsonMethod(voteIn.GetVOJSON())
	//		item["type"] = config.Wallet_tx_type_voting_reward
	//		item["hash"] = hex.EncodeToString(txid)
	//		item["is_custom_tx"] = true
	//		if voteIn.VoteType == mining.VOTE_TYPE_vote {
	//			if vouts, err := parseCustomTxEvents2VoutV0s(realtxid, addrm); err == nil {
	//				item["vout"] = vouts
	//			}
	//		}
	//	case config.Wallet_tx_type_vote_out:
	//		//社区取消质押,轻节点取消投票
	//		voteOut := txItr.(*mining.Tx_vote_out)
	//		item = JsonMethod(voteOut.GetVOJSON())
	//		item["type"] = config.Wallet_tx_type_voting_reward
	//		item["hash"] = hex.EncodeToString(txid)
	//		item["is_custom_tx"] = true
	//		if voteOut.VoteType == mining.VOTE_TYPE_vote || voteOut.VoteType == mining.VOTE_TYPE_community {
	//			if vouts, err := parseCustomTxEvents2VoutV0s(realtxid, addrm); err == nil {
	//				item["vout"] = vouts
	//			}
	//		}
	//	case config.Wallet_tx_type_voting_reward:
	//		//手动分账
	//		voteReward := txItr.(*mining.Tx_Vote_Reward)
	//		item = JsonMethod(voteReward.GetVOJSON())
	//		item["type"] = config.Wallet_tx_type_voting_reward
	//		item["hash"] = hex.EncodeToString(txid)
	//		item["is_custom_tx"] = true
	//		if vouts, err := parseCustomTxEvents2VoutV0s(realtxid, addrm); err == nil {
	//			item["vout"] = vouts
	//		}
	//	}
	//}

	if item == nil {
		item = JsonMethod(txItr.GetVOJSON())
	}

	out := make(map[string]interface{})
	item["blockhash"] = blockHashStr
	item["timestamp"] = bh.Time
	out["txinfo"] = item
	out["timestamp"] = bh.Time
	out["blockhash"] = blockHashStr
	out["blockheight"] = bh.Height
	out["upchaincode"] = code

	return out
}

// 解析一个真实交易和多个自定义交易,同时过滤地址
//func parseOneTxAndMoreCustomTx(txid []byte, addrs ...string) []map[string]interface{} {
//	txItr, code, blockHash := mining.FindTx(txid)
//	if txItr == nil {
//		return nil
//	}
//
//	if blockHash == nil {
//		return nil
//	}
//
//	//获取区块信息
//	bh, err := mining.LoadBlockHeadByHash(blockHash)
//	if err != nil {
//		return nil
//	}
//
//	var blockHashStr string
//	if blockHash != nil {
//		blockHashStr = hex.EncodeToString(*blockHash)
//	}
//
//	//vins := *(txItr.GetVin())
//	//vinAddress := ""
//	//if len(vins) > 0 {
//	//	vinAddress = vins[0].GetPukToAddr().B58String()
//	//}
//
//	addrsm := make(map[string]struct{})
//	for _, addr := range addrs {
//		addrsm[addr] = struct{}{}
//	}
//
//	items := make([]map[string]interface{}, 0)
//	if len(addrs) > 0 {
//		switch txItr.Class() {
//		//特殊处理矿工交易,vout含有所有见证人交易记录,去除非查询地址的记录
//		case config.Wallet_tx_type_mining:
//			if txReward, ok := txItr.(*mining.Tx_reward); ok {
//				newVins := []*mining.Vin{}
//				newVouts := []*mining.Vout{}
//				for i, vout := range txReward.Vout {
//					if _, ok := addrsm[vout.Address.B58String()]; ok {
//						newVouts = append(newVouts, txReward.Vout[i])
//					}
//				}
//				if len(newVouts) > 0 {
//					txReward.Vin = newVins
//					txReward.Vout = newVouts
//				}
//
//				items = append(items, JsonMethod(txReward.GetVOJSON()))
//			}
//		case config.Wallet_tx_type_vote_in:
//			//轻节点投票,伪造一笔发奖励的交易
//			voteIn := txItr.(*mining.Tx_vote_in)
//			if voteIn.VoteType == mining.VOTE_TYPE_vote {
//				if vouts, err := parseCustomTxEvents2VoutV0s(txid, addrsm); err == nil {
//					voteIn.Vout = append(voteIn.Vout, vouts...)
//				}
//			}
//			items = append(items, JsonMethod(txItr.GetVOJSON()))
//			items = append(items, JsonMethod(voteIn.GetVOJSON()))
//		case config.Wallet_tx_type_vote_out:
//			//社区取消质押,轻节点取消投票
//			voteOut := txItr.(*mining.Tx_vote_out)
//			if voteOut.VoteType == mining.VOTE_TYPE_vote || voteOut.VoteType == mining.VOTE_TYPE_community {
//				if vouts, err := parseCustomTxEvents2VoutV0s(txid, addrsm); err == nil {
//					voteOut.Vout = append(voteOut.Vout, vouts...)
//				}
//			}
//
//			items = append(items, JsonMethod(txItr.GetVOJSON()))
//			items = append(items, JsonMethod(voteOut.GetVOJSON()))
//		case config.Wallet_tx_type_voting_reward:
//			//手动分账
//			voteReward := txItr.(*mining.Tx_Vote_Reward)
//			if vouts, err := parseCustomTxEvents2VoutV0s(txid, addrsm); err == nil {
//				voteReward.Vout = append(voteReward.Vout, vouts...)
//			}
//
//			items = append(items, JsonMethod(txItr.GetVOJSON()))
//			items = append(items, JsonMethod(voteReward.GetVOJSON()))
//		}
//	}
//
//	if len(items) == 0 {
//		items = append(items, JsonMethod(txItr.GetVOJSON()))
//	}
//
//	out := make([]map[string]interface{}, 0)
//	for i, _ := range items {
//		items[i]["blockhash"] = blockHashStr
//		items[i]["timestamp"] = bh.Time
//
//		tmp := make(map[string]interface{})
//		tmp["txinfo"] = items[i]
//		tmp["timestamp"] = bh.Time
//		tmp["blockhash"] = blockHashStr
//		tmp["blockheight"] = bh.Height
//		tmp["upchaincode"] = code
//
//		out = append(out, tmp)
//	}
//
//	return out
//}

func toUint64(in float64) uint64 {
	if in > 0 {
		return uint64(in)
	} else {
		return 0
	}
}

func GetNickName(addr crypto.AddressCoin) string {
	if addrInfo, ok := config.Area.Keystore.FindAddress(addr); ok {
		return addrInfo.Nickname
	}

	return ""
}

func AllAddrBalance(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	allMainCoin, _, _ := getParamBool(rj, "allmaincoin")
	lhbalancesItr, ok := rj.Get("lhbalances")
	if !ok && !allMainCoin {
		res, err = model.Errcode(model.NoField, "lhbalances")
		return
	}
	lHBalances := []*mining.LHBalance{}
	if b, e := json.Marshal(lhbalancesItr); e == nil {
		json.Unmarshal(b, &lHBalances)
	}
	param := &mining.LHParams{
		AllMainCoin: allMainCoin,
		LHBalances:  lHBalances,
	}
	data := map[string]interface{}{}
	height, items := mining.LockHeightAllAddrBalance(param)
	data["height"] = height
	lHBalances = []*mining.LHBalance{}
	for _, item := range items {
		lHBalances = append(lHBalances, &mining.LHBalance{
			Address:         item.Address,
			ContractAddress: item.ContractAddress,
			Value:           item.Value,
		})
	}
	data["lhbalances"] = lHBalances
	return model.Tojson(data)
}

func DepositFreeGas(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	srcStr, res, err := getParamString(rj, "src")
	if err != nil {
		return res, err
	}
	src := crypto.AddressFromB58String(srcStr)
	gas, res, err := getParamUint64(rj, "gas")
	if err != nil {
		return res, err
	}
	pwd, res, err := getParamString(rj, "pwd")
	if err != nil {
		return res, err
	}
	depositAddressStr, res, err := getParamString(rj, "deposit_address")
	if err != nil {
		return res, err
	}
	depositAddress := crypto.AddressFromB58String(depositAddressStr)
	deposit, res, err := getParamUint64(rj, "deposit")
	if err != nil {
		return res, err
	}
	tx, err := mining.BuildDepositFreeGasTx(src, gas, pwd, depositAddress, deposit)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson(tx.GetVOJSON())
	return
}

func ListDepositFreeGas(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	type item struct {
		ContractAddresses string
		Owner             string
		Deposit           uint64
		LimitHeight       uint64
		LimitCount        uint64
	}
	data := []*item{}
	items := mining.GetLongChain().GetBalance().ListFreeGasAddrSet()
	for _, v := range items {
		data = append(data, &item{
			ContractAddresses: v.ContractAddresses.B58String(),
			Owner:             v.Owner.B58String(),
			Deposit:           v.Deposit,
			LimitHeight:       v.LimitHeight,
			LimitCount:        v.LimitCount,
		})
	}

	res, err = model.Tojson(data)
	return
}

func GetInternalTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	list := &[]storage.InternalTx{}

	hash, ok := rj.Get("hash")
	if !ok {
		res, err = model.Errcode(model.NoField, "hash")
		return
	}

	if !rj.VerifyType("hash", "string") {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	hByte, err := hex.DecodeString(hash.(string))
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "hash")
		return
	}

	var intHeight []byte
	height, ok := rj.Get("height")
	if !ok {
		_, _, _, blockHeight, _ := mining.FindTxJsonVo(hByte)
		intHeight = evmutils.New(0).SetUint64(blockHeight).Bytes()
	} else {
		intHeight = evmutils.New(int64(height.(float64))).Bytes()
	}

	key := append([]byte(config.DBKEY_INTERNAL_TX), intHeight...)
	body, err := db.LevelDB.GetDB().HGet(key, hByte)
	if err != nil {
		//engine.Log.Error("获取保存合约事件日志失败%s", err.Error())
		res, err = model.Tojson(list)
		return
	}
	if len(body) == 0 {
		res, err = model.Tojson(list)
		return
	}

	if err = json.Unmarshal(body, list); err != nil {
		res, err = model.Tojson(list)
		return
	}

	res, err = model.Tojson(list)
	return res, err
}
