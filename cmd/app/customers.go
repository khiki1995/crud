package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/khiki1995/crud/cmd/app/middleware"
	"github.com/khiki1995/crud/pkg/customers"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleCustomerRegistration(writer http.ResponseWriter, request *http.Request) {
	var item *customers.Registration
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
	customer, err := s.customersSvc.Register(request.Context(), item)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, customer)
}
func (s *Server) handleCustomerGetToken(writer http.ResponseWriter, request *http.Request) {
	var auth *customers.Auth
	err := json.NewDecoder(request.Body).Decode(&auth)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	token, err := s.customersSvc.GetToken(request.Context(), auth.Login, auth.Password)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, map[string]string{"token": token})
}
func (s *Server) handleCustomerValidateToken(writer http.ResponseWriter, request *http.Request) {
	var token Token
	response := make(map[string]interface{})
	err := json.NewDecoder(request.Body).Decode(&token)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := s.customersSvc.AuthentificateCustomer(request.Context(), token.Token)

	if err != nil {
		response["status"] = "fail"
		if err == customers.ErrTokenExpired {
			response["reason"] = "expired"
			responseJSON(writer, 400, response)
			return
		}
		response["reason"] = "not found"
		responseJSON(writer, 404, response)
		return
	}
	response["status"] = "ok"
	response["customerId"] = id
	responseJSON(writer, 200, response)
	return
}
func (s *Server) handleCustomerGetProducts(writer http.ResponseWriter, request *http.Request) {
	items, err := s.customersSvc.Products(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}
func (s *Server) handleCustomerGetPurchases(writer http.ResponseWriter, request *http.Request) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items, err := s.customersSvc.Purchases(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}
