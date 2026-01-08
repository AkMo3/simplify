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
	"go.uber.org/zap"
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
		logger.Error("Reconciliation Failed", err)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping reconciliation loop")
			return
		case <-ticker.C:
			if err := w.reconcile(ctx); err != nil {
				logger.Error("Reconciliation Failed", err)
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
	// Key: Container new, Value ContainerInfo
	existingContainers := make(map[string]container.ContainerInfo)

	for _, c := range containers {
		// Only track cotnainers managed by us (prefixed with "simplify-")
		if strings.HasPrefix(c.Name, "simplify-") {
			existingContainers[c.Name] = c
		}
	}

	desiredContainerNames := make(map[string]bool)

	for _, app := range apps {
		// Construct the expected container name
		// For MVP, we assume 1 replica per app with name "simplify-<appID>"
		// TODO: In the future, this loop will run 'app.Replicas' times.
		containerName := fmt.Sprintf("simplify-%s", app.ID)
		desiredContainerNames[containerName] = true

		// Check if it exists
		if info, exists := existingContainers[containerName]; exists {
			// if exists, is it running?
			if info.Status != "running" && !strings.HasPrefix(info.Status, "Up") {
				// For Podman, simpler to remove and recreate to ensure config is fresh
				// TODO: In production, we'd use Start() if config matches.
				if err := w.container.Remove(ctx, containerName, true); err != nil {
					logger.Error("Failed to remove stopped container", zap.Error(err))
					continue
				}

				// It will be recreated below since we removed it from the map (conceptually)
				// actually we just fall through to the create logic if we deleted it?
				// No, let's just trigger a re-creation in next loop or do it now.
				// For simplicity: Remove now, and let next loop create it.
				continue
			}
			// It's running. Check if config matches? (Skipped for MVP Phase 3)
		} else {
			// It does not exist. Create it.
			logger.Info("Deploying missing application", zap.String("app", app.Name))
			if err := w.deployApp(ctx, &app, containerName); err != nil {
				logger.Error("Failed to deploy app", zap.String("app", app.Name), zap.Error(err))
			}
		}
	}

	// 4. Cleanup Orphans (Actual -> Desired)
	// Any container starting with "simplify-" that is NOT in desiredContainerNames
	for name := range existingContainers {
		if !desiredContainerNames[name] {
			logger.Info("Removing orphaned container", zap.String("container", name))
			if err := w.container.Remove(ctx, name, true); err != nil {
				logger.Error("Failed to remove orphan", zap.Error(err))
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

	// Convert EnvVars map -> []string
	var env []string
	for k, v := range app.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Call Container Client
	_, err = w.container.Run(ctx, containerName, app.Image, ports, env)
	return err
}

// Helper: Parse "80:80" strings into uint16 map
func parsePorts(raw map[string]string) (map[uint16]uint16, error) {
	result := make(map[uint16]uint16)
	for host, container := range raw {
		h, err := strconv.ParseUint(host, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid host port %s: %w", host, err)
		}
		c, err := strconv.ParseUint(container, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid container port %s: %w", container, err)
		}
		result[uint16(h)] = uint16(c)
	}
	return result, nil
}
