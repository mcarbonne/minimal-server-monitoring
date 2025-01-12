package provider

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
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

func listServiceUnits(ctx context.Context, systemdConn *dbus.Conn) []dbus.UnitStatus {
	listOfUnits, _ := systemdConn.ListUnitsByPatternsContext(ctx, []string{}, []string{"*.service"})
	return listOfUnits
}

func (systemdProvider *ProviderSystemd) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	return UpdateTaskList{
		func() {
			listOfUnits := listServiceUnits(ctx, systemdProvider.systemdConn)

			for _, unit := range listOfUnits {
				metric := resultWrapper.Metric("systemd_"+unit.Name, unit.Name+"@systemd")
				if unit.ActiveState == "failed" {
					metric.PushFailure("%v is %v", unit.Name, unit.ActiveState)
				} else {
					metric.PushOK()
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
