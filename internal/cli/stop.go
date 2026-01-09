package cli

import (
	"context"
	"fmt"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [container]",
	Short: "Stop a running container",
	Long:  `Stop a running container gracefully with optional timeout.`,
	Example: `  simplify stop web
  simplify stop web --timeout 30`,
	Args: cobra.ExactArgs(1),
	RunE: stopContainer,
}

var stopTimeout uint

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().UintVarP(&stopTimeout, "timeout", "t", 10, "Seconds to wait before force killing")
}

func stopContainer(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	name := args[0]

	logger.InfoCtx(ctx, "Stopping container", "name", name, "timeout", stopTimeout)

	client, err := container.NewClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to connect to Podman", "error", err)
		return fmt.Errorf("failed to connect to Podman: %w", err)
	}

	timeout := stopTimeout
	if err := client.Stop(ctx, name, &timeout); err != nil {
		logger.ErrorCtx(ctx, "Failed to stop container", "name", name, "error", err)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("Container %s stopped\n", name)
	return nil
}
