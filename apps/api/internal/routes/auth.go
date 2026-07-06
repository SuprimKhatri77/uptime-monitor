package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/middleware"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/routes/config"

	authHandler "github.com/suprimkhatri77/uptime-monitor/api/internal/handlers/auth"
)

func setupAuthRoutes(router *gin.RouterGroup, cfg config.Config) {
	auth := router.Group("/auth")

	auth.POST("/logout", authHandler.Logout(cfg.Queries, cfg.Config))
	auth.POST("/refresh", authHandler.Refresh(cfg.Queries, cfg.Config))
	auth.GET("/me", middleware.RequireAuth(cfg.Config), middleware.RequireRole("superadmin", "user"), authHandler.Me(cfg.Queries))
	auth.GET("/github", authHandler.GithubLoginHandler(cfg.Config))
	auth.GET("/github/callback", authHandler.GithubCallbackHandler(cfg.AuthTxRepo, cfg.Queries, cfg.Config, cfg.PgxPool))
}
