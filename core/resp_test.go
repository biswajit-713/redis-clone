package core_test

import (
	"fmt"
	"testing"

	"github.com/diceclone/core"
)

func TestSimpleString(t *testing.T) {
	cases := map[string]string{
		"+OK\r\n": "OK",
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if v != value {
			t.Errorf("TestSimpleString failed: expected %s, got %v", v, value)
		}
	}
}

func TestError(t *testing.T) {
	cases := map[string]string{
		"-Error delivered\r\n": "Error delivered",
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if v != value {
			t.Errorf("TestError failed: expected %s, got %v", v, value)
		}
	}

}

func TestInt64(t *testing.T) {
	cases := map[string]int64{
		":0\r\n":    0,
		":1000\r\n": 1000,
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if value != v {
			t.Errorf("TestInt64 failed: expected %d, got %v", v, value)
		}
	}
}

func TestBulkString(t *testing.T) {
	cases := map[string]string{
		"$5\r\nhello\r\n": "hello",
		"$0\r\n\r\n":      "",
	}

	for input, want := range cases {
		got, _ := core.Decode([]byte(input))
		if got != want {
			t.Errorf("TestBulkString failed. got %s, want: %s", got, want)
		}
	}
}

func TestArray(t *testing.T) {

	cases := map[string][]interface{}{
		"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n":        {"hello", "world"},
		"*3\r\n:1\r\n:2\r\n:3\r\n":                    {int64(1), int64(2), int64(3)},
		"*5\r\n:1\r\n:2\r\n:3\r\n+OK\r\n$2\r\nOK\r\n": {int64(1), int64(2), int64(3), "OK", "OK"},
	}

	for command, want := range cases {
		value, _ := core.Decode([]byte(command))
		array, ok := value.([]interface{})
		if !ok {
			t.Errorf("expected []interface{}, got %T", value)
		}
		if len(array) != len(want) {
			t.Errorf("The result count is incorrect. got %d, want %d", len(array), len(want))
		}
		for i := range array {
			if fmt.Sprintf("%v", want[i]) != fmt.Sprintf("%v", array[i]) {
				t.Errorf("value didn't match, got %v, want %v", array[i], want[i])
			}
		}
	}

}
