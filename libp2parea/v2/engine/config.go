package engine

import (
	"github.com/gogo/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/oklog/ulid/v2"
	"net"
	"sync"
	"time"
	"web3_gui/libp2parea/v2/engine/protobuf/go_protobuf"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ShowSendAndRecvLog       = false            //是否显示发送和接收日志
	SendQueueCacheNum        = 100000           //
	MSGID_heartbeat          = 0                //心跳连接
	heartBeat_timeout        = time.Second * 8  //心跳间隔时间
	heartBeat_interval       = time.Second * 3  //心跳间隔时间
	SendTimeOut              = time.Second * 30 //默认超时时间
	MaxSendQueueDealCnt      = 20               // 同时处理协程数量
	MaxQueue                 = 100000           //伸缩发送队列最大长度
	AreaNameResult_same      = 1                //相同
	AreaNameResult_different = 2                //不同
	MaxRetryCnt              = 2                //发送最大重试次数

	HTTP_Method_connect                    = "CONNECT"
	HTTP_URL_websocket                     = "/websocket"
	HTTP_URL_rpc_json                      = "/rpc/json"
	HTTP_URL_rpc_websocket                 = "/rpc/websocket"
	HTTP_URL_Page_rpclistjson              = "/page/rpclistjson"              //rpc接口列表及参数说明页面
	HTTP_URL_Page_errors_desc              = "/page/errors_desc"              //错误编号说明
	HTTP_URL_Page_websocket_client         = "/page/websocketclient"          //websocket客户端页面
	HTTP_URL_Static_JS_vue_v3_5_13         = "/static/js_vue_v3_5_13"         //
	HTTP_URL_Static_JS_element_plus_v2_9_8 = "/static/js_element_plus_v2_9_8" //
	HTTP_URL_Static_CSS_element_plus_index = "/static/css_element_plus_index" //

	RPC_username           = "root"        //rpc请求参数中获取用户名key
	RPC_method             = "method"      //rpc请求参数中获取方法名称key
	RPC_method_constant    = "constant"    //方法名称->返回所有常量
	RPC_method_rpclist     = "rpclist"     //方法名称->显示所有rpc接口列表
	RPC_method_errors_desc = "errors_desc" //方法名称->获取所有错误编号描述

	TPL_KEY_method_rpclist        = "method_rpclist"       //模板中传值的key:
	TPL_KEY_method_errors_desc    = "method_errors_desc"   //错误编号说明
	TPL_KEY_Rpc_json_url          = "rpcJsonUrl"           //模板中传值的key:
	TPL_KEY_Rpc_websocket_url     = "rpcWebsocketUrl"      //
	TPL_KEY_vuejs                 = "vuejs"                //模板中传值的key:
	TPL_KEY_Page_rpclistjson      = "page_rpclistjson"     //
	TPL_KEY_Page_errors_desc      = "page_errors_desc"     //
	TPL_KEY_Page_websocket_client = "page_websocketclient" //

	packetLenSize        = 4                                    //包头占4字节
	package_max_size     = 1024 * 1024                          //移动端4G网络下，超过2M发送不成功。因此设置一个包最大容量1M，大于此容量采用分片续传
	package_head_size    = 1024                                 //包结构和其他字段占用大小
	package_surplus_size = package_max_size - package_head_size //

	Package_protocol_start  = uint8(1) //开始
	Package_protocol_keep   = uint8(2) //续传
	Package_protocol_end    = uint8(3) //结尾
	Package_protocol_single = uint8(4) //单独一个包

	CONN_TYPE_TCP_server  = 1 //tcp服务器连接
	CONN_TYPE_TCP_client  = 2 //tcp客户端连接
	CONN_TYPE_QUIC_server = 3 //quic服务器连接
	CONN_TYPE_QUIC_client = 4 //quic客户端连接
	CONN_TYPE_WS_server   = 5 //websocket服务器
	CONN_TYPE_WS_client   = 6 //websocket客户端

)

