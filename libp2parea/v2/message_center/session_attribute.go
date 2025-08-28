package message_center

import (
	ma "github.com/multiformats/go-multiaddr"
	"net"
	"strconv"
	"strings"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
把节点信息放入session
*/
func SetNodeInfo(s engine.Session, nodeInfo *nodeStore.NodeInfo) {
	//utils.Log.Info().Hex("会话id放入", s.GetId()).Interface("节点信息", nodeInfo).Send()
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
从session中获取节点信息
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
从session中获取节点信息
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
func SetSessionAddrMultiaddr(s engine.Session, addr ma.Multiaddr) {
	//utils.Log.Info().Hex("会话id放入", s.GetId()).Interface("节点信息", nodeInfo).Send()
	s.Set(config.Session_attribute_key_ma, addr)
}

/*
从session中获取地址信息
*/
func GetSessionAddrMultiaddr(s engine.Session) ma.Multiaddr {
	itr := s.Get(config.Session_attribute_key_ma)
	if itr == nil {
		return nil
	}
	addrM, ok := itr.(ma.Multiaddr)
	if !ok {
		//utils.Log.Error().Str("类型转换错误", reflect.TypeOf(nodeInfo).Name())
		return nil
	}
	//utils.Log.Info().Hex("会话id取出", s.GetId()).Interface("节点信息", nodeInfo).Send()
	return addrM
}
