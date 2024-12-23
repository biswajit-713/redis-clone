package core

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
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

func evalSet(args []string, c io.ReadWriter, timeProvider TimeProvider) error {

	if len(args) < 2 {
		return errors.New("missing parameters")
	}

	// build a map with the argument list
	params := buildSetParams(args)

	ttl, exists := params["EX"]
	if exists {
		Put(params["key"], NewObj(params["value"], validUntil(ttl, timeProvider)))
	} else {
		Put(params["key"], NewObj(params["value"], -1))
	}

	b := Encode("OK", true)

	_, err := c.Write(b)
	return err
}

func evalGet(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("invalid arguments")
	}
	obj := Get(args[0])
	value := valueOf(obj)

	var b []byte
	if value == nil {
		b = Encode(nil, false)
	} else {
		b = Encode(value, false)
	}
	_, err := c.Write([]byte(b))
	return err
}

func evalTtl(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("invalid arguments")
	}

	obj := Get(args[0])
	ttl := ttlOf(obj)

	b := Encode(ttl, false)
	_, err := c.Write(b)

	return err
}

func evalDel(args []string, c io.ReadWriter) error {

	var deletedKeys = 0
	for _, k := range args {
		if ok := Delete(k); ok {
			deletedKeys++
		}
	}

	b := Encode(deletedKeys, false)
	_, err := c.Write(b)
	return err
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter, timeProvider TimeProvider) error {

	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, c)
	case "SET":
		return evalSet(cmd.Args, c, timeProvider)
	case "GET":
		return evalGet(cmd.Args, c)
	case "TTL":
		return evalTtl(cmd.Args, c)
	case "DEL":
		return evalDel(cmd.Args, c)
	default:
		return evalPing(cmd.Args, c)
	}
}

func validUntil(validFor interface{}, t TimeProvider) int {
	ttl, err := strconv.Atoi(validFor.(string))
	if err != nil {
		return -1
	}
	return int(t.Now().Unix()) + ttl
}

func buildSetParams(args []string) map[string]string {
	params := make(map[string]string)

	params["key"] = strings.ToUpper(args[0])
	params["value"] = args[1]

	for i := 2; i < len(args); i = i + 2 {
		params[strings.ToUpper(args[i])] = args[i+1]
	}
	return params
}

func valueOf(obj *Obj) interface{} {
	if obj == nil {
		return nil
	}

	if obj.ValidTill == -1 {
		return obj.Value
	}

	if obj.ValidTill < int(time.Now().Unix()) {
		return nil
	}
	return obj.Value
}

func ttlOf(obj *Obj) int {
	if obj == nil {
		return -2
	}
	if obj.ValidTill == -1 {
		return -1
	}
	if obj.ValidTill < int(time.Now().Unix()) {
		return -2
	}
	return obj.ValidTill - int(time.Now().Unix())

}
