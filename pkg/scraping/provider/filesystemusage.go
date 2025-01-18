package provider

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/stats"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"
)

type ProviderFileSystemUsage struct {
	MountPrefix             string                      `json:"mountprefix" default:""` // Host root filesytem when running inside a container
	FSTypeWhitelist         []string                    `json:"fstypes" default:"[ext4, btrfs]"`
	MountPointBlacklist     []string                    `json:"mountpoint_blacklist" default:"[]"`
	MountPointWhitelist     []string                    `json:"mountpoint_whitelist" default:"[]"`
	SpaceRemainingThreshold utils.RelativeAbsoluteValue `json:"threshold" default:"20%" custom:"relative_absolute_value"`
	RateThreshold           utils.RelativeAbsoluteValue `json:"rate_threshold" default:"1g" custom:"relative_absolute_value"`
	RateThresholdWindow     time.Duration               `json:"rate_threshold_window" default:"5m"`

	mountPointStats map[string]*stats.WindowCollector[uint64]
}

func NewProviderFileSystemUsage(params map[string]any, scrapeInterval time.Duration) (Provider, error) {
	mapperCtx := configmapper.MakeContext()
	err := mapperCtx.RegisterCustomParser("relative_absolute_value", func(s string) (reflect.Value, error) {
		value, err := utils.RelativeAbsoluteValueFromString(s)
		if err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(value), nil
		}
	})
	if err != nil {
		return nil, err
	}

	cfg, err := configmapper.MapOnStructWithContext[ProviderFileSystemUsage](&mapperCtx, params)
	if err != nil {
		return nil, err
	}
	cfg.mountPointStats = make(map[string]*stats.WindowCollector[uint64])
	if cfg.RateThresholdWindow < scrapeInterval {
		return nil, fmt.Errorf("rate_threshold_window must be greater than or equal to scrape_interval (%v < %v)", cfg.RateThresholdWindow, scrapeInterval)
	}
	return &cfg, err
}

func (provider *ProviderFileSystemUsage) updateSpaceIncreaseStats(metric MetricWrapper, mountPoint string, remainingSpace, totalSpace uint64) {
	_, ok := provider.mountPointStats[mountPoint]
	if !ok {
		v := stats.MakeWindowCollector[uint64](provider.RateThresholdWindow)
		provider.mountPointStats[mountPoint] = &v
	}
	mpStats := provider.mountPointStats[mountPoint]
	mpStats.AddNew(remainingSpace)
	if mpStats.Count() >= 2 {
		first := mpStats.First()
		last := mpStats.Last()
		deltaAvailable := max(last.Data, first.Data) - min(last.Data, first.Data)
		rate := float64(deltaAvailable) / (last.Timestamp.Sub(first.Timestamp).Seconds())
		threshold := float64(provider.RateThreshold.GetValue(totalSpace)) / provider.RateThresholdWindow.Seconds()
		if rate >= threshold {
			if last.Data > first.Data {
				metric.PushFailure("available space increased rapidly by %v in %v to reach %v",
					humanize.Bytes(deltaAvailable),
					last.Timestamp.Sub(first.Timestamp).Round(time.Second),
					humanize.Bytes(last.Data))
			} else {
				metric.PushFailure("available space decreased rapidly by %v in %v to reach %v",
					humanize.Bytes(deltaAvailable),
					last.Timestamp.Sub(first.Timestamp).Round(time.Second),
					humanize.Bytes(last.Data))
			}
		} else {
			metric.PushOK("")
		}
	}
}

func (provider *ProviderFileSystemUsage) checkMountPoint(resultWrapper *ScrapeResultWrapper, mountPoint string) {
	var stat unix.Statfs_t

	err := unix.Statfs(mountPoint, &stat)

	prettyMountpoint := strings.TrimPrefix(mountPoint, provider.MountPrefix)
	if !strings.HasPrefix(prettyMountpoint, "/") {
		prettyMountpoint = "/" + prettyMountpoint
	}

	metric := resultWrapper.Metric("filesystemusage_"+prettyMountpoint, "mountpoint "+prettyMountpoint)
	metricInc := resultWrapper.Metric("filesystemusage_"+prettyMountpoint+"_rate", "mountpoint "+prettyMountpoint)

	if err != nil {
		metric.PushFailure("unable to get remaining space: %v", err)
	} else {
		remainingSpace := stat.Bavail * uint64(stat.Bsize)
		totalSpace := stat.Blocks * uint64(stat.Bsize)
		provider.updateSpaceIncreaseStats(metricInc, mountPoint, remainingSpace, totalSpace)

		if remainingSpace < provider.SpaceRemainingThreshold.GetValue(totalSpace) {
			metric.PushFailure("low space remaining (%v%% / %v)", 100*remainingSpace/totalSpace, humanize.Bytes(remainingSpace))
		} else {
			metric.PushOK("")
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
