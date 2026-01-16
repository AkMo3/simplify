package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AkMo3/simplify/internal/caddy"
	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/AkMo3/simplify/internal/permissions"
	"github.com/AkMo3/simplify/internal/reconciler"
	"github.com/AkMo3/simplify/internal/server"
	"github.com/AkMo3/simplify/internal/store"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Simplify API server",
	Long: `Start the Simplify API server which provides HTTP endpoints
for managing applications, teams, projects, and environments.

The server also runs the reconciliation loop which ensures containers
match the desired state in the database.`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	// Get config (already loaded in main.go)
	cfg := config.Get()

	logger.Info("Starting Simplify server",
		"env", cfg.Env,
		"port", cfg.Server.Port,
		"database", cfg.Database.Path,
	)

	// Ensure database directory exists and is writable
	if err := permissions.EnsureFileWritable(cfg.Database.Path); err != nil {
		logger.Error("Database path not writable", "path", cfg.Database.Path, "error", err)
		return err
	}

	// Initialize store
	s, err := store.New(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to initialize store", "error", err)
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			logger.Error("Failed to close store", "error", err)
		}
	}()

	// Initialize Podman client
	ctx := context.Background()
	podman, err := container.NewClient(ctx)
	if err != nil {
		logger.Error("Failed to connect to Podman", "error", err)
		return err
	}
	logger.Info("Connected to Podman")

	// Create context that cancels on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// Initialize Caddy manager if enabled
	var caddyMgr *caddy.Manager
	if cfg.Caddy.Enabled {
		logger.Info("Caddy integration enabled, initializing...")

		// Ensure proxy network exists
		if err := ensureProxyNetwork(ctx, podman, cfg.Caddy.ProxyNetwork); err != nil {
			logger.Error("Failed to create proxy network", "error", err)
			return err
		}

		// Create and start Caddy manager
		caddyMgr = caddy.New(podman, &cfg.Caddy, cfg.Server.Port)
		if err := caddyMgr.EnsureRunning(ctx); err != nil {
			logger.Error("Failed to start Caddy", "error", err)
			return err
		}
		logger.Info("Caddy container started")

		// Defer Caddy cleanup
		defer func() {
			if err := caddyMgr.Stop(context.Background()); err != nil {
				logger.Error("Failed to stop Caddy", "error", err)
			}
		}()
	}

	// Start reconciler in background
	worker := reconciler.New(s, podman)
	go worker.Start(ctx)
	logger.Info("Reconciler started")

	// Create and start HTTP server
	srv := server.New(cfg, s, podman)

	logger.Info("HTTP server starting",
		"addr", cfg.Server.Port,
		"healthz", "/healthz",
		"readyz", "/readyz",
		"api", "/api/v1",
	)

	// Start server (blocks until context is canceled)
	if err := srv.Start(ctx); err != nil {
		logger.Error("Server error", "error", err)
		return err
	}

	logger.Info("Simplify server stopped")
	return nil
}

// ensureProxyNetwork creates the proxy network if it doesn't exist
func ensureProxyNetwork(ctx context.Context, client container.ContainerManager, networkName string) error {
	networks, err := client.ListNetworks(ctx)
	if err != nil {
		return err
	}

	for _, n := range networks {
		if n.Name == networkName {
			logger.DebugCtx(ctx, "Proxy network already exists", "network", networkName)
			return nil
		}
	}

	logger.Info("Creating proxy network", "network", networkName)
	// Use MTU 1350 to be safe for cloud networks (like user's ens3)
	_, err = client.CreateNetwork(ctx, networkName, container.NetworkOptions{MTU: 1350})
	return err
}
