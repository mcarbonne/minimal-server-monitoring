package provider

import (
	"context"
	"slices"
	"strings"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils/configmapper"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"
)

type ProviderFileSystemUsage struct {
	MountPrefix             string   `json:"mountprefix" default:""` // Host root filesytem when running inside a container
	FSTypeWhitelist         []string `json:"fstypes" default:"[ext4, btrfs]"`
	MountPointBlacklist     []string `json:"mountpoint_blacklist" default:"[]"`
	MountPointWhitelist     []string `json:"mountpoint_whitelist" default:"[]"`
	SpaceRemainingThreshold uint     `json:"threshold_percent" default:"20"`
}

func NewProviderFileSystemUsage(params map[string]any) (Provider, error) {
	cfg, err := configmapper.MapOnStruct[ProviderFileSystemUsage](params)
	return &cfg, err
}

func (provider *ProviderFileSystemUsage) checkMountPoint(resultWrapper *ScrapeResultWrapper, mountPoint string) {
	var stat unix.Statfs_t

	err := unix.Statfs(mountPoint, &stat)

	prettyMountpoint := strings.TrimPrefix(mountPoint, provider.MountPrefix)
	if !strings.HasPrefix(prettyMountpoint, "/") {
		prettyMountpoint = "/" + prettyMountpoint
	}

	metric := resultWrapper.Metric("filesystemusage_"+prettyMountpoint, "mountpoint "+prettyMountpoint)

	if err != nil {
		metric.PushFailure("unable to get remaining space: %v", err)
	} else {
		remainingSpace := 100 * stat.Bavail / stat.Blocks
		if remainingSpace < uint64(provider.SpaceRemainingThreshold) {
			metric.PushFailure("low space remaining (%v%%)", remainingSpace)
		} else {
			metric.PushOK()
		}
	}
}

func (provider *ProviderFileSystemUsage) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	mountpoints := []string{}

	if len(provider.MountPointWhitelist) > 0 {
		mountpoints = provider.FSTypeWhitelist
	} else {
		allMountPoints, err := mountinfo.GetMounts(func(info *mountinfo.Info) (skip, stop bool) {
			return !slices.Contains(provider.FSTypeWhitelist, info.FSType) || slices.Contains(provider.MountPointBlacklist, info.Mountpoint), false
		})

		filteredMountPointsBySource := map[string]*mountinfo.Info{}

		for _, info := range allMountPoints {
			v, ok := filteredMountPointsBySource[info.Source]
			if ok {
				if len(info.Source) < len(v.Source) {
					filteredMountPointsBySource[info.Source] = info
				}
			} else {
				filteredMountPointsBySource[info.Source] = info
			}
		}

		for _, info := range filteredMountPointsBySource {
			mountpoints = append(mountpoints, info.Mountpoint)
		}

		if err != nil {
			logging.Fatal("Unable to list mountpoints: %v", err)
		}
	}
	for _, v := range mountpoints {
		logging.Info("Monitoring available disk space on %v", v)
	}

	return UpdateTaskList{
		func() {
			for _, mountpoint := range mountpoints {
				provider.checkMountPoint(resultWrapper, mountpoint)
			}
		},
	}

}

func (*ProviderFileSystemUsage) MultipleInstanceAllowed() bool {
	return true
}

func (*ProviderFileSystemUsage) Destroy() {
}
