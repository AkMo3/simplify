// Package reconciler maintains the sync between database and physical containers
package reconciler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AkMo3/simplify/internal/container"
	"github.com/AkMo3/simplify/internal/core"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/AkMo3/simplify/internal/store"
)

// Worker is responsible for reconciling desired state (DB) with actual state (Podman)
type Worker struct {
	store     *store.Store
	container container.ContainerManager
}

// New creates a new reconciler worker
func New(storeObj *store.Store, containerClient container.ContainerManager) *Worker {
	return &Worker{
		store:     storeObj,
		container: containerClient,
	}
}

// Start runs the reconciliation loop in a blocking manner
func (w *Worker) Start(ctx context.Context) {
	logger.Info("Starting reconciliation loop")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run once immediately
	if err := w.reconcile(ctx); err != nil {
		logger.Error("Reconciliation failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping reconciliation loop")
			return
		case <-ticker.C:
			if err := w.reconcile(ctx); err != nil {
				logger.Error("Reconciliation failed", "error", err)
			}
		}
	}
}

func (w *Worker) reconcile(ctx context.Context) error {
	apps, err := w.store.ListApplications()
	if err != nil {
		return fmt.Errorf("failed to list applications: %w", err)
	}

	containers, err := w.container.List(ctx, true) // true = include stopped
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Map AppID -> ContainerInfo for reconciliation
	existingApps := make(map[string]container.ContainerInfo)
	// Map ContainerName -> bool for orphan tracking
	managedContainers := make(map[string]bool)

	for _, c := range containers {
		// Check for managed label
		isManaged := false
		appID := ""

		if val, ok := c.Labels["simplify.managed"]; ok && val == "true" {
			isManaged = true
			appID = c.Labels["simplify.app.id"]
		} else if strings.HasPrefix(c.Name, "simplify-") {
			// Legacy/Fallback: Check prefix
			isManaged = true
			appID = strings.TrimPrefix(c.Name, "simplify-")
		}

		if isManaged {
			if appID != "" {
				existingApps[appID] = c
			}
			managedContainers[c.Name] = true
		}
	}

	desiredContainerNames := make(map[string]bool)

	// Use index to avoid copying large struct on each iteration
	for i := range apps {
		app := &apps[i]

		// Construct the expected container name
		// We use app.Name but sanitized
		containerName := sanitizeName(app.Name)
		if containerName == "" {
			// Fallback
			containerName = fmt.Sprintf("simplify-%s", app.ID)
		}
		// Desired state tracking
		desiredContainerNames[containerName] = true

		// Check if it exists by AppID
		if info, exists := existingApps[app.ID]; exists {
			// If exists, is it running?
			// Update the container name for tracking desired state (in case we want to reference it)
			// But wait, desiredContainerNames is used for orphan cleanup.
			// Current implementation of orphan cleanup iterates existingContainers.
			// We should track which containers are KEPT.
			desiredContainerNames[info.Name] = true

			if info.Status != "running" && !strings.HasPrefix(info.Status, "Up") {
				// For Podman, simpler to remove and recreate to ensure config is fresh
				// TODO: In production, we'd use Start() if config matches.
				if err := w.container.Remove(ctx, info.Name, true); err != nil {
					logger.Error("Failed to remove stopped container",
						"container", info.Name,
						"error", err)
					continue
				}
				// It will be recreated in next loop iteration
				continue
			}
			// It's running. Check if config matches
			if !checkPortsMatch(app.Ports, info.Ports) {
				logger.Info("Container ports mismatch, recreating",
					"container", info.Name,
					"app_id", app.ID)
				if err := w.container.Remove(ctx, info.Name, true); err != nil {
					logger.Error("Failed to remove container with outgoing ports",
						"container", info.Name,
						"error", err)
					continue
				}
				// Recreated in next loop
				continue
			}
		} else {
			// It does not exist. Create it.
			logger.Info("Deploying missing application",
				"app", app.Name,
				"id", app.ID)
			// We still use standard naming for creation, but now with labels
			desiredContainerNames[containerName] = true
			if err := w.deployApp(ctx, app, containerName); err != nil {
				logger.Error("Failed to deploy app",
					"app", app.Name,
					"error", err)
			}
		}
	}

	// Cleanup Orphans (Actual -> Desired)
	// Any container starting with "simplify-" that is NOT in desiredContainerNames
	// Cleanup Orphans (Actual -> Desired)
	// Any managed container that is NOT in desiredContainerNames
	for name := range managedContainers {
		if !desiredContainerNames[name] {
			logger.Info("Removing orphaned container", "container", name)
			if err := w.container.Remove(ctx, name, true); err != nil {
				logger.Error("Failed to remove orphan",
					"container", name,
					"error", err)
			}
		}
	}

	return nil
}

