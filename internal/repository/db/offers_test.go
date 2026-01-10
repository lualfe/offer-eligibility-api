package db

import (
	"context"
	"testing"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/lualfe/offer-eligibility-api/internal/core"
	"github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/model"
	. "github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/table"
)

func TestCreateOffer(t *testing.T) {
	ctx := context.Background()

	// Create a merchant first (foreign key constraint)
	merchantID := uuid.New().String()
	merchantStmt := Merchants.INSERT(
		Merchants.ID, Merchants.MechantName,
	).VALUES(merchantID, "Test Merchant")

	_, err := merchantStmt.ExecContext(ctx, repoSuite.db)
	if err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	t.Cleanup(func() {
		Merchants.DELETE().WHERE(Merchants.ID.EQ(CAST(String(merchantID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
	})

	t.Run("creates offer and MCCs successfully", func(t *testing.T) {
		offerID := uuid.New().String()
		mccs := []string{"5411", "5412", "5499"}

		req := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    mccs,
			Active:          true,
			MinTransactions: 5,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}

		err := repoSuite.CreateOffer(ctx, req)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		t.Cleanup(func() {
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		// Verify offer was created
		var offers []model.Offers
		err = Offers.SELECT(Offers.ID).
			WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).
			QueryContext(ctx, repoSuite.db, &offers)
		if err != nil {
			t.Fatalf("failed to query offer: %v", err)
		}
		if len(offers) != 1 {
			t.Errorf("expected 1 offer, got %d", len(offers))
		}

		// Verify MCCs were created
		var offerMccs []model.OffersMccs
		err = OffersMccs.SELECT(OffersMccs.OfferID, OffersMccs.Mcc).
			WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).
			QueryContext(ctx, repoSuite.db, &offerMccs)
		if err != nil {
			t.Fatalf("failed to query MCCs: %v", err)
		}
		if len(offerMccs) != len(mccs) {
			t.Errorf("expected %d MCCs, got %d", len(mccs), len(offerMccs))
		}
	})
}

func TestGetEligibleOffers(t *testing.T) {
	ctx := context.Background()

	// Create a merchant
	merchantID := uuid.New().String()
	merchantStmt := Merchants.INSERT(
		Merchants.ID, Merchants.MechantName,
	).VALUES(merchantID, "Test Merchant")

	_, err := merchantStmt.ExecContext(ctx, repoSuite.db)
	if err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	t.Cleanup(func() {
		Merchants.DELETE().WHERE(Merchants.ID.EQ(CAST(String(merchantID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
	})

	t.Run("returns eligible offer when user matches merchant transactions", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an active offer with MCC whitelist that does NOT match transaction MCCs
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"9999"}, // Does not match transaction MCCs
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions for the user with this merchant (matching by merchant, not MCC)
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5412",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if got.UserID != userID {
			t.Errorf("expected UserID %s, got %s", userID, got.UserID)
		}
		if len(got.Offers) != 1 {
			t.Fatalf("expected 1 offer, got %d", len(got.Offers))
		}
		if got.Offers[0].OfferID != offerID {
			t.Errorf("expected OfferID %s, got %s", offerID, got.Offers[0].OfferID)
		}
		expectedReason := ">= min_txn_count in last 30 days"
		if got.Offers[0].Reason != expectedReason {
			t.Errorf("expected Reason %q, got %q", expectedReason, got.Offers[0].Reason)
		}
	})

	t.Run("returns eligible offer when user matches MCC whitelist", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		otherMerchantID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create another merchant for transactions
		_, err := Merchants.INSERT(Merchants.ID, Merchants.MechantName).
			VALUES(otherMerchantID, "Other Merchant").
			ExecContext(ctx, repoSuite.db)
		if err != nil {
			t.Fatalf("failed to create other merchant: %v", err)
		}

		// Create an offer with MCC whitelist
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411", "5412"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err = repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions with matching MCC but different merchant
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  otherMerchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  otherMerchantID,
					MCC:         "5412",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Merchants.DELETE().WHERE(Merchants.ID.EQ(CAST(String(otherMerchantID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 1 {
			t.Fatalf("expected 1 offer, got %d", len(got.Offers))
		}
		if got.Offers[0].OfferID != offerID {
			t.Errorf("expected OfferID %s, got %s", offerID, got.Offers[0].OfferID)
		}
		expectedReason := ">= min_txn_count in last 30 days"
		if got.Offers[0].Reason != expectedReason {
			t.Errorf("expected Reason %q, got %q", expectedReason, got.Offers[0].Reason)
		}
	})

	t.Run("returns no offers when offer is inactive", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an inactive offer
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          false,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns no offers when outside date range", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC) // After offer end date

		// Create an offer that has ended
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns no offers when user has insufficient transactions", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an offer requiring 5 transactions
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 5,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create only 2 transactions
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns no offers when transactions are outside lookback period", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an offer with 7 day lookback
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    7,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions older than 7 days
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-10 * 24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-15 * 24 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns empty offers when user has no transactions", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an active offer
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		t.Cleanup(func() {
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if got.UserID != userID {
			t.Errorf("expected UserID %s, got %s", userID, got.UserID)
		}
		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns eligible offer when transaction is at edge of lookback period with different timezone", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()

		// Use a different timezone (e.g., Tokyo UTC+9)
		tokyo, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatalf("failed to load timezone: %v", err)
		}
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, tokyo)

		// Create an offer with 7 day lookback
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    7,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err = repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions: one just inside lookback (6 days 23 hours ago), one at edge (exactly 7 days ago)
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-6*24*time.Hour - 23*time.Hour), // 6 days 23 hours ago - inside lookback
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-7*24*time.Hour + 1*time.Minute), // 7 days ago + 1 minute - just inside lookback
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 1 {
			t.Fatalf("expected 1 offer, got %d", len(got.Offers))
		}
		if got.Offers[0].OfferID != offerID {
			t.Errorf("expected OfferID %s, got %s", offerID, got.Offers[0].OfferID)
		}
		expectedReason := ">= min_txn_count in last 7 days"
		if got.Offers[0].Reason != expectedReason {
			t.Errorf("expected Reason %q, got %q", expectedReason, got.Offers[0].Reason)
		}
	})

	t.Run("returns no offers when transaction is just outside lookback period with different timezone", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()

		// Use a different timezone (e.g., Los Angeles UTC-7/8)
		la, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			t.Fatalf("failed to load timezone: %v", err)
		}
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, la)

		// Create an offer with 7 day lookback
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    7,
			StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err = repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions: one inside lookback, one just outside (7 days + 1 minute ago)
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour), // 1 day ago - inside lookback
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-7*24*time.Hour - 1*time.Minute), // 7 days + 1 minute ago - outside lookback
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		// Should not be eligible because only 1 transaction is within lookback (needs 2)
		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})

	t.Run("returns eligible offer when offer start date equals now", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

		// Create an offer where start date is exactly "now"
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       now, // Start date equals now exactly
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		if len(got.Offers) != 1 {
			t.Fatalf("expected 1 offer, got %d", len(got.Offers))
		}
		if got.Offers[0].OfferID != offerID {
			t.Errorf("expected OfferID %s, got %s", offerID, got.Offers[0].OfferID)
		}
		expectedReason := ">= min_txn_count in last 30 days"
		if got.Offers[0].Reason != expectedReason {
			t.Errorf("expected Reason %q, got %q", expectedReason, got.Offers[0].Reason)
		}
	})

	t.Run("returns no offers when now is 1 second before offer start date", func(t *testing.T) {
		offerID := uuid.New().String()
		userID := uuid.New().String()
		startDate := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		now := startDate.Add(-1 * time.Second) // 1 second before start

		// Create an offer where now is just before start date
		offerReq := core.CreateOfferRequest{
			ID:              offerID,
			MerchantID:      merchantID,
			MCCWhiteList:    []string{"5411"},
			Active:          true,
			MinTransactions: 2,
			LookbackDays:    30,
			StartDate:       startDate,
			EndDate:         time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		}
		err := repoSuite.CreateOffer(ctx, offerReq)
		if err != nil {
			t.Fatalf("CreateOffer failed: %v", err)
		}

		// Create transactions
		txReq := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.New().String(),
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 2000,
					ApprovedAt:  now.Add(-48 * time.Hour),
				},
			},
		}
		err = repoSuite.CreateTransactions(ctx, txReq)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.UserID.EQ(CAST(String(userID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			OffersMccs.DELETE().WHERE(OffersMccs.OfferID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Offers.DELETE().WHERE(Offers.ID.EQ(CAST(String(offerID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		got, err := repoSuite.GetEligibleOffers(ctx, core.GetEligibleOffersRequest{
			UserID: userID,
			Now:    &now,
		})
		if err != nil {
			t.Fatalf("GetEligibleOffers failed: %v", err)
		}

		// Should not be eligible because now is before offer start date
		if len(got.Offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(got.Offers))
		}
	})
}
