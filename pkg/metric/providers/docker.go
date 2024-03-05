package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/metric"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type DockerMetric struct {
	client *client.Client

	containerRestartCount map[string]int
	containerState        map[string]string
}

func NewDockerProvider() metric.Provider {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logging.Fatal("Unable to connect to docker %v", err)
	}
	return &DockerMetric{
		client:                cli,
		containerRestartCount: make(map[string]int),
		containerState:        make(map[string]string),
	}
}

func containerDescription(ctr types.Container) string {
	return fmt.Sprintf("Container [%v] (%v)", strings.Join(ctr.Names, ", "), ctr.Image)
}

func (dc *DockerMetric) updateStateMetric(list *metric.ProviderResultList, ctr types.Container) {
	dc.containerState[ctr.ID] = ctr.State
	if ctr.State != "running" {
		list.Push(metric.TypeFailure, "container_state", "%v isn't running (%v)", containerDescription(ctr), ctr.State)
	}
}

func (dc *DockerMetric) updateImageMetric(list *metric.ProviderResultList, storage storage.Storager, ctr types.Container) {
	imageKey := fmt.Sprintf("image_id/%s", ctr.ID)

	lastKnownImage, exists := storage.Get(imageKey)
	changed := storage.Set(imageKey, ctr.ImageID)
	if changed && exists {
		list.Push(metric.TypeInfo, "container_image_update", "%v was updated from %v to %v", containerDescription(ctr), lastKnownImage, ctr.ImageID)
	}
}

func (dc *DockerMetric) updateRestartCountMetric(list *metric.ProviderResultList, ctr types.Container, inspect types.ContainerJSON) {

	lastRestartCount := dc.containerRestartCount[ctr.ID]
	dc.containerRestartCount[ctr.ID] = inspect.RestartCount
	if lastRestartCount != inspect.RestartCount && inspect.RestartCount > 0 {
		list.Push(metric.TypeFailure, "container_restarted", "%v is restarting (%v, %v)", containerDescription(ctr), inspect.RestartCount, ctr.Status)
	}
}

func (dc *DockerMetric) Update(storage storage.Storager) metric.ProviderResultList {
	resultList := metric.MakeProviderResultList()
	containers, err := dc.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		resultList.Push(metric.TypeFailure, "general", "failed to list containers")
		return resultList
	}

	for _, ctr := range containers {
		dc.updateStateMetric(&resultList, ctr)
		dc.updateImageMetric(&resultList, storage, ctr)

		inspect, err := dc.client.ContainerInspect(context.Background(), ctr.ID)
		if err == nil {
			dc.updateRestartCountMetric(&resultList, ctr, inspect)
		} else {
			resultList.Push(metric.TypeFailure, "general", "unable to inspect container: %v", err)
		}
	}

	return resultList
}
