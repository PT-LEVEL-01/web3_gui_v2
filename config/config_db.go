package config

import (
	"web3_gui/utils/utilsleveldb"
)

var (
	//在线及时消息
	DBKEY_self_userinfo               = utilsleveldb.RegDbKeyExistPanicByUint64(10001) //个人基本信息
	DBKEY_apply_local_userlist        = utilsleveldb.RegDbKeyExistPanicByUint64(10002) //添加好友列表:自己主动添加对方的好友申请列表
	DBKEY_apply_local_userlist_index  = utilsleveldb.RegDbKeyExistPanicByUint64(10003) //添加好友列表:index索引
	DBKEY_apply_local_userlist_addr   = utilsleveldb.RegDbKeyExistPanicByUint64(10004) //添加好友列表:地址索引
	DBKEY_apply_remote_userlist       = utilsleveldb.RegDbKeyExistPanicByUint64(10005) //添加好友列表:对方添加自己的好友申请列表
	DBKEY_apply_remote_userlist_index = utilsleveldb.RegDbKeyExistPanicByUint64(10006) //添加好友列表:index索引
	DBKEY_apply_remote_userlist_addr  = utilsleveldb.RegDbKeyExistPanicByUint64(10007) //添加好友列表:地址索引
	DBKEY_friend_userlist             = utilsleveldb.RegDbKeyExistPanicByUint64(10008) //好友列表:已经添加的好友
	DBKEY_message_history_index       = utilsleveldb.RegDbKeyExistPanicByUint64(10009) //聊天内容编号
	DBKEY_message_history             = utilsleveldb.RegDbKeyExistPanicByUint64(10010) //聊天内容历史
	DBKEY_message_undelivered_total   = utilsleveldb.RegDbKeyExistPanicByUint64(10011) //未发送成功消息总量
	DBKEY_message_send_id             = utilsleveldb.RegDbKeyExistPanicByUint64(10012) //保存发送ID对应的消息Index
	DBKEY_message_recv_id             = utilsleveldb.RegDbKeyExistPanicByUint64(10013) //保存接收ID对应的消息Index

	//离线消息
	DBKEY_improxy_client_list                     = utilsleveldb.RegDbKeyExistPanicByUint64(20001) //服务器列表
	DBKEY_improxy_client_orders                   = utilsleveldb.RegDbKeyExistPanicByUint64(20002) //客户端所有订单
	DBKEY_improxy_client_ordersID_not_pay         = utilsleveldb.RegDbKeyExistPanicByUint64(20003) //客户端未支付的订单ID
	DBKEY_improxy_client_ordersID_pay_notOnChain  = utilsleveldb.RegDbKeyExistPanicByUint64(20004) //客户端已支付但未上链的订单
	DBKEY_improxy_client_ordersID_inuse           = utilsleveldb.RegDbKeyExistPanicByUint64(20005) //客户端订单ID，已支付已上链，正在服务的订单
	DBKEY_improxy_client_ordersID_timeout         = utilsleveldb.RegDbKeyExistPanicByUint64(20006) //客户端订单ID，过期订单，不管是已支付还是未支付，都放在这里
	DBKEY_improxy_client_chain_count_height       = utilsleveldb.RegDbKeyExistPanicByUint64(20007) //保存服务端->统计的区块高度
	DBKEY_improxy_server_orders                   = utilsleveldb.RegDbKeyExistPanicByUint64(20501) //存储服务器->所有订单，包括已支付的，未支付的
	DBKEY_improxy_server_info_self                = utilsleveldb.RegDbKeyExistPanicByUint64(20502) //保存本节点提供服务的个人信息
	DBKEY_improxy_server_ordersID_not_pay         = utilsleveldb.RegDbKeyExistPanicByUint64(20503) //存储服务器->未支付的订单ID
	DBKEY_improxy_server_ordersID_not_pay_timeout = utilsleveldb.RegDbKeyExistPanicByUint64(20505) //存储服务器->未支付的订单ID，超时未支付
	DBKEY_improxy_server_ordersID_inuse           = utilsleveldb.RegDbKeyExistPanicByUint64(20506) //存储服务器->正在服务的订单ID
	DBKEY_improxy_server_ordersID_inuse_timeout   = utilsleveldb.RegDbKeyExistPanicByUint64(20507) //存储服务器->正在服务的订单ID，超过订单服务时间
	DBKEY_improxy_server_user_spaceSize           = utilsleveldb.RegDbKeyExistPanicByUint64(20508) //保存用户已经使用空间大小
	DBKEY_improxy_server_datachain_send_fail      = utilsleveldb.RegDbKeyExistPanicByUint64(20509) //发送失败的消息
	DBKEY_improxy_server_datachain_send_fail_id   = utilsleveldb.RegDbKeyExistPanicByUint64(20510) //发送失败的消息id索引，方便发送成功后删除
	DBKEY_improxy_server_chain_count_height       = utilsleveldb.RegDbKeyExistPanicByUint64(20511) //保存服务端统计的区块高度

	DBKEY_improxy_userinfo                           = utilsleveldb.RegDbKeyExistPanicByUint64(30001) //保存用户基本信息
	DBKEY_improxy_user_datachain                     = utilsleveldb.RegDbKeyExistPanicByUint64(30002) //保存用户的消息日志，不分客户端和服务器
	DBKEY_improxy_user_datachain_id                  = utilsleveldb.RegDbKeyExistPanicByUint64(30003) //消息日志ID
	DBKEY_improxy_user_datachain_sendid              = utilsleveldb.RegDbKeyExistPanicByUint64(30004) //数据链发送者id
	DBKEY_improxy_user_datachain_shot_index          = utilsleveldb.RegDbKeyExistPanicByUint64(30005) //快照高度
	DBKEY_improxy_user_message_history               = utilsleveldb.RegDbKeyExistPanicByUint64(30006) //聊天内容历史
	DBKEY_improxy_user_message_history_sendID        = utilsleveldb.RegDbKeyExistPanicByUint64(30007) //聊天内容保存sendid索引
	DBKEY_improxy_user_message_history_fileHash_send = utilsleveldb.RegDbKeyExistPanicByUint64(30008) //聊天内容保存文件hash索引
	DBKEY_improxy_user_message_history_fileHash_recv = utilsleveldb.RegDbKeyExistPanicByUint64(30009) //聊天内容保存文件hash索引
	DBKEY_improxy_user_datachain_send_fail           = utilsleveldb.RegDbKeyExistPanicByUint64(30010) //发送失败的消息
	DBKEY_improxy_user_datachain_send_fail_id        = utilsleveldb.RegDbKeyExistPanicByUint64(30011) //发送失败的消息id索引，方便发送成功后删除
	DBKEY_improxy_user_datachain_sendIndex_knit      = utilsleveldb.RegDbKeyExistPanicByUint64(30012) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息
	DBKEY_improxy_user_datachain_sendIndex_parse     = utilsleveldb.RegDbKeyExistPanicByUint64(30013) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息

	DBKEY_improxy_group_datachain_sendIndex_knit  = utilsleveldb.RegDbKeyExistPanicByUint64(30531) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息
	DBKEY_improxy_group_datachain_sendIndex_minor = utilsleveldb.RegDbKeyExistPanicByUint64(30532) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息
	DBKEY_improxy_group_datachain_sendIndex_parse = utilsleveldb.RegDbKeyExistPanicByUint64(30533) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息
	DBKEY_improxy_group_list_create               = utilsleveldb.RegDbKeyExistPanicByUint64(30534) //自己创建的群列表
	DBKEY_improxy_group_list_join                 = utilsleveldb.RegDbKeyExistPanicByUint64(30535) //自己加入的群列表
	DBKEY_improxy_group_list_knit                 = utilsleveldb.RegDbKeyExistPanicByUint64(30536) //自己负责构建的群
	DBKEY_improxy_group_list_minor                = utilsleveldb.RegDbKeyExistPanicByUint64(30537) //自己同步代理的群
	DBKEY_improxy_group_list_dissolve             = utilsleveldb.RegDbKeyExistPanicByUint64(30538) //解散的群列表
	DBKEY_improxy_group_members_knit              = utilsleveldb.RegDbKeyExistPanicByUint64(30539) //构建者群成员列表
	DBKEY_improxy_group_members_parser            = utilsleveldb.RegDbKeyExistPanicByUint64(30540) //解析者群成员列表
	DBKEY_improxy_group_shot_index_knit           = utilsleveldb.RegDbKeyExistPanicByUint64(30541) //群构建快照高度
	DBKEY_improxy_group_shot_index_minor          = utilsleveldb.RegDbKeyExistPanicByUint64(30542) //群副代理快照高度
	DBKEY_improxy_group_shot_index_parser         = utilsleveldb.RegDbKeyExistPanicByUint64(30543) //群解析快照高度
	DBKEY_improxy_group_datachain                 = utilsleveldb.RegDbKeyExistPanicByUint64(30544) //保存群的消息日志，不分客户端和服务器
	DBKEY_improxy_group_datachain_id              = utilsleveldb.RegDbKeyExistPanicByUint64(30545) //群消息日志ID
	DBKEY_improxy_group_datachain_sendid          = utilsleveldb.RegDbKeyExistPanicByUint64(30546) //数据链发送者id
	DBKEY_improxy_group_datachain_member_start    = utilsleveldb.RegDbKeyExistPanicByUint64(30547) //成员加入的index
	DBKEY_improxy_send_file_list                  = utilsleveldb.RegDbKeyExistPanicByUint64(30548) //待发送文件列表
	DBKEY_improxy_recv_file_list                  = utilsleveldb.RegDbKeyExistPanicByUint64(30549) //接收文件列表

	DBKEY_improxy_server_datachain_nolink    = utilsleveldb.RegDbKeyExistPanicByUint64(30701) //代理节点保存未标记接收的离线消息
	DBKEY_improxy_server_datachain_nolink_id = utilsleveldb.RegDbKeyExistPanicByUint64(30702) //代理节点保存未标记接收的离线消息ID
	DBKEY_improxy_server_datachain_sendIndex = utilsleveldb.RegDbKeyExistPanicByUint64(30703) //发送者地址和接收者地址联合key，保存每一个发送者的index，这个index必须连续，否则视为漏消息

	// DBKEY_circle_class = utilsleveldb.RegDbKeyExistPanicByUint64(2001) //我添加的分类
	DBKEY_circle_news_release = utilsleveldb.RegDbKeyExistPanicByUint64(90001) //我编辑的新闻保存到已经发布列表
	DBKEY_circle_news_draft   = utilsleveldb.RegDbKeyExistPanicByUint64(90002) //我编辑的新闻保存到草稿箱列表

	//云存储服务
	DBKEY_storage_server_GenID                    = utilsleveldb.RegDbKeyExistPanicByUint64(40001) //存储服务器全局唯一自增长ID
	DBKEY_storage_serverinfo                      = utilsleveldb.RegDbKeyExistPanicByUint64(40002) //保存本节点提供存储服务的信息
	DBKEY_storage_serverlist                      = utilsleveldb.RegDbKeyExistPanicByUint64(40003) //保存提供存储服务列表
	DBKEY_storage_server_orders                   = utilsleveldb.RegDbKeyExistPanicByUint64(40004) //存储服务器所有订单
	DBKEY_storage_server_ordersID_not_pay         = utilsleveldb.RegDbKeyExistPanicByUint64(40005) //存储服务器未支付的订单ID
	DBKEY_storage_server_ordersID_not_pay_timeout = utilsleveldb.RegDbKeyExistPanicByUint64(40006) //存储服务器未支付的订单ID，超时未支付
	DBKEY_storage_server_ordersID_inuse           = utilsleveldb.RegDbKeyExistPanicByUint64(40007) //存储服务器正在服务的订单ID
	DBKEY_storage_server_ordersID_inuse_timeout   = utilsleveldb.RegDbKeyExistPanicByUint64(40008) //存储服务器正在服务的订单ID，超过订单服务时间
	DBKEY_storage_server_user_top_dir_index       = utilsleveldb.RegDbKeyExistPanicByUint64(40009) //服务器保存用户的顶层文件目录
	DBKEY_storage_server_dir_index                = utilsleveldb.RegDbKeyExistPanicByUint64(40010) //服务器保存的文件列表
	DBKEY_storage_server_file_index               = utilsleveldb.RegDbKeyExistPanicByUint64(40011) //服务器保存的文件索引
	DBKEY_storage_server_file_chunk               = utilsleveldb.RegDbKeyExistPanicByUint64(40012) //服务器保存的文件分片数据
	DBKEY_storage_server_upload_file_index        = utilsleveldb.RegDbKeyExistPanicByUint64(40013) //服务器保存的正在上传中的文件索引ID
	DBKEY_storage_server_user_spaceSize           = utilsleveldb.RegDbKeyExistPanicByUint64(40014) //保存用户已经使用空间大小
	DBKEY_storage_client_orders                   = utilsleveldb.RegDbKeyExistPanicByUint64(40101) //客户端所有订单
	DBKEY_storage_client_ordersID_not_pay         = utilsleveldb.RegDbKeyExistPanicByUint64(40102) //客户端未支付的订单ID
	DBKEY_storage_client_ordersID_not_pay_timeout = utilsleveldb.RegDbKeyExistPanicByUint64(40103) //客户端未支付的订单ID，超时未支付
	DBKEY_storage_client_ordersID_inuse           = utilsleveldb.RegDbKeyExistPanicByUint64(40104) //客户端订单ID，正在服务的订单
	DBKEY_storage_client_ordersID_inuse_timeout   = utilsleveldb.RegDbKeyExistPanicByUint64(40105) //客户端订单ID，超过服务时间的订单
	DBKEY_storage_client_file_pwd                 = utilsleveldb.RegDbKeyExistPanicByUint64(40106) //保存文件加密密码
	DBKEY_storage_client_uploading_list           = utilsleveldb.RegDbKeyExistPanicByUint64(40107) //文件上传中的列表
	DBKEY_storage_client_upload_finish_list       = utilsleveldb.RegDbKeyExistPanicByUint64(40108) //文件上传完成列表
	DBKEY_storage_client_downloading_list         = utilsleveldb.RegDbKeyExistPanicByUint64(40109) //文件下载中的列表
	DBKEY_storage_client_download_finish_list     = utilsleveldb.RegDbKeyExistPanicByUint64(40110) //文件下载完成的列表

	//文件传输模块
	DBKEY_GenID        = utilsleveldb.RegDbKeyExistPanicByUint64(50101) //
	DBKEY_share_dir    = utilsleveldb.RegDbKeyExistPanicByUint64(50102) //保存共享文件夹目录列表
	DBKEY_white_list   = utilsleveldb.RegDbKeyExistPanicByUint64(50103) //白名单列表
	DBKEY_auto_Receive = utilsleveldb.RegDbKeyExistPanicByUint64(50104) //是否自动接收文件

	DBKEY_sharebox_price_list                   = utilsleveldb.RegDbKeyExistPanicByUint64(60001) //共享文件夹->文件价格
	DBKEY_SHAREBOX_server_order_unpaid          = utilsleveldb.RegDbKeyExistPanicByUint64(60002) //共享文件夹->未支付订单
	DBKEY_SHAREBOX_server_order_paid            = utilsleveldb.RegDbKeyExistPanicByUint64(60003) //共享文件夹->已支付订单
	DBKEY_SHAREBOX_server_order_unpaid_overtime = utilsleveldb.RegDbKeyExistPanicByUint64(60004) //共享文件夹->过期未支付订单

	DBKEY_SHAREBOX_client_order_unpaid          = utilsleveldb.RegDbKeyExistPanicByUint64(60101) //共享文件夹->未支付订单
	DBKEY_SHAREBOX_client_order_paid            = utilsleveldb.RegDbKeyExistPanicByUint64(60102) //共享文件夹->已支付订单
	DBKEY_SHAREBOX_client_order_unpaid_overtime = utilsleveldb.RegDbKeyExistPanicByUint64(60103) //共享文件夹->过期未支付订单

	DBKEY_ORDERS_goods        = utilsleveldb.RegDbKeyExistPanicByUint64(70001) //订单系统->商品
	DBKEY_ORDERS_order_unpaid = utilsleveldb.RegDbKeyExistPanicByUint64(70002) //订单系统->未支付订单
	DBKEY_ORDERS_order_paid   = utilsleveldb.RegDbKeyExistPanicByUint64(70003) //订单系统->已支付订单

	DBKEY_ChainCount_order_server = utilsleveldb.RegDbKeyExistPanicByUint64(80001) //区块统计->同步到的统计高度，服务端
	DBKEY_ChainCount_order_client = utilsleveldb.RegDbKeyExistPanicByUint64(80002) //区块统计->同步到的统计高度，客户端

)

/*
构建好友消息编号最大值
*/
func BuildKeyUserMsgIndex(selfAddr, remoteAddr []byte) utilsleveldb.LeveldbKey {
	// bs := append([]byte(DBKEY_message_history_index), selfAddr...)
	// return append(bs, remoteAddr...)
	bs, _ := DBKEY_message_history_index.BaseKey()
	bs = append(bs, selfAddr...)
	bs = append(bs, remoteAddr...)
	key, _ := utilsleveldb.BuildLeveldbKey(bs)
	return *key
}

/*
构建好友消息内容key
*/
func BuildKeyUserMsgContent(selfAddr, remoteAddr, indexBs []byte) utilsleveldb.LeveldbKey {
	// bs := append([]byte(DBKEY_message_history), selfAddr...)
	// bs = append(bs, remoteAddr...)
	// bs = append(bs, indexBs...)
	bs, _ := DBKEY_message_history.BaseKey()
	bs = append(bs, selfAddr...)
	bs = append(bs, remoteAddr...)
	bs = append(bs, indexBs...)
	key, _ := utilsleveldb.BuildLeveldbKey(bs)
	return *key
}
