package server

import (
	"bytes"
	"encoding/json"
	"github.com/generalledger/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type PaymentRequestBody struct {
	BillId    string  `json:"bill_id"`
	DueDate   string  `json:"due_date"`
	TotalPaid float64 `json:"total_paid"`
}

func (r *PaymentRequestBody) Read(p []byte) (n int, err error) {
	b, _ := json.Marshal(r)
	return bytes.NewReader(b).Read(p)
}

func TestPaymentFetch(t *testing.T) {
	// Prepare the Server & seed some data
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()
	user1 := server.SeedUser()
	user1Bill := server.SeedBill(user1["id"].(string))
	user1Payment := server.SeedPayment(user1Bill["id"].(string))
	user2 := server.SeedUser()
	user2Bill := server.SeedBill(user2["id"].(string))
	user2Payment := server.SeedPayment(user2Bill["id"].(string))

	// Validate auth is required
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payments", nil)
	server.fetchPayments()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Ensure that we are validating our requests
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/payments?from=invalid&to=invalid", user1["id"].(string), nil)
	server.fetchPayments()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusBadRequest,
			StatusText: http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{
				"From does not match the 2006-01-02 format",
				"To does not match the 2006-01-02 format",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Ensure that user1 only can see their payments
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/payments?from=2000-01-01&to=6000-01-01", user1["id"].(string), nil)
	server.fetchPayments()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result:       []interface{}{user1Payment},
		},
		response.Parse(recorder.Result().Body),
	)

	// Ensure that user2 only can see their payments
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/payments?from=2000-01-01&to=9999-01-01", user2["id"].(string), nil)
	server.fetchPayments()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result:       []interface{}{user2Payment},
		},
		response.Parse(recorder.Result().Body),
	)
}

func TestPaymentDelete(t *testing.T) {
	// Prepare the Server & seed some data
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()
	user1 := server.SeedUser()
	user1Bill := server.SeedBill(user1["id"].(string))
	user1Payment := server.SeedPayment(user1Bill["id"].(string))
	user2 := server.SeedUser()
	user2Bill := server.SeedBill(user2["id"].(string))
	user2Payment := server.SeedPayment(user2Bill["id"].(string))

	// Validate auth is required
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/payments"+user1Payment["id"].(string), nil)
	server.deletePayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Try to delete a payment that doesn't exist
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodDelete, "/payments/invalid-id", user1["id"].(string), nil)
	server.deletePayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusNotFound,
			StatusText:   http.StatusText(http.StatusNotFound),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Try to delete another users payment
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodDelete, "/payments/"+user2Payment["id"].(string), user1["id"].(string), nil)
	server.deletePayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusForbidden,
			StatusText:   http.StatusText(http.StatusForbidden),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we can successfully delete our payment
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodDelete, "/payments/"+user1Payment["id"].(string), user1["id"].(string), nil)
	server.deletePayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)
}

func TestPaymentCreate(t *testing.T) {
	// Prepare the Server & seed some data
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()
	user1 := server.SeedUser()
	user1Bill := server.SeedBill(user1["id"].(string))
	user2 := server.SeedUser()
	user2Bill := server.SeedBill(user2["id"].(string))

	// Validate auth is required
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments", nil)
	server.createPayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we cannot send a request with a misformated request body
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), strings.NewReader("-"))
	server.createPayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusBadRequest,
			StatusText:   http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{"Failed to parse the request body."},
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we are validating our request body
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), &PaymentRequestBody{
		BillId:    user1Bill["id"].(string),
		DueDate:   "invalid-due-date",
		TotalPaid: 19.99,
	})
	server.createPayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusBadRequest,
			StatusText:   http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{"DueDate does not match the 2006-01-02 format"},
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we need a valid Bill ID to create the Payment
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), &PaymentRequestBody{
		BillId:    "invalid-bill-id",
		DueDate:   "2006-01-02",
		TotalPaid: 19.99,
	})
	server.createPayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusBadRequest,
			StatusText:   http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{"There is no Bill with the specified 'BillId'."},
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that users cannot pay other users bills
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), &PaymentRequestBody{
		BillId:    user2Bill["id"].(string),
		DueDate:   "2006-01-02",
		TotalPaid: 19.99,
	})
	server.createPayment()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusForbidden,
			StatusText:   http.StatusText(http.StatusForbidden),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that users can successfully create payments for their bills
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), &PaymentRequestBody{
		BillId:    user1Bill["id"].(string),
		DueDate:   "2006-01-02",
		TotalPaid: 19.99,
	})
	server.createPayment()(recorder, req)
	resp := response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Ensure that users cannot pay the same bill for the same date twice
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPost, "/payments", user1["id"].(string), &PaymentRequestBody{
		BillId:    user1Bill["id"].(string),
		DueDate:   "2006-01-02",
		TotalPaid: 19.99,
	})
	server.createPayment()(recorder, req)
	resp = response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
