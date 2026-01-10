package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lualfe/offer-eligibility-api/internal/core"
)

type CreateOfferRequest struct {
	ID              string    `json:"id"`
	MerchantID      string    `json:"merchant_id"`
	MCCWhiteList    []string  `json:"mcc_whitelist"`
	Active          bool      `json:"active"`
	MinTransactions int       `json:"min_txn_count"`
	LookbackDays    int       `json:"lookback_days"`
	StartsDate      time.Time `json:"starts_at"`
	EndDate         time.Time `json:"ends_at"`
}

type CreateOfferResponse struct {
	ID              string    `json:"id"`
	MerchantID      string    `json:"merchant_id"`
	MCCWhiteList    []string  `json:"mcc_whitelist"`
	Active          bool      `json:"active"`
	MinTransactions int       `json:"min_txn_count"`
	LookbackDays    int       `json:"lookback_days"`
	StartsDate      time.Time `json:"starts_at"`
	EndDate         time.Time `json:"ends_at"`
}

type GetEligibleOffersResponse struct {
	UserID         string          `json:"user_id"`
	EligibleOffers []EligibleOffer `json:"eligible_offers"`
}

type EligibleOffer struct {
	OfferID string `json:"offer_id"`
	Reason  string `json:"reason"`
}

func (s Server) CreateOffer(w http.ResponseWriter, r *http.Request) {
	var req CreateOfferRequest
	json.NewDecoder(r.Body).Decode(&req)

	in := toCoreCreateOfferRequest(req)
	err := s.service.CreateOffer(r.Context(), in)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := CreateOfferResponse{
		ID:              req.ID,
		MerchantID:      req.MerchantID,
		MCCWhiteList:    req.MCCWhiteList,
		Active:          req.Active,
		MinTransactions: req.MinTransactions,
		LookbackDays:    req.LookbackDays,
		StartsDate:      req.StartsDate,
		EndDate:         req.EndDate,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s Server) GetEligibleOffers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	now := r.URL.Query().Get("now")

	req := core.GetEligibleOffersRequest{
		UserID: userID,
	}

	if now != "" {
		parsedTime, err := time.Parse(time.RFC3339, now)
		if err == nil {
			req.Now = &parsedTime
		}
	}

	offers, err := s.service.GetEligibleOffers(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := toAPIEligibleOffers(offers)
	json.NewEncoder(w).Encode(resp)
}

func toAPIEligibleOffers(coreOffers core.EligibleOffers) GetEligibleOffersResponse {
	apiOffers := GetEligibleOffersResponse{
		UserID:         coreOffers.UserID,
		EligibleOffers: []EligibleOffer{},
	}

	offers := make([]EligibleOffer, len(coreOffers.Offers))

	for i, offer := range coreOffers.Offers {
		offers[i] = EligibleOffer{
			OfferID: offer.OfferID,
			Reason:  offer.Reason,
		}
	}

	apiOffers.EligibleOffers = offers

	return apiOffers
}

func toCoreCreateOfferRequest(apiReq CreateOfferRequest) core.CreateOfferRequest {
	return core.CreateOfferRequest{
		ID:              apiReq.ID,
		MerchantID:      apiReq.MerchantID,
		MCCWhiteList:    apiReq.MCCWhiteList,
		Active:          apiReq.Active,
		MinTransactions: apiReq.MinTransactions,
		LookbackDays:    apiReq.LookbackDays,
		StartDate:       apiReq.StartsDate,
		EndDate:         apiReq.EndDate,
	}
}
