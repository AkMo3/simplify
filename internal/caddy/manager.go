// Package caddy provides management for the Caddy reverse proxy container
package caddy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/logger"
)

const (
	containerName     = "simplify-caddy"
	caddyfileTemplate = `# Simplify Caddy Configuration
# Auto-generated - do not edit manually

{
    admin :2019
}

{{if .SimplifyHost}}
{{.SimplifyHost}} {
    # API routes
    handle /api/* {
        reverse_proxy {{.APIUpstream}}
    }
    handle /healthz {
        reverse_proxy {{.APIUpstream}}
    }
    handle /readyz {
        reverse_proxy {{.APIUpstream}}
    }
    
    # Frontend SPA
    handle {
        root * /srv
        file_server
        try_files {path} /index.html
    }
    
    encode gzip
}
{{end}}

{{range .Apps}}
{{.Hostname}} {
    reverse_proxy {{.Upstream}}
}
{{end}}
`
)

// AppRoute represents a route for a user application
type AppRoute struct {
	Hostname string // e.g., "alice.example.com"
	Upstream string // e.g., "alice:3443"
}

// BuildRoutes creates AppRoutes from a list of proxy-enabled applications
func BuildRoutes(apps []core.Application) []AppRoute {
	routes := make([]AppRoute, 0)
	for _, app := range apps {
		if app.ProxyEnabled && app.ProxyHostname != "" {
			port := app.ProxyPort
			if port == 0 {
				port = 80 // Default to port 80 if not specified
			}
			routes = append(routes, AppRoute{
				Hostname: app.ProxyHostname,
				Upstream: fmt.Sprintf("%s:%d", sanitizeName(app.Name), port),
			})
		}
	}
	return routes
}

// sanitizeName ensures the container name is valid for DNS
func sanitizeName(name string) string {
	// Simple sanitization - lowercase and replace spaces
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+32) // lowercase
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '-')
		}
	}
	return string(result)
}

// Manager handles Caddy container lifecycle and configuration
type Manager struct {
	container  container.ContainerManager
	cfg        *config.CaddyConfig
	serverPort int
}

// New creates a new Caddy manager
func New(containerClient container.ContainerManager, cfg *config.CaddyConfig, serverPort int) *Manager {
	return &Manager{
		container:  containerClient,
		cfg:        cfg,
		serverPort: serverPort,
	}
}

// EnsureRunning ensures the Caddy container is running
func (m *Manager) EnsureRunning(ctx context.Context) error {
	if !m.cfg.Enabled {
		logger.DebugCtx(ctx, "Caddy is disabled, skipping")
		return nil
	}

	// Check if container exists and is running
	info, err := m.container.GetContainer(ctx, containerName)
	if err == nil && info.Status == "running" {
		logger.DebugCtx(ctx, "Caddy container already running", "id", info.ID)
		return nil
	}

	// Ensure directories exist
	if err := m.ensureDirectories(); err != nil {
		return fmt.Errorf("ensuring directories: %w", err)
	}

	// Generate initial Caddyfile
	if err := m.generateCaddyfile(nil); err != nil {
		return fmt.Errorf("generating Caddyfile: %w", err)
	}

	// Remove existing container if exists (might be stopped)
	if info != nil {
		logger.InfoCtx(ctx, "Removing existing Caddy container", "status", info.Status)
		if err := m.container.Remove(ctx, containerName, true); err != nil {
			logger.WarnCtx(ctx, "Failed to remove existing Caddy container", "error", err)
		}
	}

	// Start Caddy container
	return m.startContainer(ctx)
}

// Stop stops the Caddy container
func (m *Manager) Stop(ctx context.Context) error {
	if !m.cfg.Enabled {
		return nil
	}

	logger.InfoCtx(ctx, "Stopping Caddy container")
	timeout := uint(10)
	return m.container.Stop(ctx, containerName, &timeout)
}

