package main

import (
	"fmt"
	"gameserver/config"
	"gameserver/server"
	"gameserver/utils"
)

func main() {
	go utils.InterruptHandle()
	fmt.Println("SERVER starting...")
	go server.StartMatcher(config.ServerListenAddress, config.TCPPort)
	fmt.Println("Matching service is on!")
	go server.GameRouter(config.ServerListenAddress, config.UDPPort)
	fmt.Println("Game router is on!")
	utils.Halt()
}
