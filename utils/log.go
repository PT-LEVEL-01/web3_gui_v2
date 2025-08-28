/*
全局日志记录 使用"github.com/rs/zerolog"
TRACE(-1)：用于跟踪代码执行路径
DEBUG(0): 对故障排除有用的信息
INFO(1): 描述应用程序正常运行的信息
WARNING(2): 对于需要的记录事件，以后可能需要检查
ERROR(3): 特定操作的错误信息
FATAL(4): 应用程序无法恢复的严重错误。os.Exit(1) 在记录消息后调用
PANIC(5): 与 FATAL 类似，但只是名字改成了 PANIC()
*/

package utils

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var Log *zerolog.Logger

func init() {
	//自定义预设字段名称
	zerolog.TimestampFieldName = "t" //时间
	zerolog.LevelFieldName = "l"     //级别
	zerolog.MessageFieldName = "m"   //消息
	zerolog.CallerFieldName = "c"    //文件行号
	zerolog.ErrorFieldName = "e"     //
	LogBuildDefaultConsole()
}

/*
默认日志配置，使用json格式，输出到控制台
*/
func LogBuildDefaultConsole() {
	//var output io.Writer = &zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05.000"}
	//if os.Getenv("GO_ENV") != "development" {
	//	output = os.Stderr
	//}
	log := zerolog.New(os.Stderr).
		//Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		//Int("pid", os.Getpid()).
		//Str("go_version", buildInfo.GoVersion).
		Logger()
	Log = &log
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		_, filename := filepath.Split(file)
		return filename + ":" + strconv.Itoa(line)
	}
}

/*
默认日志配置，使用json格式，输出到文件
*/
func LogBuildDefaultFile(filename string) error {
	log, err := NewLogDefaultFile(filename)
	if err != nil {
		return err
	}
	Log = log
	return nil
}

func NewLogDefaultFile(filename string) (*zerolog.Logger, error) {
	//获取绝对路径
	fileDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		Log.Error().Str("e", err.Error())
		return nil, err
	}
	//Log.Info().Str("log file output path", filepath.Join(fileDir, filename)).Send()
	//创建文件夹
	err = CheckCreateDir(fileDir)
	if err != nil {
		return nil, err
	}
	//打开输出日志文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}

	//logLine := func() int {
	//	_, file, line, _ := runtime.Caller(1)
	//	return line
	//}
	//_, filename, line, _ := runtime.Caller(1)
	//_, filename = filepath.Split(filename)

	// 创建一个带颜色高亮显示的写入控制台的consoleWriter
	//consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	// 创建一个multiWriter，将日志同时写入文件和控制台
	multiWriter := io.MultiWriter(file, os.Stderr)
	log := zerolog.New(multiWriter).
		//Level(zerolog.DebugLevel).
		With().
		Timestamp().
		//Str("c", logLine()).
		Caller().
		//Int("pid", os.Getpid()).
		//Str("go_version", buildInfo.GoVersion).
		Logger()
	//Log.UpdateContext(func(c zerolog.Context) zerolog.Context {
	//	return c.Str("name", logLine())
	//})
	//zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return &log, nil
}

func logLine() string {
	lineStr := ""
	for i := range 10 {
		_, _, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if lineStr != "" {
			lineStr += " "
		}
		lineStr = lineStr + strconv.Itoa(i) + ":" + strconv.Itoa(line)
	}
	return lineStr
}

/*
构建为ConsoleWriter格式输出
注意不要在生产环境中使用 ， ConsoleWriter 因为它会大大减慢日志记录的速度。它只是为了帮助在开发应用程序时使日志更易于阅读。
您可以使用环境变量仅在开发中启用 ConsoleWriter 输出
*/
func LogBuildColorOutputConsole() {
	var output io.Writer = &zerolog.ConsoleWriter{Out: os.Stderr}
	//l.With().Timestamp().Caller().Logger()
	log := zerolog.New(output).
		//Level(zerolog.TraceLevel).
		With().
		//Timestamp().
		Str("t", time.Now().Format("2006-01-02 15:04:05.000")).
		Caller().
		//Int("pid", os.Getpid()).
		//Str("go_version", buildInfo.GoVersion).
		Logger()
	Log = &log
}

/*
构建为ConsoleWriter格式输出
注意不要在生产环境中使用 ， ConsoleWriter 因为它会大大减慢日志记录的速度。它只是为了帮助在开发应用程序时使日志更易于阅读。
您可以使用环境变量仅在开发中启用 ConsoleWriter 输出
*/
func LogBuildColorOutputFile(filename string) error {
	//
	fileDir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		Log.Error().Str("e", err.Error())
		return err
	}
	Log.Info().Str("log file output path", filepath.Join(fileDir, filename)).Send()
	err = CheckCreateDir(fileDir)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	var output io.Writer = &zerolog.ConsoleWriter{Out: file, TimeFormat: "2006-01-02 15:04:05.000"}
	//l.With().Timestamp().Caller().Logger()
	log := zerolog.New(output).
		//Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		//Int("pid", os.Getpid()).
		//Str("go_version", buildInfo.GoVersion).
		Logger()
	Log = &log
	return nil
}

/*
	Package log provides 全局日志记录，内部使用beego/log

	使用方法
	utils.GlobalInit("console", "", "debug", 1)
	utils.GlobalInit("file", `{"filename":"/var/log/gd/gd.log"}`, "", 1000)
	utils.Log.Debug().Msgf("session handle receive, %d, %v", msg.Code(), msg.Content())
	utils.Log.Debug().Msgf("test debug")
	utils.Log.Warn().Msgf("test warn")
	utils.Log.Error().Msgf("test error")
*/
//
//package utils
//
//import (
//	"errors"
//	"github.com/astaxie/beego/logs"
//)
//
//var (
//	logIsOpen = false
//	Log       *BeegoLog
//)
//
//func GlobalInit(kind, path, level string, length int) error {
//	logIsOpen = true
//	Log = new(BeegoLog)
//	if Log.log == nil {
//		Log.log = logs.NewLogger(int64(length))
//	}
//
//	// if Log == nil {
//	// 	Log = logs.NewLogger(int64(length))
//	// }
//
//	err := Log.log.SetLogger(kind, path)
//	if err != nil {
//		return err
//	}
//
//	switch level {
//	case "debug":
//		Log.log.SetLevel(logs.LevelDebug)
//	case "info":
//		Log.log.SetLevel(logs.LevelInfo)
//	case "warn":
//		Log.log.SetLevel(logs.LevelWarn)
//	case "error":
//		Log.log.SetLevel(logs.LevelError)
//	default:
//		return errors.New("未处理的日志记录等级")
//	}
//
//	return nil
//
//}
//
//type BeegoLog struct {
//	log *logs.BeeLogger
//}
//
//func (this *BeegoLog) Info(format string, v ...interface{}) {
//	if logIsOpen {
//		this.log.Info(format, v...)
//	}
//}
//func (this *BeegoLog) Debug(format string, v ...interface{}) {
//	if logIsOpen {
//		this.log.Debug(format, v...)
//	}
//}
//func (this *BeegoLog) Warn(format string, v ...interface{}) {
//	if logIsOpen {
//		this.log.Warn(format, v...)
//	}
//}
//func (this *BeegoLog) Error(format string, v ...interface{}) {
//	if logIsOpen {
//		this.log.Error(format, v...)
//	}
//}
