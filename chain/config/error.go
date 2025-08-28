package config

import (
	"errors"
	"fmt"
)

const (
	ERROR_fail = 5005
)

var (
	ERROR_db_not_exist             = errors.New("db key not find")          //数据库中未查询到结果
	ERROR_chain_sysn_block_fail    = errors.New("sync fail,not find block") //同步失败，未找到区块
	ERROR_chain_sync_block_timeout = errors.New("sync block timeout")       //同步区块超时
	ERROR_wait_msg_timeout         = errors.New("wait message timeout")     //等待消息返回超时
	ERROR_not_font_witness         = errors.New("not font witness")         //未找到这个见证人
	ERROR_advanced_by_build_block  = errors.New("advanced by build block")  //构建区块时间提前了
	ERROR_block_height_equality    = errors.New("block height equality")    //区块高度相同了

	ERROR_witness_deposit_exist = errors.New("Deposit cannot be paid repeatedly")                     //见证人押金已经存在
	ERROR_witness_deposit_less  = errors.New("Deposit not less than")                                 //见证人押金太少
	ERROR_deposit_witness       = errors.New("deposit shoud be:" + fmt.Sprintf("%d", Mining_deposit)) //见证人押金数量
	ERROR_deposit_not_exist     = errors.New("deposit not exist")                                     //没有缴纳押金
	ERROR_deposit_exist         = errors.New("deposit exist")                                         //押金已经存在
	ERROR_deposit_light_vote    = errors.New("tx vote feild error")                                   //轻节点押金中Vote字段因该为空
	ERROR_vote_amount_error     = errors.New("tx vote amount error")                                  //押金错误
	ERROR_rate_error            = errors.New("reward rate error")                                     //分奖比例错误
	ERROR_cancel_vote_too_often = errors.New("cancel vote too often")                                 //取消投票太频繁
	ERROR_role_type             = errors.New("error role type")                                       //错误的角色类型
	ERROR_has_no_reward         = errors.New("has no reward")                                         //没有奖励

	ERROR_password_fail            = errors.New("password fail") //密码错误
	ERROR_wallet_password_fail     = errors.New("wallet password fail")
	ERROR_not_enough               = errors.New("balance is not enough")                   //余额不足
	ERROR_token_not_enough         = errors.New("token balance is not enough")             //token余额不足
	ERROR_public_key_not_exist     = errors.New("not find public key")                     //未找到公钥
	ERROR_amount_zero              = errors.New("Transfer amount cannot be 0")             //转账金额不能为0
	ERROR_tx_not_exist             = errors.New("Transaction not found")                   //未找到交易
	ERROR_tx_format_fail           = errors.New("Error parsing transaction")               //解析交易错误
	ERROR_tx_Repetitive_vin        = errors.New("Duplicate VIN in transaction")            //交易中有重复的vin
	ERROR_tx_is_use                = errors.New("Transaction has been used")               //交易已经被使用
	ERROR_tx_fail                  = errors.New("Transaction error")                       //交易错误
	ERROR_tx_lockheight            = errors.New("Lock height error")                       //交易锁定高度错误
	ERROR_tx_frozenheight          = errors.New("frozen height error")                     //交易冻结高度错误
	ERROR_public_and_addr_notMatch = errors.New("The public key and address do not match") //公钥和地址不匹配
	ERROR_sign_fail                = errors.New("Signature error")                         //签名错误
	ERROR_check_sign_not_pass      = errors.New("check signature not pass")                //签名不满足
	ERROR_vote_exist               = errors.New("Vote already exists")                     //投票已经存在
	ERROR_pay_vin_too_much         = errors.New("vin too much")                            //交易中vin数量过多
	ERROR_pay_vout_too_much        = errors.New("vout too much")                           //交易中vou数量过多

	ERROR_name_deposit = errors.New("Domain name deposit is required at least" + fmt.Sprintf("%d", Mining_name_deposit_min/1e8)) //域名押金最少需要 n

	ERROR_name_not_self      = errors.New("Domain name does not belong to itself") //域名不属于自己
	ERROR_name_exist         = errors.New("Domain name already exists")            //域名已经存在
	ERROR_name_not_exist     = errors.New("Domain name does not exist")            //域名不存在
	ERROR_get_sign_data_fail = errors.New("Error getting signed data")             //获取签名的数据时出错
	ERROR_params_not_enough  = errors.New("params not enough")                     //参数不够
	ERROR_params_fail        = errors.New("params fail")                           //参数错误
	ERROR_token_min_fail     = errors.New("params token min fail")                 //小于token发行最少数量

	ERROR_get_node_conn_fail = errors.New("get node conn fail") //获取连接失败

	ERROR_get_reward_count_sync     = errors.New("get reward count sync")                   //正在异步统计社区节点奖励中
	ERROR_vote_reward_addr_disunity = errors.New("vote reward address disunity")            //分配给轻节点的奖励，交易中扣款地址不统一
	ERROR_pay_nonce_is_nil          = errors.New("nonce is nil")                            //
	ERROR_addr_value_big            = errors.New("The balance of the address is too large") //地址的余额太大

	ERROR_tx_nft_not_our_own      = errors.New("The NFT not our own")      //这个NFT ID不属于自己
	ERROR_tx_nft_exchange_equally = errors.New("The NFT exchange equally") //这个NFT交换ID相同

	ERROR_tx_limiter_full        = errors.New("tx limiter is full")
	ERROR_tx_check_fail          = errors.New("tx check fail")
	ERROR_token_invalid_order_id = errors.New("token invalid order id")
	ERROR_tx_multaddr_not_found  = errors.New("mult addr not found") //未找到交易
	ERROR_tx_multaddr_pending    = errors.New("mult tx pending")     //多签交易待处理
	ERROR_tx_multaddr_existed    = errors.New("mult addr existed")   //多签地址已经存在
	ERROR_tx_gas_too_little      = errors.New("tx gas too little")   //多签地址已经存在
	ERROR_tx_name_not_owner      = errors.New("name not owner")      //这个域名不属于自己
)
