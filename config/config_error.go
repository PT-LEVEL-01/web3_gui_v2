package config

import (
	"errors"
	"web3_gui/utils"
)

// 业务错误编号
var (
	ERROR_CODE_success             = utils.ERROR_CODE_success             //成功
	ERROR_CODE_system_error_remote = utils.ERROR_CODE_system_error_remote //系统错误，远程节点
	ERROR_CODE_system_error_self   = utils.ERROR_CODE_system_error_self   //系统错误，自己节点

	ERROR_CODE_params_format        = utils.RegErrCodeExistPanic(61005, "传入参数格式错误，不能解析参数") //
	ERROR_CODE_params_format_return = utils.RegErrCodeExistPanic(61006, "返回参数格式错误，不能解析参数") //
	ERROR_CODE_Not_present          = utils.RegErrCodeExistPanic(61007, "不存在")             //
	ERROR_CODE_exist                = utils.RegErrCodeExistPanic(61008, "已经存在")            //

	ERROR_CODE_password_fail = utils.RegErrCodeExistPanic(61009, "密码错误") //

	ERROR_CODE_file_transfer_classID_not_find = utils.RegErrCodeExistPanic(62001, "文件传输：传输单元ID未找到") //
	ERROR_CODE_file_transfer_file_nonexist    = utils.RegErrCodeExistPanic(62002, "文件传输：文件不存在")     //
	ERROR_CODE_file_transfer_No_permission    = utils.RegErrCodeExistPanic(62003, "文件传输：没有文件下载权限")  //
	ERROR_CODE_file_transfer_No_find_task     = utils.RegErrCodeExistPanic(62004, "文件传输：没有找到下载任务")  //

	ERROR_CODE_order_not_pay = utils.RegErrCodeExistPanic(63001, "有未支付订单") //
	ERROR_CODE_order_repeat  = utils.RegErrCodeExistPanic(63002, "订单重复")   //

	ERROR_CODE_IM_In_the_friend_list                = utils.RegErrCodeExistPanic(64001, "好友在列表中")                         //
	ERROR_CODE_IM_invalid_Agree_Add_Friend          = utils.RegErrCodeExistPanic(64002, "无效的同意添加好友，对方并未申请添加好友，你却同意。")     //
	ERROR_CODE_IM_user_not_exist                    = utils.RegErrCodeExistPanic(64003, "用户不存在")                          //
	ERROR_CODE_IM_too_many_undelivered_messages     = utils.RegErrCodeExistPanic(64004, "未送达消息太多")                        //
	ERROR_CODE_IM_not_proxy                         = utils.RegErrCodeExistPanic(64005, "不是代理节点")                         //
	ERROR_CODE_IM_check_hash_fail                   = utils.RegErrCodeExistPanic(64006, "消息hash验证失败")                     //
	ERROR_CODE_IM_index_too_small                   = utils.RegErrCodeExistPanic(64007, "当本地上传消息与代理节点消息不连续时，本地消息index太小") //
	ERROR_CODE_IM_index_too_big                     = utils.RegErrCodeExistPanic(64008, "当本地上传消息与代理节点消息不连续时，本地消息index太大") //
	ERROR_CODE_IM_index_discontinuity               = utils.RegErrCodeExistPanic(64009, "数据链index不连续")                    //
	ERROR_CODE_IM_datachain_fork                    = utils.RegErrCodeExistPanic(64010, "代理节点的消息和本地消息不一致，消息分叉了")          //
	ERROR_CODE_IM_datachain_user_different          = utils.RegErrCodeExistPanic(64011, "本次保存的多条数据链中，用户不相同")              //
	ERROR_CODE_IM_datachain_cmd_fail                = utils.RegErrCodeExistPanic(64012, "命令错误")                           //
	ERROR_CODE_IM_datachain_cmd_exist               = utils.RegErrCodeExistPanic(64013, "命令未找到")                          //
	ERROR_CODE_IM_datachain_sendIndex_discontinuity = utils.RegErrCodeExistPanic(64014, "发送者的sendIndex不连续，有消息遗漏")         //
	ERROR_CODE_IM_datachain_params_fail             = utils.RegErrCodeExistPanic(64015, "参数错误")                           //
	ERROR_CODE_IM_datachain_exist                   = utils.RegErrCodeExistPanic(64016, "本条消息已经存在")                       //
	ERROR_CODE_IM_datachain_size_max_over           = utils.RegErrCodeExistPanic(64017, "本条消息内容超过了最大限制")                  //

	ERROR_CODE_IM_group_not_admin       = utils.RegErrCodeExistPanic(64117, "不是群管理员，无操作权限")   //
	ERROR_CODE_IM_group_not_member      = utils.RegErrCodeExistPanic(64118, "不是本群成员")         //
	ERROR_CODE_IM_group_shoutup         = utils.RegErrCodeExistPanic(64119, "群禁言了")           //
	ERROR_CODE_IM_group_exist           = utils.RegErrCodeExistPanic(64120, "群已经存在")          //
	ERROR_CODE_IM_group_not_exist       = utils.RegErrCodeExistPanic(64121, "群不存在")           //
	ERROR_CODE_IM_group_dissolve        = utils.RegErrCodeExistPanic(64122, "群已经解散")          //
	ERROR_CODE_IM_dh_version_unknown    = utils.RegErrCodeExistPanic(64123, "dh公钥信息版本未知")     //
	ERROR_CODE_IM_dh_not_exist          = utils.RegErrCodeExistPanic(64124, "dh公钥不存在，未找到")    //
	ERROR_CODE_IM_imgBase64_code_fail   = utils.RegErrCodeExistPanic(64125, "图片base64编码错误")   //
	ERROR_CODE_IM_forward_proxy         = utils.RegErrCodeExistPanic(64126, "请将信息发送给自己的代理节点") //
	ERROR_CODE_IM_nickname_over_size    = utils.RegErrCodeExistPanic(64127, "昵称超过最大长度")       //
	ERROR_CODE_IM_user_invitation_exist = utils.RegErrCodeExistPanic(64128, "近期添加好友申请已经存在")   //
	ERROR_CODE_IM_group_not_del_admin   = utils.RegErrCodeExistPanic(64129, "群不能删除管理员")       //

	ERROR_CODE_sharebox_Request_path_format_error = utils.RegErrCodeExistPanic(65001, "请求访问路径错误")      //
	ERROR_CODE_sharebox_Request_path_not_found    = utils.RegErrCodeExistPanic(65002, "请求访问路径不存在")     //
	ERROR_CODE_sharebox_Request_path_not_a_folder = utils.RegErrCodeExistPanic(65003, "请求访问路径不是一个文件夹") //
	ERROR_CODE_sharebox_process_full              = utils.RegErrCodeExistPanic(65004, "任务队列满了")        //

	ERROR_CODE_storage_db_full                        = utils.RegErrCodeExistPanic(66001, "磁盘已经满了")           //
	ERROR_CODE_storage_db_path_Homologous             = utils.RegErrCodeExistPanic(66002, "数据库路径同在一个磁盘中")     //
	ERROR_CODE_storage_encry_type_Not_Supported       = utils.RegErrCodeExistPanic(66003, "不支持的加密类型")         //
	ERROR_CODE_storage_auth_file_No_permission        = utils.RegErrCodeExistPanic(66004, "没有这个文件的操作权限")      //
	ERROR_CODE_storage_Insufficient_user_space        = utils.RegErrCodeExistPanic(66005, "用户存储空间不足")         //
	ERROR_CODE_storage_orders_not_exist               = utils.RegErrCodeExistPanic(66006, "订单不存在")            //
	ERROR_CODE_storage_Service_expiration_and_closure = utils.RegErrCodeExistPanic(66007, "服务到期关闭")           //
	ERROR_CODE_storage_orders_not_overtime            = utils.RegErrCodeExistPanic(66008, "订单未到续费时间")         //
	ERROR_CODE_storage_orders_overtime                = utils.RegErrCodeExistPanic(66009, "订单到期")             //
	ERROR_CODE_storage_del_dirAndFile_NotSameFolder   = utils.RegErrCodeExistPanic(66010, "删除的文件和文件夹不是同一文件夹") //
	ERROR_CODE_storage_No_need_upload_files           = utils.RegErrCodeExistPanic(66011, "文件已经存在，不需要上传")     //
	ERROR_CODE_storage_over_free_space_limit          = utils.RegErrCodeExistPanic(66012, "超过用户空闲空间限制")       //
	ERROR_CODE_storage_over_pay_space_limit           = utils.RegErrCodeExistPanic(66013, "超过单一用户购买空间限制")     //

	ERROR_CODE_CIRCLE_classname_exist = utils.RegErrCodeExistPanic(67001, "类别名称已经存在") //

	ERROR_CODE_order_chain_not_finish = utils.RegErrCodeExistPanic(68001, "链端未准备完成") //
	ERROR_CODE_order_class_not_exist  = utils.RegErrCodeExistPanic(68002, "商品类型不存在") //
	ERROR_CODE_order_goodsId_noexist  = utils.RegErrCodeExistPanic(68003, "商品不存在")   //
	ERROR_CODE_order_goods_soldOut    = utils.RegErrCodeExistPanic(68004, "商品已经售完")  //
)

