package models

import (
	"github.com/beanpay/api/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRefreshTokenRepo(t *testing.T) {
	// Create an Ephemeral DB to run this test suite
	ephemeralDatabase, err := database.NewTestEphemeralDatabase(
		database.Config{
			MigrationsDir: "../migrations",
		},
	)
	assert.Nil(t, err)
	defer ephemeralDatabase.Terminate()

	// Create a UserRepository so we can create a user to
	// bootstrap some data for these test to work
	userRepo := UserRepository{
		DB: ephemeralDatabase.Connection(),
	}

	// Create a RefreshTokenRepository
	refreshTokenRepo := RefreshTokenRepository{
		DB: ephemeralDatabase.Connection(),
	}

	// Create a sample user so we can create RefreshTokens for them
	sampleUser := &User{
		Email:    "some-email@example.com",
		Password: "some-password",
	}
	err = userRepo.Insert(sampleUser)
	assert.Nil(t, err)
	assert.NotEqual(t, "", sampleUser.Id)

	// Create a chainID that we can use for testing refreshTokens
	const testingChainID = "14fc415c-7848-4363-967f-1a39c6031a86"

	// Create a first RefreshToken
	firstRefreshToken := &RefreshToken{
		ChainId: testingChainID,
		UserId:  sampleUser.Id,
	}
	err = refreshTokenRepo.Insert(firstRefreshToken)
	assert.Nil(t, err)
	assert.NotEqual(t, "", firstRefreshToken.Id)

	// Fetch the first one to ensure it was inserted & can be fetched
	firstFetch, err := refreshTokenRepo.FetchByID(firstRefreshToken.Id)
	assert.Nil(t, err)
	assert.Equal(t, firstRefreshToken.Id, firstFetch.Id)

	// Create a Second RefreshToken
	secondRefreshToken := &RefreshToken{
		ChainId: testingChainID,
		UserId:  sampleUser.Id,
	}
	err = refreshTokenRepo.Insert(secondRefreshToken)
	assert.Nil(t, err)
	assert.NotEqual(t, "", secondRefreshToken.Id)

	// Fetch the most recent in the chain
	mostRecentToken, err := refreshTokenRepo.FetchMostRecentInChain(testingChainID)
	assert.Nil(t, err)
	assert.Equal(t, secondRefreshToken.Id, mostRecentToken.Id)

	// Wipe the chain
	err = refreshTokenRepo.DeleteChain(testingChainID)
	assert.Nil(t, err)

	// Wipe the chain again to ensure an error is thrown
	err = refreshTokenRepo.DeleteChain(testingChainID)
	assert.NotNil(t, err)

	// Delete by a non-uuid to ensure an error is thrown
	err = refreshTokenRepo.DeleteChain("non-uuid")
	assert.NotNil(t, err)

	// Fetch the most recent in the chain to ensure that the chain was properly wiped
	mostRecentPostDelete, err := refreshTokenRepo.FetchMostRecentInChain(testingChainID)
	assert.NotNil(t, err)
	assert.Nil(t, mostRecentPostDelete)

	// Fetch the first one to ensure it was deleted and cannot be fetched
	firstFetch, err = refreshTokenRepo.FetchByID(firstRefreshToken.Id)
	assert.NotNil(t, err)
	assert.Nil(t, firstFetch)
}
