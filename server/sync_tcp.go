package server

import (
	"io"
	"log"
	"net"
	"strconv"
)

func RunSyncTCPServer(host string, port int) {
	log.Println("starting a synchronous TCP server on", host, port)

	var cons_client int = 0

	lsnr, err := net.Listen("tcp", host+":"+strconv.Itoa(port))

	if err != nil {
		panic(err)
	}

	for {
		c, err := lsnr.Accept()
		if err != nil {
			panic(err)
		}

		cons_client += 1
		log.Println("client connected with address:", c.RemoteAddr(), ", concurrent clients:", cons_client)

		for {
			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				cons_client -= 1
				log.Println("client disconnected with address:", c.RemoteAddr(), ", concurrent clients:", cons_client)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}
			log.Println("command:", cmd)
			if err := respond(c, cmd); err != nil {
				log.Println("err write:", err)
			}
		}
	}
}

func readCommand(c net.Conn) (string, error) {

	buf := make([]byte, 512)
	n, err := c.Read(buf[:])

	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func respond(c net.Conn, cmd string) error {
	log.Println("responding to command:", cmd)
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}
