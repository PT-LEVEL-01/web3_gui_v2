package object_beans

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"web3_gui/chain_boot/object_beans/object_beans_protos/protos_object_beans"
)

func init() {
	RegisterObjectClass(CLASS_SHAREBOX_goods_order, ParseOrderShareboxOrder)
}

type OrderShareboxOrder struct {
	ObjectBase
	OrderBase
}

func (this *OrderShareboxOrder) GetId() []byte {
	return this.Number
}

func (this *OrderShareboxOrder) Proto() (*[]byte, error) {
	proxyBase := this.GetProto()
	base := protos_object_beans.ObjectShareboxOrder{
		Base: proxyBase,
		OrderBase: &protos_object_beans.ObjectOrderBase{
			Id:                 this.Number,
			GoodsId:            this.GoodsId,              //商品id
			UserAddr:           this.UserAddr.GetAddr(),   //消费者地址
			ServerAddr:         this.ServerAddr.GetAddr(), //服务器地址
			TotalPrice:         this.TotalPrice,           //订单总金额
			ChainTx:            this.ChainTx,              //区块链上的交易
			TxHash:             this.TxHash,               //已经上链的交易hash
			CreateTime:         this.CreateTime,           //订单创建时间
			PayLockBlockHeight: this.PayLockBlockHeight,   //支付限制区块高度
		},
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *OrderShareboxOrder) ConverVO() *OrderShareboxOrderVO {
	oso := OrderShareboxOrderVO{
		Class:              this.Class,
		Number:             hex.EncodeToString(this.Number),
		GoodsId:            hex.EncodeToString(this.GoodsId),
		UserAddr:           this.UserAddr.B58String(),
		ServerAddr:         this.ServerAddr.B58String(),
		TotalPrice:         this.TotalPrice,
		ChainTx:            hex.EncodeToString(this.ChainTx),
		TxHash:             hex.EncodeToString(this.TxHash),
		CreateTime:         this.CreateTime,
		PayLockBlockHeight: this.PayLockBlockHeight,
	}
	return &oso
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseOrderShareboxOrder(bs []byte) (ObjectItr, error) {
	order := protos_object_beans.ObjectShareboxOrder{}
	err := proto.Unmarshal(bs, &order)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertCommonBase(order.Base)
	addFriend := OrderShareboxOrder{
		ObjectBase: *clientBase,
		OrderBase:  *ConvertOrderBase(order.OrderBase),
	}
	return &addFriend, nil
}

type OrderShareboxOrderVO struct {
	Class              uint64 //类型
	Number             string //订单编号
	GoodsId            string //商品id
	UserAddr           string //消费者地址
	ServerAddr         string //服务器地址
	TotalPrice         uint64 //订单总金额
	ChainTx            string //区块链上的交易
	TxHash             string //已经上链的交易hash
	CreateTime         int64  //订单创建时间
	PayLockBlockHeight uint64 //支付限制区块高度
}
