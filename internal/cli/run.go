package cli

import (
	"fmt"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a container",
	Long:  `Run a container with the specified image and configuration`,
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
	runCmd.Flags().StringSliceVarP(&portMappings, "port", "p", []string{}, "Port Mappings (host:container)")
	runCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Environment variables (KEY=VALUE)")

	runCmd.MarkFlagRequired("name")
	runCmd.MarkFlagRequired("image")
}

func runContainer(cmd *cobra.Command, args []string) error {
	client, err := container.NewClient()

	if err != nil {
		return fmt.Errorf("faild to connect to Podman: %w", err)
	}

	ports, err := parsePorts(portMappings)
	if err != nil {
		return err
	}

	id, err := client.Run(containerName, imageName, ports, envVars)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	fmt.Printf("Container %s started: %s\n", containerName, id[:12])

	return nil
}

/* Private Functions */
func parsePorts(mappings []string) (map[uint16]uint16, error) {
	ports := make(map[uint16]uint16)

	for _, m := range mappings {
		var host, container uint16
		_, err := fmt.Sscanf(m, "%d:%d", &host, &container)

		if err != nil {
			return nil, fmt.Errorf("invalid port mapping %q (use host:container)", m)
		}

		ports[host] = container
	}

	return ports, nil
}
