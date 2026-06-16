INSERT INTO categories(code,name,is_active) VALUES
    ('insurance','保險',true)
ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name,is_active=true;

INSERT INTO reward_units(id,code,name,symbol,precision,is_system,symbol_position,twd_rate) VALUES
    ('00000000-0000-0000-0000-000000000107','pi_coin','P幣','P',0,false,'suffix',1)
ON CONFLICT (code) DO UPDATE SET
    name=EXCLUDED.name,
    symbol=EXCLUDED.symbol,
    precision=EXCLUDED.precision,
    symbol_position=EXCLUDED.symbol_position,
    twd_rate=EXCLUDED.twd_rate;

INSERT INTO banks(id,name,code,website_url) VALUES
    ('41000000-0000-0000-0000-000000000005','上海商銀','scsb','https://www.scsb.com.tw/'),
    ('41000000-0000-0000-0000-000000000006','聯邦銀行','ubot','https://www.ubot.com.tw/')
ON CONFLICT (name) DO UPDATE SET
    code=EXCLUDED.code,
    website_url=EXCLUDED.website_url,
    is_active=true;

INSERT INTO card_products(id,bank_id,name,account_tiers,is_active) VALUES
    ('51000000-0000-0000-0000-000000000008','41000000-0000-0000-0000-000000000004','Pi 拍錢包信用卡','{}',true),
    ('51000000-0000-0000-0000-000000000009','41000000-0000-0000-0000-000000000005','小小兵回饋卡','{}',true),
    ('51000000-0000-0000-0000-000000000010','41000000-0000-0000-0000-000000000006','LINE Bank聯名卡','{}',true)
ON CONFLICT (bank_id,name) DO UPDATE SET
    account_tiers=EXCLUDED.account_tiers,
    is_active=EXCLUDED.is_active;

INSERT INTO merchants(code,name,is_system) VALUES
    ('pchome','PChome',true),
    ('taobao','淘寶',true),
    ('coupang','酷澎',true),
    ('disney_plus','Disney+',true),
    ('catchplay','CatchPlay',true),
    ('spotify','Spotify',true),
    ('line_tv','LINE TV',true),
    ('kktv','KKTV',true),
    ('kkbox','KKBOX',true),
    ('blizzard','Blizzard',true),
    ('garena','Garena',true),
    ('universal_studios','環球影城',true),
    ('funkang_supermarket','楓康超市',true),
    ('sihuwei','喜互惠',true),
    ('simple_mart','美廉社',true),
    ('a_mart','愛買',true),
    ('px_mart_plus','大全聯',true),
    ('familymart','全家',true)
ON CONFLICT (code) DO UPDATE SET
    name=EXCLUDED.name,
    is_active=true,
    is_system=true;

INSERT INTO merchant_aliases(merchant_code,alias)
SELECT code,name FROM merchants
WHERE code IN (
    'pchome','taobao','coupang','disney_plus','catchplay','spotify','line_tv','kktv','kkbox',
    'blizzard','garena','universal_studios','funkang_supermarket','sihuwei','simple_mart','a_mart','px_mart_plus','familymart'
)
ON CONFLICT DO NOTHING;

INSERT INTO merchant_categories(merchant_code,category_code) VALUES
    ('pchome','online'),
    ('taobao','online'),
    ('coupang','online'),
    ('disney_plus','online'),
    ('disney_plus','entertainment'),
    ('catchplay','online'),
    ('catchplay','entertainment'),
    ('spotify','online'),
    ('spotify','entertainment'),
    ('line_tv','online'),
    ('line_tv','entertainment'),
    ('kktv','online'),
    ('kktv','entertainment'),
    ('kkbox','online'),
    ('kkbox','entertainment'),
    ('blizzard','online'),
    ('blizzard','entertainment'),
    ('garena','online'),
    ('garena','entertainment'),
    ('universal_studios','travel'),
    ('universal_studios','entertainment'),
    ('funkang_supermarket','grocery'),
    ('sihuwei','grocery'),
    ('simple_mart','grocery'),
    ('a_mart','grocery'),
    ('px_mart_plus','grocery'),
    ('familymart','grocery')
ON CONFLICT DO NOTHING;

INSERT INTO card_activities(id,card_product_id,name,start_date,end_date,is_active,source_url,verified_at) VALUES
    ('61000000-0000-0000-0000-000000000008','51000000-0000-0000-0000-000000000008','2026 Pi 拍錢包信用卡回饋','2026-03-01','2026-08-31',true,'https://www.piapp.com.tw/picard/','2026-06-16'),
    ('61000000-0000-0000-0000-000000000009','51000000-0000-0000-0000-000000000009','2026 小小兵回饋卡優惠','2026-01-01','2026-12-31',true,'https://www.scsb.com.tw/content/card/card_1110222.html','2026-06-16'),
    ('61000000-0000-0000-0000-000000000010','51000000-0000-0000-0000-000000000010','2025-2026 LINE Bank聯名卡回饋','2025-08-01','2026-07-31',true,'https://activity.ubot.com.tw/LINEBankCard/index.htm','2026-06-16')
ON CONFLICT (id) DO UPDATE SET
    name=EXCLUDED.name,
    start_date=EXCLUDED.start_date,
    end_date=EXCLUDED.end_date,
    is_active=EXCLUDED.is_active,
    source_url=EXCLUDED.source_url,
    verified_at=EXCLUDED.verified_at;

INSERT INTO card_activity_benefits
    (id,card_activity_id,reward_unit_id,name,rate,monthly_cap,stack_group,priority,action_required,action_message)
