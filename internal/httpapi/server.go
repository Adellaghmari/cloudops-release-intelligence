package httpapi

import (
	"log/slog"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/config"
	"github.com/adell/cloudops-release-intelligence/internal/service"
	"github.com/gin-gonic/gin"
)

func NewEngine(cfg config.Config, catalog *service.Catalog, logger *slog.Logger) *gin.Engine {
	if cfg.Env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestIDMiddleware())
	r.Use(SecurityHeaders())
	r.Use(CORSMiddleware(cfg.CORSOrigins))
	r.Use(accessLog(logger))

	h := &Handler{
		catalog:     catalog,
		logger:      logger,
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
	}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
		v1.GET("/ready", h.Ready)
		v1.GET("/services", h.ListServices)
		v1.GET("/services/:id", h.GetService)
		v1.GET("/releases", h.ListReleases)
		v1.GET("/releases/:id", h.GetRelease)
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
