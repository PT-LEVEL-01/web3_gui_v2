package config

import (
	"errors"
	"runtime"
	"time"

	"web3_gui/libp2parea/v1/engine"
)

type NetModel uint8

const (
	NetType_release NetModel = 0
	NetType_test    NetModel = 1
)

// 网络地址大于本节点地址的最大连接数
var GreaterThanSelfMaxConn int = 20
var OnlyConnectList []string

const (
	AreaNameResult_same      = 1 //相同
	AreaNameResult_different = 2 //不同

	NodeIdResult_offline = 1 //节点不在线，允许连接
	NodeIdResult_online  = 2 //节点已经在线，拒绝连接
	NodeIdResult_self    = 3 //自己连接自己

	VNodeIdResult_offline = 4 //虚拟节点不在线
	VNodeIdResult_online  = 5 //虚拟节点在线
)

// 相关限制信息
const (
	MaxMachineIDLen = 128 // 最大machineID长度
)

type Config struct {
	IsFirst bool
	Addr    string
	Port    uint16
	pwd     string
}

// var NetType = NetType_release //网络类型:正式网络release/测试网络test

// func SetNetType(t NetModel) {
// 	NetType = t
// }

const (
	Version_0 = 0 //
	Version_1 = 1 //此版本广播消息机制发生改变，只广播消息hash，节点自己去同步消息到本地

	Path_configDir = "conf"               //配置文件存放目录
	Core_keystore  = "keystore.key"       //密钥文件
	SQLITE3DB_name = "p2pmessagecache.db" //sqlite3数据库文件名称

	SQL_SHOW = false //是否打印sql语句

	// NetType_release = "release" //正式网络
	// NetType_test    = "test"    //测试网络，测试网络中，局域网中节点可以直接链接。

	Mining_block_time = 10
	Addr_byte_length  = 32

	//---------------- base --------------------------
	//	MSGID_Text = 101 //显示文本消息

	MSGID_search_node              = 108 //搜索一个节点地址是否在线
	MSGID_search_node_recv         = 109 //搜索一个节点地址是否在线_返回
	MSGID_checkNodeOnline          = 110 //检查节点是否在线
	MSGID_checkNodeOnline_recv     = 111 //检查节点是否在线_返回
	MSGID_TextMsg                  = 112 //接收文本消息
	MSGID_getNearSuperIP           = 113 //从邻居节点得到自己的逻辑节点
	MSGID_getNearSuperIP_recv      = 114 //从邻居节点得到自己的逻辑节点_返回
	MSGID_multicast_online_recv    = 122 //接收节点上线广播
	MSGID_ask_close_conn           = 127 //询问关闭连接
	MSGID_ask_close_conn_recv      = 128 //询问关闭连接_返回
	MSGID_TextMsg_recv             = 129 //接收消息返回消息
	MSGID_getNearConnectVnode      = 115 //从邻居节点得到自己连接中的虚拟节点
	MSGID_getNearConnectVnode_recv = 116 //从邻居节点得到自己连接中的虚拟节点_返回
	MSGID_del_vnode                = 117 //通知邻居节点有vnode删除

	//---------------- 可靠传输加密通道协议 --------------------------
	MSGID_SearchAddr                = 130 //搜索一个节点，获取节点地址和身份公钥
	MSGID_SearchAddr_recv           = 131 //搜索一个节点，获取节点地址和身份公钥_返回
	MSGID_security_create_pipe      = 132 //发送密钥消息，对方建立通道
	MSGID_security_create_pipe_recv = 133 //发送密钥消息，对方建立通道_返回
	MSGID_security_pipe_error       = 134 //解密错误

	MSGID_searchID      = 135 //查询磁力节点的id
	MSGID_searchID_recv = 136 //查询磁力节点的id_返回

	MSGID_multicast_return = 211 //收到广播消息回复

	MSGID_multicast_offline_recv       = 212 // 接收广播告知某个节点下线
	MSGID_multicast_vnode_offline_recv = 213 // 接收广播告知某个虚拟节点下线

	MSGID_checkAddrOnline      = 214 // 询问指定的地址是否在线
	MSGID_checkAddrOnline_recv = 215 // 询问指定的地址是否在线_返回

	MSGID_checkVnodeAddrOnline      = 216 // 询问指定的虚拟地址是否在线
	MSGID_checkVnodeAddrOnline_recv = 217 // 询问指定的虚拟地址是否在线_返回

	MSGID_multicast_send_vnode_recv = 218 // 接收广播获取虚拟节点信息

	MsGID_recv_router_err = 219 // 接收广播获取虚拟节点信息

	//---------------- P2p代理 模块 --------------------------
	MSGID_sync_proxy             = 300 // 同步代理信息
	MSGID_sync_proxy_recv        = 301 // 同步代理信息_返回
	MSGID_search_addr_proxy      = 302 // 查询地址对应的代理信息
	MSGID_search_addr_proxy_recv = 303 // 查询地址对应的代理信息_返回

	//---------------- Vnode 模块 --------------------------
	MSGID_vnode_getstate            = 600 //查询一个节点是否开通了虚拟节点服务
	MSGID_vnode_getstate_recv       = 601 //查询一个节点是否开通了虚拟节点服务_返回
	MSGID_vnode_getNearSuperIP      = 602 //从邻居节点得到自己的逻辑节点
	MSGID_vnode_getNearSuperIP_recv = 603 //从邻居节点得到自己的逻辑节点_返回
	MSGID_vnode_searchID            = 604 //查询逻辑节点的id
	MSGID_vnode_searchID_recv       = 605 //查询逻辑节点的id_返回

	MSGID_MAX_ID = 700 // 最大消息id号限制，系统只能注册0-700之间的消息号消息，用户可以注册>=700消息号的消息
)

