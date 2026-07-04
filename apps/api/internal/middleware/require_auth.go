package middleware

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
)

func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			slog.Warn("missing access token",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			respond.UnauthorizedAbort(c, "Missing access token", constants.TokenNotProvided)
			return
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				slog.Error("unexpected signing method",
					"alg", token.Header["alg"],
				)
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTAccessSecret), nil
		})

		if err != nil || !token.Valid {
			slog.Warn("invalid access token",
				"error", err,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			respond.UnauthorizedAbort(c, "Invalid access token", constants.TokenInvalid)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			slog.Warn("invalid token claims structure",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			respond.UnauthorizedAbort(c, "Invalid token claims", constants.Unauthorized)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			slog.Warn("missing userID in claims",
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			respond.UnauthorizedAbort(c, "Invalid token claims", constants.Unauthorized)
			return
		}

		role, ok := claims["role"].(string)
		if !ok || (role != "admin" && role != "superadmin" && role != "student") {
			slog.Warn("invalid role in claims",
				"user_id", userID,
				"role", role,
				"path", c.FullPath(),
				"ip", c.ClientIP(),
			)

			respond.UnauthorizedAbort(c, "Invalid token claims", constants.Unauthorized)
			return
		}

		slog.Info("authenticated request",
			"user_id", userID,
			"role", role,
			"path", c.FullPath(),
			"ip", c.ClientIP(),
		)

		c.Set("userID", userID)
		c.Set("role", role)

		c.Next()
	}
}
