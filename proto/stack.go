package proto

import (
	"encoding"
	"encoding/binary"
	"log"
	"net"
)

func PacketToData(pkts ...encoding.BinaryMarshaler) (data []byte) {
	var err error
	for _, pkt := range pkts {
		var buf []byte
		if buf, err = pkt.MarshalBinary(); err != nil {
			log.Fatalf("PacketToData: %v", err)
			return
		} else {
			data = append(data, buf...)
		}
	}
	return
}

func DataToPacket(data []byte) (pkts []interface{}) {
	var pkt Packet
	pkt.UnmarshalBinary(data)
	pkts = append(pkts, pkt)
	if len(data) > pkt.Size() {
		data = data[pkt.Size():]
	}
	switch pkt.Type {
	case PacketTypeHandshake:
		var handshakePkt Handshake
		handshakePkt.UnmarshalBinary(data)
		pkts = append(pkts, handshakePkt)
	case PacketTypeEvent:
		var e Event
		e.UnmarshalBinary(data)
		pkts = append(pkts, e)
		if len(data) > e.Size() {
			data = data[e.Size():]
		}
		switch e.Type {
		case EventTypeMove:
			var em EventMove
			em.UnmarshalBinary(data)
			pkts = append(pkts, em)
		case EventTypeOver:
			var eo EventOver
			eo.UnmarshalBinary(data)
			pkts = append(pkts, eo)
		}
	}
	return
}

func ReadBuffer(conn *net.TCPConn) (data []byte, err error) {
	var sizeBuf []byte = make([]byte, 4)
	if _, err = conn.Read(sizeBuf); err != nil {
		return nil, err
	}
	data = make([]byte, binary.LittleEndian.Uint32(sizeBuf))
	_, err = conn.Read(data)
	return
}

func WriteBuffer(conn *net.TCPConn, data []byte) (n int, err error) {
	var sizeBuf []byte = make([]byte, 4)
	if _, err = conn.Write(sizeBuf); err != nil {
		return 0, err
	}
	return conn.Write(data)
}
