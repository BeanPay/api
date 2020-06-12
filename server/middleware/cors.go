package middleware

import (
	"net/http"
	"os"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", os.Getenv("APP_URL"))
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Headers", "Authorization,Keep-Alive,User-Agent,Cache-Control,Content-Type")
			w.WriteHeader(200)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
