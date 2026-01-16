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

var podCmd = &cobra.Command{
	Use:   "pod",
	Short: "Manage pods",
	Long:  `Manage pods in the container engine.`,
}

var podListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pods",
	RunE:  runPodList,
}

var podInspectCmd = &cobra.Command{
	Use:   "inspect [name]",
	Short: "Inspect a pod",
	Args:  cobra.ExactArgs(1),
	RunE:  runPodInspect,
}

var podCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new pod",
	RunE:  runPodCreate,
}

var podRmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a pod",
	Args:  cobra.ExactArgs(1),
	RunE:  runPodRm,
}

var (
	podName  string
	podPorts []string
	podForce bool
)

func init() {
	rootCmd.AddCommand(podCmd)
	podCmd.AddCommand(podListCmd)
	podCmd.AddCommand(podInspectCmd)
	podCmd.AddCommand(podCreateCmd)
	podCmd.AddCommand(podRmCmd)

	// Create flags
	podCreateCmd.Flags().StringVarP(&podName, "name", "n", "", "Pod name (required)")
	podCreateCmd.Flags().StringSliceVarP(&podPorts, "port", "p", []string{}, "Port mappings (host:container)")
	_ = podCreateCmd.MarkFlagRequired("name") //nolint:errcheck // flag registration rarely fails

	// Rm flags
	podRmCmd.Flags().BoolVarP(&podForce, "force", "f", false, "Force removal")
}

func runPodList(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	pods, err := client.ListPods(ctx)
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED")
	for _, p := range pods {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID[:12], p.Name, p.Status, p.Created.Format("2006-01-02 15:04:05"))
	}
	w.Flush()
	return nil
}

func runPodInspect(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	pod, err := client.InspectPod(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to inspect pod: %w", err)
	}

	fmt.Printf("ID:      %s\n", pod.ID)
	fmt.Printf("Name:    %s\n", pod.Name)
	fmt.Printf("Status:  %s\n", pod.Status)
	fmt.Printf("Created: %s\n", pod.Created)
	return nil
}

func runPodCreate(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	ports, err := parsePorts(podPorts)
	if err != nil {
		return err
	}

	id, err := client.CreatePod(ctx, podName, ports)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	fmt.Printf("Pod %s created (ID: %s)\n", podName, id)
	return nil
}

func runPodRm(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())
	client, err := container.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to container engine: %w", err)
	}

	if err := client.RemovePod(ctx, args[0], podForce); err != nil {
		return fmt.Errorf("failed to remove pod: %w", err)
	}

	fmt.Printf("Pod %s removed\n", args[0])
	return nil
}
