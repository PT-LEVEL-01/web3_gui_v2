package doubleratchet

import "web3_gui/utils"

var (
	ERROR_code_dial_self = utils.RegErrCodeExistPanic(500002, "自己连接自己") //
)
