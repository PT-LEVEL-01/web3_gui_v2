package chain_orders

import (
	"time"
	"web3_gui/chain/config"
)

const (
	Order_Overtime = time.Minute * 1 //订单超时时间

	GOODS_sharebox_file = 1 //商品类型->共享文件

)

var (
	Order_Overtime_Height uint64 = uint64(Order_Overtime / config.Mining_block_time) //把订单超时时间换算为区块高度
)
