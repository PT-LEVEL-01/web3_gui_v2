package chain_orders

import (
	"web3_gui/chain_boot/object_beans"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type GoodsFactory func() GoodsItr

type GoodsItr interface {
	LockOne() bool                                                          //锁定库存
	UnLockOne() bool                                                        //购买成功，解锁库存
	AddSalesTotal(goodsId []byte, salesTotalMany bool, total, price uint64) //修改价格，添加库存
	GetPrice() uint64                                                       //获取价格
	GetOrder(goodsId []byte, pullHeight uint64) (OrderItr, utils.ERROR)     //获取一个订单
}

type OrderItr interface {
	object_beans.ObjectItr
	GetGoodsId() []byte                 //获取商品id
	GetPrice() uint64                   //获取价格
	GetRecipient() nodeStore.AddressNet //获取收款地址
	GetPayLockBlockHeight() uint64      //获取未支付超时高度
}
