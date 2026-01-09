package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers",
	Long:  `List running containers. Use --all to show all containers.`,
	Example: `  simplify ps
  simplify ps --all`,
	RunE: listContainers,
}

var showAll bool

func init() {
	rootCmd.AddCommand(psCmd)

	psCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all containers (default shows just running)")
}

func listContainers(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())

	logger.DebugCtx(ctx, "Listing containers", "all", showAll)

	client, err := container.NewClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to connect to Podman", "error", err)
		return fmt.Errorf("failed to connect to Podman: %w", err)
	}

	containers, err := client.List(ctx, showAll)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to list containers", "error", err)
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("No containers found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS\tPORTS\tCREATED")

	for _, c := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			c.ID,
			c.Name,
			truncateString(c.Image, 30),
			c.Status,
			c.Ports,
			formatCreatedTime(c.Created),
		)
	}

	return w.Flush()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatCreatedTime(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}
