package server

import (
	"database/sql"
	"github.com/beanpay/api/database"
	"github.com/generalledger/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPingFailure(t *testing.T) {
	// Get a faulty connection to a database
	db, err := sql.Open("postgres",
		database.ConnectionInfo{
			Host:         "localhost",
			Port:         "5555",
			User:         "user",
			Password:     "incorrect-password",
			DatabaseName: "incorrect-database",
			SSLMode:      "require",
		}.ToURI(),
	)
	assert.Nil(t, err)

	// Prepare Server
	server := &Server{
		DB: db,
	}

	// Send Request
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	server.ping()(recorder, req)

	// Test
	parsedResponse := response.Parse(recorder.Result().Body)
	assert.Equal(t, parsedResponse.StatusCode, http.StatusInternalServerError)
}

func TestPingSuccess(t *testing.T) {
	// Get a connection to the database
	db, err := database.NewConnection(
		database.ConnectionInfo{
			Host:         os.Getenv("TEST_POSTGRES_HOST"),
			Port:         os.Getenv("TEST_POSTGRES_PORT"),
			User:         os.Getenv("TEST_POSTGRES_USER"),
			Password:     os.Getenv("TEST_POSTGRES_PASSWORD"),
			DatabaseName: os.Getenv("TEST_POSTGRES_DB"),
			SSLMode:      os.Getenv("TEST_POSTGRES_SSL_MODE"),
		}.ToURI(),
		database.Config{},
	)
	if err != nil {
		panic(err)
	}

	// Prepare Server
	server := &Server{
		DB: db,
	}

	// Send Request
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	server.ping()(recorder, req)

	// Test
	assert.Equal(t,
		response.Response{
			StatusCode: 200,
			StatusText: "OK",
			Result: map[string]interface{}{
				"database_connection": "OK",
			},
		},
		response.Parse(recorder.Result().Body),
	)
}
