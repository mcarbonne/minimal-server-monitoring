package provider

import (
	"context"
	"testing"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/stats"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
)

type mockFileSystemClient struct {
	StatfsFunc    func(path string, buf *unix.Statfs_t) error
	GetMountsFunc func(filter func(info *mountinfo.Info) (skip, stop bool)) ([]*mountinfo.Info, error)
}

func (m *mockFileSystemClient) Statfs(path string, buf *unix.Statfs_t) error {
	if m.StatfsFunc != nil {
		return m.StatfsFunc(path, buf)
	}
	return nil
}

func (m *mockFileSystemClient) GetMounts(filter func(info *mountinfo.Info) (skip, stop bool)) ([]*mountinfo.Info, error) {
	if m.GetMountsFunc != nil {
		return m.GetMountsFunc(filter)
	}
	return nil, nil
}

func TestFileSystemLowSpace(t *testing.T) {
	mockClient := &mockFileSystemClient{}

	thresh, err := utils.RelativeAbsoluteValueFromString("20%")
	assert.NilError(t, err)
	rateThresh, err := utils.RelativeAbsoluteValueFromString("1g")
	assert.NilError(t, err)

	provider := &ProviderFileSystemUsage{
		client:                  mockClient,
		FSTypeWhitelist:         []string{"ext4"},
		MountPointBlacklist:     []string{},
		MountPointWhitelist:     []string{},
		SpaceRemainingThreshold: thresh,
		RateThreshold:           rateThresh,
		RateThresholdWindow:     customtypes.Duration(5 * time.Minute),
		mountPointStats:         make(map[string]*stats.WindowCollector[uint64]),
	}

	resultChan := make(chan any, 100)
	wrapper := MakeScrapeResultWrapper("fs", resultChan)

	// Mock GetMounts to return root
	mockClient.GetMountsFunc = func(filter func(info *mountinfo.Info) (skip, stop bool)) ([]*mountinfo.Info, error) {
		return []*mountinfo.Info{
			{Mountpoint: "/", FSType: "ext4", Source: "/dev/sda1"},
		}, nil
	}

	// Mock Statfs to return low space (1%)
	mockClient.StatfsFunc = func(path string, buf *unix.Statfs_t) error {
		if path == "/" {
			buf.Bsize = 4096
			buf.Blocks = 1000 // 4MB Total
			buf.Bavail = 10   // 40KB Available (1%)
		}
		return nil
	}

	getAndExecuteTaskList(provider, context.Background(), &wrapper, storage.NewMemoryStorage())

	// Verify
	metricID := "fs_filesystemusage_/"

	val := waitForMetricState(t, resultChan, metricID)

	assert.Equal(t, Unhealthy, val.Status, "Should be unhealthy due to low space")
	assert.Assert(t, val.Description == "low space remaining (1% / 41 kB)")
}
