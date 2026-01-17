package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage networks",
	Long:  `Manage networks in the container engine.`,
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	RunE:  runNetworkList,
}

var networkCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new network",
	Args:  cobra.ExactArgs(1),
	RunE:  runNetworkCreate,
}

var networkRmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a network",
	Args:  cobra.ExactArgs(1),
	RunE:  runNetworkRm,
}

func init() {
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkRmCmd)
}

func runNetworkList(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	networks, err := client.ListNetworks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDRIVER\tSUBNET\tCREATED")
	for _, n := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", n.ID[:12], n.Name, n.Driver, n.Subnet, n.Created.Format("2006-01-02 15:04:05"))
	}
	w.Flush()
	return nil
}

func runNetworkCreate(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	id, err := client.CreateNetwork(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	fmt.Printf("Network %s created (ID: %s)\n", args[0], id)
	return nil
}

func runNetworkRm(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	if err := client.RemoveNetwork(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}

	fmt.Printf("Network %s removed\n", args[0])
	return nil
}
