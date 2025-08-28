package main

import (
	"fmt"
	"web3_gui/chain/config"
)

func main() {
	example()
}

func example() {
	ok := config.CheckAddBlacklist(3, 5)
	fmt.Println("isok:", ok)
}
