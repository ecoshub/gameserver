package main

import (
	"gameserver/config"
	"gameserver/game"
	"gameserver/simulator"
	"gameserver/utils"
)

func main() {
	startServerSimulation()
}

func startServerSimulation() {
	go game.StartMatcher(config.ServerListenAddress, config.TCPPort)
	go game.GameRouter(config.ServerListenAddress, config.UDPPort)
	testClientSize := 2
	for i := 0; i < testClientSize; i++ {
		utils.RandomSleepMillisecond(500, 2500)
		go simulator.ClientSimulation(config.ClientRequestAddress, config.TCPPort, config.UDPPort)
	}
	utils.Halt()
}
