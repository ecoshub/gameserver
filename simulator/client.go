package simulator

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"gameserver/config"
	"gameserver/frame"
	"gameserver/utils"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type SimulatedClient struct {
	ClientID uint16
	GameID   uint16
	ReadChan chan []byte
	GameOver *bool
}

func ClientSimulation(ip, TCPport, UDPport string) error {
	gameID, clientID, err := GameRequest(ip, TCPport)
	if err != nil {
		return err
	}
	gameover := false
	s := &SimulatedClient{
		GameID:   gameID,
		ClientID: clientID,
		ReadChan: make(chan []byte, 2048),
		GameOver: &gameover,
	}

	port, _ := strconv.Atoi(UDPport)
	if config.Simulation {
		port += int(clientID)
	}

	go s.ListenUDP(ip, port, clientID)

	registerPack := frame.CreateRegisterPack(s.GameID, s.ClientID)

	err = s.WriteEvent(ip, UDPport, registerPack)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s.waitForEvent(frame.Events.Start)
	fmt.Println("Game Started!")

	go s.CheckGameOver()

	for !*s.GameOver {
		time.Sleep(time.Second)
		err = s.WriteEvent(ip, UDPport, s.dummyEvent())
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Game Over!")

	return nil
}

func GameRequest(ip, port string) (uint16, uint16, error) {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		return 0, 0, err
	}
	_, err = conn.Write(initMessage())
	if err != nil {
		return 0, 0, err
	}
	fmt.Println("Game request registered. You are in the queue...")
	buffer := bufio.NewReader(conn)
	msg, err := utils.ReadNBytes(buffer, frame.PackSizeOf.GameID+frame.PackSizeOf.ClientID)
	if err != nil {
		return 0, 0, err
	}
	gameID := binary.LittleEndian.Uint16(msg[:frame.PackSizeOf.GameID])
	clientID := binary.LittleEndian.Uint16(msg[frame.PackSizeOf.GameID : frame.PackSizeOf.GameID+frame.PackSizeOf.ClientID])
	fmt.Printf("[in game] gameID: %v, clientID: %v\n", gameID, clientID)
	return gameID, clientID, err
}

func (s *SimulatedClient) WriteEvent(ip, UDPport string, p *frame.Packet) error {
	pack := frame.PacketToBytes(p)
	fmt.Printf("[Register][GID:%v, CID:%v] >>> [pack: %v]\n", s.GameID, s.ClientID, pack)
	err := s.WriteUDP(ip, UDPport, pack)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (s *SimulatedClient) WriteUDP(ip, port string, packet []byte) error {
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		return err
	}
	_, err = conn.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

func (s *SimulatedClient) ListenUDP(ip string, port int, clientID uint16) error {
	// udpPort, _ := strconv.Atoi(config.UDPPort)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(config.ServerListenAddress),
	})
	if err != nil {
		return err
	}
	for {
		buffer := make([]byte, 2048)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}
		s.ReadChan <- buffer[:n]
	}
}

func (s *SimulatedClient) CheckGameOver() {
	for {
		buffer := <-s.ReadChan
		pack := frame.BytesToPacket(buffer)
		fmt.Printf("[Game Start][GID:%v, CID:%v] <<< [pack: %v]\n", s.ClientID, s.GameID, buffer)
		if pack.IsEventPack(frame.Events.GameOver) {
			break
		}
	}
	*s.GameOver = true
}

func (s *SimulatedClient) dummyEvent() *frame.Packet {
	noe := rand.Intn(4)
	events := make([]*frame.Event, noe)
	for i := range events {
		events[i] = &frame.Event{
			ID:   uint8(rand.Intn(256)),
			Data: int32(rand.Int31()),
		}
	}
	return &frame.Packet{
		ClientID:  s.ClientID,
		GameID:    s.GameID,
		Events:    events,
		TimeStamp: time.Now(),
	}
}

func initMessage() []byte {
	return utils.CreateHash()
}

func (s *SimulatedClient) waitForEvent(e uint8) {
	for {
		buffer := <-s.ReadChan
		pack := frame.BytesToPacket(buffer)
		if pack.IsEventPack(e) {
			fmt.Printf("[Game Start][GID:%v, CID:%v] <<< [pack: %v]\n", s.ClientID, s.GameID, buffer)
			break
		}
	}
}
