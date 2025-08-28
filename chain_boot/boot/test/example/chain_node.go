package main

import (
	"web3_gui/chain_boot/boot/test"
)

func main() {
	test.InitConfig()
	ERR := test.StartChain() //
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}
