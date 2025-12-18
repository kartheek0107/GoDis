package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Connectng to GoDis on Windows...")
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Sending a manual SET command in RESP format
	fmt.Fprintf(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$8\r\nkartheek\r\n")

	// Read Response
	reader := bufio.NewReader(conn)
	line, _ := reader.ReadString('\n')
	fmt.Printf("📩 Server says: %s", line)

	// Sending a manual GET command
	fmt.Fprintf(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")
	line, _ = reader.ReadString('\n') // Read $8\r\n
	line, _ = reader.ReadString('\n') // Read kartheek\r\n
	fmt.Printf("🔍 Server says: %s", line)
}
