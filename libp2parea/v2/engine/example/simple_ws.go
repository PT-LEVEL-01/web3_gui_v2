package main

import (
	"bytes"
	"encoding/binary"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"web3_gui/utils"
)

const wsServerPort = 8080

func main() {
	utils.Log.Info().Str("start", "").Send()
	go StartWSServer()
	utils.Log.Info().Str("start server", "").Send()
	for i := range 10 {
		utils.Log.Info().Int("start client", i).Send()
		StartWSClient()
	}
}

func StartWSServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(wsServerPort))
	if err != nil {
		utils.Log.Error().Str("err", err.Error()).Send()
		return
	}
	lis, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	ws := WsServer{}
	err = http.Serve(lis, &ws)
	if err != nil {
		utils.Log.Error().Err(err).Send()
	}

}

type WsServer struct {
}

func (this *WsServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	//websocket连接
	if req.URL.String() == "/ws" {
		upgrader := &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// 我们需要检查连接的来源
			// 这将允许我们根据我们的 React 发出请求
			// 现在,我们将不检查就允许任何连接，允许跨域请求
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		ws, err := upgrader.Upgrade(resp, req, nil)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return
		}

		for {
			bs1 := make([]byte, 4+1)
			_, err = io.ReadFull(ws.NetConn(), bs1)
			if err != nil {
				utils.Log.Error().Err(err).Send()
				return
			}
			length := binary.BigEndian.Uint32(bs1[:4])
			bs := make([]byte, length)
			_, err = io.ReadFull(ws.NetConn(), bs)
			if err != nil {
				utils.Log.Error().Err(err).Send()
				return
			}
			utils.Log.Info().Str("服务器收到", string(bs)).Send()
			_, err := ws.NetConn().Write(bs)
			if err != nil {
				utils.Log.Error().Err(err).Send()
				return
			}
		}
	}
}

func StartWSClient() {
	addr := "127.0.0.1:" + strconv.Itoa(int(wsServerPort))
	wsDsn, err := url.Parse("ws://" + addr + "/ws")
	conn, _, err := websocket.DefaultDialer.Dial(wsDsn.String(), nil)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}

	hello := []byte("hello")

	buf := bytes.NewBuffer(nil)
	//发送4字节包大小
	err = binary.Write(buf, binary.BigEndian, uint32(len(hello)))
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	//再发送1字节协议号
	err = binary.Write(buf, binary.BigEndian, uint8(1))
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	//再发送包内容
	_, err = buf.Write(hello)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	//utils.Log.Info().Hex("Send data", buf.Bytes()).Send()
	_, err = conn.NetConn().Write(buf.Bytes())
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}

}
