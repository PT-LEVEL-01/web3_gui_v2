package main

import (
	"fmt"
	"web3_gui/chain/cache"
)

func main() {
	b58 := cache.To58String([]byte("ok"))
	fmt.Println(b58)
}
