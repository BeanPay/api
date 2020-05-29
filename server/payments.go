package server

import (
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"net/http"
	"time"
)

func (s *Server) fetchPayments() http.HandlerFunc {
	paymentRepo := models.PaymentRepository{DB: s.DB}
	type RequestParams struct {
		From string `json:"from" validate:"required,datetime=2006-01-02"`
		To   string `json:"to" validate:"required,datetime=2006-01-02"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()

		// Pull out our JWT Claims
		claims, ok := r.Context().Value("jwtClaims").(jwt.Claims)
		if !ok {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Validate the request params
		requestParams := &RequestParams{}
		queryParams := r.URL.Query()
		from, ok := queryParams["from"]
		if ok && (len(from) > 0) {
			requestParams.From = from[0]
		}
		to, ok := queryParams["to"]
		if ok && (len(to) > 0) {
			requestParams.To = to[0]
		}
		messages, err := s.Validator.Validate(requestParams)
		if err != nil {
			resp.SetResult(http.StatusBadRequest, nil).
				WithErrorDetails(messages...)
			return
		}

		// Parse out the dates
		fromDate, fromErr := time.Parse("2006-01-02", requestParams.From)
		toDate, toErr := time.Parse("2006-01-02", requestParams.To)
		if fromErr != nil || toErr != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// Fetch the Payments
		payments, err := paymentRepo.FetchAllUserPayments(claims.UserID, fromDate, toDate)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, payments)
	}
}
