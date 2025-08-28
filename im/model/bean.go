package model

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"math/big"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
用户基本信息
*/
type UserInfo struct {
	Addr            nodeStore.AddressNet //节点地址
	Nickname        string               //昵称
	RemarksName     string               //备注昵称
	HeadNum         uint64               //头像编号
	Status          uint64               //状态
	Time            int64                //UNIX时间戳
	CircleClass     []ClassCount         //好友的类别
	Tray            bool                 //是否打开托盘
	Proxy           *sync.Map            //代理节点
	IsGroup         bool                 //是否是群
	GroupId         []byte               //群ID
	AddrAdmin       nodeStore.AddressNet //群管理员地址
	GroupAcceptTime int64                //同意入群时间
	GroupSign       []byte               //群签名
	GroupSignPuk    []byte               //群签名用的公钥
	GroupShareKey   []byte               //群协商密码，用于群消息加解密
	GroupDHPuk      []byte               //协商密钥用公钥
	Admin           bool                 //是否是管理员
	Token           []byte               //令牌
	Notes           []byte               //备注
	ShoutUp         bool                 //群禁言
}

func NewUserInfo(addr nodeStore.AddressNet) *UserInfo {
	return &UserInfo{
		Addr:  addr,
		Proxy: new(sync.Map),
	}
}

/*
序列化
*/
func (this *UserInfo) Proto() (*[]byte, error) {
	class := make([]*go_protos.ClassCount, 0)
	for _, one := range this.CircleClass {
		classOne := go_protos.ClassCount{
			Name:  []byte(one.Name),
			Size_: one.Count,
		}
		class = append(class, &classOne)
	}
	proxy := make([][]byte, 0)
	if this.Proxy != nil {
		this.Proxy.Range(func(k, v interface{}) bool {
			key := k.(string)
			proxy = append(proxy, []byte(key))
			return true
		})
	}
	bhp := go_protos.UserInfo{
		Addr:            this.Addr.GetAddr(),
		Nickname:        []byte(this.Nickname),
		RemarksName:     []byte(this.RemarksName),
		HeadNum:         this.HeadNum,
		Status:          this.Status, //状态
		Time:            this.Time,   //UNIX时间戳
		ClassList:       class,       //
		Tray:            this.Tray,   //
		Proxy:           proxy,
		IsGroup:         this.IsGroup,
		GroupId:         this.GroupId,
		AddrAdmin:       this.AddrAdmin.GetAddr(),
		GroupAcceptTime: this.GroupAcceptTime,
		GroupSign:       this.GroupSign,
		GroupSignPuk:    this.GroupSignPuk,
		GroupShareKey:   this.GroupShareKey,
		GroupDHPuk:      this.GroupDHPuk,
		Admin:           this.Admin,
		Token:           this.Token,
		Notes:           this.Notes,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析用户基本信息
*/
func ParseUserInfo(bs *[]byte) (*UserInfo, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.UserInfo)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	class := make([]ClassCount, 0)
	for _, one := range bhp.ClassList {
		class = append(class, ClassCount{
			Name:  string(one.Name),
			Count: one.Size_,
		})
	}
	bh := UserInfo{
		Addr:            *nodeStore.NewAddressNet(bhp.Addr), //nodeStore.AddressNet(bhp.Addr),
		Nickname:        string(bhp.Nickname),
		RemarksName:     string(bhp.RemarksName),
		HeadNum:         bhp.HeadNum,
		Status:          bhp.Status, //状态
		Time:            bhp.Time,   //UNIX时间戳
		CircleClass:     class,
		Tray:            bhp.Tray,
		IsGroup:         bhp.IsGroup,
		GroupId:         bhp.GroupId,
		AddrAdmin:       *nodeStore.NewAddressNet(bhp.AddrAdmin), //bhp.AddrAdmin,
		Proxy:           new(sync.Map),
		GroupAcceptTime: bhp.GroupAcceptTime,
		GroupSign:       bhp.GroupSign,
		GroupSignPuk:    bhp.GroupSignPuk,
		GroupShareKey:   bhp.GroupShareKey,
		GroupDHPuk:      bhp.GroupDHPuk,
		Admin:           bhp.Admin,
		Token:           bhp.Token,
		Notes:           bhp.Notes,
	}
	for _, one := range bhp.Proxy {
		bh.Proxy.Store(utils.Bytes2string(one), nodeStore.NewAddressNet(one))
	}
	return &bh, nil
}

/*
用户列表
*/
type UserInfoList struct {
	UserList []*UserInfo
}

func NewUserList() *UserInfoList {
	return &UserInfoList{
		UserList: make([]*UserInfo, 0),
	}
}

/*
序列化
*/
func (this *UserInfoList) Proto() (*[]byte, error) {
	userListProto := go_protos.UserInfoList{
		UserList: make([]*go_protos.UserInfo, 0),
	}
	for i, _ := range this.UserList {
		one := this.UserList[i]
		userInfo := go_protos.UserInfo{
			Addr:            one.Addr.GetAddr(),
			Nickname:        []byte(one.Nickname),
			RemarksName:     []byte(one.RemarksName),
			HeadNum:         one.HeadNum,
			Status:          one.Status, //状态
			Time:            one.Time,   //UNIX时间戳
			Tray:            one.Tray,   //
			Proxy:           make([][]byte, 0),
			IsGroup:         one.IsGroup,
			GroupId:         one.GroupId,
			AddrAdmin:       one.AddrAdmin.GetAddr(),
			GroupAcceptTime: one.GroupAcceptTime,
			GroupSign:       one.GroupSign,
			GroupSignPuk:    one.GroupSignPuk,
			GroupShareKey:   one.GroupShareKey,
			GroupDHPuk:      one.GroupDHPuk,
			Admin:           one.Admin,
			Token:           one.Token,
			Notes:           one.Notes,
		}
		one.Proxy.Range(func(k, v interface{}) bool {
			key := k.(string)
			userInfo.Proxy = append(userInfo.Proxy, []byte(key))
			return true
		})
		userListProto.UserList = append(userListProto.UserList, &userInfo)
	}
	bs, err := userListProto.Marshal()
	return &bs, err
}

/*
解析用户基本信息
*/
func ParseUserList(bs *[]byte) (*UserInfoList, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.UserInfoList)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	userList := NewUserList()
	for i, _ := range bhp.UserList {
		one := bhp.UserList[i]
		userInfo := UserInfo{
			Addr:            *nodeStore.NewAddressNet(one.Addr),
			Nickname:        string(one.Nickname),
			RemarksName:     string(one.RemarksName),
			HeadNum:         one.HeadNum,
			Status:          one.Status, //状态
			Time:            one.Time,   //UNIX时间戳
			Tray:            one.Tray,   //
			Proxy:           new(sync.Map),
			IsGroup:         one.IsGroup,
			GroupId:         one.GroupId,
			GroupAcceptTime: one.GroupAcceptTime,
			AddrAdmin:       *nodeStore.NewAddressNet(one.AddrAdmin),
			GroupSign:       one.GroupSign,
			GroupSignPuk:    one.GroupSignPuk,
			GroupShareKey:   one.GroupShareKey,
			GroupDHPuk:      one.GroupDHPuk,
			Admin:           one.Admin,
			Token:           one.Token,
			Notes:           one.Notes,
		}
		for _, one := range one.Proxy {
			userInfo.Proxy.Store(utils.Bytes2string(one), nodeStore.NewAddressNet(one))
		}
		userList.UserList = append(userList.UserList, &userInfo)
	}
	return userList, nil
}

