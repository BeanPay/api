package server

import (
	"encoding/json"
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type authResponseBody struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

func (s *Server) login() http.HandlerFunc {
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

		// Generate a Signed JWT token
		tokenExpiration := time.Now().Add(15 * time.Minute)
		token, err := jwt.GenerateSignedToken(user.Id, tokenExpiration)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, authResponseBody{
			Token:      token,
			Expiration: tokenExpiration,
		})
	}
}
