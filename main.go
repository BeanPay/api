package main

import (
	"github.com/beanpay/api/database"
	"github.com/beanpay/api/server"
	"github.com/beanpay/api/server/jwt"
	"github.com/beanpay/api/server/validator"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"os"
)

func main() {
	godotenv.Load()

	db, err := database.NewConnection(
		os.Getenv("POSTGRES_URL"),
		database.Config{
			MigrationsDir: "./database/migrations",
		},
	)
	if err != nil {
		panic(err)
	}

	server := &server.Server{
		Version:   "0.1.1",
		Port:      os.Getenv("PORT"),
		Router:    httprouter.New(),
		Validator: validator.New(),
		JwtSignatory: &jwt.JwtSignatory{
			SigningKey: []byte(os.Getenv("JWT_SIGNING_KEY")),
		},
		DB: db,
	}
	server.Start()
}
