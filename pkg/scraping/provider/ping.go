package provider

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
)

type Pinger interface {
	Ping(target string) bool
}

type defaultPinger struct{}

func (d *defaultPinger) Ping(target string) bool {
	command := fmt.Sprintf("ping -c 1 -W 1 %s > /dev/null", target)
	_, err := exec.Command("/bin/sh", "-c", command).Output()
	if _, ok := err.(*exec.ExitError); ok {
		return false
	}
	return true
}

type ProviderPing struct {
	pinger     Pinger
	Targets    []string `json:"targets"`
	RetryCount uint     `json:"retry_count" default:"3"`
}

func (p *ProviderPing) pingRetry(target string) bool {
	for range p.RetryCount {
		if p.pinger.Ping(target) {
			return true
		}
	}
	return false
}

func NewProviderPing(params map[string]any) (Provider, error) {
	cfg, err := configmapper.MapOnStruct[ProviderPing](params)
	if err == nil {
		cfg.pinger = &defaultPinger{}
	}
	return &cfg, err
}

func (pingProvider *ProviderPing) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	taskList := UpdateTaskList{}

	for _, target := range pingProvider.Targets {
		taskList = append(taskList,
			func() {
				metric := resultWrapper.Metric("ping_"+target, "ping ["+target+"]")
				if pingProvider.pingRetry(target) {
					metric.PushOK("")
				} else {
					metric.PushFailure("unreachable")
				}
			},
		)
	}
	return taskList
}

func (pingProvider *ProviderPing) MultipleInstanceAllowed() bool {
	return true
}

func (*ProviderPing) Destroy() {
}
func init() {
	RegisterProvider("ping", func(ctx context.Context, cfg Config) (Provider, error) {
		return NewProviderPing(cfg.Params)
	})
}
