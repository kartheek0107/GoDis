package tests

import (
	"fmt"
	"net"
	"testing"
	"time"
	"strings"

	"github.com/kartheek0107/GoDis/internal/persistence"
	"github.com/kartheek0107/GoDis/internal/server"
	"github.com/kartheek0107/GoDis/internal/store"
)

func sendCommandTest(t *testing.T, conn net.Conn, cmd []string, expected string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))
	for _, arg := range cmd {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	conn.Write([]byte(sb.String()))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	resp := string(buf[:n])

	// Since sometimes multiple responses can clump up in real networking (or if we read too early)
	// let's do a basic prefix/contains check for simpler testing
	if !strings.HasPrefix(resp, expected) {
		t.Errorf("Expected %q to start with %q", resp, expected)
	}
}

func TestServerIntegration(t *testing.T) {
	// 1. Initialize Memory
	cache := store.Newstore(store.Store{})

	// 2. Initialize AOF
	aof, err := persistence.NewAof("test_database.aof")
	if err != nil {
		t.Fatalf("Error initializing AOF: %v", err)
	}
	defer aof.Close()

	// 4. Start Server on different port
	srv := server.NewServer(":6380", cache, aof)
	go srv.Start()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:6380")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	sendCommandTest(t, conn, []string{"PING"}, "+PONG\r\n")
	sendCommandTest(t, conn, []string{"SET", "k1", "v1"}, "+OK\r\n")
	sendCommandTest(t, conn, []string{"GET", "k1"}, "$2\r\nv1\r\n")
	sendCommandTest(t, conn, []string{"DEL", "k1"}, ":1\r\n")
	sendCommandTest(t, conn, []string{"GET", "k1"}, "$-1\r\n")

	sendCommandTest(t, conn, []string{"LPUSH", "l1", "a", "b"}, ":2\r\n")
	// LPOP should return 'b' since it was pushed last and thus at index 0
	sendCommandTest(t, conn, []string{"LPOP", "l1"}, "$1\r\nb\r\n")

	sendCommandTest(t, conn, []string{"HSET", "h1", "f1", "v1"}, ":1\r\n")
	sendCommandTest(t, conn, []string{"HGET", "h1", "f1"}, "$2\r\nv1\r\n")

	sendCommandTest(t, conn, []string{"SADD", "s1", "m1"}, ":1\r\n")
	sendCommandTest(t, conn, []string{"SISMEMBER", "s1", "m1"}, ":1\r\n")

	sendCommandTest(t, conn, []string{"PFADD", "pf1", "a", "b"}, ":1\r\n")

	sendCommandTest(t, conn, []string{"BF.ADD", "bf1", "a"}, ":1\r\n")
	sendCommandTest(t, conn, []string{"BF.EXISTS", "bf1", "a"}, ":1\r\n")

	sendCommandTest(t, conn, []string{"HEATMAP.SET", "hm1", "1", "2", "3"}, "+OK\r\n")
	sendCommandTest(t, conn, []string{"HEATMAP.GET", "hm1", "1", "2"}, ":3\r\n")
}
