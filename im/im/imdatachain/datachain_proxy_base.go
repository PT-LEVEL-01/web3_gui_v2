package imdatachain

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"github.com/oklog/ulid/v2"
	"math/big"
	"time"
	"web3_gui/im/protos/go_protos"
	"web3_gui/keystore/v1/base58"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type DataChainProxyItr interface {
	Proto() (*[]byte, error)                  //格式化成proto字节
	SetIndex(i big.Int)                       //设置消息的index索引
	GetIndex() (i big.Int)                    //获取消息的index索引
	GetID() []byte                            //获取唯一ID
	GetSendID() []byte                        //获取消息发送者ID
	GetAddrFrom() nodeStore.AddressNet        //获取消息发送者
	GetAddrTo() nodeStore.AddressNet          //获取消息接收者
	BuildHash()                               //构建消息的hash值
	GetHash() []byte                          //获取消息的hash值
	SetPreHash(preHash []byte)                //设置消息前一个消息的hash
	GetPreHash() []byte                       //获取消息前一个消息的hash
	CheckHash() bool                          //验证消息hash值
	GetProxyClientAddr() nodeStore.AddressNet //获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
	Forward() bool                            //本条消息是否要转发给目标用户
	SetStatus(status uint8)                   //设置消息状态
	EncryptContent(key []byte) utils.ERROR    //加密消息内容
	DecryptContent(key []byte) utils.ERROR    //解密消息内容
	GetClientItr() DataChainClientItr         //获取解密消息内容
	GetCmd() uint32                           //获取命令
	GetBase() *DataChainProxyBase             //
	GetSendTime() int64                       //
	GetRecvTime() int64                       //
}

type DataChainProxyBase struct {
	ID        []byte   //每条消息都有一个唯一ID
	SendID    []byte   //发送者的消息id，用于消息去重
	PreHash   []byte   //
	Hash      []byte   //
	Command   uint32   //日志指令
	Index     *big.Int //日志记录连续的id
	SendIndex *big.Int //聊天记录发送者的发送index
	//SendIndexReset  bool                 //是否重置sendIndex，当加好友申请和同意好友申请的时候，需要重置sendIndex
	AddrFrom        nodeStore.AddressNet //发送者
	AddrTo          nodeStore.AddressNet //发送给谁
	AddrProxyServer nodeStore.AddressNet //代理服务器节点
	GroupID         []byte               //群ID，以此判断是群消息还是个人消息，创建者地址和创建时间的Hash
	EncryptType     uint32               //加密类型。0=未加密;1=AES加密;2=;
	Content         []byte               //内容
	clientItr       DataChainClientItr   //加密内容
	Sign            []byte               //签名
	Status          uint8                //消息状态。0=保存到本地;1=发送成功;2=;
	DhPuk           []byte               //添加好友的时候，主动发送自己的公钥，用于解密
	SendTime        int64                //发送者时间
	RecvTime        int64                //接收者时间
}

func NewDataChainProxyBase(cmd uint32, addrFrom, addrTo, addrProxyServer nodeStore.AddressNet) *DataChainProxyBase {
	id := ulid.Make().Bytes()
	//fmt.Println("ulid", id)
	base := &DataChainProxyBase{
		ID:        id,
		SendID:    nil,
		PreHash:   nil,
		Hash:      nil,
		Command:   cmd,
		Index:     nil,
		SendIndex: nil, //聊天记录发送者的发送index
		//SendIndexReset:  false, //
		AddrFrom:        addrFrom,
		AddrTo:          addrTo,
		AddrProxyServer: addrProxyServer,
		GroupID:         nil, //群ID，以此判断是群消息还是个人消息，创建者地址和创建时间的Hash
		EncryptType:     0,   //加密类型。0=未加密;1=AES加密;2=;
		Content:         nil,
		Sign:            nil,
		//Status:          config.MSG_GUI_state_not_send,
		Status:   0,
		DhPuk:    nil,
		SendTime: time.Now().Unix(),
	}
	return base
}

func (this *DataChainProxyBase) GetProto() *go_protos.ImProxyBase {
	base := go_protos.ImProxyBase{
		ID:      this.ID,
		SendID:  this.SendID,
		PreHash: this.PreHash,
		Hash:    this.Hash,
		Command: uint64(this.Command),
		//Index:           this.Index.Bytes(),
		//SendIndex:       this.SendIndex.Bytes(),
		AddrFrom:        this.AddrFrom.GetAddr(),
		AddrTo:          this.AddrTo.GetAddr(),
		AddrProxyServer: this.AddrProxyServer.GetAddr(),
		GroupID:         this.GroupID,
		EncryptType:     uint64(this.EncryptType),
		Content:         this.Content,
		Sign:            this.Sign,
		Status:          uint64(this.Status),
		DhPuk:           this.DhPuk,
		SendTime:        this.SendTime,
		RecvTime:        this.RecvTime,
	}
	if this.Index != nil {
		base.Index = this.Index.Bytes()
	}
	if this.SendIndex != nil {
		base.SendIndex = this.SendIndex.Bytes()
	}
	return &base
}

