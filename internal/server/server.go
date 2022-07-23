package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/jinzhu/gorm"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/Neurostep/todo/pkg/services/todo"
	"github.com/Neurostep/todo/pkg/tools/metrics"
)

const timeout = 2 // seconds

type (
	Config struct {
		Debug       bool
		Port        int `validate:"required"`
		TodoService *todo.Service
		DB          *gorm.DB
		Logger      log.Logger

		PrometheusExporter *prometheus.Exporter
	}

	api struct {
		*http.Server
		conf   Config
		logger log.Logger
		pe     *prometheus.Exporter
	}
)

var ginMtx sync.Mutex

func New(c Config) *api {
	if !c.Debug {
		ginMtx.Lock()
		defer ginMtx.Unlock()
		gin.SetMode(gin.ReleaseMode)
	}

	r := &api{conf: c, logger: c.Logger}
	handler := &ochttp.Handler{
		Handler: r.routes(),
		FormatSpanName: func(req *http.Request) string {
			return fmt.Sprintf("recv.todo.http:%s", req.URL.Path)
		},
		IsPublicEndpoint: true,
	}
	r.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: http.TimeoutHandler(handler, time.Second, ""),

		ReadTimeout:  time.Second,
		WriteTimeout: time.Second * 2,
	}

	return r
}

func (r *api) routes() *gin.Engine {
	router := gin.New()
	router.Use(CORS)
	monitoredRouter := metrics.WrapGinRouter(router)
	router.GET("/healthz", r.healthz)
	router.GET("/readyz", r.readyz)

	// auth endpoints
	router.POST("/signin", r.signin)
	router.GET("/refresh", r.refresh)

	// API endpoints
	apiGroup := router.Group("/api/v1", authMiddleware)

	monitoredAPIGroup := metrics.WrapGinRouter(apiGroup)
	monitoredAPIGroup.Use(requireContentType(r.logger, "application/json"))

	todosGroup := metrics.WrapGinRouter(apiGroup)
	{
		todosGroup.GET("/todos", r.getTodos)
		todosGroup.POST("/todos", r.createTodo)
		todosGroup.GET("/todos/:id", r.getTodo)
		todosGroup.PUT("/todos/:id", r.updateTodo)
		todosGroup.DELETE("/todos/:id", r.deleteTodo)

		todoGroup := apiGroup.Group("todos/:id")
		todoComments := metrics.WrapGinRouter(todoGroup)
		{
			todoComments.POST("comments", r.addCommentToTodo)
			todoComments.GET("comments", r.getComments)
			todoComments.DELETE("comments/:commentId", r.removeCommentFromTodo)
		}
		todoLabels := metrics.WrapGinRouter(todoGroup)
		{
			todoLabels.POST("labels", r.addLabelToTodo)
			todoLabels.GET("labels", r.getLabels)
			todoLabels.DELETE("labels/:labelId", r.removeLabelFromTodo)
		}
	}

	if r.conf.PrometheusExporter != nil && r.setupPrometheusMetrics() == nil {
		monitoredRouter.GET("/metrics", gin.HandlerFunc(func(c *gin.Context) {
			ochttp.SetRoute(c.Request.Context(), "/metrics")
			r.conf.PrometheusExporter.ServeHTTP(c.Writer, c.Request)
		}))
	}
	return router
}

func (r *api) setupPrometheusMetrics() error {
	err := view.Register(
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,
		&view.View{
			Name:        "opencensus.io/http/server/route_latency",
			Description: "Latency distribution",
			TagKeys:     []tag.Key{ochttp.KeyServerRoute, ochttp.StatusCode},
			Measure:     ochttp.ServerLatency,
			Aggregation: ochttp.DefaultLatencyDistribution,
		},
	)
	if err != nil {
		r.logger.Log("event", "http_monitoring_view_register_failed", "error", err)
	}
	return err
}

func (r *api) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		r.logger.Log("event", "starting_server", "address", r.Addr)
		if err := r.ListenAndServe(); err != nil {
			r.logger.Log("error", err)
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		r.logger.Log("event", "shutting down...")
		c, cancel := context.WithTimeout(context.Background(), time.Second*timeout)
		defer cancel()
		err := r.Shutdown(c)

		if err != nil {
			r.logger.Log("error", err)
		}
	}
	return nil
}
