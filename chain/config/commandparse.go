package config

import (
	"flag"
	"fmt"
)

var Input_pstrName = flag.String("name", "gerry", "input ur name")
var Input_piAge = flag.Int("age", 20, "input ur age")
var Input_flagvar int

var (
	Init        = flag.String("init", "", "创建创始区块(默认：genesis.json)")
	Conf        = flag.String("conf", "conf/config.json", "指定配置文件(默认：conf/config.json)")
	Port        = flag.Int("port", 9811, "本地监听端口(默认：9811)")
	NetId       = flag.Int("netid", 20, "网络id(默认：20)")
	Ip          = flag.String("ip", "0.0.0.0", "本地IP地址(默认：0.0.0.0)")
	webAddr     = flag.String("webaddr", "0.0.0.0", "本地web服务器IP地址(默认：0.0.0.0)")
	webPort     = flag.Int("webport", 2080, "web服务器端口(默认：2080)")
	WebStatic   = flag.String("westatic", "", "本地web静态文件目录")
	WebViews    = flag.String("webviews", "", "本地web Views文件目录")
	DataDir     = flag.String("datadir", "", "指定数据目录")
	DbCache     = flag.Int("dbcache", 25, "设置数据库缓存大小，单位为兆字节（MB）（默认：25）")
	TimeOut     = flag.Int("timeout", 0, "设置连接超时，单位为毫秒")
	rpcServer   = flag.Bool("rpcserver", false, "打开或关闭JSON-RPC true/false(默认：false)")
	RpcUser     = flag.String("rpcuser", "", "JSON-RPC 连接使用的用户名")
	RpcPassword = flag.String("rpcpassword", "", "JSON-RPC 连接使用的密码")
	//WalletPwd   = flag.String("walletpwd", Wallet_keystore_default_pwd, "钱包密码")
	classpath = flag.String("classpath", "", "jar包路径")
	Load      = flag.String("load", "", "从历史区块拉起链端")
)

func Step() {
	// engine.Log.Info("开始解析参数")
	// parseConfig()
	flag.Parse()
	parseParam()
	ParseConfig()
	ParseConfigExtra()
	// parseNonFlag()
}

func StepV2(path string) {

	path = fmt.Sprintf("%s\\%s", path, Path_config)

	// engine.Log.Info("开始解析参数")
	// parseConfig()
	flag.Parse()
	parseParam()
	ParseConfigV2(path)
	// parseNonFlag()
}

func StepDLL() {
	// engine.Log.Info("开始解析参数")
	// parseConfig()
	//同步调用dll时，flag.Parse()这句代码是阻塞状态，不能执行。
	// flag.Parse()
	parseParam()
	ParseConfig()
	// parseNonFlag()
}

//	func parseNonFlag() {
//		for _, param := range flag.Args() {
//			switch param {
//			case "init":
//				// InitNode = true
//				// startblock.BuildFirstBlock()
//			case "testnet":
//				// fmt.Println("testnet")
//			}
//		}
//	}
func parseParam() {
	flag.VisitAll(func(v *flag.Flag) {
		switch v.Name {
		case "port":
			// gconfig.Init_LocalPort = uint16(*Port)
			// gconfig.Init_GatewayPort = gconfig.Init_LocalPort
		case "netid":
			// engine.Netid = uint32(*NetId)
		case "webaddr":
			// WebAddr = *WebAddr
		case "webport":
			// config.WebPort = uint16(*WebPort)
		case "westatic":
			// config.Web_path_static = *WebStatic
		case "webviews":
			// config.Web_path_views = *WebViews
		case "datadir":
			//datadir
		case "dbcache":
			//dbcache
		case "timeout":
			//timeout
		case "rpcserver":
			// engine.Log.Info("rpcserver:%t", rpcServer)
			// rpc.Server = *RpcServer
		case "rpcuser":
			// rpc.User = *RpcUser
		case "rpcpassword":
			// rpc.Password = *RpcPassword
		case "walletpwd":
			//	Wallet_keystore_default_pwd = *WalletPwd
			//rpc
		}
	})

}

// func parseConfig() {
// 	if !exists(*Conf) {
// 		return
// 	}
// 	confpath := flag.Lookup("conf").Value.String()
// 	bs, err := ioutil.ReadFile(confpath)
// 	if err != nil {
// 		panic("Read conf error: " + err.Error())
// 		return
// 	}
// 	cfi := new(Config)
// 	// err = json.Unmarshal(bs, cfi)
// 	decoder := json.NewDecoder(bytes.NewBuffer(bs))
// 	decoder.UseNumber()
// 	err = decoder.Decode(cfi)

// 	if err != nil {
// 		panic("Parse conf error: " + err.Error())
// 		return
// 	}
// 	*Port = int(cfi.Port)
// 	*NetId = int(cfi.Netid)
// 	*Ip = cfi.IP
// 	WebAddr = cfi.WebAddr
// 	WebPort = cfi.WebPort
// 	*WebStatic = cfi.WebStatic
// 	*WebViews = cfi.WebViews
// 	*RpcServer = cfi.RpcServer
// 	*RpcUser = cfi.RpcUser
// 	*RpcPassword = cfi.RpcPassword

// }
// func exists(path string) bool {
// 	_, err := os.Stat(path)
// 	if err == nil {
// 		return true
// 	}
// 	if os.IsNotExist(err) {
// 		return false
// 	}
// 	return true
// }
