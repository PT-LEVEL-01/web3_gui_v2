package grpc

import "google.golang.org/grpc"

const (
	TronGrpcMainNetWorkAddr = "grpc.trongrid.io:50051" //之前用的ankr
)

func main() {
	simple()
}

func simple() error {
	//conn, err := grpc.NewClient(TronGrpcMainNetWorkAddr, grpc.WithInsecure())
	conn, err := grpc.Dial(TronGrpcMainNetWorkAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
}
