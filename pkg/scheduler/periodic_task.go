package scheduler

import "time"

type PeriodicTask struct {
	taskFunc        func()
	refreshInterval time.Duration
	lastRun         time.Time
}

func MakePeriodicTask(taskFunc func(), refreshInterval time.Duration) Tasker {
	return &PeriodicTask{taskFunc: taskFunc,
		refreshInterval: refreshInterval}
}

func (task *PeriodicTask) NextRun() time.Time {
	return task.lastRun.Add(task.refreshInterval)
}

func (task *PeriodicTask) Run() {
	task.lastRun = time.Now()
	task.taskFunc()
}
