package server

import (
	"bytes"
	"encoding/json"
	"github.com/generalledger/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type CreateUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *CreateUserBody) Read(p []byte) (n int, err error) {
	b, _ := json.Marshal(c)
	return bytes.NewReader(b).Read(p)
}

func TestCreateUserSuccess(t *testing.T) {
	server := NewTestServer()
	defer server.Shutdown()

	// Send Request
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users",
		&CreateUserBody{
			Email:    "name@example.com",
			Password: "some-password",
		},
	)
	server.createUser()(recorder, req)

	// Test
	assert.Equal(t,
		response.Response{
			StatusCode: 200,
			StatusText: "OK",
			Result:     nil,
		},
		response.Parse(recorder.Result().Body),
	)
}