var (
	RPC_password = "pass" //rpc请求参数中获取用户密码key

	Wait_major_engine_msg = RegClassNameExistPanic([]byte{1}) //等待消息返回的消息类别

	ERROR_code_timeout_send          = utils.RegErrCodeExistPanic(30001, "发送超时")              //
	ERROR_code_timeout_recv          = utils.RegErrCodeExistPanic(30002, "接收超时")              //
	ERROR_code_packege_over_max_size = utils.RegErrCodeExistPanic(30003, "数据包超过最大值")          //
	ERROR_code_package_protocol_fail = utils.RegErrCodeExistPanic(30004, "协议号错误")             //
	ERROR_code_send_cache_full       = utils.RegErrCodeExistPanic(30005, "发送队列满了")            //
	ERROR_code_send_cache_close      = utils.RegErrCodeExistPanic(30006, "发送队列关闭")            //
	ERROR_code_tcp_listen_runing     = utils.RegErrCodeExistPanic(30007, "TCP服务器在运行状态")       //
	ERROR_code_http_listen_runing    = utils.RegErrCodeExistPanic(30008, "http服务器在运行状态")      //
	ERROR_code_ws_listen_runing      = utils.RegErrCodeExistPanic(30009, "websocket服务器在运行状态") //
	ERROR_code_quic_listen_runing    = utils.RegErrCodeExistPanic(30010, "quic服务器在运行状态")      //
	ERROR_code_tcp_listen_addr_fail  = utils.RegErrCodeExistPanic(30011, "TCP监听地址错误")         //
	ERROR_code_quic_listen_addr_fail = utils.RegErrCodeExistPanic(30012, "QUIC监听地址错误")        //
	ERROR_code_ws_listen_addr_fail   = utils.RegErrCodeExistPanic(30013, "websocket监听地址错误")   //
	ERROR_code_response_type_fail    = utils.RegErrCodeExistPanic(30014, "异步等待返回的数据类型错误")     //
	ERROR_code_unsupported_protocols = utils.RegErrCodeExistPanic(30015, "不支持的协议")            //
	ERROR_code_addr_fail             = utils.RegErrCodeExistPanic(30016, "地址错误")              //

	ERROR_code_rpc_method_repeat            = utils.RegErrCodeExistPanic(30101, "注册rpc接口方法名称重复")        //
	ERROR_code_rpc_param_not_found          = utils.RegErrCodeExistPanic(30102, "rpc接口请求参数未找到")         //
	ERROR_code_rpc_param_type_fail          = utils.RegErrCodeExistPanic(30103, "rpc接口请求参数类型错误")        //
	ERROR_code_rpc_param_length_fail        = utils.RegErrCodeExistPanic(30104, "rpc接口请求参数长度错误")        //
	ERROR_code_rpc_param_value_overstep     = utils.RegErrCodeExistPanic(30105, "rpc接口请求参数的值超过范围")      //
	ERROR_code_rpc_user_fail                = utils.RegErrCodeExistPanic(30106, "rpc接口请求用户名密码错误")       //
	ERROR_code_rpc_method_not_found         = utils.RegErrCodeExistPanic(30107, "未找到的方法")               //
	ERROR_code_rpc_method_is_future         = utils.RegErrCodeExistPanic(30108, "该方法在未来会支持，目前还不支持")     //
	ERROR_code_rpc_method_type_fail         = utils.RegErrCodeExistPanic(30109, "rpc接口方法名称类型错误，应该是字符串") //
	ERROR_code_rpc_user_exist               = utils.RegErrCodeExistPanic(30110, "rpc用户名已经存在")           //
	ERROR_code_rpc_user_not_found           = utils.RegErrCodeExistPanic(30111, "rpc用户名不存在")            //
	ERROR_code_rpc_method_fail              = utils.RegErrCodeExistPanic(30112, "注册的对象不是一个方法")          //
	ERROR_code_rpc_method_param_type_fail   = utils.RegErrCodeExistPanic(30113, "注册的方法参数类型错误")          //
	ERROR_code_rpc_method_param_total_fail  = utils.RegErrCodeExistPanic(30114, "注册的方法参数数量错误")          //
	ERROR_code_rpc_method_result_type_fail  = utils.RegErrCodeExistPanic(30115, "注册的方法返回参数类型错误")        //
	ERROR_code_rpc_method_result_total_fail = utils.RegErrCodeExistPanic(30116, "注册的方法返回参数数量错误")        //
	ERROR_code_rpc_result_param_type_fail   = utils.RegErrCodeExistPanic(30117, "rpc接口返回参数类型错误")        //
)

var msgIdRegisterMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

var RpcUserParams = map[string]string{
	//Tpl_Rpc_username_key:       RPC_username,
	//Tpl_Rpc_password_key:       RPC_password,
	//TPL_KEY_vuejs:                 html.JS_vue_global_min,
	TPL_KEY_method_rpclist:        RPC_method_rpclist,
	TPL_KEY_method_errors_desc:    RPC_method_errors_desc,
	TPL_KEY_Rpc_json_url:          HTTP_URL_rpc_json,
	TPL_KEY_Rpc_websocket_url:     HTTP_URL_rpc_websocket,
	TPL_KEY_Page_rpclistjson:      HTTP_URL_Page_rpclistjson,
	TPL_KEY_Page_errors_desc:      HTTP_URL_Page_errors_desc,
	TPL_KEY_Page_websocket_client: HTTP_URL_Page_websocket_client,
}

/*
注册一个ID，并判断是否有重复
@return    bool    是否成功
*/
func RegisterMsgId(id uint64) bool {
	_, ok := msgIdRegisterMap.LoadOrStore(id, nil)
	if ok {
		utils.Log.Error().Msgf("重复注册的消息ID:%d", id)
	}
	return !ok
}

/*
注册消息编号，有重复就panic
*/
func RegMsgIdExistPanic(id uint64) uint64 {
	if !RegisterMsgId(id) {
		panic("msg id exist")
	}
	return id
}

var waitClassNameMap *sync.Map = new(sync.Map) //key:string=[]byte;value:=;

