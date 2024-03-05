package scheduler

import (
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
)

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func Schedule(taskList []Tasker) {

	const minimalSleep = time.Second * 1 // sleep for at least 1 second (throttling)

	for {
		now := time.Now()
		for _, m := range taskList {
			mCopy := m // fixed in golang 1.22 ? (when running in //)
			if mCopy.NextRun().Before(now) {
				logging.Debug("Executing %v", m)
				mCopy.Run()
			}
		}

		now = time.Now()
		var nextSchedule time.Duration = time.Minute // run at least once every minute
		for _, m := range taskList {
			nextSchedule = min(nextSchedule, m.NextRun().Sub(now))
		}

		if nextSchedule < minimalSleep {
			nextSchedule = minimalSleep
		}
		time.Sleep(nextSchedule)

	}
}
