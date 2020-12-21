package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/khiki1995/crud/cmd/app/middleware"

	"github.com/gorilla/mux"
	"github.com/khiki1995/crud/pkg/customers"
	"github.com/khiki1995/crud/pkg/managers"
)

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	managersSvc  *managers.Service
}

type Token struct {
	Token string `json:"token"`
}

func NewServer(mux *mux.Router, customersSvc *customers.Service, managersSvc *managers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, managersSvc: managersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init() {
//	customersAuth := middleware.Authenticate(s.customersSvc.IDByToken)
	customersSR := s.mux.PathPrefix("/api/customers").Subrouter()
//	customersSR.Use(customersAuth)
	customersSR.HandleFunc("", s.handleCustomerRegistration).Methods(POST)
	customersSR.HandleFunc("/token", s.handleCustomerGetToken).Methods(POST)
	customersSR.HandleFunc("/token/validate", s.handleCustomerValidateToken).Methods(POST)
	customersSR.HandleFunc("/products", s.handleCustomerGetProducts).Methods(GET)
	customersSR.HandleFunc("/purchases", s.handleCustomerGetPurchases).Methods(GET)

	managersAuth := middleware.Authenticate(s.managersSvc.IDByToken)
	managersSR := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSR.Use(managersAuth)
	managersSR.HandleFunc("", s.handleManagerRegistration).Methods(POST)
	managersSR.HandleFunc("/token", s.handleManagerGetToken).Methods(POST)
	managersSR.HandleFunc("/sales", s.handleManagerMakeSale).Methods(POST)
	managersSR.HandleFunc("/sales", s.handleManagerGetSales).Methods(GET)
	managersSR.HandleFunc("/products", s.handleManagerChangeProduct).Methods(POST)
	managersSR.HandleFunc("/products", s.handleManagerGetProducts).Methods(GET)
	managersSR.HandleFunc("/products/{id}", s.handleManagerRemoveProductByID).Methods(DELETE)
	managersSR.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods(POST)
	managersSR.HandleFunc("/customers", s.handleManagerGetCustomers).Methods(GET)
	managersSR.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods(DELETE)
}

func responseJSON(w http.ResponseWriter, statusCode int, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
		return
	}
}
