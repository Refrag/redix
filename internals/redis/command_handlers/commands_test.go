package commandhandlers_test

import (
	"net"
	"strings"
	"testing"
	"time"
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

// Helper function to create a connection to Redis server
func getRedisConnection(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "localhost:6380")
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}
	return conn
}

// Helper function to send command and read response
func sendCommand(t *testing.T, conn net.Conn, cmd string) string {
	_, err := conn.Write([]byte(cmd + "\r\n"))
	if err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return string(buf[:n])
}

// Helper function to parse array response
func parseArrayResponse(response string) []string {
	lines := strings.Split(strings.TrimSpace(response), "\r\n")
	if !strings.HasPrefix(lines[0], "*") {
		return nil
	}
	
	var items []string
	for i := 1; i < len(lines); i += 2 {
		if strings.HasPrefix(lines[i], "$") && i+1 < len(lines) {
			items = append(items, lines[i+1])
		}
	}
	return items
}

func TestSAddAndSMembers(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Clean up any existing test data
	sendCommand(t, conn, "DEL testset")

	// Test SADD with single member
	response := sendCommand(t, conn, "SADD testset member1")
	if !strings.Contains(response, ":1") {
		t.Fatalf("expected :1, got %s", response)
	}

	// Test SADD with multiple members
	response = sendCommand(t, conn, "SADD testset member2 member3 member4")
	if !strings.Contains(response, ":3") {
		t.Fatalf("expected :3, got %s", response)
	}

	// Test SADD with duplicate member (should return 0)
	response = sendCommand(t, conn, "SADD testset member1")
	if !strings.Contains(response, ":0") {
		t.Fatalf("expected :0 for duplicate member, got %s", response)
	}

	// Test SMEMBERS
	response = sendCommand(t, conn, "SMEMBERS testset")
	members := parseArrayResponse(response)
	if len(members) != 4 {
		t.Fatalf("expected 4 members, got %d: %v", len(members), members)
	}

	expectedMembers := map[string]bool{
		"member1": true,
		"member2": true,
		"member3": true,
		"member4": true,
	}

	for _, member := range members {
		if !expectedMembers[member] {
			t.Fatalf("unexpected member: %s", member)
		}
	}

	// Clean up
	sendCommand(t, conn, "DEL testset")
}

func TestSMembersEmptySet(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Clean up any existing test data
	sendCommand(t, conn, "DEL nonexistentset")

	// Test SMEMBERS on non-existent set (should return empty array)
	response := sendCommand(t, conn, "SMEMBERS nonexistentset")
	if !strings.Contains(response, "*0") {
		t.Fatalf("expected empty array *0, got %s", response)
	}
}

func TestExpireSet(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Clean up any existing test data
	sendCommand(t, conn, "DEL expireset")

	// Create a set
	response := sendCommand(t, conn, "SADD expireset member1 member2")
	if !strings.Contains(response, ":2") {
		t.Fatalf("expected :2, got %s", response)
	}

	// Set expiration to 2 seconds
	response = sendCommand(t, conn, "EXPIRE expireset 2")
	if !strings.Contains(response, ":1") {
		t.Fatalf("expected :1 for successful expire, got %s", response)
	}

	// Check that set still exists
	response = sendCommand(t, conn, "SMEMBERS expireset")
	members := parseArrayResponse(response)
	if len(members) != 2 {
		t.Fatalf("expected 2 members before expiration, got %d", len(members))
	}

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Check that set is now empty/expired
	response = sendCommand(t, conn, "SMEMBERS expireset")
	if !strings.Contains(response, "*0") {
		t.Fatalf("expected empty set after expiration, got %s", response)
	}
}

func TestExpireNonExistentSet(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Clean up any existing test data
	sendCommand(t, conn, "DEL nonexistentset")

	// Try to expire non-existent set (should return 0)
	response := sendCommand(t, conn, "EXPIRE nonexistentset 10")
	if !strings.Contains(response, ":0") {
		t.Fatalf("expected :0 for non-existent set, got %s", response)
	}
}

func TestDelSet(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Clean up any existing test data
	sendCommand(t, conn, "DEL delset")

	// Create a set
	response := sendCommand(t, conn, "SADD delset member1 member2")
	if !strings.Contains(response, ":2") {
		t.Fatalf("expected :2, got %s", response)
	}

	// Verify set exists
	response = sendCommand(t, conn, "SMEMBERS delset")
	members := parseArrayResponse(response)
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	// Delete the set
	response = sendCommand(t, conn, "DEL delset")
	if !strings.Contains(response, ":1") {
		t.Fatalf("expected :1 for successful delete, got %s", response)
	}

	// Verify set is deleted
	response = sendCommand(t, conn, "SMEMBERS delset")
	if !strings.Contains(response, "*0") {
		t.Fatalf("expected empty set after deletion, got %s", response)
	}
}

func TestSAddInvalidArguments(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Test SADD with insufficient arguments
	response := sendCommand(t, conn, "SADD")
	if !strings.Contains(response, "ERR wrong number of arguments") {
		t.Fatalf("expected error for insufficient arguments, got %s", response)
	}

	response = sendCommand(t, conn, "SADD onlykey")
	if !strings.Contains(response, "ERR wrong number of arguments") {
		t.Fatalf("expected error for insufficient arguments, got %s", response)
	}
}

func TestSMembersInvalidArguments(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Test SMEMBERS with insufficient arguments
	response := sendCommand(t, conn, "SMEMBERS")
	if !strings.Contains(response, "ERR wrong number of arguments") {
		t.Fatalf("expected error for insufficient arguments, got %s", response)
	}
}

func TestExpireInvalidArguments(t *testing.T) {
	conn := getRedisConnection(t)
	defer conn.Close()

	// Test EXPIRE with insufficient arguments
	response := sendCommand(t, conn, "EXPIRE")
	if !strings.Contains(response, "ERR wrong number of arguments") {
		t.Fatalf("expected error for insufficient arguments, got %s", response)
	}

	response = sendCommand(t, conn, "EXPIRE onlykey")
	if !strings.Contains(response, "ERR wrong number of arguments") {
		t.Fatalf("expected error for insufficient arguments, got %s", response)
	}

	// Test EXPIRE with invalid seconds
	response = sendCommand(t, conn, "EXPIRE testkey notanumber")
	if !strings.Contains(response, "ERR value is not an integer") {
		t.Fatalf("expected error for invalid seconds, got %s", response)
	}

	// Test EXPIRE with negative seconds
	response = sendCommand(t, conn, "EXPIRE testkey -1")
	if !strings.Contains(response, "ERR invalid expire time") {
		t.Fatalf("expected error for negative seconds, got %s", response)
	}
}
