-- Seed transactions (~2,000,000)
-- Distributed across ~50,000 users, random merchants, random MCCs

-- Pre-generate arrays for fast random selection
DO $$
DECLARE
    batch_size INT := 100000;
    total_transactions INT := 2000000;
    current_batch INT := 0;
    merchant_ids UUID[];
    mccs VARCHAR(4)[] := ARRAY['5411', '5541', '5812', '5912', '5999', '7011', '4121', '5311', '5651', '5732'];
    merchant_count INT;
    mcc_count INT := 10;
BEGIN
    -- Load merchant IDs into array
    SELECT array_agg(id) INTO merchant_ids FROM merchants;
    merchant_count := array_length(merchant_ids, 1);

    WHILE current_batch * batch_size < total_transactions LOOP
        INSERT INTO transactions (id, merchant_id, user_id, mcc, amount_cents, approved_at)
        SELECT
            gen_random_uuid(),
            merchant_ids[1 + floor(random() * merchant_count)::int],
            md5(floor(random() * 50000)::text)::uuid,
            mccs[1 + floor(random() * mcc_count)::int],
            (floor(random() * 100000) + 100)::int,
            NOW() - (interval '1 minute' * floor(random() * 525600))
        FROM generate_series(1, batch_size);

        current_batch := current_batch + 1;
        RAISE NOTICE 'Inserted batch % of %', current_batch, ceil(total_transactions::float / batch_size);
    END LOOP;
END $$;

-- Update statistics for query planner
ANALYZE transactions;
ANALYZE offers;
ANALYZE offers_mccs;
ANALYZE merchants;
