package database

import (
	"os"
)

// NewTestDatabase returns an *EphemeralDatabase spun up
// against our TEST_POSTGRES environment variables, which
// centralizes this config & simplifies our test code.
func NewTestDatabase() (*EphemeralDatabase, error) {
	ephemeralDatabase, err := NewEphemeralDatabase(
		ConnectionInfo{
			Host:         os.Getenv("TEST_POSTGRES_HOST"),
			Port:         os.Getenv("TEST_POSTGRES_PORT"),
			User:         os.Getenv("TEST_POSTGRES_USER"),
			Password:     os.Getenv("TEST_POSTGRES_PASSWORD"),
			DatabaseName: os.Getenv("TEST_POSTGRES_DB"),
			SSLMode:      os.Getenv("TEST_POSTGRES_SSL_MODE"),
		},
		Config{
			MigrationsDir: "../database/migrations",
		},
	)
	if err != nil {
		return nil, err
	}
	return ephemeralDatabase, nil
}
