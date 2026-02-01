package alert

import (
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
)

type GroupingConfig struct {
	Window customtypes.Duration `json:"window" default:"15s"`
}

type Config struct {
	UnhealthyThreshold   uint                  `json:"unhealthy_threshold" default:"1"` // how many consecutive failed tests to mark metric as unhealthy, 1 means metric is unhealthy on first fail
	HealthyThreshold     uint                  `json:"healthy_threshold" default:"1"`   // how many consecutived pass tests to mark metric as healthy, 1 means metric is healthy on first fail
	FailureReminder      customtypes.Duration  `json:"failure_reminder" default:"2h"`
	DailyReminderTime    customtypes.TimeOfDay `json:"daily_reminder_time" default:"08:00"`
	FailureReminderCount uint                  `json:"failure_reminder_count" default:"3"`
	Grouping             GroupingConfig        `json:"grouping" default:"{}"`
}
