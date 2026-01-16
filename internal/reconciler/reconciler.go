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
	// 1. Reconcile Pods
	if err := w.reconcilePods(ctx); err != nil {
		return fmt.Errorf("failed to reconcile pods: %w", err)
	}

	// 2. Reconcile Applications
	return w.reconcileApps(ctx)
}

func (w *Worker) reconcilePods(ctx context.Context) error {
	pods, err := w.store.ListPods()
	if err != nil {
		return fmt.Errorf("listing db pods: %w", err)
	}

	// For each pod in DB, ensure it exists in Podman
	// Note: We don't have a ListPods in interface yet, so we just check existence.
	// Efficient logic would be to List all pods from Podman first.
	// But let's stick to simple existence check for now as we added PodExists methods.
	// Limitation: We won't remove orphaned Pods (yet).

	for _, pod := range pods {
		podName := sanitizeName(pod.Name)
		exists, err := w.container.PodExists(ctx, podName)
		if err != nil {
			logger.Error("Failed to check pod existence", "pod", podName, "error", err)
			continue
		}

		if !exists {
			logger.InfoCtx(ctx, "Creating missing pod", "pod", podName)
			// Convert ports map[string]string -> map[uint16]uint16
			ports, err := parsePorts(pod.Ports)
			if err != nil {
				logger.Error("Invalid ports for pod", "pod", podName, "error", err)
				continue
			}

			if _, err := w.container.CreatePod(ctx, podName, ports); err != nil {
				logger.Error("Failed to create pod", "pod", podName, "error", err)
			}
		}
	}
	// TODO: Cleanup orphaned pods (requires List API in ContainerManager)
	return nil
}

