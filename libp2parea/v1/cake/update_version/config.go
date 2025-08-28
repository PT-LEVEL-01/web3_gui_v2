package update_version

import "time"

var (
	Root_version_library                     = "version_library" //root节点版本文件存放根目录
	Version_dir                              = "versions"        //客户端新版本程序保持文件夹
	P2p_mgs_timeout                          = 5 * time.Second   //p2p消息超时时间为5s
	ErrNum                             int   = 5                 //传输失败重试次数 5次
	Lenth                              int64 = 2 * 1024 * 1024   //每次传输2M
	Update_version_expiration_interval       = 30 * time.Second  //定时检查更新间隔时间为30s
	TempFileSuffix                           = "_tmp"            //临时文件后缀名称
)
