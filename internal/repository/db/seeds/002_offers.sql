-- Seed offers (~10,000)
-- Each offer references a random merchant
INSERT INTO offers (id, merchant_id, active, min_transactions, lookback_days, start_date, end_date)
SELECT
    gen_random_uuid(),
    (SELECT id FROM merchants ORDER BY random() LIMIT 1),
    (random() > 0.3), -- 70% active
    (floor(random() * 10) + 1)::smallint, -- min_transactions: 1-10
    (floor(random() * 90) + 7)::smallint, -- lookback_days: 7-96
    NOW() - (interval '1 day' * floor(random() * 60)), -- start_date: up to 60 days ago
    NOW() + (interval '1 day' * floor(random() * 90 + 1)) -- end_date: 1-90 days from now
FROM generate_series(1, 10000) AS i;
