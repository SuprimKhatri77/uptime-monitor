package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func generateRandomString() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func GithubLoginHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		state, err := generateRandomString()
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.Error("Failed to process request", constants.InternalServerError))
			return
		}

		utils.SetCookie(c, "oauth_state", state, 600, cfg)

		params := url.Values{}
		params.Set("client_id", cfg.GithubClientID)
		params.Set("redirect_uri", cfg.GithubRedirectURI)
		params.Set("scope", "read:user user:email")
		params.Set("state", state)

		authURL := "https://github.com/login/oauth/authorize?" + params.Encode()

		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}
