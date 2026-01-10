package core

import (
	context "context"
	"time"
)

type CreateTransactionsRequest struct {
	Transactions []CreateTransactionRequest
}

type CreateTransactionRequest struct {
	ID          string
	UserID      string
	MerchantID  string
	MCC         string
	AmountCents int64
	ApprovedAt  time.Time
}

func (s Service) CreateTransactions(ctx context.Context, req CreateTransactionsRequest) error {
	return s.repo.CreateTransactions(ctx, req)
}
