package provider

import (
	"fmt"
	"os/exec"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
)

type ProviderPing struct {
	Targets    []string `json:"targets"`
	RetryCount uint     `json:"retry_count" default:"3"`
}

func ping(target string) bool {
	Command := fmt.Sprintf("ping -c 1 -W 1 %s > /dev/null", target)

	_, err := exec.Command("/bin/sh", "-c", Command).Output()

	if _, ok := err.(*exec.ExitError); ok {
		return false
	}
	return true
}

func pingRetry(target string, retryCount uint) bool {
	for range retryCount {
		if ping(target) {
			return true
		}
	}
	return false
}

func NewProviderPing(params map[string]any) Provider {
	cfg, err := utils.MapOnStruct[ProviderPing](params)
	if err != nil {
		logging.Fatal("Unable to load configuration for ping provider: %v", err)
	}
	return &cfg
}

func (pingProvider *ProviderPing) Update(result *ScrapeResultWrapper, storage storage.Storager) {
	for _, target := range pingProvider.Targets {
		if pingRetry(target, pingProvider.RetryCount) {
			result.PushOK("ping_" + target)
		} else {
			result.PushFailure("ping_"+target, "unable to ping %v", target)
		}
	}
}

func (pingProvider *ProviderPing) MultipleInstanceAllowed() bool {
	return true
}

func (*ProviderPing) Destroy() {
}
