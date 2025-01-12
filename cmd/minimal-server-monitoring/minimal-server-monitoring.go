package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/alert"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/config"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/logging"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/storage"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " config.json")
}

func makeStartupMessage(cfg *config.Config) notifier.Message {
	startupMsg := fmt.Sprintf("Monitoring started:\n - %v notifier(s)", len(cfg.Notifiers))
	startupMsg += fmt.Sprintf("\n - %v scraper(s):", len(cfg.Scrapers))
	for scraperName, scraperCfg := range cfg.Scrapers {
		startupMsg += fmt.Sprintf("\n  * %v (%v) every %v", scraperName, scraperCfg.Type, scraperCfg.ScrapeInterval)
	}
	return notifier.MakeMessage(notifier.Notification, "%v", startupMsg)
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	configPath := os.Args[1]
	cfg := config.LoadConfiguration(configPath)

	storage := storage.NewJSONStorage(cfg.CachePath)
	storage.Sync(true) // Test if storage can be synced

	scrapeResultChan := make(chan any, 5)
	notifyChan := make(chan notifier.Message, 5)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(2)
	// alert center
	alert.AlertCenter(ctx, cfg.Alert, scrapeResultChan, notifyChan)

	// notifier
	go func() {
		defer wg.Done()
		notifier.LoadAndRunNotifiers(ctx, cfg.MachineName, cfg.Notifiers, notifyChan)
	}()

	// Start metric scraping
	go func() {
		defer wg.Done()
		scraping.ScheduleScraping(ctx, cfg.Scrapers, storage, scrapeResultChan)
	}()

	notifyChan <- makeStartupMessage(&cfg)

	<-signalChan
	logging.Info("Exiting...")
	storage.Sync(false)
	cancel()
	logging.Info("Waiting for jobs to finish")
	if utils.WaitTimeout(&wg, 5*time.Second) {
		logging.Error("Failed to exit properly")
	}

}
