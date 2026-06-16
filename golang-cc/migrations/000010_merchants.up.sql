CREATE TABLE merchants (
    code TEXT PRIMARY KEY CHECK (code ~ '^[a-z0-9_]+$'),
    name TEXT UNIQUE NOT NULL CHECK (btrim(name) <> ''),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE merchant_aliases (
    merchant_code TEXT NOT NULL REFERENCES merchants(code) ON DELETE CASCADE,
    alias TEXT NOT NULL CHECK (btrim(alias) <> ''),
    PRIMARY KEY (merchant_code, alias)
);

INSERT INTO merchants(code,name,is_system)
SELECT 'merchant_' || substr(md5(lower(btrim(merchant_keyword))),1,16), min(btrim(merchant_keyword)), true
FROM card_activity_benefit_merchants
GROUP BY lower(btrim(merchant_keyword))
ON CONFLICT DO NOTHING;

INSERT INTO merchant_aliases(merchant_code,alias)
SELECT m.code, m.name FROM merchants m
ON CONFLICT DO NOTHING;

INSERT INTO merchant_aliases(merchant_code,alias)
SELECT code, '全聯福利中心' FROM merchants WHERE name='全聯'
ON CONFLICT DO NOTHING;

CREATE TABLE card_activity_benefit_merchants_v2 (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    merchant_code TEXT NOT NULL REFERENCES merchants(code),
    PRIMARY KEY (benefit_id,merchant_code)
);

INSERT INTO card_activity_benefit_merchants_v2(benefit_id,merchant_code)
SELECT bm.benefit_id,m.code
FROM card_activity_benefit_merchants bm
JOIN merchants m ON lower(m.name)=lower(btrim(bm.merchant_keyword))
ON CONFLICT DO NOTHING;

DROP TABLE card_activity_benefit_merchants;
ALTER TABLE card_activity_benefit_merchants_v2 RENAME TO card_activity_benefit_merchants;

ALTER TABLE transactions
    ADD COLUMN merchant_code TEXT REFERENCES merchants(code);

CREATE INDEX transactions_merchant_code_idx ON transactions(merchant_code);
