package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	var wg sync.WaitGroup
	wg.Add(2)

	var c chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// server.RunSyncTCPServer(config.Host, config.Port)
	go server.RunAsyncTCPServer(&wg)
	go server.WaitForSignal(&wg, c)

	wg.Wait()
}
