package object_beans

import (
	"web3_gui/chain_boot/object_beans/object_beans_protos/protos_object_beans"
	nodeStore "web3_gui/libp2parea/v2/node_store"
)

type OrderBase struct {
	Number             []byte               //订单编号
	GoodsId            []byte               //商品id
	UserAddr           nodeStore.AddressNet //消费者地址
	ServerAddr         nodeStore.AddressNet //服务器地址
	TotalPrice         uint64               //订单总金额
	ChainTx            []byte               //区块链上的交易
	TxHash             []byte               //已经上链的交易hash
	CreateTime         int64                //订单创建时间
	PayLockBlockHeight uint64               //支付限制区块高度
}

//func NewOrderBase() *OrderBase {
//	gb := OrderBase{
//		Number             []byte               //订单编号
//		GoodsId            []byte               //商品id
//		UserAddr           nodeStore.AddressNet //消费者地址
//		ServerAddr         nodeStore.AddressNet //服务器地址
//		TotalPrice         uint64               //订单总金额
//		ChainTx            []byte               //区块链上的交易
//		TxHash             []byte               //已经上链的交易hash
//		CreateTime         int64                //订单创建时间
//		PayLockBlockHeight uint64               //支付限制区块高度
//	}
//	return &gb
//}

func (this *OrderBase) GetPrice() uint64 {
	return this.TotalPrice
}

func (this *OrderBase) GetGoodsId() []byte {
	return this.GoodsId
}

/*
获取收款人地址
*/
func (this *OrderBase) GetRecipient() nodeStore.AddressNet {
	return this.ServerAddr
}

/*
获取未支付超时高度
*/
func (this *OrderBase) GetPayLockBlockHeight() uint64 {
	return this.PayLockBlockHeight
}

func ConvertOrderBase(base *protos_object_beans.ObjectOrderBase) *OrderBase {
	ob := &OrderBase{
		Number:  base.Id,
		GoodsId: base.GoodsId,
		//UserAddr:           base.UserAddr,
		//ServerAddr:         nodeStore.AddressNet{},
		TotalPrice:         base.TotalPrice,
		ChainTx:            base.ChainTx,
		TxHash:             base.TxHash,
		CreateTime:         base.CreateTime,
		PayLockBlockHeight: base.PayLockBlockHeight,
	}
	if base.UserAddr != nil && len(base.UserAddr) > 0 {
		ob.UserAddr = *nodeStore.NewAddressNet(base.UserAddr)
	}
	if base.ServerAddr != nil && len(base.ServerAddr) > 0 {
		ob.ServerAddr = *nodeStore.NewAddressNet(base.ServerAddr)
	}
	return ob
}
