package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"

	"github.com/accrescent/devportal/auth"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*sql.DB)
		conf := c.MustGet("oauth2_config").(*oauth2.Config)

		sessionID, err := c.Cookie(auth.SessionCookie)
		if err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		var token string
		if err := db.QueryRow(
			"SELECT access_token FROM sessions WHERE id = ? AND expiry_time > ?",
			sessionID, time.Now().Unix(),
		).Scan(&token); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = c.AbortWithError(http.StatusForbidden, err)
			} else {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}

		httpClient := conf.Client(c, &oauth2.Token{AccessToken: token})
		client := github.NewClient(httpClient)

		c.Set("session_id", sessionID)
		c.Set("gh_client", client)
	}
}
