package nodeStore

import (
	"bytes"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/ed25519"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/keystore/v1/crypto/dh"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
)

/*
保存节点的id
ip地址
不同协议的端口
*/
type Node struct {
	IdInfo               IdInfo           `json:"idinfo"`     //节点id信息，id字符串以16进制显示
	IsSuper              uint32           `json:"issuper"`    //是不是超级节点，超级节点有外网ip地址，可以为其他节点提供代理服务
	Addr                 string           `json:"addr"`       //外网ip地址
	TcpPort              uint16           `json:"tcpport"`    //TCP端口
	IsApp                bool             `json:"isapp"`      //是不是手机端节点
	MachineID            int64            `json:"mid"`        //每个节点启动的时候生成一个随机数，用作判断多个节点使用同一个key连入网络的情况【已经废弃，不再使用，为了向前兼容，所以保留】
	MachineIDStr         string           `json:"mid2"`       //设备id，由使用者控制，用作判断多个节点使用同一个key连入网络的情况
	Version              uint64           `json:"v"`          //版本号
	lastContactTimestamp time.Time        `json:"-"`          //最后检查的时间戳
	Type                 NodeClass        `json:"-"`          //
	Lock                 *sync.RWMutex    `json:"-"`          //
	Sessions             []engine.Session `json:"-"`          //节点的多个session
	SetGod               bool             `json:"setgod"`     //设置上帝节点
	SetGodTime           int64            `json:"setgodtime"` //设置上帝节点时间
	QuicPort             uint16           `json:"quicport"`   //Quic端口
}

func (this *Node) FlashOnlineTime() {
	this.lastContactTimestamp = time.Now()
}

func (this *Node) AddSession(ss engine.Session) {
	this.Lock.Lock()
	this.Sessions = append(this.Sessions, ss)
	this.Lock.Unlock()
}

func (this *Node) GetSessions() []engine.Session {
	this.Lock.Lock()
	ss := this.Sessions
	this.Lock.Unlock()
	return ss
}

func (this *Node) RemoveSession(ss engine.Session) bool {
	find := false
	newSession := make([]engine.Session, 0)
	this.Lock.Lock()
	for _, one := range this.Sessions {
		if ss.GetIndex() == one.GetIndex() {
			find = true
			continue
		}
		newSession = append(newSession, one)
	}
	this.Sessions = newSession
	this.Lock.Unlock()
	return find
}

/*
检查除了参数中传入的session之外，是否还有其他session
*/
func (this *Node) CheckHaveOtherSessions(ss engine.Session) bool {
	have := false
	this.Lock.RLock()
	for _, one := range this.Sessions {
		if one.GetIndex() != ss.GetIndex() {
			have = true
			break
		}
	}
	this.Lock.RUnlock()
	return have
}

/*
获取这个节点是否是超级节点
*/
func (this *Node) GetIsSuper() bool {
	if atomic.LoadUint32(&this.IsSuper) == 1 {
		return true
	}
	return false
}

/*
获取这个节点是否是超级节点
*/
func (this *Node) SetIsSuper(isSuper bool) {
	if isSuper {
		atomic.StoreUint32(&this.IsSuper, 1)
	} else {
		atomic.StoreUint32(&this.IsSuper, 0)
	}
}

func (this *Node) Proto() ([]byte, error) {
	idinfo := go_protobuf.IdInfo{
		Id:   this.IdInfo.Id,
		EPuk: this.IdInfo.EPuk,
		CPuk: this.IdInfo.CPuk[:],
		V:    this.IdInfo.V,
		Sign: this.IdInfo.Sign,
	}
	node := go_protobuf.Node{
		IdInfo:       &idinfo,
		IsSuper:      this.GetIsSuper(),
		Addr:         this.Addr,
		TcpPort:      uint32(this.TcpPort),
		IsApp:        this.IsApp,
		MachineID:    this.MachineID,
		Version:      this.Version,
		SetGod:       this.SetGod,
		MachineIDStr: this.MachineIDStr,
		SetGodTime:   this.SetGodTime,
		QuicPort:     uint32(this.QuicPort),
	}
	return node.Marshal()
}

func ParseNodeProto(bs []byte) (*Node, error) {
	nodep := new(go_protobuf.Node)
	err := proto.Unmarshal(bs, nodep)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], nodep.IdInfo.CPuk)
	idinfo := IdInfo{
		Id:   nodep.IdInfo.Id,
		EPuk: nodep.IdInfo.EPuk,
		CPuk: cpuk,
		V:    nodep.IdInfo.V,
		Sign: nodep.IdInfo.Sign,
	}
	isSuper := uint32(0)
	if nodep.GetIsSuper() {
		isSuper = 1
	}
	node := Node{
		IdInfo:       idinfo,
		IsSuper:      isSuper,
		Addr:         nodep.Addr,
		TcpPort:      uint16(nodep.TcpPort),
		IsApp:        nodep.IsApp,
		MachineID:    nodep.MachineID,
		Version:      nodep.Version,
		SetGod:       nodep.SetGod,
		MachineIDStr: nodep.MachineIDStr,
		SetGodTime:   nodep.SetGodTime,
		QuicPort:     uint16(nodep.QuicPort),
	}
	return &node, nil
}

