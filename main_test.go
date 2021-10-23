package main

import (
	"encoding/binary"
	"net"
	"strconv"
	"testing"

	"github.com/Snak3War/SnakeWarServer/proto"
	"github.com/Snak3War/SnakeWarServer/server"
)

func TestServer(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:"+strconv.Itoa(server.DefServerPort))
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	var dataSizeBuf, data []byte
	dataSizeBuf = make([]byte, 4)

	conn.Read(dataSizeBuf)
	dataLen := binary.ByteOrder.Uint32(binary.LittleEndian, dataSizeBuf)
	data = make([]byte, dataLen)
	conn.Read(data)

	var pkt proto.Packet
	pkt.UnmarshalBinary(data)
	if pkt.Type != proto.PacketTypeHandshake {
		t.Fail()
	}
	data = data[pkt.Size():]
	var handshakePkt proto.Handshake
	handshakePkt.UnmarshalBinary(data)
	playerID := handshakePkt.ID
	playerCount := handshakePkt.Count
	seed := handshakePkt.Seed

	t.Logf("playerID: %v playerCount: %v seed: %v", playerID, playerCount, seed)

	conn.Read(dataSizeBuf)
	dataLen = binary.LittleEndian.Uint32(dataSizeBuf)
	data = make([]byte, dataLen)
	conn.Read(data)
	pkt.UnmarshalBinary(data)
	if pkt.Type != proto.PacketTypeEvent {
		t.Fail()
	}
	var e proto.Event
	data = data[pkt.Size():]
	e.UnmarshalBinary(data)
	if e.Type != proto.EventTypeStart {
		t.Fail()
	}

	data = proto.PacketToData(
		proto.Packet{
			Version: proto.Version,
			Type:    proto.PacketTypeEvent,
			Count:   1,
		},
		proto.EventMove{
			ID:        playerID,
			Direction: 1,
		},
	)
	binary.LittleEndian.PutUint32(dataSizeBuf, uint32(len(data)))
	conn.Write(dataSizeBuf)
	conn.Write(data)

	data, _ = proto.ReadBuffer(conn)
	pkts := proto.DataToPacket(data)
	pkt = pkts[0].(proto.Packet)
	if pkt.Type != proto.PacketTypeEvent {
		t.Fail()
	}
	e = pkts[1].(proto.Event)
	t.Log(e)
}
