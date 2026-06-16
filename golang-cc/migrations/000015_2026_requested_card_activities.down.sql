DELETE FROM card_activities WHERE id IN (
    '61000000-0000-0000-0000-000000000008',
    '61000000-0000-0000-0000-000000000009',
    '61000000-0000-0000-0000-000000000010'
);

DELETE FROM card_products WHERE id IN (
    '51000000-0000-0000-0000-000000000008',
    '51000000-0000-0000-0000-000000000009',
    '51000000-0000-0000-0000-000000000010'
);

DELETE FROM merchant_categories
WHERE merchant_code IN (
    'pchome','taobao','coupang','disney_plus','catchplay','spotify','line_tv','kktv','kkbox',
    'blizzard','garena','universal_studios','funkang_supermarket','sihuwei','simple_mart','a_mart','px_mart_plus','familymart'
);

DELETE FROM merchants
WHERE code IN (
    'pchome','taobao','coupang','disney_plus','catchplay','spotify','line_tv','kktv','kkbox',
    'blizzard','garena','universal_studios','funkang_supermarket','sihuwei','simple_mart','a_mart','px_mart_plus','familymart'
);

DELETE FROM banks WHERE id IN (
    '41000000-0000-0000-0000-000000000005',
    '41000000-0000-0000-0000-000000000006'
);

DELETE FROM reward_units WHERE id='00000000-0000-0000-0000-000000000107';

DELETE FROM categories WHERE code='insurance';