VALUES
    ('71000000-0000-0000-0000-000000000024','61000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000107','一般消費 1% P幣',0.01,NULL,'pi_wallet_base',10,'account_setup','需註冊 Pi 拍錢包 App、綁定 Pi 卡並申請玉山信用卡帳單 e 化，始享 P幣回饋'),
    ('71000000-0000-0000-0000-000000000026','61000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000101','國內消費回饋 1.234%',0.01234,NULL,'minions_base',10,'none',''),
    ('71000000-0000-0000-0000-000000000027','61000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000101','國外消費回饋 2.234%',0.02234,NULL,'minions_base',20,'none',''),
    ('71000000-0000-0000-0000-000000000028','61000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000101','全球環球影城樂園最高 10%',0.10,500,'minions_base',30,'none','每期帳單回饋上限 NT$500'),
    ('71000000-0000-0000-0000-000000000029','61000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000101','生活購物最高 5%',0.05,200,'minions_base',40,'account_setup','活動期間 2026/1/1 至 2026/6/30；須辦理上海商銀帳戶自動扣繳信用卡款，每期帳單回饋上限 NT$200'),
    ('71000000-0000-0000-0000-000000000030','61000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000101','國內消費 1% 刷卡金',0.01,NULL,'linebank_base',10,'none',''),
    ('71000000-0000-0000-0000-000000000031','61000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000101','國外消費 2.5% 刷卡金',0.025,NULL,'linebank_base',20,'none',''),
    ('71000000-0000-0000-0000-000000000032','61000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000101','保險 1% 刷卡金',0.01,NULL,'linebank_base',30,'none',''),
    ('71000000-0000-0000-0000-000000000033','61000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000101','指定影音/網購/遊戲加碼 3%',0.03,300,'linebank_bonus_online_entertainment',40,'none','指定影音、網購、遊戲通路加碼，歸戶每月上限 NT$300')
ON CONFLICT (id) DO UPDATE SET
    name=EXCLUDED.name,
    rate=EXCLUDED.rate,
    monthly_cap=EXCLUDED.monthly_cap,
    stack_group=EXCLUDED.stack_group,
    priority=EXCLUDED.priority,
    action_required=EXCLUDED.action_required,
    action_message=EXCLUDED.action_message,
    updated_at=now();

INSERT INTO card_activity_benefit_categories(benefit_id,category_code) VALUES
    ('71000000-0000-0000-0000-000000000024','general'),
    ('71000000-0000-0000-0000-000000000026','general'),
    ('71000000-0000-0000-0000-000000000027','overseas'),
    ('71000000-0000-0000-0000-000000000028','travel'),
    ('71000000-0000-0000-0000-000000000028','entertainment'),
    ('71000000-0000-0000-0000-000000000029','grocery'),
    ('71000000-0000-0000-0000-000000000030','general'),
    ('71000000-0000-0000-0000-000000000031','overseas'),
    ('71000000-0000-0000-0000-000000000032','insurance'),
    ('71000000-0000-0000-0000-000000000033','online'),
    ('71000000-0000-0000-0000-000000000033','entertainment')
ON CONFLICT DO NOTHING;

INSERT INTO card_activity_benefit_merchants(benefit_id,merchant_code)
SELECT v.benefit_id::uuid,m.code
FROM (VALUES
    ('71000000-0000-0000-0000-000000000028','環球影城'),
    ('71000000-0000-0000-0000-000000000029','楓康超市'),
    ('71000000-0000-0000-0000-000000000029','喜互惠'),
    ('71000000-0000-0000-0000-000000000029','美廉社'),
    ('71000000-0000-0000-0000-000000000029','家樂福'),
    ('71000000-0000-0000-0000-000000000029','愛買'),
    ('71000000-0000-0000-0000-000000000029','全聯'),
    ('71000000-0000-0000-0000-000000000029','大全聯'),
    ('71000000-0000-0000-0000-000000000029','7-ELEVEN'),
    ('71000000-0000-0000-0000-000000000029','全家'),
    ('71000000-0000-0000-0000-000000000033','momo'),
    ('71000000-0000-0000-0000-000000000033','PChome'),
    ('71000000-0000-0000-0000-000000000033','淘寶'),
    ('71000000-0000-0000-0000-000000000033','蝦皮'),
    ('71000000-0000-0000-0000-000000000033','酷澎'),
    ('71000000-0000-0000-0000-000000000033','Netflix'),
    ('71000000-0000-0000-0000-000000000033','Disney+'),
    ('71000000-0000-0000-0000-000000000033','CatchPlay'),
    ('71000000-0000-0000-0000-000000000033','Spotify'),
    ('71000000-0000-0000-0000-000000000033','LINE TV'),
    ('71000000-0000-0000-0000-000000000033','KKTV'),
    ('71000000-0000-0000-0000-000000000033','KKBOX'),
    ('71000000-0000-0000-0000-000000000033','Steam'),
    ('71000000-0000-0000-0000-000000000033','PlayStation'),
    ('71000000-0000-0000-0000-000000000033','Nintendo'),
    ('71000000-0000-0000-0000-000000000033','Blizzard'),
    ('71000000-0000-0000-0000-000000000033','Garena')
) AS v(benefit_id,merchant_name)
JOIN card_activity_benefits b ON b.id=v.benefit_id::uuid
JOIN merchants m ON lower(m.name)=lower(v.merchant_name)
ON CONFLICT DO NOTHING;
