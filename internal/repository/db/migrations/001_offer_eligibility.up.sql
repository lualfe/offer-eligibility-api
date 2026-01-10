CREATE TABLE merchants (
    id UUID PRIMARY KEY,
    mechant_name VARCHAR(100) NOT NULL
);

CREATE TABLE offers (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    active BOOLEAN NOT NULL DEFAULT false,
    min_transactions SMALLINT NOT NULL,
    lookback_days SMALLINT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL
);

CREATE TABLE offers_mccs (
    offer_id UUID NOT NULL REFERENCES offers (id),
    mcc VARCHAR(4) NOT NULL,
    PRIMARY KEY (offer_id, mcc)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    user_id UUID NOT NULL,
    mcc VARCHAR(4) NOT NULL,
    amount_cents INT NOT NULL,
    approved_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_offers_merchant_id ON offers (merchant_id);

CREATE INDEX idx_offers_active_dates
ON offers (active, start_date, end_date) WHERE active = true;

CREATE INDEX idx_transactions_user_approved_at
ON transactions (user_id, approved_at);

CREATE INDEX idx_transactions_merchant_id ON transactions (merchant_id);

CREATE INDEX idx_transactions_mcc ON transactions (mcc);
