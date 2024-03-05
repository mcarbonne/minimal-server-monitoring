package scheduler

import "time"

type Tasker interface {
	Run()
	NextRun() time.Time
}
