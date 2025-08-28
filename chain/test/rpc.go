package main

import (
	"fmt"
	"web3_gui/chain/rpc"
)

func main() {
	fmt.Println("start...")
	rpc.RegisterRpcServer()
}
