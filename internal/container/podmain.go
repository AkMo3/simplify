package container

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	nettypes "go.podman.io/common/libnetwork/types"
)

/* Type definations */

type Client struct {
	ctx context.Context
}

/* Public Static Functions */

func NewClient() (*Client, error) {
	socketPath := getSocketPath()
	ctx, err := bindings.NewConnection(context.Background(), socketPath)

	if err != nil {
		return nil, fmt.Errorf("connecting to podman: %w", err)
	}

	return &Client{ctx: ctx}, nil
}

/* Public Class Functions */

// Context returns the connection context for direct API calls if needed
func (c *Client) Context() context.Context {
	return c.ctx
}

// Run creates and starts a container
func (c *Client) Run(name, image string, ports map[uint16]uint16, env []string) (string, error) {

	// Pull image if not needed
	exists, err := images.Exists(c.ctx, image, nil)
	if err != nil {
		return "", fmt.Errorf("checking image: %w", err)
	}

	if !exists {
		fmt.Printf("Pulling image %s...\n", image)
		_, err = images.Pull(c.ctx, image, nil)
		if err != nil {
			return "", fmt.Errorf("pulling image: %w", err)
		}
	}

	// Create spec
	s := specgen.NewSpecGenerator(image, false)
	s.Name = name
	s.Env = envSliceToMap(env)

	if len(ports) > 0 {
		s.PortMappings = make([]nettypes.PortMapping, 0, len(ports))
		for hostPort, containerPort := range ports {
			fmt.Printf("Adding port mapping (hostPort %d:containerPort %d)\n", hostPort, containerPort)
			s.PortMappings = append(s.PortMappings, nettypes.PortMapping{
				HostPort:      hostPort,
				ContainerPort: containerPort,
				Protocol:      "tcp",
			})
		}
	} else {
		fmt.Printf("No port mappings provided\n")
	}

	// Create container
	createResponse, err := containers.CreateWithSpec(c.ctx, s, nil)
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	// Start container
	if err := containers.Start(c.ctx, createResponse.ID, nil); err != nil {
		return "", fmt.Errorf("starting container: %w", err)
	}

	return createResponse.ID, nil

}

/* Utility Functions */

// Get socket path based on OS
func getSocketPath() string {
	if sock := os.Getenv("PODMAN_SOCK"); sock != "" {
		return "unix://" + sock
	}

	homeDir, _ := os.UserHomeDir()
	macSocket := fmt.Sprintf("%s/.local/share/containers/podman/machine/podman.sock", homeDir)
	if _, err := os.Stat(macSocket); err == nil {
		return "unix://" + macSocket
	}

	return fmt.Sprintf("unix://run/user/%d/podman/podman.sock", os.Getuid())
}

func envSliceToMap(env []string) map[string]string {
	result := make(map[string]string)

	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				result[e[:i]] = e[i+1:]
				break
			}
		}
	}

	return result
}
