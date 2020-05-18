package server

import (
	"encoding/json"
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 90 * (24 * time.Hour)
)

type authResponseBody struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiration time.Time `json:"access_token_expiration"`
}

func (s *Server) login() http.HandlerFunc {
	refreshTokenRepo := models.RefreshTokenRepository{DB: s.DB}
	userRepo := models.UserRepository{DB: s.DB}
	type RequestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()

		//  Parse & Validate the Body
		var requestBody RequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			resp.SetResult(http.StatusBadRequest, nil).
				WithErrorDetails("Failed to parse the request body.")
			return
		}
		messages, err := s.Validator.Validate(requestBody)
		if err != nil {
			resp.SetResult(http.StatusBadRequest, nil).
				WithErrorDetails(messages...)
			return
		}

		// Fetch the user
		user, err := userRepo.FetchByEmail(requestBody.Email)
		if err != nil {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Validate the Password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestBody.Password))
		if err != nil {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Generate a Signed JWT AccessToken
		accessTokenExpiration := time.Now().Add(accessTokenDuration)
		accessToken, err := jwt.GenerateSignedToken(user.Id, accessTokenExpiration)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// Generate a RefreshToken
		chainId := uuid.NewV4()
		refreshToken := &models.RefreshToken{
			ChainId: chainId.String(),
			UserId:  user.Id,
		}
		err = refreshTokenRepo.Insert(refreshToken)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken.Id,
			Expires:  time.Now().Add(refreshTokenDuration),
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		resp.SetResult(http.StatusOK, authResponseBody{
			AccessToken:           accessToken,
			AccessTokenExpiration: accessTokenExpiration,
		})
	}
}

func (s *Server) authRefresh() http.HandlerFunc {
	refreshTokenRepo := models.RefreshTokenRepository{DB: s.DB}

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()

		// Get the refresh token
		refreshTokenCookie, err := r.Cookie("refresh_token")
		if err != nil {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Load the Refresh Token
		refreshToken, err := refreshTokenRepo.FetchByID(refreshTokenCookie.Value)
		if err != nil {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Load the most recent RefreshToken in the chain so we can ensure
		// the latest was passed in.
		latestToken, err := refreshTokenRepo.FetchMostRecentInChain(refreshToken.ChainId)
		if err != nil {
			// This shouldn't happen, as we just verified that there is at least
			// one refresh token in this chain, so throwing a 500
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}
		// If the RefreshToken that was passed in isn't the latest, something
		// nefarious is likely happening so we wipe the entire chain to evict
		// any potential bad actors that are using an upstream token.
		// https://auth0.com/docs/tokens/concepts/refresh-token-rotation#automatic-reuse-detection
		if refreshToken.Id != latestToken.Id {
			err := refreshTokenRepo.DeleteChain(refreshToken.ChainId)
			if err != nil {
				resp.SetResult(http.StatusInternalServerError, nil)
				return
			}
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Verify that the refreshToken isn't expired
		tokenExpiry := refreshToken.CreatedAt.Add(refreshTokenDuration)
		if time.Now().After(tokenExpiry) {
			// Wipe the chain from the DB.  We've verified this is the latest
			// link in the chain and it is expired. This chain is dead now,
			// so just  clean this chain out of the DB.
			err := refreshTokenRepo.DeleteChain(refreshToken.ChainId)
			if err != nil {
				resp.SetResult(http.StatusInternalServerError, nil)
				return
			}
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Generate a new Signed JWT AccessToken
		accessTokenExpiration := time.Now().Add(accessTokenDuration)
		accessToken, err := jwt.GenerateSignedToken(refreshToken.UserId, accessTokenExpiration)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// Generate a new RefreshToken
		newRefreshToken := &models.RefreshToken{
			ChainId: refreshToken.ChainId,
			UserId:  refreshToken.UserId,
		}
		err = refreshTokenRepo.Insert(newRefreshToken)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken.Id,
			Expires:  time.Now().Add(refreshTokenDuration),
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		resp.SetResult(http.StatusOK, authResponseBody{
			AccessToken:           accessToken,
			AccessTokenExpiration: accessTokenExpiration,
		})
	}
}
