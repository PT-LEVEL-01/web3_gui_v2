package libp2parea

var (
	Path_configDir = "conf"         //配置文件存放目录
	Core_keystore  = "keystore.key" //密钥文件
)

var CommitInfo string

type Config struct {
	IsFirst bool
	Addr    string
	Port    uint16
	pwd     string
}