type MessageContent struct {
	Type           uint64               //消息类型
	FromIsSelf     bool                 //是否自己发出的
	From           nodeStore.AddressNet //发送者
	To             nodeStore.AddressNet //接收者
	Nickname       string               //发送者昵称
	Content        []byte               //消息内容
	PullAndPushID  uint64               //上传或者下载ID
	Time           int64                //时间
	SendID         []byte               //消息唯一ID
	RecvID         []byte               //消息唯一ID
	QuoteID        []byte               //引用消息ID
	State          int                  //消息状态。1=未发送;2=已送达;3=已读;
	Index          []byte               //数据库中消息索引
	EncryptType    uint32               //加密类型。0=未加密;1=AES加密;2=;
	IsGroup        bool                 //群消息
	FileMimeType   string               //文件类型
	FileSendTime   int64                //文件发送时间
	FileName       string               //文件名称
	FileType       uint64               //文件类型
	FileSize       uint64               //文件总大小
	FileHash       []byte               //文件hash
	FileBlockTotal uint64               //文件块总数
	FileBlockIndex uint64               //文件块编号，从0开始，连续增长的整数
	FileContent    [][]byte             //保存文件块的数据链index索引
	TransProgress  int                  //传送进度百分比
}

/*
创建一个base64编码的图片文件
*/
func NewMsgContentImgBase64(fromIsSelf bool, from, to nodeStore.AddressNet, createTime, fileSendTime int64, sendId []byte,
	fileName, mimeType string, size uint64, fileHash []byte, blockTotal, blockIndex uint64, content [][]byte) *MessageContent {
	return &MessageContent{
		Type:           config.MSG_type_image_base64,
		FromIsSelf:     fromIsSelf,
		From:           from,
		To:             to,
		Content:        nil,
		PullAndPushID:  0,
		Time:           createTime,
		SendID:         sendId,
		RecvID:         nil,
		QuoteID:        nil,
		State:          0,
		Index:          nil,
		EncryptType:    0,
		IsGroup:        false,
		FileMimeType:   mimeType,
		FileSendTime:   fileSendTime,
		FileName:       fileName,
		FileType:       config.FILE_type_image_base64,
		FileSize:       size,
		FileHash:       fileHash,
		FileBlockTotal: blockTotal,
		FileBlockIndex: blockIndex,
		FileContent:    content,
	}
}

