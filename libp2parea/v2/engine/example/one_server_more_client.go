package main

import (
	"strconv"
	"time"
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
	for range 10 {
		StartClientOne()
	}
}

/*
启动一个服务器
*/
func StartServer() {
	server := engine.NewEngine("server")
	server.RegisterMsg(msgid, Handler)

	ERR := server.ListenOnePort(serverPort, true)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}

func StartClientOne() {
	client := engine.NewEngine("client")
	session, ERR := client.Dial("/ip4/127.0.0.1/tcp/" + strconv.Itoa(serverPort) + "/ws")
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	hello := []byte("nihao")
	resultBs, ERR := session.SendWait(msgid, &hello, 5*time.Second)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	utils.Log.Info().Str("Result", string(*resultBs)).Send()
}

func Handler(msg *engine.Packet) {
	ERR := msg.Session.Reply(msg, &msg.Data, 0)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}
