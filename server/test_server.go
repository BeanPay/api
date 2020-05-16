package server

import (
	"github.com/beanpay/api/database"
	"github.com/beanpay/api/server/validator"
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

func NewTestServer() (*TestServer, error) {
	ephemeralDatabase, err := database.NewTestEphemeralDatabase(
		database.Config{
			MigrationsDir: "../database/migrations",
		},
	)
	if err != nil {
		return nil, err
	}
	return &TestServer{
		EphemeralDatabase: ephemeralDatabase,
		Server: Server{
			Validator: validator.New(),
			DB:        ephemeralDatabase.Connection(),
		},
	}, nil
}

func (t *TestServer) Shutdown() {
	t.EphemeralDatabase.Terminate()
}
