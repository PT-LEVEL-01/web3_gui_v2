package engine

import (
	"testing"
	"web3_gui/utils"
)

func TestTcpClient(t *testing.T) {
	example_tcpClient()
	//utils.Log.Info().Str("tcp服务器执行完成", "").Send()
}

func example_tcpClient() {
	engine := NewEngine("test_client")
	clientSession, ERR := engine.DialTcpAddr("", 8080)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Hex("sid", clientSession.GetId()).Send()
	bs := make([]byte, 1024*1024*100)
	_, ERR = clientSession.Send(msgid_hello, &bs, 0)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}
