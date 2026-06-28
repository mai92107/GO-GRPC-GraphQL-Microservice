CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS merchant;
CREATE SCHEMA IF NOT EXISTS reward;
CREATE SCHEMA IF NOT EXISTS member_profile;
CREATE SCHEMA IF NOT EXISTS recommendation;
CREATE SCHEMA IF NOT EXISTS "transaction";
CREATE SCHEMA IF NOT EXISTS integration;

ALTER TABLE banks ADD COLUMN IF NOT EXISTS logo_url TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS code TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS card_image_url TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS primary_color TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS is_open_for_application BOOLEAN NOT NULL DEFAULT true;
UPDATE card_products SET code = trim(both '_' from lower(regexp_replace(name, '[^a-zA-Z0-9]+', '_', 'g'))) || '_' || replace(id::text,'-','')
WHERE code IS NULL;
ALTER TABLE card_products ALTER COLUMN code SET NOT NULL;
ALTER TABLE card_products ALTER COLUMN code SET DEFAULT ('card_' || replace(gen_random_uuid()::text,'-',''));
CREATE UNIQUE INDEX IF NOT EXISTS card_products_code_uidx ON card_products(code);

CREATE TABLE catalog.card_networks (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL CHECK (code IN ('visa','mastercard','jcb','amex')),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);
INSERT INTO catalog.card_networks(id,code,name) VALUES
    ('00000000-0000-0000-0000-000000000201','visa','Visa'),
    ('00000000-0000-0000-0000-000000000202','mastercard','Mastercard'),
    ('00000000-0000-0000-0000-000000000203','jcb','JCB'),
    ('00000000-0000-0000-0000-000000000204','amex','American Express')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE catalog.card_product_networks (
    card_product_id UUID NOT NULL REFERENCES card_products(id) ON DELETE CASCADE,
    card_network_id UUID NOT NULL REFERENCES catalog.card_networks(id),
    PRIMARY KEY(card_product_id,card_network_id)
);
INSERT INTO catalog.card_product_networks(card_product_id,card_network_id)
SELECT cp.id,n.id FROM card_products cp CROSS JOIN catalog.card_networks n
WHERE n.code IN ('visa','mastercard','jcb')
ON CONFLICT DO NOTHING;

ALTER TABLE member_cards ADD COLUMN IF NOT EXISTS card_network_id UUID REFERENCES catalog.card_networks(id);
ALTER TABLE member_cards ADD COLUMN IF NOT EXISTS opened_on DATE;
ALTER TABLE member_cards ADD COLUMN IF NOT EXISTS closed_on DATE;
ALTER TABLE member_cards DROP CONSTRAINT IF EXISTS member_cards_user_id_card_product_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS member_cards_user_product_network_uidx
    ON member_cards(user_id,card_product_id,card_network_id) NULLS NOT DISTINCT;

CREATE OR REPLACE FUNCTION catalog.validate_member_card_network() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.card_network_id IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM catalog.card_product_networks
        WHERE card_product_id=NEW.card_product_id AND card_network_id=NEW.card_network_id
    ) THEN
        RAISE EXCEPTION 'card network is not supported by card product';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER member_card_network_check
BEFORE INSERT OR UPDATE OF card_product_id,card_network_id ON member_cards
FOR EACH ROW EXECUTE FUNCTION catalog.validate_member_card_network();