// Reload reloads Caddy configuration via Admin API
func (m *Manager) Reload(ctx context.Context, apps []AppRoute) error {
	if !m.cfg.Enabled {
		return nil
	}

	// Generate new Caddyfile
	if err := m.generateCaddyfile(apps); err != nil {
		return fmt.Errorf("generating Caddyfile: %w", err)
	}

	// Read the new config
	caddyfilePath := filepath.Join(m.cfg.DataDir, "Caddyfile")
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return fmt.Errorf("reading Caddyfile: %w", err)
	}

	// POST to Caddy admin API
	adminURL := fmt.Sprintf("http://127.0.0.1:%d/load", m.cfg.AdminPort)
	resp, err := http.Post(adminURL, "text/caddyfile", bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("calling Caddy admin API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Caddy reload failed: %s - %s", resp.Status, string(body))
	}

	logger.InfoCtx(ctx, "Caddy configuration reloaded")
	return nil
}

// ensureDirectories creates required directories
func (m *Manager) ensureDirectories() error {
	dirs := []string{
		m.cfg.DataDir,
		filepath.Join(m.cfg.DataDir, "www"),
		filepath.Join(m.cfg.DataDir, "caddy_data"),
		filepath.Join(m.cfg.DataDir, "caddy_config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateCaddyfile generates the Caddyfile from template
func (m *Manager) generateCaddyfile(apps []AppRoute) error {
	tmpl, err := template.New("caddyfile").Parse(caddyfileTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	data := struct {
		SimplifyHost string
		APIUpstream  string
		Apps         []AppRoute
	}{
		SimplifyHost: "localhost", // Default to localhost for development
		APIUpstream:  fmt.Sprintf("host.containers.internal:%d", m.serverPort),
		Apps:         apps,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	caddyfilePath := filepath.Join(m.cfg.DataDir, "Caddyfile")
	if err := os.WriteFile(caddyfilePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing Caddyfile: %w", err)
	}

	return nil
}

// startContainer starts the Caddy container with volume mounts
func (m *Manager) startContainer(ctx context.Context) error {
	logger.InfoCtx(ctx, "Starting Caddy container",
		"image", m.cfg.Image,
		"http_port", m.cfg.HTTPPort,
		"https_port", m.cfg.HTTPSPort,
	)

	// Port mappings: hostPort -> containerPort
	ports := map[uint16]uint16{
		uint16(m.cfg.HTTPPort):  80,
		uint16(m.cfg.HTTPSPort): 443,
		uint16(m.cfg.AdminPort): 2019,
	}

	// Labels for identification
	labels := map[string]string{
		"simplify.managed": "true",
		"simplify.system":  "caddy",
	}

	// Volume mounts
	mounts := []container.Mount{
		{
			Source: filepath.Join(m.cfg.DataDir, "Caddyfile"),
			Target: "/etc/caddy/Caddyfile",
		},
		{
			Source: filepath.Join(m.cfg.DataDir, "caddy_data"),
			Target: "/data",
		},
		{
			Source: filepath.Join(m.cfg.DataDir, "caddy_config"),
			Target: "/config",
		},
	}

	// Add frontend mount if FrontendPath is configured
	if m.cfg.FrontendPath != "" {
		mounts = append(mounts, container.Mount{
			Source:   m.cfg.FrontendPath,
			Target:   "/srv",
			ReadOnly: true,
		})
		logger.InfoCtx(ctx, "Mounting frontend", "path", m.cfg.FrontendPath)
	} else {
		// Default to www directory in data dir
		mounts = append(mounts, container.Mount{
			Source: filepath.Join(m.cfg.DataDir, "www"),
			Target: "/srv",
		})
	}

	opts := container.RunOptions{
		Name:   containerName,
		Image:  m.cfg.Image,
		Ports:  ports,
		Env:    []string{},
		Labels: labels,
		Mounts: mounts,
		// Do not set NetworkName initially - start on default bridge for external connectivity
	}

	_, err := m.container.RunWithMounts(ctx, opts)
	if err != nil {
		return fmt.Errorf("starting Caddy container: %w", err)
	}

	// Connect to proxy network for internal communication
	if m.cfg.ProxyNetwork != "" {
		if err := m.container.ConnectNetwork(ctx, containerName, m.cfg.ProxyNetwork); err != nil {
			logger.ErrorCtx(ctx, "Failed to connect Caddy to proxy network", "network", m.cfg.ProxyNetwork, "error", err)
			// Don't fail completely, but warn
		} else {
			logger.DebugCtx(ctx, "Connected Caddy to proxy network", "network", m.cfg.ProxyNetwork)
		}
	}

	logger.InfoCtx(ctx, "Caddy container started")
	return nil
}
