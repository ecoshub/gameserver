package server

import (
	"errors"
	"fmt"
	"gameserver/client"
	"gameserver/config"
	"gameserver/frame"
	"gameserver/utils"
	"log"
	"net"
	"strconv"
)

func (s *Server) GameRouter(ip, port string) {
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
	s.gameRoutine(conn)
}

func (s *Server) gameRoutine(conn *net.UDPConn) {
	for {
		buff := make([]byte, frame.MaxPacketSize)
		n, addr, err := conn.ReadFrom(buff)
		if err != nil {
			log.Println(err)
			continue
		}
		buff = buff[:n]
		if frame.IsValid(buff) {
			log.Println(frame.ErrInvalidEventPacket)
			continue
		}
		go s.eventRouter(buff, addr.String())
	}
}

func (s *Server) eventRouter(buffer []byte, addr string) {
	gameID := frame.GetGameID(buffer)
	players, exists := s.gameLobby[gameID]
	if !exists {
		// handle later
		// log.Printf("event comming from GID: %v.", gameID)
		return
	}

	pack := frame.Unmarshal(buffer)
	if pack.IsEventPack(frame.Events.Register) {
		player, err := selectPlayer(players, pack.ClientID)
		if err != nil {
			fmt.Println(err)
			return
		}
		// register UDP address
		registerPlayer(player, addr)
		if s.checkAllPlayerRegistered(players, pack.GameID) {
			log.Println(">>> Sending game started event")
			startEventPack := frame.CreateEventPacket(pack.GameID, frame.Events.Start, config.NullData)
			s.broadCastWithGameID(startEventPack)
			go s.simulateGameover(pack.GameID)
		}
		return
	}

	if pack.IsEventPack(frame.Events.Disconnect) {
		s.broadCastWithGameID(frame.CreateEventPacket(gameID, frame.Events.GameOver, config.NullData))
		return
	}

	someDataManipulationAndCorrectionProcess(pack)
	s.broadCastWithGameID(pack)
}

func (s *Server) broadCastWithGameID(p *frame.Packet) {
	packet := frame.Marshal(p)
	players := s.gameLobby[p.GameID]
	for _, p := range players {
		if !p.IsRegistered() {
			log.Println("error. Broadcast to unattached connection")
			return
		}

		// I need to change client udp ports because.
		// Simulation in same computer would be impossible all client has same ip and same port
		utils.SelectPort(p)

		// NOTE an attemp system might be good
		err := UDPSend(packet, p.Addr)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func registerPlayer(player *client.Client, addr string) {
	player.Addr = addr
	player.UDPRegistered = true
	log.Printf("Client UDP register success, client ID: %v\n", player.ClientID)
}

func selectPlayer(players []*client.Client, clientID uint16) (*client.Client, error) {
	for _, p := range players {
		if p.ClientID == clientID {
			return p, nil
		}
	}
	return nil, errors.New("player not found")
}

func (s *Server) checkAllPlayerRegistered(players []*client.Client, gameID uint16) bool {
	if s.gameState[gameID] {
		return false
	}
	for _, p := range players {
		if !p.UDPRegistered {
			return false
		}
	}
	s.gameState[gameID] = true
	return true
}

func (s *Server) simulateGameover(gameID uint16) {
	utils.RandomSleepMillisecond(config.MinGameOverTime, config.MaxGameOverTime)
	s.broadCastWithGameID(frame.CreateEventPacket(gameID, frame.Events.GameOver, config.NullData))
	fmt.Printf("# Game %v ended.\n", gameID)
	delete(s.gameLobby, gameID)
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

func someDataManipulationAndCorrectionProcess(p *frame.Packet) {
	log.Printf(">> incoming data from, gameID: %v, client: %v\n", p.GameID, p.ClientID)
}
