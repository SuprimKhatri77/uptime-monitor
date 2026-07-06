package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func generateAccessAndRefreshToken(c *gin.Context, cfg *config.Config, user db.CoreUser, repo repository.AuthRepository) error {

	// jti prevents two refresh/access tokens for the same user in the same second to not be identical
	jti, err := generateRandomString()
	if err != nil {
		return fmt.Errorf("generating jti: %w", err)
	}

	// creating access claims
	accessClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"name":       user.Name,
		"avatar_url": user.AvatarUrl,
		"jti":        jti,
		"exp":        time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
	if err != nil {
		return fmt.Errorf("signing access token: %w", err)
	}

	// creating refresh claims
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"jti":     jti,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
	if err != nil {
		return fmt.Errorf("signing refresh token: %w", err)
	}

	// hashing the refresh token
	hash := sha256.Sum256([]byte(refreshTokenString))
	tokenHash := hex.EncodeToString(hash[:])

	// type casting the expires_at since its type was pgtype.timestampz it wouldn't accept the standard time type
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true}

	// saving the refresh token in DB
	_, err = repo.CreateRefreshToken(c.Request.Context(), db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	// setting the auth cookies
	setAuthCookies(refreshTokenString, accessTokenString, cfg, c)

	return nil

}

func setAuthCookies(refreshToken, accessToken string, cfg *config.Config, c *gin.Context) {

	utils.SetAuthCookie(c, "access_token", accessToken, 15*60, cfg)
	utils.SetAuthCookie(c, "refresh_token", refreshToken, 30*24*60*60, cfg)
	utils.SetPublicCookie(c, "is_logged_in", "true", 30*24*60*60, cfg)

}
