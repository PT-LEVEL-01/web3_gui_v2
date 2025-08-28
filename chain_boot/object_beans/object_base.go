package object_beans

import (
	"web3_gui/chain_boot/object_beans/object_beans_protos/protos_object_beans"
)

type ObjectItr interface {
	GetClass() uint64        //获取类型
	GetId() []byte           //获取id
	Proto() (*[]byte, error) //格式化成proto字节
}

type ObjectBase struct {
	Class uint64 //类型
}

func (this *ObjectBase) GetProto() *protos_object_beans.ObjectBase {
	base := protos_object_beans.ObjectBase{
		Class: this.Class,
	}
	return &base
}

func (this *ObjectBase) GetClass() uint64 {
	return this.Class
}

func ConvertCommonBase(base *protos_object_beans.ObjectBase) *ObjectBase {
	return &ObjectBase{
		Class: base.Class,
	}
}
