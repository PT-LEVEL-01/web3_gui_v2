/*
数据库中key保存的格式
1.保存第一个区块的hash:startblock
2.保存区块头:[区块hash]
3.交易历史纪录:1_[1/2]_[自己的地址]_[目标地址]
*/
package config

import (
	"web3_gui/utils"
)

const (
// block_start_str = "startblock" //保存第一个区块hash
// History                        = "1_"             //key以“1_”前缀的含义：交易历史数据
// BlockHeight                    = "2_"             //key以“2_”前缀的含义：块高度和区块hash对应
// Name                           = "name_"          //key以“name_”前缀的含义：已经注册的域名
// Block_Highest                  = "HighestBlock"   //保存最高区块数量
// LEVELDB_Head_history_balance   = "h_b_"           //交易历史记录 格式：h_b_[分叉链编号]_[GenerateId]例：h_b_0_100
// WitnessName                    = "3_"             //key以“3_”前缀的含义：见证人名称和社区节点名称对应的地址
// WitnessAddr                    = "4_"             //key以“4_”前缀的含义：见证人地址和社区节点地址对应的名称
// AlreadyUsed_tx                 = "5_"             //key以“5_”前缀的含义：未花费的交易余额索引 格式：5_[交易hash]_[输出index]
// TokenInfo                      = "6_"             //key以“6_”前缀的含义：发布token的信息(单位，总量) 格式：6_[交易hash]
// TokenPublishTxid               = "7_"             //key以“7_”前缀的含义：发布token的合约txid 格式：7_[交易hash]_[输出index]
// DB_PRE_Tx_Not_Import           = "8_"             //key以“8_”前缀的含义：未导入区块中的交易txid，导入后删除这个key 格式：8_[交易hash]
// DB_spaces_mining_addr          = "9_"             //key以“9_”前缀的含义：质押地址对应的质押金额 格式：9_[质押地址]
// DB_community_addr_hash         = "10_"            //key以“9_”前缀的含义：成为社区节点的交易区块hash 格式：9_[社区节点地址]
// DBKEY_tx_blockhash             = "11_"            //key以“11_”前缀的含义：保存交易所属的区块hash 格式：11_[交易hash]
// DBKEY_addr_value               = "12_"            //key以“12_”前缀的含义：保存地址余额 格式：12_[地址]
// DBKEY_community_reward         = "13_"            //key以“13_”前缀的含义：保存社区节点地址对应未分配奖励余额 格式：13_[地址]
// DBKEY_community_addr           = "14_"            //key以“14_”前缀的含义：保存社区节点地址，用来查询当前状态某地址是否社区节点 格式：14_[地址]
// DBKEY_vote_addr                = "15_"            //key以“15_”前缀的含义：保存节点的投票地址 格式：15_[地址]
// DBKEY_deposit_witness_addr     = "16_"            //key以“16_”前缀的含义：保存见证人地址的押金 格式：16_[地址]
// DBKEY_deposit_light_addr       = "17_"            //key以“16_”前缀的含义：保存轻节点押金 格式：17_[地址]
// DBKEY_deposit_community_addr   = "18_"            //key以“16_”前缀的含义：保存社区节点的押金 格式：18_[地址]
// DBKEY_deposit_light_vote_value = "19_"            //key以“16_”前缀的含义：保存轻节点投票数量 格式：19_[地址]
// DBKEY_addr_nonce               = "20_"            //key以“20_”前缀的含义：保存地址的自增长nonce 格式：20_[地址]
// DBKEY_token_addr_value         = "21_"            //key以“21_”前缀的含义：保存地址Token余额 格式：21_[地址]
// EFFECTIVE_contract_count       = "contract_count" //有效的合约数量
// DBKEY_CONTRACT_OBJECT          = "stateObject_"   //合约对象 key 为addressCoin 值为stateObject
// DBKEY_contract_source          = "22_"            //key以“22_”前缀的含义：保存合约sol源代码 格式：22_[合约地址]
// DBKEY_BLOCK_EVENT              = "23_"            //key以“23_”前缀的含义：保存区块事件日志 格式：23_[区块高度]
// DBKEY_TxCONTRACT_GASUSED       = "24_"            //key以“24_”前缀的含义：保存合约交易的gasused 格式:24_[交易hash]
// DBKEY_SPECIAL_CONTRACT         = "25_"            //key以“25_”前缀的含义:保存用户特定类型合约地址，格式:25_[账户地址]
// Address_history_tx             = "address_tx_"    //地址交易历史记录 格式：address_tx_[地址]例：address_tx_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
// //Address_tx_count               = "address_tx_count_" //地址交易数量 格式：address_tx_count_[地址]例：address_tx_count_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
// Miner_history_tx        = "miner_block_"    //出块详情 格式：miner_block_[地址]例：miner_block_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
// DBKEY_Light_REWARD      = "26_"             //轻节点获得的总奖励 格式：26_[轻节点地址 社区节点地址]例：26_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
// DBKEY_REWARD_HISTORY    = "27_"             //节点获取奖励的记录 格式：27_[收益地址 来源地址] value为奖励收益
// DBKEY_ERC20_COIN        = "collect_coin"    //收藏的代币
// DBKEY_Message_Multicast = "28_"             //key以“28_”前缀的含义:保存用户特定类型合约地址，格式:28_[HASH]
// FAUCET_ADDRESS          = "29_"             //保存地址领取faucet的请求时间
// DBKEY_ERC20_CONTRACT    = "contract_erc20"  //erc20合约
// DBKEY_contract_abi      = "30_"             //key以“30_”前缀的含义：保存合约sol abi 格式：30_[合约abi]
// DBKEY_contract_bin      = "31_"             //key以“30_”前缀的含义：保存合约sol bin 格式：31_[合约bin]
// DBKEY_snapshot          = "snapshot_"       //key以“snapshot_”前缀的含义：快照碎片对象
// DBKEY_snapshot_height   = "snapshot_height" //key:snapshot_height” value：最后一次成功快照高度
)

