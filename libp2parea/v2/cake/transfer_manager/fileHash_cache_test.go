package transfer_manager

import (
	"testing"
	"web3_gui/utils"
)

func TestGetDirPaths(*testing.T) {
	dirs, ERR := GetDirPaths("D:\\迅雷下载")
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("dir", dirs[0]).Send()
}
