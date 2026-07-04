package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
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
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Refresh(queries repository.AuthRepository, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshTokenString, err := c.Cookie("refresh_token")
		if err != nil {
			handlerlog.Warn(c, "missing refresh token cookie")

			utils.ClearAuthCookies(c, cfg)

			respond.BadRequest(c, "Missing refresh token", constants.TokenNotProvided)
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
			utils.ClearAuthCookies(c, cfg)

			respond.Unauthorized(c, "Invalid token claims", constants.InvalidToken)
			return
		}

		handlerlog.Info(c, "session from claims", "session_id", sessionIDFromClaims)

		sessionID, err := utils.ConvertToUUID(sessionIDFromClaims)
		if err != nil {
			utils.ClearAuthCookies(c, cfg)

			respond.Unauthorized(c, "Invalid token claims", constants.InvalidToken)
			return
		}

		refreshTokenHash := sha256.Sum256([]byte(refreshTokenString))
		refreshTokenHashString := fmt.Sprintf("%x", refreshTokenHash)

		refreshToken, err := queries.GetRefreshTokenBySessionIDAndToken(ctx, db.GetRefreshTokenBySessionIDAndTokenParams{
			SessionID: sessionID,
			Token:     refreshTokenHashString,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {

				utils.ClearAuthCookies(c, cfg)
				respond.Unauthorized(c, "Invalid refresh token", constants.TokenInvalid)
				return
			}

			handlerlog.Error(c, "failed to fetch refresh token", err)

			respond.InternalError(c, "Something went wrong")
			return
		}

		user, err := queries.GetUserByID(ctx, refreshToken.UserID)
		if err != nil {
			handlerlog.Error(c, "failed to fetch user", err, "user_id", refreshToken.UserID)

			respond.InternalError(c, "Failed to process request")
			return
		}

		accessClaims := jwt.MapClaims{
			"userID":   user.ID,
			"role":     user.Role,
			"email":    user.Email,
			"name":     user.Name,
			"imageURL": user.ImageUrl,
			"exp":      time.Now().Add(15 * time.Minute).Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign access token", err, "user_id", user.ID)

			respond.InternalError(c, "Failed to process request")
			return
		}

		if time.Since(refreshToken.CreatedAt.Time) < 5*time.Minute {
			utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
			respond.OKMessage(c, "Access token refreshed")
			return
		}

		_, err = queries.RevokeTokenBySessionIDAndToken(ctx, db.RevokeTokenBySessionIDAndTokenParams{
			SessionID: sessionID,
			Token:     refreshTokenHashString,
		})
		if err != nil {
			handlerlog.Error(c, "failed to revoke refresh token", err, "user_id", refreshToken.UserID)

			respond.InternalError(c, "Failed to process request")
			return
		}

		refreshClaims := jwt.MapClaims{
			"user_id":    user.ID,
			"session_id": sessionID,
			"exp":        time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		newRefreshTokenString, err := newRefreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign refresh token", err, "user_id", user.ID)

			respond.InternalError(c, "Failed to process request")
			return
		}

		newHash := sha256.Sum256([]byte(newRefreshTokenString))
		newTokenHash := fmt.Sprintf("%x", newHash)

		_, err = queries.CreateToken(ctx, db.CreateTokenParams{
			UserID: user.ID,
			Token:  newTokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(30 * 24 * time.Hour),
				Valid: true,
			},
			SessionID: sessionID,
		})
		if err != nil {
			handlerlog.Error(c, "failed to persist new refresh token", err, "user_id", user.ID)

			respond.InternalError(c, "Failed to process request")
			return
		}

		utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
		utils.SetAuthCookie(c, "refresh_token", newRefreshTokenString, 30*24*60*60, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

		handlerlog.Info(c, "tokens rotated successfully", "user_id", user.ID)

		respond.OKMessage(c, "Tokens refreshed")
	}
}
