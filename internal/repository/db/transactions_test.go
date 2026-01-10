package db

import (
	"context"
	"testing"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/lualfe/offer-eligibility-api/internal/core"
	"github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/model"
	. "github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/table"
)

func TestCreateTransactions(t *testing.T) {
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
		Transactions.DELETE().WHERE(Transactions.MerchantID.EQ(CAST(String(merchantID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		Merchants.DELETE().WHERE(Merchants.ID.EQ(CAST(String(merchantID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
	})

	t.Run("creates transactions successfully", func(t *testing.T) {
		tx1ID := uuid.New().String()
		tx2ID := uuid.New().String()
		userID := uuid.New().String()
		tx1ApprovedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
		tx2ApprovedAt := time.Date(2024, 6, 16, 11, 0, 0, 0, time.UTC)

		req := core.CreateTransactionsRequest{
			Transactions: []core.CreateTransactionRequest{
				{
					ID:          tx1ID,
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5411",
					AmountCents: 1000,
					ApprovedAt:  tx1ApprovedAt,
				},
				{
					ID:          tx2ID,
					UserID:      userID,
					MerchantID:  merchantID,
					MCC:         "5412",
					AmountCents: 2500,
					ApprovedAt:  tx2ApprovedAt,
				},
			},
		}

		err := repoSuite.CreateTransactions(ctx, req)
		if err != nil {
			t.Fatalf("CreateTransactions failed: %v", err)
		}

		t.Cleanup(func() {
			Transactions.DELETE().WHERE(Transactions.ID.EQ(CAST(String(tx1ID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
			Transactions.DELETE().WHERE(Transactions.ID.EQ(CAST(String(tx2ID)).AS_UUID())).ExecContext(ctx, repoSuite.db)
		})

		merchantUUID := uuid.MustParse(merchantID)
		userUUID := uuid.MustParse(userID)

		// Verify first transaction
		want1 := model.Transactions{
			ID:          uuid.MustParse(tx1ID),
			MerchantID:  merchantUUID,
			UserID:      userUUID,
			Mcc:         "5411",
			AmountCents: 1000,
			ApprovedAt:  tx1ApprovedAt,
		}

		var got1 []model.Transactions
		err = Transactions.SELECT(Transactions.AllColumns).
			WHERE(Transactions.ID.EQ(CAST(String(tx1ID)).AS_UUID())).
			QueryContext(ctx, repoSuite.db, &got1)
		if err != nil {
			t.Fatalf("failed to query first transaction: %v", err)
		}
		if diff := cmp.Diff(want1, got1[0]); diff != "" {
			t.Errorf("first transaction mismatch (-want +got):\n%s", diff)
		}

		// Verify second transaction
		want2 := model.Transactions{
			ID:          uuid.MustParse(tx2ID),
			MerchantID:  merchantUUID,
			UserID:      userUUID,
			Mcc:         "5412",
			AmountCents: 2500,
			ApprovedAt:  tx2ApprovedAt,
		}

		var got2 []model.Transactions
		err = Transactions.SELECT(Transactions.AllColumns).
			WHERE(Transactions.ID.EQ(CAST(String(tx2ID)).AS_UUID())).
			QueryContext(ctx, repoSuite.db, &got2)
		if err != nil {
			t.Fatalf("failed to query second transaction: %v", err)
		}
		if diff := cmp.Diff(want2, got2[0]); diff != "" {
			t.Errorf("second transaction mismatch (-want +got):\n%s", diff)
		}
	})
}
