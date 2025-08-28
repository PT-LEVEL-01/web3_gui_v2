package superinfo

import (
	"time"
)

// [20020 - 20030) 超级节点信息消息号范围

// 消息号
const (
	MSGID_SUPER_INFO_MULTICAST_MSG = 20020 // 发送节点在线广播消息
)

// 常量定义
const (
	TIME_LOOP_SEND_ONLINE_INFO = time.Second * 3 // 循环发送节点在线消息间隔时间, 单位秒
	TIME_LOOP_CHECK_CACHE      = time.Second * 5 // 循环检查节点在线消息间隔时间, 单位秒
	TIME_CACHE_INVALID         = 10              // 缓存失效时间, 单位秒
)
