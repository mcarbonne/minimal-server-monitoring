package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"gotest.tools/v3/assert"
)

type mockSystemdClient struct {
	ListUnitsFunc func(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error)
	CloseFunc     func()
}

func (m *mockSystemdClient) ListUnitsByPatternsContext(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error) {
	if m.ListUnitsFunc != nil {
		return m.ListUnitsFunc(ctx, states, patterns)
	}
	return nil, nil
}

func (m *mockSystemdClient) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func TestSystemdPodmanNaming(t *testing.T) {
	mockClient := &mockSystemdClient{}
	factory := func(ctx context.Context) (SystemdClient, error) { return mockClient, nil }
	provider := &ProviderSystemd{client: mockClient, clientFactory: factory, knownUnitList: []dbus.UnitStatus{}}

	resultChan := make(chan any, 100)
	wrapper := MakeScrapeResultWrapper("systemd", resultChan)

	// Container ID: 64 hex chars
	containerID := "1111111111111111111111111111111111111111111111111111111111111111"

	mockClient.ListUnitsFunc = func(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error) {
		return []dbus.UnitStatus{
			{
				Name:        containerID + ".service",
				Description: "/usr/bin/podman healthcheck run " + containerID,
				ActiveState: "active",
			},
		}, nil
	}

	getAndExecuteTaskList(provider, context.Background(), &wrapper, storage.NewMemoryStorage())

	metric := waitForMetricState(t, resultChan, "systemd_systemd_"+containerID+".service")
	// Name should be "container <shortID> healthcheck@systemd"
	assert.Equal(t, "container 111111111111 healthcheck@systemd", metric.Name)

	provider.Destroy()
}

func TestSystemdRetry(t *testing.T) {
	mockClient := &mockSystemdClient{}

	// Track factory calls
	factoryCalls := 0
	factory := func(ctx context.Context) (SystemdClient, error) {
		factoryCalls++
		return mockClient, nil
	}

	provider := &ProviderSystemd{client: mockClient, clientFactory: factory, knownUnitList: []dbus.UnitStatus{}}

	resultChan := make(chan any, 100)
	wrapper := MakeScrapeResultWrapper("systemd", resultChan)

	// Fail 2 times, then succeed
	callCount := 0
	mockClient.ListUnitsFunc = func(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error) {
		callCount++
		if callCount <= 2 {
			return nil, errors.New("dbus error")
		}
		return []dbus.UnitStatus{{Name: "ok.service"}}, nil
	}

	getAndExecuteTaskList(provider, context.Background(), &wrapper, storage.NewMemoryStorage())

	// Verification
	metric := waitForMetricState(t, resultChan, "systemd_systemd_ok.service")

	assert.Equal(t, 3, callCount)
	assert.Equal(t, 2, factoryCalls)

	assert.Equal(t, true, metric.IsHealthy)
}

func TestSystemdDisappearance(t *testing.T) {
	mockClient := &mockSystemdClient{}

	// Dummy factory since we inject client directly
	factory := func(ctx context.Context) (SystemdClient, error) {
		return mockClient, nil
	}

	provider := &ProviderSystemd{
		client:        mockClient,
		clientFactory: factory,
		knownUnitList: []dbus.UnitStatus{},
	}

	resultChan := make(chan any, 100)
	wrapper := MakeScrapeResultWrapper("systemd", resultChan) // Prefix "systemd"

	// 1. Initial State: Service Present
	mockClient.ListUnitsFunc = func(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error) {
		return []dbus.UnitStatus{
			{Name: "test.service", ActiveState: "active"},
		}, nil
	}

	getAndExecuteTaskList(provider, context.Background(), &wrapper, storage.NewMemoryStorage())

	metric := waitForMetricState(t, resultChan, "systemd_systemd_test.service")

	assert.Equal(t, true, metric.IsHealthy)
	assert.Equal(t, "test.service@systemd", metric.Name)

	// 2. Service Disappears
	mockClient.ListUnitsFunc = func(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error) {
		return []dbus.UnitStatus{}, nil
	}

	getAndExecuteTaskList(provider, context.Background(), &wrapper, storage.NewMemoryStorage())

	metric = waitForMetricState(t, resultChan, "systemd_systemd_test.service")

	assert.Equal(t, true, metric.IsHealthy)
	assert.Equal(t, "service removed", metric.Description)
}
