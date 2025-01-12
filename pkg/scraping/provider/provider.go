package provider

import (
	"context"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
)

type UpdateTaskList []func()

type Provider interface {
	// return the list of tasks to be performed to update the provider metrics.
	// This function is called only once at startup.
	// Tasks may be executed in parallel.
	GetUpdateTaskList(ctx context.Context, resultWrapper *ScrapeResultWrapper, storage storage.Storager) UpdateTaskList
	MultipleInstanceAllowed() bool
	Destroy()
}
