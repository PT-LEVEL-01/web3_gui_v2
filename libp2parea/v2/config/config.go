package config

import (
	"github.com/gogo/protobuf/proto"
	"runtime"
	"sync/atomic"
	"time"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
)

// 相关限制信息
const (
	MaxMachineIDLen = 128 // 最大machineID长度

	Version_0 = 0 //
	Version_1 = 1 //此版本广播消息机制发生改变，只广播消息hash，节点自己去同步消息到本地

	Path_configDir = "conf"               //配置文件存放目录
	Core_keystore  = "keystore.key"       //密钥文件
	SQLITE3DB_name = "p2pmessagecache.db" //sqlite3数据库文件名称

	Mining_block_time = 10
	Addr_byte_length  = 32

	Session_attribute_key_auth        = "1" //权限
	Session_attribute_key_create_time = "2" //创建连接时间
	Session_attribute_key_nodeinfo    = "3" //节点信息
	Session_attribute_key_ma          = "4" //地址信息

	//节点连上1分钟内未注册，则关闭连接
	Not_reg_timeout = time.Minute

	//组播地址
	MulticastAddress                = "224.0.0.1"     //
	MulticastAddress_cache_timetime = time.Minute / 2 //组播接收的地址缓存时间

	RPC_method_netinfo = "netinfo"
)

var (

	// 网络地址大于本节点地址的最大连接数
	GreaterThanSelfMaxConn int = 20
	OnlyConnectList        []string

	Entry       = []string{}
	CPUNUM      = runtime.NumCPU()
	SyncNum_min = 6 //异步最少数量

	//	LnrTCP   net.Listener //获得并占用一个TCP端口
	IsOnline = &atomic.Bool{} //是否已经连接到网络中了

	//根节点公钥
	Root_publicKeyStr = []byte{}

	//默认的端口区间
	DefaultPortInterval = []PortInterval{{1901, 8}, {2901, 8}, {3901, 8}}
)

func init() {
	IsOnline.Store(false)
	//engine.GlobalInit("console", "", "debug", 1)
	//	engine.GlobalInit("file", `{"filename":"conf/log.txt"}`, "", 1000)

	//	AutoRole()
}

type Config struct {
	IsFirst bool
	Addr    string
	Port    uint16
	pwd     string
}

/*
端口区间
*/
type PortInterval struct {
	PortBase int //端口开始号
	Interval int //连续多少个号码
}

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

/*
使用protobuf把3维数组合并为1个字节数组
*/
func ByteListProto(list [][][]byte) ([]byte, error) {
	lists := go_protobuf.ByteListV2{List: make([]*go_protobuf.ByteListListV2, 0)}
	for _, list := range list {
		listlist := go_protobuf.ByteListListV2{List: list}
		lists.List = append(lists.List, &listlist)
	}
	return proto.Marshal(&lists)
}

/*
使用protobuf把3维数组合并为1个字节数组
*/
func ParseByteList(bs *[]byte) ([][][]byte, error) {
	bss := new(go_protobuf.ByteListV2)
	err := proto.Unmarshal(*bs, bss)
	if err != nil {
		return nil, err
	}
	bsss := make([][][]byte, 0)
	for _, one := range bss.List {
		bsss = append(bsss, one.List)
	}
	return bsss, nil
}
