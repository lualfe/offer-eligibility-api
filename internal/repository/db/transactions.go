package db

import (
	"context"

	"github.com/lualfe/offer-eligibility-api/internal/core"
	. "github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/table"
)

func (r Repo) CreateTransactions(ctx context.Context, req core.CreateTransactionsRequest) error {
	stmt := Transactions.INSERT(
		Transactions.ID,
		Transactions.MerchantID,
		Transactions.UserID,
		Transactions.Mcc,
		Transactions.AmountCents,
		Transactions.ApprovedAt,
	)

	for _, tx := range req.Transactions {
		stmt = stmt.VALUES(
			tx.ID,
			tx.MerchantID,
			tx.UserID,
			tx.MCC,
			tx.AmountCents,
			tx.ApprovedAt,
		)
	}

	_, err := stmt.ExecContext(ctx, r.db)
	return err
}
