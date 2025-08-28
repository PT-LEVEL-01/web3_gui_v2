package rpc

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"strconv"
	chainconfig "web3_gui/chain/config"
	"web3_gui/config"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func RegisterRPC(node *libp2parea.Node) {
	config.Node = node
	//utils.Log.Info().Str("注册rpc接口", "").Send()
	RegisterRPC_Panic(node, 1, config.RPC_Method_chain_Info, GetInfo, "获取链的基本参数信息，地址前缀、区块高度、押金信息等")

	RegisterRPC_Panic(node, 2, config.RPC_Method_chain_BlockHeight, BlockHeight, "获取本节点的区块高度")

	RegisterRPC_Panic(node, 3, config.RPC_Method_chain_Balance, GetBalance, "获取本节点总余额")

	RegisterRPC_Panic(node, 4, config.RPC_Method_chain_AddressList, AddressList, "查询本节点的地址列表",
		engine.NewParamValid_Nomast_CustomFun_Panic("startAddr", reflect.String, VlidAddressCoin, "startAddr", "查询起始地址"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 5, config.RPC_Method_chain_AddressCreate, AddressCreate, "创建一个新收款地址",
		engine.NewParamValid_Nomast_Panic("nickname", reflect.String, "给地址设置一个昵称"),
		engine.NewParamValid_Mast_Panic("seedPassword", reflect.String, "种子密码"),
		engine.NewParamValid_Nomast_Panic("addrPassword", reflect.String, "给此地址单独设置一个密码"))

	RegisterRPC_Panic(node, 6, config.RPC_Method_chain_AddressNonce, AddressNonce, "获取一个地址的nonce",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要查询的地址"))

	RegisterRPC_Panic(node, 7, config.RPC_Method_chain_AddressNonceMore, AddressNonceMore, "获取多个地址的nonce",
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "多个地址"))

	RegisterRPC_Panic(node, 8, config.RPC_Method_chain_AddressValidate, AddressValidate, "验证一个地址格式是否正确",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要验证的地址"))

	RegisterRPC_Panic(node, 9, config.RPC_Method_chain_AddressInfo, AddressInfo, "获取一个地址的角色",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要查询的地址"))

	RegisterRPC_Panic(node, 10, config.RPC_Method_chain_AddressBalance, AddressBalance, "查询全网任一地址的余额",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要查询的地址"))

	RegisterRPC_Panic(node, 11, config.RPC_Method_chain_AddressBalanceMore, AddressBalanceMore, "查询全网多个地址的余额",
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "要查询的地址"))

	RegisterRPC_Panic(node, 12, config.RPC_Method_chain_AddressAllBalanceRange, AddressAllBalanceRange, "查询全网所有地址余额",
		engine.NewParamValid_Nomast_Panic("startAddr", reflect.String, "查询起始地址"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 46, config.RPC_Method_chain_AddressVoteInfo, AddressVoteInfo, "查询一个地址的角色，给谁投票和票数",
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "要查询的地址"))

	RegisterRPC_Panic(node, 13, config.RPC_Method_chain_TransactionRecord, TransactionRecord, "查询一个地址的历史记录",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要查询的地址"),
		engine.NewParamValid_Mast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 14, config.RPC_Method_chain_MnemonicImport, MnemonicImport, "导入助记词",
		engine.NewParamValid_Mast_Panic("words", reflect.String, "助记词"),
		engine.NewParamValid_Mast_Panic("newSeedPassword", reflect.String, "新的种子密码"))

	RegisterRPC_Panic(node, 15, config.RPC_Method_chain_MnemonicExport, MnemonicExport, "导出助记词",
		engine.NewParamValid_Mast_Panic("seedPassword", reflect.String, "种子密码"))

	RegisterRPC_Panic(node, 16, config.RPC_Method_chain_MnemonicImportEncry, MnemonicImportEncry, "导入助记词，有加密",
		engine.NewParamValid_Mast_Panic("words", reflect.String, "已加密或者未加密助记词"),
		engine.NewParamValid_Nomast_Panic("wordsPassword", reflect.String, "助记词解密密码"),
		engine.NewParamValid_Mast_Panic("newSeedPassword", reflect.String, "新的种子密码"))

	RegisterRPC_Panic(node, 17, config.RPC_Method_chain_MnemonicExportEncry, MnemonicExportEncry, "导出助记词，有加密",
		engine.NewParamValid_Mast_Panic("seedPassword", reflect.String, "种子密码"),
		engine.NewParamValid_Nomast_Panic("wordsPassword", reflect.String, "助记词加密密码"))

	RegisterRPC_Panic(node, 18, config.RPC_Method_chain_SendToAddress, SendToAddress, "转账",
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "收款地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "转账金额"),
		engine.NewParamValid_Nomast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Nomast_Panic("frozen_height", reflect.Uint64, "金额冻结高度"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"),
		engine.NewParamValid_Nomast_Panic("domain", reflect.String, "地址域名"),
		engine.NewParamValid_Nomast_Panic("domain_type", reflect.Uint64, "域名类型"))

	RegisterRPC_Panic(node, 19, config.RPC_Method_chain_SendToAddressMore, SendToAddressMore, "多人转账",
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "多个接收地址"),
		engine.NewParamValid_MastSlice_Panic("amountMore", reflect.Uint64, "转账金额"),
		engine.NewParamValid_MastSlice_Panic("frozenHeightMore", reflect.Uint64, "冻结高度"),
		engine.NewParamValid_NomastSlice_Panic("domainMore", reflect.String, "域名"),
		engine.NewParamValid_NomastSlice_Panic("domainTypeMore", reflect.Uint64, "域名类型"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 20, config.RPC_Method_chain_PayOrder, PayOrder, "支付订单",
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("serverAddr", reflect.String, VlidAddressCoin, "serverAddr1", "收款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("orderId16", reflect.String, VlidStr16, "orderId161", "订单id，16进制字符串"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "转账金额"),
		engine.NewParamValid_Nomast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"))

	RegisterRPC_Panic(node, 22, config.RPC_Method_chain_TxStatusOnChain, TxStatusOnChain, "查询一笔交易是否上链",
		engine.NewParamValid_Mast_CustomFun_Panic("txHash16", reflect.String, VlidStr16, "txHash161", "交易hash，16进制字符串"))

	RegisterRPC_Panic(node, 23, config.RPC_Method_chain_TxStatusOnChainMore, TxStatusOnChainMore, "查询多笔交易是否上链",
		engine.NewParamValid_MastSlice_CustomFun_Panic("txHashMore16", reflect.String, VlidStrSlice16, "txHashMore161", "交易hash，16进制字符串"))

	RegisterRPC_Panic(node, 24, config.RPC_Method_chain_TxProto64ByHash16, TxProto64ByHash16, "通过交易hash查询交易",
		engine.NewParamValid_Mast_CustomFun_Panic("txHash16", reflect.String, VlidStr16, "txHash161", "交易hash，16进制字符串"))

	RegisterRPC_Panic(node, 25, config.RPC_Method_chain_TxJsonByHash16, TxJsonByHash16, "通过交易hash查询交易，返回可读json",
		engine.NewParamValid_Mast_CustomFun_Panic("txHash16", reflect.String, VlidStr16, "txHash161", "交易hash，16进制字符串"))

	RegisterRPC_Panic(node, 26, config.RPC_Method_chain_BlockProto64ByHash16, BlockProto64ByHash16, "通过hash查区块内容",
		engine.NewParamValid_Mast_CustomFun_Panic("blockHash16", reflect.String, VlidStr16, "blockHash161", "区块hash，16进制字符串"))

	RegisterRPC_Panic(node, 27, config.RPC_Method_chain_BlockJsonByHash16, BlockJsonByHash16, "通过hash查区块内容，返回可读json",
		engine.NewParamValid_Mast_CustomFun_Panic("blockHash16", reflect.String, VlidStr16, "blockHash161", "区块hash，16进制字符串"))

	RegisterRPC_Panic(node, 28, config.RPC_Method_chain_BlockProto64ByHeight, BlockProto64ByHeight, "通过区块高度查区块内容",
		engine.NewParamValid_Mast_Panic("height", reflect.Uint64, "区块高度"))

	RegisterRPC_Panic(node, 29, config.RPC_Method_chain_BlockJsonByHeight, BlockJsonByHeight, "通过区块高度查区块内容，返回可读json",
		engine.NewParamValid_Mast_Panic("height", reflect.Uint64, "区块高度"))

	RegisterRPC_Panic(node, 30, config.RPC_Method_chain_BlocksProto64ByHeightRange, BlocksProto64ByHeightRange, "通过高度范围，查询区块内容",
		engine.NewParamValid_Nomast_Panic("startHeight", reflect.Uint64, "查询起始高度，包含此高度"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 31, config.RPC_Method_chain_BlocksJsonByHeightRange, BlocksJsonByHeightRange, "通过高度范围，查询区块内容，返回可读json",
		engine.NewParamValid_Nomast_Panic("startHeight", reflect.Uint64, "查询起始高度，包含此高度"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 32, config.RPC_Method_chain_ChainDepositAll, ChainDepositAll, "全网所有见证人、社区、轻节点的质押金额总量")
	RegisterRPC_Panic(node, 33, config.RPC_Method_chain_ChainDepositNodeNum, ChainDepositNodeNum, "全网各见证人、社区、轻节点的质押数量")

	RegisterRPC_Panic(node, 34, config.RPC_Method_chain_WitnessDepositIn, WitnessDepositIn, "成为见证人，缴纳押金。指定钱包第一个地址作为见证人",
		engine.NewParamValid_Nomast_Panic("nickname", reflect.String, "昵称"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("rate", reflect.Uint64, "分账比例：1-100(%)"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"))

	RegisterRPC_Panic(node, 35, config.RPC_Method_chain_WitnessDepositOut, WitnessDepositOut, "取消见证人，退回见证人押金",
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"))

	RegisterRPC_Panic(node, 36, config.RPC_Method_chain_WitnessFeatureList, WitnessFeatureList, "正在出块的见证人列表")

	RegisterRPC_Panic(node, 37, config.RPC_Method_chain_WitnessFeatureInfo, WitnessFeatureInfo, "正在出块的见证人信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "见证人地址"))

	RegisterRPC_Panic(node, 38, config.RPC_Method_chain_WitnessCandidateList, WitnessCandidateList, "候选见证人列表")

	RegisterRPC_Panic(node, 39, config.RPC_Method_chain_WitnessCandidateInfo, WitnessCandidateInfo, "候选见证人信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "见证人地址"))

	RegisterRPC_Panic(node, 40, config.RPC_Method_chain_WitnessZoneDeposit, WitnessZoneDeposit, "见证人大区质押总量",
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "要查询的多个地址"))

	RegisterRPC_Panic(node, 41, config.RPC_Method_chain_WitnessList, WitnessList, "全网见证人列表",
		engine.NewParamValid_Nomast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 42, config.RPC_Method_chain_WitnessInfo, WitnessInfo, "全网查询见证人信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "见证人地址"),
		engine.NewParamValid_Nomast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 43, config.RPC_Method_chain_CommunityList, CommunityList, "全网社区节点列表",
		engine.NewParamValid_Nomast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 44, config.RPC_Method_chain_CommunityInfo, CommunityInfo, "社区节点信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "见证人地址"),
		engine.NewParamValid_Nomast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 45, config.RPC_Method_chain_CommunityDepositIn, CommunityDepositIn, "社区节点质押",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "成为社区节点地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("witness", reflect.String, VlidAddressCoin, "witness1", "绑定的见证人地址"),
		engine.NewParamValid_Mast_Panic("rate", reflect.Uint64, "分给轻节点的比例"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 46, config.RPC_Method_chain_CommunityDepositOut, CommunityDepositOut, "社区节点取消质押",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "待取消的社区节点地址"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 47, config.RPC_Method_chain_CommunityShowRewardPool, CommunityShowRewardPool, "获取社区节点累计未分配奖励",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "社区节点地址"))

	RegisterRPC_Panic(node, 48, config.RPC_Method_chain_CommunityDistributeReward, CommunityDistributeReward, "社区节点手动分配奖励",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "社区节点地址"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 49, config.RPC_Method_chain_LightList, LightList, "全网查询轻节点列表",
		engine.NewParamValid_Nomast_Panic("page", reflect.Uint64, "第几页"),
		engine.NewParamValid_Nomast_Panic("total", reflect.Uint64, "分页查询的情况下，一页查询多少条记录"))

	RegisterRPC_Panic(node, 50, config.RPC_Method_chain_LightInfo, LightInfo, "轻节点信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "见证人地址"))

	RegisterRPC_Panic(node, 51, config.RPC_Method_chain_LightDepositIn, LightDepositIn, "轻节点质押",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "成为轻节点地址"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 52, config.RPC_Method_chain_LightDepositOut, LightDepositOut, "轻节点取消质押",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "待取消的轻节点地址"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 53, config.RPC_Method_chain_LightVoteIn, LightVoteIn, "轻节点投票",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "轻节点地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("community", reflect.String, VlidAddressCoin, "community1", "社区地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "投票金额"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 54, config.RPC_Method_chain_LightVoteOut, LightVoteOut, "轻节点取消投票",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "轻节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "投票金额"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 55, config.RPC_Method_chain_LightDistributeReward, LightDistributeReward, "轻节点手动分配奖励",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "轻节点地址"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 56, config.RPC_Method_chain_TokenPublish, TokenPublish, "发布一个token",
		engine.NewParamValid_Mast_Panic("name", reflect.String, "Token名称全称"),
		engine.NewParamValid_Mast_Panic("symbol", reflect.String, "Token单位，符号"),
		engine.NewParamValid_Mast_CustomFun_Panic("supply", reflect.String, SupplyStr10, "supply1",
			"Token发行总量，最少："+chainconfig.Witness_token_supply_min.String()),
		engine.NewParamValid_Mast_Panic("accuracy", reflect.Uint64, "精度"),
		engine.NewParamValid_Nomast_CustomFun_Panic("owner", reflect.String, VlidAddressCoin, "owner1", "所有者"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 57, config.RPC_Method_chain_TokenBalance, TokenBalance, "获取一个地址的角色",
		engine.NewParamValid_Mast_CustomFun_Panic("tokenId16", reflect.String, VlidStr16, "tokenId161", "token id，16进制字符串"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "要查询的地址"))

	RegisterRPC_Panic(node, 58, config.RPC_Method_chain_TokenSendToAddress, TokenSendToAddress, "使用token转账",
		engine.NewParamValid_Mast_CustomFun_Panic("tokenId16", reflect.String, VlidStr16, "tokenId161", "tokenId，16进制字符串"),
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "收款地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "转账金额"),
		engine.NewParamValid_Nomast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Nomast_Panic("frozen_height", reflect.Uint64, "金额冻结高度"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 59, config.RPC_Method_chain_TokenSendToAddressMore, TokenSendToAddressMore, "使用token多人转账",
		engine.NewParamValid_Mast_CustomFun_Panic("tokenId16", reflect.String, VlidStr16, "tokenId161", "tokenId，16进制字符串"),
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_MastSlice_CustomFun_Panic("addressMore", reflect.String, VlidAddressCoinMore, "addressMore1", "多个接收地址"),
		engine.NewParamValid_MastSlice_Panic("amountMore", reflect.Uint64, "转账金额"),
		engine.NewParamValid_MastSlice_Panic("frozenHeightMore", reflect.Uint64, "冻结高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 60, config.RPC_Method_chain_TokenInfo, TokenInfo, "token详细信息",
		engine.NewParamValid_Mast_CustomFun_Panic("tokenId16", reflect.String, VlidStr16, "tokenId161", "tokenId，16进制字符串"))

	RegisterRPC_Panic(node, 61, config.RPC_Method_chain_TokenList, TokenList, "自己相关的token列表")

	RegisterRPC_Panic(node, 62, config.RPC_Method_chain_ContractCreate, ContractCreate, "创建合约",
		engine.NewParamValid_Mast_Panic("dataStr", reflect.String, "合约编译后的内容字符串"),
		engine.NewParamValid_Nomast_Panic("abi", reflect.String, "合约abi"),
		engine.NewParamValid_Nomast_Panic("source", reflect.String, "合约源代码"),
		engine.NewParamValid_Nomast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Nomast_Panic("class", reflect.Uint64, "合约类型"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Nomast_Panic("gasPrice", reflect.Uint64, "gasPrice"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 63, config.RPC_Method_chain_ContractPushTxProto64, ContractPushTxProto64,
		"将组装并签名好的合约交易上链，protobuf字节码base64字符串",
		engine.NewParamValid_Mast_Panic("base64StdStr", reflect.String, "离线交易base64Std字符串"),
		engine.NewParamValid_Nomast_Panic("checkBalance", reflect.Bool, "是否验证转账余额足够"))

	RegisterRPC_Panic(node, 64, config.RPC_Method_chain_ContractInfo, ContractInfo, "合约状态信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "合约地址"))

	RegisterRPC_Panic(node, 65, config.RPC_Method_chain_ContractCall, ContractCall, "调用合约中的方法",
		engine.NewParamValid_Mast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("contractAddress", reflect.String, VlidAddressCoin, "contractAddress1", "合约地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("dataStr", reflect.String, VlidStr16, "dataStr1", "调用合约参数，16进制字符串"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Nomast_Panic("gasPrice", reflect.Uint64, "gasPrice"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 66, config.RPC_Method_chain_ContractCallStack, ContractCallStack, "调用合约中的方法，本地模拟调用",
		engine.NewParamValid_Mast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("contractAddress", reflect.String, VlidAddressCoin, "contractAddress1", "合约地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("dataStr", reflect.String, VlidStr16, "dataStr1", "调用合约参数，16进制字符串"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Nomast_Panic("gasPrice", reflect.Uint64, "gasPrice"))

	RegisterRPC_Panic(node, 67, config.RPC_Method_chain_ContractEvent, ContractEvent, "根据区块高度和交易hash获取奖励合约事件",
		engine.NewParamValid_Nomast_Panic("customTxEvent", reflect.Bool, "是否查询自定义事件"),
		engine.NewParamValid_Nomast_CustomFun_Panic("txHash16", reflect.String, VlidStr16, "txHash161", "交易hash，16进制字符串"),
		engine.NewParamValid_Nomast_Panic("height", reflect.Uint64, "区块高度"))

	//RegisterRPC_Panic(node, 58, config.RPC_Method_chain_ERC20Create, ERC20Create, "创建一个ERC20合约",
	//	engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "合约地址"))

	RegisterRPC_Panic(node, 68, config.RPC_Method_chain_ERC20Info, ERC20Info, "获取ERC20合约信息",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "合约地址"))

	RegisterRPC_Panic(node, 69, config.RPC_Method_chain_GetTxGas, GetTxGas, "根据交易类型获取交易gas 默认普通交易")

	RegisterRPC_Panic(node, 201, config.RPC_Method_chain_OfflineTxSendToAddress, OfflineTxSendToAddress, "创建离线交易，转账",
		engine.NewParamValid_Mast_Panic("key_store_path", reflect.String, "钱包路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "收款地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "转账金额"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "链端当前高度"),
		engine.NewParamValid_Nomast_Panic("frozen_height", reflect.Uint64, "金额冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 202, config.RPC_Method_chain_OfflineTxCreateContract, OfflineTxCreateContract, "创建离线交易，创建合约",
		engine.NewParamValid_Mast_Panic("key_store_path", reflect.String, "钱包路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("srcaddress", reflect.String, VlidAddressCoin, "srcaddress1", "扣款地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "收款地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "转账金额"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "链端当前高度"),
		engine.NewParamValid_Nomast_Panic("frozen_height", reflect.Uint64, "金额冻结高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("gasPrice", reflect.Uint64, "gasPrice"),
		engine.NewParamValid_Mast_Panic("abi", reflect.String, "合约abi"),
		engine.NewParamValid_Nomast_Panic("source", reflect.String, "合约源代码"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 203, config.RPC_Method_chain_OfflineTxCommunityDepositIn, OfflineTxCommunityDepositIn, "社区节点质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("witness", reflect.String, VlidAddressCoin, "witness1", "绑定的见证人地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "成为社区节点地址"),
		engine.NewParamValid_Mast_Panic("rate", reflect.Uint64, "分给轻节点的比例"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "押金"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 204, config.RPC_Method_chain_OfflineTxCommunityDepositOut, OfflineTxCommunityDepositOut, "社区节点取消质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "待取消的社区节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "押金"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 205, config.RPC_Method_chain_OfflineTxLightDepositIn, OfflineTxLightDepositIn, "社区节点质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "成为轻节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "押金"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 206, config.RPC_Method_chain_OfflineTxLightDepositOut, OfflineTxLightDepositOut, "社区节点取消质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "待取消的轻节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "押金"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 207, config.RPC_Method_chain_OfflineTxLightVoteIn, OfflineTxLightVoteIn, "社区节点质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("community", reflect.String, VlidAddressCoin, "community1", "社区节点地址"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "轻节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "投票金额"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 208, config.RPC_Method_chain_OfflineTxLightVoteOut, OfflineTxLightVoteOut, "社区节点取消质押",
		engine.NewParamValid_Mast_Panic("walletPath", reflect.String, "钱包文件路径"),
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "轻节点地址"),
		engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "取消投票金额"),
		engine.NewParamValid_Mast_Panic("nonce", reflect.Uint64, "nonce"),
		engine.NewParamValid_Mast_Panic("currentHeight", reflect.Uint64, "当前节点区块高度，此参数值决定了交易上链最大高度"),
		engine.NewParamValid_Mast_Panic("gas", reflect.Uint64, "手续费"),
		engine.NewParamValid_Mast_Panic("pwd", reflect.String, "支付密码"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

	RegisterRPC_Panic(node, 299, config.RPC_Method_chain_PushTxProto64, PushTxProto64, "提交已经签名的离线交易上链",
		engine.NewParamValid_Mast_Panic("base64StdStr", reflect.String, "离线交易base64Std字符串"),
		engine.NewParamValid_Nomast_Panic("checkBalance", reflect.Bool, "是否验证转账余额足够"))

	RegisterRPC_Panic(node, 1001, config.RPC_Method_chain_TestReceiveCoin, TestReceiveCoin, "根据区块高度和交易hash获取奖励合约事件",
		engine.NewParamValid_Mast_CustomFun_Panic("address", reflect.String, VlidAddressCoin, "address1", "收款地址"),
		engine.NewParamValid_Nomast_Panic("comment", reflect.String, "备注信息"))

}

func RegisterRPC_Panic(node *libp2parea.Node, sortNumber int, rpcName string, handler any, desc string, pvs ...engine.ParamValid) {
	ERR := node.RegisterRPC(sortNumber, rpcName, handler, desc, pvs...)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}

/*
验证地址的合法性
*/
func VlidAddressCoin(params interface{}) (any, utils.ERROR) {
	//utils.Log.Info().Str("地址", "11111111").Send()
	addrStr := params.(string)
	if addrStr == "" {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	addrCoin := keystore.AddressFromB58String(addrStr)
	//utils.Log.Info().Str("地址", addrCoin.GetPre()).Send()
	if !keystore.ValidAddrCoin(addrCoin.GetPre(), addrCoin) {
		utils.Log.Info().Str("地址错误", addrStr).Send()
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	return addrCoin, utils.NewErrorSuccess()
}

/*
验证地址的合法性
*/
func VlidAddressCoinMore(params interface{}) (any, utils.ERROR) {
	paramsItr := params.([]interface{})
	addrs := make([]string, 0, len(paramsItr))
	for _, one := range paramsItr {
		addrOne := one.(string)
		addrs = append(addrs, addrOne)
	}
	addrCoins := make([]coin_address.AddressCoin, 0, len(addrs))
	for i, one := range addrs {
		if one == "" {
			return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "index:"+strconv.Itoa(i))
		}
		addrCoin := keystore.AddressFromB58String(one)
		if !keystore.ValidAddrCoin(addrCoin.GetPre(), addrCoin) {
			return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "index:"+strconv.Itoa(i))
		}
		addrCoins = append(addrCoins, addrCoin)
	}
	return addrCoins, utils.NewErrorSuccess()
}

/*
验证地址的合法性
*/
func VlidAddressNet(params interface{}) (any, utils.ERROR) {
	addrStr := params.(string)
	if addrStr == "" {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	addrNet := nodeStore.AddressFromB58String(addrStr)
	if !keystore.ValidAddrNet(addrNet.GetPre(), addrNet.GetAddr()) {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	return addrNet, utils.NewErrorSuccess()
}

/*
验证16进制字符串
*/
func VlidStr16(params interface{}) (any, utils.ERROR) {
	addrStr := params.(string)
	if addrStr == "" {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	idBs, err := hex.DecodeString(addrStr)
	if err != nil {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, err.Error())
	}
	return idBs, utils.NewErrorSuccess()
}

/*
验证16进制字符串
*/
func VlidStrSlice16(params interface{}) (any, utils.ERROR) {
	addrStrMore := params.([]string)
	if len(addrStrMore) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	bsMore := make([][]byte, 0, len(addrStrMore))
	for i, one := range addrStrMore {
		if one == "" {
			return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "index:"+strconv.Itoa(i))
		}
		idBs, err := hex.DecodeString(one)
		if err != nil {
			return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "index:"+strconv.Itoa(i))
		}
		bsMore = append(bsMore, idBs)
	}
	return bsMore, utils.NewErrorSuccess()
}

/*
验证10进制字符串
*/
func SupplyStr10(params interface{}) (any, utils.ERROR) {
	addrStr := params.(string)
	if addrStr == "" {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	supply, ok := new(big.Int).SetString(addrStr, 10)
	if !ok || supply.Cmp(chainconfig.Witness_token_supply_min) < 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	return supply, utils.NewErrorSuccess()
}
