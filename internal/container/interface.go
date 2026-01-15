package container

import (
	"context"
	"time"
)

// ContainerManager defines the interface for container operations.
// This interface is used for mocking in tests.
type ContainerManager interface {
	Run(ctx context.Context, name, image string, ports map[uint16]uint16, env []string, labels map[string]string, podName string) (string, error)
	Stop(ctx context.Context, name string, timeout *uint) error
	Remove(ctx context.Context, name string, force bool) error
	List(ctx context.Context, all bool) ([]ContainerInfo, error)
	Logs(ctx context.Context, name string, follow bool, tail string) error
	GetContainer(ctx context.Context, nameOrID string) (*ContainerInfo, error)
	InspectImage(ctx context.Context, image string) (*ImageInfo, error)
	CreatePod(ctx context.Context, name string, ports map[uint16]uint16) (string, error)
	RemovePod(ctx context.Context, nameOrID string, force bool) error
	PodExists(ctx context.Context, nameOrID string) (bool, error)
	ListPods(ctx context.Context) ([]PodInfo, error)
	InspectPod(ctx context.Context, nameOrID string) (*PodInfo, error)
	CreateNetwork(ctx context.Context, name string) (string, error)
	RemoveNetwork(ctx context.Context, nameOrID string) error
	ListNetworks(ctx context.Context) ([]NetworkInfo, error)
}

// ImageInfo holds image metadata
type ImageInfo struct {
	ID           string   `json:"id"`
	ExposedPorts []string `json:"exposed_ports"`
}

// PodInfo holds pod metadata from the container engine
type PodInfo struct {
	ID      string
	Name    string
	Status  string
	Created time.Time
}

// NetworkInfo holds network metadata from the container engine
type NetworkInfo struct {
	ID      string
	Name    string
	Driver  string
	Subnet  string
	Created time.Time
}

// Ensure Client implements ContainerManager
var _ ContainerManager = (*Client)(nil)
