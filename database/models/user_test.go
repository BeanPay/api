package models

import (
	"github.com/beanpay/api/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

var fixtureUser = User{
	Email:    "some-email@example.com",
	Password: "some-password",
}

func TestUserRepo(t *testing.T) {
	// Create a UserRepo
	ephemeralDatabase, err := database.NewTestDatabase(
		database.Config{
			MigrationsDir: "../migrations",
		},
	)
	assert.Nil(t, err)
	defer ephemeralDatabase.Terminate()
	userRepo := UserRepository{
		DB: ephemeralDatabase.Connection(),
	}

	// Create a user
	createdUser, err := userRepo.Insert(fixtureUser)
	assert.Nil(t, err)
	assert.Equal(t, fixtureUser.Email, createdUser.Email)

	// Try to create the same user to ensure it's not created
	duplicateUser, err := userRepo.Insert(fixtureUser)
	assert.Nil(t, duplicateUser)
	assert.NotNil(t, err)

	// Fetch the user by their ID
	fetchedByIdUser, err := userRepo.FetchByID(createdUser.Id)
	assert.Nil(t, err)
	assert.Equal(t, createdUser.Email, fetchedByIdUser.Email)

	// Fetch the user by their email address
	fetchedByEmailUser, err := userRepo.FetchByEmail(createdUser.Email)
	assert.Nil(t, err)
	assert.Equal(t, createdUser.Id, fetchedByEmailUser.Id)

	// Fetch a user by a non-existing ID
	failedFetchIdUser, err := userRepo.FetchByID("fake-id")
	assert.NotNil(t, err)
	assert.Nil(t, failedFetchIdUser)

	// Fetch a user by a non-existing Email
	failedFetchEmailUser, err := userRepo.FetchByEmail("fake-email")
	assert.NotNil(t, err)
	assert.Nil(t, failedFetchEmailUser)
}
