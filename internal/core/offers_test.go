package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestService_CreateOffer(t *testing.T) {
	errCreatingOffer := errors.New("database error")

	tests := []struct {
		name      string
		request   CreateOfferRequest
		setupMock func(m *MockRepository)
		wantErr   error
	}{
		{
			name: "repository returns error",
			request: CreateOfferRequest{
				ID:              "offer-123",
				MerchantID:      "merchant-456",
				MCCWhiteList:    []string{"5411"},
				Active:          true,
				MinTransactions: 10,
				LookbackDays:    60,
				StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:         time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().CreateOffer(gomock.Any(), gomock.Any()).Return(errCreatingOffer)
			},
			wantErr: errCreatingOffer,
		},
		{
			name: "success",
			request: CreateOfferRequest{
				ID:              "offer-123",
				MerchantID:      "merchant-456",
				MCCWhiteList:    []string{"5411", "5412"},
				Active:          true,
				MinTransactions: 5,
				LookbackDays:    30,
				StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:         time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().CreateOffer(gomock.Any(), CreateOfferRequest{
					ID:              "offer-123",
					MerchantID:      "merchant-456",
					MCCWhiteList:    []string{"5411", "5412"},
					Active:          true,
					MinTransactions: 5,
					LookbackDays:    30,
					StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:         time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
				}).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewService(mockRepo)

			err := service.CreateOffer(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetEligibleOffers(t *testing.T) {
	errGettingOffers := errors.New("database error")
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		request   GetEligibleOffersRequest
		timeFunc  func() time.Time
		setupMock func(m *MockRepository)
		want      EligibleOffers
		wantErr   error
	}{
		{
			name: "repository returns error",
			request: GetEligibleOffersRequest{
				UserID: "user-123",
				Now:    &fixedTime,
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().GetEligibleOffers(gomock.Any(), GetEligibleOffersRequest{
					UserID: "user-123",
					Now:    &fixedTime,
				}).Return(EligibleOffers{}, errGettingOffers)
			},
			wantErr: errGettingOffers,
		},
		{
			name: "success with provided time",
			request: GetEligibleOffersRequest{
				UserID: "user-456",
				Now:    &fixedTime,
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().GetEligibleOffers(gomock.Any(), GetEligibleOffersRequest{
					UserID: "user-456",
					Now:    &fixedTime,
				}).Return(EligibleOffers{
					UserID: "user-456",
					Offers: []EligibleOffer{
						{OfferID: "offer-1", Reason: "qualified"},
						{OfferID: "offer-2", Reason: "loyal customer"},
					},
				}, nil)
			},
			want: EligibleOffers{
				UserID: "user-456",
				Offers: []EligibleOffer{
					{OfferID: "offer-1", Reason: "qualified"},
					{OfferID: "offer-2", Reason: "loyal customer"},
				},
			},
		},
		{
			name: "success with nil time uses timeFunc",
			request: GetEligibleOffersRequest{
				UserID: "user-789",
				Now:    nil,
			},
			timeFunc: func() time.Time { return fixedTime },
			setupMock: func(m *MockRepository) {
				m.EXPECT().GetEligibleOffers(gomock.Any(), GetEligibleOffersRequest{
					UserID: "user-789",
					Now:    &fixedTime,
				}).Return(EligibleOffers{
					UserID: "user-789",
					Offers: []EligibleOffer{
						{OfferID: "offer-3", Reason: "new user bonus"},
					},
				}, nil)
			},
			want: EligibleOffers{
				UserID: "user-789",
				Offers: []EligibleOffer{
					{OfferID: "offer-3", Reason: "new user bonus"},
				},
			},
		},
		{
			name: "success with no eligible offers",
			request: GetEligibleOffersRequest{
				UserID: "user-000",
				Now:    &fixedTime,
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().GetEligibleOffers(gomock.Any(), GetEligibleOffersRequest{
					UserID: "user-000",
					Now:    &fixedTime,
				}).Return(EligibleOffers{
					UserID: "user-000",
					Offers: []EligibleOffer{},
				}, nil)
			},
			want: EligibleOffers{
				UserID: "user-000",
				Offers: []EligibleOffer{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := Service{
				repo:     mockRepo,
				timeFunc: tt.timeFunc,
			}
			if service.timeFunc == nil {
				service.timeFunc = time.Now
			}

			got, err := service.GetEligibleOffers(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetEligibleOffers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
