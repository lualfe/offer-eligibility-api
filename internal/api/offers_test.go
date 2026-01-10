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

func TestServer_CreateOffer(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		request        CreateOfferRequest
		setupMock      func(m *MockService)
		wantStatusCode int
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name: "service returns error",
			request: CreateOfferRequest{
				ID:              "offer-123",
				MerchantID:      "merchant-456",
				MCCWhiteList:    []string{"5411"},
				Active:          true,
				MinTransactions: 10,
				LookbackDays:    60,
				StartsDate:      startsAt,
				EndDate:         endsAt,
			},
			setupMock: func(m *MockService) {
				m.EXPECT().CreateOffer(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErr:        true,
			wantErrMsg:     "database error",
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
				StartsDate:      startsAt,
				EndDate:         endsAt,
			},
			setupMock: func(m *MockService) {
				m.EXPECT().CreateOffer(gomock.Any(), core.CreateOfferRequest{
					ID:              "offer-123",
					MerchantID:      "merchant-456",
					MCCWhiteList:    []string{"5411", "5412"},
					Active:          true,
					MinTransactions: 5,
					LookbackDays:    30,
					StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:         time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
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

			req := httptest.NewRequest(http.MethodPost, "/offers", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			server.CreateOffer(rec, req)

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

			var resp CreateOfferResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			want := CreateOfferResponse{
				ID:              tt.request.ID,
				MerchantID:      tt.request.MerchantID,
				MCCWhiteList:    tt.request.MCCWhiteList,
				Active:          tt.request.Active,
				MinTransactions: tt.request.MinTransactions,
				LookbackDays:    tt.request.LookbackDays,
				StartsDate:      tt.request.StartsDate,
				EndDate:         tt.request.EndDate,
			}

			if diff := cmp.Diff(want, resp); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
