package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/khiki1995/crud/pkg/managers"
)

func (s *Server) handleManagerRegistration(writer http.ResponseWriter, request *http.Request) {
	var item *managers.Registration
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	manager, err := s.managersSvc.Register(request.Context(), item)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseJSON(writer, 200, manager)
}
