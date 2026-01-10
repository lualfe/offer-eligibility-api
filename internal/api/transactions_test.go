package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lualfe/offer-eligibility-api/internal/core"
	"go.uber.org/mock/gomock"
)

func TestServer_CreateTransactions(t *testing.T) {
	approvedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		request        CreateTransactionsRequest
		setupMock      func(m *MockService)
		wantStatusCode int
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name: "service returns error",
			request: CreateTransactionsRequest{
				Transactions: []CreateTransactionRequest{
					{
						ID:          "txn-123",
						UserID:      "user-456",
						MerchantID:  "merchant-789",
						MCC:         "5411",
						AmountCents: 1000,
						ApprovedAt:  approvedAt,
					},
				},
			},
			setupMock: func(m *MockService) {
				m.EXPECT().CreateTransactions(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErr:        true,
			wantErrMsg:     "database error",
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
						ApprovedAt:  approvedAt,
					},
					{
						ID:          "txn-124",
						UserID:      "user-456",
						MerchantID:  "merchant-790",
						MCC:         "5412",
						AmountCents: 2500,
						ApprovedAt:  approvedAt,
					},
				},
			},
			setupMock: func(m *MockService) {
				m.EXPECT().CreateTransactions(gomock.Any(), core.CreateTransactionsRequest{
					Transactions: []core.CreateTransactionRequest{
						{
							ID:          "txn-123",
							UserID:      "user-456",
							MerchantID:  "merchant-789",
							MCC:         "5411",
							AmountCents: 1000,
							ApprovedAt:  approvedAt,
						},
						{
							ID:          "txn-124",
							UserID:      "user-456",
							MerchantID:  "merchant-790",
							MCC:         "5412",
							AmountCents: 2500,
							ApprovedAt:  approvedAt,
						},
					},
				}).Return(nil)
			},
			wantStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			server := NewServer(mockService)

			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			server.CreateTransactions(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("status code = %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantErr {
				var errResp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.ErrorMessage != tt.wantErrMsg {
					t.Errorf("error message = %q, want %q", errResp.ErrorMessage, tt.wantErrMsg)
				}
				return
			}

			var resp CreateTransactionsResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			want := CreateTransactionsResponse{
				Inserted: len(tt.request.Transactions),
			}

			if diff := cmp.Diff(want, resp); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
