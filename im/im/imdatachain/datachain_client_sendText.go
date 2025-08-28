package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"time"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterCmdClient(config.IMPROXY_Command_client_sendText, ParseSendTextFactory)
}

/*
发送文本消息
*/
type DataChainSendText struct {
	DataChainClientBase
	//Type          uint64               //消息类型
	//FromIsSelf    bool                 //是否自己发出的
	//From          nodeStore.AddressNet //发送者
	//To            nodeStore.AddressNet //接收者
	Content []byte //消息内容
	//PullAndPushID uint64               //上传或者下载ID
	SendTime int64 //发送的时间
	//SendID        []byte               //消息唯一ID
	//RecvID        []byte               //消息唯一ID
	QuoteID []byte //引用消息ID
	//State         int                  //消息状态。1=未发送;2=已送达;3=已读;
	//Index         uint64               //数据库中消息索引
	//EncryptType   uint32               //加密类型。0=未加密;1=AES加密;2=;
}

func (this *DataChainSendText) CheckCmd() bool {
	return true
}

func (this *DataChainSendText) Proto() (*[]byte, error) {
	base := this.DataChainClientBase.GetProto()
	sendText := go_protos.ImDataChainSendText{
		Base:     base,
		Content:  this.Content,
		SendTime: this.SendTime,
		QuoteID:  this.QuoteID,
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
func ParseSendTextFactory(bs []byte) (DataChainClientItr, error) {
	base := go_protos.ImDataChainSendText{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertClientBase(base.Base)
	sendText := DataChainSendText{
		DataChainClientBase: *clientBase,
		Content:             base.Content,  //消息内容
		SendTime:            base.SendTime, //发送的时间
		QuoteID:             base.QuoteID,  //引用消息ID
	}
	return &sendText, nil
}

func NewDataChainSendText(addrSelf, addrFirend nodeStore.AddressNet, text string, quotoId []byte) *DataChainSendText {
	forward := NewDataChainProxyForward(addrSelf, addrFirend)
	clientBase := NewDataChainClientBase(config.IMPROXY_Command_client_sendText, addrSelf, addrFirend)
	sendText := DataChainSendText{DataChainClientBase: *clientBase}
	sendText.Content = []byte(text)
	sendText.SendTime = time.Now().Unix()
	sendText.QuoteID = quotoId
	forward.clientItr = &sendText
	sendText.SetProxyItr(forward)
	return &sendText
}
