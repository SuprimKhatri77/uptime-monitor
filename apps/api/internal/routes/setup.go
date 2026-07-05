package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/routes/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
)

func Setup(r *gin.Engine, cfg config.Config) {
	router := r.Group("/api/v1")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.Success("Server is up and running", nil))
	})

	setupAuthRoutes(router, cfg)
	setupDocsRoutes(router, cfg)
}
