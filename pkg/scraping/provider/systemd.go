package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper"
)

type ProviderSystemd struct {
	systemdConn *dbus.Conn
}

func NewProviderSystemd(ctx context.Context, params map[string]any) (Provider, error) {
	cfg, err := configmapper.MapOnStruct[ProviderSystemd](params)
	if err == nil {
		cfg.systemdConn, err = dbus.NewSystemdConnectionContext(ctx)
	}
	return &cfg, err
}

func listServiceUnits(ctx context.Context, systemdConn *dbus.Conn) ([]dbus.UnitStatus, error) {
	return systemdConn.ListUnitsByPatternsContext(ctx, []string{}, []string{"*.service"})
}

func extractPodmanHealthCheckPrettyName(unit dbus.UnitStatus) *string {
	podmanHealthCheckServiceRegex := `^[\da-f]{64}\.service$`
	unitNameMatched, err := regexp.MatchString(podmanHealthCheckServiceRegex, unit.Name)
	if err == nil && unitNameMatched {
		containerId := unit.Name[:64]
		expectedDescription := fmt.Sprintf("/usr/bin/podman healthcheck run %v", containerId)
		if unit.Description == expectedDescription {
			return utils.Ptr(fmt.Sprintf("container %v healthcheck", containerId[:12]))
		}
	}
	return nil
}

func getServicePrettyName(unit dbus.UnitStatus) string {
	if prettyName := extractPodmanHealthCheckPrettyName(unit); prettyName != nil {
		return *prettyName
	} else {
		return unit.Name
	}
}

func (systemdProvider *ProviderSystemd) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	return UpdateTaskList{
		func() {
			metricListServices := resultWrapper.Metric("list_services", "list services")
			listOfUnits, err := listServiceUnits(ctx, systemdProvider.systemdConn)
			if err != nil {
				metricListServices.PushFailure("failed to list services: %v", err)
				return
			} else {
				metricListServices.PushOK("")
			}

			for _, unit := range listOfUnits {
				prettyName := getServicePrettyName(unit)
				metric := resultWrapper.Metric("systemd_"+unit.Name, prettyName+"@systemd")
				if unit.ActiveState == "failed" {
					metric.PushFailure("")
				} else {
					metric.PushOK("")
				}
			}
		},
	}
}

func (*ProviderSystemd) MultipleInstanceAllowed() bool {
	return false
}

func (systemdProvider *ProviderSystemd) Destroy() {
	systemdProvider.systemdConn.Close()
}
