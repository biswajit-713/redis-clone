package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/diceclone/config"
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

func evalExpire(args []string, c io.ReadWriter, t TimeProvider) error {

	if len(args) != 2 {
		return errors.New("EXPIRE command - invalid arguments")
	}

	_, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errors.New("EXPIRE command - invalid arguments")
	}

	var b []byte

	v := Get(args[0])
	if v == nil {
		b = Encode(0, false)
	} else if v.HasExpired() {
		b = Encode(0, false)
	} else {
		evalSet([]string{args[0], v.Value.(string), "ex", args[1]}, c, t)
		b = Encode(1, false)
	}

	_, err = c.Write(b)
	return err
}

func evalIncrement(args []string, c io.ReadWriter) error {

	// fetch the value from store
	v := Get(args[0])

	var value int64 = -1
	var err error
	if v == nil {
		value = 0
	} else {
		value, err = strconv.ParseInt(v.Value.(string), 10, 64)
		if err != nil {
			return errors.New("ERR value is not an integer or out of range")
		}
	}

	b := Encode(value+1, false)
	_, err = c.Write(b)

	return err
}

func evalBackgroundRewriteAof() error {

	aofFile := config.AppendOnlyFile

	_, err := os.Stat(aofFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if _, err := os.Create(aofFile); err != nil {
				fmt.Println("Unable to create aof: ", err)
			}
		} else {
			fmt.Println("Error retrieving file info: ", err)
		}
	}

	// TODO - refactor this code to move decisioning on AOF to an outside process, preferably in the async_tcp.go
	// TODO - where the BGREWRITEAOF can be called based on the modification time
	// fileInfo, _ := os.Stat(aofFile)
	// modTime := fileInfo.ModTime()
	// if time.Since(modTime) < 5*time.Minute {
	// 	fmt.Printf("Skipping writing to aof. The existing aof is still fresh")
	// 	return nil
	// }

	tempAofFile := fmt.Sprintf("%d-%s", time.Now().Unix(), config.AppendOnlyFile)
	file, err := os.Create(tempAofFile)
	if err != nil {
		fmt.Println("Error creating file: ", err)
		return err
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 4096)

	for pair := range IterateStore() {
		_, err := writer.Write([]byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(pair.Key), pair.Key, len(pair.Value.Value.(string)), pair.Value.Value)))
		if err != nil {
			fmt.Println("Error writing to file: ", err)
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing writer: ", err)
		return err
	}
	err = os.Rename(tempAofFile, aofFile)
	if err != nil {
		fmt.Println("Error renaming file: ", err)
		return err
	}
	return nil
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
	case "EXPIRE":
		return evalExpire(cmd.Args, c, timeProvider)
	case "INCR":
		return evalIncrement(cmd.Args, c)
	case "BGREWRITEAOF":
		return evalBackgroundRewriteAof()
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
