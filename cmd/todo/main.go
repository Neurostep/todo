package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-kit/kit/log"
	"go.opencensus.io/stats/view"
	"golang.org/x/sync/errgroup"

	"github.com/Neurostep/todo/config"
	"github.com/Neurostep/todo/internal/server"
	"github.com/Neurostep/todo/pkg/database"
	"github.com/Neurostep/todo/pkg/services/todo"
	"github.com/Neurostep/todo/pkg/tools/metrics"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var waitStopping sync.WaitGroup
	waitStopping.Add(1)
	go func() {
		Run(ctx)
		waitStopping.Done()
	}()

	select {
	case <-ctx.Done():
	case <-interrupt:
		cancelFunc()
	}
	waitStopping.Wait()
}

func Run(ctx context.Context) {
	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	configPath := flag.String("cfg", "", "config file")
	flag.Parse()

	cfg, err := config.ReadConfigFile(*configPath)
	if err != nil {
		logger.Log("error", "failed to read config file", "configPath", configPath, "cause", err)
		os.Exit(1)
	}

	cfgDatabase := database.Config{
		Address: cfg.Database.Address,
	}
	db, err := database.New(cfgDatabase, log.With(logger, "service", "database"))
	if err != nil {
		logger.Log("error", "failed to setup database connection", "cause", err)
		os.Exit(1)
	}
	defer db.Close()

	cfgMetrics := metrics.Config{
		TracingEnable: cfg.Metrics.TracingEnable,
	}
	tracingDone, err := metrics.SetupTracer("todo_service", cfgMetrics, log.With(logger, "service", "tracing"))
	if err != nil {
		logger.Log("error", "failed to setup tracing", "cause", err)
	}
	defer tracingDone()

	prometheusExporter, err := metrics.SetupMonitoring("todo_service", cfgMetrics, log.With(logger, "service", "monitoring"))
	if err != nil {
		logger.Log("error", "failed to setup prometheus monitoring", "cause", err)
		prometheusExporter = nil
		// continue
	} else {
		defer func() {
			view.UnregisterExporter(prometheusExporter)
		}()
	}

	todoService := todo.New(todo.Config{
		DB:     db,
		Logger: log.With(logger, "service", "todo"),
	})

	serverCfg := server.Config{
		Debug:              cfg.Server.Debug,
		Port:               cfg.Server.Port,
		DB:                 db,
		TodoService:        todoService,
		Logger:             log.With(logger, "service", "http"),
		PrometheusExporter: prometheusExporter,
	}
	s := server.New(serverCfg)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return s.Start(groupCtx) })
	if err := group.Wait(); err != nil {
		logger.Log("event", "error", "cause", err)
	}
}
