package metric

import (
	"log"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/scheduler"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

func ScheduleMetrics(metrics []*Metric, storageInstance storage.Storager, eventChan chan<- Event) {

	taskList := []scheduler.Tasker{}

	taskList = append(taskList, scheduler.MakePeriodicTask(func() { storageInstance.Sync(false) }, time.Second*30))
	for _, m := range metrics {
		taskList = append(taskList, scheduler.MakePeriodicTask(func() {
			m.Update(storage.NewSubStorage(storageInstance, "metrics/"+m.name+"/"), eventChan)
		}, m.policy.refreshInterval))
	}
	log.Printf("Starting metrics scheduler (%d metrics)", len(metrics))
	scheduler.Schedule(taskList)
}
