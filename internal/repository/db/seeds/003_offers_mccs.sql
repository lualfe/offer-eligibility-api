-- Common MCCs for realistic data
-- 5411: Supermarkets, 5541: Gas stations, 5812: Restaurants
-- 5912: Pharmacies, 5999: Misc stores, 7011: Hotels
-- 4121: Taxis/Rideshare, 5311: Department stores, 5651: Clothing, 5732: Electronics

-- Create temp table with MCCs
CREATE TEMP TABLE temp_mccs (mcc VARCHAR(4));
INSERT INTO temp_mccs VALUES
    ('5411'), ('5541'), ('5812'), ('5912'), ('5999'),
    ('7011'), ('4121'), ('5311'), ('5651'), ('5732');

-- Each offer gets 1-4 random MCCs (only ~40% of offers have MCC restrictions)
INSERT INTO offers_mccs (offer_id, mcc)
SELECT DISTINCT
    o.id,
    t.mcc
FROM offers o
CROSS JOIN LATERAL (
    SELECT mcc
    FROM temp_mccs
    ORDER BY random()
    LIMIT (floor(random() * 4) + 1)::int
) t
WHERE random() > 0.6 -- ~40% of offers have MCC restrictions
ON CONFLICT (offer_id, mcc) DO NOTHING;

DROP TABLE temp_mccs;
