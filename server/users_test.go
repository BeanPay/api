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

type CreateUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *CreateUserBody) Read(p []byte) (n int, err error) {
	b, _ := json.Marshal(c)
	return bytes.NewReader(b).Read(p)
}

func TestCreateUser(t *testing.T) {
	// Prepare the Server
	server := NewTestServer()
	defer server.Shutdown()

	// Test that we throw an error with a misformatted request
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("-"))
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusBadRequest,
			StatusText: http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{
				"Failed to parse the request body.",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we are requiring the Email & Password fields
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{},
	)
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusBadRequest,
			StatusText: http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{
				"Email is a required field",
				"Password is a required field",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we are only accepting valid email addresses
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{
			Email:    "invalid-email",
			Password: "some-password",
		},
	)
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusBadRequest,
			StatusText: http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{
				"Email must be a valid email address",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that passwords must be at least 8 characters
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{
			Email:    "name@example.com",
			Password: "1234567",
		},
	)
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusBadRequest,
			StatusText: http.StatusText(http.StatusBadRequest),
			ErrorDetails: &[]string{
				"Password must be at least 8 characters in length",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test the Successful creation of a user
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{
			Email:    "name@example.com",
			Password: "some-password",
		},
	)
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: 200,
			StatusText: "OK",
			Result:     nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that we cannot create the same user twice
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{
			Email:    "name@example.com",
			Password: "some-password",
		},
	)
	server.createUser()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode: http.StatusConflict,
			StatusText: http.StatusText(http.StatusConflict),
			ErrorDetails: &[]string{
				"Email is already in use by another user",
			},
			Result: nil,
		},
		response.Parse(recorder.Result().Body),
	)

}
