package rpc_server

import (
	"context"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/libp2parea/adapter/engine"
)

type RpcServer struct {
	chain      *mining.Chain
	grpcServer *grpc.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewRpcServer() *RpcServer {
	server := grpc.NewServer()
	chain := mining.GetLongChain()
	return &RpcServer{
		chain:      chain,
		grpcServer: server,
	}
}
func (s *RpcServer) Start() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	//注册api
	s.RegisterApi()
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(int(config.RpcPort)))
	if err != nil {
		engine.Log.Error("tcp监听失败:%s", err.Error())
	}
	engine.Log.Info("rpc server开始启动 listen on %v", ":"+strconv.Itoa(int(config.RpcPort)))
	go func() {
		err = s.grpcServer.Serve(conn)
		if err != nil {
			engine.Log.Error("grpc start失败:%s", err.Error())
		}
	}()
}
func (s *RpcServer) Stop() {
	s.cancel()
	s.grpcServer.GracefulStop()
}
func (s *RpcServer) RegisterApi() {
	go_protos.RegisterSubscriberServer(s.grpcServer, NewSubscriberService(s.ctx, s.chain))
}
