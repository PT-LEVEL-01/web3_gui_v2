package engine

import (
	"net"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestTcpListen(t *testing.T) {
	go example_tcpListen()
	time.Sleep(time.Second * 1)
	example_tcpClient()
	//time.Sleep(time.Second * 50)
	//utils.Log.Info().Str("tcp服务器执行完成", "").Send()
}

func example_tcpListen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	if err != nil {
		utils.Log.Error().Str("err", err.Error()).Send()
		return
	}

	engine := NewEngine("test_server")
	engine.RegisterMsg(msgid_hello, helloHandler)
	tcpListen := NewTcpListen(engine)
	ERR := tcpListen.Listen(tcpAddr, false)
	utils.Log.Info().Str("ERR", ERR.String()).Send()
}

var msgid_hello = uint64(15)
var msgid_hello_recv = uint64(16)

func helloHandler(msg *Packet) {
	utils.Log.Info().Int("数据长度", len(msg.Data)).Send()
}
func helloHandler_recv(msg Packet) {

}
