package main

import (
	"fmt"
	"path/filepath"

	"github.com/ao-go-server/internal/server"
	"github.com/spf13/cobra"
)

var env string
var port string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting server in %s environment on port %s...\n", env, port)

		// En un entorno real, usaríamos 'env' para elegir el path de recursos
		resourcesPath := "resources"
		if env == "dev" {
			// Si se ejecuta desde cmd/launcher durante el desarrollo
			resourcesPath = "resources"
		}

		absPath, _ := filepath.Abs(resourcesPath)
		fmt.Printf("Using resources from: %s\n", absPath)

		s := server.NewServer(":"+port, resourcesPath)
		if err := s.Start(); err != nil {
			fmt.Printf("Server failed: %v\n", err)
		}
	},
}

func init() {
	startCmd.Flags().StringVarP(&env, "env", "e", "dev", "Environment (dev, prod)")
	startCmd.Flags().StringVarP(&port, "port", "p", "7666", "Port to listen on")
	rootCmd.AddCommand(startCmd)
}
