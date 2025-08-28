package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"web3_gui/utils"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	Path_config         = "config.json"                                //配置文件名称
	Core_keystore       = "keystore.key"                               //密钥文件
	Path_configDir      = "conf"                                       //配置文件存放目录
	KeystoreFileAbsPath = filepath.Join(Path_configDir, Core_keystore) //密钥文件存放地址

	AddrPre               = ""                                       //收款地址前缀
	AreaName              = sha256.Sum256([]byte("icom_chain_test")) //名称不同节点无法连接
	Init                  = false                                    //是否创世节点
	Init_LocalIP          = ""                                       //rpc 服务ip
	Init_LocalPort uint16 = 19981                                    // rpc 服务端口

	WebAddr                = "0.0.0.0"
	WebPort         uint16 = 2080
	Web_path_static        = "./static"
	Web_path_views         = "./view"
	RpcServer              = false
	RPCUser                = "test"
	RPCPassword            = "testp"

	NetType_release = "release" //正式网络
	NetType         = "test"    //网络类型:正式网络release/测试网络test

	MachineId = "" // 客户端机器Id
)

// TODO 解析本地配置文件,根据不同的模块需求自定义配置
func Step() {
	ParseConfig()
}

func ParseConfig() {
	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		panic("检查配置文件错误：" + err.Error())
	}

	if !ok {
		panic("检查配置文件错误")
	}

	bs, err := ioutil.ReadFile(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		panic("读取配置文件错误：" + err.Error())
	}

	cfi := new(Config)

	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(cfi)
	if err != nil {
		panic("解析配置文件错误：" + err.Error())
	}

	Init_LocalIP = cfi.IP
	Init_LocalPort = cfi.Port
	Web_path_static = cfi.WebStatic
	Web_path_views = cfi.WebViews
	// engine. Netid = cfi.Netid
	WebAddr = cfi.WebAddr
	WebPort = cfi.WebPort
	NetType = cfi.NetType
	//	NetType = NetType_release //正式版所有节点都为release
	AddrPre = cfi.AddrPre
	RPCPassword = cfi.RpcPassword
	RpcServer = cfi.RpcServer
	MachineId = cfi.MachineID
}

type Config struct {
	// Netid       uint32 `json:"netid"`       //
	IP          string `json:"ip"`          //ip地址
	Port        uint16 `json:"port"`        //监听端口
	WebAddr     string `json:"WebAddr"`     //
	WebPort     uint16 `json:"WebPort"`     //
	WebStatic   string `json:"WebStatic"`   //
	WebViews    string `json:"WebViews"`    //
	RpcServer   bool   `json:"RpcServer"`   //
	RpcUser     string `json:"RpcUser"`     //
	RpcPassword string `json:"RpcPassword"` //
	NetType     string `json:"NetType"`     //正式网络release/测试网络test
	AddrPre     string `json:"AddrPre"`     //收款地址前缀
	MachineID   string `json:"MachineId"`   //设备机器码
}

func SetLibp2pareaConfig() string {
	// 在此添加需要传递给libp2parea的参数
	b, _ := json.Marshal(map[string]interface{}{
		"port":        Init_LocalPort,
		"WebAddr":     WebAddr,
		"WebPort":     WebPort,
		"WebStatic":   Web_path_static,
		"WebViews":    Web_path_views,
		"RpcServer":   RpcServer,
		"RpcUser":     RPCUser,
		"RpcPassword": RPCPassword,
	})
	return string(b)
}
