package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [container]",
	Short: "Fetch container logs",
	Long:  `Fetch logs from a container. Use --follow to stream logs in real-time.`,
	Example: `  simplify logs web
  simplify logs web --follow
  simplify logs web --tail 100`,
	Args: cobra.ExactArgs(1),
	RunE: getContainerLogs,
}

var (
	followLogs bool
	tailLines  string
)

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVarP(&tailLines, "tail", "n", "", "Number of lines to show from the end")
}

func getContainerLogs(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	name := args[0]

	logger.DebugCtx(ctx, "Getting container logs",
		"name", name,
		"follow", followLogs,
		"tail", tailLines,
	)

	client, err := container.NewClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to connect to Podman", "error", err)
		return fmt.Errorf("failed to connect to Podman: %w", err)
	}

	// Handle Ctrl+C gracefully
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := client.Logs(ctx, name, followLogs, tailLines); err != nil {
		if ctx.Err() != nil {
			// User canceled with Ctrl+C
			return nil
		}
		logger.ErrorCtx(ctx, "Failed to get logs", "name", name, "error", err)
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
