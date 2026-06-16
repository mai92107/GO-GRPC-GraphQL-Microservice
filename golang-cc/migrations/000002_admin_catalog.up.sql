CREATE TABLE banks (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE,
    website_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE card_products (
    id UUID PRIMARY KEY,
    bank_id UUID NOT NULL REFERENCES banks(id),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (bank_id, name)
);

CREATE TABLE member_cards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_product_id UUID NOT NULL REFERENCES card_products(id),
    nickname TEXT,
    last_four CHAR(4),
    is_active BOOLEAN NOT NULL DEFAULT true,
    statement_day SMALLINT CHECK (statement_day BETWEEN 1 AND 31),
    payment_due_day SMALLINT CHECK (payment_due_day BETWEEN 1 AND 31),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, card_product_id),
    UNIQUE (id, user_id),
    CHECK (last_four IS NULL OR last_four ~ '^[0-9]{4}$')
);

CREATE TABLE card_activities (
    id UUID PRIMARY KEY,
    card_product_id UUID NOT NULL REFERENCES card_products(id),
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    name TEXT NOT NULL,
    rate NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    monthly_cap NUMERIC(18,6) CHECK (monthly_cap > 0),
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);

CREATE TABLE card_activity_categories (
    card_activity_id UUID NOT NULL REFERENCES card_activities(id) ON DELETE CASCADE,
    category_code TEXT NOT NULL REFERENCES categories(code),
    PRIMARY KEY (card_activity_id, category_code)
);

ALTER TABLE categories ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE reward_units ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Preserve existing demo/member data by promoting each distinct issuer into a bank,
-- each owned card into a catalog product, and each rule into a shared activity.
INSERT INTO banks (id, name)
SELECT md5('bank:' || issuer)::uuid, issuer FROM cards GROUP BY issuer
ON CONFLICT (name) DO NOTHING;

INSERT INTO card_products (id, bank_id, name, is_active, created_at, updated_at)
SELECT c.id, b.id, c.name, c.is_active, c.created_at, c.updated_at
FROM cards c JOIN banks b ON b.name = c.issuer
ON CONFLICT DO NOTHING;

INSERT INTO member_cards (id, user_id, card_product_id, nickname, last_four, is_active, created_at, updated_at)
SELECT c.id, c.user_id, c.id, c.name, c.last_four, c.is_active, c.created_at, c.updated_at
FROM cards c
ON CONFLICT DO NOTHING;

INSERT INTO card_activities (id, card_product_id, reward_unit_id, name, rate, monthly_cap, start_date, end_date, is_active, created_at, updated_at)
SELECT r.id, r.card_id, r.reward_unit_id, r.name, r.rate, r.monthly_cap, r.start_date, r.end_date, r.is_active, r.created_at, r.updated_at
FROM reward_rules r
ON CONFLICT DO NOTHING;

INSERT INTO card_activity_categories (card_activity_id, category_code)
SELECT reward_rule_id, category_code FROM reward_rule_categories
ON CONFLICT DO NOTHING;

ALTER TABLE transactions DROP CONSTRAINT transactions_card_id_user_id_fkey;
ALTER TABLE transactions ADD CONSTRAINT transactions_member_card_fkey
    FOREIGN KEY (card_id, user_id) REFERENCES member_cards(id, user_id);

ALTER TABLE reward_allocations DROP CONSTRAINT reward_allocations_reward_rule_id_fkey;
ALTER TABLE reward_allocations ADD CONSTRAINT reward_allocations_activity_fkey
    FOREIGN KEY (reward_rule_id) REFERENCES card_activities(id);

DROP TABLE reward_rule_categories;
DROP TABLE reward_rules;
DROP TABLE cards;

CREATE INDEX card_products_bank_active_idx ON card_products(bank_id, is_active);
CREATE INDEX member_cards_user_active_idx ON member_cards(user_id, is_active);
CREATE INDEX card_activities_product_active_idx ON card_activities(card_product_id, is_active);
CREATE INDEX card_activities_dates_idx ON card_activities(start_date, end_date);
