ALTER TABLE identity.users
ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE reward.activities
    ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS published_by UUID REFERENCES identity.users(id),
    ADD COLUMN IF NOT EXISTS published_checksum TEXT;

CREATE TABLE reward.published_activities (
    id UUID PRIMARY KEY,
    source_activity_id UUID NOT NULL REFERENCES reward.activities(id) ON DELETE CASCADE,
    bank_id UUID NOT NULL REFERENCES public.banks(id),
    card_product_id UUID NOT NULL REFERENCES catalog.card_products(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_by UUID REFERENCES identity.users(id),
    source_checksum TEXT NOT NULL,
    CHECK (effective_to >= effective_from)
);

CREATE TABLE reward.published_reward_rules (
    id UUID PRIMARY KEY,
    published_activity_id UUID NOT NULL REFERENCES reward.published_activities(id) ON DELETE CASCADE,
    source_activity_id UUID NOT NULL REFERENCES reward.activities(id) ON DELETE CASCADE,
    source_component_id UUID NOT NULL,
    source_benefit_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    layer INTEGER NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    stack_group TEXT NOT NULL,
    stack_policy TEXT NOT NULL CHECK (stack_policy IN ('stack','best_of_group','exclusive')),
    priority INTEGER NOT NULL DEFAULT 0,
    effect_type TEXT NOT NULL CHECK (effect_type IN ('ADD_RATE','SET_RATE','MULTIPLY_RATE','ADD_CASH','DISCOUNT')),
    reward_value NUMERIC(18,8) NOT NULL CHECK (reward_value > 0),
    reward_unit_id UUID NOT NULL REFERENCES public.reward_units(id),
    cap_amount NUMERIC(18,6),
    cap_formula TEXT,
    cap_period TEXT,
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (effective_to >= effective_from),
    CHECK (cap_amount IS NOT NULL OR NULLIF(btrim(COALESCE(cap_formula,'')), '') IS NOT NULL OR cap_period IS NULL)
);

CREATE TABLE reward.published_rule_requirements (
    id UUID PRIMARY KEY,
    published_rule_id UUID NOT NULL REFERENCES reward.published_reward_rules(id) ON DELETE CASCADE,
    source_requirement_id UUID NOT NULL,
    requirement_type TEXT NOT NULL,
    operator TEXT NOT NULL CHECK (operator IN ('IN','NOT_IN','EQ','GTE','LTE','BETWEEN')),
    configuration_json JSONB NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reward.published_rule_benefits (
    id UUID PRIMARY KEY,
    published_rule_id UUID NOT NULL REFERENCES reward.published_reward_rules(id) ON DELETE CASCADE,
    source_benefit_id UUID NOT NULL,
    benefit_type TEXT NOT NULL,
    value NUMERIC(18,8) NOT NULL CHECK (value > 0),
    reward_unit_id UUID NOT NULL REFERENCES public.reward_units(id),
    cap_amount NUMERIC(18,6),
    cap_formula TEXT,
    cap_period TEXT,
    description TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX published_activities_source_activity_idx
    ON reward.published_activities(source_activity_id);
CREATE INDEX published_activities_card_date_idx
    ON reward.published_activities(card_product_id, effective_from, effective_to);
CREATE INDEX published_reward_rules_activity_date_idx
    ON reward.published_reward_rules(published_activity_id, effective_from, effective_to, is_active);
CREATE INDEX published_rule_requirements_rule_type_idx
    ON reward.published_rule_requirements(published_rule_id, requirement_type);
CREATE INDEX published_rule_benefits_rule_idx
    ON reward.published_rule_benefits(published_rule_id);
