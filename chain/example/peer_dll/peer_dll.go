// go build -buildmode=c-shared -o ../dll/peer.dll dll.go
package main

import (
	"C"
	jsoniter "github.com/json-iterator/go"
	"net"
	"time"
	"web3_gui/chain/boot"
	"web3_gui/chain/config"
	chain "web3_gui/chain/sdk"
	"web3_gui/keystore/adapter"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
	// config.StepDLL()
	config.ParseConfig()
	k := keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre)
	// engine.Log.Info("密钥路径:%s %s", config.KeystoreFileAbsPath, config.AddrPre)
	err := k.ImportMnemonic(seed, password, password, password)
	// err := config.Area.Keystore.ImportMnemonic(seed, password, password, password)
	if err != nil {
		code = 1
		return
	}
	/*rs := kstore.Import(config.KeystoreFileAbsPath, password, seed, config.AddrPre)
	di := kstore.ParseDataInfo(rs)
	if di.Code == 500 {
		code = 1
	}*/
	return
}

//export StartUPC
func StartUPC(word, params *C.char) (code int) {
	passwd := C.GoString(word)
	paramJSON := C.GoString(params)

	c := new(Config)
	err := json.Unmarshal([]byte(paramJSON), c)
	if err != nil {
		return 1
	}
	// config.Step()
	config.LoadNode = c.Load
	config.InitNode = c.Init

	// fmt.Println(gconfig.LoadNode, gconfig.InitNode)

	return StartUPJAVA(passwd)
}

//export StartUPJAVA
func StartUPJAVA(passwd string) (code int) {
	defer func() {
		if x := recover(); x != nil {
			code = 2
			return
		}
		return
	}()
	// gconfig.Wallet_keystore_default_pwd = passwd
	//print("*****start****", msg, num, b, "\n")
	//	ip := beego.AppConfig.DefaultString("ip", "0.0.0.0")
	//	portStr := beego.AppConfig.DefaultString("port", "0")
	boot.StartWithArea(passwd)
	return 0
}

//export CreateOfflineTx
func CreateOfflineTx(keyStorePath *C.char, srcaddress *C.char, address *C.char, pwd *C.char, comment *C.char, amount *C.char, gas *C.char, frozenHeight *C.char, nonce *C.char, currentHeight *C.char, domain *C.char, domainType *C.char) *C.char {
	return C.CString(chain.CreateOfflineTx(C.GoString(keyStorePath), C.GoString(srcaddress), C.GoString(address), C.GoString(pwd), C.GoString(comment), C.GoString(amount), C.GoString(gas), C.GoString(frozenHeight), C.GoString(nonce), C.GoString(currentHeight), C.GoString(domain), C.GoString(domainType)))
}

//export CreateOfflineContractTx
func CreateOfflineContractTx(keyStorePath *C.char, srcaddress *C.char, address *C.char, pwd *C.char, comment *C.char, amount *C.char, gas *C.char, frozenHeight *C.char, gasPrice *C.char, nonce *C.char, currentHeight *C.char, domain *C.char, domainType *C.char, abi *C.char, source *C.char) *C.char {
	return C.CString(chain.CreateOfflineContractTx(C.GoString(keyStorePath), C.GoString(srcaddress), C.GoString(address), C.GoString(pwd), C.GoString(comment), C.GoString(amount), C.GoString(gas), C.GoString(frozenHeight), C.GoString(gasPrice), C.GoString(nonce), C.GoString(currentHeight), C.GoString(domain), C.GoString(domainType), C.GoString(abi), C.GoString(source)))
}

//export GetComment
func GetComment(tag *C.char, jsonData *C.char) *C.char {
	return C.CString(chain.GetComment(C.GoString(tag), C.GoString(jsonData)))
}

//export MultDeal
func MultDeal(tag *C.char, jsonData *C.char, keyStorePath *C.char, srcaddress *C.char, address *C.char, pwd *C.char, comment *C.char, amount C.ulonglong, gas C.ulonglong, frozenHeight C.ulonglong, gasPrice C.ulonglong, nonce C.ulonglong, currentHeight C.ulonglong, domain *C.char, domainType C.ulonglong) *C.char {
	return C.CString(chain.MultDeal(C.GoString(tag), C.GoString(jsonData), C.GoString(jsonData), C.GoString(srcaddress), C.GoString(address), C.GoString(pwd), C.GoString(comment), uint64(amount), uint64(gas), uint64(frozenHeight), uint64(gasPrice), uint64(nonce), uint64(currentHeight), C.GoString(domain), uint64(domainType)))
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

// //export RemoveKey
// func RemoveKey(conf *C.char) *C.char {
// 	resultStr := ""
// 	confParam := C.GoString(conf)
// 	// err := os.RemoveAll(config.)
// 	if err != nil {
// 		resultStr = Out(OutStatusFail, OutStatusFailText)
// 	} else {
// 		resultStr = Out(OutStatusOK, OutStatusOkText)
// 	}
// 	return C.CString(resultStr)
// }

// const (
// 	OutStatusOK       int    = 200    //返回成功状态
// 	OutStatusOkText   string = "ok"   //成功反回说明
// 	OutStatusFail     int    = 500    //返回失败状态
// 	OutStatusFailText string = "fail" //失败反回说明
// )

// //返回数据格式
// type DataInfo struct {
// 	Code int         `json:"code"`
// 	Data interface{} `json:"data"`
// }

// func (d *DataInfo) Json() string {
// 	rs, _ := json.Marshal(d)
// 	return string(rs)
// }
// func ParseDataInfo(bs string) DataInfo {
// 	di := DataInfo{}
// 	err := json.Unmarshal([]byte(bs), &di)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return di
// }
// func Out(code int, data interface{}) string {
// 	d := DataInfo{Code: code, Data: data}
// 	return d.Json()
// }
