package transfer_manager

import (
	"time"
	"web3_gui/utils"
)

const (
	MinLenth                          int64 = 100 * 1024       //每次传最小（100kb）
	MaxLenth                          int64 = 20 * 1024 * 1024 //每次传输最大（20M）
	Transfer_slice_max_time_diff      int64 = 100              //每次传输最大时间差阀值（100ms）
	ErrNum                            int   = 5                //传输失败重试次数 5次
	Second                            int64 = 1                //传输速度统计时间间隔 1秒
	Recfilepath                             = "files"
	Transferlog_push_task_db_key            = "transfer_push_task_db_key"
	Transferlog_pull_task_db_key            = "transfer_pull_task_db_key"
	Transfer_push_task_id_max               = "transfer_push_task_id_max"
	Transfer_pull_task_id_max               = "transfer_pull_task_id_max"
	Transfer_task_expiration_interval       = 24 * 60 * 60 * time.Second          //任务有效期24小时
	Transfer_push_task_sharing_dirs         = "transfer_push_task_sharing_dirs"   // 共享文件夹
	Transfer_pull_addr_white_list           = "transfer_pull_addr_white_list"     //有权限拉取文件的地址白名单
	Transfer_pull_task_if_atuo_db_key       = "transfer_pull_task_if_atuo_db_key" //
	Transfer_p2p_mgs_timeout                = 5 * time.Second                     //p2p消息超时时间为5s
)

const (
	Transfer_pull_task_stautus_pending_confirmation = "pendingConfirmation" //任务待确认
	Transfer_pull_task_stautus_running              = "running"             // 运行中
	Transfer_pull_task_stautus_stop                 = "stop"                // 已停止
)

var msg_id_p2p_transfer_push uint64 = 1001      //请求传输文件
var msg_id_p2p_transfer_push_recv uint64 = 1002 //请求传输文件回复

var msg_id_p2p_transfer_pull uint64 = 1003      //请求拉取文件流
var msg_id_p2p_transfer_pull_recv uint64 = 1004 //请求拉取文件流回复

var msg_id_p2p_transfer_new_pull uint64 = 1005      //创建拉取任务
var msg_id_p2p_transfer_new_pull_recv uint64 = 1006 //创建拉取任务回复

var (
	ERROR_CODE_task_not_exist         = utils.RegErrCodeExistPanic(80001, "传输任务不存在")      //
	ERROR_CODE_file_pull_fail         = utils.RegErrCodeExistPanic(80002, "文件拉取失败")       //
	ERROR_CODE_task_list_zero         = utils.RegErrCodeExistPanic(80003, "没有任务列表")       //
	ERROR_CODE_task_repeat            = utils.RegErrCodeExistPanic(80004, "任务重复")         //
	ERROR_CODE_not_update_dir         = utils.RegErrCodeExistPanic(80005, "不能修改存储目录")     //
	ERROR_CODE_not_stop               = utils.RegErrCodeExistPanic(80006, "不能停止该任务")      //
	ERROR_CODE_file_path_not_abs      = utils.RegErrCodeExistPanic(80007, "文件路径不是绝对路径")   //
	ERROR_CODE_file_content_size_zero = utils.RegErrCodeExistPanic(80008, "文件内容为空")       //
	ERROR_CODE_task_params_fail       = utils.RegErrCodeExistPanic(80009, "任务参数错误")       //
	ERROR_CODE_create_task_fail       = utils.RegErrCodeExistPanic(80010, "创建任务失败")       //
	ERROR_CODE_task_not_auth          = utils.RegErrCodeExistPanic(80011, "任务无权限")        //
	ERROR_CODE_task_overtime          = utils.RegErrCodeExistPanic(80012, "任务过期")         //
	ERROR_CODE_dir_path_fail          = utils.RegErrCodeExistPanic(80013, "目录路径错误")       //
	ERROR_CODE_dir_name_repeat        = utils.RegErrCodeExistPanic(80014, "路径的最后文件夹名称重复") //
	ERROR_CODE_file_damage            = utils.RegErrCodeExistPanic(80015, "文件已损坏")        //
	ERROR_CODE_create_push_fail       = utils.RegErrCodeExistPanic(80016, "创建推送任务失败")     //
)
