package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/config"
	"web3_gui/libp2parea/adapter/ntp"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Path_config_extra  = "config_extra.json" //额外配置文件名称
	Core_startup_key   = "startup.pem"       //startup 密钥文件
	Core_addr_prk      = "addr_ec_prk.pem"   //地址私钥文件名称
	Core_addr_puk      = "addr_ec_puk.pem"   //地址公钥文件名称
	Core_addr_prk_type = "EC PRIVATE KEY"    //地址私钥文件抬头
	Core_addr_puk_type = "EC PUBLIC KEY"     //地址公钥文件抬头
)

const (
	Name_prk = "name_ec_prk.pem" //地址私钥文件名称
	Name_puk = "name_ec_puk.pem" //地址公钥文件名称
)

const (
	Store_path_dir            = "store" //本地共享文件存储目录名称
	Store_path_fileinfo_self  = "self"  //自己上传的文件索引存储目录名称
	Store_path_fileinfo_local = "local" //本地下载过的文件索引存储目录名称
	Store_path_fileinfo_net   = "net"   //网络需要保存的文件索引存储目录名称
	Store_path_fileinfo_cache = "cache" //缓存中保存的文件索引存储目录名称
	Store_path_temp           = "temp"  //临时文件夹，本地上传存放目录，存放未切片的完整文件
	Store_path_files          = "files" //带扩展名的完整文件
	IsRemoveStore             = false   //启动时删除本地所有文件分片及分片索引
	//	IsCreateId                = true    //启动时是否要创建新的id

	HashCode    = utils.SHA3_256 //
	NodeIDLevel = 256            //节点id比特位数

	Model_complete = "complete"
	Model_light    = "light"
)

var (
	Path_config = "config.json"                       //配置文件名称
	AreaName    = sha256.Sum256([]byte("icom_chain")) //名称不同节点无法连接

	Area            *libp2parea.Area //
	Init_LocalIP                     = ""
	Init_LocalPort  uint16           = 19981
	WebAddr                          = ""             //
	WebPort         uint16           = 0              //本地监听端口
	Web_path_static                  = ""             //网页静态文件路径
	Web_path_views                   = ""             //网页模板文件路径
	AddrPre                          = ""             //收款地址前缀
	NetIds                           = []byte{0}      //网络号
	NetType_release                  = "release"      //正式网络
	NetType                          = "test"         //网络类型:正式网络release/测试网络test
	RpcServer                        = false          //
	RPCUser                          = ""             //
	RPCPassword                      = ""             //
	Model                            = Model_complete //默认是完整版模式

	Wallet_txitem_save_db        = false //是否把未花费的余额保存在数据库，来降低内存
	Entry                        = []string{}
	CPUNUM                       = runtime.NumCPU()
	OS                           = runtime.GOOS //操作系统
	RpcPort               uint16 = 8888

	Path_configDir = "conf" //配置文件存放目录

	WitnessMinCpuNum        = 4          // 见证者节点拥有的最小cpu数量
	WitnessMinMem    uint64 = 8388608000 // 见证者节点拥有的最小内存

	CheckStartBlockHash = []byte{} //用于验证的创始区块hash
	//EVM_Reward_Enable   = false    //虚拟机合约发奖励使能,缺省关闭

	//硬件检查
	MinCPUCores         = uint32(0) //uint32(3)        //最低cpu核数
	MinFreeMemory       = uint32(0) //uint32(8 * 1024) //最低可用内存(MB)
	MinFreeDisk         = uint32(0) //uint32(6 * 1024) //最低可用磁盘空间(MB)
	MinNetworkBandwidth = uint32(0) //uint32(10)       //最低带宽(MB/s)
)

var (
	//Core_keystore       = "keystore.key"     //密钥文件
	KeystoreFileAbsPath = filepath.Join(Path_configDir, "keystore.key") //密钥文件存放地址

	Store_dir            string = filepath.Join(Store_path_dir)                            //本地共享文件存储目录路径
	Store_fileinfo_self  string = filepath.Join(Store_path_dir, Store_path_fileinfo_self)  //自己上传的文件索引存储目录路径
	Store_fileinfo_local string = filepath.Join(Store_path_dir, Store_path_fileinfo_local) //本地下载过的文件索引存储目录路径
	Store_fileinfo_net   string = filepath.Join(Store_path_dir, Store_path_fileinfo_net)   //网络需要保存的文件索引存储目录路径
	Store_fileinfo_cache string = filepath.Join(Store_path_dir, Store_path_fileinfo_cache) //缓存中保存的文件索引存储目录路径
	Store_temp           string = filepath.Join(Store_path_dir, Store_path_temp)           //临时文件夹，本地上传存放目录，存放未切片的完整文件
	Store_files          string = filepath.Join(Store_path_dir, Store_path_files)          //存放带扩展名的完整文件

)

