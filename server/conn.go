package server

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"net"

	"github.com/Snak3War/SnakeWarServer/proto"
)

type connectionHandler struct {
	Online  bool
	running bool
	id      int
	conn    *net.TCPConn
	stopSig chan struct{}
	server  *Server
	seed    uint16
}

type eventWithData struct {
	proto.Event
	Data []byte
}

func newConnectionHandler(s *Server, id int, conn *net.TCPConn, seed uint16) *connectionHandler {
	return &connectionHandler{
		server:  s,
		Online:  true,
		id:      id,
		conn:    conn,
		stopSig: make(chan struct{}),
		seed:    seed,
	}
}

func (c *connectionHandler) Stop() {
	if c.running {
		c.stopSig <- struct{}{}
	}
	if c.Online {
		c.Online = false
	}
}

func (c *connectionHandler) Send(data []byte) (n int, err error) {
	if !c.Online {
		return 0, errors.New("handler die")
	}
	var bufSize []byte = make([]byte, 4)
	binary.LittleEndian.PutUint32(bufSize, uint32(len(data)))
	if n, err = c.conn.Write(bufSize); n != 4 || err != nil {
		c.Stop()
		return 0, err
	}
	n, err = c.conn.Write(data)
	return
}

func (c *connectionHandler) Run(ctx context.Context, events chan eventWithData) {
	c.running = true
	defer func() {
		c.running = false
		c.Online = false
	}()
	defer c.conn.Close()
	c.handshake()
	for {
		var size chan uint32
		go func() {
			var buf []byte = make([]byte, 4)
			n, err := c.conn.Read(buf)
			if n != 4 || err != nil {
				log.Println("connection lost", n, err)
				c.Stop()
				return
			}
			size <- binary.ByteOrder.Uint32(binary.LittleEndian, buf)
		}()
		var toRead uint32
		select {
		case <-c.stopSig:
			return
		case <-ctx.Done():
			return
		case toRead = <-size:
			var buf []byte = make([]byte, toRead)
			n, err := c.conn.Read(buf)
			if uint32(n) != toRead || err != nil {
				log.Printf("%v: invalid packet n=%v err=%v\n", c.conn.RemoteAddr().String(), n, err)
				break
			}
			e, err := c.handle(buf)
			if err == nil {
				events <- eventWithData{Event: e, Data: buf[e.Size():]}
			} else {
				return
			}
		}
	}
}

func (c *connectionHandler) handshake() {
	c.Send(proto.PacketToData(
		proto.Packet{
			Version: proto.Version,
			Type:    proto.PacketTypeHandshake,
			Count:   1,
		},
		proto.Handshake{
			Seed:  c.seed,
			ID:    byte(c.id),
			Count: byte(c.server.cfg.Count),
		},
	))
}

func (c *connectionHandler) handle(data []byte) (e proto.Event, err error) {
	var pkt proto.Packet
	if err = pkt.UnmarshalBinary(data); err != nil {
		log.Printf("%v: invalid packet err=%v\n", c.conn.RemoteAddr().String(), err)
		return
	}
	if pkt.Type != proto.PacketTypeEvent {
		log.Printf("%v: packet type not supportted type=%v\n", c.conn.RemoteAddr().String(), pkt.Type)
		return
	}

	data = data[pkt.Size():]
	if err = e.UnmarshalBinary(data); err != nil {
		log.Printf("%v: invalid packet err=%v\n", c.conn.RemoteAddr().String(), err)
		return
	}
	return
}
