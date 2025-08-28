package imdatachain

import (
	"fmt"
	"testing"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func TestProto(*testing.T) {
	exampleProto()
}

func exampleProto() {
	base := go_protos.ImDataChainProxyBase{
		Base: &go_protos.ImProxyBase{Command: uint64(1)},
	}
	bs1, err := base.Marshal()
	fmt.Println(bs1, err)

	fmt.Println("序列化测试1")
	init := NewFirendListInit(nodeStore.AddressNet{}, nodeStore.AddressNet{})
	fmt.Printf("序列化对象:%+v"+"\n", *init)
	//fmt.Println("序列化对象", *init)
	bs, err := init.Proto()
	fmt.Println("init序列化字节序", bs, err)

	itr, ERR := ParseDataChain(*bs)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Printf("解析后的对象:%+v"+"\n", itr)

	fmt.Println("序列化测试2")
	bs2 := []byte{10, 57, 10, 16, 1, 144, 103, 178, 145, 192, 204, 59, 112, 189, 221, 86, 167, 120, 17, 191, 32, 1, 42,
		1, 1, 50, 32, 213, 195, 173, 202, 29, 8, 14, 4, 182, 211, 34, 219, 227, 26, 161, 3, 11, 1, 110, 48, 179, 132, 244, 81, 63, 216, 191, 244, 156, 101, 163, 17}
	fmt.Println("init序列化字节序", bs2)
	itr, ERR = ParseDataChain(bs2)
	if !ERR.CheckSuccess() {
		panic(ERR.String())
	}
	fmt.Printf("解析后的对象:%+v"+"\n", itr)

}
