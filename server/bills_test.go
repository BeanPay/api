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

type BillRequestBody struct {
	Name              string  `json:"name,omitempty"`
	PaymentURL        string  `json:"payment_url,omitempty"`
	Frequency         string  `json:"frequency,omitempty"`
	EstimatedTotalDue float64 `json:"estimated_total_due,omitempty"`
	FirstDueDate      string  `json:"first_due_date,omitempty"`
}

func (r *BillRequestBody) Read(p []byte) (n int, err error) {
	b, _ := json.Marshal(r)
	return bytes.NewReader(b).Read(p)
}

func TestBillFetch(t *testing.T) {
	// Prepare the Server & seed some data
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()
	user1 := server.SeedUser()
	user2 := server.SeedUser()
	user1Bill1 := server.SeedBill(user1["id"].(string))
	user1Bill2 := server.SeedBill(user1["id"].(string))
	user2Bill1 := server.SeedBill(user2["id"].(string))

	// Validate auth is required
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bills", nil)
	server.fetchBills()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Validate that we error out with a misformatted UUID
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/bills", "some-fake-uuid", nil)
	server.fetchBills()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusInternalServerError,
			StatusText:   http.StatusText(http.StatusInternalServerError),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Validate that Fetch for user1 returns only their bills
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/bills", user1["id"].(string), nil)
	server.fetchBills()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result: []interface{}{
				user1Bill1,
				user1Bill2,
			},
		},
		response.Parse(recorder.Result().Body),
	)

	// Validate that Fetch for user2 returns only their bills
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodGet, "/bills", user2["id"].(string), nil)
	server.fetchBills()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result: []interface{}{
				user2Bill1,
			},
		},
		response.Parse(recorder.Result().Body),
	)
}

func TestBillUpdate(t *testing.T) {
	// Prepare the Server & seed some data
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()
	user1 := server.SeedUser()
	bill1 := server.SeedBill(user1["id"].(string))
	user2 := server.SeedUser()
	bill2 := server.SeedBill(user2["id"].(string))

	// Validate auth is required for [PUT] /bills
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/bills/"+bill1["id"].(string), nil)
	server.updateBill()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Validate that we cannot update a bill that does not exist
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPut, "/bills/fake-bill-id", user1["id"].(string), nil)
	server.updateBill()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusNotFound,
			StatusText:   http.StatusText(http.StatusNotFound),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Validate that we cannot update a bill that does not belong to us
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPut, "/bills/"+bill2["id"].(string), user1["id"].(string), nil)
	server.updateBill()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusForbidden,
			StatusText:   http.StatusText(http.StatusForbidden),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we cannot send a request with a misformated request body
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPut, "/bills/"+bill1["id"].(string), user1["id"].(string), strings.NewReader("-"))
	server.updateBill()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusBadRequest,
			StatusText:   http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{"Failed to parse the request body."},
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we are performing field level validation checks
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPut, "/bills/"+bill1["id"].(string), user1["id"].(string), &BillRequestBody{
		PaymentURL: "not-a-url",
	})
	server.updateBill()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusBadRequest,
			StatusText:   http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{"PaymentURL must be a valid URL"},
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we can successfully update a bill
	recorder = httptest.NewRecorder()
	req = server.NewAuthenticatedRequest(http.MethodPut, "/bills/"+bill1["id"].(string), user1["id"].(string), &BillRequestBody{
		Name:              "Some Bill",
		PaymentURL:        "https://some-url.com",
		Frequency:         "annually",
		EstimatedTotalDue: 29.99,
		FirstDueDate:      "2020-01-01",
	})
	server.updateBill()(recorder, req)
	response := response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
