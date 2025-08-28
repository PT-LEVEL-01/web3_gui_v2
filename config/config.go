package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

const (
	GUITitle          = "web3.0"                                       //GUI应用显示title名称
	VersionName       = "v0.0.1"                                       //版本名称
	VersionIndex      = 1                                              //版本号
	Path_configDir    = "conf"                                         //配置文件存放目录
	Path_config       = "config.json"                                  //配置文件名称
	Core_keystore     = "wallet.key"                                   //密钥文件
	versionUpdateAddr = "HUBHo3J7LPcv27GUVvQqEMNiprZuPJqDh8F3hRopMYVc" //版本更新节点地址
	UpdateProcName    = "./update.exe"                                 //更新重启程序可执行文件命令
)

const (
	ORDER_status_not_pay          = 1 //未支付
	ORDER_status_pay_not_onchain  = 2 //已支付，但未上链
	ORDER_status_pay_onchain      = 3 //已支付，已上链
	ORDER_status_not_pay_overtime = 4 //未支付超时
)

var (
	AreaName_str                               = "test" + strconv.Itoa(1)
	AreaName                                   = sha256.Sum256([]byte(AreaName_str))          //
	KeystoreFileAbsPath                        = filepath.Join(Path_configDir, Core_keystore) //密钥文件存放地址
	FILEPATH_config                            = filepath.Join(Path_configDir, Path_config)   //密钥文件存放地址
	AddrPre                                    = "TEST"                                       //收款地址前缀
	Init_LocalPort       uint16                = 19981                                        //
	ConfigJSONStr                              = ""                                           //
	ParamJSONstr                               = ""                                           //
	InitNode                                   = false                                        //本节点是否是创世节点
	NetAddr                                    = ""                                           //
	LevelDB              *utilsleveldb.LevelDB                                                //
	DownloadFileDir      = "download"                                                         //下载文件默认保存地址
	VersionUpdateAddress nodeStore.AddressNet                                                 //版本更新节点地址
	DiscoverPeerAddr     []string              = []string{}                                   //

	Node *libp2parea.Node //
)

func init() {
	VersionUpdateAddress = nodeStore.AddressFromB58String(versionUpdateAddr)
}

/*
读取本地config文件内容
*/
func LoadConfigJSONString(filePath string) (string, error) {
	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// panic("检查配置文件错误：" + err.Error())
		return "", err
	}
	if !ok {
		return "", errors.New("config file not exist")
	}
	bs, err := os.ReadFile(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// panic("读取配置文件错误：" + err.Error())
		return "", err
	}
	return string(bs), nil
}

/*
解析配置参数内容
@configStr    string    需要把配置文件中的内容传过来，本方法不读配置文件
@paramsStr    string    配置中是否有init参数，有则用初始节点方式启动
*/
func ParseConfig(configStr, paramsStr string) error {
	ConfigJSONStr = configStr
	ParamJSONstr = paramsStr

	if paramsStr != "" {
		bs := []byte(paramsStr)
		err := parseParamsInit(bs)
		if err != nil {
			return err
		}
	}

	bs := []byte(configStr)
	return parseConfigJSON(bs)
}

type Config struct {
	Init          bool     `json:"init"`          //= flag.String("init", "", "创建创始区块(默认：genesis.json)")
	Load          bool     `json:"load"`          // = flag.String("load", "", "从历史区块拉起链端")
	AreaName      string   `json:"AreaName"`      //AreaName
	AddrPre       string   `json:"AddrPre"`       //收款地址前缀
	Port          uint16   `json:"Port"`          //监听端口
	RpcServer     bool     `json:"RpcServer"`     //RPC服务是否打开
	AdminPassword string   `json:"AdminPassword"` //RPC管理员初始密码
	RegistAddress []string `json:"RegistAddress"` //节点注册地址
	Download      string   `json:"download"`      //下载文件默认保存地址
}

/*
判断是否有init参数
*/
func ParseHaveInitFlag() bool {
	if InitNode {
		return true
	}
	for _, param := range flag.Args() {
		switch param {
		case "init":
			InitNode = true
			// Model = Model_complete
			return true
		case "load":
			// LoadNode = true
			// Model = Model_complete
			return true
		}
	}
	return false
}

// type Config struct {
// 	Init        bool   //= flag.String("init", "", "创建创始区块(默认：genesis.json)")
// 	Conf        string //= flag.String("conf", "conf/config.json", "指定配置文件(默认：conf/config.json)")
// 	Port        int    //= flag.Int("port", 9811, "本地监听端口(默认：9811)")
// 	NetId       int    //= flag.Int("netid", 20, "网络id(默认：20)")
// 	Ip          string //= flag.String("ip", "0.0.0.0", "本地IP地址(默认：0.0.0.0)")
// 	WebAddr     string //= flag.String("webaddr", "0.0.0.0", "本地web服务器IP地址(默认：0.0.0.0)")
// 	WebPort     int    //= flag.Int("webport", 2080, "web服务器端口(默认：2080)")
// 	WebStatic   string //= flag.String("westatic", "", "本地web静态文件目录")
// 	WebViews    string //= flag.String("webviews", "", "本地web Views文件目录")
// 	DataDir     string //= flag.String("datadir", "", "指定数据目录")
// 	DbCache     int    //= flag.Int("dbcache", 25, "设置数据库缓存大小，单位为兆字节（MB）（默认：25）")
// 	TimeOut     int    //= flag.Int("timeout", 0, "设置连接超时，单位为毫秒")
// 	RpcServer   bool   //= flag.Bool("rpcserver", false, "打开或关闭JSON-RPC true/false(默认：false)")
// 	RpcUser     string //= flag.String("rpcuser", "", "JSON-RPC 连接使用的用户名")
// 	RpcPassword string //= flag.String("rpcpassword", "", "JSON-RPC 连接使用的密码")
// 	WalletPwd   string //= flag.String("walletpwd", config.Wallet_keystore_default_pwd, "钱包密码")
// 	classpath   string //= flag.String("classpath", "", "jar包路径")
// 	Load        bool   // = flag.String("load", "", "从历史区块拉起链端")
// }

/*
解析参数中是否有Init的值
*/
func parseParamsInit(bs []byte) error {
	cfi := new(Config)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(cfi)
	if err != nil {
		return err
	}
	InitNode = cfi.Init
	return nil
}

/*
解析json格式的配置项目
*/
func parseConfigJSON(bs []byte) error {
	cfi := new(Config)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(cfi)
	if err != nil {
		return err
	}

	//utils.Log.Info().Msgf("initNode %t", InitNode)
	if cfi.AreaName != "" {
		AreaName = sha256.Sum256([]byte(cfi.AreaName))
	}
	//Init_LocalIP = cfi.IP
	Init_LocalPort = cfi.Port
	//NetType = cfi.NetType
	AddrPre = cfi.AddrPre
	//	NetType = NetType_release //正式版所有节点都为release
	DiscoverPeerAddr = cfi.RegistAddress
	//utils.Log.Info().Msgf("DiscoverPeerAddr %+v", cfi.RegistAddress)
	cfi.Download = strings.Trim(cfi.Download, " ") //去掉前后空格
	if cfi.Download != "" {
		cfi.Download = filepath.Clean(cfi.Download) //
		DownloadFileDir = cfi.Download              //
	}
	return nil
}

/*
获取版本号
*/
func GetVersion() (string, int) {
	return VersionName, VersionIndex
}

func SetAreaName(areaName string) {
	AreaName_str = areaName
	AreaName = sha256.Sum256([]byte(AreaName_str))
}
