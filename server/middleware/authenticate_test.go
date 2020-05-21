package middleware

import (
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A simple jwtSignatory to generate JWT Tokens for
// use against our Authenticate middleware.
var jwtSignatory = &jwt.JwtSignatory{
	SigningKey: []byte("some-key"),
}

// A simple http.HandlerFunc which is wrapped in our Authenticate
// middleware that is used in the tests below.
func authenticatedHandler() http.HandlerFunc {
	return Authenticate(
		jwtSignatory,
		func(w http.ResponseWriter, r *http.Request) {
			resp := response.New(w)
			defer resp.Output()
			claims, ok := r.Context().Value("jwtClaims").(jwt.Claims)
			if !ok {
				resp.SetResult(http.StatusInternalServerError, nil)
				return
			}
			resp.SetResult(http.StatusOK, claims.UserID)
		},
	)
}

// TestAuthenticateMiddlewareNoAuth tests that when we are given
// Unauthorized status thrown from the Middleware when there is no
// token passed.
func TestAuthenticateMiddlewareNoAuth(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	authenticatedHandler()(recorder, req)
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

// TestAuthenticateMiddlewareIncorrectAuth tests that we are given a
// Unauthorized status thrown from the Middleware when there is an
// invalid token passed.
func TestAuthenticateMiddlewareIncorrectAuth(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	authenticatedHandler()(recorder, req)
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

// TestAuthenticateMiddlewareSuccess tests the success path of
// our Authenticate middleware.
//
// Additionally, this validates that the jwtClaims are being passed down
// the request chain appropriately.
func TestAuthenticateMiddlewareSuccess(t *testing.T) {
	// Generate a JWT Token signed by the same signatory that our Authenticate middleware uses
	jwtToken, err := jwtSignatory.GenerateSignedToken("some-user-id", time.Now().Add(time.Second*10))
	assert.Nil(t, err)

	// Verify request is OK
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	authenticatedHandler()(recorder, req)
	assert.Equal(t,
		response.Response{
			StatusCode:   http.StatusOK,
			StatusText:   http.StatusText(http.StatusOK),
			ErrorDetails: nil,
			Result:       "some-user-id",
		},
		response.Parse(recorder.Result().Body),
	)
}

// TestAuthenticateMiddlewareIncorrectlySignedToken tests that the middleware
// rejects any JWTs signed by a different Key.
func TestAuthenticateMiddlewareIncorrectlySignedToken(t *testing.T) {
	// Generate a JWT Token signed by a different Signatory, which provides
	// a JWT Token signed by a different Key as compared to what our middleware expects.
	var nefariousJwtSignatory = &jwt.JwtSignatory{
		SigningKey: []byte("incorrect-key"),
	}
	jwtToken, err := nefariousJwtSignatory.GenerateSignedToken("some-user-id", time.Now().Add(time.Second*10))
	assert.Nil(t, err)

	// Verify request is OK
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	authenticatedHandler()(recorder, req)
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
