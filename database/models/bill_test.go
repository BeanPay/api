package models

import (
	"github.com/beanpay/api/database"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBillRepo(t *testing.T) {
	// Create a Database for testing
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
	billRepo := BillRepository{
		DB: ephemeralDatabase.Connection(),
	}

	// Create a user so we can create some bills
	newUser := &User{
		Email:    "some-email@example.com",
		Password: "some-password",
	}
	err = userRepo.Insert(newUser)
	assert.Nil(t, err)
	assert.NotEqual(t, "", newUser.Id)

	// Create our first bill
	firstBill := &Bill{
		UserId:            newUser.Id,
		Name:              "First Bill",
		PaymentURL:        "https://example.com",
		Frequency:         "monthly",
		EstimatedTotalDue: 100.25,
		FirstDueDate:      time.Now(),
	}
	err = billRepo.Insert(firstBill)
	assert.Nil(t, err)

	// Create our second bill
	secondBill := &Bill{
		UserId:            newUser.Id,
		Name:              "Second Bill",
		PaymentURL:        "https://example.com",
		Frequency:         "monthly",
		EstimatedTotalDue: 200.25,
		FirstDueDate:      time.Now(),
	}
	err = billRepo.Insert(secondBill)
	assert.Nil(t, err)

	// Test that FetchAllUserBills returns two bills
	allBills, err := billRepo.FetchAllUserBills(newUser.Id)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allBills))
	assert.Equal(t, firstBill.Name, allBills[0].Name)

	// Update the First Bill
	firstBill.Name = "First Bill (Updated)"
	err = billRepo.Update(firstBill)
	assert.Nil(t, err)

	// Fetch the first bill by it's ID to ensure it's been updated in the DB
	fetchedBill, err := billRepo.FetchByID(firstBill.Id)
	assert.Nil(t, err)
	assert.Equal(t, "First Bill (Updated)", fetchedBill.Name)

	// Delete the fetched bill, ensure it can't be fetched
	err = billRepo.Delete(fetchedBill)
	assert.Nil(t, err)
	bill, err := billRepo.FetchByID(fetchedBill.Id)
	assert.Nil(t, bill)
	assert.NotNil(t, err)

	// Ensure we can't delete it again
	err = billRepo.Delete(fetchedBill)
	assert.NotNil(t, err)

	// Try to delete a bill with an invalid ID
	err = billRepo.Delete(&Bill{Id: "not-a-uuid"})
	assert.NotNil(t, err)

	// Test out the fetch method error handling for a invalid query
	bills, err := billRepo.fetch("SELECT *;")
	assert.NotNil(t, err)
	assert.Nil(t, bills)

	// Test out the fetch method error handling for a param # mismatch
	bills, err = billRepo.fetch("SELECT id, name FROM bills;")
	assert.NotNil(t, err)
	assert.Nil(t, bills)
}
