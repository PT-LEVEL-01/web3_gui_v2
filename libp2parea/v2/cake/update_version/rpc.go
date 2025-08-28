package update_version

const (
	NameError = 4005
)

var updateVersion *UpdateVersion

func Start(uv *UpdateVersion) {
	updateVersion = uv
	//RegisterRPC()
	//routers.RegisterRpc()
}

//
//func RegisterRPC() {
//	model.RegisterErrcode(NameError, "name error")
//
//	rpc.RegisterRPC("checklatestversion", CheckLatestVersion)     //检查最新版本
//	rpc.RegisterRPC("getversionfile", GetVersionFile)             //获取版本文件
//	rpc.RegisterRPC("getplatform", GetPlatform)                   //获取平台
//	rpc.RegisterRPC("setplatform", SetPlatform)                   //设置平台
//	rpc.RegisterRPC("getrootnetaddr", GetRootNetAddr)             //获取版本库远程主机
//	rpc.RegisterRPC("setrootnetaddr", SetRootNetAddr)             //设置版本库远程主机
//	rpc.RegisterRPC("locallatestversion", LocalLatestVersion)     //获取本机最新版本
//	rpc.RegisterRPC("originversionlibrary", OriginVersionLibrary) //获取远端版本文件列表
//}
//
//func CheckLatestVersion(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	ok, _, ver, _, code, err := updateVersion.CheckLatestVersion()
//	if err != nil {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	} else {
//		res, err = model.Tojson(map[string]interface{}{"latest": ok, "latestName": ver, "version_code": code})
//	}
//	return
//}
//
//func GetVersionFile(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	fileName, ok := rj.Get("fileName")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "fileName")
//		return
//	}
//	err = updateVersion.GetVersionFile(fileName.(string))
//	if err == nil {
//		res, err = model.Tojson("success")
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func GetPlatform(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	platform := updateVersion.GetPlatform()
//	res, err = model.Tojson(platform)
//	return
//}
//
//func SetPlatform(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	platform, ok := rj.Get("platform")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "platform")
//		return
//	}
//	updateVersion.SetPlatform(platform.(string))
//	res, err = model.Tojson("success")
//	return
//}
//
//func GetRootNetAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	platform := updateVersion.GetRootNetAddr()
//	res, err = model.Tojson(platform)
//	return
//}
//
//func SetRootNetAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	addr, ok := rj.Get("addr")
//	if !ok {
//		res, err = model.Errcode(model.NoField, "addr")
//		return
//	}
//	updateVersion.SetRootNetAddr(addr.(string))
//	res, err = model.Tojson("success")
//	return
//}
//
//func LocalLatestVersion(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	fn, code, _, err := updateVersion.LocalLatestVersion()
//	if err == nil {
//		res, err = model.Tojson(map[string]interface{}{"latestName": fn, "version_code": code})
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
//
//func OriginVersionLibrary(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
//	list, err := updateVersion.OriginVersionLibrary()
//	if err == nil {
//		res, err = model.Tojson(list)
//	} else {
//		res, err = model.Errcode(model.Nomarl, err.Error())
//	}
//	return
//}
