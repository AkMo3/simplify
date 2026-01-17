// Package container provides container management operations using Podman.
package container

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AkMo3/simplify/internal/logger"
	"github.com/containers/podman/v5/libpod/define"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/bindings/network"
	"github.com/containers/podman/v5/pkg/bindings/pods"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/containers/podman/v5/pkg/specgen"
	nettypes "go.podman.io/common/libnetwork/types"
)

// Client wraps the Podman bindings
type Client struct {
	ctx context.Context
}

// ContainerInfo holds container information for listing
type ContainerInfo struct {
	Created      time.Time
	Ports        map[string]string
	Labels       map[string]string
	ID           string
	Name         string
	Image        string
	Status       string
	IPAddress    string
	ExposedPorts []string
	PodID        string
	Networks     []string
}

// NewClient creates a new Podman client
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

// Context returns the connection context for direct API calls if needed
func (c *Client) Context() context.Context {
	return c.ctx
}

// Run creates and starts a container
func (c *Client) Run(ctx context.Context, name, image string, ports map[uint16]uint16, env []string, labels map[string]string, podName, networkName string) (string, error) {
	logger.DebugCtx(ctx, "Checking if image exists", "image", image)

	// Pull image if not exists
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
	s.Labels = labels

	switch {
	case podName != "":
		s.Pod = podName
		// When in a pod, ports are ignored here (handled by pod) typically,
		// but if we want to expose ports from the container specifically (uncommon in shared net),
		// we can. However, usually ports are on the Pod.
		// For now, if PodName is set, we skip port mapping on the container
		// OR we can assume the caller knows what they are doing.
		// If we are sharing net namespace, ports on container usually conflict or simply don't make sense
		// if they are meant to be external.
		// BUT the user plan says: "Pods define external ports; Apps define internal ports."
		// So if podName is set, we likely SHOULD NOT set PortMappings on the container spec
		// unless we want double mapping or something.
		// Let's omit port mappings if in a Pod, to be safe.
		if len(ports) > 0 {
			logger.DebugCtx(ctx, "Ignoring container ports because running in a Pod", "pod", podName)
		}
	case len(ports) > 0:
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
	default:
		logger.DebugCtx(ctx, "No port mappings provided")
	}

	if networkName != "" {
		logger.DebugCtx(ctx, "Setting network", "network", networkName)
		s.CNINetworks = []string{networkName}
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

// RunWithMounts creates and starts a container with volume mounts
func (c *Client) RunWithMounts(ctx context.Context, opts RunOptions) (string, error) {
	logger.DebugCtx(ctx, "Checking if image exists", "image", opts.Image)

	// Pull image if not exists
	exists, err := images.Exists(c.ctx, opts.Image, nil)
	if err != nil {
		return "", fmt.Errorf("checking image: %w", err)
	}

	if !exists {
		logger.InfoCtx(ctx, "Pulling image", "image", opts.Image)
		_, err = images.Pull(c.ctx, opts.Image, nil)
		if err != nil {
			return "", fmt.Errorf("pulling image: %w", err)
		}
		logger.DebugCtx(ctx, "Image pulled successfully", "image", opts.Image)
	}

	// Create spec
	s := specgen.NewSpecGenerator(opts.Image, false)
	s.Name = opts.Name
	s.Env = envSliceToMap(opts.Env)
	s.Labels = opts.Labels

	// Handle pod or ports
	switch {
	case opts.PodName != "":
		s.Pod = opts.PodName
		if len(opts.Ports) > 0 {
			logger.DebugCtx(ctx, "Ignoring container ports because running in a Pod", "pod", opts.PodName)
		}
	case len(opts.Ports) > 0:
		s.PortMappings = make([]nettypes.PortMapping, 0, len(opts.Ports))
		for hostPort, containerPort := range opts.Ports {
			logger.DebugCtx(ctx, "Adding port mapping",
				"host_port", hostPort,
				"container_port", containerPort,
			)
			s.PortMappings = append(s.PortMappings, nettypes.PortMapping{
				HostIP:        "0.0.0.0", // Caddy needs external access
				HostPort:      hostPort,
				ContainerPort: containerPort,
				Protocol:      "tcp",
			})
		}
	}

	// Handle network
	if opts.NetworkName != "" {
		logger.DebugCtx(ctx, "Setting network", "network", opts.NetworkName)
		s.CNINetworks = []string{opts.NetworkName}
	}

	// Handle mounts using Volumes (bind mounts)
	if len(opts.Mounts) > 0 {
		for _, m := range opts.Mounts {
			logger.DebugCtx(ctx, "Adding bind mount",
				"source", m.Source,
				"target", m.Target,
				"readonly", m.ReadOnly,
			)
			// Use Volumes for bind mounts in the format source:dest[:options]
			mountStr := m.Source + ":" + m.Target
			if m.ReadOnly {
				mountStr += ":ro"
			}
			s.Volumes = append(s.Volumes, &specgen.NamedVolume{
				Name: m.Source,
				Dest: m.Target,
			})
		}
	}

	// Create container
	logger.DebugCtx(ctx, "Creating container", "name", opts.Name)
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
		"name", opts.Name,
		"id", createResponse.ID[:12],
		"mounts", len(opts.Mounts),
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
	for i := range listContainers {
		name := ""
		ctr := &listContainers[i]

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
			Labels:  ctr.Labels,
			Created: ctr.Created,
			PodID:   ctr.Pod,
		})

		// Populate IP for running containers using Inspect (List doesn't provide it detailed enough)
		// This is N+1 but necessary for IP display until we find a better way or use events
		if ctr.State == "running" {
			// We modify the last element
			idx := len(result) - 1
			inspectData, err := containers.Inspect(c.ctx, ctr.ID, nil)
			if err == nil { // Ignore error, just don't show IP
				result[idx].IPAddress = getIPAddress(inspectData.NetworkSettings.Networks)
				result[idx].ExposedPorts = getExposedPorts(inspectData.Config.ExposedPorts)
				result[idx].Networks = getNetworkNames(inspectData.NetworkSettings.Networks)
			}
		}
	}

	logger.DebugCtx(ctx, "Found containers", "count", len(result))
	return result, nil
}

