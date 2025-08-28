package main

import (
	"web3_gui/chain_boot/boot/hive"
	"web3_gui/chain_boot/boot/test"
)

func main() {
	StartChain()
}

func StartChain() {
	hive.InitConfig()
	ERR := test.StartChain() // .StartWithArea("")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}
