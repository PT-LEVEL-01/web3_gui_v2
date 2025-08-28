package file_transfer

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
)

/*
注册文件传输Handers
*/
func (this *Manager) RegisterHandlers() {
	this.area.Register_p2p(NET_CODE_apply_download_shareBox, this.downloadShareBox) //协商下载共享文件夹中的文件
	//this.area.Register_p2p(NET_CODE_apply_download_shareBox_recv, this.downloadShareBox_recv)    //协商下载共享文件夹中的文件 返回
	this.area.Register_p2p(NET_CODE_apply_send_file, this.applySendFile) //协商主动发送文件
	//this.area.Register_p2p(NET_CODE_apply_send_file_recv, this.applySendFile_recv)               //协商主动发送文件 返回
	this.area.Register_p2p(NET_CODE_get_file_chunk, this.getFileChunk) //获取文件块数据
	//this.area.Register_p2p(NET_CODE_get_file_chunk_recv, this.getFileChunk_recv)                 //获取文件块数据 返回
	this.area.Register_p2p(NET_CODE_notice_transfer_finish, this.noticeTransferFinish) //通知文件传输完成
	//this.area.Register_p2p(NET_CODE_notice_transfer_finish_recv, this.noticeTransferFinish_recv) //通知文件传输完成 返回

}

/*
申请下载共享文件夹中的文件
*/
func (this *Manager) downloadShareBox(message *message_center.MessageBase) {
	//replyMsgID := uint64(NET_CODE_apply_download_shareBox_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	fi, err := ParseFileIndex(np.Data)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	//
	ns := strings.SplitN(fi.RemoteFilePath, "\\", 2)
	if len(ns) != 2 {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_file_nonexist, "")
		return
	}
	t := this.GetClass(fi.ClassID)
	if t == nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_classID_not_find, "")
		return
	}
	path := t.GetShareDir(ns[0])
	if path == "" {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_file_nonexist, "")
		return
	}
	//验证白名单
	ok := t.CheckWhiteList(*message.SenderAddr)
	if !ok {
		utils.Log.Info().Msgf("白名单拦截，没有权限")
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_No_permission, "")
		return
	}
	//创建任务
	ft, ERR := NewFileTransferTask(this.area.GetLevelDB(), fi.ClassID, filepath.Join(path, ns[1]), t.uploadFile, fi.PullID)
	if !ERR.CheckSuccess() {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	//先把任务保存好
	t.lock.Lock()
	t.transferTaskByID[utils.Bytes2string(ft.SupplierID)] = ft
	t.lock.Unlock()
	fileIndex, err := NewFileIndex(ft)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	bs, err := fileIndex.Proto()
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	config.ReplySuccess(this.area, config.NET_protocol_version_v1, message, bs)
}

/*
申请下载共享文件夹中的文件 返回
*/
//func (this *Manager) downloadShareBox_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(FLOOD_CLASS_key, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
申请发送文件
*/
func (this *Manager) applySendFile(message *message_center.MessageBase) {
	//replyMsgID := uint64(NET_CODE_apply_download_shareBox_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	fi, err := ParseFileIndex(np.Data)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}

	//utils.Log.Info().Msgf("文件路径:%+v", fi)
	//ns := strings.SplitN(fi.RemoteFilePath, "/", 2)
	//if len(ns) != 2 {
	//	config.ReplyError(this.area, config.NET_protocol_version_v1, message,  config.ERROR_CODE_file_transfer_file_nonexist, "")
	//	return
	//}

	t := this.GetClass(fi.ClassID)
	if t == nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_classID_not_find, "")
		return
	}

	//验证白名单
	ok := t.CheckWhiteList(*message.SenderAddr)
	if !ok {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_No_permission, "")
		return
	}

	localPath := filepath.Join(t.defaultDownloadPath, fi.Name)
	//idBs, err := GetGenID(this.area.GetLevelDB())
	//if err != nil {
	//	config.ReplyError(this.area, config.NET_protocol_version_v1, message,  config.ERROR_CODE_system_error_self, err.Error())
	//	return
	//}
	//fi.PullID = ulid.Make().Bytes()

	ds := NewDownloadStep(this.area, *message.SenderAddr, fi, localPath, this.pullFinishSignalMonitor, t.downloadFile, t)
	//utils.Log.Info().Msgf("申请发送文件:%v", ds.FileIndex.PullID)
	t.lock.Lock()
	t.downloadSteps[utils.Bytes2string(ds.FileIndex.PullID)] = ds
	t.lock.Unlock()
	if t.GetAutoReceive() {
		ds.Start()
	} else {
		ds.StartWait()
	}
	config.ReplySuccess(this.area, config.NET_protocol_version_v1, message, nil)
}

