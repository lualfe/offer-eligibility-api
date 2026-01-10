package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s Server) InitHandlers() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/offers", s.CreateOffer).Methods("POST")
	r.HandleFunc("/transactions", s.CreateTransactions).Methods("POST")
	r.HandleFunc("/users/{user_id}/eligible-offers", s.GetEligibleOffers).Methods("GET")

	return r
}
