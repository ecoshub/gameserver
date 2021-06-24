package main

import (
	"fmt"
	"gameserver/config"
	"gameserver/server"
	"gameserver/utils"
)

func main() {

	s := server.NewServer()

	go s.InterruptHandle()
	fmt.Println("SERVER starting...")
	go s.StartMatcher(config.ServerListenAddress, config.TCPPort)
	fmt.Println("Matching service is on!")
	go s.GameRouter(config.ServerListenAddress, config.UDPPort)
	fmt.Println("Game router is on!")
	utils.Halt()

}
