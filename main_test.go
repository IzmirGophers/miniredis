package main

import (
	"net"
	"testing"
)

/*
func init(t *testing.T) {
	testStore := &store{}
	testStore.l = make(map[string]string)

	testFile, err := os.Create("test.db")

	if err != nil {
		os.Exit(1)
	}
}
*/

func listenTCP(t *testing.T) net.Conn {
	srv, err := net.Listen("tcp", "127.0.0.1:1234")

	if err != nil {
		t.Error(err)
	}

	return srv
}

func TestGet(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1234")

	if err != nil {
		t.Error(err)
	}

	conn.Write([]byte("SET riza 1"))
}