var (

	//leveldb
	Key_block_start = []byte{0, 0} //key以“0_”前缀的含义：保存第一个区块hash
	DBKEY_BlockHead = []byte{0, 1} //key以“1_”前缀的含义：保存区块头,NOTE：直接读写db,不走增量数据
	DBKEY_Block_tx  = []byte{0, 2} //key以“2_”前缀的含义：保存交易,NOTE：直接读写db,不走增量数据

	Block_Highest            = []byte{0, 3}  //key以“3_”前缀的含义：保存最高区块数量
	BlockHeight              = []byte{0, 4}  //key以“4_”前缀的含义：块高度和区块hash对应
	DB_PRE_Tx_Not_Import     = []byte{0, 5}  //key以“5_”前缀的含义：未导入区块中的交易txid，导入后删除这个key 格式：5_[交易hash]
	EFFECTIVE_contract_count = []byte{0, 6}  //key以“6_”前缀的含义：有效的合约数量
	DBKEY_contract_source    = []byte{0, 7}  //key以“7_”前缀的含义：保存合约sol源代码 格式：7_[合约地址]
	DBKEY_BLOCK_EVENT        = []byte{0, 8}  //key以“8_”前缀的含义：保存区块事件日志 格式：8_[区块高度]
	DBKEY_TxCONTRACT_GASUSED = []byte{0, 9}  //key以“9_”前缀的含义：保存合约交易的gasused 格式:9_[交易hash]
	DBKEY_SPECIAL_CONTRACT   = []byte{0, 10} //key以“10_”前缀的含义:保存用户特定类型合约地址，格式:10_[账户地址]
	DBKEY_REWARD_HISTORY     = []byte{0, 11} //key以“11_”前缀的含义:节点获取奖励的记录 格式：11_[收益地址 来源地址] value为奖励收益
	DBKEY_ERC20_COIN         = []byte{0, 12} //key以“12_”前缀的含义:收藏的代币
	FAUCET_ADDRESS           = []byte{0, 13} //key以“13_”前缀的含义:保存地址领取faucet的请求时间
	DBKEY_ERC20_CONTRACT     = []byte{0, 14} //key以“14_”前缀的含义:erc20合约
	DBKEY_contract_abi       = []byte{0, 15} //key以“15_”前缀的含义:保存合约sol abi 格式：15_[合约abi]
	DBKEY_contract_bin       = []byte{0, 16} //key以“16_”前缀的含义:保存合约sol bin 格式：16_[合约bin]
	DBKEY_blockhead_bloom    = []byte{0, 17} //key以“17_”前缀的含义:保存区块的bloom值
	DBKEY_tx_bloom           = []byte{0, 18} //key以“18_”前缀的含义:保存交易的bloom值
	DBKEY_ERC721_CONTRACT    = []byte{0, 19} //key以“19_”前缀的含义:erc721合约
	DBKEY_ERC1155_CONTRACT   = []byte{0, 20} //key以“20_”前缀的含义:erc1155合约
	DBKEY_INTERNAL_TX        = []byte{0, 21} //key以“21_”前缀的含义:internal tx

	//tempdb
	History                      = []byte{1, 1} //key以“1_”前缀的含义：交易历史数据
	Name                         = []byte{1, 2} //key以“2_”前缀的含义：已经注册的域名
	LEVELDB_Head_history_balance = []byte{1, 3} //key以“3_”前缀的含义：交易历史记录 格式：[分叉链编号]_[GenerateId]例：3_0_100
	//WitnessName                    = []byte{1, 4}  //key以“4_”前缀的含义：见证人名称和社区节点名称对应的地址
	WitnessAddr                    = []byte{1, 5}  //key以“5_”前缀的含义：见证人地址和社区节点地址对应的名称
	AlreadyUsed_tx                 = []byte{1, 6}  //key以“6_”前缀的含义：未花费的交易余额索引 格式：6_[交易hash]_[输出index]
	TokenInfo                      = []byte{1, 7}  //key以“7_”前缀的含义：发布token的信息(单位，总量) 格式：7_[交易hash]
	TokenPublishTxid               = []byte{1, 8}  //key以“8_”前缀的含义：发布token的合约txid 格式：8_[交易hash]_[输出index]
	DB_spaces_mining_addr          = []byte{1, 9}  //key以“9_”前缀的含义：质押地址对应的质押金额 格式：9_[质押地址]
	DB_community_addr_hash         = []byte{1, 10} //key以“10_”前缀的含义：成为社区节点的交易区块hash 格式：10_[社区节点地址]
	DBKEY_tx_blockhash             = []byte{1, 11} //key以“11_”前缀的含义：保存交易所属的区块hash 格式：11_[交易hash]
	DBKEY_addr_value               = []byte{1, 12} //key以“12_”前缀的含义：保存地址余额 格式：12_[地址]
	DBKEY_community_reward         = []byte{1, 13} //key以“13_”前缀的含义：保存社区节点地址对应未分配奖励余额 格式：13_[地址]
	DBKEY_community_addr           = []byte{1, 14} //key以“14_”前缀的含义：保存社区节点地址，用来查询当前状态某地址是否社区节点 格式：14_[地址] 格式：14_[地址]
	DBKEY_vote_addr                = []byte{1, 15} //key以“15_”前缀的含义：保存节点的投票地址 格式：15_[地址]
	DBKEY_deposit_witness_addr     = []byte{1, 16} //key以“16_”前缀的含义：保存见证人地址的押金 格式：16_[地址]
	DBKEY_deposit_light_addr       = []byte{1, 17} //key以“17_”前缀的含义：保存轻节点押金 格式：17_[地址]
	DBKEY_deposit_community_addr   = []byte{1, 18} //key以“18_”前缀的含义：保存社区节点的押金 格式：18_[地址]
	DBKEY_deposit_light_vote_value = []byte{1, 19} //key以“19_”前缀的含义：保存轻节点投票数量 格式：19_[地址]
	DBKEY_addr_nonce               = []byte{1, 20} //key以“20_”前缀的含义：保存地址的自增长nonce 格式：20_[地址]
	DBKEY_token_addr_value         = []byte{1, 21} //key以“21_”前缀的含义：保存地址Token余额 格式：21_[地址]
	DBKEY_CONTRACT_OBJECT          = []byte{1, 22} //key以“22_”前缀的含义：合约对象 key 为addressCoin 值为stateObject
	Address_history_tx             = []byte{1, 23} //key以“23_”前缀的含义：地址交易历史记录 格式：23_[地址]例：23_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
	Miner_history_tx               = []byte{1, 24} //key以“24_”前缀的含义：出块详情 格式：24_[地址]例：24_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
	DBKEY_Light_REWARD             = []byte{1, 25} //key以“25_”前缀的含义：轻节点获得的总奖励 格式：25_[轻节点地址 社区节点地址]例：25_MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4MMSLQbionTB915GTsaU8hiXMBCUD2uaYTBXV4
	DBKEY_Message_Multicast        = []byte{1, 26} //key以“26_”前缀的含义：保存用户特定类型合约地址，格式:26_[HASH]
	DBKEY_snapshot                 = []byte{1, 27} //key以“27_”前缀的含义：快照碎片对象
	DBKEY_snapshot_height          = []byte{1, 28} //key以“28_”前缀的含义：key:28_height” value：最后一次成功快照高度,同时也是一个存在快照的标记

	DBKEY_zset_frozen_height          = []byte{1, 29} //key以“29_”前缀的含义：保存以区块高度排序的 格式：29_[地址]
	DBKEY_zset_frozen_time            = []byte{1, 30} //key以“30_”前缀的含义：保存地址Token余额 格式：30_[地址]
	DBKEY_zset_frozen_height_children = []byte{1, 31} //key以“31_”前缀的含义：保存以区块高度排序的 格式：31_[地址]
	DBKEY_zset_frozen_time_children   = []byte{1, 32} //key以“32_”前缀的含义：保存地址Token余额 格式：32_[地址]
	DBKEY_addr_frozen_value           = []byte{1, 33} //key以“33_”前缀的含义：保存地址Token余额 格式：33_[地址]
	DBKEY_addr_value_big              = []byte{1, 34} //key以“34_”前缀的含义：大余额地址，先冻结处理
	DBKEY_hash_key_block_height       = []byte{1, 35} //key以“35_”前缀的含义：保存冻结高度
	DBKEY_hash_key_block_time         = []byte{1, 36} //key以“36_”前缀的含义：保存冻结时间
	DBKEY_hash_key_nft_owner          = []byte{1, 37} //key以“37_”前缀的含义：保存用户的NFT
	DBKEY_hash_key_addr_nft           = []byte{1, 38} //key以“38_”前缀的含义：保存每个地址拥有的NFT
	DBKEY_light_cache                 = []byte{1, 39} //key以“39_”前缀的含义：保存轻节点缓存
	//DBKEY_light_last_vote_height      = []byte{1, 40} //key以“40_”前缀的含义：获取轻节点最后投票高度
	DBKEY_light_all_reward        = []byte{1, 41} //key以“41_”前缀的含义：保存轻节点地址对应累计奖励 格式：41_[地址]
	DBKEY_community_all_reward    = []byte{1, 42} //key以“42_”前缀的含义：保存社区节点地址对应累计奖励 格式：42_[地址]
	DBKEY_community_reward_pool   = []byte{1, 43} //key以“43_”前缀的含义：保存社区节点地址奖励池 格式：43_[地址]
	Address_history_custom_tx     = []byte{1, 44} //key以“44_”前缀的含义：细化地址交易历史记录
	DBKEY_token_addr_frozen_value = []byte{1, 45} //key以“45_”前缀的含义：保存地址Token冻结金额
	DBKEY_token_addr_locked_value = []byte{1, 46} //key以“46_”前缀的含义：保存地址Token锁定金额
	DBKEY_token_order             = []byte{1, 47} //key以“47_”前缀的含义：保存订单信息
	DBKEY_swap_tx                 = []byte{1, 48} //key以“48_”前缀的含义：swap交易记录
	DBKEY_swap_tx_pool            = []byte{1, 49} //key以“49_”前缀的含义：swap交易池记录
	DBKEY_address_tx_bind         = []byte{1, 50} //key以“50_”前缀的含义：地址绑定
	DBKEY_address_tx_frozen       = []byte{1, 51} //key以“50_”前缀的含义：地址绑定冻结
	DBKEY_multsign_set            = []byte{1, 52} //key以“52_”前缀的含义：保存创建的多签集合
	DBKEY_multsign_request_tx     = []byte{1, 53} //key以“53_”前缀的含义：多签请求交易
)

