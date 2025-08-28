package utils

import (
	"testing"
)

func TestLog(t *testing.T) {
	//logExample()
	//logExample_maxSize()
}

func logExample() {
	defer PrintPanicStack(Log)
	Log.Info().Int("age", 12).Msg("nihao")
	err := LogBuildDefaultFile("logs/log.txt")
	if err != nil {
		Log.Fatal().Err(err).Msg("fail")
	}
	Log.Info().Int("age", 12).Msg("nihao")
	Log.Info().Msgf("age:%d", 5)
	Log.Info().Str("msg", "end").Send()
	//Log.Info("number:%d", 555)
	//切换为带颜色高亮显示的日志
	LogBuildColorOutputConsole()
	Log.Info().Int("age", 12).Msg("nihao")
	Log.Info().Int("age", 12).Msg("nihao")
	LogBuildColorOutputFile("logs/log.txt")
	Log.Info().Int("age", 12).Msg("nihao")
	panic("haha")
}

/*
测试日志分割大小
*/
func logExample_maxSize() {
	err := LogBuildDefaultFile("logs/log.txt")
	if err != nil {
		Log.Fatal().Err(err).Msg("fail")
	}
	for i := 0; i < 10000*10000; i++ {
		Log.Info().Int("age", 12).Msg("nihao")
	}
}
