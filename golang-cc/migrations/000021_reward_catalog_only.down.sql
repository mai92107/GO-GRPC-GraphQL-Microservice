CREATE TABLE card_activities (
    id UUID PRIMARY KEY,
    card_product_id UUID NOT NULL REFERENCES card_products(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_url TEXT,
    verified_at DATE,
    shared_monthly_caps JSONB NOT NULL DEFAULT '{}'::jsonb
);

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
    selectable_type TEXT,
    action_required TEXT NOT NULL DEFAULT 'none'
        CHECK (action_required IN ('none', 'registration', 'app_switch', 'account_setup')),
    action_message TEXT NOT NULL DEFAULT '',
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
    merchant_code TEXT NOT NULL REFERENCES merchants(code),
    PRIMARY KEY (benefit_id, merchant_code)
);

CREATE TABLE card_activity_benefit_payment_methods (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    payment_method_code TEXT NOT NULL REFERENCES payment_methods(code),
    PRIMARY KEY (benefit_id,payment_method_code)
);

CREATE TABLE card_activity_networks (
    card_activity_id UUID NOT NULL REFERENCES card_activities(id) ON DELETE CASCADE,
    card_network_id UUID NOT NULL REFERENCES catalog.card_networks(id) ON DELETE CASCADE,
    PRIMARY KEY(card_activity_id,card_network_id)
);

ALTER TABLE reward_allocations DROP CONSTRAINT IF EXISTS reward_allocations_component_fkey;
ALTER TABLE reward_allocations
    ADD CONSTRAINT reward_allocations_benefit_fkey
    FOREIGN KEY (benefit_id) REFERENCES card_activity_benefits(id);