type ContractClass uint64

const (
	NORMAL_CONTRACT              ContractClass = iota //正常合约
	FAUCET_CONTRACT                                   //水龙头合约
	CLOUD_STORAGE_PROXY_CONTRACT                      //云存储代理合约
	END_CONTRACT                                      //用于合法性检查
)

func (c ContractClass) IsLegal() bool {
	return c < END_CONTRACT
}

///*
//	构建交易历史转入key
//*/
//func BuildHistoryInKey(self, tag string) []byte {
//	return []byte(History + In + self + "_" + tag)
//}

///*
//	构建交易历史转出key
//*/
//func BuildHistoryOutKey(self, tag string) []byte {
//	return []byte(History + Out + self + "_" + tag)
//}

/*
构建未导入区块的交易key标记
将未导入的区块中的交易使用此key保存到数据库中作为标记，如果已经导入过区块，则删除此标记
用作验证已存在的交易hash
*/
func BuildTxNotImport(txid []byte) []byte {
	return append([]byte(DB_PRE_Tx_Not_Import), txid...)
}

/*
成为社区节点的起始区块高度
*/
func BuildCommunityAddrStartHeight(addr []byte) []byte {
	return append([]byte(DB_community_addr_hash), addr...)
}

/*
保存交易所属的区块hash
*/
func BuildTxToBlockHash(txid []byte) []byte {
	return append([]byte(DBKEY_tx_blockhash), txid...)
}

