package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"io"
	"os"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type SendFileInfo struct {
	Addr        nodeStore.AddressNet //好友地址
	FileNameAbs string               //真实文件路径
	SendTime    int64                //发送时间
	Name        string               //文件名称
	MimeType    string               //文件类型
	FileSize    uint64               //文件总大小
	BlockSize   uint64               //块大小
	Hash        []byte               //文件hash
	BlockTotal  uint64               //文件块总数
	BlockIndex  uint64               //文件块编号，从0开始，连续增长的整数
	Block       []byte               //文件块内容
	IsGroup     bool                 //是否是群
}

func (this SendFileInfo) BuildDatachainFile(addrSelf nodeStore.AddressNet) (*DatachainFile, utils.ERROR) {

	start := this.BlockSize * this.BlockIndex
	end := this.BlockSize * (this.BlockIndex + 1)
	if end > this.FileSize {
		end = this.FileSize
	}

	var blockContent []byte
	if len(this.Block) == 0 {
		file, err := os.Open(this.FileNameAbs)
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return nil, utils.NewErrorSysSelf(err)
		}
		defer file.Close()

		utils.Log.Info().Msgf("文件大小:%d %d %d", this.FileSize, start, end)
		buf := make([]byte, config.DataChainBlockContentSize)

		n, err := file.ReadAt(buf, int64(start))
		if err != nil && err != io.EOF {
			return nil, utils.NewErrorSysSelf(err)
		}
		blockContent = buf[:n]
	} else {
		blockContent = this.Block[start:end]
	}
	utils.Log.Info().Str("好友地址", this.Addr.B58String()).Send()
	fileBlock := NewDatachainFile(addrSelf, this.Addr, this.SendTime, this.Name, this.MimeType,
		this.FileSize, this.Hash, this.BlockTotal, this.BlockIndex, blockContent)
	if this.IsGroup {
		fileBlock.GetProxyItr().GetBase().GroupID = this.Addr.GetAddr()
	}

	//proxyItr := fileBlock.GetProxyItr()
	//sendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	//proxyItr.GetBase().SendIndex = sendIndex
	//ERR = proxyItr.EncryptContent(shareKey)
	//if !ERR.CheckSuccess() {
	//	utils.Log.Error().Msgf("修改群 错误:%s", ERR.String())
	//	return ERR
	//}
	return fileBlock, utils.NewErrorSuccess()
}

func (this *SendFileInfo) Proto() (*[]byte, error) {
	sendText := go_protos.SendFileInfo{
		Addr:        this.Addr.GetAddr(),      //
		FileNameAbs: []byte(this.FileNameAbs), //真实文件路径
		SendTime:    this.SendTime,            //发送时间
		Name:        []byte(this.Name),        //文件名称
		MimeType:    []byte(this.MimeType),    //文件类型
		FileSize:    this.FileSize,            //文件总大小
		BlockSize:   this.BlockSize,           //块大小
		Hash:        this.Hash,                //文件hash
		BlockTotal:  this.BlockTotal,          //文件块总数
		BlockIndex:  this.BlockIndex,          //文件块编号，从0开始，连续增长的整数
		Block:       this.Block,               //文件块内容
		IsGroup:     this.IsGroup,             //是否是群
	}
	bs, err := sendText.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseSendFileInfo(bs []byte) (*SendFileInfo, error) {
	base := go_protos.SendFileInfo{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	addFriend := SendFileInfo{
		Addr:        *nodeStore.NewAddressNet(base.Addr), //
		FileNameAbs: string(base.FileNameAbs),            //真实文件路径
		SendTime:    base.SendTime,                       //发送时间
		Name:        string(base.Name),                   //文件名称
		MimeType:    string(base.MimeType),               //文件类型
		FileSize:    base.FileSize,                       //文件总大小
		BlockSize:   base.BlockSize,                      //块大小
		Hash:        base.Hash,                           //文件hash
		BlockTotal:  base.BlockTotal,                     //文件块总数
		BlockIndex:  base.BlockIndex,                     //文件块编号，从0开始，连续增长的整数
		Block:       base.Block,                          //文件块内容
		IsGroup:     base.IsGroup,                        //是否是群
	}
	return &addFriend, nil
}

/*
开始计算块切片总量
@fileSize    uint64    文件总大小
@fileSize    uint64    文件总大小
*/
func BuildBlockTotal(fileSize, blockSize uint64) uint64 {
	blockTotal := fileSize / blockSize
	if fileSize%blockSize != 0 {
		blockTotal++
	}
	return blockTotal
}
