UPDATE card_products
SET account_tiers = ARRAY['大大','大戶','大戶Plus']
WHERE id = '51000000-0000-0000-0000-000000000001';

UPDATE card_activities
SET source_url = 'https://bank.sinopac.com/sinopacBT/webevents/dawho/about/card/',
    verified_at = DATE '2026-06-15'
WHERE id = '61000000-0000-0000-0000-000000000001';

UPDATE card_activity_benefits
SET name = '國內一般消費 1%',
    rate = 0.01,
    monthly_cap = NULL,
    stack_group = 'dawho_base',
    priority = 10,
    required_account_tiers = ARRAY['大大','大戶','大戶Plus'],
    action_required = 'none',
    action_message = '',
    updated_at = now()
WHERE id = '71000000-0000-0000-0000-000000000001';

UPDATE card_activity_benefits
SET name = '大戶指定任務加碼 2.5%',
    rate = 0.025,
    monthly_cap = 400,
    stack_group = 'dawho_tier_bonus',
    priority = 20,
    required_account_tiers = ARRAY['大戶'],
    action_required = 'account_setup',
    action_message = '需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單',
    updated_at = now()
WHERE id = '71000000-0000-0000-0000-000000000002';

INSERT INTO card_activity_benefits
    (id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group,priority,required_account_tiers,action_required,action_message)
VALUES
    ('71000000-0000-0000-0000-000000000020','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','國外一般消費 2%',0.02,NULL,'dawho_base',10,ARRAY['大大','大戶','大戶Plus'],'none',''),
    ('71000000-0000-0000-0000-000000000021','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','大戶Plus 指定任務加碼 4%',0.04,1000,'dawho_tier_bonus',20,ARRAY['大戶Plus'],'account_setup','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    rate = EXCLUDED.rate,
    monthly_cap = EXCLUDED.monthly_cap,
    stack_group = EXCLUDED.stack_group,
    priority = EXCLUDED.priority,
    required_account_tiers = EXCLUDED.required_account_tiers,
    action_required = EXCLUDED.action_required,
    action_message = EXCLUDED.action_message,
    updated_at = now();

DELETE FROM card_activity_benefit_categories
WHERE benefit_id IN (
    '71000000-0000-0000-0000-000000000001',
    '71000000-0000-0000-0000-000000000002',
    '71000000-0000-0000-0000-000000000020',
    '71000000-0000-0000-0000-000000000021'
);

INSERT INTO card_activity_benefit_categories(benefit_id,category_code)
SELECT '71000000-0000-0000-0000-000000000001', code
FROM categories
WHERE code NOT IN ('general','overseas')
ON CONFLICT DO NOTHING;

INSERT INTO card_activity_benefit_categories(benefit_id,category_code) VALUES
    ('71000000-0000-0000-0000-000000000002','general'),
    ('71000000-0000-0000-0000-000000000020','overseas'),
    ('71000000-0000-0000-0000-000000000021','general')
ON CONFLICT DO NOTHING;
