package provider

import (
	"context"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils/configmapper"
	"golang.org/x/sys/unix"
)

type ProviderFileSystemUsage struct {
	MountPoints             []string `json:"mountpoints"`
	SpaceRemainingThreshold uint     `json:"threshold_percent" default:"20"`
}

func NewProviderFileSystemUsage(params map[string]any) (Provider, error) {
	cfg, err := configmapper.MapOnStruct[ProviderFileSystemUsage](params)
	return &cfg, err
}

func checkMountPoint(resultWrapper *ScrapeResultWrapper, mountPoint string, threshold uint) {
	var stat unix.Statfs_t

	err := unix.Statfs(mountPoint, &stat)

	metric := resultWrapper.Metric("filesystemusage_"+mountPoint, "mountpoint ["+mountPoint+"]")

	if err != nil {
		metric.PushFailure("unable to get remaining space for %v: %v", mountPoint, err)
	} else {
		remainingSpace := 100 * stat.Bavail / stat.Blocks
		if remainingSpace < uint64(threshold) {
			metric.PushFailure("low space remaining on %v: %v%%", mountPoint, remainingSpace)
		} else {
			metric.PushOK()
		}
	}
}

func (provider *ProviderFileSystemUsage) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	return UpdateTaskList{
		func() {
			for _, mountpoint := range provider.MountPoints {
				checkMountPoint(resultWrapper, mountpoint, provider.SpaceRemainingThreshold)
			}
		},
	}

}

func (*ProviderFileSystemUsage) MultipleInstanceAllowed() bool {
	return true
}

func (*ProviderFileSystemUsage) Destroy() {
}
