package proxyrec

import (
	"time"
)

const (
	MAX_REC_PROXY_CNT    = 3               // 最大存储节点数量
	ProxySyncTime        = 5 * time.Second // 节点同步时间
	ProxyDataInvalieTime = 2               // 节点数据失效次数
	CacheValidTime       = time.Second     // 代理缓存有效时间
)
