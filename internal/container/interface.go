package container

import (
	"context"
)

// ContainerManager defines the interface for container operations.
// This interface is used for mocking in tests.
type ContainerManager interface {
	Run(ctx context.Context, name, image string, ports map[uint16]uint16, env []string) (string, error)
	Stop(ctx context.Context, name string, timeout *uint) error
	Remove(ctx context.Context, name string, force bool) error
	List(ctx context.Context, all bool) ([]ContainerInfo, error)
	Logs(ctx context.Context, name string, follow bool, tail string) error
}

// Ensure Client implements ContainerManager
var _ ContainerManager = (*Client)(nil)
