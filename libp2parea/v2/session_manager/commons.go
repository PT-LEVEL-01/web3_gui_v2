package session_manager

import (
	"bytes"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
连接状态
*/
type ConnStatus struct {
	lock    *sync.RWMutex             //
	connMap map[string]*ConnStatusOne //
	Log     **zerolog.Logger          //日志
}

func NewConnStatus(log **zerolog.Logger) *ConnStatus {
	cs := ConnStatus{
		lock:    new(sync.RWMutex),
		connMap: make(map[string]*ConnStatusOne), //new(sync.Map),
		Log:     log,
	}
	return &cs
}

/*
查询连接
如果正在连接，则等待连接结果
如果没有连接，则直接返回
如果有连接，返回连接是否成功
@return    bool           是否存在
@return    utils.ERROR    错误
*/
func (this *ConnStatus) QueryConn(multiaddr ma.Multiaddr) bool {
	this.lock.RLock()
	_, ok := this.connMap[multiaddr.String()]
	//itr, ok := this.connMap.Load(multiaddr.String())
	this.lock.RUnlock()
	if ok {
		//one := itr.(*ConnStatusOne)
		return true
	}
	return false
}

/*
查询连接，如果不存在则保存
@return    bool           是否存在
@return    utils.ERROR    错误
*/
func (this *ConnStatus) LoadOrStore(multiaddr ma.Multiaddr, sid []byte) (bool, utils.ERROR) {
	this.lock.Lock()
	one, ok := this.connMap[multiaddr.String()]
	if !ok {
		one = NewConnStatusOne(multiaddr, sid, this)
		one.SetStatus(utils.NewErrorSuccess())
		//(*this.Log).Info().Str("添加连接", multiaddr.String()).Send()
		this.connMap[multiaddr.String()] = one
	}
	//itr, ok := this.connMap.LoadOrStore(multiaddr.String(), one)
	this.lock.Unlock()
	if ok {
		//one = itr.(*ConnStatusOne)
		//ERR := one.GetStatus()
		return true, utils.NewErrorSuccess()
	}
	return false, utils.NewErrorSuccess()
}

/*
连接
*/
func (this *ConnStatus) Conn(multiaddr ma.Multiaddr, f func() (engine.Session, utils.ERROR)) utils.ERROR {
	//(*this.Log).Info().Str("ConnStatus 连接是否存在", multiaddr.String()).Send()
	this.lock.Lock()
	one, ok := this.connMap[multiaddr.String()]
	if !ok {
		one = NewConnStatusOne(multiaddr, nil, this)
		this.connMap[multiaddr.String()] = one
	}
	//(*this.Log).Info().Str("ConnStatus 连接是否存在", multiaddr.String()).Bool("ConnStatus 连接是否存在", ok).Send()
	//itr, ok := this.connMap.LoadOrStore(multiaddr.String(), one)
	this.lock.Unlock()
	if ok {
		//one = itr.(*ConnStatusOne)
		return one.GetStatus()
	}
	ss, ERR := f()
	if ERR.CheckFail() {
		//连不上，则销毁和删除
		one.Destroy(ERR)
		return ERR
	}
	one.sid = ss.GetId()
	//(*this.Log).Info().Str("添加连接", multiaddr.String()).Hex("sid", one.sid).Send()
	//(*this.Log).Info().Str("ConnStatus 连接是否存在", multiaddr.String()).Str("连接", ERR.String()).Send()
	one.SetStatus(ERR)
	if ERR.CheckFail() {
		//(*this.Log).Info().Str("ConnStatus 连接是否存在 销毁", multiaddr.String()).Send()
		//连不上，则销毁和删除
		one.Destroy(ERR)
	}
	return ERR
}

/*
关闭
*/
func (this *ConnStatus) Close(multiaddr ma.Multiaddr, sid []byte) {
	//(*this.Log).Info().Str("关闭连接", multiaddr.String()).Send()
	this.lock.RLock()
	one, ok := this.connMap[multiaddr.String()]
	//itr, ok := this.connMap.Load(multiaddr.String())
	this.lock.RUnlock()
	//itr, ok := this.connMap.Load(multiaddr.String())
	//(*this.Log).Info().Str("关闭连接", multiaddr.String()).Send()
	if ok {
		//(*this.Log).Info().Str("关闭连接", multiaddr.String()).Hex("sid1", sid).Hex("sid2", one.sid).Send()
		if !bytes.Equal(sid, one.sid) {
			return
		}
		//(*this.Log).Info().Str("关闭连接", multiaddr.String()).Send()
		//one := itr.(*ConnStatusOne)
		one.Destroy(utils.NewErrorBus(config.ERROR_code_session_close, ""))
	}
	//(*this.Log).Info().Str("关闭连接", multiaddr.String()).Send()
}

type ConnStatusOne struct {
	//lock *sync.Mutex  //
	sid        []byte                      //
	multiaddr  ma.Multiaddr                //
	c          chan bool                   //
	ERR        atomic.Pointer[utils.ERROR] //
	connStatus *ConnStatus                 //
}

func NewConnStatusOne(multiaddr ma.Multiaddr, sid []byte, connStatus *ConnStatus) *ConnStatusOne {
	one := ConnStatusOne{
		//lock: new(sync.Mutex),
		sid:        sid,
		multiaddr:  multiaddr,
		c:          make(chan bool, 1),
		connStatus: connStatus,
	}
	return &one
}

/*
获取状态
*/
func (this *ConnStatusOne) GetStatus() utils.ERROR {
	isClose, open := <-this.c
	if open && !isClose {
		select {
		case this.c <- false:
		default:
		}
	}
	return *this.ERR.Load()
}

/*
设置状态
*/
func (this *ConnStatusOne) SetStatus(ERR utils.ERROR) {
	this.ERR.Store(&ERR)
	//this.ERR = &ERR
	this.c <- false
}

/*
销毁
*/
func (this *ConnStatusOne) Destroy(ERR utils.ERROR) {
	this.ERR.Store(&ERR)
	select {
	case <-this.c:
	default:
	}
	this.c <- true
	this.connStatus.lock.Lock()
	defer this.connStatus.lock.Unlock()
	//(*this.connStatus.Log).Info().Str("删除连接", this.multiaddr.String()).Send()
	delete(this.connStatus.connMap, this.multiaddr.String())
}

/*
把节点信息放入session
*/
func SetNodeInfo(s engine.Session, nodeInfo *nodeStore.NodeInfo) {
	//utils.Log.Info().Hex("会话id放入", s.GetId()).Interface("节点信息", nodeInfo).Send()
	nodeInfo.AddSession(s)
	s.Set(config.Session_attribute_key_nodeinfo, nodeInfo)
}

/*
从session中获取节点信息
*/
func GetNodeInfo(s engine.Session) *nodeStore.NodeInfo {
	itr := s.Get(config.Session_attribute_key_nodeinfo)
	if itr == nil {
		return nil
	}
	nodeInfo, ok := itr.(*nodeStore.NodeInfo)
	if !ok {
		//utils.Log.Error().Str("类型转换错误", reflect.TypeOf(nodeInfo).Name())
		return nil
	}
	//utils.Log.Info().Hex("会话id取出", s.GetId()).Interface("节点信息", nodeInfo).Send()
	return nodeInfo
}

/*
拼接一个节点的端口信息
*/
func BuildMultiaddrBySession(s engine.Session) (ma.Multiaddr, utils.ERROR) {
	/*
		/ip4/127.0.0.1/tcp/1234
		/ip4/127.0.0.1/udp/1234/quic
		/ip4/127.0.0.1/tcp/1234/ws
	*/
	mAddr := ""
	//IP版本
	ip := strings.SplitN(s.GetRemoteHost(), ":", 2)[0]
	parsedIP := net.ParseIP(ip)
	if parsedIP.To4() != nil {
		mAddr += "/ip4"
	} else if parsedIP.To16() != nil {
		mAddr += "/ip6"
	}
	//IP地址
	mAddr += "/" + ip
	//协议
	suffix := ""
	if s.GetConnType() == engine.CONN_TYPE_TCP_server || s.GetConnType() == engine.CONN_TYPE_TCP_client {
		mAddr += "/tcp"
	}
	if s.GetConnType() == engine.CONN_TYPE_QUIC_server || s.GetConnType() == engine.CONN_TYPE_QUIC_client {
		mAddr += "/udp"
		suffix = "/quic"
	}
	if s.GetConnType() == engine.CONN_TYPE_WS_server || s.GetConnType() == engine.CONN_TYPE_WS_client {
		mAddr += "/tcp"
		suffix = "/ws"
	}
	//端口
	nodeRemote := GetNodeInfo(s)
	mAddr += "/" + strconv.Itoa(int(nodeRemote.Port))
	//添加后缀
	mAddr += suffix
	a, err := ma.NewMultiaddr(mAddr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return a, utils.NewErrorSuccess()
}

/*
把会话创建时间放入session
*/
func SetSessionCreateTime(ss engine.Session, createTime int64) {
	ss.Set(config.Session_attribute_key_create_time, createTime)
}

/*
获取会话创建时间
*/
func GetSessionCreateTime(ss engine.Session) int64 {
	itr := ss.Get(config.Session_attribute_key_create_time)
	if itr == nil {
		return 0
	}
	createTime, ok := itr.(int64)
	if !ok {
		//utils.Log.Error().Str("类型转换错误", reflect.TypeOf(nodeInfo).Name())
		return 0
	}
	return createTime
}

/*
把地址信息放入session
*/
func SetSessionAddrMultiaddr(s engine.Session, addrInfo *engine.AddrInfo) {
	//utils.Log.Info().Hex("会话id放入", s.GetId()).Interface("节点信息", nodeInfo).Send()
	s.Set(config.Session_attribute_key_ma, addrInfo)
}

/*
从session中获取地址信息
*/
func GetSessionAddrMultiaddr(s engine.Session) *engine.AddrInfo {
	itr := s.Get(config.Session_attribute_key_ma)
	if itr == nil {
		return nil
	}
	addrInfo, ok := itr.(*engine.AddrInfo)
	if !ok {
		//utils.Log.Error().Str("类型转换错误", reflect.TypeOf(nodeInfo).Name())
		return nil
	}
	//utils.Log.Info().Hex("会话id取出", s.GetId()).Interface("节点信息", nodeInfo).Send()
	return addrInfo
}

/*
设置可连接节点的公开端口
*/
func SetSessionCanConnPort(s engine.Session, addrInfo *engine.AddrInfo) {
	//utils.Log.Info().Hex("会话id放入", s.GetId()).Interface("节点信息", nodeInfo).Send()
	s.Set(config.Session_attribute_key_ma, addrInfo)
}

/*
获取可连接节点的公开端口
*/
func GetSessionCanConnPort(s engine.Session) *engine.AddrInfo {
	itr := s.Get(config.Session_attribute_key_ma)
	if itr == nil {
		return nil
	}
	addrInfo, ok := itr.(*engine.AddrInfo)
	if !ok {
		//utils.Log.Error().Str("类型转换错误", reflect.TypeOf(nodeInfo).Name())
		return nil
	}
	//utils.Log.Info().Hex("会话id取出", s.GetId()).Interface("节点信息", nodeInfo).Send()
	return addrInfo
}
