package scraping

import (
	"context"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scheduler"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
)

const maxParallelScrapingJobs uint = 8

func ScheduleScraping(ctx context.Context, providerCfgList map[string]provider.Config, storageInstance storage.Storager, resultChan chan<- any) {
	taskList := []scheduler.Tasker{}
	taskList = append(taskList, scheduler.MakePeriodicTask(func() { storageInstance.Sync(false) }, time.Second*30))

	providerList := make([]provider.Provider, 0, len(providerCfgList))
	instanciatedProviderTypeMap := map[string]int{}
	// Load and schedule providers
	for providerName, providerCfg := range providerCfgList {
		if !utils.IsNameValid(providerName) {
			logging.Fatal("Unable to setup scraper: forbidden characters in name '%v'", providerName)
		}

		providerInstance, err := provider.LoadProviderFromConfig(ctx, providerCfg)

		if err != nil {
			logging.Fatal("Unable to setup scraper '%v': %v", providerName, err)
		}

		providerList = append(providerList, providerInstance)
		instanciatedProviderTypeMap[providerCfg.Type]++
		if !providerInstance.MultipleInstanceAllowed() {
			if instanciatedProviderTypeMap[providerCfg.Type] >= 2 {
				logging.Fatal("Cannot instantiate provider %v multiple times", providerCfg.Type)
			}
		}
		resultWrapper := provider.MakeScrapeResultWrapper(providerName, resultChan)

		updateTaskList := providerInstance.GetUpdateTaskList(ctx, &resultWrapper, storage.NewSubStorage(storageInstance, providerName))

		for _, updateTask := range updateTaskList {
			taskList = append(taskList, scheduler.MakePeriodicTask(updateTask, providerCfg.ScrapeInterval))
		}
	}

	scheduler := scheduler.MakeScheduler(taskList)

	logging.Info("Start collecting metrics (%d providers, %d max threads)", len(providerCfgList), maxParallelScrapingJobs)
	scheduler.ScheduleAsync(ctx, maxParallelScrapingJobs)
	logging.Info("Exiting scraping...")
	for _, providerInstance := range providerList {
		providerInstance.Destroy()
	}
	logging.Info("Done")
}
