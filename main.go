package main

import (
	"flag"
	"log"

	"github.com/diceclone/config"
	"github.com/diceclone/server"
)

func setUpFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for dicedb server")
	flag.IntVar(&config.Port, "port", 7379, "port for dicedb server")

	flag.Parse()
}

func main() {
	setUpFlags()
	log.Println("rolling the dice")

	server.RunSyncTCPServer("0.0.0.0", 7379)
}
