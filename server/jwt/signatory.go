package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type JwtSignatory struct {
	SigningKey []byte
}

func (s *JwtSignatory) GenerateSignedToken(userID string, expiration time.Time) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&Claims{
			UserID: userID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expiration.Unix(),
			},
		},
	)
	return token.SignedString(s.SigningKey)
}

func (s *JwtSignatory) ParseToken(token string) (*Claims, error) {
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(jwtToken *jwt.Token) (interface{}, error) {
			if jwtToken.Method.Alg() != "HS256" {
				msg := fmt.Errorf("Unexpected signing method: %v", jwtToken.Header["alg"])
				return nil, msg
			}
			return s.SigningKey, nil
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
