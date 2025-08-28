package config

import (
	"web3_gui/utils"
)

var (
	ERROR_code_dial_self                     = utils.RegErrCodeExistPanic(100002, "自己连接自己")      //
	ERROR_code_areaname_different            = utils.RegErrCodeExistPanic(100003, "域名称不同")       //
	ERROR_code_offline                       = utils.RegErrCodeExistPanic(100004, "节点离线")        //
	ERROR_code_params_fail                   = utils.RegErrCodeExistPanic(100005, "参数错误")        //
	ERROR_code_get_session_fail              = utils.RegErrCodeExistPanic(100006, "获取会话失败")      //
	ERROR_code_wait_msg_timeout              = utils.RegErrCodeExistPanic(100007, "等待消息返回超时")    //
	ERROR_code_Repeated_connections          = utils.RegErrCodeExistPanic(100008, "重复的连接")       //
	ERROR_code_session_close                 = utils.RegErrCodeExistPanic(100009, "会话已经关闭")      //
	ERROR_code_search_node_info_not_exist    = utils.RegErrCodeExistPanic(100010, "搜索用户")        //
	ERROR_code_security_store_not_exist      = utils.RegErrCodeExistPanic(100011, "双棘轮缓存中不存在")   //
	ERROR_code_send_ratchet_not_exist        = utils.RegErrCodeExistPanic(100012, "发送棘轮不存在")     //
	ERROR_code_recv_security_store_not_exist = utils.RegErrCodeExistPanic(100013, "接收双棘轮缓存中不存在") //
	ERROR_code_recv_ratchet_not_exist        = utils.RegErrCodeExistPanic(100014, "接收棘轮不存在")     //
	ERROR_code_read_or_write_size_fail       = utils.RegErrCodeExistPanic(100015, "读或者写的长度不对")   //
	ERROR_code_nodeinfo_illegal              = utils.RegErrCodeExistPanic(100016, "节点信息不合法")     //
)
