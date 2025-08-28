package main

import (
	"crypto/sha256"
	json2 "encoding/json"
	"path/filepath"
	"reflect"
	"strconv"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

var (
	addrPre      = "TEST"                          //收款地址前缀
	areaName     = sha256.Sum256([]byte("nihaoa")) //域网络名称
	keyPwd       = "123456789"                     //钱包密码
	basePort     = uint16(19960)                   //节点本地监听的端口
	dirRoot      = "D:/test/temp"                  //
	rpc_username = engine.RPC_username
	rpc_password = engine.RPC_password
)

const (
	rpcname_getinfo = "getinfo"
	rpcname_print18 = "print18"
)

func main() {
	StartServer()
	StartClient()
}

/*
启动一个节点，作为rpc服务器
*/
func StartServer() {
	//日志文件路径
	logPath := filepath.Join(dirRoot, "log", "log0.txt")
	utils.LogBuildDefaultFile(logPath)
	//utils.Log.Info().Str("start", "11111111").Send()
	//密钥库文件路径
	keyPath := filepath.Join(dirRoot, "keystore", "keystore0.key")
	//数据库文件路径
	dbPath := filepath.Join(dirRoot, "db", "db0")
	//初始化一个密钥库，并指定路径和前缀
	key1 := keystore.NewKeystoreSingle(keyPath, addrPre)
	//加载路径中的密钥库
	ERR := key1.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == keystore.ERROR_code_wallet_file_not_exist {
			//创建一个新的密钥库
			ERR = key1.CreateRand(keyPwd, keyPwd, keyPwd, keyPwd)
		}
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建key1错误", ERR.String()).Send()
			return
		}
	}
	//utils.Log.Info().Str("start", "11111111").Send()
	//创建一个域节点
	node, ERR := libp2parea.NewNode(areaName, addrPre, key1, keyPwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("创建area错误", ERR.String()).Send()
		return
	}
	//设置数据库路径
	node.SetLeveldbPath(dbPath)

	node.UpdateRpcUser(rpc_username, rpc_password)
	node.RegisterRPC(1, rpcname_getinfo, GetInfo_rpc, "")
	node.RegisterRPC(2, rpcname_print18, PrintAge18_rpc, "",
		engine.NewParamValid_UnsupportedTypePanic(true, "age", reflect.String, false, nil, ""))

	//utils.Log.Info().Str("start", "11111111").Send()
	//指定本地监听端口，并启动节点
	ERR = node.StartUP(basePort)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//utils.Log.Info().Str("END", "11111111").Send()
	//等待节点组网完成，当首节点无其他节点可以连接时，尝试连接所有节点失败后组网完成。
	node.WaitAutonomyFinish()
	//utils.Log.Info().Int("节点组网完成", 11111).Send()
}

/*
启动一个节点，作为rpc客户端
*/
func StartClient() {
	//先获取接口列表
	postResult, err := engine.Post("127.0.0.1:"+strconv.Itoa(int(basePort)), rpc_username, rpc_password, engine.RPC_method_rpclist, nil)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	ERR := postResult.ConverERR()
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return
	}
	listItr := postResult.Data["list"]
	//data := dataItr.(string)
	utils.Log.Info().Interface("返回值", listItr).Send()

	//请求getinfo接口
	params := make(map[string]interface{})
	params["key"] = "hello"
	postResult, err = engine.Post("127.0.0.1:"+strconv.Itoa(int(basePort)), rpc_username, rpc_password, rpcname_getinfo, params)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	ERR = postResult.ConverERR()
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return
	}
	dataItr := postResult.Data["data"]
	data := dataItr.(string)
	utils.Log.Info().Str("返回值", data).Send()

	//请求print18接口
	params = make(map[string]interface{})
	params["age"] = 18
	postResult, err = engine.Post("127.0.0.1:"+strconv.Itoa(int(basePort)), rpc_username, rpc_password, rpcname_print18, params)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	ERR = postResult.ConverERR()
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return
	}
	dataItr = postResult.Data["data"]
	data = dataItr.(string)
	utils.Log.Info().Str("返回值", data).Send()
}

func GetInfo_rpc(params map[string]interface{}) *engine.PostResult {
	pr := engine.NewPostResult()
	itr, ok := params["key"]
	if !ok {
		pr.Code = 10000001
		return pr
	}
	value := itr.(string)
	utils.Log.Info().Str("参数", value).Send()
	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = "hello"
	return pr
}

func PrintAge18_rpc(params map[string]interface{}) *engine.PostResult {
	pr := engine.NewPostResult()
	ageItr := params["age"]
	ageStr := ageItr.(string)
	age, err := json2.Number(ageStr).Int64()
	if err != nil {
		pr.Code = 11111
		pr.Msg = "age"
		return pr
	}
	utils.Log.Info().Int64("年龄", age).Send()
	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = "hello"
	return pr
}
