UPDATE card_activities
SET source_url = 'https://dawho.tw/about/card/',
    verified_at = DATE '2026-06-16'
WHERE id = '61000000-0000-0000-0000-000000000001';

INSERT INTO card_activity_benefits
    (id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group,priority,required_account_tiers,action_required,action_message)
VALUES
    ('71000000-0000-0000-0000-000000000022','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','悠遊卡自動加值 3%',0.03,100,'exclusive:dawho_easycard_autoload',30,ARRAY['大戶'],'account_setup','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費'),
    ('71000000-0000-0000-0000-000000000023','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','大戶Plus 悠遊卡自動加值 5%',0.05,500,'exclusive:dawho_easycard_autoload',30,ARRAY['大戶Plus'],'account_setup','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費')
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

INSERT INTO card_activity_benefit_categories(benefit_id,category_code) VALUES
    ('71000000-0000-0000-0000-000000000022','general'),
    ('71000000-0000-0000-0000-000000000023','general')
ON CONFLICT DO NOTHING;

INSERT INTO card_activity_benefit_payment_methods(benefit_id,payment_method_code) VALUES
    ('71000000-0000-0000-0000-000000000022','easycard'),
    ('71000000-0000-0000-0000-000000000023','easycard')
ON CONFLICT DO NOTHING;
