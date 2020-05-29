package models

import (
	"github.com/beanpay/api/database"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPaymentRepo(t *testing.T) {
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
	paymentRepo := PaymentRepository{
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

	// Create our first Payment
	dueDate, _ := time.Parse("2006-01-02", "2020-05-28")
	firstPayment := &Payment{
		BillId:    firstBill.Id,
		DueDate:   dueDate,
		TotalPaid: 100.50,
	}
	err = paymentRepo.Insert(firstPayment)
	assert.Nil(t, err)

	// Ensure we can't pay the same bill twice
	err = paymentRepo.Insert(firstPayment)
	assert.NotNil(t, err)

	// Create a second payment
	dueDate, _ = time.Parse("2006-01-02", "2020-06-28")
	secondPayment := &Payment{
		BillId:    firstBill.Id,
		DueDate:   dueDate,
		TotalPaid: 100.50,
	}
	err = paymentRepo.Insert(secondPayment)
	assert.Nil(t, err)

	// Fetch a via invalid ID
	_, err = paymentRepo.FetchByID("invalid-id")
	assert.NotNil(t, err)

	// Fetch a payment via ID
	fetchedPayment, err := paymentRepo.FetchByID(firstPayment.Id)
	assert.Nil(t, err)
	assert.Equal(t, firstPayment, fetchedPayment)

	// Fetch May 2020 Payments
	from, _ := time.Parse("2006-01-02", "2020-05-01")
	to, _ := time.Parse("2006-01-02", "2020-06-01")
	payments, err := paymentRepo.FetchAllUserPayments(newUser.Id, from, to)
	assert.Nil(t, err)
	assert.Equal(t, payments[0], firstPayment)

	// Fetch payments via invalid ID
	_, err = paymentRepo.FetchAllUserPayments("invalid-user-id", from, to)
	assert.NotNil(t, err)

	// Delete a Payment
	err = paymentRepo.Delete(firstPayment)
	assert.Nil(t, err)

	// Ensure we cannot delete it again
	err = paymentRepo.Delete(firstPayment)
	assert.NotNil(t, err)

	// Ensure we cannot delete a payment with an invalid ID
	firstPayment.Id = "invalid-payment-id"
	err = paymentRepo.Delete(firstPayment)
	assert.NotNil(t, err)
}
