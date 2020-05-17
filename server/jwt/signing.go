package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
	"time"
)

var jwtSigningKey = []byte(os.Getenv("JWT_SIGNING_KEY"))

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateSignedToken(userID string, expiration time.Time) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&Claims{
			UserID: userID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expiration.Unix(),
			},
		},
	)
	return token.SignedString(jwtSigningKey)
}

func ParseToken(token string) (*Claims, error) {
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(jwtToken *jwt.Token) (interface{}, error) {
			if jwtToken.Method.Alg() != "HS256" {
				msg := fmt.Errorf("Unexpected signing method: %v", jwtToken.Header["alg"])
				return nil, msg
			}
			return jwtSigningKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("Token Invalid")
	}
	return claims, nil
}
