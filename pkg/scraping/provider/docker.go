package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type ProviderDocker struct {
	client *client.Client

	containerRestartCount map[string]int
	containerState        map[string]string
}

func NewProviderDocker() Provider {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logging.Fatal("Unable to connect to docker %v", err)
	}
	return &ProviderDocker{
		client:                cli,
		containerRestartCount: make(map[string]int),
		containerState:        make(map[string]string),
	}
}

func containerDescription(ctr types.Container) string {
	return fmt.Sprintf("Container [%v] (%v)", strings.Join(ctr.Names, ", "), ctr.Image)
}

func (dockerProvider *ProviderDocker) updateStateMetric(list *ScrapeResult, ctr types.Container) {
	metric_id := fmt.Sprintf("container_state_%v", ctr.ID)
	dockerProvider.containerState[ctr.ID] = ctr.State
	if ctr.State != "running" {
		list.PushFailure(metric_id, "%v isn't running (%v)", containerDescription(ctr), ctr.State)
	} else {
		list.PushOK(metric_id)
	}
}

func (dockerProvider *ProviderDocker) updateImageMetric(list *ScrapeResult, storage storage.Storager, ctr types.Container) {
	imageKey := fmt.Sprintf("image_id/%s", ctr.ID)
	metric_id := fmt.Sprintf("container_image_update_%v", ctr.ID)

	lastKnownImage, exists := storage.Get(imageKey)
	changed := storage.Set(imageKey, ctr.ImageID)
	if changed && exists {
		list.PushMessage(metric_id, "%v was updated from %v to %v", containerDescription(ctr), lastKnownImage, ctr.ImageID)
	}
}

func (dockerProvider *ProviderDocker) updateRestartCountMetric(list *ScrapeResult, ctr types.Container, inspect types.ContainerJSON) {
	metric_id := fmt.Sprintf("container_restarted_%v", ctr.ID)

	lastRestartCount := dockerProvider.containerRestartCount[ctr.ID]
	dockerProvider.containerRestartCount[ctr.ID] = inspect.RestartCount
	if lastRestartCount != inspect.RestartCount && inspect.RestartCount > 0 {
		list.PushFailure(metric_id, "%v is restarting (%v, %v)", containerDescription(ctr), inspect.RestartCount, ctr.Status)
	} else {
		list.PushOK(metric_id)
	}
}

func (dockerProvider *ProviderDocker) Update(result *ScrapeResult, storage storage.Storager) {
	containers, err := dockerProvider.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		result.PushFailure("general_list_container", "failed to list containers")
		return
	} else {
		result.PushOK("general_list_container")
	}

	var inspectErrorList []error
	for _, ctr := range containers {
		dockerProvider.updateStateMetric(result, ctr)
		dockerProvider.updateImageMetric(result, storage, ctr)

		inspect, err := dockerProvider.client.ContainerInspect(context.Background(), ctr.ID)
		if err == nil {
			dockerProvider.updateRestartCountMetric(result, ctr, inspect)
		} else {
			inspectErrorList = append(inspectErrorList, err)
		}
	}

	if len(inspectErrorList) > 0 {
		result.PushFailure("general_inspect_container", "unable to inspect containers: %v", inspectErrorList)
	} else {
		result.PushOK("general_inspect_container")
	}
}

func (dockerProvider *ProviderDocker) MultipleInstanceAllowed() bool {
	return false
}
