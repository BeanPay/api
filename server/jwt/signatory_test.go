package jwt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwtSignatory(t *testing.T) {
	// Create a Signatory
	signatoryOne := &JwtSignatory{
		SigningKey: []byte("sig-one"),
	}

	// Generate a Signed Token
	jwt, err := signatoryOne.GenerateSignedToken("some-user-id", time.Now().Add(time.Millisecond*10))
	assert.Nil(t, err)

	// Parse & Validated the generated token
	claims, err := signatoryOne.ParseToken(jwt)
	assert.Nil(t, err)
	assert.Equal(t, "some-user-id", claims.UserID)

	// Let the token expire, Parse & Validate it again
	time.Sleep(time.Second * 1)
	claims, err = signatoryOne.ParseToken(jwt)
	assert.Nil(t, claims)
	assert.NotNil(t, err)
	assert.Equal(t, "token is expired by 1s", err.Error())

	// Verify that the proper algorithm must be used
	_, err = signatoryOne.ParseToken("eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.e30.vf3TqzF5aHxl9AHEBKqyWF2GINAWx2UUabmm2-_3qJNj6wwMHCde04OteqQi8UIq7b330WbeW7NgXKe6oKzplQ")
	assert.NotNil(t, err)
	assert.Equal(t, "Unexpected signing method: HS512", err.Error())

	// Create a Second Signatory
	signatoryTwo := &JwtSignatory{
		SigningKey: []byte("sig-two"),
	}

	// Sign a JWT
	jwt, err = signatoryTwo.GenerateSignedToken("some-user-id", time.Now().Add(time.Millisecond*10))
	assert.Nil(t, err)
	assert.NotEqual(t, "", jwt)

	// Verify that signatoryTwo can Parse the token, but signatoryOne cannot
	_, err = signatoryTwo.ParseToken(jwt)
	assert.Nil(t, err)
	_, err = signatoryOne.ParseToken(jwt)
	assert.NotNil(t, err)
	assert.Equal(t, "signature is invalid", err.Error())
}
