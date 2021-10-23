package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Snak3War/SnakeWarServer/server"
)

func main() {
	tickclk := flag.Int("tickclk", server.DefTickClk, "tick clk")
	port := flag.Int("port", server.DefServerPort, "listen port")
	count := flag.Int("count", server.DefCount, "player count")
	logFile := flag.String("log", "", "log file")
	flag.Parse()

	if *logFile != "" {
		if logFileFp, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
			log.Fatal(err)
		} else {
			log.SetOutput(io.MultiWriter(os.Stdout, logFileFp))
		}
	}

	rand.Seed(time.Now().Unix())
	s := server.New(&server.Config{
		TickClk: *tickclk,
		Port:    *port,
		Count:   *count,
	})
	for {
		s.Run()
	}
}
