package utils

import (
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
)

func SafeClose(closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		logging.Error("failed to close file: %v", err)
	}
}
