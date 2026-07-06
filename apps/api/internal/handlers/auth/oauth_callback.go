package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

type GithubCallbackHandlerParams struct {
	Code  string `form:"code" binding:"required,not_blank,min=1,max=200"`
	State string `form:"state" binding:"required,not_blank,min=1,max=50"`
}

func GithubCallbackHandler(txRepo repository.AuthTxRepository, repo repository.AuthRepository, cfg *config.Config, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var params GithubCallbackHandlerParams
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, types.Error("Invalid query parameters", constants.InvalidQueryParam))
			return
		}

		utils.SetCookie(c, "oauth_state", "", -1, cfg) // clear regardless of match or mismatch
		cookieState, err := c.Cookie("oauth_state")
		if err != nil || cookieState != params.State {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/auth?error=invalid_state")
			return
		}

		// exchanging short lived token for a longer one which we need to get the user info
		accessToken, err := exchangeCodeForToken(ctx, cfg, params.Code)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/auth?error=auth_failed")
			return
		}

		// fetching the user profile
		ghUser, err := fetchGithubUser(ctx, accessToken)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/auth?error=auth_failed")
			return
		}

		var user db.CoreUser
		err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
			qtx := txRepo.WithTx(tx)

			existingUser, err := qtx.GetUserByEmail(ctx, ghUser.Email)
			if err != nil {

				if !errors.Is(err, pgx.ErrNoRows) {
					return err // only reaches here for genuine DB errors, not ErrNoRows
				}

				// inserting the user in DB instead of returning error response
				newUser, err := qtx.CreateUser(ctx, db.CreateUserParams{
					Name:      utils.ToNullableText(ghUser.Name),
					Email:     ghUser.Email,
					AvatarUrl: utils.ToNullableText(ghUser.AvatarURL),
				})

				if err != nil {
					return err
				}

				_, err = qtx.CreateAccount(ctx, db.CreateAccountParams{
					UserID:            newUser.ID,
					Provider:          "github",
					ProviderAccountID: strconv.FormatInt(ghUser.ID, 10),
				})

				if err != nil {
					return err
				}

				user = newUser
				return nil

			}

			user = existingUser
			return nil
		})

		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/auth?error=auth_failed")
			return
		}

		if err := generateAccessAndRefreshToken(c, cfg, user, repo); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/auth?error=auth_failed")
			return
		}

		// TODO: check if the user belongs to a tenant and redirect accordingly
		c.Redirect(http.StatusTemporaryRedirect, cfg.FrontendURL+"/onboarding")

	}
}
