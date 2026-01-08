package cli

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/AkMo3/simplify/internal/reconciler"
	"github.com/AkMo3/simplify/internal/server"
	"github.com/AkMo3/simplify/internal/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Simplify API Server",
	Run: func(cmd *cobra.Command, args []string) {
		err := logger.Init()
		if err != nil {
			log.Fatal(err)
		}

		// Store DB in home directory for now
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatal("failed to get home directory: %w", err)
			return
		}
		dbPath := filepath.Join(home, ".simplify", "simplify.db")

		s, err := store.New(dbPath)
		if err != nil {
			logger.Fatal("Failed to initialize store", err)
		}

		defer s.Close()

		ctx := context.Background()
		podman, err := container.NewClient(ctx)
		if err != nil {
			log.Fatal("Failed to connect to Podman", zap.Error(err))
		}

		worker := reconciler.New(s, podman)
		go worker.Start(ctx)

		srv := server.New(s)

		// TODO: Make port configurable
		if err = srv.Start(":8080"); err != nil {
			logger.Fatal("Server failed to start", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
