package core_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/diceclone/core"
)

// MockReadWriter is a mock implementation of io.ReadWriter
type MockReadWriter struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
	LastWrite   []byte
}

func (m *MockReadWriter) Read(b []byte) (n int, e error) {
	return m.ReadBuffer.Read(b)
}

func (m *MockReadWriter) Write(b []byte) (n int, e error) {
	m.LastWrite = append([]byte(nil), b...)
	return m.WriteBuffer.Write(b)
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

	mockReadWriter := &MockReadWriter{
		ReadBuffer:  bytes.NewBufferString(""),
		WriteBuffer: bytes.NewBufferString(""),
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := core.EvalAndRespond(
				&core.RedisCmd{Cmd: tc.command, Args: tc.argument},
				mockReadWriter,
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
		mockReadWriter := &MockReadWriter{
			ReadBuffer:  bytes.NewBufferString(""),
			WriteBuffer: bytes.NewBufferString(""),
		}
		t.Run(tc.name, func(t *testing.T) {
			got := core.EvalAndRespond(&core.RedisCmd{
				Cmd: tc.command, Args: tc.argument,
			}, mockReadWriter)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", reflect.TypeOf(got), reflect.TypeOf(tc.want))
			}

			if !bytes.Equal(mockReadWriter.LastWrite, tc.wantWrite) {
				t.Errorf("got %v, want %v", string(mockReadWriter.LastWrite), string(tc.wantWrite))
			}
		})
	}
}
