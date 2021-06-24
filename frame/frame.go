package frame

import (
	"encoding/binary"
	"errors"
	"gameserver/config"
	"gameserver/utils"
	"time"
)

type Event struct {
	ID   uint8
	Data int32
}
type Packet struct {
	ClientID  uint16
	GameID    uint16
	Events    []*Event
	TimeStamp time.Time
}

var (
	Events = struct {
		Register   uint8
		Start      uint8
		Data       uint8
		End        uint8
		Disconnect uint8
		GameOver   uint8
	}{
		Data:       0,
		Register:   1,
		Start:      2,
		End:        3,
		Disconnect: 254,
		GameOver:   255,
	}

	EventName map[uint8]string = map[uint8]string{
		Events.Data:       "data",
		Events.Register:   "register",
		Events.Start:      "start",
		Events.End:        "end",
		Events.Disconnect: "disconnect",
		Events.GameOver:   "gameover",
	}

	PackSizeOf = struct {
		ClientID      int
		GameID        int
		EventID       int
		numberOfEvent int
		Data          int
		TimeStamp     int
	}{
		ClientID:      2,
		GameID:        2,
		numberOfEvent: 1,
		EventID:       1,
		Data:          4,
		TimeStamp:     8,
	}

	MinPacketSize   int = PackSizeOf.ClientID + PackSizeOf.GameID + PackSizeOf.numberOfEvent + PackSizeOf.EventID + PackSizeOf.Data + PackSizeOf.TimeStamp
	EventPacketSize int = PackSizeOf.EventID + PackSizeOf.Data
	MaxPacketSize   int = MinPacketSize + EventPacketSize*255

	ErrInvalidEventPacket error = errors.New("invalid event data packet size")
)

// Custom data package system for game event communication
// |-------------------------------------------------------------------------------------------
// |                header              |                events...                |  time     |
// |-------------------------------------------------------------------------------------------
// |clientID | gameID | number of event | eventID  |  data | eventID  |  data |...| timeStamp |
// |-------------------------------------------------------------------------------------------
// |  2byte  | 2byte  |     1byte       |   1byte  | 4byte |   1byte  | 4byte |...|   8byte   |
// |  16bit  | 16bit  |     8bit        |   8bit   | 32bit |   8bit   | 32bit |...|   64bit   |
// |-------------------------------------------------------------------------------------------

func Unmarshal(p []byte) *Packet {
	clientID := GetClientID(p)
	gameID := GetGameID(p)
	events := GetEvents(p)
	timeStamp := GetTimeStamp(p)
	return &Packet{
		ClientID:  clientID,
		GameID:    gameID,
		Events:    events,
		TimeStamp: timeStamp,
	}
}

func GetClientID(packet []byte) uint16 {
	clientIDPack := packet[:PackSizeOf.ClientID]
	return binary.LittleEndian.Uint16(clientIDPack)
}

func GetGameID(packet []byte) uint16 {
	gameIDPack := packet[PackSizeOf.ClientID : PackSizeOf.ClientID+PackSizeOf.GameID]
	return binary.LittleEndian.Uint16(gameIDPack)
}

func GetNOF(packet []byte) uint8 {
	numberOfEventPack := packet[PackSizeOf.ClientID+PackSizeOf.GameID : PackSizeOf.ClientID+PackSizeOf.GameID+PackSizeOf.numberOfEvent]
	return uint8(numberOfEventPack[0])
}

func GetEvents(packet []byte) []*Event {
	n := GetNOF(packet)
	events := make([]*Event, 0, n)
	startByte := PackSizeOf.ClientID + PackSizeOf.GameID + PackSizeOf.numberOfEvent
	for i := uint8(0); i < n; i++ {
		eventIDPack := packet[startByte : startByte+PackSizeOf.EventID]
		dataPack := packet[startByte+PackSizeOf.EventID : startByte+PackSizeOf.TimeStamp]
		data := binary.LittleEndian.Uint32(dataPack)
		e := &Event{
			ID:   eventIDPack[0],
			Data: int32(data),
		}
		events = append(events, e)
		startByte += EventPacketSize
	}
	return events
}

func GetTimeStamp(packet []byte) time.Time {
	timePack := packet[len(packet)-PackSizeOf.TimeStamp:]
	timeRead := binary.LittleEndian.Uint64(timePack)
	return time.Unix(0, int64(timeRead))
}

func Marshal(p *Packet) []byte {
	buffer := make([]byte, 0, MaxPacketSize)

	clientIDBytes, _ := utils.ToBytes(p.ClientID)
	clientIDBytes = clientIDBytes[:PackSizeOf.ClientID]
	buffer = append(buffer, clientIDBytes...)

	gameIDBytes, _ := utils.ToBytes(p.GameID)
	gameIDBytes = gameIDBytes[:PackSizeOf.GameID]
	buffer = append(buffer, gameIDBytes...)

	buffer = append(buffer, uint8(len(p.Events)))

	for _, e := range p.Events {
		buffer = append(buffer, eventToBytes(e)...)
	}

	t, _ := utils.ToBytes(p.TimeStamp.UnixNano())
	t = t[:PackSizeOf.TimeStamp]
	buffer = append(buffer, t...)

	return buffer
}

func (p *Packet) IsEventPack(eventID uint8) bool {
	if len(p.Events) == 1 {
		if p.Events[0].ID == eventID {
			return true
		}
	}
	return false
}

func eventToBytes(e *Event) []byte {
	pack := make([]byte, 0, EventPacketSize)
	pack = append(pack, e.ID)
	data, _ := utils.ToBytes(e.Data)
	data = data[:PackSizeOf.Data]
	pack = append(pack, data...)
	return pack
}

func PackGameIDAndClientID(gameID, clientID uint16) []byte {
	gameIDPack, _ := utils.ToBytes(gameID)
	clientIDPack, _ := utils.ToBytes(clientID)
	pack := make([]byte, 0, PackSizeOf.GameID+PackSizeOf.ClientID)
	pack = append(pack, gameIDPack[:PackSizeOf.GameID]...)
	pack = append(pack, clientIDPack[:PackSizeOf.ClientID]...)
	return pack
}

func CreateEventPacket(gameID uint16, event uint8, data int32) *Packet {
	return &Packet{
		ClientID: config.ServerID,
		GameID:   gameID,
		Events: []*Event{
			{
				ID:   event,
				Data: data,
			},
		},
		TimeStamp: time.Now(),
	}
}

func CreatePack(gameID, clientID uint16, event uint8) *Packet {
	return &Packet{
		ClientID: clientID,
		GameID:   gameID,
		Events: []*Event{
			{
				ID:   event,
				Data: int32(clientID),
			},
		},
		TimeStamp: time.Now(),
	}
}

func IsValid(packet []byte) bool {
	if len(packet) < MinPacketSize {
		return false
	}
	if len(packet) > MaxPacketSize {
		return false
	}
	if (len(packet) - MinPacketSize%EventPacketSize) != 0 {
		return false
	}
	return true
}
