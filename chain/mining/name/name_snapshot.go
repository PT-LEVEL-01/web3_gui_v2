package name

import (
	"web3_gui/chain/mining/snapshot"
	"web3_gui/chain/protos/go_protos"

	"github.com/gogo/protobuf/proto"
)

func init() {
	snapshot.Add(new(NameSnapshot))
}

/*
负责域名快照
*/
type NameSnapshot struct {
}

/*
快照碎片对象名称
*/
func (this *NameSnapshot) SnapshotName() string {
	return "name"
}

/*
序列化内存对象
*/
func (this *NameSnapshot) SnapshotSerialize() ([]byte, error) {
	bss := go_protos.RepeatedBytes{
		Bss: make([][]byte, 0),
	}
	var err error
	names.Range(func(k, v interface{}) bool {
		nameOne := v.(Nameinfo)
		bsOne, e := nameOne.Proto()
		if e != nil {
			err = e
			return false
		}
		bss.Bss = append(bss.Bss, bsOne)
		return true
	})
	if err != nil {
		return nil, err
	}
	return bss.Marshal()
}

/*
反序列化,还原内存对象
*/
func (this *NameSnapshot) SnapshotDeSerialize(bs []byte) error {
	if bs == nil || len(bs) == 0 {
		return nil
	}
	bss := new(go_protos.RepeatedBytes)
	err := proto.Unmarshal(bs, bss)
	if err != nil {
		return err
	}
	for _, one := range bss.Bss {
		nameinfo, err := ParseNameinfo(one)
		if err != nil {
			return err
		}
		AddName(*nameinfo)
	}
	return nil
}
