package engine

import (
	"net"
	"testing"
	"web3_gui/utils"
)

func TestWsListen(t *testing.T) {
	wsListen_example()
}

func wsListen_example() {
	addr, err := net.ResolveTCPAddr("tcp4", ":8080")
	if err != nil {
		utils.Log.Error().Str("err", err.Error()).Send()
		return
	}

	engine := NewEngine("test")
	wsServer := NewWebsocketListen(engine)
	ERR := wsServer.Listen(addr, false)
	utils.Log.Info().Str("ERR", ERR.String()).Send()
}
