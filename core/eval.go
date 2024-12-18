package core

import (
	"errors"
	"io"
	"strings"
)

func evalPing(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) > 1 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func evalSet(args []string, c io.ReadWriter) error {

	if len(args) != 2 {
		return errors.New("missing parameters")
	}

	Put(strings.ToUpper(args[0]), NewObj(args[1]))

	b := Encode("OK", true)

	_, err := c.Write(b)
	return err
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {

	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, c)
	case "SET":
		return evalSet(cmd.Args, c)
	default:
		return evalPing(cmd.Args, c)
	}
}
