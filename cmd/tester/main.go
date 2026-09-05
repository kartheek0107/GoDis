package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func sendCommand(conn net.Conn, cmd []string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))
	for _, arg := range cmd {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	conn.Write([]byte(sb.String()))
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	commands := [][]string{
		{"PING"},
		{"SET", "key1", "val1"},
		{"GET", "key1"},
		{"DEL", "key1"},
		{"GET", "key1"},
		{"LPUSH", "mylist", "a", "b", "c"},
		{"LRANGE", "mylist", "0", "-1"},
		{"LPOP", "mylist"},
		{"LRANGE", "mylist", "0", "-1"},
		{"HSET", "myhash", "f1", "v1"},
		{"HGET", "myhash", "f1"},
		{"HGETALL", "myhash"},
		{"SADD", "myset", "m1", "m2"},
		{"SMEMBERS", "myset"},
		{"SISMEMBER", "myset", "m1"},
		{"PFADD", "myhll", "a", "b", "c", "c"},
		{"PFCOUNT", "myhll"},
		{"BF.ADD", "mybf", "apple"},
		{"BF.EXISTS", "mybf", "apple"},
		{"BF.EXISTS", "mybf", "banana"},
		{"HEATMAP.SET", "myhm", "10", "20", "100"},
		{"HEATMAP.GET", "myhm", "10", "20"},
	}

	for _, cmd := range commands {
		sendCommand(conn, cmd)
		// Give small amount of time or just read line
		reader := bufio.NewReader(conn)
		// For simplicity we just read whatever is there, it's not a full RESP parser in the test
		buf := make([]byte, 1024)
		n, _ := reader.Read(buf)
		fmt.Printf("Cmd: %v\nResp: %q\n\n", cmd, string(buf[:n]))
	}
}