/*
构建收款地址key
*/
func BuildAddrValue(addr []byte) []byte {
	return append([]byte(DBKEY_addr_value), addr...)
}

/*
构建社区节点收款地址key
*/
func BuildDBKeyCommunityAddrFrozen(addr []byte) *[]byte {
	key := append([]byte(DBKEY_community_reward), addr...)
	return &key
}

/*
构建轻节点奖励key
*/
func BuildDBKeyLightVoteReward(lightAddr, communityAddr []byte) *[]byte {
	keyBs := []byte(DBKEY_Light_REWARD)
	key := make([]byte, 0, len(keyBs)+len(lightAddr)+len(communityAddr))
	key = append(keyBs, lightAddr...)
	key = append(key, communityAddr...)
	return &key
}

/*
构建社区节点地址，记录当前状态某地址是否是社区节点
*/
func BuildDBKeyCommunityAddr(addr []byte) *[]byte {
	key := append([]byte(DBKEY_community_addr), addr...)
	return &key
}

/*
构建投票地址
*/
func BuildDBKeyVoteAddr(addr []byte) *[]byte {
	key := append([]byte(DBKEY_vote_addr), addr...)
	return &key
}

/*
构建投票地址
*/
func BuildDBKeyDepositWitnessAddr(addr []byte) *[]byte {
	key := append([]byte(DBKEY_deposit_witness_addr), addr...)
	return &key
}

