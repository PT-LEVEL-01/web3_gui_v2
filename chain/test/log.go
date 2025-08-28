package main

import (
	"time"
	"web3_gui/libp2parea/adapter/engine"
)

func main() {
	engine.NLog.Debug(engine.LOG_file, "%s", "nihao")
	engine.NLog.Error(engine.LOG_file, "%s", "我不好")
	time.Sleep(time.Second)
}
