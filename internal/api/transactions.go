package api

import (
	"encoding/json"
	"net/http"
	"time"

	core "github.com/lualfe/offer-eligibility-api/internal/core"
)

type CreateTransactionsRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions"`
}

type CreateTransactionRequest struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	MerchantID  string    `json:"merchant_id"`
	MCC         string    `json:"mcc"`
	AmountCents int64     `json:"amount_cents"`
	ApprovedAt  time.Time `json:"approved_at"`
}

type CreateTransactionsResponse struct {
	Inserted int `json:"inserted"`
}

func (s Server) CreateTransactions(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionsRequest
	json.NewDecoder(r.Body).Decode(&req)

	err := s.service.CreateTransactions(r.Context(), toCoreCreateTransactionsRequest(req))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	resp := CreateTransactionsResponse{
		Inserted: len(req.Transactions),
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func toCoreCreateTransactionsRequest(req CreateTransactionsRequest) core.CreateTransactionsRequest {
	var coreReq core.CreateTransactionsRequest
	transactions := make([]core.CreateTransactionRequest, len(req.Transactions))
	for i, t := range req.Transactions {
		transactions[i] = core.CreateTransactionRequest{
			ID:          t.ID,
			UserID:      t.UserID,
			MerchantID:  t.MerchantID,
			MCC:         t.MCC,
			AmountCents: t.AmountCents,
			ApprovedAt:  t.ApprovedAt,
		}
	}

	coreReq.Transactions = transactions
	return coreReq
}
