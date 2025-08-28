package config

import (
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
)

const (
	NET_protocol_version_v1 = 1 //网络协议版本号

	MSGID_multicast_message_recv    = 35000 //广播消息
	MSGID_get_friend_info           = 35001 //获取好友基本信息
	MSGID_get_friend_info_recv      = 35002 //获取好友基本信息 返回
	MSGID_add_friend                = 35003 //申请添加好友
	MSGID_add_friend_recv           = 35004 //申请添加好友 返回
	MSGID_agree_add_friend          = 35005 //同意添加好友
	MSGID_agree_add_friend_recv     = 35006 //同意添加好友 返回
	MSGID_send_friend_message       = 35007 //发送私聊消息
	MSGID_send_friend_message_recv  = 35008 //发送私聊消息 返回
	MSGID_file_getShareboxList      = 35009 //查询共享目录列表
	MSGID_file_getShareboxList_recv = 35010 //查询共享目录列表 返回
	MSGID_multicast_online_recv     = 35011 //用户上线广播消息

	MSGID_IM_PROXY_multicast_nodeinfo_recv = 36001 //提供代理节点广播消息
	MSGID_IM_PROXY_multicast_search        = 36002 //广播搜索某节点的代理节点
	MSGID_IM_PROXY_multicast_search_recv   = 36003 //广播搜索某节点的代理节点 返回
	MSGID_IM_PROXY_getorders               = 36004 //客户端向代理获取一个租用时间的订单
	MSGID_IM_PROXY_getorders_recv          = 36005 //客户端向代理获取一个租用空间的订单 返回
	MSGID_IM_PROXY_getRenewalOrders        = 36006 //客户端向代理获取续费订单
	MSGID_IM_PROXY_getRenewalOrders_recv   = 36007 //客户端向代理获取续费订单 返回

	MSGID_IM_PROXY_getDataChainIndex                   = 36008 //查询代理节点已经同步的数据链索引
	MSGID_IM_PROXY_getDataChainIndex_recv              = 36009 //查询代理节点已经同步的数据链索引 返回
	MSGID_IM_PROXY_upload_datachain                    = 36010 //上传数据链
	MSGID_IM_PROXY_upload_datachain_recv               = 36011 //上传数据链 返回
	MSGID_IM_PROXY_sync_datachain                      = 36012 //代理服务器获取客户端的数据链
	MSGID_IM_PROXY_sync_datachain_recv                 = 36013 //代理服务器获取客户端的数据链 返回
	MSGID_IM_PROXY_send_datachain                      = 36014 //发送数据链
	MSGID_IM_PROXY_send_datachain_recv                 = 36015 //发送数据链 返回
	MSGID_IM_PROXY_download_datachain                  = 36016 //下载数据链
	MSGID_IM_PROXY_download_datachain_recv             = 36017 //下载数据链 返回
	MSGID_IM_PROXY_group_send_datachain                = 36018 //发送群数据链消息
	MSGID_IM_PROXY_group_send_datachain_recv           = 36019 //发送群数据链消息 返回
	MSGID_IM_PROXY_group_multicast_datachain           = 36020 //群内成员广播数据链消息
	MSGID_IM_PROXY_group_multicast_datachain_recv      = 36021 //群内成员广播数据链消息 返回
	MSGID_IM_PROXY_group_download_datachain_start      = 36022 //成员查询群数据链起始记录
	MSGID_IM_PROXY_group_download_datachain_start_recv = 36023 //成员查询群数据链起始记录 返回
	MSGID_IM_PROXY_group_download_datachain            = 36024 //下载群数据链记录
	MSGID_IM_PROXY_group_download_datachain_recv       = 36025 //下载群数据链记录 返回
	MSGID_IM_PROXY_get_sendindex                       = 36026 //从好友获取自己的sendIndex
	MSGID_IM_PROXY_get_sendindex_recv                  = 36027 //从好友获取自己的sendIndex 返回

	MSGID_circle_getClassNames       = 45011 //查询好友博客圈子类别
	MSGID_circle_getClassNames_recv  = 45012 //查询好友博客圈子类别 返回
	MSGID_circle_multicast_news_recv = 45013 //其他用户广播的博客 接收广播

	//文件传输模块保留消息编号
	MSGID_file_transfer_1 = 56001 //
	MSGID_file_transfer_2 = 56002 //
	MSGID_file_transfer_3 = 56003 //
	MSGID_file_transfer_4 = 56004 //
	MSGID_file_transfer_5 = 56005 //
	MSGID_file_transfer_6 = 56006 //
	MSGID_file_transfer_7 = 56007 //
	MSGID_file_transfer_8 = 56008 //

	MSGID_file_transfer_11 = 56011 //
	MSGID_file_transfer_12 = 56012 //
	MSGID_file_transfer_13 = 56013 //
	MSGID_file_transfer_14 = 56014 //
	MSGID_file_transfer_15 = 56015 //
	MSGID_file_transfer_16 = 56016 //

	//版本更新模块保留消息标号
	MSGID_version_update_1 = 57001 //
	MSGID_version_update_2 = 57002 //
	MSGID_version_update_3 = 57003 //
	MSGID_version_update_4 = 57004 //
	MSGID_version_update_5 = 57005 //
	MSGID_version_update_6 = 57006 //

	MSGID_STORAGE_multicast_nodeinfo_recv = 67000 //存储空提供者间节点广播消息
	MSGID_STORAGE_getorders               = 67001 //客户端向存储提供商获取一个租用空间的订单
	MSGID_STORAGE_getorders_recv          = 67002 //客户端向存储提供商获取一个租用空间的订单 返回
	MSGID_STORAGE_getRenewalOrders        = 67003 //客户端向存储提供商获取续费订单
	MSGID_STORAGE_getRenewalOrders_recv   = 67004 //客户端向存储提供商获取续费订单 返回
	MSGID_STORAGE_getFreeSpace            = 67005 //客户端向存储提供商获取自己的可用空间
	MSGID_STORAGE_getFreeSpace_recv       = 67006 //客户端向存储提供商获取自己的可用空间 返回
	MSGID_STORAGE_getFileList             = 67007 //客户端向存储提供商获取文件列表
	MSGID_STORAGE_getFileList_recv        = 67008 //客户端向存储提供商获取文件列表 返回

	MSGID_STORAGE_upload_fileindex                   = 67009 //客户端向存储提供商发送要上传的文件索引
	MSGID_STORAGE_upload_fileindex_recv              = 67010 //客户端向存储提供商发送要上传的文件索引 返回
	MSGID_STORAGE_upload_transfer_finish             = 67011 //服务端告诉客户端传输完成
	MSGID_STORAGE_upload_transfer_finish_recv        = 67012 //服务端告诉客户端传输完成 返回
	MSGID_STORAGE_upload_savedb_finish               = 67013 //服务端告诉客户端保存数据库完成
	MSGID_STORAGE_upload_savedb_finish_recv          = 67014 //服务端告诉客户端保存数据库完成 返回
	MSGID_STORAGE_upload_stop                        = 67015 //客户端暂停上传
	MSGID_STORAGE_upload_stop_recv                   = 67016 //客户端暂停上传 返回
	MSGID_STORAGE_upload_reset                       = 67017 //客户端暂停上传后恢复
	MSGID_STORAGE_upload_reset_recv                  = 67018 //客户端暂停上传后恢复 返回
	MSGID_STORAGE_upload_delete_dir_and_file         = 67019 //客户端删除文件和文件夹
	MSGID_STORAGE_upload_delete_dir_and_file_recv    = 67020 //客户端删除文件和文件夹 返回
	MSGID_STORAGE_upload_create_dir                  = 67021 //客户端创建文件夹
	MSGID_STORAGE_upload_create_dir_recv             = 67022 //客户端创建文件夹 返回
	MSGID_STORAGE_download_select_dir_fileindex      = 67023 //递归查询多个文件夹中的文件列表
	MSGID_STORAGE_download_select_dir_fileindex_recv = 67024 //递归查询多个文件夹中的文件列表 返回
	MSGID_STORAGE_download_apply                     = 67025 //申请下载文件
	MSGID_STORAGE_download_apply_recv                = 67026 //申请下载文件 返回
	MSGID_STORAGE_upload_update_name                 = 67027 //客户端修改文件和文件夹名称
	MSGID_STORAGE_upload_update_name_recv            = 67028 //客户端修改文件和文件夹名称 返回

	MSGID_SHAREBOX_Order_getOrder = 68001 //获取订单

)

/*
给对方返回一个系统错误
*/
func ReplyError(area *libp2parea.Node, version uint64, message *message_center.MessageBase, code uint64, msg string) {
	nr := utils.NewNetResult(version, code, msg, nil)
	bs, err := nr.Proto()
	if err != nil {
		area.SendP2pReplyMsgHE(message, nil)
		return
	}
	area.SendP2pReplyMsgHE(message, bs)
}

/*
给对方返回成功
*/
func ReplySuccess(area *libp2parea.Node, version uint64, message *message_center.MessageBase, bs *[]byte) {
	resultBs := []byte{}
	if bs != nil {
		resultBs = *bs
	}
	nr := utils.NewNetResult(version, ERROR_CODE_success, "", resultBs)
	bs, err := nr.Proto()
	if err != nil {
		area.SendP2pReplyMsgHE(message, nil)
		return
	}
	area.SendP2pReplyMsgHE(message, bs)
}
