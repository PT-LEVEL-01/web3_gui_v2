package file_transfer

import (
	"github.com/gogo/protobuf/proto"
	"os"
	"time"
	"web3_gui/im/protos/go_protos"
	"web3_gui/utils"
)

type FileIndex struct {
	ClassID        uint64 //文件传输单元ID
	ID             []byte //文件加密后的hash值
	UserAddr       []byte //所属用户地址
	Version        uint16 //版本号
	Name           string //真实文件名称
	FileSize       uint64 //文件总大小
	Time           int64  //创建时间
	RemoteFilePath string //远端文件的路径
	SupplierID     []byte //提供者ID
	PullID         []byte //文件下载者ID
	OffsetIndex    uint64 //已经下载的大小
}

func NewFileIndex(ft *FileTransferTask) (*FileIndex, error) {
	fileInfo, err := os.Stat(ft.LocalFilePath)
	if err != nil {
		return nil, err
	}
	hashBs, err := utils.FileSHA3_256(ft.LocalFilePath)
	if err != nil {
		return nil, err
	}
	fileIndex := FileIndex{
		ClassID: ft.ClassID,
		ID:      hashBs, //文件加密后的hash值
		//UserAddr []byte //所属用户地址
		//Version  uint16 //版本号
		Name:     fileInfo.Name(),         //真实文件名称
		FileSize: uint64(fileInfo.Size()), //文件总大小
		Time:     time.Now().Unix(),       //创建时间
		//RemoteFilePath: ft.LocalFilePath,        //文件的路径
		SupplierID: ft.SupplierID, //
		PullID:     ft.PullID,     //
	}
	return &fileIndex, nil
}

func (this *FileIndex) Proto() (*[]byte, error) {
	bhp := go_protos.FileTransferTask{
		ClassID:       this.ClassID,         //文件传输单元ID
		ID:            this.ID,              //文件加密后的hash值
		RemoteAddr:    this.UserAddr,        //所属用户地址
		Version:       uint64(this.Version), //版本号
		Name:          this.Name,            //真实文件名称
		FileSize:      this.FileSize,        //文件总大小
		CreateTime:    this.Time,            //
		LocalFilePath: this.RemoteFilePath,  //文件的路径
		SupplierID:    this.SupplierID,      //提供者ID
		PullID:        this.PullID,          //文件下载者ID
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析为文件索引
*/
func ParseFileIndex(bs []byte) (*FileIndex, error) {
	if bs == nil || len(bs) == 0 {
		return nil, nil
	}
	sdi := new(go_protos.FileTransferTask)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	fi := FileIndex{
		ClassID:        sdi.ClassID,         //文件传输单元ID
		ID:             sdi.ID,              //文件加密后的hash值
		UserAddr:       sdi.RemoteAddr,      //所属用户地址
		Version:        uint16(sdi.Version), //版本号
		Name:           sdi.Name,            //真实文件名称
		FileSize:       sdi.FileSize,        //文件总大小
		Time:           sdi.CreateTime,      //创建时间
		RemoteFilePath: sdi.LocalFilePath,   //文件的路径
		SupplierID:     sdi.SupplierID,      //提供者ID
		PullID:         sdi.PullID,          //文件下载者ID
	}
	return &fi, nil
}

type FileChunk struct {
	ClassID     uint64 //文件传输单元ID
	SupplierID  []byte //提供者ID
	PullID      []byte //文件下载者ID
	OffsetIndex uint64 //块起始位置偏移量
	ChunkSize   uint64 //块大小
	Data        []byte //文件块数据
}

func (this *FileChunk) Proto() (*[]byte, error) {
	bhp := go_protos.FileChunk{
		ClassID:     this.ClassID,     //文件传输单元ID
		SupplierID:  this.SupplierID,  //提供者ID
		PullID:      this.PullID,      //文件下载者ID
		OffsetIndex: this.OffsetIndex, //块起始位置偏移量
		ChunkSize:   this.ChunkSize,   //块大小
		Data:        this.Data,        //文件块数据
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析为文件块
*/
func ParseFileChunk(bs []byte) (*FileChunk, error) {
	if bs == nil || len(bs) == 0 {
		return nil, nil
	}
	sdi := new(go_protos.FileChunk)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	fi := FileChunk{
		ClassID:     sdi.ClassID,     //文件传输单元ID
		SupplierID:  sdi.SupplierID,  //提供者ID
		PullID:      sdi.PullID,      //文件下载者ID
		OffsetIndex: sdi.OffsetIndex, //块起始位置偏移量
		ChunkSize:   sdi.ChunkSize,   //块大小
		Data:        sdi.Data,        //文件块数据
	}
	return &fi, nil
}
