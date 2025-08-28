package module_a

import (
	"net/http"

	"web3_gui/libp2parea/v1"
	rpc "web3_gui/libp2parea/v1/sdk/jsonrpc2"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
)

const (
	NameError = 4005
)

func Start(area *libp2parea.Area) {
	RegisterRPC()
}

func RegisterRPC() {
	model.RegisterErrcode(NameError, "name error")

	rpc.RegisterRPC("getsuccess", GetSuccess)
	rpc.RegisterRPC("getinfo", GetInfo)
	rpc.RegisterRPC("geterror", GetError)
}

/*
获取成功
*/
func GetSuccess(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	res, err = model.Tojson("success")
	return
}

/*
获取信息
*/
func GetInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	outMap := make(map[string]interface{})
	outMap["info"] = "info"
	outMap["code"] = 0
	res, err = model.Tojson(outMap)
	return
}

/*
获取错误
*/
func GetError(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	res, err = model.Errcode(NameError, "name")
	res, err = model.Errcode(model.TypeWrong, "name")
	return
}