CREATE TABLE catalog.card_plans (
    id UUID PRIMARY KEY,
    card_product_id UUID NOT NULL REFERENCES card_products(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    plan_type TEXT NOT NULL CHECK (plan_type IN ('selectable','qualified')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(card_product_id,code)
);
CREATE TABLE catalog.card_plan_versions (
    id UUID PRIMARY KEY,
    card_plan_id UUID NOT NULL REFERENCES catalog.card_plans(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    reminder_text TEXT NOT NULL DEFAULT '',
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    announced_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    supersedes_version_id UUID REFERENCES catalog.card_plan_versions(id),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (card_plan_id WITH =, tstzrange(effective_from,effective_to,'[)') WITH &&)
);
CREATE TABLE catalog.member_card_qualification_statuses (
    id UUID PRIMARY KEY,
    member_card_id UUID NOT NULL REFERENCES member_cards(id) ON DELETE CASCADE,
    card_plan_id UUID NOT NULL REFERENCES catalog.card_plans(id) ON DELETE CASCADE,
    is_qualified BOOLEAN NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    updated_by_user_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (member_card_id WITH =, card_plan_id WITH =, tstzrange(effective_from,effective_to,'[)') WITH &&)
);

INSERT INTO catalog.card_plans(id,card_product_id,code,plan_type,display_order)
SELECT md5('qualified-plan:'||cp.id::text||':'||tier)::uuid,cp.id,
       'qualified_'||lower(regexp_replace(tier,'[^a-zA-Z0-9]+','_','g'))||'_'||left(md5(tier),6),
       'qualified',ordinality::integer
FROM card_products cp
CROSS JOIN LATERAL unnest(cp.account_tiers) WITH ORDINALITY AS value(tier,ordinality)
ON CONFLICT DO NOTHING;
INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,effective_from,published_at)
SELECT md5('qualified-plan-version:'||p.id::text)::uuid,p.id,value.tier,'由會員自行確認是否符合銀行資格',
       '2000-01-01 00:00:00+00',now()
FROM catalog.card_plans p
JOIN card_products cp ON cp.id=p.card_product_id
CROSS JOIN LATERAL unnest(cp.account_tiers) value(tier)
WHERE p.id=md5('qualified-plan:'||cp.id::text||':'||value.tier)::uuid
ON CONFLICT DO NOTHING;
INSERT INTO catalog.member_card_qualification_statuses(
    id,member_card_id,card_plan_id,is_qualified,effective_from,updated_by_user_at,created_at)
SELECT md5('legacy-qualification:'||mc.id::text||':'||p.id::text)::uuid,mc.id,p.id,
       (pv.name=mc.account_tier),'2000-01-01 00:00:00+00',mc.updated_at,mc.created_at
FROM member_cards mc
JOIN catalog.card_plans p ON p.card_product_id=mc.card_product_id AND p.plan_type='qualified'
JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id
ON CONFLICT DO NOTHING;

ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS type TEXT;
ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;
UPDATE payment_methods SET type = CASE
    WHEN code IN ('physical_card') THEN 'physical_card'
    WHEN code IN ('online_card') THEN 'online_card'
    WHEN code IN ('easycard','icash','ipass') THEN 'electronic_ticket'
    ELSE 'mobile_payment' END
WHERE type IS NULL;
ALTER TABLE payment_methods ALTER COLUMN type SET NOT NULL;
ALTER TABLE payment_methods ALTER COLUMN type SET DEFAULT 'mobile_payment';
ALTER TABLE payment_methods ADD CONSTRAINT payment_methods_type_check
    CHECK (type IN ('physical_card','online_card','mobile_payment','electronic_ticket'));

CREATE TABLE member_profile.user_payment_methods (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_method_code TEXT NOT NULL REFERENCES payment_methods(code) ON DELETE CASCADE,
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id,payment_method_code)
);
INSERT INTO member_profile.user_payment_methods(id,user_id,payment_method_code,is_available)
SELECT md5('user-payment:'||u.id::text||':'||p.code)::uuid,u.id,p.code,true
FROM users u CROSS JOIN payment_methods p
WHERE u.role='member' AND p.type IN ('mobile_payment','electronic_ticket')
ON CONFLICT DO NOTHING;

CREATE TABLE reward.programs (
    id UUID PRIMARY KEY,
    card_product_id UUID NOT NULL REFERENCES card_products(id),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    source_url TEXT,
    verified_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('draft','published','archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(card_product_id,code)
);
CREATE TABLE reward.components (
    id UUID PRIMARY KEY,
    reward_program_id UUID NOT NULL REFERENCES reward.programs(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    stack_group TEXT NOT NULL,
    stack_policy TEXT NOT NULL DEFAULT 'best_of_group' CHECK (stack_policy IN ('stack','best_of_group','exclusive')),
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(reward_program_id,code)
);
CREATE TABLE reward.component_versions (
    id UUID PRIMARY KEY,
    reward_component_id UUID NOT NULL REFERENCES reward.components(id) ON DELETE CASCADE,
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    rate_kind TEXT NOT NULL DEFAULT 'percentage' CHECK (rate_kind IN ('percentage','fixed_amount','points_per_amount')),
    rate NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    announced_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    supersedes_version_id UUID REFERENCES reward.component_versions(id),
    change_reason TEXT NOT NULL DEFAULT '',
    display_change_until TIMESTAMPTZ,
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (reward_component_id WITH =, tstzrange(effective_from,effective_to,'[)') WITH &&)
);
CREATE TABLE reward.conditions (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    condition_type TEXT NOT NULL CHECK (condition_type IN ('category','merchant','payment_method','card_network','card_plan','country','currency','channel','amount_range')),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE reward.condition_versions (
    id UUID PRIMARY KEY,
    reward_condition_id UUID NOT NULL REFERENCES reward.conditions(id) ON DELETE CASCADE,
    operator TEXT NOT NULL CHECK (operator IN ('equals','in','not_in','gte','lte','between')),
    configuration_json JSONB NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    supersedes_version_id UUID REFERENCES reward.condition_versions(id),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (reward_condition_id WITH =, tstzrange(effective_from,effective_to,'[)') WITH &&)
);
CREATE TABLE reward.requirements (
    id UUID PRIMARY KEY,
    reward_component_id UUID NOT NULL REFERENCES reward.components(id) ON DELETE CASCADE,
    reward_condition_id UUID NOT NULL REFERENCES reward.conditions(id) ON DELETE CASCADE,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    display_order INTEGER NOT NULL DEFAULT 0,
    CHECK (effective_to IS NULL OR effective_to > effective_from)
);
CREATE TABLE reward.caps (
    id UUID PRIMARY KEY,
    reward_program_id UUID NOT NULL REFERENCES reward.programs(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN ('component','stack_group','program','card','customer')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(reward_program_id,code)
);
CREATE TABLE reward.cap_versions (
    id UUID PRIMARY KEY,
    reward_cap_id UUID NOT NULL REFERENCES reward.caps(id) ON DELETE CASCADE,
    cap_type TEXT NOT NULL CHECK (cap_type IN ('reward_amount','spending_amount','transaction_count')),
    limit_value NUMERIC(18,6) NOT NULL CHECK (limit_value > 0),
    reward_unit_id UUID REFERENCES reward_units(id),
    period_type TEXT NOT NULL CHECK (period_type IN ('calendar_month','statement_cycle','campaign','year')),
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    supersedes_version_id UUID REFERENCES reward.cap_versions(id),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (reward_cap_id WITH =, tstzrange(effective_from,effective_to,'[)') WITH &&)
);
CREATE TABLE reward.component_caps (
    reward_component_id UUID NOT NULL REFERENCES reward.components(id) ON DELETE CASCADE,
    reward_cap_id UUID NOT NULL REFERENCES reward.caps(id) ON DELETE CASCADE,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    PRIMARY KEY(reward_component_id,reward_cap_id,effective_from)
);

INSERT INTO reward.programs(id,card_product_id,code,name,source_url,verified_at,status,created_at,updated_at)
SELECT a.id,a.card_product_id,'legacy_'||replace(a.id::text,'-',''),a.name,a.source_url,a.verified_at::timestamptz,
       CASE WHEN a.is_active THEN 'published' ELSE 'archived' END,a.created_at,a.updated_at
FROM card_activities a ON CONFLICT DO NOTHING;
INSERT INTO reward.components(id,reward_program_id,code,stack_group,stack_policy,priority,is_active,created_at)
SELECT b.id,b.card_activity_id,'legacy_'||replace(b.id::text,'-',''),b.stack_group,'best_of_group',b.priority,b.is_active,b.created_at
FROM card_activity_benefits b ON CONFLICT DO NOTHING;
INSERT INTO reward.component_versions(id,reward_component_id,reward_unit_id,name,rate,effective_from,effective_to,published_at)
SELECT md5('component-version:'||b.id::text)::uuid,b.id,b.reward_unit_id,b.name,b.rate,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',
       (a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',COALESCE(a.created_at,now())
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
ON CONFLICT DO NOTHING;

INSERT INTO reward.conditions(id,code,condition_type,name)
SELECT md5('category-condition:'||b.id::text)::uuid,'category_'||b.id::text,'category',b.name||'／消費分類'
FROM card_activity_benefits b
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_categories x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;
INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('category-condition-version:'||b.id::text)::uuid,md5('category-condition:'||b.id::text)::uuid,'in',
       jsonb_build_object('category_codes',jsonb_agg(c.category_code ORDER BY c.category_code)),
       string_agg(c.category_code,', ' ORDER BY c.category_code),
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
JOIN card_activity_benefit_categories c ON c.benefit_id=b.id
GROUP BY b.id,a.start_date,a.end_date ON CONFLICT DO NOTHING;
INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('category-requirement:'||b.id::text)::uuid,b.id,md5('category-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',10
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_categories x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;

INSERT INTO reward.conditions(id,code,condition_type,name)
SELECT md5('qualified-condition:'||b.id::text)::uuid,'qualified_'||b.id::text,'card_plan',b.name||'／會員資格'
FROM card_activity_benefits b WHERE cardinality(b.required_account_tiers)>0
ON CONFLICT DO NOTHING;
INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('qualified-condition-version:'||b.id::text)::uuid,md5('qualified-condition:'||b.id::text)::uuid,'in',
       jsonb_build_object('card_plan_ids',jsonb_agg(p.id ORDER BY p.id)),
       '需符合：'||array_to_string(b.required_account_tiers,' / '),
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
JOIN catalog.card_plans p ON p.card_product_id=a.card_product_id
JOIN catalog.card_plan_versions pv ON pv.card_plan_id=p.id AND pv.name=ANY(b.required_account_tiers)
WHERE cardinality(b.required_account_tiers)>0
GROUP BY b.id,a.start_date,a.end_date,b.required_account_tiers
ON CONFLICT DO NOTHING;
INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('qualified-requirement:'||b.id::text)::uuid,b.id,md5('qualified-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',30
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE cardinality(b.required_account_tiers)>0
ON CONFLICT DO NOTHING;

INSERT INTO catalog.card_plans(id,card_product_id,code,plan_type,display_order)
SELECT md5('selectable-plan:'||b.id::text)::uuid,a.card_product_id,'selectable_'||replace(b.id::text,'-',''),'selectable',b.priority
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required='app_switch' ON CONFLICT DO NOTHING;
INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,reminder_text,effective_from,effective_to,published_at)
SELECT md5('selectable-plan-version:'||b.id::text)::uuid,md5('selectable-plan:'||b.id::text)::uuid,b.name,b.action_message,b.action_message,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',a.created_at
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required='app_switch' ON CONFLICT DO NOTHING;
INSERT INTO reward.conditions(id,code,condition_type,name)
SELECT md5('selectable-condition:'||b.id::text)::uuid,'selectable_'||b.id::text,'card_plan',b.name||'／切換方案'
FROM card_activity_benefits b WHERE b.action_required='app_switch'
ON CONFLICT DO NOTHING;
INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('selectable-condition-version:'||b.id::text)::uuid,md5('selectable-condition:'||b.id::text)::uuid,'equals',
       jsonb_build_object('card_plan_ids',jsonb_build_array(md5('selectable-plan:'||b.id::text)::uuid)),
       b.action_message,a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required='app_switch' ON CONFLICT DO NOTHING;
INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('selectable-requirement:'||b.id::text)::uuid,b.id,md5('selectable-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',40
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required='app_switch' ON CONFLICT DO NOTHING;

INSERT INTO reward.conditions(id,code,condition_type,name)
SELECT md5('payment-condition:'||b.id::text)::uuid,'payment_'||b.id::text,'payment_method',b.name||'／支付方式'
FROM card_activity_benefits b
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_payment_methods x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;
INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('payment-condition-version:'||b.id::text)::uuid,md5('payment-condition:'||b.id::text)::uuid,'in',
       jsonb_build_object('payment_method_codes',jsonb_agg(p.payment_method_code ORDER BY p.payment_method_code)),
       string_agg(pm.name,'、' ORDER BY pm.name),
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
JOIN card_activity_benefit_payment_methods p ON p.benefit_id=b.id
JOIN payment_methods pm ON pm.code=p.payment_method_code
GROUP BY b.id,a.start_date,a.end_date ON CONFLICT DO NOTHING;
INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('payment-requirement:'||b.id::text)::uuid,b.id,md5('payment-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',20
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_payment_methods x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;

INSERT INTO reward.conditions(id,code,condition_type,name)
SELECT md5('merchant-condition:'||b.id::text)::uuid,'merchant_'||b.id::text,'merchant',b.name||'／指定店家'
FROM card_activity_benefits b
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_merchants x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;
INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('merchant-condition-version:'||b.id::text)::uuid,md5('merchant-condition:'||b.id::text)::uuid,'in',
       jsonb_build_object('merchant_codes',jsonb_agg(m.merchant_code ORDER BY m.merchant_code)),
       string_agg(me.name,'、' ORDER BY me.name),
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
JOIN card_activity_benefit_merchants m ON m.benefit_id=b.id
JOIN merchants me ON me.code=m.merchant_code
GROUP BY b.id,a.start_date,a.end_date ON CONFLICT DO NOTHING;
INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('merchant-requirement:'||b.id::text)::uuid,b.id,md5('merchant-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',25
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE EXISTS (SELECT 1 FROM card_activity_benefit_merchants x WHERE x.benefit_id=b.id)
ON CONFLICT DO NOTHING;

INSERT INTO reward.caps(id,reward_program_id,code,name,scope)
SELECT md5('cap:'||b.id::text)::uuid,b.card_activity_id,'legacy_'||replace(b.id::text,'-',''),b.name||'每月上限','component'
FROM card_activity_benefits b WHERE b.monthly_cap IS NOT NULL ON CONFLICT DO NOTHING;
INSERT INTO reward.cap_versions(id,reward_cap_id,cap_type,limit_value,reward_unit_id,period_type,effective_from,effective_to)
SELECT md5('cap-version:'||b.id::text)::uuid,md5('cap:'||b.id::text)::uuid,'reward_amount',b.monthly_cap,b.reward_unit_id,'calendar_month',
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',(a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.monthly_cap IS NOT NULL ON CONFLICT DO NOTHING;
INSERT INTO reward.component_caps(reward_component_id,reward_cap_id,effective_from,effective_to)
SELECT b.id,md5('cap:'||b.id::text)::uuid,a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',
       (a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.monthly_cap IS NOT NULL ON CONFLICT DO NOTHING;

CREATE TABLE "transaction".reward_calculations (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    calculation_no INTEGER NOT NULL,
    calculation_type TEXT NOT NULL CHECK (calculation_type IN ('original','manual_recalculation','adjustment')),
    transaction_at TIMESTAMPTZ NOT NULL,
    total_effective_rate NUMERIC(18,8) NOT NULL DEFAULT 0,
    total_reward_value NUMERIC(18,6) NOT NULL DEFAULT 0,
    engine_version TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active','superseded')),
    supersedes_calculation_id UUID REFERENCES "transaction".reward_calculations(id),
    input_snapshot_json JSONB NOT NULL DEFAULT '{}',
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(transaction_id,calculation_no)
);
CREATE UNIQUE INDEX reward_calculation_active_uidx ON "transaction".reward_calculations(transaction_id) WHERE status='active';
CREATE TABLE "transaction".reward_calculation_components (
    id UUID PRIMARY KEY,
    reward_calculation_id UUID NOT NULL REFERENCES "transaction".reward_calculations(id) ON DELETE CASCADE,
    reward_component_id UUID REFERENCES reward.components(id),
    reward_component_version_id UUID REFERENCES reward.component_versions(id),
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    suggested_card_plan_id UUID REFERENCES catalog.card_plans(id),
    qualified_card_plan_id UUID REFERENCES catalog.card_plans(id),
    member_qualification_status_id UUID REFERENCES catalog.member_card_qualification_statuses(id),
    rate NUMERIC(18,8) NOT NULL,
    uncapped_reward NUMERIC(18,6) NOT NULL,
    allocated_reward NUMERIC(18,6) NOT NULL,
    cap_used_before NUMERIC(18,6) NOT NULL DEFAULT 0,
    cap_remaining_before NUMERIC(18,6),
    condition_snapshot_json JSONB NOT NULL DEFAULT '[]',
    reminder_snapshot_json JSONB NOT NULL DEFAULT '[]'
);

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS card_network_id UUID REFERENCES catalog.card_networks(id);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS currency_code CHAR(3) NOT NULL DEFAULT 'TWD';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS country_code CHAR(2) NOT NULL DEFAULT 'TW';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS channel TEXT NOT NULL DEFAULT 'physical' CHECK (channel IN ('online','physical'));
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('pending','confirmed','cancelled','refunded'));

INSERT INTO "transaction".reward_calculations(id,transaction_id,calculation_no,calculation_type,transaction_at,total_reward_value,engine_version,status,input_snapshot_json,calculated_at)
SELECT md5('legacy-calculation:'||t.id::text)::uuid,t.id,1,'original',t.transaction_date::timestamp AT TIME ZONE 'Asia/Taipei',
       COALESCE(sum(ra.score),0),'legacy-v1','active',jsonb_build_object('legacy',true),t.created_at
FROM transactions t LEFT JOIN reward_allocations ra ON ra.transaction_id=t.id
GROUP BY t.id ON CONFLICT DO NOTHING;
INSERT INTO "transaction".reward_calculation_components(
    id,reward_calculation_id,reward_component_id,reward_component_version_id,reward_unit_id,rate,
    uncapped_reward,allocated_reward,condition_snapshot_json,reminder_snapshot_json)
SELECT ra.id,md5('legacy-calculation:'||ra.transaction_id::text)::uuid,ra.benefit_id,
       md5('component-version:'||ra.benefit_id::text)::uuid,ra.reward_unit_id,b.rate,
       ra.uncapped_reward,ra.allocated_reward,jsonb_build_array(jsonb_build_object('legacy',true)),'[]'
FROM reward_allocations ra JOIN card_activity_benefits b ON b.id=ra.benefit_id
ON CONFLICT DO NOTHING;

CREATE TABLE integration.outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload_json JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
CREATE INDEX outbox_unpublished_idx ON integration.outbox_events(occurred_at) WHERE published_at IS NULL;

CREATE TABLE member_profile.reward_usage_adjustments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_card_id UUID NOT NULL REFERENCES member_cards(id) ON DELETE CASCADE,
    reward_cap_id UUID NOT NULL REFERENCES reward.caps(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    amount NUMERIC(18,6) NOT NULL,
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (period_end >= period_start)
);
