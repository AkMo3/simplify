package container

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AkMo3/simplify/internal/logger"
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

// ContainerInfo holds container information for listing
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created time.Time
}

/* Public Static Functions */

func NewClient(ctx context.Context) (*Client, error) {
	logger.DebugCtx(ctx, "Connecting to Podman socket")

	socketPath := getSocketPath()
	ctx, err := bindings.NewConnection(ctx, socketPath)

	if err != nil {
		return nil, fmt.Errorf("connecting to podman: %w", err)
	}

	logger.DebugCtx(ctx, "Connected to Podman", "socket", socketPath)
	return &Client{ctx: ctx}, nil
}

/* Public Class Functions */

// Context returns the connection context for direct API calls if needed
func (c *Client) Context() context.Context {
	return c.ctx
}

// Run creates and starts a container
func (c *Client) Run(ctx context.Context, name, image string, ports map[uint16]uint16, env []string) (string, error) {
	logger.DebugCtx(ctx, "Checking if image exists", "image", image)

	// Pull image if not needed
	exists, err := images.Exists(c.ctx, image, nil)
	if err != nil {
		return "", fmt.Errorf("checking image: %w", err)
	}

	if !exists {
		logger.InfoCtx(ctx, "Pulling image", "image", image)
		_, err = images.Pull(c.ctx, image, nil)
		if err != nil {
			return "", fmt.Errorf("pulling image: %w", err)
		}
		logger.DebugCtx(ctx, "Image pulled successfully", "image", image)
	}

	// Create spec
	s := specgen.NewSpecGenerator(image, false)
	s.Name = name
	s.Env = envSliceToMap(env)

	if len(ports) > 0 {
		s.PortMappings = make([]nettypes.PortMapping, 0, len(ports))
		for hostPort, containerPort := range ports {
			logger.DebugCtx(ctx, "Adding port mapping",
				"host_port", hostPort,
				"container_port", containerPort,
			)
			s.PortMappings = append(s.PortMappings, nettypes.PortMapping{
				HostIP:        "127.0.0.1",
				HostPort:      hostPort,
				ContainerPort: containerPort,
				Protocol:      "tcp",
			})
		}
	} else {
		logger.DebugCtx(ctx, "No port mappings provided")
	}

	// Create container
	logger.DebugCtx(ctx, "Creating container", "name", name)
	createResponse, err := containers.CreateWithSpec(c.ctx, s, nil)
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	// Start container
	logger.DebugCtx(ctx, "Starting container", "id", createResponse.ID[:12])
	if err := containers.Start(c.ctx, createResponse.ID, nil); err != nil {
		return "", fmt.Errorf("starting container: %w", err)
	}

	logger.InfoCtx(ctx, "Container running",
		"name", name,
		"id", createResponse.ID[:12],
	)

	return createResponse.ID, nil
}

// Stop stops a running container
func (c *Client) Stop(ctx context.Context, name string, timeout *uint) error {
	logger.DebugCtx(ctx, "Stopping container", "name", name)

	if err := containers.Stop(c.ctx, name, &containers.StopOptions{Timeout: timeout}); err != nil {
		return fmt.Errorf("stopping container: %w", err)
	}

	logger.InfoCtx(ctx, "Container stopped", "name", name)
	return nil
}

// Remove removes a container
func (c *Client) Remove(ctx context.Context, name string, force bool) error {
	logger.DebugCtx(ctx, "Removing container", "name", name, "force", force)

	_, err := containers.Remove(c.ctx, name, &containers.RemoveOptions{Force: &force})
	if err != nil {
		return fmt.Errorf("removing container: %w", err)
	}

	logger.InfoCtx(ctx, "Container removed", "name", name)
	return nil
}

// List returns containers based on filters
func (c *Client) List(ctx context.Context, all bool) ([]ContainerInfo, error) {
	logger.DebugCtx(ctx, "Listing containers", "all", all)

	listContainers, err := containers.List(c.ctx, &containers.ListOptions{All: &all})
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	result := make([]ContainerInfo, 0, len(listContainers))
	for _, ctr := range listContainers {
		name := ""
		if len(ctr.Names) > 0 {
			name = ctr.Names[0]
		}

		ports := formatPorts(ctr.Ports)

		result = append(result, ContainerInfo{
			ID:      ctr.ID[:12],
			Name:    name,
			Image:   ctr.Image,
			Status:  ctr.State,
			Ports:   ports,
			Created: ctr.Created,
		})
	}

	logger.DebugCtx(ctx, "Found containers", "count", len(result))
	return result, nil
}

// Logs streams container Logs
func (c *Client) Logs(ctx context.Context, name string, follow bool, tail string) error {
	logger.DebugCtx(ctx, "Getting container logs",
		"name", name,
		"follow", follow,
		"tail", tail,
	)

	stdoutCh := make(chan string)
	stderrCh := make(chan string)

	opts := &containers.LogOptions{
		Follow:     &follow,
		Stdout:     ptrBool(true),
		Stderr:     ptrBool(true),
		Timestamps: ptrBool(true),
	}

	if tail != "" {
		opts.Tail = &tail
	}

	go func() {
		if err := containers.Logs(c.ctx, name, opts, stdoutCh, stderrCh); err != nil {
			logger.ErrorCtx(ctx, "error streaming logs", "error", err)
		}

		close(stdoutCh)
		close(stderrCh)
	}()

	// Print logs from both channels
	for {
		select {
		case line, ok := <-stdoutCh:
			if !ok {
				stdoutCh = nil
			} else {
				fmt.Println(line)
			}

		case line, ok := <-stderrCh:
			if !ok {
				stderrCh = nil
			} else {
				fmt.Println(line)
			}
		case <-ctx.Done():
			return ctx.Err()
		}

		if stdoutCh == nil && stderrCh == nil {
			break
		}
	}

	return nil
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

// Helper function to format ports for display
func formatPorts(ports []nettypes.PortMapping) string {
	if len(ports) == 0 {
		return ""
	}

	var portStrs []string
	for _, p := range ports {
		if p.HostPort > 0 {
			portStrs = append(portStrs, fmt.Sprintf("%s:%d->%d/%s",
				p.HostIP, p.HostPort, p.ContainerPort, p.Protocol))
		} else {
			portStrs = append(portStrs, fmt.Sprintf("%d/%s",
				p.ContainerPort, p.Protocol))
		}
	}
	return strings.Join(portStrs, ", ")
}

// Helper to create bool pointer
func ptrBool(b bool) *bool {
	return &b
}
