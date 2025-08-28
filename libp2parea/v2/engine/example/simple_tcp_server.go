package main

import (
	"net"
	"strconv"
	"time"
	"web3_gui/libp2parea/v2/v2/engine"
	"web3_gui/utils"
)

func main() {
	tcpServer()
}

func tcpServer() {
	ip := ""
	port := 8848
	listenHost := ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenHost)
	if err != nil {
		utils.Log.Error().Str("error", err.Error()).Send()
		//Log.Error("%v", err)
		return
	}

	lis, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		utils.Log.Error().Str("error", err.Error()).Send()
		//Log.Error("%v", err)
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			utils.Log.Error().Str("error", err.Error()).Send()
			continue
		}
		go tcpReadServer(conn)
		go tcpWriteServer(conn)
	}
}

func tcpReadServer(conn net.Conn) {
	for {
		bs, err := engine.ReadStream(conn)
		if err != nil {
			utils.Log.Error().Str("error", err.Error()).Send()
			break
		}
		utils.Log.Debug().Str("data", string(bs)).Send()
	}
}

func tcpWriteServer(conn net.Conn) {
	bs, err := engine.BuildStream([]byte("nihao"))
	if err != nil {
		utils.Log.Error().Str("error", err.Error()).Send()
		return
	}
	for {
		_, err = conn.Write(bs)
		if err != nil {
			utils.Log.Error().Str("error", err.Error()).Send()
			break
		}
		time.Sleep(time.Second * 5)
	}
}
