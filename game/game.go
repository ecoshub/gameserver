package game

import (
	"fmt"
	"gameserver/client"
	"gameserver/config"
	"gameserver/event"
	"gameserver/utils"
	"log"
	"net"
	"strconv"
)

var (
	simulation bool = true
)

func GameRouter(ip, port string) {
	udpPort, _ := strconv.Atoi(config.UDPPort)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: udpPort,
		IP:   net.ParseIP(config.ServerListenAddress),
	})
	if err != nil {
		// listen error must be handle.
		// it can control by an error channel
		log.Println(err)
		return
	}
	defer conn.Close()
	gameRoutine(conn)
}

func gameRoutine(conn *net.UDPConn) {
	for {
		buff := make([]byte, event.MaxPacketSize)
		n, addr, err := conn.ReadFrom(buff)
		if err != nil {
			log.Println(err)
			continue
		}
		buff = buff[:n]
		if event.IsValid(buff) {
			log.Println(event.ErrInvalidEventPacket)
			continue
		}
		go eventRouter(buff, addr.String())
	}
}

func eventRouter(buffer []byte, addr string) {
	gameID := event.GetGameID(buffer)
	players, exists := gameList[gameID]
	if !exists {
		log.Printf("there is no game with ID: %v, package must be broken.", gameID)
		return
	}
	pack := event.BytesToPacket(buffer)
	if IsUDPRegisterRequest(players, pack, addr) {
		return
	}
	someDataManipulationAndCorrectionProcess(pack)
	broadCastWithGameID(pack)
}

func broadCastWithGameID(p *event.Packet) {
	packet := event.PacketToBytes(p)
	players := gameList[p.GameID]
	for _, p := range players {
		if !p.UDPRegistered {
			log.Println("broadcast to unattached connection")
			return
		}
		if simulation {
			p.Addr = getIP(p.Addr) + ":" + strconv.Itoa(9090+int(p.ClientID))
		} else {
			p.Addr = getIP(p.Addr) + ":9090"
		}
		err := UDPSend(packet, p.Addr)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func IsUDPRegisterRequest(players []*client.Client, pack *event.Packet, addr string) bool {
	if len(pack.Events) == 1 {
		if pack.Events[0].ID == event.Events.Register {
			allAttached := true
			for _, p := range players {
				if p.ClientID == pack.ClientID {
					if p.UDPRegistered {
						continue
					}
					p.Addr = addr
					p.UDPRegistered = true
					log.Printf("connection attached, client ID: %v\n", p.ClientID)
				}
				if !p.UDPRegistered {
					allAttached = false
				}
			}
			if allAttached {
				log.Println("sending game started event")
				broadCastWithGameID(event.CreateEventPacket(pack.GameID, event.Events.Start, 0))
				// gameover condition simulator
				go func() {
					utils.RandomSleepMillisecond(10000, 15000)
					broadCastWithGameID(event.CreateEventPacket(pack.GameID, event.Events.GameOver, 0))
				}()
			}
			return true
		}
	}
	return false
}

func UDPSend(msg []byte, addr string) error {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}
	return nil
}

func someDataManipulationAndCorrectionProcess(p *event.Packet) {
	fmt.Printf("data processing, gameID: %v\n", p.GameID)
}