// deployApp handles the specific logic of converting App struct to Container args
func (w *Worker) deployApp(ctx context.Context, app *core.Application, containerName string) error {
	// Convert Ports map[string]string -> map[uint16]uint16
	// Format "8080:80" -> Host:Container
	ports, err := parsePorts(app.Ports)
	if err != nil {
		return fmt.Errorf("invalid ports: %w", err)
	}

	// Convert EnvVars map -> []string with pre-allocation
	env := make([]string, 0, len(app.EnvVars))
	for k, v := range app.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Define Labels
	labels := map[string]string{
		"simplify.managed":  "true",
		"simplify.app.id":   app.ID,
		"simplify.app.name": app.Name,
	}

	// Call Container Client
	_, err = w.container.Run(ctx, containerName, app.Image, ports, env, labels)
	return err
}

// parsePorts converts "80:80" strings into uint16 map
func parsePorts(raw map[string]string) (map[uint16]uint16, error) {
	result := make(map[uint16]uint16, len(raw))
	for host, containerPort := range raw {
		// Clean host port
		hStr := cleanPortString(host)
		h, err := strconv.ParseUint(hStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid host port %s (cleaned: %s): %w", host, hStr, err)
		}

		// Clean container port
		cStr := cleanPortString(containerPort)
		c, err := strconv.ParseUint(cStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid container port %s (cleaned: %s): %w", containerPort, cStr, err)
		}

		result[uint16(h)] = uint16(c)
	}
	return result, nil
}

// cleanPortString strips protocol and IP to return just the port number
func cleanPortString(s string) string {
	// Remove protocol suffix (e.g., /tcp)
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}

	// Remove IP prefix (e.g., 127.0.0.1:8080)
	// We take the last part after the last colon
	if idx := strings.LastIndex(s, ":"); idx != -1 {
		s = s[idx+1:]
	}

	return s
}

// sanitizeName ensures the container name is valid for Podman/Docker
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")
	// Remove invalid chars
	// Only [a-z0-9][a-z0-9_.-]* are allowed
	// For simplicity, we just keep a-z0-9 and -
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// checkPortsMatch checks if the application ports match the container ports
// appPorts: map[HostPort]ContainerPort (e.g. "8080": "80")
// containerPorts: map[ContainerPort/Proto]HostIP:HostPort (e.g. "80/tcp": "0.0.0.0:8080")
func checkPortsMatch(appPorts, containerPorts map[string]string) bool {
	// If counts don't match, they differ
	// Note: containerPorts might have extra entries for different protocols or bindings?
	// For simplicity, we check if every appPort exists in containerPorts with correct mapping.
	// And stricter: count should match effectively.
	// However, containerPorts keys includes protocol (80/tcp). App ports don't specify protocol yet (assume tcp).

	if len(appPorts) != len(containerPorts) {
		return false
	}

	for hostPort, containerPort := range appPorts {
		// key in containerPorts should be "containerPort/tcp" (default)
		// We need to iterate containerPorts to find the matching container port ignoring protocol?
		// Or assume TCP for now as per current limitation.
		key := fmt.Sprintf("%s/tcp", containerPort)

		val, ok := containerPorts[key]
		if !ok {
			// Try without protocol if simple key? Unlikely given podmain implementation
			return false
		}

		// Val is "IP:HostPort" or just "HostPort" (if no IP, but podmain adds IP)
		// We need to check if hostPort is in val
		if !strings.HasSuffix(val, ":"+hostPort) && val != hostPort {
			return false
		}
	}

	return true
}
