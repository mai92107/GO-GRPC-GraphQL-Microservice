INSERT INTO categories(code, name) VALUES
    ('overseas', '海外消費'),
    ('sports', '運動健身'),
    ('travel', '旅遊'),
    ('entertainment', '娛樂'),
    ('grocery', '量販超市')
ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name;

INSERT INTO reward_units(id, code, name, symbol, precision, is_system) VALUES
    ('00000000-0000-0000-0000-000000000104', 'sinopac_points', '永豐豐點', '點', 2, false),
    ('00000000-0000-0000-0000-000000000105', 'openpoint', 'OPENPOINT', '點', 2, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO banks(id, name, code, website_url) VALUES
    ('41000000-0000-0000-0000-000000000001', '永豐銀行', 'sinopac', 'https://bank.sinopac.com/'),
    ('41000000-0000-0000-0000-000000000002', '台新銀行', 'taishin', 'https://www.taishinbank.com.tw/'),
    ('41000000-0000-0000-0000-000000000003', '中國信託', 'ctbc', 'https://www.ctbcbank.com/')
ON CONFLICT (name) DO UPDATE SET website_url=EXCLUDED.website_url;

INSERT INTO card_products(id, bank_id, name, account_tiers) VALUES
    ('51000000-0000-0000-0000-000000000001','41000000-0000-0000-0000-000000000001','DAWHO 現金回饋信用卡',ARRAY['大戶','大戶Plus']),
    ('51000000-0000-0000-0000-000000000002','41000000-0000-0000-0000-000000000001','SPORT 卡','{}'),
    ('51000000-0000-0000-0000-000000000003','41000000-0000-0000-0000-000000000002','Richart 卡','{}'),
    ('51000000-0000-0000-0000-000000000004','41000000-0000-0000-0000-000000000003','uniopen 聯名卡','{}')
ON CONFLICT (bank_id,name) DO UPDATE SET account_tiers=EXCLUDED.account_tiers;

INSERT INTO card_activities(id,card_product_id,name,start_date,end_date,is_active,source_url,verified_at) VALUES
    ('61000000-0000-0000-0000-000000000001','51000000-0000-0000-0000-000000000001','2026 上半年核心權益','2026-01-01','2026-06-30',true,'https://dawho.tw/hot/2026h1offer/','2026-06-14'),
    ('61000000-0000-0000-0000-000000000002','51000000-0000-0000-0000-000000000002','2026 上半年核心權益','2026-01-01','2026-06-30',true,'https://bank.sinopac.com/sinopacBT/webevents/2008_SportCard2/index.html?Branch=SP2026H1','2026-06-14'),
    ('61000000-0000-0000-0000-000000000003','51000000-0000-0000-0000-000000000003','2026 上半年核心權益','2026-01-01','2026-06-30',true,'https://www.taishinbank.com.tw/','2026-06-14'),
    ('61000000-0000-0000-0000-000000000004','51000000-0000-0000-0000-000000000004','2026 上半年核心權益','2026-01-01','2026-06-30',true,'https://www.ctbcbank.com/content/twrbo/zh_tw/cc_index/cc_feedback/cc_feedback_other_index/cc_feedback_uniopen.html','2026-06-14')
ON CONFLICT (id) DO NOTHING;

INSERT INTO card_activity_benefits(id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group,priority,required_account_tiers,action_required,action_message) VALUES
    ('71000000-0000-0000-0000-000000000001','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','大戶基本回饋',0.01,NULL,'base',10,ARRAY['大戶','大戶Plus'],'none',''),
    ('71000000-0000-0000-0000-000000000002','61000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','大戶Plus 加碼',0.02,600,'tier_bonus',20,ARRAY['大戶Plus'],'none',''),
    ('71000000-0000-0000-0000-000000000003','61000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000104','一般消費豐點',0.01,NULL,'base',10,'{}','none',''),
    ('71000000-0000-0000-0000-000000000004','61000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000104','行動支付加碼',0.04,300,'payment_bonus',20,'{}','registration','需先完成當期活動登錄'),
    ('71000000-0000-0000-0000-000000000005','61000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000101','一般消費回饋',0.003,NULL,'base',10,'{}','none',''),
    ('71000000-0000-0000-0000-000000000006','61000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000101','Richart 指定方案加碼',0.035,1000,'richart_plan',20,'{}','app_switch','需於 Richart Life APP 切換至符合消費情境的方案'),
    ('71000000-0000-0000-0000-000000000007','61000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000105','國內一般消費',0.01,NULL,'base',10,'{}','none',''),
    ('71000000-0000-0000-0000-000000000008','61000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000105','海外消費加碼',0.10,500,'channel_bonus',20,'{}','none',''),
    ('71000000-0000-0000-0000-000000000009','61000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000105','統一集團加碼',0.10,500,'channel_bonus',20,'{}','none','')
ON CONFLICT (id) DO NOTHING;

INSERT INTO card_activity_benefit_categories(benefit_id,category_code) VALUES
    ('71000000-0000-0000-0000-000000000001','general'),
    ('71000000-0000-0000-0000-000000000002','general'),
    ('71000000-0000-0000-0000-000000000003','general'),
    ('71000000-0000-0000-0000-000000000004','general'),
    ('71000000-0000-0000-0000-000000000005','general'),
    ('71000000-0000-0000-0000-000000000006','dining'),
    ('71000000-0000-0000-0000-000000000006','online'),
    ('71000000-0000-0000-0000-000000000006','travel'),
    ('71000000-0000-0000-0000-000000000006','entertainment'),
    ('71000000-0000-0000-0000-000000000007','general'),
    ('71000000-0000-0000-0000-000000000008','overseas'),
    ('71000000-0000-0000-0000-000000000009','grocery')
ON CONFLICT DO NOTHING;

UPDATE card_activity_benefits
SET payment_methods=ARRAY['Apple Pay','Google Pay','Samsung Pay','LINE Pay','街口支付']
WHERE id='71000000-0000-0000-0000-000000000004';

INSERT INTO card_activity_benefit_merchants(benefit_id,merchant_keyword) VALUES
    ('71000000-0000-0000-0000-000000000009','7-ELEVEN'),
    ('71000000-0000-0000-0000-000000000009','星巴克'),
    ('71000000-0000-0000-0000-000000000009','家樂福')
ON CONFLICT DO NOTHING;
