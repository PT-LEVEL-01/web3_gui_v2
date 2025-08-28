package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"github.com/oklog/ulid/v2"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_group_members, ParseDataChainGroupMemberFactory)
}

/*
群成员
*/
type DataChainGroupMember struct {
	DataChainProxyBase                           //
	ProxyMajor            nodeStore.AddressNet   //指定一个代理节点作为群数据链构建者
	ShoutUp               bool                   //是否禁言
	Nickname              string                 //群名称
	CreateTime            int64                  //创建时间
	MemberAddrsMerkleRoot []byte                 //好友地址默克尔hash
	MembersAddr           []nodeStore.AddressNet //群成员地址
	MembersTime           []int64                //群成员同意加入群的时间
	MembersSign           [][]byte               //成员同意加入群的签名
	MembersSignPuk        [][]byte               //成员同意加入群的签名用公钥
	MembersShareKey       [][]byte               //每个成员的公钥和管理员的私钥作为交换密钥加密的key
	MembersDHPuk          [][]byte               //群成员公钥
	Status                int                    //群状态。1=正常;2=解散;3=;
}

func (this *DataChainGroupMember) Proto() (*[]byte, error) {
	base := this.DataChainProxyBase.GetProto()
	list := go_protos.ImProxyGroupCreate{
		Base:                  base,
		ProxyMajor:            this.ProxyMajor.GetAddr(),
		ShoutUp:               this.ShoutUp,
		Nickname:              []byte(this.Nickname),
		CreateTime:            this.CreateTime,
		MemberAddrsMerkleRoot: this.MemberAddrsMerkleRoot,
		MembersAddr:           make([][]byte, 0, len(this.MembersAddr)),
		MembersTime:           this.MembersTime,
		MembersSign:           this.MembersSign,
		MembersSignPuk:        this.MembersSignPuk,
		MembersShareKey:       this.MembersShareKey,
		MembersDHPuk:          this.MembersDHPuk,
		Status:                int64(this.Status),
	}
	for _, one := range this.MembersAddr {
		list.MembersAddr = append(list.MembersAddr, one.GetAddr())
	}
	bs, err := list.Marshal()
	if err != nil {
		return nil, err
	}
	//fmt.Println("序列化", bs, err)
	return &bs, err
}

/*
获得被代理地址。代理节点需要验证地址是否属于自己代理的节点
*/
func (this *DataChainGroupMember) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

/*
构建消息的hash值
*/
func (this *DataChainGroupMember) BuildHash() {
	this.Hash = ulid.Make().Bytes()
	return
}

/*
 */
func (this *DataChainGroupMember) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDataChainGroupMemberFactory(bs []byte) (DataChainProxyItr, error) {
	return ParseDataChainGroupMember(bs)
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDataChainGroupMember(bs []byte) (*DataChainGroupMember, error) {
	base := go_protos.ImProxyGroupCreate{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	proxyBase := ConvertProxyBase(base.Base)
	fli := &DataChainGroupMember{
		DataChainProxyBase:    *proxyBase,                                             //
		ProxyMajor:            *nodeStore.NewAddressNet(base.ProxyMajor),              //指定一个代理节点作为群数据链构建者
		ShoutUp:               base.ShoutUp,                                           //是否禁言
		Nickname:              string(base.Nickname),                                  //群名称
		CreateTime:            base.CreateTime,                                        //创建时间
		MemberAddrsMerkleRoot: base.MemberAddrsMerkleRoot,                             //好友地址默克尔hash
		MembersAddr:           make([]nodeStore.AddressNet, 0, len(base.MembersAddr)), //群成员地址
		MembersTime:           base.MembersTime,                                       //
		MembersShareKey:       base.MembersShareKey,                                   //每个成员的公钥和管理员的私钥作为交换密钥加密的key
		MembersSign:           base.MembersSign,                                       //成员同意加入群的签名
		MembersSignPuk:        base.MembersSignPuk,                                    //成员同意加入群的签名
		MembersDHPuk:          base.MembersDHPuk,                                      //群成员公钥
		Status:                int(base.Status),                                       //群状态。1=正常;2=解散;3=;
	}
	//for _, one := range base.Base.EncryptType {
	//	fli.DataChainProxyBase.EncryptType = append(fli.DataChainProxyBase.EncryptType, uint32(one))
	//}
	for _, one := range base.MembersAddr {
		fli.MembersAddr = append(fli.MembersAddr, *nodeStore.NewAddressNet(one))
	}
	return fli, nil
}

func NewDatachainGroupMember(addrSelf nodeStore.AddressNet, proxyAddr nodeStore.AddressNet, nickname string,
	createTime int64, shoutUp bool, groupId []byte, members []model.UserInfo) *DataChainGroupMember {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_group_members, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	proxyBase.GroupID = groupId
	addFriend := DataChainGroupMember{
		DataChainProxyBase: *proxyBase,
		ProxyMajor:         proxyAddr,  //指定一个代理节点作为群数据链构建者
		ShoutUp:            shoutUp,    //是否禁言
		Nickname:           nickname,   //群名称
		CreateTime:         createTime, //创建时间
		//MemberAddrsMerkleRoot []byte                 //好友地址默克尔hash
		MembersAddr:     make([]nodeStore.AddressNet, 0, len(members)), //群成员地址
		MembersTime:     make([]int64, 0, len(members)),                //
		MembersShareKey: make([][]byte, 0, len(members)),               //每个成员的公钥和管理员的私钥作为交换密钥加密的key
		MembersSign:     make([][]byte, 0, len(members)),               //成员同意加入群的签名
		MembersSignPuk:  make([][]byte, 0, len(members)),               //成员同意加入群的签名
		MembersDHPuk:    make([][]byte, 0, len(members)),               //群成员公钥
		Status:          1,                                             //群状态。1=正常;2=解散;3=;
	}
	addrList := make([][]byte, 0, len(members))
	for _, one := range members {
		addFriend.MembersAddr = append(addFriend.MembersAddr, one.Addr)
		addFriend.MembersTime = append(addFriend.MembersTime, one.GroupAcceptTime)
		addFriend.MembersShareKey = append(addFriend.MembersShareKey, one.GroupShareKey)
		addFriend.MembersSign = append(addFriend.MembersSign, one.GroupSign)
		addFriend.MembersSignPuk = append(addFriend.MembersSignPuk, one.GroupSignPuk)
		addFriend.MembersDHPuk = append(addFriend.MembersDHPuk, one.GroupDHPuk)
		addrList = append(addrList, one.Addr.GetAddr())
	}
	addFriend.MemberAddrsMerkleRoot = utils.BuildMerkleRoot(addrList)
	return &addFriend
}

/*
创建一个群
*/
type DataChainGroupMemberVO struct {
	DataChainProxyBaseVO        //
	ShutUp               bool   //是否禁言
	Nickname             string //群名称
	CreateTime           int64  //创建时间
}

func (this *DataChainGroupMember) ConverVO() *DataChainGroupMemberVO {
	vo := DataChainGroupMemberVO{
		DataChainProxyBaseVO: *this.DataChainProxyBase.ConverVO(),
		ShutUp:               this.ShoutUp,    //是否禁言
		Nickname:             this.Nickname,   //群名称
		CreateTime:           this.CreateTime, //创建时间
	}
	return &vo
}
