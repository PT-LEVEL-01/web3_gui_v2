package libp2parea

//111111
import (
	"net"
	"strconv"
	"web3_gui/libp2parea/v1/socks5"
	"web3_gui/utils"
)

func InitSocks5Server() {
	port := 1080
	var listener net.Listener
	var err error
	for {
		listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(port)))
		if err != nil {
			port += 1
			continue
		}
		break
	}
	utils.Log.Debug().Msgf("socks5服务监听端口 %d", port)

	defer listener.Close()
	if server, err := socks5.NewSocks5Server(socks5.Direct); err == nil {
		server.Serve(listener)
	}
}
