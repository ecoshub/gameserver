package test

import (
	"gameserver/event"
	"gameserver/utils"
	"reflect"
	"testing"
	"time"
)

func TestEventPackage(t *testing.T) {
	events := []*event.Event{
		{
			ID:   72,
			Data: 100,
		},
		{
			ID:   31,
			Data: 1512,
		},
		{
			ID:   70,
			Data: 2014,
		},
	}

	packObject := &event.Packet{
		ClientID:  31,
		GameID:    1555,
		Events:    events,
		TimeStamp: time.Now(),
	}

	// to packet
	pack := event.PacketToBytes(packObject)

	// to packet object
	newPacket := event.BytesToPacket(pack)
	packObject.TimeStamp = packObject.TimeStamp.Round(time.Nanosecond * 100)
	newPacket.TimeStamp = newPacket.TimeStamp.Round(time.Nanosecond * 100)

	if !reflect.DeepEqual(packObject, newPacket) {
		t.Log("ERROR: event object -> event packet failed")
		t.Log("incoming object:")
		t.Log(utils.SprintStruct(packObject))
		t.Log("unpacked object:")
		t.Log(utils.SprintStruct(newPacket))
		t.Log(pack)
		t.Fail()
	}

}
