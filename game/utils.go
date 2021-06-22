package game

import (
	"fmt"
	"gameserver/client"
	"gameserver/config"
	"strconv"
)

func getIP(addr string) string {
	separator := 0
	for i := range addr {
		if addr[i] == ':' {
			separator = i
		}
	}
	return addr[:separator]
}

func selectPort(c *client.Client) {
	udpPortInt, _ := strconv.Atoi(config.UDPPort)
	if simulation {
		c.Addr = fmt.Sprintf("%v:%v", getIP(c.Addr), udpPortInt+int(c.ClientID))
		return
	}
	c.Addr = getIP(c.Addr) + ":9090"
}
