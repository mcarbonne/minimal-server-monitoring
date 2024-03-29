package alert

type Config struct {
	UnhealthyThreshold uint `json:"unhealthy_threshold" default:"1"` // how many consecutive failed tests to mark metric as unhealthy, 1 means metric is unhealthy on first fail
	HealthyThreshold   uint `json:"healthy_threshold" default:"1"`   // how many consecutived pass tests to mark metric as healthy, 1 means metric is healthy on first fail
}
