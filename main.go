package main

import (
	"github.com/beanpay/api/database"
	"github.com/beanpay/api/server"
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
		Port:      os.Getenv("PORT"),
		Router:    httprouter.New(),
		Validator: validator.New(),
		DB:        db,
	}
	server.Start()
}