func (this *DataChainProxyBase) Proto() (*[]byte, error) {
	proxyBase := this.GetProto()
	base := go_protos.ImDataChainProxyBase{
		Base: proxyBase,
	}
	if this.Index != nil {
		base.Base.Index = this.Index.Bytes()
	}
	if this.SendIndex != nil {
		base.Base.SendIndex = this.SendIndex.Bytes()
	}
	//for _, one := range this.EncryptType {
	//	base.Base.EncryptType = append(base.Base.EncryptType, uint64(one))
	//}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *DataChainProxyBase) SetIndex(i big.Int) {
	this.Index = &i
}
func (this *DataChainProxyBase) GetIndex() (i big.Int) {

	return *this.Index
}
func (this *DataChainProxyBase) GetID() []byte {
	return this.ID
}
func (this *DataChainProxyBase) GetSendID() []byte {
	return this.SendID
}
func (this *DataChainProxyBase) GetAddrFrom() nodeStore.AddressNet {
	return this.AddrFrom
}
func (this *DataChainProxyBase) GetAddrTo() nodeStore.AddressNet {
	return this.AddrTo
}

/*
构建消息的hash值
*/
func (this *DataChainProxyBase) BuildHash() {
	this.Hash = ulid.Make().Bytes()
	return
}

func (this *DataChainProxyBase) GetHash() []byte {
	return this.Hash
}
func (this *DataChainProxyBase) SetPreHash(preHash []byte) {
	this.PreHash = preHash
}
func (this *DataChainProxyBase) GetPreHash() []byte {
	return this.PreHash
}
func (this *DataChainProxyBase) CheckHash() bool {
	return true
}
func (this *DataChainProxyBase) Forward() bool {
	return false
}

/*
设置消息状态
*/
func (this *DataChainProxyBase) SetStatus(status uint8) {
	this.Status = status
}

