package provider

import (
	"context"
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/containerapi"
	"gotest.tools/v3/assert"
)

// Mock Client Implementation
type mockContainerClient struct {
	ListFunc    func(ctx context.Context) ([]containerapi.Container, error)
	InspectFunc func(ctx context.Context, containerId string) (containerapi.ContainerInspect, error)
}

func (m *mockContainerClient) ContainerList(ctx context.Context) ([]containerapi.Container, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *mockContainerClient) ContainerInspect(ctx context.Context, containerId string) (containerapi.ContainerInspect, error) {
	if m.InspectFunc != nil {
		return m.InspectFunc(ctx, containerId)
	}
	return containerapi.ContainerInspect{}, nil
}

func TestContainerDisappearance(t *testing.T) {
	// Setup
	mockClient := &mockContainerClient{}
	provider := &ProviderContainer{
		client:                mockClient,
		containerRestartCount: make(map[string]int),
		containerState:        make(map[string]string),
	}

	resultChan := make(chan any, 10)
	wrapper := MakeScrapeResultWrapper("test", resultChan) // Prefix "test"
	memStorage := storage.NewMemoryStorage()

	// 1. Initial State: Container Running
	mockClient.ListFunc = func(ctx context.Context) ([]containerapi.Container, error) {
		return []containerapi.Container{
			{
				ID:      "container123",
				Names:   []string{"my-app"},
				Image:   "my-image:latest",
				ImageID: "sha256:1111",
				State:   "running",
				Status:  "Up 2 hours",
			},
		}, nil
	}
	mockClient.InspectFunc = func(ctx context.Context, id string) (containerapi.ContainerInspect, error) {
		return containerapi.ContainerInspect{RestartCount: 0}, nil
	}

	// Run Scrape
	taskList := provider.GetUpdateTaskList(context.Background(), &wrapper, memStorage)
	assert.Equal(t, 1, len(taskList))
	taskList[0]()

	// Verify Metrics
	stateMetric := waitForMetricState(t, resultChan, "test_container_state_container123")

	assert.Equal(t, Healthy, stateMetric.Status, "Container should be healthy")
	assert.Equal(t, "my-app@container (my-image:latest) state", stateMetric.Name)
	assert.Equal(t, "", stateMetric.Description)

	// 2. State Change: Container Disappears
	mockClient.ListFunc = func(ctx context.Context) ([]containerapi.Container, error) {
		return []containerapi.Container{}, nil // Empty list
	}

	taskList[0]() // Run Scrape again

	// Verify "Disappeared" Metric
	stateMetric = waitForMetricState(t, resultChan, "test_container_state_container123")

	assert.Equal(t, Removed, stateMetric.Status, "Disappeared container should be Removed")
	assert.Equal(t, "container removed", stateMetric.Description)
}

func TestContainerImageUpdate(t *testing.T) {
	mockClient := &mockContainerClient{}
	provider := &ProviderContainer{
		client:                mockClient,
		containerRestartCount: make(map[string]int),
		containerState:        make(map[string]string),
	}

	resultChan := make(chan any, 10)
	wrapper := MakeScrapeResultWrapper("test", resultChan)
	memStorage := storage.NewMemoryStorage()

	// 1. Initial Scrape: Image A
	mockClient.ListFunc = func(ctx context.Context) ([]containerapi.Container, error) {
		return []containerapi.Container{
			{
				ID:      "container123",
				Names:   []string{"my-app"},
				Image:   "my-image:latest",
				ImageID: "sha256:old_hash",
				State:   "running",
			},
		}, nil
	}

	taskList := provider.GetUpdateTaskList(context.Background(), &wrapper, memStorage)
	taskList[0]()

	// Consume initial metrics (we don't expect an update message yet)
	drainChannel(resultChan)

	// 2. Update: Image B
	mockClient.ListFunc = func(ctx context.Context) ([]containerapi.Container, error) {
		return []containerapi.Container{
			{
				ID:      "container123",
				Names:   []string{"my-app"},
				Image:   "my-image:latest",
				ImageID: "sha256:new_hash", // Changed
				State:   "running",
			},
		}, nil
	}

	taskList[0]()

	// Verify "image was updated" message
	msg := waitForMessage(t, resultChan, "test_container_image_update_container123")
	assert.Equal(t, "image was updated", msg.Description)
}
