package simulator

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"gameserver/event"
	"gameserver/utils"
	"math/rand"
	"net"
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

	go s.ListenUDP(ip, UDPport, clientID)

	registerPack := event.CreateRegisterPack(s.GameID, s.ClientID)

	err = s.WriteEvent(ip, UDPport, registerPack)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s.waitForEvent(event.Events.Start)
	fmt.Println("Game Started!")

	go s.CheckGameOver()

	for !*s.GameOver {
		time.Sleep(time.Second)
		// time.Sleep(randomEventTime())
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
	msg, err := utils.ReadNBytes(buffer, event.PackSizeOf.GameID+event.PackSizeOf.ClientID)
	if err != nil {
		return 0, 0, err
	}
	gameID := binary.LittleEndian.Uint16(msg[:event.PackSizeOf.GameID])
	clientID := binary.LittleEndian.Uint16(msg[event.PackSizeOf.GameID : event.PackSizeOf.GameID+event.PackSizeOf.ClientID])
	fmt.Printf("[in game] gameID: %v, clientID: %v\n", gameID, clientID)
	return gameID, clientID, err
}

func (s *SimulatedClient) WriteEvent(ip, UDPport string, p *event.Packet) error {
	pack := event.PacketToBytes(p)
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

func (s *SimulatedClient) ListenUDP(ip, port string, clientID uint16) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 9090 + int(clientID),
		IP:   net.ParseIP("0.0.0.0"),
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
		pack := event.BytesToPacket(buffer)
		fmt.Printf("[Game Start][GID:%v, CID:%v] <<< [pack: %v]\n", s.ClientID, s.GameID, buffer)
		if len(pack.Events) == 1 {
			if pack.Events[0].ID == event.Events.GameOver {
				break
			}
		}
	}
	*s.GameOver = true
}

func (s *SimulatedClient) dummyEvent() *event.Packet {
	noe := rand.Intn(4)
	events := make([]*event.Event, noe)
	for i := range events {
		events[i] = &event.Event{
			ID:   uint8(rand.Intn(256)),
			Data: int32(rand.Int31()),
		}
	}
	return &event.Packet{
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
		pack := event.BytesToPacket(buffer)
		if len(pack.Events) == 1 {
			if pack.Events[0].ID == e {
				fmt.Printf("[Game Start][GID:%v, CID:%v] <<< [pack: %v]\n", s.ClientID, s.GameID, buffer)
				break
			}
		}
	}
}
