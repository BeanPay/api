package server

import (
	"encoding/json"
	"github.com/beanpay/api/database/models"
	"github.com/generalledger/response"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func (s *Server) createUser() http.HandlerFunc {
	userRepo := models.UserRepository{DB: s.DB}
	type RequestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
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

		// Encrypt the Password
		pwBytes, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), 14)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// Create the user record
		_, err = userRepo.Insert(models.User{
			Email:    requestBody.Email,
			Password: string(pwBytes),
		})
		if err != nil {
			pqErr, ok := err.(*pq.Error)
			if ok {
				if pqErr.Code.Name() == "unique_violation" && pqErr.Constraint == "users_email_key" {
					resp.SetResult(http.StatusConflict, nil).
						WithErrorDetails("Email is already in use by another user")
					return
				}
			}
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, nil)
	}
}
