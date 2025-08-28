package main

import (
	"crypto/sha256"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/example"
	"web3_gui/utils"
)

var (
	addrPre = "TEST" //收款地址前缀
	//areaName = sha256.Sum256([]byte("nihaoa a a!")) //域网络名称
	areaName = sha256.Sum256([]byte(example.AreaNameStr + strconv.Itoa(1)))
	keyPwd   = "123456789"   //钱包密码
	basePort = uint16(25331) //节点本地监听的端口
	//日志文件路径
	filePath_log = filepath.Join("logs", "log.txt")
	//密钥库文件路径
	filePath_keystore = filepath.Join("conf", "keystore"+strconv.Itoa(0)+".key")
	//数据库文件路径
	filePath_db = filepath.Join("db")

	rpcname_getinfo = "getinfo"
)

func main() {

	utils.LogBuildDefaultFile(filePath_log)
	//utils.Log.Info().Str("start", "11111111").Send()

	//初始化一个密钥库，并指定路径和前缀
	key1 := keystore.NewKeystoreSingle(filePath_keystore, addrPre)
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
	node.SetLeveldbPath(filePath_db)
	//utils.Log.Info().Str("start", "11111111").Send()

	//设置发现节点的地址和端口
	node.SetDiscoverPeer([]string{"/ip4/47.112.178.87/tcp/" + strconv.Itoa(int(basePort)) + "/ws"})

	//node.UpdateRpcUser(engine.RPC_username, "testp")
	node.RegisterRPC(1, rpcname_getinfo, GetInfo_rpc, "获取基本信息",
		engine.NewParamValid_Mast_Panic("age", reflect.String, "年龄"))
	//指定本地监听端口，并启动节点
	ERR = node.StartUP(basePort)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//utils.Log.Info().Str("END", "11111111").Send()
	//等待节点组网完成，当首节点无其他节点可以连接时，尝试连接所有节点失败后组网完成。
	node.WaitAutonomyFinish()
	utils.Log.Info().Int("节点组网完成", 11111).Send()
	time.Sleep(time.Second * 20)

	nodes := []*libp2parea.Node{node}
	//打印实际逻辑节点
	example.PrintActualAreaname(nodes)
	//打印各个节点的所有域节点连接
	example.PrintSessionNodeAreaname(nodes)

	select {}
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
