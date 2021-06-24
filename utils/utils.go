package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gameserver/client"
	"gameserver/config"
	"math/rand"
	"strconv"
	"time"
)

func PrintStruct(i interface{}) {
	enc, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(enc))
}

func SprintStruct(i interface{}) string {
	enc, _ := json.MarshalIndent(i, "", "  ")
	return string(enc)
}

func RandomSleepMillisecond(min, max int) {
	t := rand.Intn(max-min) + min
	time.Sleep(time.Millisecond * time.Duration(t))
}

func Halt() {
	for {
		time.Sleep(time.Hour)
	}
}

func ReadNBytes(b *bufio.Reader, n int) ([]byte, error) {
	buffer := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		singleByte, err := b.ReadByte()
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, singleByte)
	}
	return buffer, nil
}

func GetIP(addr string) string {
	separator := 0
	for i := range addr {
		if addr[i] == ':' {
			separator = i
		}
	}
	return addr[:separator]
}

func SelectPort(c *client.Client) {
	udpPortInt, _ := strconv.Atoi(config.UDPPort)
	if config.Simulation {
		c.Addr = fmt.Sprintf("%v:%v", GetIP(c.Addr), udpPortInt+int(c.ClientID))
		return
	}
	c.Addr = GetIP(c.Addr) + ":9090"
}
