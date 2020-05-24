package server

import (
	"encoding/json"
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"net/http"
	"time"
)

func (s *Server) fetchBills() http.HandlerFunc {
	billRepo := models.BillRepository{DB: s.DB}
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()

		// Pull out our JWT Claims
		claims, ok := r.Context().Value("jwtClaims").(jwt.Claims)
		if !ok {
			resp.SetResult(http.StatusUnauthorized, nil)
			return
		}

		// Fetch the bills
		bills, err := billRepo.FetchAllUserBills(claims.UserID)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, bills)
	}
}

func (s *Server) createBill() http.HandlerFunc {
	billRepo := models.BillRepository{DB: s.DB}
	type RequestBody struct {
		Name              string  `json:"name" validate:"required"`
		PaymentURL        string  `json:"payment_url" validate:"required,url"`
		Frequency         string  `json:"frequency" validate:"required,oneof=monthly quarterly biannually annually"`
		EstimatedTotalDue float64 `json:"estimated_total_due" validate:"required,gte=0"`
		FirstDueDate      string  `json:"first_due_date" validate:"required,datetime=2006-01-02"`
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
		firstDueDate, err := time.Parse("2006-01-02", requestBody.FirstDueDate)
		if err != nil {
			// We just validated this would be in the right format, so if this
			// error happens something is wrong with our Validator internals.
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// Create a new Bill Record
		newBill := &models.Bill{
			UserId:            claims.UserID,
			Name:              requestBody.Name,
			PaymentURL:        requestBody.PaymentURL,
			Frequency:         requestBody.Frequency,
			EstimatedTotalDue: requestBody.EstimatedTotalDue,
			FirstDueDate:      firstDueDate,
		}
		err = billRepo.Insert(newBill)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, newBill)
	}
}
