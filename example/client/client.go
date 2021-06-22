package main

import (
	"gameserver/config"
	"gameserver/simulator"
)

func main() {
	simulator.ClientSimulation(config.ClientRequestAddress, config.TCPPort, config.UDPPort)
}