/*
构建投票地址
*/
func BuildDBKeyDepositLightAddr(addr []byte) *[]byte {
	key := append([]byte(DBKEY_deposit_light_addr), addr...)
	return &key
}

/*
构建投票地址
*/
func BuildDBKeyDepositCommunityAddr(addr []byte) *[]byte {
	key := append([]byte(DBKEY_deposit_community_addr), addr...)
	return &key
}

/*
轻节点投票总金额
*/
func BuildDBKeyDepositLightVoteValue(lightAddr, communityAddr []byte) *[]byte {
	keyBs := []byte(DBKEY_deposit_light_vote_value)
	key := make([]byte, 0, len(keyBs)+len(lightAddr)+len(communityAddr))
	key = append(keyBs, lightAddr...)
	key = append(key, communityAddr...)
	return &key

	// key := append([]byte(DBKEY_deposit_light_vote_value), addr...)
	// return &key
}

/*
构建地址的nonce
*/
func BuildDBKeyAddrNonce(addr []byte) *[]byte {
	key := append([]byte(DBKEY_addr_nonce), addr...)
	return &key
}

/*
构建地址Token冻结余额
*/
func BuildDBKeyTokenAddrValue(token, addr []byte) *[]byte {
	keyBs := []byte(DBKEY_token_addr_value)
	key := make([]byte, 0, len(keyBs)+len(token)+len(addr))
	key = append(keyBs, token...)
	key = append(key, addr...)
	return &key
}

/*
构建地址Token冻结余额
*/
func BuildDBKeyTokenAddrFrozenValue(token, addr []byte) *[]byte {
	keyBs := []byte(DBKEY_token_addr_frozen_value)
	key := make([]byte, 0, len(keyBs)+len(token)+len(addr))
	key = append(keyBs, token...)
	key = append(key, addr...)
	return &key
}

/*
构建地址Token锁定余额
*/
func BuildDBKeyTokenAddrLockedValue(token, addr []byte) *[]byte {
	keyBs := []byte(DBKEY_token_addr_frozen_value)
	key := make([]byte, 0, len(keyBs)+len(token)+len(addr))
	key = append(keyBs, token...)
	key = append(key, addr...)
	return &key
}

/*
构建Token订单
订单交易的Hash作为订单Id
*/
func BuildDBKeyTokenOrder(orderId []byte) *[]byte {
	keyBs := []byte(DBKEY_token_order)
	key := make([]byte, 0, len(keyBs)+len(orderId))
	key = append(key, keyBs...)
	key = append(key, orderId...)
	return &key
}

