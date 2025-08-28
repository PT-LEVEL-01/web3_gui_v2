package object_beans

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/chain_boot/object_beans/object_beans_protos/protos_object_beans"
)

func init() {
	RegisterObjectClass(CLASS_COMMON_order_number, ParseCommonOrder)
}

type CommonOrder struct {
	ObjectBase
	Number []byte //订单编号
}

func NewCommonOrder(number []byte) *CommonOrder {
	return &CommonOrder{
		ObjectBase: ObjectBase{CLASS_COMMON_order_number},
		Number:     number,
	}
}

func (this *CommonOrder) Proto() (*[]byte, error) {
	proxyBase := this.GetProto()
	base := protos_object_beans.ObjectCommonOrderNumber{
		Base: proxyBase,
		Id:   this.Number,
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *CommonOrder) GetId() []byte {
	return this.Number
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseCommonOrder(bs []byte) (ObjectItr, error) {
	order := protos_object_beans.ObjectCommonOrderNumber{}
	err := proto.Unmarshal(bs, &order)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertCommonBase(order.Base)
	addFriend := CommonOrder{
		ObjectBase: *clientBase,
		Number:     order.Id,
	}
	return &addFriend, nil
}
