package storage

import (
	"testing"
	"web3_gui/utils"
)

func TestGetFilePathForDir(*testing.T) {
	dirNames, _, err := GetFilePathForDir("D:\\迅雷下载")
	if err != nil {
		utils.Log.Error().Str("ERR", err.Error()).Send()
		return
	}
	utils.Log.Info().Interface("文件夹名称列表", len(dirNames)).Send()
}

func TestGetFileInfo(*testing.T) {
	fps, ERR := GetFileInfo([]string{"D:\\迅雷下载"})
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Int("文件数量", len(fps)).Send()
}
