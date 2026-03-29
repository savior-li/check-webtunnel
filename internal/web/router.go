package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tor-bridge-collector/internal/config"
	"tor-bridge-collector/internal/exporter"
	"tor-bridge-collector/internal/storage"
	"tor-bridge-collector/internal/validator"
	"tor-bridge-collector/pkg/models"
)

func NewRouter(s *storage.Storage, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(corsMiddleware())

	api := router.Group("/api")
	{
		api.GET("/bridges", func(c *gin.Context) {
			handleGetBridges(c, s)
		})
		api.GET("/bridges/:id", func(c *gin.Context) {
			handleGetBridge(c, s)
		})
		api.GET("/bridges/:id/history", func(c *gin.Context) {
			handleGetBridgeHistory(c, s)
		})
		api.GET("/stats", func(c *gin.Context) {
			handleGetStats(c, s)
		})
		api.POST("/bridges/:id/validate", func(c *gin.Context) {
			handleValidateBridge(c, s, cfg)
		})
		api.POST("/export", func(c *gin.Context) {
			handleExport(c, s)
		})
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func handleGetBridges(c *gin.Context, s *storage.Storage) {
	filter := &models.BridgeFilter{}

	if transport := c.Query("transport"); transport != "" {
		filter.Transport = transport
	}
	if isValid := c.Query("is_valid"); isValid != "" {
		val := isValid == "true"
		filter.IsValid = &val
	}
	if page := c.Query("page"); page != "" {
		filter.Page, _ = strconv.Atoi(page)
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		filter.PageSize, _ = strconv.Atoi(pageSize)
	}

	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}

	bridges, total, err := s.GetAllBridges(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bridges":   bridges,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

func handleGetBridge(c *gin.Context, s *storage.Storage) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	bridge, err := s.GetBridgeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bridge not found"})
		return
	}

	c.JSON(http.StatusOK, bridge)
}

func handleGetBridgeHistory(c *gin.Context, s *storage.Storage) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	histories, err := s.GetValidationHistory(uint(id), 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, histories)
}

func handleGetStats(c *gin.Context, s *storage.Storage) {
	stats, err := s.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func handleValidateBridge(c *gin.Context, s *storage.Storage, cfg *config.Config) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	bridge, err := s.GetBridgeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bridge not found"})
		return
	}

	v := validator.New(s, cfg.Validation.Timeout, cfg.Validation.Concurrency, cfg.Validation.Retry)
	result := v.ValidateBridge(c.Request.Context(), bridge)

	bridge.IsValid = result.IsReachable
	if result.IsReachable {
		bridge.AvgLatency = result.Latency
		bridge.SuccessRate = 100
	}
	s.UpdateBridge(bridge)

	history := &models.ValidationHistory{
		BridgeID:    bridge.ID,
		TestedAt:    time.Now(),
		Latency:     result.Latency,
		IsReachable: result.IsReachable,
		ErrorMsg:    result.ErrorMsg,
	}
	s.CreateValidationHistory(history)

	c.JSON(http.StatusOK, result)
}

func handleExport(c *gin.Context, s *storage.Storage) {
	format := c.DefaultQuery("format", "torrc")
	path := c.DefaultQuery("path", "")

	exp := exporter.New(s)
	count, err := exp.Export(path, models.ExportFormat(format), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count, "path": path})
}
