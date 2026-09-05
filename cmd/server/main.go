package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kartheek0107/GoDis/internal/persistence" // Import new package
	"github.com/kartheek0107/GoDis/internal/protocol"
	"github.com/kartheek0107/GoDis/internal/server"
	"github.com/kartheek0107/GoDis/internal/store"
)

func main() {
	fmt.Println("🚀 GoDis Engine Starting...")

	// 1. Initialize Memory
	cache := store.Newstore(store.Store{})

	// 2. Initialize AOF
	aof, err := persistence.NewAof("database.aof")
	if err != nil {
		fmt.Println("Error initializing AOF:", err)
		return
	}
	defer aof.Close()

	// 3. RECOVERY: Read the journal to rebuild memory
	// Notice: We reuse the exact same Parser logic!
	file, err := os.Open("database.aof")
	if err == nil {
		fmt.Println("📜 Found AOF file. Replaying commands...")
		parser := protocol.NewParser(file)
		for {
			cmd, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error parsing AOF:", err)
				break
			}
			// Execute the replayed command
			// For simplicity we could just route to the server handler, but doing minimal replay here
			switch strings.ToUpper(cmd[0]) {
			case "SET":
				cache.Set(cmd[1], cmd[2])
			case "DEL":
				cache.Delete(cmd[1])
			case "LPUSH":
				cache.LPush(cmd[1], cmd[2:]...)
			case "RPUSH":
				cache.RPush(cmd[1], cmd[2:]...)
			case "LPOP":
				cache.LPop(cmd[1])
			case "RPOP":
				cache.RPop(cmd[1])
			case "HSET":
				cache.HSet(cmd[1], cmd[2], cmd[3])
			case "SADD":
				cache.SAdd(cmd[1], cmd[2:]...)
			case "PFADD":
				cache.PfAdd(cmd[1], cmd[2:]...)
			case "BF.ADD":
				cache.BfAdd(cmd[1], cmd[2])
			// Simplified heatmap replay for now, assuming args are valid integers from AOF
			}
		}
		file.Close()
		fmt.Println("✅ Recovery Complete.")
	}

	// 4. Start Server
	srv := server.NewServer(":6379", cache, aof)
	srv.Start()
}
