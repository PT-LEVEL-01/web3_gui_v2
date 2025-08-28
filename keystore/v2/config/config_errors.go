package config

import (
	"web3_gui/utils"
)

var (
	ERROR_code_Invalid_mnenomic           = utils.RegErrCodeExistPanic(40001, "助记词无效")               //
	ERROR_code_password_fail              = utils.RegErrCodeExistPanic(40002, "密码错误")                //
	ERROR_code_seed_password_fail         = utils.RegErrCodeExistPanic(40003, "种子密码错误")              //
	ERROR_code_coinAddr_password_fail     = utils.RegErrCodeExistPanic(40004, "钱包地址密码错误")            //
	ERROR_code_netAddr_password_fail      = utils.RegErrCodeExistPanic(40005, "网络地址密码错误")            //
	ERROR_code_dhkey_password_fail        = utils.RegErrCodeExistPanic(40006, "协商密钥密码错误")            //
	ERROR_code_wallet_incomplete          = utils.RegErrCodeExistPanic(40007, "密钥文件损坏，不完整")          //
	ERROR_code_keystore_index_maximum     = utils.RegErrCodeExistPanic(40008, "传入的密钥库索引太大，对应密钥库不存在") //
	ERROR_code_wallet_file_not_exist      = utils.RegErrCodeExistPanic(40009, "钱包文件不存在")             //
	ERROR_code_wallet_file_exist          = utils.RegErrCodeExistPanic(40010, "钱包文件已存在")             //
	ERROR_code_addr_not_found             = utils.RegErrCodeExistPanic(40011, "未找到此地址")              //
	ERROR_code_size_too_small             = utils.RegErrCodeExistPanic(40012, "字节长度太小")              //
	ERROR_code_salt_size_too_small        = utils.RegErrCodeExistPanic(40013, "加解密盐字节长度太小")          //
	ERROR_code_version_old                = utils.RegErrCodeExistPanic(40014, "版本太旧，需要升级新版本")        //
	ERROR_code_coin_type_not_exist        = utils.RegErrCodeExistPanic(40015, "CoinType不存在")         //
	ERROR_code_coin_type_addr_not_achieve = utils.RegErrCodeExistPanic(40016, "CoinType地址生成算法未实现")   //
	ERROR_code_unusable_seed              = utils.RegErrCodeExistPanic(40017, "不能使用的种子")             //
)
