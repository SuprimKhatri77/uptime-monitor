package auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Logout(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenFromCookie, err := c.Cookie("refresh_token")
		if err != nil {
			handlerlog.Warn(c, "missing refresh token on logout")

			c.JSON(http.StatusUnauthorized, types.Error("Missing refresh token", constants.TokenNotProvided))
			return
		}

		token, err := jwt.Parse(refreshTokenFromCookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				handlerlog.Error(c, "unexpected signing method during logout", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]), "alg", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTRefreshSecret), nil
		})

		if err != nil || !token.Valid {
			handlerlog.Warn(c, "invalid refresh token on logout", "error", err)

			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.Error("Invalid refresh token", constants.TokenInvalid))
			return
		}

		hash := sha256.Sum256([]byte(refreshTokenFromCookie))
		tokenHash := fmt.Sprintf("%x", hash)

		err = queries.RevokeRefreshToken(ctx, tokenHash)
		if err != nil {
			handlerlog.Error(c, "failed to revoke refresh token on logout", err)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to logout", constants.InternalServerError))
			return
		}

		utils.SetAuthCookie(c, "access_token", "", -1, cfg)
		utils.SetAuthCookie(c, "refresh_token", "", -1, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "", -1, cfg)

		c.JSON(http.StatusOK, types.Success("Logged out successfully", nil))
	}
}
