package storage

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
)

func RegisterHandlers() {
	Node.Register_p2pHE(config.MSGID_STORAGE_multicast_nodeinfo_recv, MulticastStorageServer_recv) //提供离线消息存储节点广播消息 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_getorders, GetOrders)                                 //客户端向存储提供商获取一个租用空间的订单
	//Node.Register_p2pHE(config.MSGID_STORAGE_getorders_recv, GetOrders_recv)                       //客户端向存储提供商获取一个租用空间的订单 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_getRenewalOrders, GetRenewalOrders) //获取续费订单
	//Node.Register_p2pHE(config.MSGID_STORAGE_getRenewalOrders_recv, GetRenewalOrders_recv) //获取续费订单 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_getFreeSpace, GetFreeSpace_net) //获取可用空间
	//Node.Register_p2pHE(config.MSGID_STORAGE_getFreeSpace_recv, GetFreeSpace_net_recv)     //获取可用空间 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_getFileList, GetRemoteFileList) //获取文件列表
	//Node.Register_p2pHE(config.MSGID_STORAGE_getFileList_recv, GetRemoteFileList_recv)     //获取文件列表 返回

	Node.Register_p2pHE(config.MSGID_STORAGE_upload_fileindex, UploadFileIndex) //客户端向存储提供商发送要上传的文件索引
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_fileindex_recv, UploadFileIndex_recv) //客户端向存储提供商发送要上传的文件索引 返回
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_transfer_finish, UploadTransferFinish)           //服务端告诉客户端传输完成
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_transfer_finish_recv, UploadTransferFinish_recv) //服务端告诉客户端传输完成 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_savedb_finish, UploadSavedbFinish) //服务端告诉客户端保存数据库完成
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_savedb_finish_recv, UploadSavedbFinish_recv)                //服务端告诉客户端保存数据库完成 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_stop, UploadStop) //客户端暂停上传
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_stop_recv, UploadStop_recv)                                 //客户端暂停上传 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_reset, UploadReset) //客户端暂停上传后恢复
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_reset_recv, UploadReset_recv)                               //客户端暂停上传后恢复 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_delete_dir_and_file, DeleteDirAndFiles) //客户端删除多个文件夹和文件
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_delete_dir_and_file_recv, DeleteDirAndFiles_recv)           //客户端删除多个文件夹和文件 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_create_dir, CreateDir) //创建目录
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_create_dir_recv, CreateDir_recv)                            //创建目录 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_download_select_dir_fileindex, SelectDirWithinFileindex) //递归查询多个文件夹中的文件列表
	//Node.Register_p2pHE(config.MSGID_STORAGE_download_select_dir_fileindex_recv, SelectDirWithinFileindex_recv) //递归查询多个文件夹中的文件列表 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_download_apply, ApplyDownload) //申请下载文件
	//Node.Register_p2pHE(config.MSGID_STORAGE_download_apply_recv, ApplyDownload_recv)                           //申请下载文件 返回
	Node.Register_p2pHE(config.MSGID_STORAGE_upload_update_name, UpdateFileAndDirName) //客户端修改文件和文件夹名称
	//Node.Register_p2pHE(config.MSGID_STORAGE_upload_update_name_recv, UpdateFileAndDirName_recv)                //客户端修改文件和文件夹名称 返回

}

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
	nr := utils.NewNetResult(version, config.ERROR_CODE_success, "", resultBs)
	bs, err := nr.Proto()
	if err != nil {
		area.SendP2pReplyMsgHE(message, nil)
		return
	}
	area.SendP2pReplyMsgHE(message, bs)
}

/*
接收云存储空间提供者广播 接收广播
*/
func MulticastStorageServer_recv(message *message_center.MessageBase) {
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收云存储空间提供者广播，解析参数出错:%s", err.Error())
		return
	}
	sinfo, err := model.ParseStorageServerInfo(np.Data)
	if err != nil {
		return
	}
	sinfo.Addr = *message.SenderAddr
	AddStorageServerList(sinfo)
}

