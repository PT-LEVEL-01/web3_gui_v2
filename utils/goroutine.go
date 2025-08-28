package utils

import (
	"fmt"
	"github.com/rs/zerolog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func Go(f func(), log *zerolog.Logger) {
	thislog := log
	if thislog == nil {
		thislog = Log
	}
	go func() {
		defer PrintPanicStack(thislog)
		f()
	}()
}

/*
错误处理
@return    bool    是否有错误
*/
func PrintPanicStack(log *zerolog.Logger) bool {
	thislog := log
	if thislog == nil {
		thislog = Log
	}
	if x := recover(); x != nil {
		thislog.Error().Int("n", 0).Str("file", fmt.Sprintf("panic:%v", x)).Send()
		for i := 0; i < 10; i++ {
			//funcName, filename, line, ok := runtime.Caller(i)
			funcName, filename, line, ok := runtime.Caller(i)
			if ok {
				fnName := runtime.FuncForPC(funcName).Name()
				fnName = strings.SplitAfterN(fnName, ".", 2)[1]
				_, filename = filepath.Split(filename)
				thislog.Error().
					Int("n", i+1).
					Str("file", filename+":"+fnName+":"+strconv.Itoa(line)).
					Send()
				//Log.Error("%d frame :[func:%s,filename:%s,line:%d]\n", i, runtime.FuncForPC(funcName).Name(), filename, line)
			}
		}
		return true
	} else {
		return false
	}
}
