package main

import (
	"net"
	"strconv"
	"time"
	"web3_gui/libp2parea/v2/v2/engine"
	"web3_gui/utils"
)

func main() {
	tcpClient()
}

func tcpClient() {
	//ip := "45.32.227.1"
	ip := "127.0.0.1"
	port := 8848
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(int(port)), time.Second*3)
	if err != nil {
		utils.Log.Error().Str("connect error", err.Error()).Send()
		//Log.Error("connect error:%s", err.Error())
		return
	}
	go tcpWriteClient(conn)
	tcpReadClient(conn)
}

func tcpReadClient(conn net.Conn) {
	for {
		bs, err := engine.ReadStream(conn)
		if err != nil {
			utils.Log.Error().Str("error", err.Error()).Send()
			break
		}
		utils.Log.Debug().Str("data", string(bs)).Send()
	}
}

func tcpWriteClient(conn net.Conn) {
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