func SetConfigDir(dir string) {
	if dir == "" {
		return
	}

	Path_configDir = dir
}

func setDefaultImConfig(dir string) {
	cfi := Config{
		IP:          "127.0.0.1",
		Port:        19981,
		WebAddr:     "0.0.0.0",
		WebPort:     2080,
		WebStatic:   "./static",
		WebViews:    "./views",
		RpcServer:   true,
		RpcUser:     "test",
		RpcPassword: "testp",
		Miner:       true,
		NetType:     "release",
		AddrPre:     "ICOM",
		BalanceDB:   false,
		Model:       "",
		RpcPort:     8888,
	}
	bs, err := json.Marshal(cfi)
	if err != nil {
		panic("marshal config ：" + err.Error())
		return
	}

	_, err = os.Stat(filepath.Dir(dir))
	if os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(dir), 0770); err != nil {
			return
		}
	}

	f, err := os.OpenFile(dir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic("create file：" + err.Error())
		return
	}
	_, err = f.Write(bs)
	if err != nil {
		panic("write file：" + err.Error())
		return
	}
	f.Close()

}

// func ParseConfig() {

// 	if OS == "windows" {
// 		Wallet_print_serialize_hex = false
// 	} else {
// 		Wallet_print_serialize_hex = false
// 	}

// 	fmt.Println("---------------Path_configDir-----------------------")
// 	fmt.Println(filepath.Join(Path_configDir, Path_config))

// 	// fmt.Println("1111111111111")
// 	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config))
// 	if err != nil {
// 		// fmt.Println("22222222222222")
// 		panic("检查配置文件错误：" + err.Error())
// 		return
// 	}
// 	// fmt.Println("3333333333333")
// 	if !ok {
// 		setDefaultImConfig(filepath.Join(Path_configDir, Path_config))
// 		// fmt.Println("4444444444444")
// 		//cfi := new(Config)
// 		//cfi.Port = 9981
// 		//bs, _ := json.Marshal(cfi)
// 		//// fmt.Println("5555555555555555")
// 		//f, err := os.OpenFile(filepath.Join(Path_configDir, Path_config), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
// 		//if err != nil {
// 		//	// fmt.Println("666666666666666666")
// 		//	panic("创建配置文件错误：" + err.Error())
// 		//	return
// 		//}
// 		//// fmt.Println("77777777777777777")
// 		//_, err = f.Write(bs)
// 		//if err != nil {
// 		//	// fmt.Println("88888888888888")
// 		//	panic("写入配置文件错误：" + err.Error())
// 		//	return
// 		//}
// 		//// fmt.Println("999999999999999")
// 		//f.Close()
// 	}
// 	// fmt.Println("11111111111111111111111111")
// 	bs, err := ioutil.ReadFile(filepath.Join(Path_configDir, Path_config))
// 	if err != nil {
// 		// fmt.Println("2222222222222222222222222222")
// 		panic("读取配置文件错误：" + err.Error())
// 		return
// 	}

// 	// fmt.Println("3333333333333333333")
// 	cfi := new(Config)
// 	// err = json.Unmarshal(bs, cfi)

// 	decoder := json.NewDecoder(bytes.NewBuffer(bs))
// 	decoder.UseNumber()
// 	err = decoder.Decode(cfi)
// 	if err != nil {
// 		// fmt.Println("44444444444444444444444444")
// 		return
// 	}
// 	// fmt.Printf("55555555555555555555555 %+v", cfi)
// 	if cfi.AreaName != "" {
// 		AreaName = sha256.Sum256([]byte(cfi.AreaName))
// 	}
// 	Init_LocalIP = cfi.IP
// 	Init_LocalPort = cfi.Port
// 	config.Init_GatewayPort = cfi.Port
// 	Web_path_static = cfi.WebStatic
// 	Web_path_views = cfi.WebViews
// 	// engine. Netid = cfi.Netid
// 	WebAddr = cfi.WebAddr
// 	WebPort = cfi.WebPort
// 	Miner = cfi.Miner
// 	NetType = cfi.NetType
// 	//	NetType = NetType_release //正式版所有节点都为release
// 	AddrPre = cfi.AddrPre
// 	RpcServer = cfi.RpcServer
// 	RPCUser = cfi.RpcUser
// 	RPCPassword = cfi.RpcPassword
// 	Wallet_txitem_save_db = cfi.BalanceDB
// 	DisableSnapshot = cfi.DisableSnapshot

// 	if cfi.Model == Model_light {
// 		Model = Model_light
// 	}
// 	RpcPort = cfi.RpcPort

// 	if cfi.CheckStartBlockHash != "" {
// 		CheckStartBlockHash, err = hex.DecodeString(cfi.CheckStartBlockHash)
// 		if err != nil {
// 			panic("配置文件中CheckStartBlockHash解析错误：" + err.Error())
// 		}
// 	}
// }