/*
创建一个发送文件消息
*/
func NewMsgContentFile(fromIsSelf bool, from, to nodeStore.AddressNet, createTime, fileSendTime int64, sendId []byte,
	fileName, mimeType string, size uint64, fileHash []byte, blockTotal, blockIndex uint64, content [][]byte) *MessageContent {
	return &MessageContent{
		Type:           config.MSG_type_file,
		FromIsSelf:     fromIsSelf,
		From:           from,
		To:             to,
		Content:        nil,
		PullAndPushID:  0,
		Time:           createTime,
		SendID:         sendId,
		RecvID:         nil,
		QuoteID:        nil,
		State:          0,
		Index:          nil,
		EncryptType:    0,
		IsGroup:        false,
		FileMimeType:   mimeType,
		FileSendTime:   fileSendTime,
		FileName:       fileName,
		FileType:       config.FILE_type_file,
		FileSize:       size,
		FileHash:       fileHash,
		FileBlockTotal: blockTotal,
		FileBlockIndex: blockIndex,
		FileContent:    content,
	}
}

func ConvertMessageContentToProto(this *MessageContent) *go_protos.MessageContent {
	mc := go_protos.MessageContent{
		Type:           this.Type,
		FromIsSelf:     this.FromIsSelf,     //是否自己发出的
		From:           this.From.GetAddr(), //发送者
		To:             this.To.GetAddr(),   //接收者
		Content:        this.Content,
		Time:           uint64(this.Time),
		PullAndPushID:  this.PullAndPushID,
		SendID:         this.SendID,
		RecvID:         this.RecvID,
		QuoteID:        this.QuoteID,
		State:          uint32(this.State),
		Index:          this.Index,
		IsGroup:        this.IsGroup,
		FileMimeType:   []byte(this.FileMimeType),
		FileSendTime:   uint64(this.FileSendTime),
		FileName:       []byte(this.FileName),
		FileType:       uint64(this.FileType),
		FileSize:       this.FileSize,
		FileHash:       []byte(this.FileHash),
		FileBlockTotal: this.FileBlockTotal,
		FileBlockIndex: this.FileBlockIndex,
		FileContent:    this.FileContent,
		TransProgress:  uint64(this.TransProgress), //传送进度百分比
	}
	return &mc
}

func ConvertProtoToMessageContent(this *go_protos.MessageContent) *MessageContent {
	mc := MessageContent{
		Type:           this.Type,
		FromIsSelf:     this.FromIsSelf,                     //是否自己发出的
		From:           *nodeStore.NewAddressNet(this.From), //发送者
		To:             *nodeStore.NewAddressNet(this.To),   //接收者
		Content:        this.Content,
		Time:           int64(this.Time),
		PullAndPushID:  this.PullAndPushID,
		SendID:         this.SendID,
		RecvID:         this.RecvID,
		QuoteID:        this.QuoteID,
		State:          int(this.State),
		Index:          this.Index,
		IsGroup:        this.IsGroup,
		FileMimeType:   string(this.FileMimeType),
		FileSendTime:   int64(this.FileSendTime),
		FileName:       string(this.FileName),
		FileType:       this.FileType,
		FileSize:       this.FileSize,
		FileHash:       this.FileHash,
		FileBlockTotal: this.FileBlockTotal,
		FileBlockIndex: this.FileBlockIndex,
		FileContent:    this.FileContent,
		TransProgress:  int(this.TransProgress), //传送进度百分比
	}
	return &mc
}

/*
序列化
*/
func (this *MessageContent) Proto() (*[]byte, error) {
	bhp := ConvertMessageContentToProto(this)
	bs, err := bhp.Marshal()
	return &bs, err
}

