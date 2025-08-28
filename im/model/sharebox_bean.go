package model

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"web3_gui/im/protos/go_protos"
)

/*
文件信息
*/
type FilePrice struct {
	Name    string //真实文件名
	Hash    []byte //hash值
	Price   uint64 //价格
	Size    int64  //文件大小
	Time    int64  //文件创建时间
	DirPath string //父文件夹路径
	IsDir   bool   //是否文件夹
}

/*
序列化
*/
func (this *FilePrice) Proto() (*[]byte, error) {
	bhp := go_protos.FilePrice{
		Name:     this.Name,
		Hash:     this.Hash,
		Price:    this.Price,
		FileSize: this.Size,
		Time:     this.Time,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

func ParseFilePrice(bs []byte) (*FilePrice, error) {
	isInfo := new(go_protos.FilePrice)
	err := proto.Unmarshal(bs, isInfo)
	if err != nil {
		return nil, err
	}
	info := FilePrice{
		Name:  isInfo.Name,     //真实文件名
		Hash:  isInfo.Hash,     //hash值
		Price: isInfo.Price,    //价格
		Size:  isInfo.FileSize, //文件大小
		Time:  isInfo.Time,     //文件创建时间
	}
	return &info, nil
}

/*
文件信息
*/
type FilePriceVO struct {
	Name    string //真实文件名
	Hash    string //hash值
	Price   uint64 //价格
	Size    int64  //文件大小
	Time    int64  //文件创建时间
	DirPath string //父文件夹路径
	IsDir   bool   //是否文件夹
}

/*
转化为VO对象
*/
func (this *FilePrice) ConverVO() *FilePriceVO {
	fpVO := FilePriceVO{
		Name:    this.Name,
		Hash:    hex.EncodeToString(this.Hash),
		Price:   this.Price,
		Size:    this.Size,
		Time:    this.Time,
		DirPath: this.DirPath,
		IsDir:   this.IsDir,
	}
	return &fpVO
}
