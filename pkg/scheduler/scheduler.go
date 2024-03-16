package scheduler

import (
	"sync"
	"time"
)

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

type Scheduler struct {
	taskList []Tasker
}

func MakeScheduler(taskList []Tasker) Scheduler {
	return Scheduler{
		taskList: taskList,
	}
}

func (s *Scheduler) Schedule() {

	const minimalSleep = time.Second * 1 // sleep for at least 1 second (throttling)

	for {
		now := time.Now()
		for _, m := range s.taskList {
			if m.NextRun().Before(now) {
				m.Run()
			}
		}

		now = time.Now()
		var nextSchedule time.Duration = time.Minute // run at least once every minute
		for _, m := range s.taskList {
			nextSchedule = min(nextSchedule, m.NextRun().Sub(now))
		}

		if nextSchedule < minimalSleep {
			nextSchedule = minimalSleep
		}
		time.Sleep(nextSchedule)

	}
}

func (s *Scheduler) ScheduleAsync(maxJobInParallel uint) {

	const minimalSleep = time.Second * 1 // sleep for at least 1 second (throttling)

	jobQueue := make(chan func())
	for range maxJobInParallel {
		go func() {
			for task := range jobQueue {
				task()
			}

		}()
	}

	for {
		now := time.Now()
		iterationWg := sync.WaitGroup{}
		for _, m := range s.taskList {
			if m.NextRun().Before(now) {
				iterationWg.Add(1)
				jobQueue <- func() {
					defer iterationWg.Done()
					m.Run()
				}
			}
		}

		iterationWg.Wait() // This isn't optimal, but required as Tasker interface isn't threadsafe

		now = time.Now()
		var nextSchedule time.Duration = time.Minute // run at least once every minute
		for _, m := range s.taskList {
			nextSchedule = min(nextSchedule, m.NextRun().Sub(now))
		}

		if nextSchedule < minimalSleep {
			nextSchedule = minimalSleep
		}
		time.Sleep(nextSchedule)

	}
}
