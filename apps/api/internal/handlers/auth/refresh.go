package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Refresh(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenString, err := c.Cookie("refresh_token")
		if err != nil {
			handlerlog.Warn(c, "missing refresh token cookie")

			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusBadRequest, types.Error("Missing refresh token", constants.TokenNotProvided))
			return
		}

		token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				handlerlog.Error(c, "unexpected signing method", fmt.Errorf("unexpected signing method: %v", token.Header["alg"]), "alg", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(cfg.JWTRefreshSecret), nil
		})

		if err != nil || !token.Valid {
			handlerlog.Warn(c, "invalid refresh token", "error", err)

			utils.ClearAuthCookies(c, cfg)

			c.JSON(http.StatusUnauthorized, types.Error("Invalid refresh token", constants.TokenInvalid))
			return
		}

		refreshTokenHash := sha256.Sum256([]byte(refreshTokenString))
		refreshTokenHashString := fmt.Sprintf("%x", refreshTokenHash)

		refreshToken, err := queries.GetRefreshTokenByHash(ctx, refreshTokenHashString)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {

				utils.ClearAuthCookies(c, cfg)
				c.JSON(http.StatusUnauthorized, types.Error("Invalid refresh token", constants.TokenInvalid))
				return
			}

			handlerlog.Error(c, "failed to fetch refresh token", err)

			c.JSON(http.StatusInternalServerError, types.Error("Something went wrong", constants.InternalServerError))
			return
		}

		user, err := queries.GetUserByID(ctx, refreshToken.UserID)
		if err != nil {
			handlerlog.Error(c, "failed to fetch user", err, "user_id", refreshToken.UserID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		accessClaims := jwt.MapClaims{
			"user_id":    user.ID,
			"role":       user.Role,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarUrl,
			"exp":        time.Now().Add(15 * time.Minute).Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign access token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		if time.Since(refreshToken.CreatedAt.Time) < 5*time.Minute {
			utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
			c.JSON(http.StatusOK, types.Success("Access token refreshed", nil))
			return
		}

		err = queries.RevokeRefreshToken(ctx, refreshTokenHashString)
		if err != nil {
			handlerlog.Error(c, "failed to revoke refresh token", err, "user_id", refreshToken.UserID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		refreshClaims := jwt.MapClaims{
			"user_id": user.ID,
			"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		newRefreshTokenString, err := newRefreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		newHash := sha256.Sum256([]byte(newRefreshTokenString))
		newTokenHash := fmt.Sprintf("%x", newHash)

		_, err = queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: newTokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(30 * 24 * time.Hour),
				Valid: true,
			},
		})
		if err != nil {
			handlerlog.Error(c, "failed to persist new refresh token", err, "user_id", user.ID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
		utils.SetAuthCookie(c, "refresh_token", newRefreshTokenString, 30*24*60*60, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

		handlerlog.Info(c, "tokens rotated successfully", "user_id", user.ID)

		c.JSON(http.StatusOK, types.Success("Tokens refreshed", nil))
	}
}
