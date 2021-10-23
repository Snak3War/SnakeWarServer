package server

import (
	"context"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/Snak3War/SnakeWarServer/proto"
)

// Config ...
type Config struct {
	Count   int
	Host    string
	Port    int
	TickClk int
}

// Server ...
type Server struct {
	cfg        *Config
	addr       *net.TCPAddr
	listener   *net.TCPListener
	handlers   []connectionHandler
	events     chan eventWithData
	tick       uint32
	stopSig    chan struct{}
	onlineMask uint64
}

const (
	DefCount      = 1
	DefHost       = "0.0.0.0"
	DefServerPort = 21739
	DefTickClk    = 17
)

func defaultConfig(cfg *Config) {
	if cfg.Count == 0 {
		cfg.Count = DefCount
	}
	if cfg.Host == "" {
		cfg.Host = DefHost
	}
	if cfg.Port == 0 {
		cfg.Port = DefServerPort
	}
	if cfg.TickClk == 0 {
		cfg.TickClk = DefTickClk
	}
}

// New snake war game server
func New(cfg *Config) *Server {
	defaultConfig(cfg)
	log.Printf("%+v\n", cfg)
	return &Server{
		cfg:      cfg,
		handlers: make([]connectionHandler, cfg.Count),
		events:   make(chan eventWithData, cfg.Count),
	}
}

func (s *Server) Stop() {
	s.stopSig <- struct{}{}
}

func (s *Server) Run() {
	var err error
	s.tick = 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.addr, err = net.ResolveTCPAddr("tcp", s.cfg.Host+":"+strconv.Itoa(s.cfg.Port))
	if err != nil {
		log.Fatalf("ResolveTCPAddr fail %v\n", err)
	}
	s.listener, err = net.ListenTCP("tcp", s.addr)
	defer s.listener.Close()
	if err != nil {
		log.Fatalf("ListenTCP fail %v\n", err)
	}
	log.Println("snake war server listen on", s.addr.String())

	seed := uint16(rand.Uint32())
	var targetMask uint64 = (1 << s.cfg.Count) - 1
	for s.onlineMask != targetMask {
		for i := 0; i < s.cfg.Count; i++ {
			if !s.handlers[i].Online {
				conn, err := s.listener.AcceptTCP()
				if err != nil {
					log.Printf("AcceptTCP fail %v\n", err)
					break
				}
				s.handlers[i] = *newConnectionHandler(s, i, conn, seed)
				go s.handlers[i].Run(ctx, s.events)
				defer s.handlers[i].Stop()
				s.onlineMask |= 1 << i
			} else {
				s.onlineMask &= ^(1 << i)
			}
		}
	}

	// send start game
	buf := proto.PacketToData(
		proto.Packet{Version: proto.Version, Type: proto.PacketTypeEvent, Count: 1},
		proto.Event{Tick: s.tick, Type: proto.EventTypeStart},
	)
	s.Broadcast(-1, buf)
	log.Println("game started")

	for {
		select {
		case <-time.After(time.Duration(s.cfg.TickClk) * time.Millisecond):
			if s.onlineMask == 0 {
				return
			}
			s.tick++
			s.notifyNextTick()
		case <-s.stopSig:
			return
		}
	}
}

func (s *Server) notifyNextTick() {
	var data []byte
	pkt := proto.Packet{
		Version: proto.Version,
		Type:    proto.PacketTypeEvent,
		Count:   0,
	}
Loop:
	for {
		var ed eventWithData
		select {
		case ed = <-s.events:
			pkt.Count++
			ed.Event.Tick = s.tick
			data = append(data, proto.PacketToData(ed.Event)...)
			data = append(data, ed.Data...)
		default:
			break Loop
		}
	}
	if pkt.Count == 0 {
		data = append(data, proto.PacketToData(
			proto.Event{
				Tick:     s.tick,
				Type:     proto.EventTypeHeartbeat,
				DataSize: 0,
			},
		)...)
		pkt.Count++
	}
	data = append(proto.PacketToData(pkt), data...)
	s.Broadcast(-1, data)
}

func (s *Server) Broadcast(except int, data []byte) {
	for i, v := range s.handlers {
		if i != except && v.Online {
			go func(i int) {
				n, err := s.handlers[i].Send(data)
				if err != nil || n != len(data) {
					log.Printf("%v: broadcast fail n=%v err=%v\n", v.conn.RemoteAddr().String(), n, err)
					s.onlineMask &= ^(1 << i)
				}
			}(i)
		}
	}
}
