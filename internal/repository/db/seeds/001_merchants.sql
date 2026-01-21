-- Seed merchants (~100)
INSERT INTO merchants (id, mechant_name)
SELECT
    gen_random_uuid(),
    'Merchant ' || i
FROM generate_series(1, 100) AS i
ON CONFLICT (id) DO NOTHING;
