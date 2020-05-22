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

type AuthBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *AuthBody) Read(p []byte) (n int, err error) {
	b, _ := json.Marshal(l)
	return bytes.NewReader(b).Read(p)
}

const (
	realUserEmail    = "user@domain.com"
	realUserPassword = "some-great-password"
)

func TestAuthRefresh(t *testing.T) {
	// Prepare the Server
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()

	// Create a User for testing
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users",
		&AuthBody{
			Email:    realUserEmail,
			Password: realUserPassword,
		},
	)
	server.createUser()(recorder, req)
	resp := response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login the user to get a Refresh Token
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{
			Email:    realUserEmail,
			Password: realUserPassword,
		},
	)
	server.login()(recorder, req)
	resp = response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	hasRefreshCookie := false
	firstRefreshToken := ""
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "refresh_token" && cookie.Value != "" {
			hasRefreshCookie = true
			firstRefreshToken = cookie.Value
		}
	}
	assert.True(t, hasRefreshCookie)

	// Refresh our Auth
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: firstRefreshToken,
	})
	server.authRefresh()(recorder, req)
	resp = response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.Result.(map[string]interface{})["access_token"])
	assert.NotNil(t, resp.Result.(map[string]interface{})["access_token_expiration"])
	secondRefreshToken := ""
	hasRefreshCookie = false
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "refresh_token" && cookie.Value != "" {
			hasRefreshCookie = true
			secondRefreshToken = cookie.Value
		}
	}
	assert.True(t, hasRefreshCookie)

	// Test with an invalid refresh token
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid-refresh-token",
	})
	server.authRefresh()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test without a refresh token cookie set
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	server.authRefresh()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test with the first refresh token in the chain, this should trigger
	// the entire chain to be wiped as this token has already been used.
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: firstRefreshToken,
	})
	server.authRefresh()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test that the entire chain was wiped in the last request above.
	// If the chain was still intact, this request below will be successful,
	// but because the last request triggered the chain to be wiped we
	// should actually expect an error.
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: secondRefreshToken,
	})
	server.authRefresh()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)
}

func TestLoginUser(t *testing.T) {
	// Prepare the Server
	server, err := NewTestServer()
	assert.Nil(t, err)
	defer server.Shutdown()

	// Create a User for testing
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users",
		&AuthBody{
			Email:    realUserEmail,
			Password: realUserPassword,
		},
	)
	server.createUser()(recorder, req)
	resp := response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test that we throw an error with a misformatted request
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("-"))
	server.login()(recorder, req)
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
	// Test that we are requiring the Email & Password fields
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{},
	)
	server.login()(recorder, req)
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
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{
			Email:    "some-invalid-email",
			Password: "some-password",
		},
	)
	server.login()(recorder, req)
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

	// Test a failed login of a user where the user does not exist
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{
			Email:    "hello@example.com",
			Password: "some-password",
		},
	)
	server.login()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test a failed login of a user where the password is incorrect
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{
			Email:    realUserEmail,
			Password: "some-invalid-password",
		},
	)
	server.login()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusUnauthorized,
			StatusText:   http.StatusText(http.StatusUnauthorized),
			ErrorDetails: nil,
			Result:       nil,
		},
		response.Parse(recorder.Result().Body),
	)

	// Test the Successful login of a user
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login",
		&AuthBody{
			Email:    realUserEmail,
			Password: realUserPassword,
		},
	)
	server.login()(recorder, req)
	resp = response.Parse(recorder.Result().Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.Result.(map[string]interface{})["access_token"])
	assert.NotNil(t, resp.Result.(map[string]interface{})["access_token_expiration"])
	hasRefreshCookie := false
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == "refresh_token" && cookie.Value != "" {
			hasRefreshCookie = true
		}
	}
	assert.True(t, hasRefreshCookie)
}
