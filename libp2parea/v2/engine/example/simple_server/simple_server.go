package main

import (
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

const serverPort = 8080
const msgid = 1

/*
一个服务器，多个客户端来连接
*/
func main() {
	StartServer()
}

/*
启动一个服务器
*/
func StartServer() {
	server := engine.NewEngine("server")
	server.RegisterMsg(msgid, Handler)

	ERR := server.ListenOnePort(serverPort, false)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}

func Handler(msg *engine.Packet) {
	ERR := msg.Session.Reply(msg, &msg.Data, 0)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}
