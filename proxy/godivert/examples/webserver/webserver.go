package main

import "godivert/utils"

func main() {
	utils.RunServer()
	select {}
}
