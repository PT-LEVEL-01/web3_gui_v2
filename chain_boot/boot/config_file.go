package boot

import (
	"crypto/rand"
	"math/big"
	"path/filepath"
	"strings"
	"web3_gui/chain/config"
	gconfig "web3_gui/config"
	"web3_gui/utils"
)

/*
保存配置文件
若文件已经存在，则不做任何修改
*/
func SaveConfigFile(filePath string) error {
	//若文件已经存在，则退出
	ok, err := utils.PathExists(filePath)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	//随机生成8位的密码,作为RPC管理员初始密码
	password, err := GeneratePassword(8)
	if err != nil {
		return err
	}
	cfi := Config{
		AreaName:      gconfig.AreaName_str,                    //
		Port:          config.Init_LocalPort,                   //
		RpcServer:     true,                                    //
		AddrPre:       config.AddrPre,                          //
		AdminPassword: password,                                //RPC管理员初始密码
		RegistAddress: []string{"/ip4/127.0.0.1/tcp/25331/ws"}, //节点注册地址
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	dir, _ := filepath.Split(absPath)
	//若文件夹不存在，则创建
	err = utils.CheckCreateDir(dir)
	if err != nil {
		return err
	}
	return utils.SaveJsonFile(filePath, cfi)
}

type Config struct {
	AreaName      string   `json:"AreaName"`      //AreaName
	AddrPre       string   `json:"AddrPre"`       //收款地址前缀
	Port          uint16   `json:"Port"`          //监听端口
	RpcServer     bool     `json:"RpcServer"`     //RPC服务是否打开
	AdminPassword string   `json:"AdminPassword"` //RPC管理员初始密码
	RegistAddress []string `json:"RegistAddress"` //节点注册地址
}

const (
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits           = "0123456789"
	specialChars     = "!@#$%^&*()-_=+,.?/:;{}[]~"
)

func GeneratePassword(length int) (string, error) {
	// 将所有可能的字符组合在一起
	allChars := lowercaseLetters + uppercaseLetters + digits + specialChars
	var password strings.Builder
	for i := 0; i < length; i++ {
		// 随机选择一个字符
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}
	return password.String(), nil
}

func randomChar(chars string) (byte, error) {
	// 生成一个随机索引
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, err
	}
	return chars[n.Int64()], nil
}