/*
加载本地配置文件，并解析配置文件
*/
func ParseConfig() {
	if OS == "windows" {
		Wallet_print_serialize_hex = false
	} else {
		Wallet_print_serialize_hex = false
	}

	bs, err := loadConfigLocal()
	if err != nil {
		panic("读取配置文件错误：" + err.Error())
	}
	err = parseConfigJSON(bs)
	if err != nil {
		panic("解析配置文件错误：" + err.Error())
	}
}

/*
从本地路径中加载配置文件
*/
func loadConfigLocal() ([]byte, error) {
	//fmt.Println("---------------Path_configDir-----------------------")
	//fmt.Println(filepath.Join(Path_configDir, Path_config))

	// fmt.Println("1111111111111")
	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// fmt.Println("22222222222222")
		return nil, err
	}
	// fmt.Println("3333333333333")
	if !ok {
		setDefaultImConfig(filepath.Join(Path_configDir, Path_config))
		// fmt.Println("4444444444444")
		//cfi := new(Config)
		//cfi.Port = 9981
		//bs, _ := json.Marshal(cfi)
		//// fmt.Println("5555555555555555")
		//f, err := os.OpenFile(filepath.Join(Path_configDir, Path_config), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		//if err != nil {
		//	// fmt.Println("666666666666666666")
		//	panic("创建配置文件错误：" + err.Error())
		//	return
		//}
		//// fmt.Println("77777777777777777")
		//_, err = f.Write(bs)
		//if err != nil {
		//	// fmt.Println("88888888888888")
		//	panic("写入配置文件错误：" + err.Error())
		//	return
		//}
		//// fmt.Println("999999999999999")
		//f.Close()
	}
	// fmt.Println("11111111111111111111111111")
	bs, err := ioutil.ReadFile(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// fmt.Println("2222222222222222222222222222")
		return nil, err
	}
	return bs, nil
}

