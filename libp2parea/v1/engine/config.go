package engine

import (
	"errors"
	"time"
)

const (
	SendQueueCacheNum        = 100000
	MSGID_heartbeat          = 0               //心跳连接
	heartBeat_timeout        = time.Second * 8 //心跳间隔时间
	heartBeat_interval       = time.Second * 3 //心跳间隔时间
	MaxSendQueueDealCnt      = 20              // 同时处理协程数量
	MaxQueue                 = 100000          //伸缩发送队列最大长度
	AreaNameResult_same      = 1               //相同
	AreaNameResult_different = 2               //不同
	MaxRetryCnt              = 2               //发送最大重试次数
)

const (
	AddressUp   = "up"       //addSession方法添加upSession标识
	AddressDown = "down"     //addSession方法添加downSession标识
	BothMod     = "both"     //AddClientConn方法全规则添加模式
	OnebyoneMod = "onebyone" //AddClientConn方法onebyone规则添加模式
	KadMod      = "kad"      //AddClientConn方法kad规则添加模式
	VnodeMod    = "vnode"    //AddClientConn方法vnode规则添加模式
)

var (
	ERROR_send_cache_full  = errors.New("send cache full error") //发送队列满了
	ERROR_send_cache_close = errors.New("send cache close")      //发送队列关闭
	ERROR_send_timeout     = errors.New("send timeout error")    //发送超时
	Error_different_netid  = errors.New("netid different")       //
	Error_node_unwanted    = errors.New("node unwante")          //不需要的节点
	SendTimeOut            = time.Second * 10                    //默认超时时间
)

const (
	CONN_TYPE_TCP  = 0 // tcp连接
	CONN_TYPE_QUIC = 1 // quic连接
)
