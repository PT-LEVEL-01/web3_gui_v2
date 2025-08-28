package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_file, ParseDatachainFileFactory)
}

type DatachainFile struct {
	DataChainClientBase
	SendTime   int64  //发送时间
	Name       string //文件名称
	MimeType   string //文件类型
	Size       uint64 //文件总大小
	Hash       []byte //文件hash
	BlockTotal uint64 //文件块总数
	BlockIndex uint64 //文件块编号，从0开始，连续增长的整数
	Block      []byte //文件块内容
}

func (this *DatachainFile) CheckCmd() bool {
	return true
}

func (this *DatachainFile) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainFile{
		Base:       base,
		SendTime:   uint64(this.SendTime), //
		Name:       []byte(this.Name),     //文件名称
		MimeType:   []byte(this.MimeType), //文件类型
		Size_:      this.Size,             //文件总大小
		Hash:       this.Hash,             //文件hash
		BlockTotal: this.BlockTotal,       //文件块总数
		BlockIndex: this.BlockIndex,       //文件块编号，从0开始，连续增长的整数
		Block:      this.Block,            //文件块内容
	}
	bs, err := sendText.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDatachainFileFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainFile{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainFile{
		DataChainClientBase: *clientBase,
		SendTime:            int64(base.SendTime),  //
		Name:                string(base.Name),     //文件名称
		MimeType:            string(base.MimeType), //文件类型
		Size:                base.Size_,            //文件总大小
		Hash:                base.Hash,             //文件hash
		BlockTotal:          base.BlockTotal,       //文件块总数
		BlockIndex:          base.BlockIndex,       //文件块编号，从0开始，连续增长的整数
		Block:               base.Block,            //文件块内容
	}
	return &addFriend, nil
}

func NewDatachainFile(addrSelf, addrFirend nodeStore.AddressNet, sendTime int64, name, mimeType string, size uint64, hash []byte,
	blockTotal, blockIndex uint64, block []byte) *DatachainFile {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_file, addrSelf, addrFirend)
	addFriend := DatachainFile{
		DataChainClientBase: *clientBase,
		SendTime:            sendTime,   //
		Name:                name,       //文件名称
		MimeType:            mimeType,   //文件类型
		Size:                size,       //文件总大小
		Hash:                hash,       //文件hash
		BlockTotal:          blockTotal, //文件块总数
		BlockIndex:          blockIndex, //文件块编号，从0开始，连续增长的整数
		Block:               block,      //文件块内容
	}
	forward.clientItr = &addFriend
	addFriend.SetProxyItr(forward)
	return &addFriend
}
