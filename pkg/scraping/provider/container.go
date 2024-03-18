package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type ProviderContainer struct {
	client *client.Client

	containerRestartCount map[string]int
	containerState        map[string]string
}

func NewProviderContainer() (Provider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &ProviderContainer{
		client:                cli,
		containerRestartCount: make(map[string]int),
		containerState:        make(map[string]string),
	}, nil
}

func containerPrettyName(ctr types.Container) string {
	return strings.Join(ctr.Names, ", ") + "@container (" + ctr.Image + ")"
}

func (containerProvider *ProviderContainer) updateStateMetric(resultWrapper *ScrapeResultWrapper, ctr types.Container) {
	metric := resultWrapper.Metric("container_state_"+ctr.ID, containerPrettyName(ctr)+" state")
	containerProvider.containerState[ctr.ID] = ctr.State
	if ctr.State != "running" {
		metric.PushFailure("container isn't running (%v)", ctr.State)
	} else {
		metric.PushOK()
	}
}

func (containerProvider *ProviderContainer) updateImageMetric(resultWrapper *ScrapeResultWrapper, storage storage.Storager, ctr types.Container) {
	imageKey := fmt.Sprintf("container/%v/image_id", ctr.Names)
	metric := resultWrapper.Metric("container_image_update_"+ctr.ID, containerPrettyName(ctr)+" image update")

	_, exists := storage.Get(imageKey)
	changed := storage.Set(imageKey, ctr.ImageID)
	if changed && exists {
		metric.PushMessage("image was updated")
	}
}

func (containerProvider *ProviderContainer) updateRestartCountMetric(resultWrapper *ScrapeResultWrapper, ctr types.Container, inspect types.ContainerJSON) {
	metric := resultWrapper.Metric("container_restarted_"+ctr.ID, containerPrettyName(ctr)+" restart")

	lastRestartCount := containerProvider.containerRestartCount[ctr.ID]
	containerProvider.containerRestartCount[ctr.ID] = inspect.RestartCount
	if lastRestartCount != inspect.RestartCount && inspect.RestartCount > 0 {
		metric.PushFailure("container is restarting (%v, %v)", inspect.RestartCount, ctr.Status)
	} else {
		metric.PushOK()
	}
}

func (containerProvider *ProviderContainer) GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList {
	return UpdateTaskList{
		func() {
			containers, err := containerProvider.client.ContainerList(ctx, container.ListOptions{})

			metricListContainer := resultWrapper.Metric("general_list_container", "container provider")
			if err != nil {
				metricListContainer.PushFailure("failed to list containers")
				return
			} else {
				metricListContainer.PushOK()
			}

			var inspectErrorList []error
			for _, ctr := range containers {
				containerProvider.updateStateMetric(resultWrapper, ctr)
				containerProvider.updateImageMetric(resultWrapper, storage, ctr)

				inspect, err := containerProvider.client.ContainerInspect(ctx, ctr.ID)
				if err == nil {
					containerProvider.updateRestartCountMetric(resultWrapper, ctr, inspect)
				} else {
					inspectErrorList = append(inspectErrorList, err)
				}
			}

			metricInspectContainer := resultWrapper.Metric("general_inspect_container", "container provider")
			if len(inspectErrorList) > 0 {
				metricInspectContainer.PushFailure("unable to inspect containers: %v", inspectErrorList)
			} else {
				metricInspectContainer.PushOK()
			}
		},
	}
}

func (containerProvider *ProviderContainer) MultipleInstanceAllowed() bool {
	return false
}

func (*ProviderContainer) Destroy() {
}
