package config

import "web3_gui/utils"

var (
	ERROR_CODE_File_not_exist     = utils.RegErrCodeExistPanic(10000001, "文件不存在")
	ERROR_CODE_File_exist         = utils.RegErrCodeExistPanic(10000002, "文件已存在")
	ERROR_CODE_conn_node_fail_trx = utils.RegErrCodeExistPanic(10001001, "连接波场节点失败")
)
