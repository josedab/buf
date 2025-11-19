// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bufremoteplugindocker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Container represents a warm Docker container in the pool.
type Container struct {
	ID        string
	Image     string
	Ready     bool
	CreatedAt time.Time
	LastUsed  time.Time
	client    *client.Client
	mu        sync.Mutex
}

// Execute runs a command in the container.
func (c *Container) Execute(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Ready {
		return errors.New("container is not ready")
	}

	// Create exec instance
	execConfig := containertypes.ExecOptions{
		Cmd:          cmd,
		AttachStdin:  stdin != nil,
		AttachStdout: stdout != nil,
		AttachStderr: stderr != nil,
	}

	execResp, err := c.client.ContainerExecCreate(ctx, c.ID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec
	attachResp, err := c.client.ContainerExecAttach(ctx, execResp.ID, containertypes.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Handle I/O
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Copy stdin
	if stdin != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(attachResp.Conn, stdin)
			if err != nil {
				errChan <- fmt.Errorf("stdin copy error: %w", err)
			}
			// Close stdin to signal EOF
			if closer, ok := attachResp.Conn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
		}()
	}

	// Copy stdout/stderr
	if stdout != nil || stderr != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Docker multiplexes stdout/stderr when not using TTY
			_, err := io.Copy(stdout, attachResp.Reader)
			if err != nil && err != io.EOF {
				errChan <- fmt.Errorf("output copy error: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// Check exec exit code
	inspectResp, err := c.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	if inspectResp.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", inspectResp.ExitCode)
	}

	if len(errs) > 0 {
		return fmt.Errorf("exec errors: %v", errs)
	}

	c.LastUsed = time.Now()
	return nil
}

// ContainerPool manages a pool of warm Docker containers for frequently-used plugins.
type ContainerPool struct {
	client     *client.Client
	containers map[string]*Container // image -> warm container
	mu         sync.RWMutex
	maxSize    int
	logger     *slog.Logger
	stats      PoolStats
}

// PoolStats tracks container pool statistics.
type PoolStats struct {
	TotalContainers int
	ActiveContainers int
	Hits            int64
	Misses          int64
}

// ContainerPoolOption configures a ContainerPool.
type ContainerPoolOption func(*ContainerPool)

// WithPoolMaxSize sets the maximum number of containers in the pool.
func WithPoolMaxSize(size int) ContainerPoolOption {
	return func(p *ContainerPool) {
		p.maxSize = size
	}
}

// WithPoolLogger sets the logger for the pool.
func WithPoolLogger(logger *slog.Logger) ContainerPoolOption {
	return func(p *ContainerPool) {
		p.logger = logger
	}
}

// NewContainerPool creates a new container pool.
func NewContainerPool(dockerClient *client.Client, opts ...ContainerPoolOption) (*ContainerPool, error) {
	if dockerClient == nil {
		return nil, errors.New("docker client is required")
	}

	pool := &ContainerPool{
		client:     dockerClient,
		containers: make(map[string]*Container),
		maxSize:    5, // default
		logger:     slog.Default(),
	}

	for _, opt := range opts {
		opt(pool)
	}

	return pool, nil
}

// Get returns a warm container for the specified image, creating one if necessary.
func (p *ContainerPool) Get(ctx context.Context, image string) (*Container, error) {
	p.mu.RLock()
	if container, ok := p.containers[image]; ok {
		p.mu.RUnlock()
		p.recordHit()
		p.logger.Debug("container pool hit",
			slog.String("image", image),
			slog.String("containerID", container.ID),
		)
		return container, nil
	}
	p.mu.RUnlock()

	p.recordMiss()

	// Create new container
	container, err := p.createWarmContainer(ctx, image)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check again in case another goroutine created it
	if existing, ok := p.containers[image]; ok {
		// Clean up the one we just created
		go p.removeContainer(context.Background(), container.ID)
		return existing, nil
	}

	// Evict if at capacity
	if len(p.containers) >= p.maxSize {
		p.evictOldest()
	}

	p.containers[image] = container
	p.stats.TotalContainers = len(p.containers)

	p.logger.Debug("added container to pool",
		slog.String("image", image),
		slog.String("containerID", container.ID),
		slog.Int("poolSize", len(p.containers)),
	)

	return container, nil
}

// createWarmContainer creates a new warm container for the given image.
func (p *ContainerPool) createWarmContainer(ctx context.Context, image string) (*Container, error) {
	p.logger.Debug("creating warm container", slog.String("image", image))

	// Create container with sleep infinity to keep it alive
	createResp, err := p.client.ContainerCreate(ctx, &containertypes.Config{
		Image: image,
		Cmd:   []string{"sleep", "infinity"},
		Tty:   false,
	}, &containertypes.HostConfig{
		AutoRemove: false,
	}, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := p.client.ContainerStart(ctx, createResp.ID, containertypes.StartOptions{}); err != nil {
		// Clean up the created container
		p.client.ContainerRemove(ctx, createResp.ID, containertypes.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	p.logger.Debug("warm container created and started",
		slog.String("image", image),
		slog.String("containerID", createResp.ID),
	)

	return &Container{
		ID:        createResp.ID,
		Image:     image,
		Ready:     true,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		client:    p.client,
	}, nil
}

// evictOldest removes the least recently used container from the pool.
// Must be called with mu held.
func (p *ContainerPool) evictOldest() {
	var oldestImage string
	var oldestTime time.Time

	for image, container := range p.containers {
		if oldestImage == "" || container.LastUsed.Before(oldestTime) {
			oldestImage = image
			oldestTime = container.LastUsed
		}
	}

	if oldestImage != "" {
		container := p.containers[oldestImage]
		delete(p.containers, oldestImage)
		go p.removeContainer(context.Background(), container.ID)
		p.logger.Debug("evicted container from pool",
			slog.String("image", oldestImage),
			slog.String("containerID", container.ID),
		)
	}
}

// removeContainer stops and removes a container.
func (p *ContainerPool) removeContainer(ctx context.Context, containerID string) {
	// Stop container (with timeout)
	timeout := 10
	if err := p.client.ContainerStop(ctx, containerID, containertypes.StopOptions{Timeout: &timeout}); err != nil {
		p.logger.Warn("failed to stop container",
			slog.Any("error", err),
			slog.String("containerID", containerID),
		)
	}

	// Remove container
	if err := p.client.ContainerRemove(ctx, containerID, containertypes.RemoveOptions{Force: true}); err != nil {
		p.logger.Warn("failed to remove container",
			slog.Any("error", err),
			slog.String("containerID", containerID),
		)
	}
}

// Release returns a container to the pool for reuse.
func (p *ContainerPool) Release(container *Container) {
	container.mu.Lock()
	container.LastUsed = time.Now()
	container.mu.Unlock()
}

// Remove removes a specific container from the pool.
func (p *ContainerPool) Remove(image string) error {
	p.mu.Lock()
	container, ok := p.containers[image]
	if !ok {
		p.mu.Unlock()
		return nil
	}
	delete(p.containers, image)
	p.stats.TotalContainers = len(p.containers)
	p.mu.Unlock()

	p.removeContainer(context.Background(), container.ID)
	return nil
}

// Clear removes all containers from the pool.
func (p *ContainerPool) Clear() error {
	p.mu.Lock()
	containers := make([]*Container, 0, len(p.containers))
	for _, container := range p.containers {
		containers = append(containers, container)
	}
	p.containers = make(map[string]*Container)
	p.stats.TotalContainers = 0
	p.mu.Unlock()

	var errs []error
	for _, container := range containers {
		// Stop container
		timeout := 10
		if err := p.client.ContainerStop(context.Background(), container.ID, containertypes.StopOptions{Timeout: &timeout}); err != nil {
			errs = append(errs, err)
		}
		// Remove container
		if err := p.client.ContainerRemove(context.Background(), container.ID, containertypes.RemoveOptions{Force: true}); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors clearing pool: %v", errs)
	}

	return nil
}

// Stats returns the current pool statistics.
func (p *ContainerPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	stats := p.stats
	stats.TotalContainers = len(p.containers)
	return stats
}

// List returns information about all containers in the pool.
func (p *ContainerPool) List() []Container {
	p.mu.RLock()
	defer p.mu.RUnlock()

	containers := make([]Container, 0, len(p.containers))
	for _, container := range p.containers {
		containers = append(containers, *container)
	}
	return containers
}

// Prune removes containers that haven't been used for the specified duration.
func (p *ContainerPool) Prune(maxIdleTime time.Duration) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-maxIdleTime)
	var removed int

	for image, container := range p.containers {
		if container.LastUsed.Before(cutoff) {
			delete(p.containers, image)
			go p.removeContainer(context.Background(), container.ID)
			removed++
		}
	}

	p.stats.TotalContainers = len(p.containers)
	return removed
}

// Close cleans up all containers in the pool.
func (p *ContainerPool) Close() error {
	return p.Clear()
}

func (p *ContainerPool) recordHit() {
	p.mu.Lock()
	p.stats.Hits++
	p.mu.Unlock()
}

func (p *ContainerPool) recordMiss() {
	p.mu.Lock()
	p.stats.Misses++
	p.mu.Unlock()
}
