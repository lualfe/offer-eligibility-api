package core

import (
	"context"
	"time"
)

type CreateOfferRequest struct {
	ID              string
	MerchantID      string
	MCCWhiteList    []string
	Active          bool
	MinTransactions int
	LookbackDays    int
	StartDate       time.Time
	EndDate         time.Time
}

type GetEligibleOffersRequest struct {
	UserID string
	Now    *time.Time
}

type EligibleOffers struct {
	UserID string
	Offers []EligibleOffer
}

type EligibleOffer struct {
	OfferID string
	Reason  string
}

func (s Service) CreateOffer(ctx context.Context, req CreateOfferRequest) error {
	return s.repo.CreateOffer(ctx, req)
}

func (s Service) GetEligibleOffers(ctx context.Context, req GetEligibleOffersRequest) (EligibleOffers, error) {
	if req.Now == nil {
		now := s.timeFunc()
		req.Now = &now
	}
	return s.repo.GetEligibleOffers(ctx, req)
}
