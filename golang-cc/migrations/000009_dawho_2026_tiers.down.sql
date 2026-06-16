DELETE FROM card_activity_benefits
WHERE id IN (
    '71000000-0000-0000-0000-000000000020',
    '71000000-0000-0000-0000-000000000021'
);

UPDATE card_products
SET account_tiers = ARRAY['大戶','大戶Plus']
WHERE id = '51000000-0000-0000-0000-000000000001';

UPDATE card_activities
SET source_url = 'https://dawho.tw/hot/2026h1offer/',
    verified_at = DATE '2026-06-14'
WHERE id = '61000000-0000-0000-0000-000000000001';

UPDATE card_activity_benefits
SET name = '大戶基本回饋',
    rate = 0.01,
    monthly_cap = NULL,
    stack_group = 'base',
    priority = 10,
    required_account_tiers = ARRAY['大戶','大戶Plus'],
    action_required = 'none',
    action_message = '',
    updated_at = now()
WHERE id = '71000000-0000-0000-0000-000000000001';

UPDATE card_activity_benefits
SET name = '大戶Plus 加碼',
    rate = 0.02,
    monthly_cap = 600,
    stack_group = 'tier_bonus',
    priority = 20,
    required_account_tiers = ARRAY['大戶Plus'],
    action_required = 'none',
    action_message = '',
    updated_at = now()
WHERE id = '71000000-0000-0000-0000-000000000002';

DELETE FROM card_activity_benefit_categories
WHERE benefit_id IN (
    '71000000-0000-0000-0000-000000000001',
    '71000000-0000-0000-0000-000000000002'
);

INSERT INTO card_activity_benefit_categories(benefit_id,category_code) VALUES
    ('71000000-0000-0000-0000-000000000001','general'),
    ('71000000-0000-0000-0000-000000000002','general')
ON CONFLICT DO NOTHING;
