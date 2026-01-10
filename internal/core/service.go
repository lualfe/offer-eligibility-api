package core

import (
	"context"
	"time"
)

type Service struct {
	repo     Repository
	timeFunc func() time.Time
}

func NewService(repo Repository) Service {
	return Service{
		repo:     repo,
		timeFunc: time.Now,
	}
}

type Repository interface {
	CreateOffer(ctx context.Context, req CreateOfferRequest) error
	CreateTransactions(ctx context.Context, req CreateTransactionsRequest) error
	GetEligibleOffers(ctx context.Context, req GetEligibleOffersRequest) (EligibleOffers, error)
}
