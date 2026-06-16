ALTER TABLE reward_units
    ADD COLUMN symbol_position TEXT NOT NULL DEFAULT 'prefix'
        CHECK (symbol_position IN ('prefix','suffix')),
    ADD COLUMN twd_rate NUMERIC(18,6) NOT NULL DEFAULT 1
        CHECK (twd_rate > 0);

UPDATE reward_units
SET symbol_position = CASE WHEN code = 'cash_twd' THEN 'prefix' ELSE 'suffix' END;

INSERT INTO payment_methods(code,name,is_active,is_system)
VALUES ('any_payment','不限支付方式',true,true)
ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name,is_active=true,is_system=true;

INSERT INTO categories(code,name,is_active)
VALUES ('other','其他',true)
ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name,is_active=true;

UPDATE categories SET name='網路服務' WHERE code='online';

CREATE TABLE merchant_categories (
    merchant_code TEXT NOT NULL REFERENCES merchants(code) ON DELETE CASCADE,
    category_code TEXT NOT NULL REFERENCES categories(code),
    PRIMARY KEY (merchant_code,category_code)
);

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT DISTINCT bm.merchant_code,bc.category_code
FROM card_activity_benefit_merchants bm
JOIN card_activity_benefit_categories bc ON bc.benefit_id=bm.benefit_id
WHERE bc.category_code <> 'general'
ON CONFLICT DO NOTHING;

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT m.code,'other'
FROM merchants m
WHERE NOT EXISTS (
    SELECT 1 FROM merchant_categories mc WHERE mc.merchant_code=m.code
)
ON CONFLICT DO NOTHING;

WITH ranked AS (
    SELECT user_id,reward_unit_id,
        dense_rank() OVER (PARTITION BY user_id ORDER BY weight ASC) AS tier
    FROM reward_preferences
)
UPDATE reward_preferences p
SET weight = 1 + ((ranked.tier - 1) * 0.1)
FROM ranked
WHERE p.user_id=ranked.user_id AND p.reward_unit_id=ranked.reward_unit_id;
