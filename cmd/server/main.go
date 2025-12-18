package main

import (
	"fmt"

	"github.com/kartheek0107/GoDis/internal/server"
	"github.com/kartheek0107/GoDis/internal/store"
)

func main() {
	cache := store.Newstore(store.Store{})
	server := server.NewServer(":6379", cache)

	fmt.Println("🚀 GoDis Engine Starting...")
	if err := server.Start(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
