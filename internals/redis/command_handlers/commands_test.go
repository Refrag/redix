package commandhandlers_test

import (
	"net"
	"testing"
)

func TestPing(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:6380")
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}
	defer conn.Close()

	conn.Write([]byte("PING\r\n"))
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read from redis: %v", err)
	}

	t.Log(string(buf[:n]))
	if string(buf[:n]) != "+PONG\r\n" {
		t.Fatalf("expected PONG, got %s", string(buf[:n]))
	}
}
