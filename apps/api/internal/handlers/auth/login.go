package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=50"`
}

func Login(
	queries repository.AuthRepository,
	cfg *config.Config,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var loginRequest LoginRequest
		if !respond.BindJSON(c, &loginRequest, "Invalid request data") {
			handlerlog.Warn(c, "invalid request payload")
			return
		}

		handlerlog.Info(c, "login attempt")

		user, err := queries.GetUserByEmail(ctx, loginRequest.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				handlerlog.Warn(c, "invalid credentials (user not found)")

				respond.Unauthorized(c, "Invalid credentials", constants.InvalidCredentials)
				return
			}

			handlerlog.Error(c, "failed to fetch user", err)

			respond.InternalError(c, "Something went wrong")
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte(loginRequest.Password),
		)
		if err != nil {
			handlerlog.Warn(c, "invalid credentials (password mismatch)", "user_id", user.ID)

			respond.Unauthorized(c, "Invalid credentials", constants.InvalidCredentials)
			return
		}

		handlerlog.Info(c, "password verified", "user_id", user.ID)

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

			respond.InternalError(c, "Something went wrong")
			return
		}

		sessionID := uuid.New()

		refreshClaims := jwt.MapClaims{
			"user_id":    user.ID,
			"session_id": sessionID,
			"exp":        time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		if err != nil {
			handlerlog.Error(c, "failed to sign refresh token", err, "user_id", user.ID)

			respond.InternalError(c, "Something went wrong")
			return
		}

		expiresAt := pgtype.Timestamptz{
			Time:  time.Now().Add(30 * 24 * time.Hour),
			Valid: true,
		}

		refreshTokenHash := sha256.Sum256([]byte(refreshTokenString))
		refreshTokenHashString := fmt.Sprintf("%x", refreshTokenHash)

		_, err = queries.CreateToken(ctx, db.CreateTokenParams{
			UserID:    user.ID,
			Token:     refreshTokenHashString,
			SessionID: pgtype.UUID{Bytes: sessionID, Valid: true},
			ExpiresAt: expiresAt,
		})
		if err != nil {
			handlerlog.Error(c, "failed to store refresh token", err, "user_id", user.ID)

			respond.InternalError(c, "Something went wrong")
			return
		}

		handlerlog.Info(c, "tokens issued", "user_id", user.ID)

		utils.SetAuthCookie(c, "access_token", accessTokenString, 15*60, cfg)
		utils.SetAuthCookie(c, "refresh_token", refreshTokenString, 30*24*60*60, cfg)
		utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

		handlerlog.Info(c, "login successful", "user_id", user.ID)

		respond.OK(c, "logged in successfully", db.User{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Role:     user.Role,
			ImageUrl: user.ImageUrl,
		})
	}
}
