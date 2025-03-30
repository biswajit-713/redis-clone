package core_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/diceclone/config"
	"github.com/diceclone/core"
)

// MockReadWriter is a mock implementation of io.ReadWriter
type MockReadWriter struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
	LastWrite   []byte
}

type MockTimeProvider struct {
	MockTime time.Time
}

func (m MockTimeProvider) Now() time.Time {
	return time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func (m *MockReadWriter) Read(b []byte) (n int, e error) {
	return m.ReadBuffer.Read(b)
}

func (m *MockReadWriter) Write(b []byte) (n int, e error) {
	m.LastWrite = append([]byte(nil), b...)
	return m.WriteBuffer.Write(b)
}

func setupTest() (*MockReadWriter, MockTimeProvider) {
	mockReadWriter := &MockReadWriter{
		ReadBuffer:  bytes.NewBufferString(""),
		WriteBuffer: bytes.NewBufferString(""),
	}

	mockTime := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	timeProvider := MockTimeProvider{
		MockTime: mockTime,
	}

	return mockReadWriter, timeProvider
}

func TestPINGCommand(t *testing.T) {
	cases := []struct {
		name      string
		command   string
		argument  []string
		want      interface{}
		wantWrite []byte
	}{
		{
			name:      "PING with argument",
			command:   "PING",
			argument:  []string{"hello"},
			want:      nil,
			wantWrite: []byte("$5\r\nhello\r\n"),
		},
		{
			name:      "PING without argument",
			command:   "PING",
			argument:  []string{},
			want:      nil,
			wantWrite: []byte("+PONG\r\n"),
		},
	}

	mockReadWriter, timeProvider := setupTest()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := core.EvalAndRespond(
				&core.RedisCmd{Cmd: tc.command, Args: tc.argument},
				mockReadWriter,
				timeProvider,
			)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
			if !bytes.Equal(mockReadWriter.LastWrite, tc.wantWrite) {
				t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(tc.wantWrite))
			}
		})
	}
}

func TestSETCommand(t *testing.T) {
	cases := []struct {
		name      string
		command   string
		argument  []string
		want      interface{}
		wantWrite []byte
	}{
		{
			name:      "SET with key and value",
			command:   "SET",
			argument:  []string{"key", "value"},
			want:      nil,
			wantWrite: []byte("+OK\r\n"),
		},
		{
			name:      "SET with only key and no value",
			command:   "SET",
			argument:  []string{"key"},
			want:      errors.New("missing parameters"),
			wantWrite: []byte(""),
		},
	}

	for _, tc := range cases {
		mockReadWriter, timeProvider := setupTest()
		t.Run(tc.name, func(t *testing.T) {
			got := core.EvalAndRespond(&core.RedisCmd{
				Cmd: tc.command, Args: tc.argument,
			}, mockReadWriter, timeProvider)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", reflect.TypeOf(got), reflect.TypeOf(tc.want))
			}

			if !bytes.Equal(mockReadWriter.LastWrite, tc.wantWrite) {
				t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(tc.wantWrite))
			}
		})
	}
}

func TestGETCommand(t *testing.T) {
	cases := []struct {
		name     string
		command  string
		argument string
		want     []byte
	}{
		{
			name:     "GET value of a key when it is not expired",
			command:  "GET",
			argument: "key",
			want:     []byte("$5\r\nvalue\r\n"),
		},
		{
			name:     "GET value of a key that does not exist",
			command:  "GET",
			argument: "nonexistent",
			want:     []byte("$-1\r\n"),
		},
	}

	mockReadWriter, timeProvider := setupTest()
	core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"key", "value"}}, mockReadWriter, timeProvider)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := core.EvalAndRespond(&core.RedisCmd{
				Cmd:  tc.command,
				Args: []string{tc.argument},
			}, mockReadWriter, timeProvider)

			if !bytes.Equal(mockReadWriter.LastWrite, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGETCommandWithValueExpired(t *testing.T) {

	mockReadWriter, timeProvider := setupTest()
	t.Run("GET value of a key when it is expired", func(t *testing.T) {

		core.EvalAndRespond(&core.RedisCmd{Cmd: "GET", Args: []string{"expired"}}, mockReadWriter, timeProvider)

		want := "$-1\r\n"
		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})
}

func TestTTLCommand(t *testing.T) {

	t.Run("TTL when key has not expired", func(t *testing.T) {
		mockReadWriter, _ := setupTest()
		timeProvider := core.RealTimeProvider{}

		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"key", "value", "ex", "100"}}, mockReadWriter, timeProvider)
		want := ":100\r\n"
		core.EvalAndRespond(&core.RedisCmd{Cmd: "TTL", Args: []string{"key"}}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})

	t.Run("TTL when key has expired", func(t *testing.T) {
		mockReadWriter, timeProvider := setupTest()
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"key", "value", "ex", "10"}}, mockReadWriter, timeProvider)
		want := ":-2\r\n"

		core.EvalAndRespond(&core.RedisCmd{Cmd: "TTL", Args: []string{"key"}}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})

	t.Run("TTL when key has not expiry set", func(t *testing.T) {
		mockReadWriter, timeProvider := setupTest()
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"key", "value"}}, mockReadWriter, timeProvider)
		want := ":-1\r\n"

		core.EvalAndRespond(&core.RedisCmd{Cmd: "TTL", Args: []string{"key"}}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})

	t.Run("TTL when key does not exist", func(t *testing.T) {
		mockReadWriter, timeProvider := setupTest()

		want := ":-2\r\n"
		core.EvalAndRespond(&core.RedisCmd{Cmd: "TTL", Args: []string{"nonexistentkey"}}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})
}

