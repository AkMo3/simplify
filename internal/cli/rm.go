package cli

import (
	"context"
	"fmt"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [containers...]",
	Short: "Remove one or more containers",
	Long:  `Remove one or more containers. Use --force to remove running containers.`,
	Example: `  simplify rm web
  simplify rm web api worker
  simplify rm --force web`,
	Args: cobra.MinimumNArgs(1),
	RunE: removeContainers,
}

var forceRemove bool

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force remove running containers")
}

func removeContainers(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())

	logger.InfoCtx(ctx, "Removing containers", "names", args, "force", forceRemove)

	client, err := container.NewClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to connect to Podman", "error", err)
		return fmt.Errorf("failed to connect to Podman: %w", err)
	}

	var errors []error
	for _, name := range args {
		if err := client.Remove(ctx, name, forceRemove); err != nil {
			logger.ErrorCtx(ctx, "Failed to remove container", "name", name, "error", err)
			errors = append(errors, fmt.Errorf("%s: %w", name, err))
		} else {
			fmt.Printf("Container %s removed\n", name)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove %d container(s)", len(errors))
	}

	return nil
}
