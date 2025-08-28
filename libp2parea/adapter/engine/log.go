package engine

import (
	"web3_gui/utils"
)

var logOpen = true //日志开关
var Log = &LogAdapter{}

func SetLogPath(logPath string) {
	utils.LogBuildDefaultFile(logPath)
}

/*
关闭日志
*/
func LogClose() {
	logOpen = false
}

type LogItr interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type LogAdapter struct {
}

func (this *LogAdapter) Debug(format string, v ...interface{}) {
	if !logOpen {
		return
	}
	if len(v) == 0 {
		utils.Log.Debug().Msgf(format)
		return
	}
	utils.Log.Debug().Msgf(format, v...)
}
func (this *LogAdapter) Info(format string, v ...interface{}) {
	if !logOpen {
		return
	}
	if len(v) == 0 {
		utils.Log.Info().Msgf(format)
		return
	}
	utils.Log.Info().Msgf(format, v...)
}
func (this *LogAdapter) Warn(format string, v ...interface{}) {
	if !logOpen {
		return
	}
	if len(v) == 0 {
		utils.Log.Warn().Msgf(format)
		return
	}
	utils.Log.Warn().Msgf(format, v...)
}
func (this *LogAdapter) Error(format string, v ...interface{}) {
	if !logOpen {
		return
	}
	if len(v) == 0 {
		utils.Log.Error().Msgf(format)
		return
	}
	utils.Log.Error().Msgf(format, v...)
}
