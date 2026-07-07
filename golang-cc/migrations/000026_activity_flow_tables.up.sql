CREATE SCHEMA IF NOT EXISTS reward;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS identity;

ALTER TABLE IF EXISTS public.users SET SCHEMA identity;
ALTER TABLE IF EXISTS public.sessions SET SCHEMA identity;
ALTER TABLE IF EXISTS public.invitations SET SCHEMA identity;
ALTER TABLE IF EXISTS public.password_reset_tokens SET SCHEMA identity;
ALTER TABLE IF EXISTS public.card_products SET SCHEMA catalog;

CREATE OR REPLACE VIEW public.users AS SELECT * FROM identity.users;
CREATE OR REPLACE VIEW public.sessions AS SELECT * FROM identity.sessions;
CREATE OR REPLACE VIEW public.invitations AS SELECT * FROM identity.invitations;
CREATE OR REPLACE VIEW public.password_reset_tokens AS SELECT * FROM identity.password_reset_tokens;
CREATE OR REPLACE VIEW public.card_products AS SELECT * FROM catalog.card_products;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'banks_pkey' AND conrelid = to_regclass('public.banks')
    ) THEN
        ALTER TABLE public.banks ADD CONSTRAINT banks_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'card_products_pkey' AND conrelid = to_regclass('catalog.card_products')
    ) THEN
        ALTER TABLE catalog.card_products ADD CONSTRAINT card_products_pkey PRIMARY KEY (id);
    END IF;
END $$;

CREATE TABLE reward.activities (
    id UUID PRIMARY KEY,
    bank_id UUID NOT NULL REFERENCES public.banks(id),
    card_product_id UUID NOT NULL REFERENCES catalog.card_products(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (effective_to >= effective_from)
);

CREATE TABLE reward.activity_groups (
    id UUID PRIMARY KEY,
    activity_id UUID NOT NULL REFERENCES reward.activities(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reward.activity_components (
    id UUID PRIMARY KEY,
    reward_group_id UUID NOT NULL REFERENCES reward.activity_groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    layer INTEGER NOT NULL,
    stack_group TEXT NOT NULL,
    stack_mode TEXT NOT NULL CHECK (stack_mode IN ('ADDITIVE','BEST_ONLY','EXCLUSIVE')),
    priority INTEGER NOT NULL DEFAULT 0,
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (layer > 0),
    CHECK (effective_to >= effective_from)
);

CREATE TABLE reward.activity_requirements (
    id UUID PRIMARY KEY,
    reward_component_id UUID NOT NULL REFERENCES reward.activity_components(id) ON DELETE CASCADE,
    requirement_type TEXT NOT NULL,
    operator TEXT NOT NULL CHECK (operator IN ('IN','NOT_IN','EQ','GTE','LTE','BETWEEN')),
    configuration_json JSONB NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'reward_units_pkey' AND conrelid = to_regclass('public.reward_units')
    ) THEN
        ALTER TABLE public.reward_units ADD CONSTRAINT reward_units_pkey PRIMARY KEY (id);
    END IF;
END $$;

CREATE TABLE reward.activity_benefits (
    id UUID PRIMARY KEY,
    reward_component_id UUID NOT NULL REFERENCES reward.activity_components(id) ON DELETE CASCADE,
    benefit_type TEXT NOT NULL,
    value NUMERIC(18,8) NOT NULL,
    reward_unit_id UUID NOT NULL REFERENCES public.reward_units(id),
    cap_amount NUMERIC(18,6),
    cap_period TEXT,
    description TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX reward_activities_card_active_idx
    ON reward.activities(card_product_id, effective_from, effective_to, is_active);
CREATE INDEX reward_activity_groups_activity_order_idx
    ON reward.activity_groups(activity_id, display_order);
CREATE INDEX reward_activity_components_group_order_idx
    ON reward.activity_components(reward_group_id, layer, priority);
CREATE INDEX reward_activity_requirements_component_type_idx
    ON reward.activity_requirements(reward_component_id, requirement_type);
CREATE INDEX reward_activity_benefits_component_type_idx
    ON reward.activity_benefits(reward_component_id, benefit_type);
