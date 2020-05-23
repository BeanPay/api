package validator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type RequestBody struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"min=8"`
}

func TestValidator(t *testing.T) {
	v := New()

	// Test that errors are thrown when failing against the validator
	messages, err := v.Validate(RequestBody{
		Email:    "not an email",
		Password: "short",
	})
	assert.NotNil(t, err)
	assert.Equal(t, []string{
		"Email must be a valid email address",
		"Password must be at least 8 characters in length",
	}, messages)

	// Test that errors aren't thrown for a successful validation
	messages, err = v.Validate(RequestBody{
		Email:    "real-email@example.com",
		Password: "a-very-secure-password",
	})
	assert.Nil(t, err)
	assert.Equal(t, []string{}, messages)
}
