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

func evalPing(args []string, c io.ReadWriter) []byte {
	var b []byte

	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalSet(args []string, c io.ReadWriter, timeProvider TimeProvider) []byte {

	if len(args) < 2 {
		return Encode(errors.New("missing parameters"), false)
	}

	// build a map with the argument list
	params := buildSetParams(args)
	oType, oEncoding := deduceTypeEncoding(params["value"])

	ttl, exists := params["EX"]
	if exists {
		Put(params["key"], NewObj(params["value"], calculateDuration(ttl, timeProvider), oType, oEncoding))
	} else {
		Put(params["key"], NewObj(params["value"], -1, oType, oEncoding))
	}

	return Encode("OK", true)
}

func evalGet(args []string, c io.ReadWriter) []byte {
	if len(args) != 1 {
		return Encode(errors.New("invalid arguments"), false)
	}
	obj := Get(args[0])
	value := valueOf(obj)

	var b []byte
	if value == nil {
		b = Encode(nil, false)
	} else {
		b = Encode(value, false)
	}
	return b
}

func evalTtl(args []string, c io.ReadWriter) []byte {
	if len(args) != 1 {
		return Encode(errors.New("invalid arguments"), false)
	}

	obj := Get(args[0])
	ttl := ttlOf(obj)

	return Encode(ttl, false)

}

func evalDel(args []string, c io.ReadWriter) []byte {

	var deletedKeys = 0
	for _, k := range args {
		if ok := Delete(k); ok {
			deletedKeys++
		}
	}

	return Encode(deletedKeys, false)
}

func evalExpire(args []string, c io.ReadWriter, t TimeProvider) []byte {

	if len(args) != 2 {
		return Encode(errors.New("EXPIRE command - invalid arguments"), false)
	}

	_, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("EXPIRE command - invalid arguments"), false)
	}

	v := Get(args[0])
	if v == nil {
		return Encode(0, false)
	} else if v.HasExpired() {
		return Encode(0, false)
	} else {
		evalSet([]string{args[0], v.Value.(string), "ex", args[1]}, c, t)
		return Encode(1, false)
	}

}

func evalIncrement(args []string) []byte {

	// fetch the value from store
	v := Get(args[0])
	if v == nil {
		Put(args[0], NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT))
	}

	v = Get(args[0])

	if !assertType(v.TypeEncoding, OBJ_TYPE_STRING) {
		return Encode(errors.New("operation not permitted on this type"), false)
	}
	// if the encoding is not integer, throw error
	if !assertEncoding(v.TypeEncoding, OBJ_ENCODING_INT) {
		return Encode(errors.New("operation not permitted on this encoding"), false)
	}
	result, _ := strconv.ParseInt(v.Value.(string), 10, 64)
	// convert the value to integer, increment the value and return it
	v.Value = strconv.FormatInt(result+1, 10)

	return Encode(result+1, false)
}

func evalBackgroundRewriteAof() []byte {

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
		return Encode(err, false)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 4096)

	for pair := range IterateStore() {
		_, err := writer.Write([]byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(pair.Key), pair.Key, len(pair.Value.Value.(string)), pair.Value.Value)))
		if err != nil {
			fmt.Println("Error writing to file: ", err)
			return Encode(err, false)
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing writer: ", err)
		return Encode(err, false)
	}
	err = os.Rename(tempAofFile, aofFile)
	if err != nil {
		fmt.Println("Error renaming file: ", err)
		return Encode(err, false)
	}
	return Encode("OK", true)
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter, timeProvider TimeProvider) error {
	var buf []byte
	switch cmd.Cmd {
	case "PING":
		buf = evalPing(cmd.Args, c)
	case "SET":
		buf = evalSet(cmd.Args, c, timeProvider)
	case "GET":
		buf = evalGet(cmd.Args, c)
	case "TTL":
		buf = evalTtl(cmd.Args, c)
	case "DEL":
		buf = evalDel(cmd.Args, c)
	case "EXPIRE":
		buf = evalExpire(cmd.Args, c, timeProvider)
	case "INCR":
		buf = evalIncrement(cmd.Args)
	case "BGREWRITEAOF":
		buf = evalBackgroundRewriteAof()
	default:
		buf = evalPing(cmd.Args, c)
	}

	_, err := c.Write(buf)
	return err
}

func calculateDuration(validFor interface{}, t TimeProvider) int {
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

func deduceTypeEncoding(v string) (uint8, uint8) {
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return OBJ_TYPE_STRING, OBJ_ENCODING_INT
	}

	if len(v) <= 44 {
		return OBJ_TYPE_STRING, OBJ_ENCODING_EMBSTR
	}
	return OBJ_TYPE_STRING, OBJ_ENCODING_RAW
}
