package mining

import (
	"errors"
)

var (
	ERROR_repeat_import_block                = errors.New("Repeat import block")                                                 //导入重复的区块
	ERROR_fork_import_block                  = errors.New("The front block cannot be found, and the new block is discontinuous") //导入的区块分叉了
	ERROR_import_block_height_not_continuity = errors.New("Import block height discontinuity")                                   //导入的区块高度不连续
	ERROR_not_strict_import_block            = errors.New("Not strict import block")                                             //不严格的导入区块顺序
	ERROR_swap_tx_check_fail                 = errors.New("swap tx check fail")                                                  //撮合交易检查失败
	ERROR_import_block_illegal_tx            = errors.New("import block illegal transaction")                                    //导入的区块交易不合法
)
