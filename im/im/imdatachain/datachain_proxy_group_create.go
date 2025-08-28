package imdatachain

import (
	"bytes"
	"math/big"
	"time"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_group_create, ParseDataChainCreateGroupFactory)
}

/*
创建一个群
*/
type DataChainCreateGroup struct {
	DataChainGroupMember
}

/*
构建群ID
*/
func (this *DataChainCreateGroup) BuildGroupId() []byte {
	bs := make([]byte, 0)
	bs = append(bs, this.AddrFrom.GetAddr()...)
	createTimeBs := big.NewInt(this.CreateTime).Bytes()
	bs = append(bs, createTimeBs...)
	hashBs := utils.Hash_SHA3_256(bs)
	return hashBs
}

/*
验证群ID
*/
func (this *DataChainCreateGroup) CheckGroupId() bool {
	groupId := this.BuildGroupId()
	if bytes.Equal(groupId, this.GroupID) {
		return true
	}
	return false
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDataChainCreateGroupFactory(bs []byte) (DataChainProxyItr, error) {
	mamberGroup, err := ParseDataChainGroupMember(bs)
	if err != nil {
		return nil, err
	}
	createGroup := DataChainCreateGroup{DataChainGroupMember: *mamberGroup}
	return &createGroup, nil
}

/*
创建群昵称和是否禁言
*/
func NewDataChainCreateGroup(addrSelf nodeStore.AddressNet, proxyAddr nodeStore.AddressNet, nickname string, shoutUp bool,
	shareKey, signPuk, dhPuk []byte) *DataChainCreateGroup {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_group_create, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	createTime := time.Now().Unix()
	group := DataChainGroupMember{
		DataChainProxyBase: *proxyBase,
		ProxyMajor:         proxyAddr,  //指定一个代理节点作为群数据链构建者
		ShoutUp:            shoutUp,    //是否禁言
		Nickname:           nickname,   //群名称
		CreateTime:         createTime, //创建时间
		//MemberAddrsMerkleRoot []byte                 //好友地址默克尔hash
		MembersAddr:     make([]nodeStore.AddressNet, 0, 1), //群成员地址
		MembersTime:     make([]int64, 0, 1),                //
		MembersShareKey: make([][]byte, 0, 1),               //每个成员的公钥和管理员的私钥作为交换密钥加密的key
		MembersSign:     make([][]byte, 0, 1),               //成员同意加入群的签名
		MembersSignPuk:  make([][]byte, 0, 1),               //成员同意加入群的签名用公钥
		MembersDHPuk:    make([][]byte, 0, 1),               //群成员公钥
		Status:          1,                                  //群状态。1=正常;2=解散;3=;
	}

	group.MembersAddr = append(group.MembersAddr, addrSelf)
	group.MembersTime = append(group.MembersTime, createTime)
	group.MembersShareKey = append(group.MembersShareKey, shareKey)
	//group.MembersSign = append(group.MembersSign, sign)
	group.MembersSignPuk = append(group.MembersSignPuk, signPuk)
	group.MembersDHPuk = append(group.MembersDHPuk, dhPuk)

	addrList := make([][]byte, 0, 1)
	addrList = append(addrList, addrSelf.GetAddr())
	group.MemberAddrsMerkleRoot = utils.BuildMerkleRoot(addrList)
	createGroup := DataChainCreateGroup{
		DataChainGroupMember: group,
	}
	createGroup.SendIndex = big.NewInt(1)
	groupId := createGroup.BuildGroupId()
	createGroup.GroupID = groupId
	//utils.Log.Info().Msgf("创建群:%+v", createGroup)
	return &createGroup
}

/*
创建一个群
*/
type DataChainCreateGroupVO struct {
	DataChainProxyBaseVO        //
	ShutUp               bool   //是否禁言
	Nickname             string //群名称
	CreateTime           int64  //创建时间
}

func (this *DataChainCreateGroup) ConverVO() *DataChainCreateGroupVO {
	vo := DataChainCreateGroupVO{
		DataChainProxyBaseVO: *this.DataChainProxyBase.ConverVO(),
		ShutUp:               this.ShoutUp,    //是否禁言
		Nickname:             this.Nickname,   //群名称
		CreateTime:           this.CreateTime, //创建时间
	}
	return &vo
}
