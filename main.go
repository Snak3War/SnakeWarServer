package main

import (
	"flag"

	"github.com/Snak3War/SnakeWarServer/server"
)

func main() {
	tickclk := flag.Int("tickclk", server.DefTickClk, "tick clk")
	port := flag.Int("port", server.DefServerPort, "listen port")
	count := flag.Int("count", server.DefCount, "player count")
	flag.Parse()

	s := server.New(&server.Config{
		TickClk: *tickclk,
		Port:    *port,
		Count:   *count,
	})
	for {
		s.Run()
	}
}
