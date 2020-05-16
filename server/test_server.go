package server

import (
	"github.com/beanpay/api/database"
	"github.com/beanpay/api/server/validator"
	"os"
)

// TestServer is a utility Server wrapper that is used to simplify
// writing integration tests. This really serves two purposes:
//
// 1. Plug all of the dependencies up in one place
// 2. Spin up an EphemeralDatabase that can be Terminated w/ a Shutdown function
type TestServer struct {
	Server
	EphemeralDatabase *database.EphemeralDatabase
}

func NewTestServer() *TestServer {
	// Get a connection to the database
	ephemeralDatabase, err := database.NewEphemeralDatabase(
		database.ConnectionInfo{
			Host:         os.Getenv("TEST_POSTGRES_HOST"),
			Port:         os.Getenv("TEST_POSTGRES_PORT"),
			User:         os.Getenv("TEST_POSTGRES_USER"),
			Password:     os.Getenv("TEST_POSTGRES_PASSWORD"),
			DatabaseName: os.Getenv("TEST_POSTGRES_DB"),
			SSLMode:      os.Getenv("TEST_POSTGRES_SSL_MODE"),
		},
		database.Config{
			MigrationsDir: "../database/migrations",
		},
	)
	if err != nil {
		panic(err)
	}

	return &TestServer{
		EphemeralDatabase: ephemeralDatabase,
		Server: Server{
			Validator: validator.New(),
			DB:        ephemeralDatabase.Connection(),
		},
	}
}

func (t *TestServer) Shutdown() {
	t.EphemeralDatabase.Terminate()
}
