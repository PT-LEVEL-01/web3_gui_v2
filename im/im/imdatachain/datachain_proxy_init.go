package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"github.com/oklog/ulid/v2"
	"math/big"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_init, ParseFirendListInitFactory)
}

/*
初始化好友列表
*/
type FirendListInit struct {
	DataChainProxyBase                            //
	Version               uint64                  //版本号
	PreVersion            uint64                  //前置版本号
	FirendAddrsMerkleRoot []byte                  //好友地址默克尔hash
	FirendAddrs           []*nodeStore.AddressNet //好友地址
	UserList              []*model.UserInfo       //好友信息
	GroupIDsMerkleRoot    []byte                  //群id默克尔hash
	GroupIDs              [][]byte                //群id
	GroupVersions         []uint64                //群版本号
	GroupMembers          []GroupMember           //群成员信息
	Sign                  []byte                  //签名
}

func NewFirendListInit(addrSelf, addrFriend nodeStore.AddressNet) *FirendListInit {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_init, addrSelf, addrFriend, nodeStore.AddressNet{})
	f := FirendListInit{DataChainProxyBase: *proxyBase}
	return &f
}

func (this *FirendListInit) Proto() (*[]byte, error) {
	base := this.DataChainProxyBase.GetProto()
	list := go_protos.FirendListInit{
		Base:                  base,
		Version:               0,
		PreVersion:            0,
		FirendAddrsMerkleRoot: nil,
		FirendAddrs:           nil,
		UserList:              nil,
		GroupIDsMerkleRoot:    nil,
		GroupIDs:              nil,
		GroupVersions:         nil,
		GroupMembers:          nil,
		Sign:                  nil,
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
func (this *FirendListInit) GetProxyClientAddr() nodeStore.AddressNet {
	return this.AddrFrom
}

/*
构建消息的hash值
*/
func (this *FirendListInit) BuildHash() {
	this.Hash = ulid.Make().Bytes()
	return
}

/*
 */
func (this *FirendListInit) GetBase() *DataChainProxyBase {
	return &this.DataChainProxyBase
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseFirendListInitFactory(bs []byte) (DataChainProxyItr, error) {
	base := go_protos.FirendListInit{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, err
	}
	fli := &FirendListInit{
		DataChainProxyBase: DataChainProxyBase{
			ID:              base.Base.ID,
			PreHash:         base.Base.PreHash,
			Hash:            base.Base.Hash,
			Command:         uint32(base.Base.Command),
			Index:           new(big.Int).SetBytes(base.Base.Index),
			AddrFrom:        *nodeStore.NewAddressNet(base.Base.AddrFrom),
			AddrTo:          *nodeStore.NewAddressNet(base.Base.AddrTo),
			AddrProxyServer: *nodeStore.NewAddressNet(base.Base.AddrProxyServer),
			EncryptType:     uint32(base.Base.EncryptType),
			Content:         base.Base.Content,
			//clientItr:       base.Base.,
			Sign:   base.Base.Sign,
			Status: uint8(base.Base.Status),
		}, //
		Version:               base.Version,               //版本号
		PreVersion:            base.PreVersion,            //前置版本号
		FirendAddrsMerkleRoot: base.FirendAddrsMerkleRoot, //好友地址默克尔hash
		//FirendAddrs:           nodeStore.AddressNet(base.FirendAddrs), //好友地址
		//UserList:              base.UserList,                          //好友信息
		GroupIDsMerkleRoot: base.GroupIDsMerkleRoot, //群id默克尔hash
		GroupIDs:           base.GroupIDs,           //群id
		GroupVersions:      base.GroupVersions,      //群版本号
		//GroupMembers:       base.GroupMembers,       //群成员信息
		Sign: base.Sign, //签名
	}
	//for _, one := range base.Base.EncryptType {
	//	fli.DataChainProxyBase.EncryptType = append(fli.DataChainProxyBase.EncryptType, uint32(one))
	//}
	for _, one := range base.GroupMembers {
		gm := GroupMember{
			GroupID:               one.GroupID,
			Version:               one.Version,
			PreVersion:            one.PreVersion,
			AdminAddr:             *nodeStore.NewAddressNet(one.AdminAddr),
			MemberAddrsMerkleRoot: one.MemberAddrsMerkleRoot,
			//MemberAddrs:           nodeStore.AddressNet(one.MemberAddrs),
			GroupIDsMerkleRoot: one.GroupIDsMerkleRoot,
			//UserList:           one.UserList,
			Sign:   one.Sign,
			Status: int(one.Status),
		}
		for _, addrOne := range one.MemberAddrs {
			gm.MemberAddrs = append(gm.MemberAddrs, *nodeStore.NewAddressNet(addrOne))
		}
		//for _, uOne := range one.UserList {
		//	model.UserInfo{
		//		Addr:        nil,
		//		Nickname:    "",
		//		RemarksName: "",
		//		HeadNum:     0,
		//		Status:      0,
		//		Time:        0,
		//		CircleClass: nil,
		//		Tray:        false,
		//		Proxy:       nil,
		//	}
		//	gm.UserList = append(gm.UserList, uOne)
		//}
		fli.GroupMembers = append(fli.GroupMembers, gm)
	}
	return fli, nil
}

/*
群成员
*/
type GroupMember struct {
	GroupID               []byte                 //群id
	Version               uint64                 //版本号
	PreVersion            uint64                 //前置版本号
	AdminAddr             nodeStore.AddressNet   //管理员地址
	MemberAddrsMerkleRoot []byte                 //好友地址默克尔hash
	MemberAddrs           []nodeStore.AddressNet //成员地址
	GroupIDsMerkleRoot    []byte                 //群id默克尔hash
	UserList              []*model.UserInfo      //成员信息，包括管理员
	Sign                  []byte                 //签名
	Status                int                    //群状态。1=正常;2=普通成员禁言;3=解散;
}
