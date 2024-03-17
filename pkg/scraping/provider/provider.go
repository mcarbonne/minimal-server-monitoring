package provider

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type Provider interface {
	Update(resultWrapper *ScrapeResultWrapper, storage storage.Storager)
	MultipleInstanceAllowed() bool
	Destroy()
}
