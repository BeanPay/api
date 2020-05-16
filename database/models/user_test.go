package models

import (
	"github.com/beanpay/api/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepo(t *testing.T) {
	// Create a UserRepo
	ephemeralDatabase, err := database.NewTestEphemeralDatabase(
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
	newUser := &User{
		Email:    "some-email@example.com",
		Password: "some-password",
	}
	err = userRepo.Insert(newUser)
	assert.Nil(t, err)
	assert.NotEqual(t, "", newUser.Id)

	// Try to create the same user to ensure it's not created
	err = userRepo.Insert(newUser)
	assert.NotNil(t, err)

	// Fetch the user by their ID
	fetchedByIdUser, err := userRepo.FetchByID(newUser.Id)
	assert.Nil(t, err)
	assert.Equal(t, newUser.Email, fetchedByIdUser.Email)

	// Fetch the user by their email address
	fetchedByEmailUser, err := userRepo.FetchByEmail(newUser.Email)
	assert.Nil(t, err)
	assert.Equal(t, newUser.Id, fetchedByEmailUser.Id)

	// Fetch a user by a non-existing ID
	failedFetchIdUser, err := userRepo.FetchByID("fake-id")
	assert.NotNil(t, err)
	assert.Nil(t, failedFetchIdUser)

	// Fetch a user by a non-existing Email
	failedFetchEmailUser, err := userRepo.FetchByEmail("fake-email")
	assert.NotNil(t, err)
	assert.Nil(t, failedFetchEmailUser)

	// Update the users email
	newUser.Email = "new-email@example.com"
	err = userRepo.Update(newUser)
	assert.Nil(t, err)

	// Fetch the user by their new address
	fetchedUpdatedUser, err := userRepo.FetchByEmail("new-email@example.com")
	assert.Nil(t, err)
	assert.Equal(t, newUser.Id, fetchedUpdatedUser.Id)

	// Delete it
	err = userRepo.Delete(newUser)
	assert.Nil(t, err)

	// Try to delete it again to ensure there's an error
	err = userRepo.Delete(newUser)
	assert.NotNil(t, err)

	// Verify we can no longer fetch this user
	fetchedDeletedUser, err := userRepo.FetchByID(newUser.Id)
	assert.Nil(t, fetchedDeletedUser)
	assert.NotNil(t, err)
}
