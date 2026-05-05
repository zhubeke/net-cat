package main

import (
	"fmt"
	"os"
	"net-cat/internal/server"
)

func main() {
	port := "8989" // default port
	
	// Check command line arguments
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	
	// Create and start server
	chatServer := server.NewServer()
	err := chatServer.Start(port)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
