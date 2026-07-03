CREATE TABLE IF NOT EXISTS catalog.regions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS catalog.user_qualifications (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reward.requirement_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    value_key TEXT NOT NULL,
    value_source TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO catalog.regions(id, name) VALUES
    ('TW', '台灣'),
    ('OVERSEAS', '海外')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, updated_at=now();

INSERT INTO catalog.user_qualifications(id, name) VALUES
    ('NEW_USER', '新戶'),
    ('PAYROLL', '薪轉戶'),
    ('DIGITAL_ACCOUNT', '數位帳戶戶'),
    ('REGISTERED', '已登錄')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, updated_at=now();

INSERT INTO reward.requirement_types(id, name, value_key, value_source, display_order) VALUES
    ('PAYMENT_METHOD', '支付方式', 'payment_method_codes', 'payment_methods', 10),
    ('CARD_NETWORK', '卡組織', 'network_codes', 'card_networks', 20),
    ('CARD_PLAN', '卡方案', 'card_plan_ids', 'card_plans', 30),
    ('CARD_PRODUCT', '信用卡', 'card_product_ids', 'card_products', 40),
    ('MERCHANT', '店家', 'merchant_ids', 'merchants', 50),
    ('MERCHANT_CATEGORY', '店家分類', 'category_ids', 'categories', 60),
    ('CONSUMPTION_CATEGORY', '消費分類', 'category_ids', 'categories', 70),
    ('AMOUNT', '金額', 'amount', 'number', 80),
    ('ACCOUNT_TIER', '帳戶等級', 'tiers', 'account_tiers', 90),
    ('USER_QUALIFICATION', '會員資格', 'qualification_codes', 'user_qualifications', 100),
    ('DATE_RANGE', '日期區間', 'date_range', 'date_range', 110),
    ('WEEKDAY', '星期', 'weekdays', 'weekdays', 120),
    ('TIME_RANGE', '時間區間', 'time_range', 'time_range', 130),
    ('CHANNEL', '通路', 'channels', 'channels', 140),
    ('REGION', '地區', 'regions', 'regions', 150),
    ('ACTION_REQUIRED', '必要操作', 'action_codes', 'actions', 160)
ON CONFLICT (id) DO UPDATE SET
    name=EXCLUDED.name,
    value_key=EXCLUDED.value_key,
    value_source=EXCLUDED.value_source,
    display_order=EXCLUDED.display_order,
    updated_at=now();
