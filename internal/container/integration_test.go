package container

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AkMo3/simplify/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Initialize Test
func TestMain(m *testing.M) {
	// Setup: Initialize logger before any tests run
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown: Sync logger
	logger.Sync()

	os.Exit(code)
}

// Integration tests require Podman to be running.
// Skip these tests if SKIP_INTEGRATION is set or Podman is not available.

func skipIfNoIntegration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test (SKIP_INTEGRATION is set)")
	}
}

func skipIfNoPodman(t *testing.T, ctx context.Context) *Client {
	skipIfNoIntegration(t)

	client, err := NewClient(ctx)
	if err != nil {
		t.Skipf("Skipping integration test (Podman not available): %v", err)
	}
	return client
}

// uniqueName generates a unique container name for tests
func uniqueName(prefix string) string {
	return prefix + "-" + time.Now().Format("150405")
}

// TestIntegration_RunAndRemove tests the full container lifecycle
func TestIntegration_RunAndRemove(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify")

	// Cleanup in case previous test failed
	_ = client.Remove(ctx, containerName, true)

	// Run a container
	id, err := client.Run(ctx, containerName, "docker.io/library/alpine:latest", nil, []string{"TEST_VAR=hello"}, nil, "")
	require.NoError(t, err, "Failed to run container")
	assert.NotEmpty(t, id, "Container ID should not be empty")
	assert.Len(t, id, 64, "Container ID should be 64 characters")

	// Cleanup
	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err, "Failed to remove container")
}

// TestIntegration_RunWithPorts tests running a container with port mappings
func TestIntegration_RunWithPorts(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-ports")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run with port mapping
	ports := map[uint16]uint16{
		18080: 80,
	}

	id, err := client.Run(ctx, containerName, "docker.io/library/nginx:alpine", ports, nil, nil, "")
	require.NoError(t, err, "Failed to run container with ports")
	assert.NotEmpty(t, id)

	// List and verify port is shown
	containers, err := client.List(ctx, true)
	require.NoError(t, err)

	found := false
	for _, c := range containers {
		if c.Name != containerName {
			continue
		}
		found = true
		break
	}
	assert.True(t, found, "Container should be in list")

	// Cleanup
	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err)
}

// TestIntegration_StopContainer tests stopping a running container
func TestIntegration_StopContainer(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-stop")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run a container (nginx stays running)
	_, err := client.Run(ctx, containerName, "docker.io/library/nginx:alpine", nil, nil, nil, "")
	require.NoError(t, err)

	// Verify it's running
	containers, err := client.List(ctx, false) // only running
	require.NoError(t, err)

	found := false
	for _, c := range containers {
		if c.Name == containerName {
			found = true
			break
		}
	}
	assert.True(t, found, "Container should be running")

	// Stop it
	timeout := uint(5)
	err = client.Stop(ctx, containerName, &timeout)
	require.NoError(t, err)

	// Verify it's not in running list
	containers, err = client.List(ctx, false)
	require.NoError(t, err)

	found = false
	for _, c := range containers {
		if c.Name == containerName {
			found = true
			break
		}
	}
	assert.False(t, found, "Container should not be running after stop")

	// But should be in all list
	containers, err = client.List(ctx, true)
	require.NoError(t, err)

	found = false
	for _, c := range containers {
		if c.Name == containerName {
			found = true
			break
		}
	}
	assert.True(t, found, "Container should be in all list")

	// Cleanup
	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err)
}

// TestIntegration_List tests listing containers
func TestIntegration_List(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-list")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run a container
	_, err := client.Run(ctx, containerName, "docker.io/library/alpine:latest", nil, nil, nil, "")
	require.NoError(t, err)

	// List all containers
	containers, err := client.List(ctx, true)
	require.NoError(t, err)

	// Find our container
	found := false
	for _, c := range containers {
		if c.Name != containerName {
			continue
		}
		found = true
		assert.Equal(t, 12, len(c.ID), "ID should be truncated to 12 chars")
		assert.Contains(t, c.Image, "alpine")
		assert.NotZero(t, c.Created)
		break
	}
	assert.True(t, found, "Our container should be in the list")

	// Cleanup
	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err)
}

// TestIntegration_RemoveForce tests force removing a running container
func TestIntegration_RemoveForce(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-force")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run a container (nginx stays running)
	_, err := client.Run(ctx, containerName, "docker.io/library/nginx:alpine", nil, nil, nil, "")
	require.NoError(t, err)

	// Try to remove without force - should fail
	err = client.Remove(ctx, containerName, false)
	assert.Error(t, err, "Should fail to remove running container without force")

	// Remove with force - should succeed
	err = client.Remove(ctx, containerName, true)
	assert.NoError(t, err, "Should succeed with force")

	// Verify it's gone
	containers, err := client.List(ctx, true)
	require.NoError(t, err)

	for _, c := range containers {
		assert.NotEqual(t, containerName, c.Name, "Container should be removed")
	}
}

// TestIntegration_RemoveNonExistent tests removing a container that doesn't exist
func TestIntegration_RemoveNonExistent(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	err := client.Remove(ctx, "non-existent-container-12345", false)
	assert.Error(t, err, "Should fail when container doesn't exist")
}

// TestIntegration_RunDuplicateName tests that running with duplicate name fails
func TestIntegration_RunDuplicateName(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-dup")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run first container
	_, err := client.Run(ctx, containerName, "docker.io/library/alpine:latest", nil, nil, nil, "")
	require.NoError(t, err)

	// Try to run with same name - should fail
	_, err = client.Run(ctx, containerName, "docker.io/library/alpine:latest", nil, nil, nil, "")
	assert.Error(t, err, "Should fail when container with same name exists")

	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err)
}

// TestIntegration_RunWithLabels tests running a container with labels
func TestIntegration_RunWithLabels(t *testing.T) {
	ctx := context.Background()
	client := skipIfNoPodman(t, ctx)

	containerName := uniqueName("test-simplify-labels")

	// Cleanup
	_ = client.Remove(ctx, containerName, true)

	// Run with labels
	labels := map[string]string{
		"com.example.managed": "true",
		"com.example.id":      "12345",
	}

	id, err := client.Run(ctx, containerName, "docker.io/library/alpine:latest", nil, nil, labels, "")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// List and verify labels
	containers, err := client.List(ctx, true)
	require.NoError(t, err)

	found := false
	for _, c := range containers {
		if c.Name != containerName {
			continue
		}
		found = true
		assert.Equal(t, "true", c.Labels["com.example.managed"])
		assert.Equal(t, "12345", c.Labels["com.example.id"])
		break
	}
	assert.True(t, found, "Container should be in list")

	// GetContainer and verify labels
	info, err := client.GetContainer(ctx, containerName)
	require.NoError(t, err)
	assert.Equal(t, "true", info.Labels["com.example.managed"])
	assert.Equal(t, "12345", info.Labels["com.example.id"])

	// Cleanup
	err = client.Remove(ctx, containerName, true)
	require.NoError(t, err)
}