/*
加密消息内容
*/
func (this *DataChainProxyBase) EncryptContent(key []byte) utils.ERROR {
	encryptType := 0
	if key != nil || len(key) > 0 {
		encryptType = 1
	}
	//this.Content = make([][]byte, 0, len(this.clientItr))
	//this.EncryptType = make([]uint32, 0, len(this.clientItr))
	//for _, one := range this.clientItr {
	if this.clientItr == nil {
		return utils.NewErrorSuccess()
	}
	plainText, err := this.clientItr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	switch encryptType {
	case 0:
		//明文保存，未加密
	case 1:
		//对内容加密
		cipherText, err := utils.AesCTR_Encrypt(key, nil, *plainText)
		if err != nil {
			utils.Log.Info().Msgf("发送消息 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Msgf("加密内容:%s %d %s %d %s", hex.EncodeToString(key), len(*plainText),
		//	hex.EncodeToString(*plainText), len(cipherText), hex.EncodeToString(cipherText))

		plainText = &cipherText
	}
	this.Content = *plainText
	this.EncryptType = uint32(encryptType)
	//}
	return utils.NewErrorSuccess()
}

/*
解密消息内容
*/
func (this *DataChainProxyBase) DecryptContent(key []byte) utils.ERROR {
	//this.clientItr = make([]DataChainClientItr, 0, len(this.EncryptType))
	//for i, one := range this.EncryptType {
	var plaintext []byte
	var err error
	switch this.EncryptType {
	case 0:
		//明文保存，未加密
		plaintext = this.Content
	case 1:
		plaintext, err = utils.AesCTR_Decrypt(key, nil, this.Content)
		if err != nil {
			utils.Log.Info().Msgf("发送消息 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Msgf("解密内容:%s %d %s %d %s", hex.EncodeToString(key), len(this.Content),
		//	hex.EncodeToString(this.Content), len(plaintext), hex.EncodeToString(plaintext))

	case 2:
	default:
		utils.Log.Error().Msgf("不支持的加密类型")
	}
	clientItr, ERR := ParseDataChainClient(plaintext)
	if !ERR.CheckSuccess() {
		return ERR
	}
	this.clientItr = clientItr
	//}
	return utils.NewErrorSuccess()
}

/*
获取解密消息内容
*/
func (this *DataChainProxyBase) GetClientItr() DataChainClientItr {
	return this.clientItr
}

/*
获取命令
*/
func (this *DataChainProxyBase) GetCmd() uint32 {
	return this.Command
}

/*
获取本交易用作签名的序列化
[上一个交易GetVoutSignSerialize()返回]+[本交易类型]+[本交易输入总数]+[本交易输入index]+
[本交易输出总数]+[vouts序列化]+[锁定区块高度]
@voutBs    *[]byte    上一个交易GetVoutSignSerialize()返回
*/
func (this *DataChainProxyBase) GetSignSerialize() *[]byte {
	//
	//
	//voutBssLenght := 0
	//voutBss := make([]*[]byte, 0, len(this.Vout))
	//for _, one := range this.Vout {
	//	voutBsOne := one.Serialize()
	//	voutBss = append(voutBss, voutBsOne)
	//	voutBssLenght += len(*voutBsOne)
	//}
	//
	//var bs []byte
	//if voutBs == nil {
	//	bs = make([]byte, 0, 8+8+8+8+voutBssLenght+8+len(this.Payload))
	//} else {
	//	bs = make([]byte, 0, len(*voutBs)+8+8+8+8+voutBssLenght+8+len(this.Payload))
	//	bs = append(bs, *voutBs...)
	//}
	//
	//bs = append(bs, utils.Uint64ToBytes(this.Type)...)
	//bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
	//bs = append(bs, utils.Uint64ToBytes(vinIndex)...)
	//bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
	//for _, one := range voutBss {
	//	bs = append(bs, *one...)
	//}
	//bs = append(bs, utils.Uint64ToBytes(this.Gas)...)
	//bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
	//bs = append(bs, this.Payload...)
	//return &bs
	bs := []byte{}
	return &bs
}

/*
获取签名
*/
func (this *DataChainProxyBase) GetSign() *[]byte {
	//signDst := this.GetSignSerialize(nil, vinIndex)
	//
	//// utils.Log.Info().Msgf("签名字符序列化 耗时 %s", config.TimeNow().Sub(start))
	//// fmt.Println("签名前的字节", len(*signDst), hex.EncodeToString(*signDst), "\n")
	//// fmt.Printf("签名前的字节 len=%d signDst=%s key=%s \n", len(*signDst), hex.EncodeToString(*signDst), hex.EncodeToString(*key))
	//sign := keystore.Sign(*key, *signDst)
	//return &sign
	bs := []byte{}
	return &bs
}

func (this *DataChainProxyBase) GetSendTime() int64 {
	return this.SendTime
}
func (this *DataChainProxyBase) GetRecvTime() int64 {
	return this.RecvTime
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDatachainProxyBase(bs []byte) (*DataChainProxyBase, error) {
	basePro := go_protos.ImDataChainProxyBase{}
	err := proto.Unmarshal(bs, &basePro)
	if err != nil {
		return nil, err
	}
	proxyBase := ConvertProxyBase(basePro.Base)
	return proxyBase, nil
}

func ConvertProxyBase(basePro *go_protos.ImProxyBase) *DataChainProxyBase {
	base := DataChainProxyBase{
		ID:        basePro.ID,
		SendID:    basePro.SendID,
		PreHash:   basePro.PreHash,
		Hash:      basePro.Hash,
		Command:   uint32(basePro.Command),
		Index:     new(big.Int).SetBytes(basePro.Index),
		SendIndex: new(big.Int).SetBytes(basePro.SendIndex),
		//SendIndexReset:  basePro.SendIndexReset,
		AddrFrom:        *nodeStore.NewAddressNet(basePro.AddrFrom),
		AddrTo:          *nodeStore.NewAddressNet(basePro.AddrTo),
		AddrProxyServer: *nodeStore.NewAddressNet(basePro.AddrProxyServer),
		GroupID:         basePro.GroupID,
		EncryptType:     uint32(basePro.EncryptType),
		Content:         basePro.Content,
		clientItr:       nil,
		Sign:            basePro.Sign,
		Status:          uint8(basePro.Status),
		DhPuk:           basePro.DhPuk,
		SendTime:        basePro.SendTime,
		RecvTime:        basePro.RecvTime,
	}
	return &base
}

type DataChainProxyBaseVO struct {
	ID              string //每条消息都有一个唯一ID
	PreHash         string //
	Hash            string //
	Command         uint32 //日志指令
	Index           string //日志记录连续的id
	SendIndex       string //聊天记录发送者的发送index
	AddrFrom        string //发送者
	AddrTo          string //发送给谁
	AddrProxyServer string //代理服务器节点
	GroupID         string //群ID，以此判断是群消息还是个人消息，创建者地址和创建时间的Hash
	EncryptType     uint32 //加密类型。0=未加密;1=AES加密;2=;
	//Content         [][]byte             //多条内容
	//clientItr       []DataChainClientItr //加密内容
	Sign   string //签名
	Status uint8  //消息状态。0=保存到本地;1=发送成功;2=;
}

func (this *DataChainProxyBase) ConverVO() *DataChainProxyBaseVO {
	vo := DataChainProxyBaseVO{
		ID:              hex.EncodeToString(this.ID),
		PreHash:         hex.EncodeToString(this.PreHash),
		Hash:            hex.EncodeToString(this.Hash),
		Command:         this.Command,
		Index:           hex.EncodeToString(this.Index.Bytes()),
		SendIndex:       hex.EncodeToString(this.SendIndex.Bytes()),
		AddrFrom:        this.AddrFrom.B58String(),
		AddrTo:          this.AddrTo.B58String(),
		AddrProxyServer: this.AddrProxyServer.B58String(),
		GroupID:         string(base58.Encode(this.GroupID)),
		EncryptType:     this.EncryptType,
		//Content:         nil,
		//clientItr:       nil,
		Sign:   hex.EncodeToString(this.Sign),
		Status: this.Status,
	}
	return &vo
}
