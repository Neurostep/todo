package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
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
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

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
		AuthEnabled:        cfg.Server.AuthEnabled,
		DB:                 db,
		TodoService:        todoService,
		Logger:             log.With(logger, "service", "http"),
		PrometheusExporter: prometheusExporter,
	}
	s := server.New(serverCfg)

	group, groupCtx := errgroup.WithContext(ctx)

	// server start
	group.Go(func() error {
		err := s.Start(groupCtx)

		return err
	})

	// signal handlers
	interrupt := make(chan os.Signal, 2)
	cancel := make(chan struct{})
	defer close(interrupt)

	group.Go(func() error {
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-interrupt:
			logger.Log("event", "captured signal", sig)
			cancelFunc()
		case <-groupCtx.Done():
		case <-cancel:
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		logger.Log("event", "error", "cause", err)
		close(cancel)
		signal.Stop(interrupt)
	}
}
