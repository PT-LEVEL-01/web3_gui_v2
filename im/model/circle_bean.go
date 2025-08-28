package model

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"web3_gui/im/protos/go_protos"
)

type News struct {
	Class      string //
	Index      string //
	Title      string
	Content    string
	CreateTime int64
	UpdateTime int64
}

func (this *News) Proto() (*[]byte, error) {
	index, err := hex.DecodeString(this.Index)
	if err != nil {
		return nil, err
	}
	bhp := go_protos.News{
		Index:      index,
		Title:      []byte(this.Title),
		Content:    []byte(this.Content),
		CreateTime: this.CreateTime,
		UpdateTime: this.UpdateTime,
		Class:      []byte(this.Class),
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析一个新闻条目
*/
func ParseNews(bs *[]byte) (*News, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.News)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	bh := News{
		Class:      string(bhp.Class),
		Index:      hex.EncodeToString(bhp.Index),
		Title:      string(bhp.Title),
		Content:    string(bhp.Content),
		CreateTime: bhp.CreateTime,
		UpdateTime: bhp.UpdateTime,
	}
	return &bh, nil
}

type ClassCount struct {
	Name  string
	Count uint64
}

/*
解析一个博客圈子列表
*/
func ParseClassNames(bs *[]byte) ([]ClassCount, error) {
	if bs == nil {
		return nil, nil
	}
	cns := new(go_protos.ClassNames)
	// cns.ClassList = make([]*go_protos.Class,0)
	err := proto.Unmarshal(*bs, cns)
	if err != nil {
		return nil, err
	}
	classCounts := make([]ClassCount, 0)
	for _, one := range cns.ClassList {
		classOne := ClassCount{
			Name:  string(one.Name),
			Count: one.Size_,
		}
		classCounts = append(classCounts, classOne)
	}
	return classCounts, nil
}
