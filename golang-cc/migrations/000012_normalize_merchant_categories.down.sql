DELETE FROM merchant_categories;

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT DISTINCT bm.merchant_code,bc.category_code
FROM card_activity_benefit_merchants bm
JOIN card_activity_benefit_categories bc ON bc.benefit_id=bm.benefit_id
WHERE bc.category_code <> 'general'
ON CONFLICT DO NOTHING;

DELETE FROM card_activity_benefit_merchants;

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT m.code,'other'
FROM merchants m
WHERE NOT EXISTS (
    SELECT 1 FROM merchant_categories mc WHERE mc.merchant_code=m.code
)
ON CONFLICT DO NOTHING;