var (
	ERROR_byte_nomarl = []byte{2} //
	ERROR_byte_exist  = []byte{3} //

	ERROR_remote_peer_error = errors.New("ERROR_remote_peer_error") //
	ERROR_user_not_exist    = errors.New("user not exist")          //用户不存在

	ERROR_storage_user_remain_space_max       = errors.New("user remain space max")       //用户购买剩余空间超过了最大值
	ERROR_storage_user_remain_space_total_max = errors.New("user remain space total max") //用户购买总空间超过了最大值
	//ERROR_storage_user_renewal_space_total_max   = errors.New("user remain space total max")    //用户续费空间超过了最大值
	ERROR_storage_Service_expiration_and_closure = errors.New("Service expiration and closure") //服务到期关闭
	ERROR_storage_orders_not_exist               = errors.New("orders not exist")               //订单不存在
	ERROR_storage_orders_not_overtime            = errors.New("orders not over time")           //订单未到期
	ERROR_storage_orders_overtime                = errors.New("orders over time")               //订单到期

	ERROR_db_full                      = errors.New("disk full")                    //磁盘已经满了
	ERROR_db_path_error                = errors.New("db path error")                //数据库路径错误
	ERROR_db_path_Homologous           = errors.New("db path homologous")           //数据库路径同在一个磁盘中
	ERROR_encry_type_Not_Supported     = errors.New("Unsupported encryption type")  //不支持的加密类型
	ERROR_Insufficient_user_space_size = errors.New("Insufficient user space size") //用户的空间不足

	ERROR_net_param_format = errors.New("unknown error code") //参数不能解析

	ERROR_unknown_code = errors.New("unknown error code") //未知的错误编号
)
