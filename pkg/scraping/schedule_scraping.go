package scraping

import (
	"log"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scheduler"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
)

const maxParallelScrapingJobs uint = 8

func ScheduleScraping(providerCfgList map[string]provider.Config, storageInstance storage.Storager, resultChan chan<- provider.ScrapeResult) {
	taskList := []scheduler.Tasker{}
	taskList = append(taskList, scheduler.MakePeriodicTask(func() { storageInstance.Sync(false) }, time.Second*30))

	instanciatedProviderTypeMap := map[string]int{}
	// Load and schedule providers
	for providerName, providerCfg := range providerCfgList {
		providerInstance := provider.LoadProviderFromConfig(providerCfg)
		instanciatedProviderTypeMap[providerCfg.Type]++
		if !providerInstance.MultipleInstanceAllowed() {
			if instanciatedProviderTypeMap[providerCfg.Type] >= 2 {
				logging.Fatal("Cannot instantiate provider %v multiple times", providerCfg.Type)
			}
		}
		taskList = append(taskList, scheduler.MakePeriodicTask(func() {
			result := provider.MakeScrapeResult(providerName)
			providerInstance.Update(&result, storage.NewSubStorage(storageInstance, providerName+"/"))
			resultChan <- result
		}, time.Second*time.Duration(providerCfg.ScrapeInterval)))
	}

	scheduler := scheduler.MakeScheduler(taskList)

	log.Printf("Start collecting metrics (%d providers, %d max threads)", len(providerCfgList), maxParallelScrapingJobs)
	scheduler.ScheduleAsync(maxParallelScrapingJobs)
}
