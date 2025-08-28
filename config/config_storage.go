package config

import (
	"time"
)

const (
	STORAGE_server_RenewalTime = 30  //最后30天可以续费
	STORAGE_server_UseTimeMax  = 365 //一个用户租用最大时间 单位：天

	STORAGE_server_dbpath_name = "storageleveldb" //云存储数据存放目录名称
)

const (
	Clean_Interval = time.Second * 10

	FileName_dir_temp_name      = "temp"                                //临时文件夹名称
	FileName_dir_filechunk_name = FileName_dir_temp_name + "/filechunk" //文件的加密切片文件夹名称

	FileChunk_version      = 1                //文件分片存储版本号
	EncryptionType_AES_ctr = 1                //文件分片加密类型
	AES_CTR_DEFAULT_KEY    = "morenmima"      //加密默认密码
	Chunk_size             = 1024 * 1024 * 16 //16M一个块

	PermissionType_self       = 0 //仅自己可见
	PermissionType_Designated = 1 //指定的用户可以访问
	PermissionType_all        = 2 //所有用户可以访问

	UploadFile_status_start       = 0 //上传文件状态，等待切片
	UploadFile_status_diced       = 1 //上传文件状态，切片状态
	UploadFile_status_upload      = 2 //上传文件状态，上传到网络中状态
	UploadFile_status_saveDB      = 3 //上传文件状态，服务器保存到数据库状态
	UploadFile_status_finish      = 4 //上传文件状态，完成状态
	UploadFile_status_finish_over = 5 //上传文件状态，删除列表状态
	UploadFile_status_stop        = 6 //上传文件状态，暂停上传状态

	UploadFile_server_status_start       = 10 //上传文件状态，等待传输
	UploadFile_server_status_transfer    = 11 //上传文件状态，开始传输
	UploadFile_server_status_savedb      = 12 //上传文件状态，开始保存到数据库
	UploadFile_server_status_finish      = 13 //上传文件状态，完成状态
	UploadFile_server_status_finish_over = 14 //上传文件状态，删除列表状态
	UploadFile_server_status_stop        = 15 //上传文件状态，暂停状态

	UploadFile_Diced_total = 1024 * 1024 * 1024 //上传文件切片文件夹总大小

)

var (
	MSG_timeout_interval = []int64{20, 20, 20, 60, 60, 60 * 60}
)
