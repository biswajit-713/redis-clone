package core_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

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
	core.EvalAndRespond(&core.RedisCmd{Cmd: "SET", Args: []string{"expired", "value", "ex", "10"}}, mockReadWriter, timeProvider)
	t.Run("GET value of a key when it is expired", func(t *testing.T) {

		core.EvalAndRespond(&core.RedisCmd{Cmd: "GET", Args: []string{"expired"}}, mockReadWriter, timeProvider)

		want := "$-1\r\n"
		if !bytes.Equal(mockReadWriter.LastWrite, []byte(want)) {
			t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), want)
		}
	})
}