func (this *MessageContent) ConverVO() *MessageContentVO {
	from := nodeStore.AddressNet(this.From)
	to := nodeStore.AddressNet(this.To)
	indexSubOne := new(big.Int).Sub(new(big.Int).SetBytes(this.Index), big.NewInt(1)).Bytes()
	mcVO := &MessageContentVO{
		Type:       this.Type,        //
		FromIsSelf: this.FromIsSelf,  //是否自己发出的
		From:       from.B58String(), //
		To:         to.B58String(),   //
		Nickname:   this.Nickname,
		//Content:        string(this.Content),                              //
		Time:           this.Time,                       //
		PullAndPushID:  this.PullAndPushID,              //
		Data:           make(map[string]interface{}),    //
		SendID:         hex.EncodeToString(this.SendID), //
		RecvID:         hex.EncodeToString(this.RecvID), //
		State:          this.State,                      //
		Index:          hex.EncodeToString(this.Index),  //
		IndexSubOne:    hex.EncodeToString(indexSubOne), //
		IsGroup:        this.IsGroup,                    //
		FileMimeType:   this.FileMimeType,               //
		FileSendTime:   uint64(this.FileSendTime),       //
		FileName:       this.FileName,
		FileType:       uint64(this.FileType),
		FileSize:       this.FileSize,
		FileHash:       hex.EncodeToString(this.FileHash),
		FileBlockTotal: this.FileBlockTotal,
		FileBlockIndex: this.FileBlockIndex,
		//FileContent:    this.FileContent,
		TransProgress: this.TransProgress, //传送进度百分比
	}
	//utils.Log.Info().Msgf("是否图片:%d %d %d", mcVO.FileType, mcVO.FileBlockTotal, mcVO.FileBlockIndex)
	if mcVO.Type == config.MSG_type_text {
		mcVO.Content = string(this.Content)
	} else if mcVO.Type == config.MSG_type_voice {
		mcVO.Content = string(this.Content)
	} else if mcVO.FileType == config.FILE_type_image_base64 && mcVO.FileBlockTotal == mcVO.FileBlockIndex+1 {
		//是base64编码图片，并且已经传完
		//mcVO.Content = base64.StdEncoding.EncodeToString(this.Content)
		mcVO.Content = string(this.Content)
	} else {

	}
	return mcVO
}

/*
解析好友消息内容
*/
func ParseMessageContent(bs *[]byte) (*MessageContent, error) {
	if bs == nil {
		return nil, nil
	}
	mc := new(go_protos.MessageContent)
	err := proto.Unmarshal(*bs, mc)
	if err != nil {
		return nil, err
	}
	bh := ConvertProtoToMessageContent(mc)
	return bh, nil
}

type MessageContentList struct {
	List []*MessageContent
}

func NewMessageContentList() *MessageContentList {
	return &MessageContentList{
		List: make([]*MessageContent, 0),
	}
}

/*
序列化
*/
func (this *MessageContentList) Proto() (*[]byte, error) {
	userListProto := go_protos.MessageContentList{
		List: make([]*go_protos.MessageContent, 0),
	}
	for i, _ := range this.List {
		mc := ConvertMessageContentToProto(this.List[i])
		userListProto.List = append(userListProto.List, mc)
	}
	bs, err := userListProto.Marshal()
	return &bs, err
}

/*
解析好友消息内容
*/
func ParseMessageContentList(bs *[]byte) (*MessageContentList, error) {
	if bs == nil {
		return nil, nil
	}
	mcl := new(go_protos.MessageContentList)
	err := proto.Unmarshal(*bs, mcl)
	if err != nil {
		return nil, err
	}
	mcList := MessageContentList{
		List: make([]*MessageContent, 0),
	}
	for i, _ := range mcl.List {
		one := mcl.List[i]
		mcOne := ConvertProtoToMessageContent(one)
		mcList.List = append(mcList.List, mcOne)
	}
	return &mcList, nil

}

type FileList struct {
	List      []FileVO
	ErrorCode uint64 //错误编码
}

type FileVO struct {
	Name       string //文件名称
	Hash       string //文件hash
	IsDir      bool   //是否是文件夹
	UpdateTime string //最后修改时间
	Size       uint64 //文件大小
	Price      uint64 //价格
}

/*
解析文件列表
*/
func ParseFilelist(bs *[]byte) (*FileList, error) {
	if bs == nil {
		return nil, nil
	}
	fl := new(go_protos.Filelist)
	err := proto.Unmarshal(*bs, fl)
	if err != nil {
		return nil, err
	}
	files := make([]FileVO, 0)
	for _, one := range fl.FileList {
		upodateTime := time.Unix(one.UpdateTime, 0).Format("2006-01-02 15:04:05")
		file := FileVO{
			Name:       string(one.Name),             //文件名称
			Hash:       hex.EncodeToString(one.Hash), //
			IsDir:      one.IsDir,                    //是否是文件夹
			UpdateTime: upodateTime,                  //最后修改时间
			Size:       one.Length,                   //文件大小
			Price:      one.Price,                    //
		}
		files = append(files, file)
	}
	fileList := FileList{
		List:      files,
		ErrorCode: fl.ErrorCode,
	}
	return &fileList, nil
}
