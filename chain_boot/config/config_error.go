package config

import "web3_gui/utils"

var (
	ERROR_CODE_CHAIN_not_ready                          = utils.RegErrCodeExistPanic(90001, "链端未准备就绪")
	ERROR_CODE_CHAIN_balance_not_enough                 = utils.RegErrCodeExistPanic(90002, "余额不足")
	ERROR_CODE_CHAIN_address_not_exist                  = utils.RegErrCodeExistPanic(90003, "地址不存在")
	ERROR_CODE_CHAIN_leveldb_null                       = utils.RegErrCodeExistPanic(90004, "数据库是空")
	ERROR_CODE_CHAIN_not_found                          = utils.RegErrCodeExistPanic(90005, "查询内容未找到")
	ERROR_CODE_CHAIN_witness_deposit_exist              = utils.RegErrCodeExistPanic(90006, "见证人押金已经存在")
	ERROR_CODE_CHAIN_witness_deposit_not_exist          = utils.RegErrCodeExistPanic(90007, "见证人押金已经不存在")
	ERROR_CODE_CHAIN_witness_deposit_quantity_incorrect = utils.RegErrCodeExistPanic(90008, "见证人押金数量不对")
	ERROR_CODE_CHAIN_Memory_not_enough                  = utils.RegErrCodeExistPanic(90009, "内存不足")
	ERROR_CODE_CHAIN_too_low                            = utils.RegErrCodeExistPanic(90010, "太低")
	ERROR_CODE_CHAIN_too_high                           = utils.RegErrCodeExistPanic(90011, "太高")

	ERROR_CODE_CHAIN_not_test_net            = utils.RegErrCodeExistPanic(90012, "不是测试网络")
	ERROR_CODE_CHAIN_recv_interval_too_short = utils.RegErrCodeExistPanic(90013, "领取测试币间隔时间太短")

	ERROR_CODE_CHAIN_contract_Payload_size_fail  = utils.RegErrCodeExistPanic(90014, "合约交易payload长度错误")
	ERROR_CODE_CHAIN_comment_size_too_long       = utils.RegErrCodeExistPanic(90015, "交易comment信息长度太长")
	ERROR_CODE_CHAIN_community_depositin_close   = utils.RegErrCodeExistPanic(90016, "社区节点质押未开放")
	ERROR_CODE_CHAIN_light_depositin_close       = utils.RegErrCodeExistPanic(90017, "轻节点质押未开放")
	ERROR_CODE_CHAIN_address_not_light           = utils.RegErrCodeExistPanic(90018, "地址不是轻节点")
	ERROR_CODE_CHAIN_address_not_community       = utils.RegErrCodeExistPanic(90019, "地址不是社区节点")
	ERROR_CODE_CHAIN_address_not_witness         = utils.RegErrCodeExistPanic(90020, "地址不是见证人节点")
	ERROR_CODE_CHAIN_community_reward_load_async = utils.RegErrCodeExistPanic(90021, "社区奖励正在异步加载")

	ERROR_CODE_CHAIN_token_publish_min        = utils.RegErrCodeExistPanic(91001, "发布token最少数量")
	ERROR_CODE_CHAIN_token_owner_nil          = utils.RegErrCodeExistPanic(91002, "发布的token所有人是空")
	ERROR_CODE_CHAIN_token_balance_not_enough = utils.RegErrCodeExistPanic(91003, "token转账余额不足")

	ERROR_CODE_coinAddr_fail        = utils.RegErrCodeExistPanic(70001, "地址不完整，地址格式错误") //
	ERROR_CODE_coinAddr_not_exist   = utils.RegErrCodeExistPanic(70002, "地址不存在")        //
	ERROR_CODE_BalanceNotEnough     = utils.RegErrCodeExistPanic(70003, "余额不足")         //
	ERROR_CODE_not_be_zero          = utils.RegErrCodeExistPanic(70004, "转账金额不能为0")     //
	ERROR_CODE_VoteNotOpen          = utils.RegErrCodeExistPanic(70005, "投票功能还未开放")     //
	ERROR_CODE_WitnessDepositExist  = utils.RegErrCodeExistPanic(70006, "见证人质押已经存在")    //
	ERROR_CODE_WitnessDepositLess   = utils.RegErrCodeExistPanic(70007, "见证人押金数量不对")    //
	ERROR_CODE_name_exist           = utils.RegErrCodeExistPanic(70008, "域名已经存在")       //
	ERROR_CODE_params_format_fail   = utils.RegErrCodeExistPanic(70009, "参数格式错误")       //
	ERROR_CODE_GasTooLittle         = utils.RegErrCodeExistPanic(70010, "手续费太少")        //
	ERROR_CODE_Tx_LockHeight_TooLow = utils.RegErrCodeExistPanic(70011, "交易的上链高度太低")    //
	ERROR_CODE_Tx_Nonce_fail        = utils.RegErrCodeExistPanic(70012, "交易中nonce错误")   //
	ERROR_CODE_Tx_Vin_fail          = utils.RegErrCodeExistPanic(70013, "交易中vin错误")     //
	ERROR_CODE_Tx_Sign_fail         = utils.RegErrCodeExistPanic(70014, "交易签名错误")       //
	ERROR_CODE_Tx_Domain_fail       = utils.RegErrCodeExistPanic(70015, "交易域名错误")       //

)
