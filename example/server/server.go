package main

import (
	"fmt"
	"gameserver/config"
	"gameserver/game"
	"gameserver/utils"
)

func main() {
	fmt.Println("SERVER starting...")
	go game.StartMatcher(config.ServerListenAddress, config.TCPPort)
	fmt.Println("Matching service is on!")
	go game.GameRouter(config.ServerListenAddress, config.UDPPort)
	fmt.Println("Game router is on!")
	utils.Halt()
}
