package server

import (
	"database/sql"
	"fmt"
	"github.com/beanpay/api/server/jwt"
	"github.com/beanpay/api/server/middleware"
	"github.com/beanpay/api/server/validator"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Server is a struct responsible for managing a *httprouter.Router.
// All HandlerFunc Closures hang off of this struct, so all HandlerFunc's
// have access to the server values.
type Server struct {
	Port         string
	Router       *httprouter.Router
	Validator    validator.Validator
	JwtSignatory *jwt.JwtSignatory
	DB           *sql.DB
}

// registerRoutes is responsible for wiring up all of our HandlerFunc
// to our server's router.
func (s *Server) registerRoutes() {
	requireAuth := middleware.GetRequireAuthMiddleware(s.JwtSignatory)
	s.Router.HandlerFunc(http.MethodGet, "/ping", s.ping())

	// Payments Endpoints
	s.Router.HandlerFunc(http.MethodGet, "/payments", requireAuth(s.fetchPayments()))

	// Bills Endpoints
	s.Router.HandlerFunc(http.MethodGet, "/bills", requireAuth(s.fetchBills()))
	s.Router.HandlerFunc(http.MethodPost, "/bills", requireAuth(s.createBill()))
	s.Router.HandlerFunc(http.MethodPut, "/bills/:id", requireAuth(s.updateBill()))
	s.Router.HandlerFunc(http.MethodDelete, "/bills/:id", requireAuth(s.deleteBill()))

	// Auth Endpoints
	s.Router.HandlerFunc(http.MethodPost, "/users", s.createUser())
	s.Router.HandlerFunc(http.MethodPost, "/auth/login", s.login())
	s.Router.HandlerFunc(http.MethodPost, "/auth/refresh", s.authRefresh())
}

// Start binds all routes to our router and then serves our
// router to handle all requests on incoming connections.
func (s *Server) Start() {
	s.registerRoutes()
	fmt.Println(fmt.Sprintf("Starting server on :%v", s.Port))
	http.ListenAndServe(":"+s.Port, s.Router)
}
