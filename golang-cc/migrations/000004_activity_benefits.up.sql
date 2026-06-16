CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE card_products
    ADD COLUMN account_tiers TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE member_cards
    ADD COLUMN account_tier TEXT;

ALTER TABLE card_activities
    ADD COLUMN source_url TEXT,
    ADD COLUMN verified_at DATE,
    ADD COLUMN shared_monthly_caps JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE card_activities
SET start_date = COALESCE(start_date, DATE '2000-01-01'),
    end_date = COALESCE(end_date, DATE '2099-12-31');

ALTER TABLE card_activities
    ALTER COLUMN start_date SET NOT NULL,
    ALTER COLUMN end_date SET NOT NULL;

ALTER TABLE card_activities
    ADD CONSTRAINT card_activities_no_overlapping_periods
    EXCLUDE USING gist (
        card_product_id WITH =,
        daterange(start_date, end_date, '[]') WITH &&
    ) WHERE (is_active);

CREATE TABLE card_activity_benefits (
    id UUID PRIMARY KEY,
    card_activity_id UUID NOT NULL REFERENCES card_activities(id) ON DELETE CASCADE,
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    name TEXT NOT NULL,
    rate NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    monthly_cap NUMERIC(18,6) CHECK (monthly_cap > 0),
    stack_group TEXT NOT NULL DEFAULT 'base' CHECK (btrim(stack_group) <> ''),
    priority INTEGER NOT NULL DEFAULT 100,
    required_account_tiers TEXT[] NOT NULL DEFAULT '{}',
    action_required TEXT NOT NULL DEFAULT 'none'
        CHECK (action_required IN ('none', 'registration', 'app_switch')),
    action_message TEXT NOT NULL DEFAULT '',
    payment_methods TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE card_activity_benefit_categories (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    category_code TEXT NOT NULL REFERENCES categories(code),
    PRIMARY KEY (benefit_id, category_code)
);

CREATE TABLE card_activity_benefit_merchants (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    merchant_keyword TEXT NOT NULL CHECK (btrim(merchant_keyword) <> ''),
    PRIMARY KEY (benefit_id, merchant_keyword)
);

INSERT INTO card_activity_benefits
    (id, card_activity_id, reward_unit_id, name, rate, monthly_cap, stack_group, priority, is_active, created_at, updated_at)
SELECT id, id, reward_unit_id, name, rate, monthly_cap, 'legacy', 100, is_active, created_at, updated_at
FROM card_activities;

INSERT INTO card_activity_benefit_categories(benefit_id, category_code)
SELECT card_activity_id, category_code FROM card_activity_categories;

INSERT INTO card_activity_benefit_merchants(benefit_id, merchant_keyword)
SELECT card_activity_id, merchant_keyword FROM card_activity_merchants;

ALTER TABLE reward_allocations DROP CONSTRAINT reward_allocations_activity_fkey;
ALTER TABLE reward_allocations RENAME COLUMN reward_rule_id TO benefit_id;
ALTER TABLE reward_allocations
    ADD CONSTRAINT reward_allocations_benefit_fkey
    FOREIGN KEY (benefit_id) REFERENCES card_activity_benefits(id);
ALTER TABLE reward_allocations
    DROP CONSTRAINT reward_allocations_transaction_id_reward_rule_id_key;
ALTER TABLE reward_allocations
    ADD UNIQUE (transaction_id, benefit_id);

DROP INDEX IF EXISTS reward_allocations_rule_transaction_idx;
CREATE INDEX reward_allocations_benefit_transaction_idx
    ON reward_allocations(benefit_id, transaction_id);
CREATE INDEX card_activity_benefits_activity_idx
    ON card_activity_benefits(card_activity_id, is_active);
CREATE INDEX card_activity_benefit_merchants_keyword_idx
    ON card_activity_benefit_merchants(lower(merchant_keyword));

DROP TABLE card_activity_merchants;
DROP TABLE card_activity_categories;

ALTER TABLE card_activities
    DROP COLUMN reward_unit_id,
    DROP COLUMN rate,
    DROP COLUMN monthly_cap;

ALTER TABLE member_cards
    ADD CONSTRAINT member_cards_account_tier_valid
    CHECK (account_tier IS NULL OR btrim(account_tier) <> '');

