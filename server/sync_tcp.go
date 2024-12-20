package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/diceclone/core"
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
			log.Println("error accepting connection", err)
			return
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
			respond(c, cmd)
		}
	}
}

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {

	buf := make([]byte, 512)
	n, err := c.Read(buf[:])

	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil

}

func respond(c io.ReadWriter, cmd *core.RedisCmd) {
	realTimeProvider := core.NewRealTimeProvider()
	err := core.EvalAndRespond(cmd, c, realTimeProvider)
	if err != nil {
		respondError(err, c)
	}
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}
