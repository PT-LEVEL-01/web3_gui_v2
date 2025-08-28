package main

import (
	chainboot "web3_gui/chain_boot/boot/test"
	"web3_gui/chain_boot/boot/three"
)

func main() {
	three.InitConfig()
	ERR := chainboot.StartChain() //
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}
