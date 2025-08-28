package nodeStore

import (
	"bytes"
	"github.com/gogo/protobuf/proto"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
保存节点的id
ip地址
不同协议的端口
*/
type NodeInfo struct {
	Version              uint64                     //版本号
	AreaName             []byte                     //域名称
	IdInfo               IdInfo                     //节点id信息，id字符串以16进制显示
	MachineID            []byte                     //每个节点启动的时候生成一个随机数，用作判断自己连接自己的情况
	lastContactTimestamp time.Time                  //最后检查的时间戳
	RemoteMultiaddr      ma.Multiaddr               //主动连接对方时，需要标明对方的可连接端口
	Port                 uint16                     //自己节点的监听端口
	Lock                 *sync.RWMutex              //
	MultiaddrLAN         map[string]engine.AddrInfo //自己的局域网地址。key:string=ma.Multiaddr;value:engine.AddrInfo=;
	MultiaddrWAN         map[string]engine.AddrInfo //自己的广域网地址。key:string=ma.Multiaddr;value:engine.AddrInfo=;
	Sessions             map[string]engine.Session  //本节点的会话。key:string=会话id;value:engine.Session=;
}

func NewNodeInfo(areaName []byte, idinfo *IdInfo, mID []byte) *NodeInfo {
	nodeInfo := NodeInfo{
		Version:              config.Version_1,
		AreaName:             areaName,
		IdInfo:               *idinfo,
		MachineID:            mID,
		lastContactTimestamp: time.Time{},
		RemoteMultiaddr:      nil,
		Port:                 0,
		Lock:                 new(sync.RWMutex),
		MultiaddrLAN:         make(map[string]engine.AddrInfo),
		MultiaddrWAN:         make(map[string]engine.AddrInfo),
		Sessions:             make(map[string]engine.Session),
	}
	return &nodeInfo
}

func (this *NodeInfo) FlashOnlineTime() {
	this.lastContactTimestamp = time.Now()
}

func (this *NodeInfo) AddSession(ss engine.Session) {
	this.Lock.Lock()
	this.Sessions[utils.Bytes2string(ss.GetId())] = ss
	this.Lock.Unlock()
}

func (this *NodeInfo) AddMultiaddrLAN(addrInfo *engine.AddrInfo) {
	//utils.Log.Info().Str("AddMultiaddrLAN", addrInfo.Multiaddr.String()).Send()
	this.Lock.Lock()
	this.MultiaddrLAN[utils.Bytes2string(addrInfo.Multiaddr.Bytes())] = *addrInfo
	this.Lock.Unlock()
}

func (this *NodeInfo) AddMultiaddrWAN(addrInfo *engine.AddrInfo) {
	//utils.Log.Info().Str("AddMultiaddrWAN", addrInfo.Multiaddr.String()).Send()
	this.Lock.Lock()
	this.MultiaddrWAN[utils.Bytes2string(addrInfo.Multiaddr.Bytes())] = *addrInfo
	this.Lock.Unlock()
}

func (this *NodeInfo) GetMultiaddrLAN() []engine.AddrInfo {
	//utils.Log.Info().Str("GetMultiaddrLAN", "").Send()
	this.Lock.Lock()
	addrInfos := make([]engine.AddrInfo, 0, len(this.MultiaddrLAN))
	for _, addrInfo := range this.MultiaddrLAN {
		addrInfos = append(addrInfos, addrInfo)
	}
	this.Lock.Unlock()
	return addrInfos
}

func (this *NodeInfo) GetMultiaddrWAN() []engine.AddrInfo {
	//utils.Log.Info().Str("GetMultiaddrWAN", "").Send()
	this.Lock.Lock()
	addrInfos := make([]engine.AddrInfo, 0, len(this.MultiaddrWAN))
	for _, addrInfo := range this.MultiaddrWAN {
		addrInfos = append(addrInfos, addrInfo)
	}
	this.Lock.Unlock()
	return addrInfos
}

func (this *NodeInfo) GetSessions() []engine.Session {
	this.Lock.Lock()
	ss := make([]engine.Session, 0, len(this.Sessions))
	for _, v := range this.Sessions {
		ss = append(ss, v)
	}
	this.Lock.Unlock()
	return ss
}

func (this *NodeInfo) RemoveSession(ss engine.Session) {
	this.Lock.Lock()
	delete(this.Sessions, utils.Bytes2string(ss.GetId()))
	this.Lock.Unlock()
}

/*
检查除了参数中传入的session之外，是否还有其他session
*/
func (this *NodeInfo) CheckHaveOtherSessions(ss engine.Session) bool {
	have := false
	this.Lock.RLock()
	for _, one := range this.Sessions {
		if !bytes.Equal(one.GetId(), ss.GetId()) {
			have = true
			break
		}
	}
	this.Lock.RUnlock()
	return have
}

/*
验证节点合法性
@return   true:合法;false:不合法;
*/
func (this *NodeInfo) Validate() (bool, utils.ERROR) {
	ok, ERR := CheckIdInfo(this.IdInfo)
	if ERR.CheckFail() {
		return false, ERR
	}
	if !ok {
		return false, utils.NewErrorSuccess()
	}
	return true, utils.NewErrorSuccess()
}

func (this *NodeInfo) Conver() *go_protobuf.NodeV2 {
	idinfo := this.IdInfo.Conver()
	node := go_protobuf.NodeV2{
		AreaName:     this.AreaName,
		IdInfo:       idinfo,
		MachineID:    this.MachineID,
		Version:      this.Version,
		Port:         uint32(this.Port),
		MultiaddrLAN: make([][]byte, 0, len(this.MultiaddrLAN)),
		MultiaddrWAN: make([][]byte, 0, len(this.MultiaddrWAN)),
	}
	if this.RemoteMultiaddr != nil {
		node.RemoteMultiaddr = this.RemoteMultiaddr.Bytes()
	}
	this.Lock.RLock()
	for _, one := range this.MultiaddrLAN {
		node.MultiaddrLAN = append(node.MultiaddrLAN, one.Multiaddr.Bytes())
	}
	for _, one := range this.MultiaddrWAN {
		node.MultiaddrWAN = append(node.MultiaddrWAN, one.Multiaddr.Bytes())
	}
	this.Lock.RUnlock()
	return &node
}

func (this *NodeInfo) Proto() (*[]byte, error) {
	node := this.Conver()
	bs2, err := node.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs2, nil
}

func ConverNode(nodep *go_protobuf.NodeV2) (*NodeInfo, error) {
	idinfo := ConverIdInfo(nodep.IdInfo)
	node := NodeInfo{
		AreaName:     nodep.AreaName,
		IdInfo:       *idinfo,
		MachineID:    nodep.MachineID,
		Version:      nodep.Version,
		Port:         uint16(nodep.Port),
		Lock:         new(sync.RWMutex),
		MultiaddrLAN: make(map[string]engine.AddrInfo),
		MultiaddrWAN: make(map[string]engine.AddrInfo),
		Sessions:     make(map[string]engine.Session),
	}
	//utils.Log.Info().Hex("RemoteMultiaddr", nodep.RemoteMultiaddr).Send()
	if nodep.RemoteMultiaddr != nil && len(nodep.RemoteMultiaddr) > 0 {
		mAddr, err := ma.NewMultiaddrBytes(nodep.RemoteMultiaddr)
		if err != nil {
			return nil, err
		}
		node.RemoteMultiaddr = mAddr
	}
	for _, one := range nodep.MultiaddrLAN {
		a, err := ma.NewMultiaddrBytes(one)
		if err != nil {
			continue
		}
		addrInfo, ERR := engine.CheckAddr(a)
		if ERR.CheckFail() {
			continue
		}
		node.MultiaddrLAN[utils.Bytes2string(addrInfo.Multiaddr.Bytes())] = *addrInfo
	}
	for _, one := range nodep.MultiaddrWAN {
		a, err := ma.NewMultiaddrBytes(one)
		if err != nil {
			continue
		}
		addrInfo, ERR := engine.CheckAddr(a)
		if ERR.CheckFail() {
			continue
		}
		node.MultiaddrWAN[utils.Bytes2string(addrInfo.Multiaddr.Bytes())] = *addrInfo
	}
	return &node, nil
}

func ParseNodeProto(bs []byte) (*NodeInfo, error) {
	nodep := new(go_protobuf.NodeV2)
	err := proto.Unmarshal(bs, nodep)
	if err != nil {
		return nil, err
	}
	node, err := ConverNode(nodep)
	return node, err
}

type NodeInfoList struct {
	List []NodeInfo
}

func NewNodeInfoList(nodeInfos []NodeInfo) *NodeInfoList {
	nodes := NodeInfoList{List: nodeInfos}
	return &nodes
}

func (this *NodeInfoList) Proto() (*[]byte, error) {
	nodesp := go_protobuf.NodeRepeatedV2{
		Nodes: make([]*go_protobuf.NodeV2, 0, len(this.List)),
	}
	for _, one := range this.List {
		node := one.Conver()
		nodesp.Nodes = append(nodesp.Nodes, node)
	}
	bs2, err := nodesp.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs2, nil
}

func ParseNodesProto(bs *[]byte) ([]NodeInfo, error) {
	nodes := make([]NodeInfo, 0)
	if bs == nil {
		return nodes, nil
	}
	nodesp := new(go_protobuf.NodeRepeatedV2)
	err := proto.Unmarshal(*bs, nodesp)
	if err != nil {
		return nil, err
	}
	for _, nodep := range nodesp.Nodes {
		node, err := ConverNode(nodep)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *node)
	}
	return nodes, nil
}
