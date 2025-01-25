package alert

import "time"

type GroupingConfig struct {
	Window time.Duration `json:"window" default:"15s"`
}

type Config struct {
	UnhealthyThreshold uint           `json:"unhealthy_threshold" default:"1"` // how many consecutive failed tests to mark metric as unhealthy, 1 means metric is unhealthy on first fail
	HealthyThreshold   uint           `json:"healthy_threshold" default:"1"`   // how many consecutived pass tests to mark metric as healthy, 1 means metric is healthy on first fail
	FailureReminder    time.Duration  `json:"failure_reminder" default:"2h"`
	Grouping           GroupingConfig `json:"grouping" default:"{}"`
}
