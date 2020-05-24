package middleware

import (
	"context"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"net/http"
	"strings"
)

// GetRequireAuthMiddleware returns the middleware function for Authenticate.
// The reason why we have split this out is because there is an
// external dependency (jwt.JwtSignatory) that we don't want to
// have to pass in every time we use this middleware.
func GetRequireAuthMiddleware(jwtSignatory *jwt.JwtSignatory) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			resp := response.New(w)

			// Parse & validate the token
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			claims, err := jwtSignatory.ParseToken(token)
			if err != nil {
				resp.SetResult(http.StatusUnauthorized, nil).Output()
				return
			}

			// Serve next with Claims context
			ctx := context.WithValue(r.Context(), "jwtClaims", *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
