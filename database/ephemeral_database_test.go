package database

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEphemeralDatabaseBadConfig(t *testing.T) {
	// Spin up a new Ephemeral Database with a bad config
	_, err := NewTestEphemeralDatabase(
		Config{
			MigrationsDir: "../incorect/path/to/migrations",
		},
	)
	assert.NotNil(t, err)
}

func TestEphemeralDatabaseSuccess(t *testing.T) {
	// Spin up a new Ephemeral Database
	ephemeralDatabase, err := NewTestEphemeralDatabase(
		Config{
			MigrationsDir: "../database/migrations",
		},
	)
	assert.Nil(t, err)
	defer ephemeralDatabase.Terminate()

	// Test that the database works & we can ping it
	err = ephemeralDatabase.Connection().Ping()
	assert.Nil(t, err)

	// Shut it down
	err = ephemeralDatabase.Terminate()
	assert.Nil(t, err)

	// Verify that we can't ping it
	err = ephemeralDatabase.Connection().Ping()
	assert.NotNil(t, err)
}
