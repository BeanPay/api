package server

import (
	"github.com/generalledger/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
