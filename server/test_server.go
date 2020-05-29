package server

import (
	"context"
	"encoding/json"
	"github.com/beanpay/api/database"
	"github.com/beanpay/api/database/models"
	"github.com/beanpay/api/server/jwt"
	"github.com/beanpay/api/server/validator"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func NewTestServer() (*TestServer, error) {
	ephemeralDatabase, err := database.NewTestEphemeralDatabase(
		database.Config{
			MigrationsDir: "../database/migrations",
		},
	)
	if err != nil {
		return nil, err
	}
	return &TestServer{
		EphemeralDatabase: ephemeralDatabase,
		Server: Server{
			Validator: validator.New(),
			DB:        ephemeralDatabase.Connection(),
			JwtSignatory: &jwt.JwtSignatory{
				SigningKey: []byte("test-signing-key"),
			},
		},
	}, nil
}

// TestServer is a utility Server wrapper that is used to simplify
// writing integration tests. This really serves two purposes:
//
// 1. Plug all of the dependencies up in one place
// 2. Spin up an EphemeralDatabase that can be Terminated w/ a Shutdown function
type TestServer struct {
	Server
	EphemeralDatabase *database.EphemeralDatabase
}

// NewAuthenticatedRequest generates a request that's authenticated as if the
// user with userId has actually gone through a proper authentication flow.
func (t *TestServer) NewAuthenticatedRequest(method, target, userId string, body io.Reader) *http.Request {
	// Generate a token valid for 1 second
	token, err := t.JwtSignatory.GenerateSignedToken(userId, time.Now().Add(time.Second*1))
	if err != nil {
		panic(err)
	}
	claims, err := t.JwtSignatory.ParseToken(token)
	if err != nil {
		panic(err)
	}

	// Create the request and add the context
	req := httptest.NewRequest(method, target, body)
	ctx := req.Context()
	ctx = context.WithValue(ctx, "jwtClaims", *claims)
	return req.WithContext(ctx)
}

// Seeds a random User into our TestServer's database.
// This returns a map[string]interface{} for ease of result comparison,
// as the output of a parsed HTTPRecorder's result will be untyped.
func (t *TestServer) SeedUser() map[string]interface{} {
	// Create a User in our Database
	userRepo := models.UserRepository{DB: t.EphemeralDatabase.Connection()}
	user := &models.User{
		Email:    uuid.NewV4().String() + "@example.com",
		Password: uuid.NewV4().String(),
	}
	err := userRepo.Insert(user)
	if err != nil {
		panic(err)
	}

	// Convert to a map[string]interface{}
	var result map[string]interface{}
	b, _ := json.Marshal(user)
	json.Unmarshal(b, &result)
	return result
}

// Seeds a random Bill into our TestServer's database.
// This returns a map[string]interface{} for ease of result comparison,
// as the output of a parsed HTTPRecorder's result will be untyped.
func (t *TestServer) SeedBill(userId string) map[string]interface{} {
	// Create a Bill in our Database
	billRepo := models.BillRepository{DB: t.EphemeralDatabase.Connection()}
	bill := &models.Bill{
		UserId:            userId,
		Name:              uuid.NewV4().String(),
		PaymentURL:        "https://" + uuid.NewV4().String() + ".com",
		Frequency:         "monthly",
		EstimatedTotalDue: 19.99,
		FirstDueDate:      time.Now().Add(time.Hour * 24 * 10),
	}
	err := billRepo.Insert(bill)
	if err != nil {
		panic(err)
	}

	// Convert to a map[string]interface{}
	var result map[string]interface{}
	b, _ := json.Marshal(bill)
	json.Unmarshal(b, &result)
	return result
}

// Seeds a random Payment into our TestServer's database.
// This returns a map[string]interface{} for ease of result comparison,
// as the output of a parsed HTTPRecorder's result will be untyped.
func (t *TestServer) SeedPayment(billId string) map[string]interface{} {
	// Create a Payment in our Database
	paymentRepo := models.PaymentRepository{DB: t.EphemeralDatabase.Connection()}
	payment := &models.Payment{
		BillId:    billId,
		DueDate:   time.Now().Add(time.Hour * 24 * 10),
		TotalPaid: 19.99,
	}
	err := paymentRepo.Insert(payment)
	if err != nil {
		panic(err)
	}

	// Convert to a map[string]interface{}
	var result map[string]interface{}
	b, _ := json.Marshal(payment)
	json.Unmarshal(b, &result)
	return result
}

func (t *TestServer) Shutdown() {
	t.EphemeralDatabase.Terminate()
}
