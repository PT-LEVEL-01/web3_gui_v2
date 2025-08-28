package utils

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"
	"web3_gui/chain/config"

	"web3_gui/libp2parea/adapter/engine"
)

/*
为调试统计各步骤的耗时
*/
type elapsed struct {
	enable     bool
	height     uint64
	startAt    time.Time
	laststepAt time.Time
	steps      strings.Builder
}

// enable 使能
// 只能在调试模式打印
func NewElapsed(height uint64) *elapsed {
	nowtime := config.TimeNow()
	return &elapsed{
		enable:     false,
		height:     height,
		startAt:    nowtime,
		laststepAt: nowtime,
		steps:      strings.Builder{},
	}
}

// 记录点
func (e *elapsed) Step(names ...string) {
	if !e.enable {
		return
	}

	stepname := strings.Join(names, ",")
	_, file, line, _ := runtime.Caller(1)
	_, filename := path.Split(file)
	msg := fmt.Sprintf("%s:%d %s", filename, line, stepname)
	nowtime := config.TimeNow()
	steptime := nowtime.Sub(e.laststepAt)
	totaltime := nowtime.Sub(e.startAt)
	e.steps.WriteString(fmt.Sprintf("\nElapsed [%d][%10s / %10s] %s", e.height, steptime, totaltime, msg))
	e.laststepAt = nowtime
	if totaltime.Milliseconds() > 500 {
		e.enable = true
	}
}

// 结果打印
// 打印大于500ms的耗时日志
func (e *elapsed) Print() {
	if !e.enable {
		return
	}

	engine.Log.Debug("%s", e.steps.String())
}