/*
客户端向存储提供商获取一个租用空间的订单
*/
func GetOrders(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_getorders_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, err := model.ParseOrderForm(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	utils.Log.Info().Msgf("本次购买空间:%+v %+v", of, np.Data)
	of, ERR := StServer.CreateOrders(of.UserAddr, utils.Byte(of.SpaceTotal), of.UseTime)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := of.Proto()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
客户端向存储提供商获取一个租用空间的订单 返回
*/
//func GetOrders_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端向存储提供商获取续费订单
*/
func GetRenewalOrders(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_getRenewalOrders_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, err := model.ParseOrderForm(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, ERR := StServer.CreateRenewalOrders(of.PreNumber, of.SpaceTotal, of.UseTime)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := of.Proto()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
客户端向存储提供商获取续费订单 返回
*/
//func GetRenewalOrders_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
获取用户可用空间
*/
func GetFreeSpace_net(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_getFreeSpace_recv)
	_, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	spacesTotal, spacesUse, _ := StServer.AuthManager.QueryUserSpaces(*message.SenderAddr)
	bs := utils.Uint64ToBytes(uint64(spacesTotal))
	bs = append(bs, utils.Uint64ToBytes(uint64(spacesUse))...)
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
获取用户可用空间 返回
*/
//func GetFreeSpace_net_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端向存储提供商获取续费订单
*/
func GetRemoteFileList(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_getFileList_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	var dir *model.DirectoryIndex
	var ERR utils.ERROR
	dirPath := np.Data
	if dirPath == nil || len(dirPath) == 0 {
		//传空则是查询最顶层文件目录
		dir, ERR = db.StorageServer_GetUserTopDirIndex(*message.SenderAddr)
	} else {
		dir, ERR = db.StorageServer_GetDirIndex(dirPath)
	}
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	//utils.Log.Info().Msgf("查询的文件夹:%+v", dir)
	//查询文件夹中的子文件夹
	dirs, ERR := db.StorageServer_GetDirIndexMore(dir.DirsID...)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	//utils.Log.Info().Msgf("查询的文件夹列表:%+v", dirs)
	dir.Dirs = dirs
	//查询文件夹中的子文件
	files, ERR := db.StorageServer_GetFileIndexMore(dir.FilesID...)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	for _, one := range files {
		one.FilterUser(*message.SenderAddr)
	}
	//utils.Log.Info().Msgf("查询的文件列表:%+v", files)
	dir.Files = files
	bs, err := dir.ProtoDirAndFile()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
客户端向存储提供商获取续费订单 返回
*/
//func GetRemoteFileList_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
上传文件索引
*/
func UploadFileIndex(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_fileindex_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	fileindex, err := model.ParseFileIndex(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	fileindex.UserAddr = [][]byte{message.SenderAddr.GetAddr()}
	ERR := StServer.UploadAddFileIndex(*message.SenderAddr, *fileindex)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
上传文件索引 返回
*/
//func UploadFileIndex_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
服务端告诉客户端传输完成
*/
//func UploadTransferFinish(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	StClient.UpdateFileUploadFinish(*message.Head.Sender, *message.Body.Content)
//	Node.SendP2pReplyMsgHE(message, config.MSGID_STORAGE_upload_transfer_finish_recv, nil)
//}

/*
服务端告诉客户端传输完成 返回
*/
//func UploadTransferFinish_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
服务端告诉客户端保存数据库完成
*/
func UploadSavedbFinish(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_savedb_finish_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	ERR, err := utils.ParseERROR(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	if !ERR.CheckSuccess() {
		ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
		return
	}

	//nr, err := model.ParseNetResult(np.Data)
	//if err != nil {
	//	ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
	//	return
	//}
	err = StClient.UpdateFileSavedbFinish(*message.SenderAddr, []byte(ERR.Msg))
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
服务端告诉客户端保存数据库完成 返回
*/
//func UploadSavedbFinish_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端暂停上传
*/
func UploadStop(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_stop_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	err = StServer.UploadStop(*message.SenderAddr, np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
客户端暂停上传 返回
*/
//func UploadStop_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端暂停上传后恢复
*/
func UploadReset(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_reset_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	err = StServer.UploadReset(*message.SenderAddr, np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
客户端暂停上传后恢复 返回
*/
//func UploadReset_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端删除多个文件夹和文件
*/
func DeleteDirAndFiles(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_delete_dir_and_file_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	bss := new(go_protos.Bytes)
	err = proto.Unmarshal(np.Data, bss)
	if err != nil {
		//参数无法解析
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	ERR := StServer.DelDirAndFile(*message.SenderAddr, bss.List, bss.List2)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
客户端删除多个文件夹和文件 返回
*/
//func DeleteDirAndFiles_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
创建目录
*/
func CreateDir(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_create_dir_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	dirIndex, err := model.ParseDirectoryIndex(np.Data)
	dirIndex.UAddr = *message.SenderAddr
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	utils.Log.Info().Msgf("收到创建文件夹消息:%+v", dirIndex)
	ERR := StServer.CreateDir(*dirIndex)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建文件夹错误:%s", err.Error())
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		//Node.SendP2pReplyMsgHE(message, config.MSGID_STORAGE_upload_create_dir_recv, nil)
		return
	}
	utils.Log.Info().Msgf("创建文件夹成功")
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
	//Node.SendP2pReplyMsgHE(message, config.MSGID_STORAGE_upload_create_dir_recv, nil)
}

/*
创建目录 返回
*/
//func CreateDir_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
递归查询多个文件夹中的文件列表
*/
func SelectDirWithinFileindex(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_download_select_dir_fileindex_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	bss := new(go_protos.Bytes)
	err = proto.Unmarshal(np.Data, bss)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		//Node.SendP2pReplyMsgHE(message, config.MSGID_STORAGE_download_select_dir_fileindex_recv, nil)
		return
	}

	dirIndexAll, _, _, _, ERR := db.StorageServer_GetDirIndexRecursion(bss.List)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}

	bss.List = make([][]byte, 0)
	for _, one := range dirIndexAll {
		bsOne, err := one.ProtoDirAndFile()
		if err != nil {
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
			return
		}
		bss.List = append(bss.List, *bsOne)
	}
	bs, err := bss.Marshal()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
递归查询多个文件夹中的文件列表 返回
*/
//func SelectDirWithinFileindex_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
请求下载文件
*/
func ApplyDownload(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_download_apply_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	fileIndex, err := model.ParseFileIndex(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	err = StServer.DownloadFile(*message.SenderAddr, fileIndex)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
请求下载文件 返回
*/
//func ApplyDownload_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端修改文件和文件夹名称
*/
func UpdateFileAndDirName(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_STORAGE_upload_update_name_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	updateInfo, err := model.ParseDirectoryIndex(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	var ERR utils.ERROR
	if len(updateInfo.ID) > 0 {
		//修改文件
		ERR = StServer.UpdateFileName(*message.SenderAddr, updateInfo.ID, updateInfo.Name)
	} else {
		//修改文件夹
		ERR = StServer.UpdateDirName(*message.SenderAddr, updateInfo.ParentID, updateInfo.Name)
	}
	if ERR.CheckFail() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
客户端修改文件和文件夹名称 返回
*/
//func UpdateFileAndDirName_recv(message *message_center.MessageBase) {
//	engine.ResponseByteKey(config.FLOOD_key_file_transfer, message.SendID, &message.Content)
//	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
