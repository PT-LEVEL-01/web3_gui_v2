package config

import "path/filepath"

const (
	FILE_TRANSFER_CLASS_storage      = 100 //云存储使用的文件传输单元
	FILE_TRANSFER_CLASS_im_send_file = 101 //im聊天主动发送文件传输单元

)

var (
	FILE_TRANSFER_storage_dir           = "temp"                                               //文件传输临时目录
	FILE_TRANSFER_storage_download_path = filepath.Join(FILE_TRANSFER_storage_dir, "download") //云存储传输的临时文件存放目录

	FLOOD_key_file_transfer = BuildFloodKey(2000001) //文件传输
)
