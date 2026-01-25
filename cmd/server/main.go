package main

import (

	"log/slog"

	"github.com/ao-go-server/internal/server"

)



func main() {

	s := server.NewServer(":7666", "../../resources")

	if err := s.Start(); err != nil {

		slog.Error("Server failed", "error", err)

	}

}