/*
申请发送文件 返回
*/
//func (this *Manager) applySendFile_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(FLOOD_CLASS_key, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
获取文件块数据
*/
func (this *Manager) getFileChunk(message *message_center.MessageBase) {
	//utils.Log.Info().Str("收到文件块下载信息", "").Send()
	//replyMsgID := uint64(NET_CODE_get_file_chunk_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	fileChunk, err := ParseFileChunk(np.Data)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	//utils.Log.Info().Msgf("请求文件块参数:%+v", fileChunk)
	t := this.GetClass(fileChunk.ClassID)
	if t == nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_classID_not_find, "")
		return
	}
	//验证白名单
	//ok := t.CheckWhiteList(*message.Head.Sender)
	//if ok {
	//	config.ReplyError(this.area, config.NET_protocol_version_v1, message,  config.ERROR_CODE_file_transfer_No_permission, err.Error())
	//	return
	//}

	//utils.Log.Info().Str("收到文件块下载信息", "").Send()
	ft := t.FindTransferTask(fileChunk.SupplierID)
	if ft == nil {
		utils.Log.Info().Msgf("未找到这个下载任务")
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_No_find_task, "")
		return
	}

	//utils.Log.Info().Str("收到文件块下载信息", "").Send()
	if time.Now().Unix() > ft.OverTime {
		//超时
		utils.Log.Info().Msgf("下载超时")
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_No_find_task, "")
		return
	}

	//设置已经上传的大小
	ft.OffsetIndex = fileChunk.OffsetIndex

	//fileIndex := t.GetDownloadByPullID(ft.PullID)
	//if fileIndex != nil {
	//	fileIndex.OffsetIndex = fileChunk.OffsetIndex
	//}
	//utils.Log.Info().Str("收到文件块下载信息", "").Send()
	//判断文件大小
	bs, err := readFromFile(ft.LocalFilePath, int64(fileChunk.OffsetIndex), int64(fileChunk.ChunkSize), ft.uploadFile)
	if err != nil {
		utils.Log.Info().Msgf("读取文件出错:%s", err.Error())
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}

	//utils.Log.Info().Msgf("返回文件块:%d", len(bs))
	config.ReplySuccess(this.area, config.NET_protocol_version_v1, message, &bs)
}

/*
获取文件块数据 返回
*/
//func (this *Manager) getFileChunk_recv(message *message_center.MessageBase) {
//	//utils.Log.Info().Msgf("收到返回文件块信息")
//	engine.ResponseByteKey(FLOOD_CLASS_key, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
通知传输完成
*/
func (this *Manager) noticeTransferFinish(message *message_center.MessageBase) {
	//utils.Log.Info().Msgf("传输完成 1111111")
	//replyMsgID := uint64(NET_CODE_notice_transfer_finish_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	//utils.Log.Info().Msgf("传输完成 222222")
	fileChunk, err := ParseFileChunk(np.Data)
	if err != nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	//utils.Log.Info().Msgf("传输完成 333333")
	t := this.GetClass(fileChunk.ClassID)
	if t == nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_classID_not_find, "")
		return
	}
	//utils.Log.Info().Msgf("传输完成 444444")
	//验证白名单
	//ok := t.CheckWhiteList(*message.Head.Sender)
	//if ok {
	//	config.ReplyError(this.area, config.NET_protocol_version_v1, message,  config.ERROR_CODE_file_transfer_No_permission, err.Error())
	//	return
	//}

	ft := t.FindTransferTask(fileChunk.SupplierID)
	if ft == nil {
		config.ReplyError(this.area, config.NET_protocol_version_v1, message, config.ERROR_CODE_file_transfer_No_find_task, "")
		return
	}
	atomic.StoreUint64(&ft.OffsetIndex, fileChunk.OffsetIndex)
	//utils.Log.Info().Msgf("传输完成 555555")
	config.ReplySuccess(this.area, config.NET_protocol_version_v1, message, nil)
	this.pushFinishSignalMonitor <- ft
	//utils.Log.Info().Msgf("传输完成 666666")
}

/*
通知传输完成 返回
*/
//func (this *Manager) noticeTransferFinish_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(FLOOD_CLASS_key, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

func readFromFile(filename string, offset int64, length int64, uploadFile UploadFile) ([]byte, error) {
	if uploadFile != nil {
		return uploadFile.Read(filename, offset, length)
	} else {
		file, err := os.OpenFile(filename, os.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if fileInfo.Size() < offset+length {
			length = fileInfo.Size() - offset
		}
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		data := make([]byte, length)
		_, err = io.ReadFull(file, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func writeFileChunk(filename string, offset int64, data []byte, downloadFile DownloadFile) error {
	if downloadFile != nil {
		return downloadFile.Writer(filename, data, offset)
	} else {
		// 打开文件
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
		// 指定位置写入数据（例如从第10个字节位置开始写入）
		_, err = file.WriteAt(data, offset)
		if err != nil {
			return err
		}
		return nil
	}
}
