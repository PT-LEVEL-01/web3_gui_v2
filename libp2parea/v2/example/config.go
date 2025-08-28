package example

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"web3_gui/utils"
)

var (
	AddrPre     = "TEST"                             //收款地址前缀
	AreaNameStr = "nihaoa"                           //域网络名称
	AreaName    = sha256.Sum256([]byte(AreaNameStr)) //域网络名称
	keyPwd      = "123456789"                        //钱包密码
	ip          = "127.0.0.1"                        //
	//ip       = "47.112.178.87" //
	basePort = uint16(19970) //节点本地监听的端口

	Path_rootDir = "D:/test/temp"
	//Path_rootDir       = ""
	Path_log_root      = filepath.Join(Path_rootDir, "logs")
	Path_db_root       = filepath.Join(Path_rootDir, "dbs")
	Path_keystore_root = filepath.Join(Path_rootDir, "keystore")

	defaultLogPath = filepath.Join(Path_log_root, "log.txt")
)

const (
	MSGID_hello    = 1 //
	MSGID_hello_HE = 2 //加密消息
)

func init() {
	//删除日志目录
	os.RemoveAll(Path_log_root)
	//日志文件路径
	utils.LogBuildDefaultFile(defaultLogPath)
}
