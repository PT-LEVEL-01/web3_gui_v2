package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_voice, ParseDatachainVoiceFactory)
}

type DatachainVoice struct {
	DataChainClientBase
	Name        string //文件名称
	MimeType    string //文件类型
	Second      int64  //语音录制了多少秒时间
	BlockBinary []byte //文件块内容
	BlockCoding string //文件块编码内容
}

func (this *DatachainVoice) CheckCmd() bool {
	return true
}

func (this *DatachainVoice) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainVoice{
		Base:        base,
		Name:        []byte(this.Name),     //文件名称
		MimeType:    []byte(this.MimeType), //文件类型
		Second:      this.Second,           //
		BlockBinary: this.BlockBinary,      //文件二进制内容
		BlockCoding: this.BlockCoding,      //文件编码内容
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
func ParseDatachainVoiceFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainVoice{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	addFriend := DatachainVoice{
		DataChainClientBase: *clientBase,
		Name:                string(base.Name),     //文件名称
		MimeType:            string(base.MimeType), //文件类型
		Second:              base.Second,           //
		BlockBinary:         base.BlockBinary,      //文件块内容
		BlockCoding:         base.BlockCoding,      //文件块编码内容
	}
	return &addFriend, nil
}

func NewDatachainVoice(addrSelf, addrFirend nodeStore.AddressNet, name, mimeType string, second int64, blockBinary []byte,
	blockCoding string) *DatachainVoice {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_voice, addrSelf, addrFirend)
	voice := DatachainVoice{
		DataChainClientBase: *clientBase,
		Name:                name,        //文件名称
		MimeType:            mimeType,    //文件类型
		Second:              second,      //
		BlockBinary:         blockBinary, //文件块内容
		BlockCoding:         blockCoding, //文件块编码内容
	}
	forward.clientItr = &voice
	voice.SetProxyItr(forward)
	return &voice
}
