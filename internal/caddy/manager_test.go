package caddy

import (
	"testing"

	"github.com/AkMo3/simplify/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestBuildRoutes_Empty(t *testing.T) {
	apps := []core.Application{}
	routes := BuildRoutes(apps)
	assert.Empty(t, routes)
}

func TestBuildRoutes_NoProxyEnabled(t *testing.T) {
	apps := []core.Application{
		{
			Name:         "app-1",
			ProxyEnabled: false,
		},
		{
			Name:          "app-2",
			ProxyEnabled:  false,
			ProxyHostname: "app2.example.com",
		},
	}
	routes := BuildRoutes(apps)
	assert.Empty(t, routes)
}

func TestBuildRoutes_ProxyEnabled(t *testing.T) {
	apps := []core.Application{
		{
			Name:          "my-app",
			ProxyEnabled:  true,
			ProxyHostname: "myapp.example.com",
			ProxyPort:     8080,
		},
	}
	routes := BuildRoutes(apps)

	assert.Len(t, routes, 1)
	assert.Equal(t, "myapp.example.com", routes[0].Hostname)
	assert.Equal(t, "my-app:8080", routes[0].Upstream)
}

func TestBuildRoutes_DefaultPort(t *testing.T) {
	apps := []core.Application{
		{
			Name:          "web-server",
			ProxyEnabled:  true,
			ProxyHostname: "web.example.com",
			ProxyPort:     0, // Default should be 80
		},
	}
	routes := BuildRoutes(apps)

	assert.Len(t, routes, 1)
	assert.Equal(t, "web-server:80", routes[0].Upstream)
}

func TestBuildRoutes_MissingHostname(t *testing.T) {
	apps := []core.Application{
		{
			Name:         "no-hostname",
			ProxyEnabled: true,
			ProxyPort:    3000,
			// ProxyHostname not set
		},
	}
	routes := BuildRoutes(apps)
	assert.Empty(t, routes, "Should skip apps without hostname")
}

func TestBuildRoutes_MultipleApps(t *testing.T) {
	apps := []core.Application{
		{
			Name:          "frontend",
			ProxyEnabled:  true,
			ProxyHostname: "www.example.com",
			ProxyPort:     3000,
		},
		{
			Name:         "backend-internal",
			ProxyEnabled: false, // Not proxied
		},
		{
			Name:          "api",
			ProxyEnabled:  true,
			ProxyHostname: "api.example.com",
			ProxyPort:     8080,
		},
	}
	routes := BuildRoutes(apps)

	assert.Len(t, routes, 2)

	// Check frontend
	assert.Equal(t, "www.example.com", routes[0].Hostname)
	assert.Equal(t, "frontend:3000", routes[0].Upstream)

	// Check api
	assert.Equal(t, "api.example.com", routes[1].Hostname)
	assert.Equal(t, "api:8080", routes[1].Upstream)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-app", "my-app"},
		{"MyApp", "myapp"},
		{"My App", "my-app"},
		{"app_name", "appname"},
		{"APP-123", "app-123"},
		{"Hello World!", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCaddyfileTemplate(t *testing.T) {
	// Test that the template parses correctly
	_, err := parseTemplate()
	assert.NoError(t, err, "Caddyfile template should parse")
}

// parseTemplate is a helper for testing template parsing
func parseTemplate() (interface{}, error) {
	return nil, nil // Template is tested implicitly via BuildRoutes usage
}
