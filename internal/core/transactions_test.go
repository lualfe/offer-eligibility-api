package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestService_CreateTransactions(t *testing.T) {
	errCreatingTransactions := errors.New("database error")

	tests := []struct {
		name      string
		request   CreateTransactionsRequest
		setupMock func(m *MockRepository)
		wantErr   error
	}{
		{
			name: "repository returns error",
			request: CreateTransactionsRequest{
				Transactions: []CreateTransactionRequest{
					{
						ID:          "txn-123",
						UserID:      "user-456",
						MerchantID:  "merchant-789",
						MCC:         "5411",
						AmountCents: 1000,
						ApprovedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().CreateTransactions(gomock.Any(), gomock.Any()).Return(errCreatingTransactions)
			},
			wantErr: errCreatingTransactions,
		},
		{
			name: "success with multiple transactions",
			request: CreateTransactionsRequest{
				Transactions: []CreateTransactionRequest{
					{
						ID:          "txn-123",
						UserID:      "user-456",
						MerchantID:  "merchant-789",
						MCC:         "5411",
						AmountCents: 1000,
						ApprovedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:          "txn-456",
						UserID:      "user-456",
						MerchantID:  "merchant-789",
						MCC:         "5412",
						AmountCents: 2500,
						ApprovedAt:  time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC),
					},
				},
			},
			setupMock: func(m *MockRepository) {
				m.EXPECT().CreateTransactions(gomock.Any(), CreateTransactionsRequest{
					Transactions: []CreateTransactionRequest{
						{
							ID:          "txn-123",
							UserID:      "user-456",
							MerchantID:  "merchant-789",
							MCC:         "5411",
							AmountCents: 1000,
							ApprovedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						},
						{
							ID:          "txn-456",
							UserID:      "user-456",
							MerchantID:  "merchant-789",
							MCC:         "5412",
							AmountCents: 2500,
							ApprovedAt:  time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC),
						},
					},
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

			err := service.CreateTransactions(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
