package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/routes/config"

	authHandler "github.com/suprimkhatri77/uptime-monitor/api/internal/handlers/auth"
)

func setupAuthRoutes(router *gin.RouterGroup, cfg config.Config) {
	auth := router.Group("/auth")

	auth.POST("/logout", authHandler.Logout(cfg.Queries, cfg.Config))
	auth.POST("/refresh", authHandler.Refresh(cfg.Queries, cfg.Config))
	auth.POST("/me", authHandler.Me(cfg.Queries))
}