/*
解析json格式的配置项目
*/
func parseConfigJSON(bs []byte) error {
	// fmt.Println("3333333333333333333")
	cfi := new(Config)
	// err = json.Unmarshal(bs, cfi)

	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(cfi)
	if err != nil {
		// fmt.Println("44444444444444444444444444")
		return err
	}
	// fmt.Printf("55555555555555555555555 %+v", cfi)
	if cfi.AreaName != "" {
		AreaName = sha256.Sum256([]byte(cfi.AreaName))
	}
	Init_LocalIP = cfi.IP
	Init_LocalPort = cfi.Port
	config.Init_GatewayPort = cfi.Port
	Web_path_static = cfi.WebStatic
	Web_path_views = cfi.WebViews
	// engine. Netid = cfi.Netid
	WebAddr = cfi.WebAddr
	WebPort = cfi.WebPort
	Miner = cfi.Miner
	NetType = cfi.NetType
	//	NetType = NetType_release //正式版所有节点都为release
	AddrPre = cfi.AddrPre
	RpcServer = cfi.RpcServer
	RPCUser = cfi.RpcUser
	RPCPassword = cfi.RpcPassword
	Wallet_txitem_save_db = cfi.BalanceDB
	DisableSnapshot = cfi.DisableSnapshot
	EnableRestart = cfi.EnableRestart
	EnableStartupWeb = cfi.EnableStartupWeb
	if cfi.Model == Model_light {
		Model = Model_light
	}
	RpcPort = cfi.RpcPort
	//EVM_Reward_Enable = cfi.EvmRewardEnable
	//MinCPUCores = cfi.MinCPUCores
	//MinFreeMemory = cfi.MinFreeMemory
	//MinFreeDisk = cfi.MinFreeDisk
	//MinNetworkBandwidth = cfi.MinNetworkBandwidth

	if cfi.CheckStartBlockHash != "" {
		CheckStartBlockHash, err = hex.DecodeString(cfi.CheckStartBlockHash)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseConfigV2(configPath string) {

	if OS == "windows" {
		Wallet_print_serialize_hex = false
	} else {
		Wallet_print_serialize_hex = false
	}

	// fmt.Println("1111111111111")
	ok, err := utils.PathExists(configPath)

	fmt.Println("ok:", ok)
	fmt.Println("configPath:", configPath)

	if err != nil {
		// fmt.Println("22222222222222")
		panic("检查配置文件错误：" + err.Error())
		return
	}
	// fmt.Println("3333333333333")
	if !ok {
		setDefaultImConfig(configPath)
		// fmt.Println("4444444444444")
		//cfi := new(Config)
		//cfi.Port = 9981
		//bs, _ := json.Marshal(cfi)
		//// fmt.Println("5555555555555555")
		//f, err := os.OpenFile(filepath.Join(Path_configDir, Path_config), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		//if err != nil {
		//	// fmt.Println("666666666666666666")
		//	panic("创建配置文件错误：" + err.Error())
		//	return
		//}
		//// fmt.Println("77777777777777777")
		//_, err = f.Write(bs)
		//if err != nil {
		//	// fmt.Println("88888888888888")
		//	panic("写入配置文件错误：" + err.Error())
		//	return
		//}
		//// fmt.Println("999999999999999")
		//f.Close()
	}
	fmt.Println("11111111111111111111111111:", configPath)
	bs, err := ioutil.ReadFile(configPath)
	if err != nil {
		// fmt.Println("2222222222222222222222222222")
		panic("读取配置文件错误：" + err.Error())
		return
	}

	// fmt.Println("3333333333333333333")
	cfi := new(Config)

	// err = json.Unmarshal(bs, cfi)

	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(cfi)
	if err != nil {
		// fmt.Println("44444444444444444444444444")
		panic("解析配置文件错误：" + err.Error())
		return
	}

	fmt.Println("------------new(Config)----------------------")
	fmt.Println(cfi.AddrPre)

	// fmt.Printf("55555555555555555555555 %+v", cfi)
	if cfi.AreaName != "" {
		AreaName = sha256.Sum256([]byte(cfi.AreaName))
	}
	Init_LocalIP = cfi.IP
	Init_LocalPort = cfi.Port
	config.Init_GatewayPort = cfi.Port
	Web_path_static = cfi.WebStatic
	Web_path_views = cfi.WebViews
	// engine. Netid = cfi.Netid
	WebAddr = cfi.WebAddr
	WebPort = cfi.WebPort
	Miner = cfi.Miner
	NetType = cfi.NetType
	//	NetType = NetType_release //正式版所有节点都为release
	AddrPre = cfi.AddrPre
	RpcServer = cfi.RpcServer
	RPCUser = cfi.RpcUser
	RPCPassword = cfi.RpcPassword
	Wallet_txitem_save_db = cfi.BalanceDB
	if cfi.Model == Model_light {
		Model = Model_light
	}
	RpcPort = cfi.RpcPort
}

type Config struct {
	Init bool `json:"init"` //= flag.String("init", "", "创建创始区块(默认：genesis.json)")
	Load bool // = flag.String("load", "", "从历史区块拉起链端")
	// Netid       uint32 `json:"netid"`       //
	IP                  string `json:"ip"`                  //ip地址
	AreaName            string `json:"AreaName"`            //AreaName
	Port                uint16 `json:"port"`                //监听端口
	WebAddr             string `json:"WebAddr"`             //
	WebPort             uint16 `json:"WebPort"`             //
	WebStatic           string `json:"WebStatic"`           //
	WebViews            string `json:"WebViews"`            //
	RpcServer           bool   `json:"RpcServer"`           //
	RpcUser             string `json:"RpcUser"`             //
	RpcPassword         string `json:"RpcPassword"`         //
	Miner               bool   `json:"miner"`               //本节点是否是矿工
	NetType             string `json:"NetType"`             //正式网络release/测试网络test
	AddrPre             string `json:"AddrPre"`             //收款地址前缀
	BalanceDB           bool   `json:"balancedb"`           //是否开启余额保存在数据库的模式
	Model               string `json:"model"`               //是否开启轻节点S模式
	RpcPort             uint16 `json:"RpcPort"`             //grpc服务端口
	CheckStartBlockHash string `json:"CheckStartBlockHash"` //创始区块hash，用于验证
	DisableSnapshot     bool   `json:"DisableSnapshot"`     //是否禁用快照功能
	EvmRewardEnable     bool   `json:"EvmRewardEnable"`     //是否开启主链奖励虚拟机模式
	EnableRestart       bool   `json:"EnableRestart"`       //是否启用自重启功能
	EnableStartupWeb    bool   `json:"EnableStartupWeb"`    //是否启用StartupWEB
	//MinCPUCores         uint32 `json:"MinCPUCores"`         //最低cpu核数
	//MinFreeMemory       uint32 `json:"MinFreeMemory"`       //最低可用内存(MB)
	//MinFreeDisk         uint32 `json:"MinFreeDisk"`         //最低可用磁盘空间(MB)
	//MinNetworkBandwidth uint32 `json:"MinNetworkBandwidth"` //最低带宽(MB/s)
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

// 获取当前时间

func TimeNow() time.Time {
	return ntp.GetNtpTime()
}

/*
解析配置参数内容
@configStr    string    需要把配置文件中的内容传过来，本方法不读配置文件
@paramsStr    string    配置中是否有init参数，有则用初始节点方式启动
*/
func ParseConfigWithConfigJSON(configStr, paramsStr string) error {
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
