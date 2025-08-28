package engine

import "web3_gui/utils"

type LogItr interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type LogAdapter struct {
}

func (this *LogAdapter) Debug(format string, v ...interface{}) {
	utils.Log.Debug().Msgf(format, v...)
}
func (this *LogAdapter) Info(format string, v ...interface{}) {
	utils.Log.Info().Msgf(format, v...)
}
func (this *LogAdapter) Warn(format string, v ...interface{}) {
	utils.Log.Warn().Msgf(format, v...)
}
func (this *LogAdapter) Error(format string, v ...interface{}) {
	utils.Log.Error().Msgf(format, v...)
}