// Logs streams container logs
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

// GetContainer returns information about a specific container
func (c *Client) GetContainer(ctx context.Context, nameOrID string) (*ContainerInfo, error) {
	logger.DebugCtx(ctx, "Getting container info", "id", nameOrID)

	data, err := containers.Inspect(c.ctx, nameOrID, nil)
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	// formatInspectPorts formats port mappings from inspect data
	ports := formatInspectPorts(data.NetworkSettings.Ports)
	ip := getIPAddress(data.NetworkSettings.Networks)
	exposed := getExposedPorts(data.Config.ExposedPorts)
	networks := getNetworkNames(data.NetworkSettings.Networks)

	return &ContainerInfo{
		ID:           data.ID[:12],
		Name:         data.Name,
		Image:        data.ImageName,
		Status:       data.State.Status,
		Ports:        ports,
		Labels:       data.Config.Labels,
		Created:      data.Created,
		IPAddress:    ip,
		ExposedPorts: exposed,
		PodID:        data.Pod,
		Networks:     networks,
	}, nil
}

// InspectImage returns information about an image
func (c *Client) InspectImage(ctx context.Context, name string) (*ImageInfo, error) {
	logger.DebugCtx(ctx, "Inspecting image", "image", name)

	// Pull if not exists (optional, but good for inspection)
	exists, err := images.Exists(c.ctx, name, nil)
	if err != nil {
		return nil, fmt.Errorf("checking image: %w", err)
	}
	if !exists {
		logger.InfoCtx(ctx, "Pulling image for inspection", "image", name)
		_, err = images.Pull(c.ctx, name, nil)
		if err != nil {
			return nil, fmt.Errorf("pulling image: %w", err)
		}
	}

	data, err := images.GetImage(c.ctx, name, nil)
	if err != nil {
		return nil, fmt.Errorf("getting image info: %w", err)
	}

	exposedPorts := make([]string, 0)
	if data.Config != nil && len(data.Config.ExposedPorts) > 0 {
		for port := range data.Config.ExposedPorts {
			exposedPorts = append(exposedPorts, port)
		}
	}

	return &ImageInfo{
		ID:           data.ID[:12],
		ExposedPorts: exposedPorts,
	}, nil
}

// formatInspectPorts formats port mappings from inspect data
func formatInspectPorts(ports map[string][]define.InspectHostPort) map[string]string {
	result := make(map[string]string)
	if len(ports) == 0 {
		return result
	}

	for containerPort, hostPorts := range ports {
		if len(hostPorts) > 0 {
			// Just take the first one for now
			p := hostPorts[0]
			result[containerPort] = fmt.Sprintf("%s:%s", p.HostIP, p.HostPort)
		} else {
			result[containerPort] = ""
		}
	}
	return result
}

