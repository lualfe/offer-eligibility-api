package api

import (
	"context"

	"github.com/lualfe/offer-eligibility-api/internal/core"
)

type Server struct {
	service Service
}

type Service interface {
	CreateOffer(ctx context.Context, req core.CreateOfferRequest) error
	CreateTransactions(ctx context.Context, req core.CreateTransactionsRequest) error
	GetEligibleOffers(ctx context.Context, req core.GetEligibleOffersRequest) (core.EligibleOffers, error)
}

func NewServer(service Service) Server {
	return Server{
		service: service,
	}
}
