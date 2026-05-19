package main

import (
	"flag"
	"fmt"
	"monitoring_server/pkg/collector"
	"monitoring_server/pkg/common/log"
	"monitoring_server/pkg/writer"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	Version = "dev-"
)

var (
	timezone    = flag.String("timezone", "", "timezone")
	configFile  = flag.String("config", "", "config file")
	versionFlag = flag.Bool("version", false, "show version")
	logFile     = flag.String("log", "", "log path")
	logMaxAge   = flag.Int("log-max-age", 30, "log max age(days)")
)

func main() {
	flag.Parse()

	if versionFlag != nil && *versionFlag {
		fmt.Printf("Collector %s\n", Version)
		os.Exit(0)
	}

	maxAge := 0
	if logMaxAge != nil {
		maxAge = *logMaxAge
	}

	if logFile != nil && *logFile != "" {
		err := log.SetLogFile(*logFile, maxAge)
		if err != nil {
			fmt.Printf("Log File Error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	logger := log.GetLogger()

	cfg := defaultConfig
	if configFile != nil && *configFile != "" {
		var err error
		err = loadYamlFile(*configFile, cfg)
		if err != nil {
			logger.WithError(err).Errorf("Can't load file %s", *configFile)
			os.Exit(1)
		}
	}

	if timezone != nil && *timezone != "" {
		loc, err := time.LoadLocation(*timezone)
		if err != nil {
			logger.WithError(err).Errorf("Unknown timezone %s. Using system's local time.", *timezone)
		} else {
			time.Local = loc
		}
	}

	var err error

	writer.SetLogger(logger)
	err = writer.Load(cfg.Writer)
	if err != nil {
		logger.WithError(err).Error("failed to load writers")
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collector.SetLogger(logger)
	err = collector.Load(cfg.Collector)
	if err != nil {
		logger.WithError(err).Error("failed to load collectors")
		os.Exit(1)
	}
	logger.Infof("Start Collector (version:%s)\n", Version)
	wg := sync.WaitGroup{}
	err = collector.StartAll(&wg)
	if err != nil {
		logger.WithError(err).Error("failed to start all collectors")
		wg.Wait()
		os.Exit(1)
	}

	select {
	case <-quit:
		logger.Infof("Server is waiting for stop\n")
		collector.StopAll()
	}

	wg.Wait()
}