var (
	AddrPre                     = "TEST"
	Wallet_keystore_default_pwd = "123456789"

	VNODE_get_neighbor_vnode_tiker = []time.Duration{time.Second * 1}  //定时获取邻居节点的虚拟节点地址
	VNODE_tiker_sync_logical_vnode = []time.Duration{time.Second * 10} //每个虚拟节点定时从自己的逻辑节点查询逻辑节点
	VNODE_heartbeat_timeout        = time.Second * 60                  //
	Entry                          = []string{}

	CPUNUM      = runtime.NumCPU()
	SyncNum_min = 6 //异步最少数量

	CLASS_wallet_broadcast_return = "CLASS_wallet_broadcast_return" //广播消息回复
	CLASS_im_security_create_pipe = "CLASS_im_security_create_pipe" //创建加密通道消息
	CLASS_security_searchAddr     = "CLASS_security_searchAddr"     //加密通信搜索节点
	CLASS_engine_multicast_sync   = "CLASS_engine_multicast_sync"   //广播消息同步
	CLASS_get_MachineID           = "CLASS_get_MachineID"           //获取节点的机器id
	CLASS_im_msg_come             = "CLASS_im_msg_come"             //消息到达
	CLASS_router_err              = "CLASS_router_err"              //消息转发时，发给消息接收者失败
)

var (
	ERROR_send_to_sender     = errors.New("need send back to sender")              //发回给sender，消息发送失败
	ERROR_not_in_waitRequest = errors.New("not in waitRequest")                    //waitRequest中没加载出来
	ERROR_wait_msg_timeout   = errors.New("wait message timeout")                  //等待消息返回超时
	ERROR_get_node_conn_fail = errors.New("get node conn fail")                    //获取连接失败
	ERROR_offline            = errors.New("node offline")                          //节点离线
	ERROR_online             = errors.New("node online")                           //节点已经在线
	ERROR_sent_myself        = errors.New("sent to myself")                        //消息发送给自己
	ERROR_no_neighbor        = errors.New("no neighbor")                           //没有可用的邻居节点
	ERROR_no_super           = errors.New("no super")                              //没有可用的超级节点
	ERROR_conn_self          = errors.New("Connect yourself, disconnect yourself") //自己连接自己
	ERROR_conn_exists        = errors.New("This link already exists")              //这个连接已经存在
	ERROR_router_fail        = errors.New("message router fail")                   //消息中转失败
	ERROR_params_fail        = errors.New("params fail")                           //参数错误
	Error_retry_over         = errors.New("retry send error")                      //重试发送失败
	Error_node_connecting    = errors.New("Node is connnecting")                   //和对方节点正在建立连接中,本次连接重复
)

