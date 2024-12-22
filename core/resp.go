package core

import (
	"errors"
	"fmt"
	"strconv"
)

func readLength(data []byte) (int, int) {

	var c []byte
	var pos int = 0
	for _, value := range data {
		if value == '\r' {
			break
		}
		c = append(c, value)
		pos++
	}

	len, _ := strconv.ParseInt(string(c), 10, 64)

	return int(len), int(pos + 2)
}

func readSimpleString(data []byte) (string, int, error) {
	pos := 0
	for ; data[pos] != '\r'; pos++ {

	}
	return string(data[:pos]), pos + 3, nil
}

func readInt64(data []byte) (int64, int, error) {
	var result []byte
	pos := 0
	for ; data[pos] != '\r'; pos++ {
		result = append(result, data[pos])
	}
	parsedValue, _ := strconv.ParseInt(string(result), 10, 64)
	return parsedValue, pos + 3, nil
}

func readBulkstring(data []byte) (string, int, error) {

	length, delta := readLength(data)

	var bulkstringStartPosition = 3
	var result []byte
	for i := bulkstringStartPosition; data[i] != '\r' && i < bulkstringStartPosition+int(length); i++ {
		result = append(result, data[i])
	}

	return string(result), length + delta + 3, nil
}

func readArray(data []byte) (interface{}, int, error) {

	count, nextPos := readLength(data)

	var result []interface{}

	for i := 0; i < count; i++ {
		response, delta, err := DecodeOne(data[nextPos:])
		if err != nil {
			return nil, 0, err
		}
		result = append(result, response)
		nextPos += delta
	}

	return result, 0, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {

	identifier := data[0]

	switch identifier {
	case '+':
		return readSimpleString(data[1:])
	case '-':
		return readSimpleString(data[1:])
	case ':':
		return readInt64(data[1:])
	case '$':
		return readBulkstring(data[1:])
	case '*':
		return readArray(data[1:])
	default:
		return nil, 0, errors.New("invalid command")
	}

}

func Decode(data []byte) (interface{}, error) {

	if len(data) == 0 {
		return nil, errors.New("invalid command")
	}

	result, _, err := DecodeOne(data)
	return result, err
}

func Encode(value interface{}, isSimple bool) []byte {
	if value == nil {
		return []byte("$-1\r\n")
	}

	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))

	case int, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	default:
		return []byte("$-1\r\n")
	}
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}
	ts := value.([]interface{})
	tokens := make([]string, len(ts))
	for i := range ts {
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}
