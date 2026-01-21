package db

import (
	"context"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/lualfe/offer-eligibility-api/internal/core"
	. "github.com/lualfe/offer-eligibility-api/internal/repository/db/.gen/offer-eligibility/public/table"
)

func (r Repo) CreateOffer(ctx context.Context, req core.CreateOfferRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(core.WrapErrFmt, core.ErrInternal, err)
	}
	defer tx.Rollback()

	stmt := Offers.INSERT(
		Offers.ID, Offers.Active, Offers.MerchantID, Offers.MinTransactions, Offers.LookbackDays, Offers.StartDate, Offers.EndDate,
	).VALUES(
		req.ID,
		req.Active,
		req.MerchantID,
		req.MinTransactions,
		req.LookbackDays,
		req.StartDate,
		req.EndDate,
	)

	_, err = stmt.ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf(core.WrapErrFmt, core.ErrInternal, err)
	}

	mccStmt := OffersMccs.INSERT(
		OffersMccs.OfferID, OffersMccs.Mcc,
	)

	for _, mcc := range req.MCCWhiteList {
		mccStmt = mccStmt.VALUES(
			req.ID,
			mcc,
		)
	}

	_, err = mccStmt.ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf(core.WrapErrFmt, core.ErrInternal, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(core.WrapErrFmt, core.ErrInternal, err)
	}

	return nil
}

func (r Repo) GetEligibleOffers(ctx context.Context, req core.GetEligibleOffersRequest) (core.EligibleOffers, error) {
	if req.Now == nil {
		return core.EligibleOffers{}, nil
	}
	now := *req.Now

	reasonExpr := CONCAT(String(">= min_txn_count in last "), Offers.LookbackDays, String(" days")).AS("reason")

	transactionsFilter := SELECT(Int(1)).
		FROM(Transactions).
		WHERE(
			AND(
				Transactions.UserID.EQ(CAST(String(req.UserID)).AS_UUID()),
				Transactions.ApprovedAt.GT_EQ(
					TimestampzT(now).SUB(INTERVAL(1, DAY).MUL(IntExp(Offers.LookbackDays))),
				),
			),
		).
		GROUP_BY(Transactions.UserID).
		HAVING(COUNT(STAR).GT_EQ(Offers.MinTransactions))

	activeOffersFilter := AND(
		Offers.Active.EQ(Bool(true)),
		Offers.StartDate.LT_EQ(TimestampzT(now)),
		Offers.EndDate.GT_EQ(TimestampzT(now)),
		EXISTS(transactionsFilter),
	)

	merchantBranch := SELECT(Offers.ID.AS("offer_id"), reasonExpr).
		DISTINCT().
		FROM(Offers.INNER_JOIN(Transactions, Offers.MerchantID.EQ(Transactions.MerchantID))).
		WHERE(activeOffersFilter)

	mccBranch := SELECT(Offers.ID.AS("offer_id"), reasonExpr).
		DISTINCT().
		FROM(
			Offers.
				INNER_JOIN(OffersMccs, OffersMccs.OfferID.EQ(Offers.ID)).
				INNER_JOIN(Transactions, OffersMccs.Mcc.EQ(Transactions.Mcc)),
		).
		WHERE(activeOffersFilter)

	stmt := UNION(merchantBranch, mccBranch)

	var dest []struct {
		OfferID string `sql:"offer_id"`
		Reason  string `sql:"reason"`
	}
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		return core.EligibleOffers{}, fmt.Errorf(core.WrapErrFmt, core.ErrInternal, err)
	}

	res := make([]core.EligibleOffer, len(dest))
	for i, offer := range dest {
		res[i] = core.EligibleOffer{
			OfferID: offer.OfferID,
			Reason:  offer.Reason,
		}
	}

	return core.EligibleOffers{
		UserID: req.UserID,
		Offers: res,
	}, nil
}
