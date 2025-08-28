package engine

import (
	"testing"
	"web3_gui/utils"
)

func TestWsClient(t *testing.T) {
	wsclient_example()
}

func wsclient_example() {
	engine := NewEngine("test")
	_, ERR := NewWsClientConn(engine, "127.0.0.1", "8080")
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}