func (w *Worker) reconcileApps(ctx context.Context) error {
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

	for i := range containers {
		c := &containers[i]
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
				existingApps[appID] = *c
			}
			managedContainers[c.Name] = true
		}
	}

	desiredContainerNames := make(map[string]bool)

	for i := range apps {
		app := &apps[i]

		// Construct the expected container name
		containerName := sanitizeName(app.Name)
		if containerName == "" {
			containerName = fmt.Sprintf("simplify-%s", app.ID)
		}
		desiredContainerNames[containerName] = true

		// Check if it exists by AppID
		info, exists := existingApps[app.ID]
		if exists {
			desiredContainerNames[info.Name] = true

			// Check if we need to recreate
			needsRecreate := false
			switch {
			case info.Status != "running" && !strings.HasPrefix(info.Status, "Up"):
				needsRecreate = true
			case app.PodID != "":
				// App should be in a Pod.
				// app.PodID is the DB ID. We need to check if the container is in the CORRECT physical pod.
				// Podman container info gives us the PHYSICAL Pod ID.
				// Issues arise if DB Pod ID != Physical Pod ID (e.g. manual recreation).
				// Strict check: info.PodID should match app.PodID.
				// Robust check: Resolve app.PodID -> Pod Name -> Current Physical Pod ID.

				pod, err := w.store.GetPod(app.PodID)
				if err != nil {
					// DB Pod missing? If strict, we might want to fail or detach.
					// For now, let's assume if DB pod is missing, we can't enforce pod constraints.
					logger.Warn("App assigned to non-existent pod in DB", "app", app.Name, "pod_id", app.PodID)
				} else {
					// We have the expected Pod Name.
					// Let's get the CURRENT Physical Pod ID for this name.
					// We can InspectPod or ListPods. Inspect is cheaper if singular.
					physicalPod, err := w.container.InspectPod(ctx, sanitizeName(pod.Name))
					if err != nil {
						// Physical Pod missing?
						// reconcilePods should have created it, but maybe it failed or race condition.
						// If physical pod missing, we definitely need recreation logic to trigger correct deployment path?
						// Actually, if physical pod is missing, deployApp will fail anyway.
						// But here we are checking if CURRENT container is valid.
						// If physical pod missing, current container CANNOT be in it (unless stale info).
						needsRecreate = true
						logger.Info("Physical pod missing", "app", app.Name, "pod_name", pod.Name)
					} else {
						// We have physical ID. Compare with info.PodID.
						// Note: IDs might be short (12 chars) or full (64 chars). compare prefix.
						match := false
						switch {
						case info.PodID == physicalPod.ID:
							match = true
						case len(info.PodID) > len(physicalPod.ID) && strings.HasPrefix(info.PodID, physicalPod.ID):
							match = true
						case len(physicalPod.ID) > len(info.PodID) && strings.HasPrefix(physicalPod.ID, info.PodID):
							match = true
						}

						if !match {
							needsRecreate = true
							logger.Info("Pod mismatch", "app", app.Name, "expected_pod_name", pod.Name, "expected_pod_id", physicalPod.ID, "actual_pod_id", info.PodID)

						}
					}
				}
			case app.PodID == "" && info.PodID != "":
				// App should NOT be in a Pod, but IS
				needsRecreate = true
				logger.Info("Pod mismatch (should be standalone)", "app", app.Name, "actual_pod", info.PodID)
			case app.NetworkID != "":
				// App should be in a Network
				// We need to look up network name from DB ID to check against info.Networks names
				// This requires unnecessary DB lookup inside the loop?
				// Optimization: Pre-fetch networks map?
				// Or... simpler: Just assume if networks list is empty, it's wrong (unless default bridge is implied but usually we want explicit)
				// Actually, we can just check if we need to reconcile networks.
				// For now, let's skip complex name resolution and rely on "if not in correct pod" and "should be in network".
				// If we implement network check properly later.
				// BUT checking if existing network (like 'bridge') matches is good.
				// Note: info.Networks contains names.
				// We can just trigger recreate if we know the network name.
				// Let's defer strict network name check to avoid N+1 DB lookup here,
				// UNLESS we pre-load networks.
				// Given the user issue, the main problem was Pod vs Network switch.
				// The above "app.PodID == "" && info.PodID != """ check COVERS the specific user case!
				// User switched from Pod -> Network. So App.PodID is empty, but Info.PodID is set.
				// So that case is handled.

				// Port check only if standalone
				if !checkPortsMatch(app.Ports, info.Ports) {
					needsRecreate = true
					logger.Info("Ports mismatch", "app", app.Name, "info_ports", info.Ports)
				}
			case app.PodID == "" && !checkPortsMatch(app.Ports, info.Ports):
				// Standalone default bridge
				needsRecreate = true
				logger.Info("Ports mismatch", "app", app.Name, "info_ports", info.Ports)
			}

			if needsRecreate {
				logger.Info("Recreating container", "container", info.Name)
				if err := w.container.Remove(ctx, info.Name, true); err != nil {
					logger.Error("Failed to remove container for update", "container", info.Name, "error", err)
					continue
				}
				// Mark as missing so we fall through to deploy logic
				exists = false
			}
		}

		if !exists {
			// Missing or just removed, deploy
			logger.Info("Deploying missing application", "app", app.Name)
			if err := w.deployApp(ctx, app, containerName); err != nil {
				logger.Error("Failed to deploy app", "app", app.Name, "error", err)
			}
		}
	}

	// Cleanup Orphans
	for name := range managedContainers {
		if !desiredContainerNames[name] {
			logger.Info("Removing orphaned container", "container", name)
			if err := w.container.Remove(ctx, name, true); err != nil {
				logger.Error("Failed to remove orphan", "container", name, "error", err)
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

	// Determine Pod Name if valid
	podName := ""
	if app.PodID != "" {
		pod, err := w.store.GetPod(app.PodID)
		if err != nil {
			logger.WarnCtx(ctx, "App assigned to non-existent pod", "app", app.Name, "pod_id", app.PodID)
			// Decide: Fail or run standalone?
			// Let's run standalone but log warning, OR better: fail to deploy until Pod is ready.
			return fmt.Errorf("pod %s does not exist", app.PodID)
		}
		// Use Pod Name. Podman uses name for associating containers.
		podName = sanitizeName(pod.Name)
	}

	// Determine Network Name if valid
	networkName := ""
	if app.NetworkID != "" {
		net, err := w.store.GetNetwork(app.NetworkID)
		if err != nil {
			logger.WarnCtx(ctx, "App assigned to non-existent network", "app", app.Name, "network_id", app.NetworkID)
			// Decide: Fail or run with default?
			// Let's fail because if user wants custom network, falling back to bridge might be confusing security-wise.
			return fmt.Errorf("network %s does not exist", app.NetworkID)
		}
		networkName = net.Name
	}

	// Call Container Client
	_, err = w.container.Run(ctx, containerName, app.Image, ports, env, labels, podName, networkName)
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
