package server

import (
	"encoding/json"
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"net/http"
	"strings"
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

func (s *Server) updateBill() http.HandlerFunc {
	billRepo := models.BillRepository{DB: s.DB}
	type RequestBody struct {
		Name              string  `json:"name" validate:"omitempty"`
		PaymentURL        string  `json:"payment_url" validate:"omitempty,url"`
		Frequency         string  `json:"frequency" validate:"omitempty,oneof=monthly quarterly biannually annually"`
		EstimatedTotalDue float64 `json:"estimated_total_due" validate:"omitempty,gte=0"`
		FirstDueDate      string  `json:"first_due_date" validate:"omitempty,datetime=2006-01-02"`
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

		// Fetch the Bill
		billId := strings.Split(r.URL.Path, "/")[2]
		bill, err := billRepo.FetchByID(billId)
		if err != nil {
			resp.SetResult(http.StatusNotFound, nil)
			return
		}

		// Verify the authorized user owns the bill
		if bill.UserId != claims.UserID {
			resp.SetResult(http.StatusForbidden, nil)
			return
		}

		//  Parse & Validate the Body
		var requestBody RequestBody
		err = json.NewDecoder(r.Body).Decode(&requestBody)
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

		// Update the Bill
		if requestBody.Name != "" {
			bill.Name = requestBody.Name
		}
		if requestBody.PaymentURL != "" {
			bill.PaymentURL = requestBody.PaymentURL
		}
		if requestBody.Frequency != "" {
			bill.Frequency = requestBody.Frequency
		}
		if requestBody.EstimatedTotalDue != 0 {
			bill.EstimatedTotalDue = requestBody.EstimatedTotalDue
		}
		if requestBody.FirstDueDate != "" {
			firstDueDate, err := time.Parse("2006-01-02", requestBody.FirstDueDate)
			if err != nil {
				// We just validated this would be in the right format, so if this
				// error happens something is wrong with our Validator internals.
				resp.SetResult(http.StatusInternalServerError, nil)
				return
			}
			bill.FirstDueDate = firstDueDate
		}
		err = billRepo.Update(bill)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, bill)
	}
}

func (s *Server) deleteBill() http.HandlerFunc {
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

		// Fetch the Bill
		billId := strings.Split(r.URL.Path, "/")[2]
		bill, err := billRepo.FetchByID(billId)
		if err != nil {
			resp.SetResult(http.StatusNotFound, nil)
			return
		}

		// Verify the authorized user owns the bill
		if bill.UserId != claims.UserID {
			resp.SetResult(http.StatusForbidden, nil)
			return
		}

		// Delete the bill
		err = billRepo.Delete(bill)
		if err != nil {
			resp.SetResult(http.StatusInternalServerError, nil)
			return
		}

		// OK
		resp.SetResult(http.StatusOK, nil)
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
