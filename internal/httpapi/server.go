package httpapi

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/config"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/service"
	"github.com/gin-gonic/gin"
)

type EngineConfig struct {
	Config    config.Config
	Catalog   *service.Catalog
	Logger    *slog.Logger
	Processor *events.Processor
	Store     repository.Store
	Bus       events.Bus
	PutRaw    func(ctx context.Context, e events.Envelope) error
}

func NewEngine(cfg config.Config, catalog *service.Catalog, logger *slog.Logger, proc *events.Processor) *gin.Engine {
	return NewEngineWith(EngineConfig{
		Config: cfg, Catalog: catalog, Logger: logger, Processor: proc,
	})
}

func NewEngineWith(opt EngineConfig) *gin.Engine {
	cfg := opt.Config
	if cfg.Env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestIDMiddleware())
	r.Use(SecurityHeaders())
	r.Use(CORSMiddleware(cfg.CORSOrigins))
	r.Use(accessLog(opt.Logger))

	storeName := cfg.StoreDriver
	if storeName == "" {
		storeName = "memory"
	}
	h := &Handler{
		catalog:     opt.Catalog,
		processor:   opt.Processor,
		logger:      opt.Logger,
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
		gitSHA:      cfg.GitSHA,
		storeName:   storeName,
		store:       opt.Store,
		bus:         opt.Bus,
		putRaw:      opt.PutRaw,
		reset:       resetGate{every: 15 * time.Second},
	}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
		v1.GET("/ready", h.Ready)
		v1.GET("/status", h.SystemStatus)
		v1.GET("/services", h.ListServices)
		v1.GET("/services/:id", h.GetService)
		v1.GET("/releases", h.ListReleases)
		v1.GET("/releases/:id", h.GetRelease)
		v1.GET("/releases/:id/risk", h.GetReleaseRisk)
		v1.GET("/releases/:id/health", h.GetReleaseHealth)
		v1.GET("/releases/:id/impact", h.GetReleaseImpact)
		v1.GET("/releases/:id/policy", h.GetReleasePolicy)
		v1.GET("/releases/:id/rollback", h.GetReleaseRollback)
		v1.GET("/releases/:id/timeline", h.GetReleaseTimeline)
		v1.GET("/replay", h.Replay)
		v1.POST("/events", h.IngestEvent)
		v1.POST("/demo/reset", h.ResetDemo)
	}
	r.NoRoute(func(c *gin.Context) {
		writeError(c, 404, "NOT_FOUND", "route not found")
	})
	return r
}

func accessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.InfoContext(c.Request.Context(), "http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	}
}

type resetGate struct {
	mu    sync.Mutex
	last  time.Time
	every time.Duration
}

func (g *resetGate) allow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	if !g.last.IsZero() && now.Sub(g.last) < g.every {
		return false
	}
	g.last = now
	return true
}