// getIPAddress extracts the primary IP address from networks
// We prioritize the bridge network or the user-defined network
func getIPAddress(networks map[string]*define.InspectAdditionalNetwork) string {
	if len(networks) == 0 {
		return ""
	}
	// Return the first non-empty IP found
	for _, net := range networks {
		if net.IPAddress != "" {
			return net.IPAddress
		}
	}
	return ""
}

// getSocketPath returns the Podman socket path based on environment
func getSocketPath() string {
	// Check CONTAINER_HOST first (used by podman-remote and systemd services)
	if sock := os.Getenv("CONTAINER_HOST"); sock != "" {
		return sock
	}

	// Check PODMAN_SOCK for backwards compatibility
	if sock := os.Getenv("PODMAN_SOCK"); sock != "" {
		return "unix://" + sock
	}

	// Check for system Podman socket (used when running as root or with podman.socket)
	systemSocket := "/run/podman/podman.sock"
	if _, err := os.Stat(systemSocket); err == nil {
		return "unix://" + systemSocket
	}

	// Check for macOS Podman Machine socket
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	macSocket := fmt.Sprintf("%s/.local/share/containers/podman/machine/podman.sock", homeDir)
	if _, err := os.Stat(macSocket); err == nil {
		return "unix://" + macSocket
	}

	// Default to rootless user socket
	return fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", os.Getuid())
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

// formatPorts returns port mappings as a map
func formatPorts(ports []nettypes.PortMapping) map[string]string {
	result := make(map[string]string)
	if len(ports) == 0 {
		return result
	}

	for _, p := range ports {
		containerPort := fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol)
		hostVal := ""
		if p.HostPort > 0 {
			if p.HostIP != "" {
				hostVal = fmt.Sprintf("%s:%d", p.HostIP, p.HostPort)
			} else {
				hostVal = fmt.Sprintf("%d", p.HostPort)
			}
		}
		result[containerPort] = hostVal
	}
	return result
}

// ptrBool creates a bool pointer
func ptrBool(b bool) *bool {
	return &b
}

// CreatePod creates a new pod
func (c *Client) CreatePod(ctx context.Context, name string, ports map[uint16]uint16) (string, error) {
	logger.DebugCtx(ctx, "Creating pod", "name", name)

	s := specgen.NewPodSpecGenerator()
	s.Name = name

	// Configure ports
	if len(ports) > 0 {
		s.PortMappings = make([]nettypes.PortMapping, 0, len(ports))
		for hostPort, containerPort := range ports {
			s.PortMappings = append(s.PortMappings, nettypes.PortMapping{
				HostIP:        "127.0.0.1", // Default to localhost for safety
				HostPort:      hostPort,
				ContainerPort: containerPort,
				Protocol:      "tcp",
			})
		}
	}

	// Create the pod
	// CreatePodFromSpec expects entities.PodSpec which wraps PodSpecGen
	spec := &entities.PodSpec{
		PodSpecGen: *s,
	}

	response, err := pods.CreatePodFromSpec(c.ctx, spec)
	if err != nil {
		return "", fmt.Errorf("creating pod: %w", err)
	}

	logger.InfoCtx(ctx, "Pod created", "name", name, "id", response.Id[:12])
	return response.Id, nil
}

// RemovePod removes a pod
func (c *Client) RemovePod(ctx context.Context, nameOrID string, force bool) error {
	logger.DebugCtx(ctx, "Removing pod", "name", nameOrID, "force", force)

	_, err := pods.Remove(c.ctx, nameOrID, &pods.RemoveOptions{Force: &force})
	if err != nil {
		return fmt.Errorf("removing pod: %w", err)
	}

	return nil
}

// PodExists checks if a pod exists
func (c *Client) PodExists(ctx context.Context, nameOrID string) (bool, error) {
	exists, err := pods.Exists(c.ctx, nameOrID, nil)
	if err != nil {
		return false, fmt.Errorf("checking pod existence: %w", err)
	}
	return exists, nil
}

