package model

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/im/protos/go_protos"
)

func BytesProto(list1, list2 [][]byte) ([]byte, error) {
	bss := go_protos.Bytes{
		List:  list1,
		List2: list2,
	}
	return bss.Marshal()
}

func ParseBytes(bs []byte) ([][]byte, [][]byte, error) {
	bss := new(go_protos.Bytes)
	err := proto.Unmarshal(bs, bss)
	if err != nil {
		return nil, nil, err
	}
	return bss.List, bss.List2, nil
}
