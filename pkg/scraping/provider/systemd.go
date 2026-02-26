package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
)

const NB_RETRIES = 3

type SystemdClient interface {
	ListUnitsByPatternsContext(ctx context.Context, states []string, patterns []string) ([]dbus.UnitStatus, error)
	Close()
}

type ProviderSystemd struct {
	client        SystemdClient
	clientFactory func(context.Context) (SystemdClient, error)
	knownUnitList []dbus.UnitStatus
}

func (provider *ProviderSystemd) reset(ctx context.Context) error {
	var err error
	provider.client.Close()
	provider.client, err = provider.clientFactory(ctx)
	return err
}

func defaultSystemdFactory(ctx context.Context) (SystemdClient, error) {
	return dbus.NewSystemdConnectionContext(ctx)
}

func NewProviderSystemd(ctx context.Context, params map[string]any) (Provider, error) {
	cfg, err := configmapper.MapOnStruct[ProviderSystemd](params)
	if err == nil {
		cfg.clientFactory = defaultSystemdFactory
		cfg.client, err = cfg.clientFactory(ctx)
	}
	return &cfg, err
}

func (provider *ProviderSystemd) listServiceUnits(ctx context.Context) ([]dbus.UnitStatus, error) {
	result, err := provider.client.ListUnitsByPatternsContext(ctx, []string{}, []string{"*.service"})
	for i := 0; i < NB_RETRIES && err != nil; i++ {
		logging.Warning("resetting dbus connection (%v)", err)
		err = provider.reset(ctx)
		if err != nil {
			logging.Warning("reset failed: %v", err)
		}

		result, err = provider.client.ListUnitsByPatternsContext(ctx, []string{}, []string{"*.service"})
	}
	return result, err
}

func extractPodmanHealthCheckPrettyName(unit dbus.UnitStatus) (string, bool) {
	podmanHealthCheckServiceRegex := `^[\da-f]{64}\.service$`
	unitNameMatched, err := regexp.MatchString(podmanHealthCheckServiceRegex, unit.Name)
	if err == nil && unitNameMatched {
		containerId := unit.Name[:64]
		expectedDescription := fmt.Sprintf("/usr/bin/podman healthcheck run %v", containerId)
		if unit.Description == expectedDescription {
			return fmt.Sprintf("container %v healthcheck", containerId[:12]), true
		}
	}
	return "", false
}

func getServicePrettyName(unit dbus.UnitStatus) string {
	if prettyName, isPodmanHealthCheck := extractPodmanHealthCheckPrettyName(unit); isPodmanHealthCheck {
		return prettyName
	} else {
		return unit.Name
	}
}

func (systemdProvider *ProviderSystemd) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	return UpdateTaskList{
		func() {
			metricListServices := resultWrapper.Metric("list_services", "list services")
			listOfUnits, err := systemdProvider.listServiceUnits(ctx)
			if err != nil {
				metricListServices.PushFailure("failed to list services: %v", err)
				return
			}
			metricListServices.PushOK("")

			// For O(1) lookup
			currentUnitsMap := make(map[string]struct{}, len(listOfUnits))

			for _, unit := range listOfUnits {
				currentUnitsMap[unit.Name] = struct{}{}
				prettyName := getServicePrettyName(unit)
				metric := resultWrapper.Metric("systemd_"+unit.Name, prettyName+"@systemd")
				if unit.ActiveState == "failed" {
					metric.PushFailure("")
				} else {
					metric.PushOK("")
				}
			}

			for _, knownUnit := range systemdProvider.knownUnitList {
				if _, exists := currentUnitsMap[knownUnit.Name]; !exists {
					prettyName := getServicePrettyName(knownUnit)
					metric := resultWrapper.Metric("systemd_"+knownUnit.Name, prettyName+"@systemd")
					metric.PushOK("service removed")
				}
			}
			systemdProvider.knownUnitList = listOfUnits
		},
	}
}

func (*ProviderSystemd) MultipleInstanceAllowed() bool {
	return false
}

func (systemdProvider *ProviderSystemd) Destroy() {
	systemdProvider.client.Close()
}

func init() {
	RegisterProvider("systemd", func(ctx context.Context, cfg Config) (Provider, error) {
		return NewProviderSystemd(ctx, cfg.Params)
	})
}
