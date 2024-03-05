package metric

import (
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

type Provider interface {
	Update(storage storage.Storager) ProviderResultList
}