/*
构建收款地址key
*/
func BuildAddrFrozen(addr []byte) []byte {
	return append(DBKEY_addr_frozen_value, addr...)
}

/*
	构建轻节点给社区节点投票总共押金
	@lightAddr        []byte    轻节点地址
	@communityAddr    []byte    社区节点地址
*/
// func BuildDBKeyAddrVoteLock(lightAddr, communityAddr []byte) *[]byte {
// 	key := make([]byte, 0, len(DBKEY_addr_vote_lock)+len(lightAddr)+len(communityAddr))
// 	key = append(key, lightAddr...)
// 	key = append(key, communityAddr...)
// 	return &key
// }

/*
构建大额地址黑名单
*/
func BuildAddrValueBig(addr []byte) []byte {
	return append(DBKEY_addr_value_big, addr...)
}

/*
构建冻结高度锁定表key
*/
func BuildFrozenHeight(height uint64) []byte {
	bs := utils.Uint64ToBytes(height)
	return append(DBKEY_hash_key_block_height, bs...)
}

/*
构建冻结时间锁定表key
*/
func BuildFrozenTime(ftime int64) []byte {
	bs := utils.Int64ToBytes(ftime)
	return append(DBKEY_hash_key_block_time, bs...)
}

/*
构建NFT拥有者
*/
func BuildNftOwner(nftid []byte) []byte {
	return append(DBKEY_hash_key_nft_owner, nftid...)
}

/*
构建每个地址拥有的NFT
*/
func BuildAddrToNft(addr []byte) []byte {
	return append(DBKEY_hash_key_addr_nft, addr...)
}

// 构建合约源代码
func BuildContractSource(addr []byte) []byte {
	return append([]byte(DBKEY_contract_source), addr...)
}

// 构建合约abi
func BuildContractAbi(addr []byte) []byte {
	return append([]byte(DBKEY_contract_abi), addr...)
}

// 构建合约bin
func BuildContractBin(addr []byte) []byte {
	return append([]byte(DBKEY_contract_bin), addr...)
}

// 构建合约奖励
func BuildRewardHistory(addr []byte) []byte {
	return append([]byte(DBKEY_REWARD_HISTORY), addr...)
}

// 构建区块头哈希
func BuildBlockHead(bhash []byte) []byte {
	return append(DBKEY_BlockHead, bhash...)
}

// 构建区块交易哈希
func BuildBlockTx(txhash []byte) []byte {
	return append(DBKEY_Block_tx, txhash...)
}

//// 构建轻节点最后投票高度
//func BuildLightLastVoteHeight(addr []byte) []byte {
//	return append(DBKEY_light_last_vote_height, addr...)
//}

// 构建社区奖励池
func BuildCommunityRewardPool(addr []byte) []byte {
	return append(DBKEY_community_reward_pool, addr...)
}

// 构建社区累计奖励
func BuildCommunityAllReward(addr []byte) []byte {
	return append(DBKEY_community_all_reward, addr...)
}

// 构建轻节点累计奖励
func BuildLightAllReward(addr []byte) []byte {
	return append(DBKEY_light_all_reward, addr...)
}

// 构建自定义交易事件
func BuildCustomTxEvent(txid []byte) []byte {
	return append(Address_history_custom_tx, txid...)
}

/*
构建发布token的信息
*/
func BuildKeyForPublishToken(txid []byte) []byte {
	return append([]byte(TokenInfo), txid...) //[]byte(config.TokenInfo + txidStr)
}

/*
构建swap交易记录key
*/
func BuildSwapTxKey(txid []byte) []byte {
	return append(DBKEY_swap_tx, txid...)
}

/*
构建地址绑定key
*/
func BuildAddressTxBindKey(addr []byte) []byte {
	return append(DBKEY_address_tx_bind, addr...)
}

/*
构建地址绑定冻结key
*/
func BuildAddressTxFrozenKey(addr []byte) []byte {
	return append(DBKEY_address_tx_frozen, addr...)
}

/*
构建多签集合地址key
*/
func BuildMultsignAddrSet(addr []byte) []byte {
	return append(DBKEY_multsign_set, addr...)
}

/*
构建多签请求交易key
TxItr
*/
func BuildMultsignRequestTx(txid []byte) []byte {
	return append(DBKEY_multsign_request_tx, txid...)
}
