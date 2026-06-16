DELETE FROM merchant_categories;

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT m.code,v.category_code
FROM (VALUES
    ('7-ELEVEN','grocery'),
    ('Agoda','online'),
    ('Agoda','travel'),
    ('Booking.com','online'),
    ('Booking.com','travel'),
    ('ChatGPT','online'),
    ('Gemini','online'),
    ('Netflix','online'),
    ('Netflix','entertainment'),
    ('Nintendo','online'),
    ('Nintendo','entertainment'),
    ('PlayStation','online'),
    ('PlayStation','entertainment'),
    ('Steam','online'),
    ('Steam','entertainment'),
    ('UNIQLO','other'),
    ('Uber','transport'),
    ('momo','online'),
    ('全聯','grocery'),
    ('台灣中油','transport'),
    ('家樂福','grocery'),
    ('新光三越','other'),
    ('星巴克','dining'),
    ('蝦皮','online'),
    ('高鐵','transport'),
    ('高鐵','travel')
) AS v(merchant_name,category_code)
JOIN merchants m ON lower(m.name)=lower(v.merchant_name)
ON CONFLICT DO NOTHING;

INSERT INTO merchant_categories(merchant_code,category_code)
SELECT m.code,'other'
FROM merchants m
WHERE NOT EXISTS (
    SELECT 1 FROM merchant_categories mc WHERE mc.merchant_code=m.code
)
ON CONFLICT DO NOTHING;

DELETE FROM card_activity_benefit_merchants;

INSERT INTO card_activity_benefit_merchants(benefit_id,merchant_code)
SELECT v.benefit_id::uuid,m.code
FROM (VALUES
    ('30000000-0000-0000-0000-000000000001','全聯'),
    ('71000000-0000-0000-0000-000000000009','7-ELEVEN'),
    ('71000000-0000-0000-0000-000000000009','星巴克'),
    ('71000000-0000-0000-0000-000000000009','家樂福'),
    ('71000000-0000-0000-0000-000000000012','Netflix'),
    ('71000000-0000-0000-0000-000000000012','ChatGPT'),
    ('71000000-0000-0000-0000-000000000012','Gemini'),
    ('71000000-0000-0000-0000-000000000012','Steam'),
    ('71000000-0000-0000-0000-000000000012','Nintendo'),
    ('71000000-0000-0000-0000-000000000012','PlayStation'),
    ('71000000-0000-0000-0000-000000000015','台灣中油'),
    ('71000000-0000-0000-0000-000000000015','高鐵'),
    ('71000000-0000-0000-0000-000000000015','Uber'),
    ('71000000-0000-0000-0000-000000000015','新光三越'),
    ('71000000-0000-0000-0000-000000000015','momo'),
    ('71000000-0000-0000-0000-000000000015','蝦皮'),
    ('71000000-0000-0000-0000-000000000015','家樂福'),
    ('71000000-0000-0000-0000-000000000015','UNIQLO'),
    ('71000000-0000-0000-0000-000000000015','Booking.com'),
    ('71000000-0000-0000-0000-000000000015','Agoda')
) AS v(benefit_id,merchant_name)
JOIN card_activity_benefits b ON b.id=v.benefit_id::uuid
JOIN merchants m ON lower(m.name)=lower(v.merchant_name)
ON CONFLICT DO NOTHING;
