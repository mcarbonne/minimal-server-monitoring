package provider

import (
	"context"
	"reflect"
	"slices"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils/configmapper"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"
)

type ProviderFileSystemUsage struct {
	MountPrefix             string                      `json:"mountprefix" default:""` // Host root filesytem when running inside a container
	FSTypeWhitelist         []string                    `json:"fstypes" default:"[ext4, btrfs]"`
	MountPointBlacklist     []string                    `json:"mountpoint_blacklist" default:"[]"`
	MountPointWhitelist     []string                    `json:"mountpoint_whitelist" default:"[]"`
	SpaceRemainingThreshold utils.RelativeAbsoluteValue `json:"threshold" default:"20%" custom:"relative_absolute_value"`
}

func NewProviderFileSystemUsage(params map[string]any) (Provider, error) {
	mapperCtx := configmapper.MakeContext()
	mapperCtx.RegisterCustomParser("relative_absolute_value", func(s string) (reflect.Value, error) {
		value, err := utils.RelativeAbsoluteValueFromString(s)
		if err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(value), nil
		}
	})
	cfg, err := configmapper.MapOnStructWithContext[ProviderFileSystemUsage](&mapperCtx, params)
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
		remainingSpace := stat.Bavail * uint64(stat.Bsize)
		totalSpace := stat.Blocks * uint64(stat.Bsize)
		if remainingSpace < provider.SpaceRemainingThreshold.GetValue(totalSpace) {
			metric.PushFailure("low space remaining (%v%% / %v)", 100*remainingSpace/totalSpace, humanize.Bytes(remainingSpace))
		} else {
			metric.PushOK()
		}
	}
}

func (provider *ProviderFileSystemUsage) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	mountpoints := []string{}

	if len(provider.MountPointWhitelist) > 0 {
		mountpoints = provider.MountPointWhitelist
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
