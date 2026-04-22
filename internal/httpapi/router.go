package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/YY404NF/pq-backend/internal/config"
	"github.com/YY404NF/pq-backend/internal/query"
)

func NewRouter(cfg config.Config, service *query.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	router.GET("/api/health", func(c *gin.Context) {
		version, err := service.Version(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"server":         cfg.ServerName,
			"party":          cfg.Party,
			"status":         "ok",
			"datasetVersion": version.DatasetVersion,
		})
	})

	router.GET("/api/catalog/version", func(c *gin.Context) {
		version, err := service.Version(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, version)
	})

	router.GET("/api/catalog/list", func(c *gin.Context) {
		items, err := service.CatalogItems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	router.POST("/api/query/eval", func(c *gin.Context) {
		var req query.EvalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		resp, err := Eval(ctx, cfg, service, req)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, query.ErrVersionMismatch) {
				status = http.StatusConflict
			} else if errors.Is(err, ErrBadRequest) {
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{"error": err.Error(), "elapsedMs": time.Since(start).Milliseconds()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
