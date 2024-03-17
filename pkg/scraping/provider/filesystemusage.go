package provider

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"golang.org/x/sys/unix"
)

type ProviderFileSystemUsage struct {
	MountPoints             []string `json:"mountpoints"`
	SpaceRemainingThreshold uint     `json:"threshold_percent" default:"20"`
}

func NewProviderFileSystemUsage(params map[string]any) Provider {
	cfg, err := utils.MapOnStruct[ProviderFileSystemUsage](params)
	if err != nil {
		logging.Fatal("Unable to load configuration for filesystemusage provider: %v", err)
	}

	if err != nil {
		logging.Fatal("Unable to setup systemd provider: %v", err)
	}

	return &cfg
}

func checkMountPoint(result *ScrapeResult, mountPoint string, threshold uint) {
	var stat unix.Statfs_t

	err := unix.Statfs(mountPoint, &stat)

	if err != nil {
		result.PushFailure("filesystemusage_"+mountPoint, "Unable to get remaining space for %v: %v", mountPoint, err)
	} else {
		remainingSpace := 100 * stat.Bavail / stat.Blocks
		if remainingSpace < uint64(threshold) {
			result.PushFailure("filesystemusage_"+mountPoint, "Low space remaining on %v: %v%%", mountPoint, remainingSpace)
		} else {
			result.PushOK("filesystemusage_" + mountPoint)
		}
	}
}

func (provider *ProviderFileSystemUsage) Update(result *ScrapeResult, storage storage.Storager) {
	for _, mountpoint := range provider.MountPoints {
		checkMountPoint(result, mountpoint, provider.SpaceRemainingThreshold)
	}

}

func (*ProviderFileSystemUsage) MultipleInstanceAllowed() bool {
	return true
}

func (*ProviderFileSystemUsage) Destroy() {
}
