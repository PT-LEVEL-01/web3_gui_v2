package file_transfer

import (
	"path/filepath"
	"time"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

const (
	FILEPATH_temp = "temp"

	Transfer_task_overtime = time.Hour * 24 * 7 //一个文件传输任务，不能超过7天
	//DBKEY_white_list     = "DBKEY_white_list" //白名单列表
	//DBKEY_share_dir      = "DBKEY_share_dir"  //保存共享文件夹目录列表

)

var (
	FILEPATH_Recfilepath = filepath.Join(FILEPATH_temp, "download")

	DBKEY_GenID        = config.DBKEY_GenID        //
	DBKEY_share_dir    = config.DBKEY_share_dir    //保存共享文件夹目录列表
	DBKEY_white_list   = config.DBKEY_white_list   //白名单列表
	DBKEY_auto_Receive = config.DBKEY_auto_Receive //是否自动接收文件

	FLOOD_CLASS_key = engine.RegClassNameExistPanic([]byte{2, 0, 0, 0, 0, 0, 1})

	NET_CODE_apply_download_shareBox uint64 = config.MSGID_file_transfer_1 //协商下载共享文件夹中的文件
	//NET_CODE_apply_download_shareBox_recv uint64 = config.MSGID_file_transfer_2 //协商下载共享文件夹中的文件 返回
	NET_CODE_apply_send_file uint64 = config.MSGID_file_transfer_3 //协商主动发送文件
	//NET_CODE_apply_send_file_recv         uint64 = config.MSGID_file_transfer_4 //协商主动发送文件 返回
	NET_CODE_get_file_chunk uint64 = config.MSGID_file_transfer_5 //获取文件块数据
	//NET_CODE_get_file_chunk_recv          uint64 = config.MSGID_file_transfer_6 //获取文件块数据 返回
	NET_CODE_notice_transfer_finish uint64 = config.MSGID_file_transfer_7 //通知文件传输完成
	//NET_CODE_notice_transfer_finish_recv  uint64 = config.MSGID_file_transfer_8 //通知文件传输完成 返回

	NET_overtime_retry = []time.Duration{time.Second, time.Second * 2, time.Second * 4, time.Second * 8,
		time.Second * 60, time.Second * 60, time.Second * 60 * 10}
	NET_timeout = time.Second * 10
)

/*
一个不会出错的方法
*/
func BuildDbKey(key uint64) utilsleveldb.LeveldbKey {
	dbkey, _ := utilsleveldb.BuildLeveldbKey(utils.Uint64ToBytesByBigEndian(key))
	return *dbkey
}

/*
计算块大小
*/
//func ChunkSizeCalculators(size utils.Byte, spend time.Duration) utils.Byte {
//	if size == 0 && spend == 0 {
//		return utils.KB
//	}
//	if spend > time.Second/100 {
//		return size / 2
//	} else {
//		return size * 2
//	}
//}
