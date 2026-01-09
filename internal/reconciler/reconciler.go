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
	container *container.Client
}

// New creates a new reconciler worker
func New(storeObj *store.Store, containerClient *container.Client) *Worker {
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

	// Create map of existing containers for quick lookup
	// Key: Container name, Value: ContainerInfo
	existingContainers := make(map[string]container.ContainerInfo)

	for _, c := range containers {
		// Only track containers managed by us (prefixed with "simplify-")
		if strings.HasPrefix(c.Name, "simplify-") {
			existingContainers[c.Name] = c
		}
	}

	desiredContainerNames := make(map[string]bool)

	// Use index to avoid copying large struct on each iteration
	for i := range apps {
		app := &apps[i]

		// Construct the expected container name
		// For MVP, we assume 1 replica per app with name "simplify-<appID>"
		// TODO: In the future, this loop will run 'app.Replicas' times.
		containerName := fmt.Sprintf("simplify-%s", app.ID)
		desiredContainerNames[containerName] = true

		// Check if it exists
		if info, exists := existingContainers[containerName]; exists {
			// If exists, is it running?
			if info.Status != "running" && !strings.HasPrefix(info.Status, "Up") {
				// For Podman, simpler to remove and recreate to ensure config is fresh
				// TODO: In production, we'd use Start() if config matches.
				if err := w.container.Remove(ctx, containerName, true); err != nil {
					logger.Error("Failed to remove stopped container",
						"container", containerName,
						"error", err)
					continue
				}
				// It will be recreated in next loop iteration
				continue
			}
			// It's running. Check if config matches? (Skipped for MVP Phase 3)
		} else {
			// It does not exist. Create it.
			logger.Info("Deploying missing application",
				"app", app.Name,
				"id", app.ID)
			if err := w.deployApp(ctx, app, containerName); err != nil {
				logger.Error("Failed to deploy app",
					"app", app.Name,
					"error", err)
			}
		}
	}

	// Cleanup Orphans (Actual -> Desired)
	// Any container starting with "simplify-" that is NOT in desiredContainerNames
	for name := range existingContainers {
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

	// Call Container Client
	_, err = w.container.Run(ctx, containerName, app.Image, ports, env)
	return err
}

// parsePorts converts "80:80" strings into uint16 map
func parsePorts(raw map[string]string) (map[uint16]uint16, error) {
	result := make(map[uint16]uint16, len(raw))
	for host, containerPort := range raw {
		h, err := strconv.ParseUint(host, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid host port %s: %w", host, err)
		}
		c, err := strconv.ParseUint(containerPort, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid container port %s: %w", containerPort, err)
		}
		result[uint16(h)] = uint16(c)
	}
	return result, nil
}
