CREATE TABLE payment_methods (
    code TEXT PRIMARY KEY CHECK (btrim(code) <> ''),
    name TEXT UNIQUE NOT NULL CHECK (btrim(name) <> ''),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO payment_methods(code,name,is_system) VALUES
    ('physical_card','實體信用卡',true),
    ('online_card','線上刷卡',true),
    ('apple_pay','Apple Pay',true),
    ('google_pay','Google Pay',true),
    ('samsung_pay','Samsung Pay',true),
    ('line_pay','LINE Pay',true),
    ('jkopay','街口支付',true),
    ('easy_wallet','悠遊付',true),
    ('px_pay_plus','全支付',true),
    ('fami_pay','全盈+PAY',true),
    ('ipass_money','iPASS MONEY',true),
    ('icash_pay','icash Pay',true),
    ('open_wallet','OPEN錢包',true),
    ('esun_wallet','玉山Wallet',true),
    ('easycard','悠遊卡功能',true),
    ('icash','icash 功能',true),
    ('ipass','一卡通功能',true);

CREATE TABLE card_activity_benefit_payment_methods (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    payment_method_code TEXT NOT NULL REFERENCES payment_methods(code),
    PRIMARY KEY (benefit_id,payment_method_code)
);

INSERT INTO card_activity_benefit_payment_methods(benefit_id,payment_method_code)
SELECT b.id, p.code
FROM card_activity_benefits b
JOIN LATERAL unnest(b.payment_methods) value ON true
JOIN payment_methods p ON lower(p.name)=lower(value)
ON CONFLICT DO NOTHING;

ALTER TABLE card_activity_benefits DROP COLUMN payment_methods;

ALTER TABLE transactions
    ADD COLUMN payment_method_code TEXT NOT NULL DEFAULT 'physical_card'
    REFERENCES payment_methods(code);

ALTER TABLE transactions ALTER COLUMN payment_method_code DROP DEFAULT;
CREATE INDEX transactions_payment_method_idx ON transactions(payment_method_code);

ALTER TABLE card_activity_benefits DROP CONSTRAINT card_activity_benefits_action_required_check;
ALTER TABLE card_activity_benefits ADD CONSTRAINT card_activity_benefits_action_required_check
    CHECK (action_required IN ('none','registration','app_switch','account_setup'));
