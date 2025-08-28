package transfer_manager

const (
	NameError = 4005
)

var TransferMangerStatic *TransferManger

func Start(transferManger *TransferManger) {
	TransferMangerStatic = transferManger
	//RegisterRPC()
	//routers.RegisterRpc()
}

//
//func RegisterRPC() {
//	model.RegisterErrcode(NameError, "name error")
//
//	rpc.RegisterRPC("transfernewpushtask", TransferNewPushTask)   //发送文件给一个节点
//	rpc.RegisterRPC("transferpushtasklist", TransferPushTaskList) //推送任务列表
//	rpc.RegisterRPC("transfergetpushtask", GetPushTaskByTaskId)   //获取一个推送任务
//
//	rpc.RegisterRPC("transferpushtasksharingdirs", TransferPushTaskSharingDirs)       //共享目录列表
//	rpc.RegisterRPC("transferpushtasksharingdirsadd", TransferPushTaskSharingDirsAdd) //共享目录添加
//	rpc.RegisterRPC("transferpushtasksharingdirsdel", TransferPushTaskSharingDirsDel) //共享目录删除
//	rpc.RegisterRPC("transferpulladdrwhitelist", TransferPullAddrWhiteList)           //授权白名单地址
//	rpc.RegisterRPC("transferpulladdrwhitelistadd", TransferPullAddrWhiteListAdd)     //授权白名单地址添加
//	rpc.RegisterRPC("transferpulladdrwhitelistdel", TransferPullAddrWhiteListDel)     //授权白名单地址删除
//
//	rpc.RegisterRPC("transfernewpulltask", TransferNewPullTask)             //向一个节点发起拉取文件任务
//	rpc.RegisterRPC("transferpulltaskisautoset", TransferPullTaskIsAutoSet) //设置是否自动拉取
//	rpc.RegisterRPC("transferpulltaskisautoget", TransferPullTaskIsAutoGet) //获取是否自动拉取状态
//	rpc.RegisterRPC("transferfilepulltasklist", TransferFilePullTaskList)   //获取拉取文件任务列表
//	rpc.RegisterRPC("transfergetpulltask", GetPullTaskByTaskId)             //获取一个拉取任务
//	rpc.RegisterRPC("transferfilepulltaskstop", TransferFilePullTaskStop)   //拉取文件任务停止
//	rpc.RegisterRPC("transferfilepulltaskstart", TransferFilePullTaskStart) //拉取文件任务开启
//	rpc.RegisterRPC("transferfilepulltaskdel", TransferFilePullTaskDel)     //拉取文件任务删除
//}
//
//func TransferNewPushTask(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	path, ok := rj.Get("path")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "path")
//		return
//	}
//
//	toAddressNet, ok := rj.Get("toAddressNet")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "toAddressNet")
//		return
//	}
//
//	_, err = TransferMangerStatic.NewPushTask(path.(string), nodeStore.AddressFromB58String(toAddressNet.(string)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	} else {
//		res, err = model.Tojson("success")
//	}
//
//	return
//}
//
//func TransferPushTaskList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	list := TransferMangerStatic.PushTaskList()
//	if err == nil {
//		res, err = model.Tojson(list)
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func GetPushTaskByTaskId(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	taskId, ok := rj.Get("taskId")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "taskId")
//		return
//	}
//
//	data, err := TransferMangerStatic.GetPushTaskByTaskId(uint64(taskId.(float64)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, "获取任务失败:"+err.Error())
//	} else {
//		res, err = model.Tojson(data)
//	}
//
//	return
//}
//
//func TransferPushTaskSharingDirs(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	list, err := TransferMangerStatic.TransferPushTaskSharingDirs()
//	if err == nil {
//		res, err = model.Tojson(list)
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPushTaskSharingDirsAdd(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	dir, ok := rj.Get("dir")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "dir")
//		return
//	}
//	err = TransferMangerStatic.TransferPushTaskSharingDirsAdd(dir.(string))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPushTaskSharingDirsDel(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	dir, ok := rj.Get("dir")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "dir")
//		return
//	}
//	err = TransferMangerStatic.TransferPushTaskSharingDirsDel(dir.(string))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPullAddrWhiteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	list, err := TransferMangerStatic.TransferPullAddrWhiteList()
//	if err == nil {
//		res, err = model.Tojson(list)
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPullTaskIsAutoSet(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	auto, ok := rj.Get("auto")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "auto")
//		return
//	}
//
//	err = TransferMangerStatic.PullTaskIsAutoSet(auto.(bool))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPullTaskIsAutoGet(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	autoB, err := TransferMangerStatic.PullTaskIsAutoGet()
//	if err == nil {
//		res, err = model.Tojson(autoB)
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferPullAddrWhiteListDel(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	addr, ok := rj.Get("addr")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "addr")
//		return
//	}
//
//	err = TransferMangerStatic.TransferPullAddrWhiteListDel(nodeStore.AddressFromB58String(addr.(string)))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferNewPullTask(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	source, ok := rj.Get("source")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "source")
//		return
//	}
//
//	path, ok := rj.Get("path")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "path")
//		return
//	}
//
//	fromAddressNet, ok := rj.Get("fromAddressNet")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "fromAddressNet")
//		return
//	}
//
//	err = TransferMangerStatic.NewPullTask(source.(string), path.(string), nodeStore.AddressFromB58String(fromAddressNet.(string)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	} else {
//		res, err = model.Tojson("success")
//	}
//
//	return
//}
//
//func TransferPullAddrWhiteListAdd(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	addr, ok := rj.Get("addr")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "addr")
//		return
//	}
//
//	err = TransferMangerStatic.TransferPullAddrWhiteListAdd(nodeStore.AddressFromB58String(addr.(string)))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func TransferFilePullTaskList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	list := TransferMangerStatic.PullTaskList()
//	res, err = model.Tojson(list)
//	return
//}
//
//func GetPullTaskByTaskId(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	taskId, ok := rj.Get("taskId")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "taskId")
//		return
//	}
//
//	data, err := TransferMangerStatic.GetPullTaskByTaskId(uint64(taskId.(float64)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, "获取任务失败:"+err.Error())
//	} else {
//		res, err = model.Tojson(data)
//	}
//
//	return
//}
//
//func TransferFilePullTaskStop(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	taskId, ok := rj.Get("taskId")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "taskId")
//		return
//	}
//
//	err = TransferMangerStatic.PullTaskStop(uint64(taskId.(float64)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, "关闭任务失败:"+err.Error())
//	} else {
//		res, err = model.Tojson("success")
//	}
//
//	return
//}
//
//func TransferFilePullTaskStart(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	taskId, ok := rj.Get("taskId")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "taskId")
//		return
//	}
//
//	//path 可选参数
//	pathStr := ""
//	path, ok := rj.Get("path")
//	if ok {
//		pathStr = path.(string)
//	}
//
//	err = TransferMangerStatic.PullTaskStart(uint64(taskId.(float64)), pathStr)
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, "开启任务失败"+err.Error())
//	} else {
//		res, err = model.Tojson("success")
//	}
//
//	return
//}
//
//func TransferFilePullTaskDel(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	taskId, ok := rj.Get("taskId")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "taskId")
//		return
//	}
//
//	err = TransferMangerStatic.PullTaskDel(uint64(taskId.(float64)))
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, "删除任务失败"+err.Error())
//	} else {
//		res, err = model.Tojson("success")
//	}
//
//	return
//}
