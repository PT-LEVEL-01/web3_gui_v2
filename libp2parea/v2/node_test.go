package libp2parea

import (
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"testing"
	"web3_gui/keystore/v2"
	"web3_gui/utils"
)

var (
	addrPre    = "TEST"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	clientHost = "127.0.0.1"
	basePort   = uint16(19960)
)

func TestArea(t *testing.T) {
	logPath := filepath.Join("D:/test", "log.txt")
	utils.LogBuildDefaultFile(logPath)
	utils.Log.Info().Str("start", "11111111").Send()
	keyPath := filepath.Join("D:/test/keystore", "keystore"+strconv.Itoa(0)+".key")

	key1 := keystore.NewKeystoreSingle(keyPath, addrPre)
	ERR := key1.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == keystore.ERROR_code_wallet_file_not_exist {
			ERR = key1.CreateRand(keyPwd, keyPwd, keyPwd, keyPwd)
		}
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建key1错误", ERR.String()).Send()
			return
		}
	}

	utils.Log.Info().Str("start", "11111111").Send()
	area, ERR := NewNode(areaName, addrPre, key1, keyPwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("创建area错误", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("start", "11111111").Send()
	ERR = area.StartUP(basePort)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("END", "11111111").Send()
}
