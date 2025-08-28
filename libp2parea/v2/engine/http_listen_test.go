package engine

import (
	"net"
	"testing"
	"web3_gui/utils"
)

func TestHttpListen(t *testing.T) {
	httpListen_example()
}

func httpListen_example() {
	httpListen := NewHttpListen(&utils.Log)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	if err != nil {
		utils.Log.Error().Str("err", err.Error()).Send()
		return
	}

	ERR := httpListen.Listen(tcpAddr, false)
	utils.Log.Info().Str("ERR", ERR.String()).Send()
}