func TestDELCommand(t *testing.T) {

	mockReadWriter, timeProvider := setupTest()

	t.Run("delete multiple keys", func(t *testing.T) {
		keysToDelete := []string{"k1", "k2", "k3", "k4"}
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"k1", "v1"}}, mockReadWriter, timeProvider)
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"k2", "v2"}}, mockReadWriter, timeProvider)
		want := []byte(":2\r\n")

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "DEL",
			Args: keysToDelete,
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}
	})

	t.Run("delete a non existent key", func(t *testing.T) {
		want := []byte(":0\r\n")
		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "DEL",
			Args: []string{"nonexistentkey"},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}
	})

	t.Run("delete command no arguments passed", func(t *testing.T) {
		want := []byte(":0\r\n")
		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "DEL",
			Args: []string{},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}

	})
}

func TestEXPIRECommand(t *testing.T) {

	mockReadWriter, timeProvider := setupTest()

	t.Run("expire a key with no ttl set", func(t *testing.T) {
		want := []byte(":1\r\n")
		timeProvider := core.RealTimeProvider{}
		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "SET",
			Args: []string{"keyWithNoTtl", "value"},
		}, mockReadWriter, timeProvider)

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{"keyWithNoTtl", "100"},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "GET",
			Args: []string{"keyWithNoTtl"},
		}, mockReadWriter, timeProvider)
		want = []byte("$5\r\nvalue\r\n")
		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("compare values: got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}
	})

	// expire a key with a valid ttl - update the ttl to the new value -> return 1
	t.Run("expire a key with a valid ttl", func(t *testing.T) {
		timeProvider := core.RealTimeProvider{}

		want := []byte(":1\r\n")
		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "SET",
			Args: []string{"keyWithTtl", "value", "ex", "30"},
		}, mockReadWriter, timeProvider)

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{"keyWithTtl", "20"},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "GET",
			Args: []string{"keyWithTtl"},
		}, mockReadWriter, timeProvider)
		want = []byte("$5\r\nvalue\r\n")
		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("compare values: got %v, want %v", string(mockReadWriter.LastWrite), string(want))
		}

	})

	t.Run("set expire for a key that does not exist", func(t *testing.T) {
		want := []byte(":0\r\n")

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{"nonExistentKey", "10"},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", mockReadWriter.LastWrite, want)
		}
	})

	// expire a key who ttl has already expired -> return 0
	t.Run("set expiry for a key that has already expired", func(t *testing.T) {
		want := []byte(":0\r\n")

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "SET",
			Args: []string{"k", "v", "ex", "20"},
		}, mockReadWriter, timeProvider)

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{"k", "30"},
		}, mockReadWriter, timeProvider)

		if !bytes.Equal(mockReadWriter.LastWrite, want) {
			t.Errorf("got %v, want %v", mockReadWriter.LastWrite, want)
		}
	})

	t.Run("expire with missing arguments", func(t *testing.T) {
		want := errors.New("EXPIRE command - invalid arguments")

		got := core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{},
		}, mockReadWriter, timeProvider)

		if got.Error() != want.Error() {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("expire with invalid ttl", func(t *testing.T) {
		want := errors.New("EXPIRE command - invalid arguments")

		got := core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "EXPIRE",
			Args: []string{"key", "invalid ttl"},
		}, mockReadWriter, timeProvider)

		if got.Error() != want.Error() {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestBGREWRITEAOFCommand(t *testing.T) {

	// TODO - the test should verify that the bgrewrite is invoked
	t.Run("rewrite state to AOF in background", func(t *testing.T) {
		mockReadWriter, timeProvider := setupTest()
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"K1", "V1"}}, mockReadWriter, timeProvider)
		core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"K2", "V2"}}, mockReadWriter, timeProvider)

		core.EvalAndRespond(&core.RedisCmd{
			Cmd:  "BGREWRITEAOF",
			Args: []string{},
		}, &MockReadWriter{}, MockTimeProvider{})

		// Verify the AOF file content
		content, _ := os.ReadFile(config.AppendOnlyFile)
		os.Remove(config.AppendOnlyFile)
		expectedContents := []string{"*3\r\n$3\r\nSET\r\n$2\r\nK1\r\n$2\r\nV1\r\n", "*3\r\n$3\r\nSET\r\n$2\r\nK2\r\n$2\r\nV2\r\n"}
		for _, want := range expectedContents {
			if !bytes.Contains(content, []byte(want)) {
				t.Errorf("AOF content does not contain expected entry:\nGot:\n%s\nWant:\n%s", string(content), want)
			}
		}
	})
}
