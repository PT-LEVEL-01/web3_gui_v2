package rpc

import (
	"net/http"
	"sync"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

type ServerHandler func(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) ([]byte, error)

//访问接口，统一header里传user,password
// var rpcHandler = map[string]ServerHandler{
// 	"stopservice": StopService, //关闭服务器
// }

var rpcHandler = new(sync.Map) //

func RegisterRPC(rpcName string, handler ServerHandler) {
	_, ok := rpcHandler.LoadOrStore(rpcName, handler)
	if ok {
		utils.Log.Info().Msgf("register rpc name exist:%s", rpcName)
	}
}

// 获取RPC的Handlers
func GetRPCHandlers() *sync.Map {
	return rpcHandler
}

// 移除已注册的handler
// 缺省移除全部，有参数则移除指定handler
func RemoveRegisterRPC(rpcNames ...string) {
	if len(rpcNames) == 0 {
		rpcHandler = new(sync.Map)
		return
	}

	for _, rpcName := range rpcNames {
		rpcHandler.Delete(rpcName)
	}
}

/*
关闭服务
*/
func StopService(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	utils.StopService()
	res, err = model.Tojson("success")
	return
}