// ListPods returns a list of all pods
func (c *Client) ListPods(ctx context.Context) ([]PodInfo, error) {
	logger.DebugCtx(ctx, "Listing pods")

	reports, err := pods.List(c.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	result := make([]PodInfo, 0, len(reports))
	for _, p := range reports {
		result = append(result, PodInfo{
			ID:      p.Id[:12],
			Name:    p.Name,
			Status:  p.Status,
			Created: p.Created,
		})
	}

	return result, nil
}

// InspectPod returns information about a specific pod
func (c *Client) InspectPod(ctx context.Context, nameOrID string) (*PodInfo, error) {
	logger.DebugCtx(ctx, "Inspecting pod", "name", nameOrID)

	data, err := pods.Inspect(c.ctx, nameOrID, nil)
	if err != nil {
		return nil, fmt.Errorf("inspecting pod: %w", err)
	}

	return &PodInfo{
		ID:      data.ID[:12],
		Name:    data.Name,
		Status:  data.State,
		Created: data.Created,
	}, nil
}

// CreateNetwork creates a new bridge network
func (c *Client) CreateNetwork(ctx context.Context, name string) (string, error) {
	logger.DebugCtx(ctx, "Creating network", "name", name)

	// In this version of bindings, it seems we pass the Network struct directly?
	// Based on error: want (context.Context, *"go.podman.io/common/libnetwork/types".Network)
	net := &nettypes.Network{
		Name:   name,
		Driver: "bridge",
	}

	// Assuming network.Create returns (*types.NetworkCreateReport, error) or similar
	// Let's try matching the signature "want (context.Context, *types.Network)"
	// Wait, network.Create in bindings v5 usually takes (*types.Network, *network.CreateOptions) or similar?
	// The error says it WANTS (ctx, *Network).
	// Let's try just passing the network struct.

	// Note: We might need to check if response is just the network struct back or a report.
	// If it returns (newNet, err), then we use newNet.ID.

	newNet, err := network.Create(c.ctx, net)
	if err != nil {
		return "", fmt.Errorf("creating network: %w", err)
	}

	logger.InfoCtx(ctx, "Network created", "name", name, "id", newNet.ID)
	return newNet.ID, nil
}

// RemoveNetwork removes a network
func (c *Client) RemoveNetwork(ctx context.Context, nameOrID string) error {
	logger.DebugCtx(ctx, "Removing network", "name", nameOrID)

	// Force removal? Maybe careful.
	force := false
	_, err := network.Remove(c.ctx, nameOrID, &network.RemoveOptions{Force: &force})
	if err != nil {
		return fmt.Errorf("removing network: %w", err)
	}

	return nil
}

// ListNetworks lists all networks
func (c *Client) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	logger.DebugCtx(ctx, "Listing networks")

	reports, err := network.List(c.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing networks: %w", err)
	}

	result := make([]NetworkInfo, 0, len(reports))
	for i := range reports {
		n := &reports[i]
		// Filter out standard networks if we want to show only user-created?
		// For now show all, maybe filter 'host', 'none' in UI
		subnet := ""
		if len(n.Subnets) > 0 {
			subnet = n.Subnets[0].Subnet.String()
		}

		result = append(result, NetworkInfo{
			ID:      n.ID[:12],
			Name:    n.Name,
			Driver:  n.Driver,
			Subnet:  subnet,
			Created: n.Created,
		})
	}
	return result, nil
}

// getExposedPorts extracts the exposed ports keys
func getExposedPorts(ports map[string]struct{}) []string {
	result := make([]string, 0, len(ports))
	for k := range ports {
		result = append(result, k)
	}
	return result
}

// getNetworkNames extracts the network names
func getNetworkNames(networks map[string]*define.InspectAdditionalNetwork) []string {
	result := make([]string, 0, len(networks))
	for k := range networks {
		result = append(result, k)
	}
	return result
}

// ConnectNetwork connects a container to a network
func (c *Client) ConnectNetwork(ctx context.Context, containerName, networkName string) error {
	logger.DebugCtx(ctx, "Connecting container to network", "container", containerName, "network", networkName)

	if err := network.Connect(c.ctx, networkName, containerName, nil); err != nil {
		return fmt.Errorf("connecting container %s to network %s: %w", containerName, networkName, err)
	}

	logger.InfoCtx(ctx, "Container connected to network", "container", containerName, "network", networkName)
	return nil
}

// DisconnectNetwork disconnects a container from a network
func (c *Client) DisconnectNetwork(ctx context.Context, containerName, networkName string) error {
	logger.DebugCtx(ctx, "Disconnecting container from network", "container", containerName, "network", networkName)

	if err := network.Disconnect(c.ctx, networkName, containerName, nil); err != nil {
		return fmt.Errorf("disconnecting container %s from network %s: %w", containerName, networkName, err)
	}

	logger.InfoCtx(ctx, "Container disconnected from network", "container", containerName, "network", networkName)
	return nil
}
