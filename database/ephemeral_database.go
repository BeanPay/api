package database

import (
	"database/sql"
	"fmt"
	"github.com/lithammer/shortuuid/v3"
	"os"
	"strings"
)

// NewTestDatabase returns an *EphemeralDatabase spun up
// against our TEST_POSTGRES environment variables, which
// centralizes this config & simplifies our test code.
func NewTestEphemeralDatabase(config Config) (*EphemeralDatabase, error) {
	ephemeralDatabase, err := NewEphemeralDatabase(
		ConnectionInfo{
			Host:         os.Getenv("TEST_POSTGRES_HOST"),
			Port:         os.Getenv("TEST_POSTGRES_PORT"),
			User:         os.Getenv("TEST_POSTGRES_USER"),
			Password:     os.Getenv("TEST_POSTGRES_PASSWORD"),
			DatabaseName: os.Getenv("TEST_POSTGRES_DB"),
			SSLMode:      os.Getenv("TEST_POSTGRES_SSL_MODE"),
		},
		config,
	)
	if err != nil {
		return nil, err
	}
	return ephemeralDatabase, nil
}

func NewEphemeralDatabase(bootstrapConnection ConnectionInfo, config Config) (*EphemeralDatabase, error) {
	// Get a connection to the primary test database, which is only
	// to be used to bootstrap and terminate our ephemeral database
	// connection.
	bootstrapDB, err := sql.Open("postgres", bootstrapConnection.ToURI())
	if err != nil {
		return nil, err
	}
	err = bootstrapDB.Ping()
	if err != nil {
		return nil, err
	}

	// Create the actual database.
	//
	// Can't use the proper injection-safe version of this query here.
	// This is safe in this scenario, as the database name here is far
	// from being in the path of user input.
	//
	// https://github.com/lib/pq/issues/694
	// > This is a limitation of postgres, not lib/pq. Only some types
	// of statements support parameters. Generally anything that
	// modifies schemas doesn't support them.
	dbName := "ephemeral_db_" + strings.ToLower(shortuuid.New())
	_, err = bootstrapDB.Exec(fmt.Sprintf("CREATE DATABASE %v;", dbName))
	if err != nil {
		return nil, err
	}

	// Open the connection to the new Database
	newConnInfo := bootstrapConnection
	newConnInfo.DatabaseName = dbName
	newConnection, err := NewConnection(newConnInfo.ToURI(), config)
	if err != nil {
		if newConnection != nil {
			// The DB was setup, but there was an issue with something else
			// (likely the migration process). We want to report this to the
			// consumer of this function as an error, and not provide them
			// with a database. We need to silently clean up the side-effects
			// from what just happened, and wipe out the actual database under
			// the hood that was just created.
			(&EphemeralDatabase{
				databaseName:        dbName,
				connection:          newConnection,
				bootstrapConnection: bootstrapDB,
			}).Terminate()
		}
		return nil, err
	}

	// OK
	return &EphemeralDatabase{
		bootstrapConnection: bootstrapDB,
		databaseName:        dbName,
		connection:          newConnection,
	}, nil
}

// EphemeralDatabase represents
type EphemeralDatabase struct {
	databaseName        string
	bootstrapConnection *sql.DB
	connection          *sql.DB
}

// Terminate entirely deletes the ephemeral database, and also closes out the connection.
func (e *EphemeralDatabase) Terminate() error {
	// Close the connection to the Ephemeral Database
	err := e.connection.Close()
	if err != nil {
		return err
	}

	// Boot all connections off the database so it can be deleted.
	_, err = e.bootstrapConnection.Exec(fmt.Sprintf("REVOKE CONNECT ON DATABASE %v FROM public;", e.databaseName))
	if err != nil {
		return err
	}
	_, err = e.bootstrapConnection.Exec(
		fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%v';`,
			e.databaseName,
		),
	)
	if err != nil {
		return err
	}

	// Delete the Database
	_, err = e.bootstrapConnection.Exec(fmt.Sprintf("DROP DATABASE %v;", e.databaseName))
	if err != nil {
		return err
	}

	// Close out the bootstrap connection
	err = e.bootstrapConnection.Close()
	return err
}

// Connection returns a reference to the connection to the ephemeral database
func (e *EphemeralDatabase) Connection() *sql.DB {
	return e.connection
}
