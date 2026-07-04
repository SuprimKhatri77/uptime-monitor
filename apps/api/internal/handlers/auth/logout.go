package auth

import (
	"crypto/sha256"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Logout(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenFromCookie, err := c.Cookie("refresh_token")
		if err != nil {
			handlerlog.Warn(c, "missing refresh token on logout")

			respond.Unauthorized(c, "Missing refresh token", constants.TokenNotProvided)
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

			respond.Unauthorized(c, "Invalid refresh token", constants.TokenInvalid)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			handlerlog.Warn(c, "invalid token claims")

			respond.Unauthorized(c, "Invalid token", constants.InvalidToken)
			return
		}

		sessionIDFromClaims, ok := claims["session_id"].(string)
		if !ok {
			respond.Unauthorized(c, "Invalid token claims", constants.InvalidToken)
			return
		}

		sessionID, err := utils.ConvertToUUID(sessionIDFromClaims)
		if err != nil {
			respond.Unauthorized(c, "Invalid token claims", constants.InvalidToken)
			return
		}

		hash := sha256.Sum256([]byte(refreshTokenFromCookie))
		tokenHash := fmt.Sprintf("%x", hash)

		_, err = queries.RevokeTokenBySessionIDAndToken(ctx, db.RevokeTokenBySessionIDAndTokenParams{
			Token:     tokenHash,
			SessionID: sessionID,
		})
		if err != nil {
			handlerlog.Error(c, "failed to revoke refresh token on logout", err)

			respond.InternalError(c, "Failed to logout")
			return
		}

		utils.SetAuthCookie(c, "access_token", "", -1, cfg)
		utils.SetAuthCookie(c, "refresh_token", "", -1, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "", -1, cfg)

		handlerlog.Info(c, "user logged out")

		respond.OKMessage(c, "Logged out successfully")
	}
}
