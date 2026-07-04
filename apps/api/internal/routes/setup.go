package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/routes/config"
)

func Setup(r *gin.Engine, cfg config.Config) {
	router := r.Group("/api/v1")

	router.GET("/health", func(c *gin.Context) {
		respond.OKMessage(c, "Server is up and running")
	})

	setupAuthRoutes(router, cfg)
	setupDocsRoutes(router, cfg)
}
