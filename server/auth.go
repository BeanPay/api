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
			Secure:   true,
			HttpOnly: true,
		})
		resp.SetResult(http.StatusOK, authResponseBody{
			AccessToken:           accessToken,
			AccessTokenExpiration: accessTokenExpiration,
		})
	}
}

func (s *Server) authRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()
		resp.SetResult(http.StatusOK, nil)
	}
}
