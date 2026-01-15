package cli

import (
	"context"
	"fmt"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a container",
	Long:  `Run a container with the specified image and configuration.`,
	Example: `  simplify run --name web --image nginx:latest --port 8080:80
  simplify run --name api --image myapp:v1 --port 3000:3000 --env DB_HOST=localhost`,
	RunE: runContainer,
}

var (
	containerName string
	imageName     string
	portMappings  []string
	envVars       []string
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&containerName, "name", "n", "", "Container name (required)")
	runCmd.Flags().StringVarP(&imageName, "image", "i", "", "Container image (required)")
	runCmd.Flags().StringSliceVarP(&portMappings, "port", "p", []string{}, "Port mappings (host:container)")
	runCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Environment variables (KEY=VALUE)")

	_ = runCmd.MarkFlagRequired("name")  //nolint:errcheck // flag registration rarely fails
	_ = runCmd.MarkFlagRequired("image") //nolint:errcheck // flag registration rarely fails
}

func runContainer(cmd *cobra.Command, args []string) error {
	ctx := logger.WithOperationID(context.Background())

	logger.InfoCtx(ctx, "Running container",
		"name", containerName,
		"image", imageName,
	)

	client, err := container.NewClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to connect to Podman", "error", err)
		return fmt.Errorf("failed to connect to Podman: %w", err)
	}

	ports, err := parsePorts(portMappings)
	if err != nil {
		logger.ErrorCtx(ctx, "Invalid port mapping", "error", err)
		return err
	}

	logger.DebugCtx(ctx, "Parsed configuration",
		"ports", ports,
		"env_count", len(envVars),
	)

	id, err := client.Run(ctx, containerName, imageName, ports, envVars, nil, "")
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to run container", "error", err)
		return fmt.Errorf("failed to run container: %w", err)
	}

	fmt.Printf("Container %s started (ID: %s)\n", containerName, id[:12])
	return nil
}

// parsePorts converts "host:container" strings to a map
func parsePorts(mappings []string) (map[uint16]uint16, error) {
	ports := make(map[uint16]uint16)

	for _, m := range mappings {
		var host, containerPort uint16
		_, err := fmt.Sscanf(m, "%d:%d", &host, &containerPort)
		if err != nil {
			return nil, fmt.Errorf("invalid port mapping %q (use host:container format): %w", m, err)
		}
		ports[host] = containerPort
	}

	return ports, nil
}
