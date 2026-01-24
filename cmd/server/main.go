package main

import (
	"fmt"
	"github.com/ao-go-server/internal/server"
)

func main() {
	s := server.NewServer(":7666", "../../resources")
	if err := s.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}