package main

import (
	"flag"
	"log"

	"github.com/diceclone/server"
)

var (
	configHost string
	configPort int
)

func setUpFlags() {
	flag.StringVar(&configHost, "host", "0.0.0.0", "host for dicedb server")
	flag.IntVar(&configPort, "port", 7379, "port for dicedb server")

	flag.Parse()
}

func main() {
	setUpFlags()
	log.Println("rolling the dice")

	server.RunSyncTCPServer("0.0.0.0", 7379)
}
