// go build -buildmode=c-shared -o ../dll/peer.dll dll.go
package main

import (
	"C"
	"net"
	"time"
	"web3_gui/im/boot"
)

//export TryConnHost
func TryConnHost(host *C.char) (code int) {
	// fmt.Println("conn start")
	hostStr := C.GoString(host)
	//设置1秒钟超时
	conn, err := net.DialTimeout("tcp", hostStr, time.Second*1)
	if err != nil {
		// conn.Close()
		return 1
	}
	// fmt.Println("conn ok")
	conn.Close()
	return 0
}

//export RpcPost
func RpcPost(host *C.char) (result *C.char) {
	//C.free(unsafe.Pointer(cstr))
	return C.CString("result 1111111")
}

//export Import
func Import(password, seed *C.char) (code int) {
	passwd := C.GoString(password)
	seedStr := C.GoString(seed)
	return ImportJAVA(passwd, seedStr)
}

//export ImportJAVA
func ImportJAVA(password, seed string) (code int) {
	// gconfig.StepDLL()
	// rs := kstore.Import(gconfig.KeystoreFileAbsPath, password, seed, gconfig.AddrPre)
	// di := kstore.ParseDataInfo(rs)
	// if di.Code == 500 {
	// 	code = 1
	// }
	return 1
}

//export StartUPC
func StartUPC(word, params *C.char) (code int) {
	passwd := C.GoString(word)
	paramJSON := C.GoString(params)
	return StartUPJAVA(passwd, paramJSON)
}

//export StartUPJAVA
func StartUPJAVA(passwd, params string) (code int) {
	defer func() {
		if x := recover(); x != nil {
			code = 2
			return
		}
		return
	}()
	boot.StartUP(passwd, params)
	return 0
}

func main() {
	//StartUP()
}

type Config struct {
	Init        bool   //= flag.String("init", "", "创建创始区块(默认：genesis.json)")
	Conf        string //= flag.String("conf", "conf/config.json", "指定配置文件(默认：conf/config.json)")
	Port        int    //= flag.Int("port", 9811, "本地监听端口(默认：9811)")
	NetId       int    //= flag.Int("netid", 20, "网络id(默认：20)")
	Ip          string //= flag.String("ip", "0.0.0.0", "本地IP地址(默认：0.0.0.0)")
	WebAddr     string //= flag.String("webaddr", "0.0.0.0", "本地web服务器IP地址(默认：0.0.0.0)")
	WebPort     int    //= flag.Int("webport", 2080, "web服务器端口(默认：2080)")
	WebStatic   string //= flag.String("westatic", "", "本地web静态文件目录")
	WebViews    string //= flag.String("webviews", "", "本地web Views文件目录")
	DataDir     string //= flag.String("datadir", "", "指定数据目录")
	DbCache     int    //= flag.Int("dbcache", 25, "设置数据库缓存大小，单位为兆字节（MB）（默认：25）")
	TimeOut     int    //= flag.Int("timeout", 0, "设置连接超时，单位为毫秒")
	RpcServer   bool   //= flag.Bool("rpcserver", false, "打开或关闭JSON-RPC true/false(默认：false)")
	RpcUser     string //= flag.String("rpcuser", "", "JSON-RPC 连接使用的用户名")
	RpcPassword string //= flag.String("rpcpassword", "", "JSON-RPC 连接使用的密码")
	WalletPwd   string //= flag.String("walletpwd", config.Wallet_keystore_default_pwd, "钱包密码")
	classpath   string //= flag.String("classpath", "", "jar包路径")
	Load        bool   // = flag.String("load", "", "从历史区块拉起链端")
}
