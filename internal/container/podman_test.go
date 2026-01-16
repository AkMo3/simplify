package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
	nettypes "go.podman.io/common/libnetwork/types"
)

// TestEnvSliceToMap tests the envionment variable parsing
func TestEnvSliceToMap(t *testing.T) {
	tests := []struct {
		expected map[string]string
		name     string
		input    []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:  "single variable",
			input: []string{"FOO=bar"},
			expected: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name:  "multiple variables",
			input: []string{"FOO=bar", "BAZ=qux", "DB_HOST=localhost"},
			expected: map[string]string{
				"FOO":     "bar",
				"BAZ":     "qux",
				"DB_HOST": "localhost",
			},
		},
		{
			name:  "value with equals sign",
			input: []string{"CONNECTION=host=localhost;port=5432"},
			expected: map[string]string{
				"CONNECTION": "host=localhost;port=5432",
			},
		},
		{
			name:  "empty value",
			input: []string{"EMPTY="},
			expected: map[string]string{
				"EMPTY": "",
			},
		},
		{
			name:     "no equals sign (invalid, should be skipped)",
			input:    []string{"INVALID"},
			expected: map[string]string{},
		},
		{
			name:  "mixed valid and invalid",
			input: []string{"VALID=value", "INVALID", "ALSO_VALID=123"},
			expected: map[string]string{
				"VALID":      "value",
				"ALSO_VALID": "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := envSliceToMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

}

// TestFormatPorts tests the port formatting for display
func TestFormatPorts(t *testing.T) {
	tests := []struct {
		expected map[string]string
		name     string
		input    []nettypes.PortMapping
	}{
		{
			name:     "empty ports",
			input:    []nettypes.PortMapping{},
			expected: map[string]string{},
		},
		{
			name: "single port with host binding",
			input: []nettypes.PortMapping{
				{HostIP: "0.0.0.0", HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
			expected: map[string]string{
				"80/tcp": "0.0.0.0:8080",
			},
		},
		{
			name: "multiple ports",
			input: []nettypes.PortMapping{
				{HostIP: "0.0.0.0", HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
				{HostIP: "0.0.0.0", HostPort: 443, ContainerPort: 443, Protocol: "tcp"},
			},
			expected: map[string]string{
				"80/tcp":  "0.0.0.0:8080",
				"443/tcp": "0.0.0.0:443",
			},
		},
		{
			name: "container port only (no host binding)",
			input: []nettypes.PortMapping{
				{HostIP: "", HostPort: 0, ContainerPort: 80, Protocol: "tcp"},
			},
			expected: map[string]string{
				"80/tcp": "",
			},
		},
		{
			name: "localhost binding",
			input: []nettypes.PortMapping{
				{HostIP: "127.0.0.1", HostPort: 3000, ContainerPort: 3000, Protocol: "tcp"},
			},
			expected: map[string]string{
				"3000/tcp": "127.0.0.1:3000",
			},
		},
		{
			name: "udp protocol",
			input: []nettypes.PortMapping{
				{HostIP: "0.0.0.0", HostPort: 53, ContainerPort: 53, Protocol: "udp"},
			},
			expected: map[string]string{
				"53/udp": "0.0.0.0:53",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPorts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPtrBool tests the pointer helper
func TestPtrBool(t *testing.T) {
	truePtr := ptrBool(true)
	assert.NotNil(t, truePtr)
	assert.True(t, *truePtr)

	falsePtr := ptrBool(false)
	assert.NotNil(t, falsePtr)
	assert.False(t, *falsePtr)
}

// TestGetSocketPath tests socket path detection
func TestGetSocketPath(t *testing.T) {
	// Just verify it returns something non-empty
	// Actual path depends on the system
	socketPath := getSocketPath()
	assert.NotEmpty(t, socketPath)
	assert.Contains(t, socketPath, "unix://")
}

// TestGetSocketPath_WithEnvVar tests PODMAN_SOCK override
func TestGetSocketPath_WithEnvVar(t *testing.T) {
	t.Setenv("PODMAN_SOCK", "/custom/path/podman.sock")

	socketPath := getSocketPath()
	assert.Equal(t, "unix:///custom/path/podman.sock", socketPath)
}
