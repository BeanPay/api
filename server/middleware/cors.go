package middleware

import (
	"net/http"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Headers", "Authorization,Keep-Alive,User-Agent,Cache-Control,Content-Type")
			w.WriteHeader(200)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
