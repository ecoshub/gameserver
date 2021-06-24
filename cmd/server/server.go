package main

import (
	"fmt"
	"gameserver/config"
	"gameserver/server"
	"gameserver/utils"
)

func main() {

	s := server.NewServer()

	// to catch SIGINT SIGTERM SIGQUIT
	// when an interrupt occurred
	// server is sending a gameover event to all clients
	go s.InterruptHandle()

	// tcp server
	fmt.Println("SERVER starting...")
	go s.StartMatcher(config.ServerListenAddress, config.TCPPort)

	// udp server
	fmt.Println("Matching service is on!")
	go s.GameRouter(config.ServerListenAddress, config.UDPPort)

	fmt.Println("Game router is on!")
	// to block main thread
	utils.Halt()
}
