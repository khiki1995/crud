package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/khiki1995/crud/cmd/app/middleware"
	"github.com/khiki1995/crud/pkg/customers"
	"github.com/khiki1995/crud/pkg/managers"
)

func (s *Server) handleManagerRegistration(writer http.ResponseWriter, request *http.Request) {
	var reg *managers.Registration
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if !s.managersSvc.IsAdmin(request.Context(), id) {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	err = json.NewDecoder(request.Body).Decode(&reg)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	token, err := s.managersSvc.Register(request.Context(), reg)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, Token{Token: token})
}

func (s *Server) handleManagerGetToken(writer http.ResponseWriter, request *http.Request) {
	var manager *managers.Manager
	err := json.NewDecoder(request.Body).Decode(&manager)

	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := s.managersSvc.GetToken(request.Context(), manager.Phone, manager.Password)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerChangeProduct(writer http.ResponseWriter, request *http.Request) {
	var product *managers.Product
	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(request.Body).Decode(&product)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.managersSvc.SaveProduct(request.Context(), product)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, item)
}

func (s *Server) handleManagerMakeSale(writer http.ResponseWriter, request *http.Request) {
	var sale *managers.Sale

	id, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(request.Body).Decode(&sale)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	sale.Manager_id = id

	item, err := s.managersSvc.MakeSale(request.Context(), sale)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, item)
}

func (s *Server) handleManagerGetSales(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	total, err := s.managersSvc.GetSales(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, map[string]interface{}{"manager_id": id, "total": total})
}

func (s *Server) handleManagerGetProducts(writer http.ResponseWriter, request *http.Request) {
	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	products, err := s.managersSvc.GetProducts(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, products)
}

func (s *Server) handleManagerRemoveProductByID(writer http.ResponseWriter, request *http.Request) {
	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	product, err := s.managersSvc.RemoveProductByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, product)
}

func (s *Server) handleManagerChangeCustomer(writer http.ResponseWriter, request *http.Request) {
	var customer *customers.Customer

	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(request.Body).Decode(&customer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.managersSvc.ChangeCustomer(request.Context(), customer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, item)
}

func (s *Server) handleManagerGetCustomers(writer http.ResponseWriter, request *http.Request) {
	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	customers, err := s.managersSvc.GetCustomers(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, customers)
}

func (s *Server) handleManagerRemoveCustomerByID(writer http.ResponseWriter, request *http.Request) {
	_, err := middleware.Authentication(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	customer, err := s.managersSvc.RemoveCustomerByID(request.Context(), id)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responseJSON(writer, 200, customer)
}
