package model

import (
	"encoding/hex"
	"strconv"
	"web3_gui/config"
	"web3_gui/keystore/v2/base58"
	"web3_gui/libp2parea/v2/node_store"
)

/*
用户基本信息
*/
type UserInfoVO struct {
	Addr        string       //节点地址
	Nickname    string       //昵称
	RemarksName string       //备注昵称
	HeadNum     uint64       //头像编号
	Status      uint64       //状态
	Time        int64        //UNIX时间戳
	ClassCount  []ClassCount //
	Tray        bool         //
	IsGroup     bool         //是否是群
	GroupId     string       //群ID
	GroupSign   string       //群签名
	AddrAdmin   string       //群管理员地址
	ProxyAddr   string       //代理地址
	Token       string       //令牌
	Nodes       string       //备注
	ShoutUp     bool         //群禁言
}

func ConverUserInfoVO(userinfo *UserInfo) *UserInfoVO {
	user := &UserInfoVO{
		Addr:        userinfo.Addr.B58String(),                 //节点地址
		Nickname:    userinfo.Nickname,                         //昵称
		RemarksName: userinfo.RemarksName,                      //
		HeadNum:     userinfo.HeadNum,                          //头像编号
		Status:      userinfo.Status,                           //状态
		Time:        int64(userinfo.Time),                      //UNIX时间戳
		ClassCount:  userinfo.CircleClass,                      //
		Tray:        userinfo.Tray,                             //
		IsGroup:     userinfo.IsGroup,                          //
		GroupId:     string(base58.Encode(userinfo.GroupId)),   //
		GroupSign:   string(base58.Encode(userinfo.GroupSign)), //
		AddrAdmin:   userinfo.AddrAdmin.B58String(),            //
		Token:       hex.EncodeToString(userinfo.Token),        //
		Nodes:       string(userinfo.Notes),                    //
		ShoutUp:     userinfo.ShoutUp,                          //
	}
	userinfo.Proxy.Range(func(key, value any) bool {
		addr := value.(nodeStore.AddressNet)
		user.ProxyAddr = addr.B58String()
		return false
	})
	return user
}

type UserinfoListVO struct {
	UserList []*UserInfoVO
}

func ConverUserListVO(userList *UserInfoList) *UserinfoListVO {
	listVO := UserinfoListVO{
		UserList: make([]*UserInfoVO, 0),
	}
	for _, one := range userList.UserList {
		listVO.UserList = append(listVO.UserList, ConverUserInfoVO(one))
	}
	return &listVO
}

//type MessageContentVO struct {
//	Type          uint64 //消息类型
//	FromIsSelf    bool   //是否自己发出的
//	From          string //发送者
//	To            string //接收者
//	Content       string //消息内容
//	Time          int64  //时间
//	PullAndPushID uint64 //上传或者下载ID
//}

/*
广播消息
*/
type MessageContentVO struct {
	Subscription   uint64                 //通知类型
	Type           uint64                 //消息类型
	FromIsSelf     bool                   //是否自己发出的
	From           string                 //发送者
	To             string                 //接收者
	Nickname       string                 //发送者昵称
	Content        string                 //消息内容
	Time           int64                  //时间
	PullAndPushID  uint64                 //上传或者下载ID
	Data           map[string]interface{} //
	SendID         string                 //
	RecvID         string                 //
	State          int                    //消息状态1=发送未送达;2=发送失败;3=发送成功;4=已读;
	Index          string                 //数据库中消息索引
	IndexSubOne    string                 //数据库中消息索引-1
	IsGroup        bool                   //
	FileMimeType   string                 //
	FileSendTime   uint64                 //文件发送时间
	FileName       string                 //文件名称
	FileType       uint64                 //文件类型
	FileSize       uint64                 //文件总大小
	FileHash       string                 //文件hash
	FileBlockTotal uint64                 //文件块总数
	FileBlockIndex uint64                 //文件块编号，从0开始，连续增长的整数
	//FileContent    string                 //base64图片编码
	TransProgress int //传送进度百分比
}

// 获取通知类型
func (this *MessageContentVO) GetSubscription() uint64 {
	return this.Subscription
}

// 获取通知key
func (this *MessageContentVO) GetNoticeKey() string {
	key := strconv.Itoa(int(this.Subscription))
	switch this.Subscription {
	case config.SUBSCRIPTION_type_msg:
		//自己发出去的消息不闪烁
		if this.FromIsSelf {
			return ""
		}
		//文件头和尾才闪烁，文件体部传输不闪烁
		if this.FileSize > 0 && this.FileBlockIndex != 0 && this.FileBlockTotal != this.FileBlockIndex+1 {
			return ""
		}
		if this.IsGroup {
			key += this.To
		} else {
			key += this.From
		}
	case config.SUBSCRIPTION_type_addFriend:
	case config.SUBSCRIPTION_type_agreeFriend:
	}
	return key
}

func NewMessageInfo(Subscription, t uint64, fromIsSelf bool, from, to, content string, time int64, pullAndPushID uint64,
	data map[string]interface{}, sendID, recvID string, state int, index []byte) *MessageContentVO {
	return &MessageContentVO{
		Subscription:  Subscription,
		Type:          t,          //
		FromIsSelf:    fromIsSelf, //是否自己发出的
		From:          from,
		To:            to,
		Content:       content,
		Time:          time,
		PullAndPushID: pullAndPushID,
		Data:          data,
		SendID:        sendID,
		RecvID:        recvID,
		State:         state,
		Index:         hex.EncodeToString(index),
	}
}

//func ConverMessageContentVO(msgOne *MessageContent) *MessageContentVO {
//	from := nodeStore.AddressNet(msgOne.From)
//	to := nodeStore.AddressNet(msgOne.To)
//	return &MessageContentVO{
//		Type:          msgOne.Type,            //消息类型
//		FromIsSelf:    msgOne.FromIsSelf,      //是否自己发出的
//		From:          from.B58String(),       //发送者
//		To:            to.B58String(),         //接收者
//		Content:       string(msgOne.Content), //消息内容
//		Time:          msgOne.Time,            //时间
//		PullAndPushID: msgOne.PullAndPushID,   //上传或者下载ID
//
//	}
//}

type MessageContentListVO struct {
	MessageList []*MessageContentVO
}

func ConverMessageContentListVO(msgList *MessageContentList) *MessageContentListVO {
	listVO := MessageContentListVO{
		MessageList: make([]*MessageContentVO, 0),
	}
	for _, one := range msgList.List {
		listVO.MessageList = append(listVO.MessageList, one.ConverVO())
	}
	return &listVO
}
