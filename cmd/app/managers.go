package 

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/khiki1995/crud/cmd/app/middleware"
	"github.com/khiki1995/crud/pkg/managers"
	"golang.org/x/crypto/bcrypt"
)

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

type Server struct {
	mux          *mux.Router
	managersSvc  *managers.Service
}

func NewServer(mux *mux.Router, managersSvc *managers.Service) *Server {
	return &Server{mux: mux, managersSvc: managersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
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

func (s *Server) Init() {
	managersAuthMW := middleware.Authenticate(s.managersSvc.IDByToken)

	managersSR :=	s.mux.PathPrefix("/api/managers").Subrouter()
	managersSR.Use(managersAuthMW)
	managersSR.HandleFunc("", s.handleManagerRegistration).Methods(POST)
	managersSR.HandleFunc("/token", s.handleeManagerGetToken).Methods(POST)
	managersSR.HandleFunc("/sales", s.handleManagerMakeSale).Methods(POST)
	managersSR.HandleFunc("/products", s.handleManagerChangeProduct).Methods(POST)
	managersSR.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods(POST)

	managersSR.HandleFunc("/products", s.handleManagerGetProducts).Methods(GET)
	managersSR.HandleFunc("/sales", s.handleManagerGetSales).Methods(GET)
	managersSR.HandleFunc("/customers", s.handleManagerGetCustomers).Methods(GET)

	managersSR.HandleFunc("/products/{id}", s.handleManagerRemoveProductByID).Methods(DELETE)
	managersSR.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods(DELETE)
}

func (s *Server) handleManagerRegistration(writer http.ResponseWriter, request *http.Request) {
	var item *managers.Registration
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	item.Password = string(hash)
	manager, err := s.managersSvc.Register(request.Context(), item)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, manager)
}