func ParseNodesProto(bs *[]byte) ([]Node, error) {
	nodes := make([]Node, 0)
	if bs == nil {
		return nodes, nil
	}
	nodesp := new(go_protobuf.NodeRepeated)
	err := proto.Unmarshal(*bs, nodesp)
	if err != nil {
		return nil, err
	}
	for _, nodep := range nodesp.Nodes {
		var cpuk dh.Key = [32]byte{}
		copy(cpuk[:], nodep.IdInfo.CPuk)
		idinfo := IdInfo{
			Id:   nodep.IdInfo.Id,
			EPuk: nodep.IdInfo.EPuk,
			CPuk: cpuk,
			V:    nodep.IdInfo.V,
			Sign: nodep.IdInfo.Sign,
		}
		isSuper := uint32(0)
		if nodep.GetIsSuper() {
			isSuper = 1
		}
		node := Node{
			IdInfo:       idinfo,
			IsSuper:      isSuper,
			Addr:         nodep.Addr,
			TcpPort:      uint16(nodep.TcpPort),
			IsApp:        nodep.IsApp,
			MachineID:    nodep.MachineID,
			Version:      nodep.Version,
			SetGod:       nodep.SetGod,
			MachineIDStr: nodep.MachineIDStr,
			QuicPort:     uint16(nodep.QuicPort),
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Id信息
type IdInfo struct {
	Id   AddressNet        `json:"id"`   //id，节点网络地址
	EPuk ed25519.PublicKey `json:"epuk"` //ed25519公钥，身份密钥的公钥
	CPuk dh.Key            `json:"cpuk"` //curve25519公钥,DH公钥
	V    uint32            `json:"v"`    //DH公钥版本，低版本将被弃用，用于自动升级更换DH公钥协议
	Sign []byte            `json:"sign"` //ed25519私钥签名,Sign(V + CPuk)
	// Ctype string           `json:"ctype"` //签名方法 如ecdsa256 ecdsa512
}

/*
给idInfo签名
*/
func (this *IdInfo) SignDHPuk(prk ed25519.PrivateKey) {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, this.V)
	buf.Write(this.CPuk[:])
	this.Sign = ed25519.Sign(prk, buf.Bytes())
}

/*
验证签名
*/
func (this *IdInfo) CheckSignDHPuk() bool {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, this.V)
	buf.Write(this.CPuk[:])
	return ed25519.Verify(this.EPuk, buf.Bytes(), this.Sign)
}

func (this *IdInfo) Proto() ([]byte, error) {
	idinfo := go_protobuf.IdInfo{
		Id:   this.Id,
		EPuk: this.EPuk,
		CPuk: this.CPuk[:],
		V:    this.V,
		Sign: this.Sign,
	}
	return idinfo.Marshal()
}

/*
检查idInfo是否合法
1.地址生成合法
2.签名正确
@return   true:合法;false:不合法;
*/
func CheckIdInfo(idInfo IdInfo) bool {
	//验证签名
	ok := idInfo.CheckSignDHPuk()
	if !ok {
		return false
	}
	//验证地址
	return CheckPukAddr(idInfo.EPuk, idInfo.Id)
}

func ParseIdInfo(bs []byte) (*IdInfo, error) {
	iip := new(go_protobuf.IdInfo)
	err := proto.Unmarshal(bs, iip)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], iip.CPuk)
	idInfo := IdInfo{
		Id:   iip.Id,   //id，节点网络地址
		EPuk: iip.EPuk, //ed25519公钥，身份密钥的公钥
		CPuk: cpuk,     //curve25519公钥,DH公钥
		V:    iip.V,    //DH公钥版本，低版本将被弃用，用于自动升级更换DH公钥协议
		Sign: iip.Sign, //ed25519私钥签名,Sign(V + CPuk)
	}
	return &idInfo, nil
}

/*
临时id
*/
type TempId struct {
	SuperPeerId *AddressNet `json:"superpeerid"` //更新在线时间
	PeerId      *AddressNet `json:"peerid"`      //更新在线时间
	UpdateTime  int64       `json:"updatetime"`  //更新在线时间
}

/*
创建一个临时id
*/
func NewTempId(superId, peerId *AddressNet) *TempId {
	return &TempId{
		SuperPeerId: superId,
		PeerId:      peerId,
		UpdateTime:  time.Now().Unix(),
	}
}
