package middleware

import (
	"context"
	"github.com/beanpay/api/server/jwt"
	"github.com/generalledger/response"
	"net/http"
	"strings"
)

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)

		// Parse & validate the token
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims, err := jwt.ParseToken(token)
		if err != nil {
			resp.SetResult(http.StatusUnauthorized, nil).Output()
			return
		}

		// Serve next with Claims context
		ctx := context.WithValue(r.Context(), "jwtClaims", *claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