const (
	C_Server_name = "libp2parea" //网络名称
	C_root_name   = "root"       //网络管理员域名

	//服务器角色，只有局域网开发模式才能用
	C_role_client = "client" //客户端模式
	C_role_super  = "super"  //超级节点模式

)

var (
	Init_IsGlobalOnlyAddress = false //本地ip是否是公网全球唯一ip
	// Init_LocalIP                    = ""    //本地ip地址(局域网ip或公网全球唯一ip)
	// Init_LocalPort           uint16 = 9981  //本地监听端口

	Init_IsMapping             = false //是否映射了端口
	Init_GatewayAddress        = ""    //网关地址
	Init_GatewayPort    uint16 = 9981  //网关端口

	Mode_local = true          //是否是局域网开发模式
	Init_role  = C_role_client //服务器角色，当为开发模式时可用

	//	LnrTCP   net.Listener //获得并占用一个TCP端口
	IsOnline      = false //是否已经连接到网络中了
	IsNotInternet = true  //是否没有因特网

	//根节点公钥
	Root_publicKeyStr = []byte{}
)

var (
// SuperNodeIp   string           //超级节点ip地址
// SuperNodePort int              //超级节点端口
//
//	TCPListener *net.TCPListener //本地监听TCP端口
)

func init() {
	engine.GlobalInit("console", "", "debug", 1)
	//	engine.GlobalInit("file", `{"filename":"conf/log.txt"}`, "", 1000)

	// utils.Log.Debug().Msgf("session handle receive, %d, %v", msg.Code(), msg.Content())
	// utils.Log.Debug().Msgf("test debug")
	// utils.Log.Warn().Msgf("test warn")
	// utils.Log.Error().Msgf("test error")

	//	AutoRole()
}

/*
	获得一个TCP监听
*/
//func GetTCPListener(ip string, port int) (*net.TCPListener, error) {
//	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip+":"+strconv.Itoa(int(port)))
//	if err != nil {
//		// Log.Error("这个地址不符合规范：%s", ip+":"+strconv.Itoa(int(port)))
//		return nil, err
//	}
//	var listener *net.TCPListener
//	listener, err = net.ListenTCP("tcp4", tcpAddr)
//	if err != nil {
//		// Log.Error("监听一个地址失败：%s", ip+":"+strconv.Itoa(int(port)))
//		// Log.Error("%v", err)
//		return nil, err
//	}
//	// Log.Debug("监听一个地址：%s", ip+":"+strconv.Itoa(int(port)))
//	// fmt.Println("监听一个地址：", ip+":"+strconv.Itoa(int(port)))
//	// fmt.Println(ip + ":" + strconv.Itoa(int(port)) + "成功启动服务器")
//	return listener, nil
//}

/*
获得本机是否是超级节点
*/
func CheckIsSuperPeer() bool {
	if Mode_local {
		switch Init_role {
		case C_role_client:
			return false
		case C_role_super:
			return true
		}
		return false
	}
	if Init_IsGlobalOnlyAddress {
		return true
	}
	if Init_IsMapping {
		return true
	}
	return false
}

/*
	获得自己的节点地址
	@return   string    ip地址
	@return   int       端口号
*/
// func GetHost() (string, uint16) {
// 	//局域网开发模式
// 	if Mode_local {
// 		return Init_LocalIP, Init_LocalPort
// 	}
// 	if Init_IsGlobalOnlyAddress {
// 		return Init_LocalIP, Init_LocalPort
// 	}
// 	if Init_IsMapping {
// 		return Init_GatewayAddress, Init_GatewayPort
// 	}
// 	return "", 0
// }

/*
根据CPU数量
*/
func GetCPUSyncNum() int {
	num := SyncNum_min
	if CPUNUM*2 > num {
		num = CPUNUM * 2
	}
	return num
}

// p2p连接方式
const (
	CONN_TYPE_TCP  = 1 << 0 // tcp连接
	CONN_TYPE_QUIC = 1 << 1 // UDP
	CONN_TYPE_ALL  = 0xFF   // 全部连接
)
