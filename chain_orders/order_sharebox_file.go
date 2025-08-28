package chain_orders

import (
	"github.com/oklog/ulid/v2"
	"time"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	"web3_gui/im/db"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func init() {
	RegisterGoodsClass(GOODS_sharebox_file, NewShareboxFileGoods)
}

type ShareboxFileGoods struct {
	GoodsBase
}

func NewShareboxFileGoods() GoodsItr {
	sfg := ShareboxFileGoods{
		GoodsBase: *NewGoodsBase(),
	}
	return &sfg
}

/*
获取价格
*/
func (this *ShareboxFileGoods) GetOrder(goodsId []byte, pullHeight uint64) (OrderItr, utils.ERROR) {
	if !this.LockOne() {
		return nil, utils.NewErrorBus(config.ERROR_CODE_order_goods_soldOut, "")
	}
	addr := config.Node.Keystore.GetCoinAddrAll()[0]
	oso := OrderShareboxFile{
		object_beans.OrderShareboxOrder{
			ObjectBase: object_beans.ObjectBase{Class: object_beans.CLASS_SHAREBOX_goods_order},
			OrderBase: object_beans.OrderBase{
				Number:             ulid.Make().Bytes(),
				GoodsId:            goodsId,
				UserAddr:           nodeStore.AddressNet{},
				ServerAddr:         *nodeStore.NewAddressNet(addr.Addr.Bytes()),
				TotalPrice:         this.GetPrice(),
				ChainTx:            nil,
				TxHash:             nil,
				CreateTime:         time.Now().Unix(),
				PayLockBlockHeight: pullHeight + Order_Overtime_Height,
			},
		},
	}
	ERR := db.Sharebox_server_SaveOrderUnpaid(&oso.OrderShareboxOrder)
	if ERR.CheckFail() {
		this.UnLockOne()
		return nil, ERR
	}
	return &oso, utils.NewErrorSuccess()
}

type OrderShareboxFile struct {
	object_beans.OrderShareboxOrder
}

func (this *OrderShareboxFile) GetPrice() uint64 {
	return this.TotalPrice
}

func (this *OrderShareboxFile) GetGoodsId() []byte {
	return this.GoodsId
}

/*
获取收款人地址
*/
func (this *OrderShareboxFile) GetRecipient() nodeStore.AddressNet {
	return this.ServerAddr
}

/*
获取未支付超时高度
*/
func (this *OrderShareboxFile) GetPayLockBlockHeight() uint64 {
	return this.PayLockBlockHeight
}