/*
注册一个ID，并判断是否有重复
@return    bool    是否成功
*/
func RegisterClassName(className []byte) bool {
	_, ok := waitClassNameMap.LoadOrStore(utils.Bytes2string(className), nil)
	if ok {
		utils.Log.Error().Msgf("重复注册的异步等待类名称:%+v", className)
	}
	return !ok
}

/*
注册消息编号，有重复就panic
*/
func RegClassNameExistPanic(className []byte) []byte {
	if !RegisterClassName(className) {
		panic("className id exist")
	}
	return className
}

type ListenConfig struct {
	TcpAddr     *net.TCPAddr //tcp服务器监听地址
	WsAddr      *net.TCPAddr //websocket监听地址
	QuicAddr    *net.UDPAddr //quic协议监听地址
	RpcHttpAddr *net.TCPAddr //rpc http服务器监听地址
	RpcWsAddr   *net.TCPAddr //rpc websocket服务器监听地址
}

type Packet struct {
	MsgID   uint64  //协议编号，0保留做心跳，从1开始为用户定义
	sendID  []byte  //发送者id
	replyID []byte  //回复者id
	Data    []byte  //数据内容
	Session Session //
}

func NewPacket(msgId uint64, data *[]byte) *Packet {
	packet := new(Packet)
	packet.MsgID = msgId
	packet.sendID = ulid.Make().Bytes()
	if data != nil {
		packet.Data = *data
	}
	return packet
}

func (this *Packet) GetSendId() []byte {
	return this.sendID
}

func (this *Packet) Proto() (*[]byte, error) {
	packetProto := go_protobuf.PacketV2{
		MsgID:   this.MsgID,
		SendID:  this.sendID,
		ReplyID: this.replyID,
		Data:    this.Data,
	}
	bs, err := packetProto.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, nil
}

func (this *Packet) Reply(data *[]byte, timeout time.Duration) utils.ERROR {
	return this.Session.Reply(this, data, timeout)
}

func ParsePacket(bs []byte) (*Packet, error) {
	packet := new(Packet)
	packetProto := go_protobuf.PacketV2{}
	err := proto.Unmarshal(bs, &packetProto)
	if err != nil {
		//utils.Log.Error().Msgf("recv packet proto unmarshal error:%s", err.Error())
		return nil, err
	}
	packet.MsgID = packetProto.MsgID
	packet.sendID = packetProto.SendID
	packet.replyID = packetProto.ReplyID
	packet.Data = packetProto.Data
	return packet, nil
}

/*
检查地址是否符合规范
*/
func CheckAddr(addr ma.Multiaddr) (*AddrInfo, utils.ERROR) {
	aInfo := new(AddrInfo)
	aInfo.Multiaddr = addr
	aInfo.IsDNS = false
	c, m := ma.SplitFirst(addr)
	//检查协议号
	if c.Protocol().Code == ma.P_IP4 {
		aInfo.NetType = "4"
	} else if c.Protocol().Code == ma.P_IP6 {
		aInfo.NetType = "6"
	} else if c.Protocol().Code == ma.P_DNS {
		aInfo.IsDNS = true
	} else {
		return nil, utils.NewErrorBus(ERROR_code_addr_fail, "1 Not Supported "+c.Protocol().Name)
	}
	aInfo.Addr = c.Value()
	if m == nil {
		return nil, utils.NewErrorBus(ERROR_code_addr_fail, "Addr Incomplete")
	}
	c, m = ma.SplitFirst(m)
	if c.Protocol().Code == ma.P_TCP {
		aInfo.NetType = "tcp" + aInfo.NetType
		aInfo.Port = c.Value()
		if m == nil {
			return aInfo, utils.NewErrorSuccess()
		}
		c, m = ma.SplitFirst(m)
		if c.Protocol().Code != ma.P_HTTP && c.Protocol().Code != ma.P_WS {
			return nil, utils.NewErrorBus(ERROR_code_addr_fail, "2 Not Supported "+c.Protocol().Name)
		}
		aInfo.Proto = c.Protocol().Name
		return aInfo, utils.NewErrorSuccess()
	} else if c.Protocol().Code == ma.P_UDP {
		if m == nil {
			return nil, utils.NewErrorBus(ERROR_code_addr_fail, "Addr Incomplete")
		}
		c, m = ma.SplitFirst(m)
		if c.Protocol().Code != ma.P_QUIC {
			return nil, utils.NewErrorBus(ERROR_code_addr_fail, "3 Not Supported "+c.Protocol().Name)
		}
		aInfo.Proto = c.Protocol().Name
		return aInfo, utils.NewErrorSuccess()
	} else {
		return nil, utils.NewErrorBus(ERROR_code_addr_fail, "4 Not Supported "+c.Protocol().Name)
	}
}

type AddrInfo struct {
	IsDNS     bool         //是否DNS
	NetType   string       //网络类型
	Proto     string       //协议类型
	Addr      string       //地址
	Port      string       //端口
	Multiaddr ma.Multiaddr //
}

/*
是否公网地址
*/
func (this *AddrInfo) IsOnlyIp() bool {
	if this.IsDNS {
		return true
	}
	return utils.IsOnlyIp(this.Addr)
}
