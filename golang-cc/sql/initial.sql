--
-- PostgreSQL database dump
--

\restrict vHn0oR01hWcC1gOcKIOUrxgrXwJ9sMCUcEUWQBykt2OeOML3gnBRF06t401iweP

-- Dumped from database version 17.10
-- Dumped by pg_dump version 17.10

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: catalog; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA catalog;


--
-- Name: identity; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA identity;


--
-- Name: integration; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA integration;


--
-- Name: member_profile; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA member_profile;


--
-- Name: merchant; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA merchant;


--
-- Name: recommendation; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA recommendation;


--
-- Name: reward; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA reward;


--
-- Name: transaction; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA transaction;


--
-- Name: btree_gist; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;


--
-- Name: EXTENSION btree_gist; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION btree_gist IS 'support for indexing common datatypes in GiST';


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: card_product_network_values(text); Type: FUNCTION; Schema: catalog; Owner: -
--

CREATE FUNCTION catalog.card_product_network_values(value text) RETURNS text[]
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT COALESCE(array_agg(DISTINCT trimmed ORDER BY trimmed), ARRAY[]::TEXT[])
    FROM (
        SELECT btrim(item) AS trimmed
        FROM regexp_split_to_table(COALESCE(value, ''), ',') item
        WHERE btrim(item) <> ''
    ) network_values
$$;


--
-- Name: validate_member_card_network(); Type: FUNCTION; Schema: catalog; Owner: -
--

CREATE FUNCTION catalog.validate_member_card_network() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    product_networks TEXT[];
BEGIN
    SELECT catalog.card_product_network_values(networks)
    INTO product_networks
    FROM catalog.card_products
    WHERE id = NEW.card_product_id;

    IF COALESCE(NEW.network, '') = '' THEN
        RAISE EXCEPTION 'card network is required';
    END IF;

    IF NOT NEW.network = ANY(product_networks) THEN
        RAISE EXCEPTION 'card network is not supported by card product';
    END IF;

    RETURN NEW;
END $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: card_networks; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.card_networks (
    id uuid NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: card_plan_versions; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.card_plan_versions (
    id uuid NOT NULL,
    card_plan_id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    reminder_text text DEFAULT ''::text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    announced_at timestamp with time zone,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    supersedes_version_id uuid,
    CONSTRAINT card_plan_versions_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from)))
);


--
-- Name: card_plans; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.card_plans (
    id uuid NOT NULL,
    card_product_id uuid NOT NULL,
    plan_type text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT card_plans_plan_type_check CHECK ((plan_type = ANY (ARRAY['selectable'::text, 'qualified'::text])))
);


--
-- Name: card_products; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.card_products (
    id uuid NOT NULL,
    bank_id uuid NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    account_tiers text[] DEFAULT '{}'::text[] NOT NULL,
    description text,
    card_image_url text,
    primary_color text,
    is_open_for_application boolean DEFAULT true NOT NULL,
    qualified_type text,
    selectable_type text,
    networks text DEFAULT ''::text NOT NULL
);


--
-- Name: member_card_qualification_statuses; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.member_card_qualification_statuses (
    id uuid NOT NULL,
    member_card_id uuid NOT NULL,
    card_plan_id uuid NOT NULL,
    is_qualified boolean NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    updated_by_user_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT member_card_qualification_statuses_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from)))
);


--
-- Name: regions; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.regions (
    id text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_qualifications; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.user_qualifications (
    id text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: invitations; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.invitations (
    id uuid NOT NULL,
    email public.citext NOT NULL,
    token_hash bytea NOT NULL,
    invited_by uuid NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: password_reset_tokens; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.password_reset_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: sessions; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.sessions (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.users (
    id uuid NOT NULL,
    email public.citext NOT NULL,
    password_hash text NOT NULL,
    display_name text NOT NULL,
    role text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY['admin'::text, 'member'::text]))),
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text])))
);


--
-- Name: outbox_events; Type: TABLE; Schema: integration; Owner: -
--

CREATE TABLE integration.outbox_events (
    id uuid NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    event_type text NOT NULL,
    payload_json jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone
);


--
-- Name: user_payment_methods; Type: TABLE; Schema: member_profile; Owner: -
--

CREATE TABLE member_profile.user_payment_methods (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    payment_method_id text NOT NULL,
    is_available boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: banks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.banks (
    id uuid NOT NULL,
    name text NOT NULL,
    code text,
    website_url text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    logo_url text
);


--
-- Name: card_products; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.card_products AS
 SELECT id,
    bank_id,
    name,
    is_active,
    created_at,
    updated_at,
    account_tiers,
    description,
    card_image_url,
    primary_color,
    is_open_for_application,
    qualified_type,
    selectable_type,
    networks
   FROM catalog.card_products;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: invitations; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.invitations AS
 SELECT id,
    email,
    token_hash,
    invited_by,
    expires_at,
    accepted_at,
    created_at
   FROM identity.invitations;


--
-- Name: member_card_credit_limits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.member_card_credit_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    member_card_id uuid NOT NULL,
    credit_limit numeric(18,2) NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT member_card_credit_limits_credit_limit_check CHECK ((credit_limit > (0)::numeric)),
    CONSTRAINT member_card_credit_limits_effective_range_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from)))
);


--
-- Name: member_cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.member_cards (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    card_product_id uuid NOT NULL,
    nickname text,
    last_four character(4),
    is_active boolean DEFAULT true NOT NULL,
    statement_day smallint,
    payment_due_day smallint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    account_tier text,
    opened_on date,
    closed_on date,
    credit_limit numeric(18,2),
    network text DEFAULT ''::text NOT NULL,
    CONSTRAINT member_cards_account_tier_valid CHECK (((account_tier IS NULL) OR (btrim(account_tier) <> ''::text))),
    CONSTRAINT member_cards_credit_limit_check CHECK (((credit_limit IS NULL) OR (credit_limit > (0)::numeric))),
    CONSTRAINT member_cards_last_four_check CHECK (((last_four IS NULL) OR (last_four ~ '^[0-9]{4}$'::text))),
    CONSTRAINT member_cards_payment_due_day_check CHECK (((payment_due_day >= 1) AND (payment_due_day <= 31))),
    CONSTRAINT member_cards_statement_day_check CHECK (((statement_day >= 1) AND (statement_day <= 31)))
);


--
-- Name: merchant_aliases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.merchant_aliases (
    alias text NOT NULL,
    merchant_id uuid NOT NULL,
    CONSTRAINT merchant_aliases_alias_check CHECK ((btrim(alias) <> ''::text))
);


--
-- Name: merchant_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.merchant_categories (
    category_id text NOT NULL,
    merchant_id uuid NOT NULL
);


--
-- Name: merchants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.merchants (
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    CONSTRAINT merchants_name_check CHECK ((btrim(name) <> ''::text))
);


--
-- Name: password_reset_tokens; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.password_reset_tokens AS
 SELECT id,
    user_id,
    token_hash,
    expires_at,
    used_at,
    created_at
   FROM identity.password_reset_tokens;


--
-- Name: payment_methods; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_methods (
    id text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    type text DEFAULT 'mobile_payment'::text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    CONSTRAINT payment_methods_code_check CHECK ((btrim(id) <> ''::text)),
    CONSTRAINT payment_methods_name_check CHECK ((btrim(name) <> ''::text)),
    CONSTRAINT payment_methods_type_check CHECK ((type = ANY (ARRAY['physical_card'::text, 'online_card'::text, 'mobile_payment'::text, 'electronic_ticket'::text])))
);


--
-- Name: reward_allocations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reward_allocations (
    id uuid NOT NULL,
    transaction_id uuid NOT NULL,
    benefit_id uuid NOT NULL,
    reward_unit_id uuid NOT NULL,
    uncapped_reward numeric(18,6) NOT NULL,
    allocated_reward numeric(18,6) NOT NULL,
    preference_weight numeric(18,6) NOT NULL,
    score numeric(18,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reward_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reward_preferences (
    user_id uuid NOT NULL,
    reward_unit_id uuid NOT NULL,
    weight numeric(18,6) NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reward_preferences_weight_check CHECK ((weight >= (0)::numeric))
);


--
-- Name: reward_units; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reward_units (
    id uuid NOT NULL,
    name text NOT NULL,
    symbol text NOT NULL,
    "precision" smallint DEFAULT 2 NOT NULL,
    is_system boolean DEFAULT true NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    symbol_position text DEFAULT 'prefix'::text NOT NULL,
    twd_rate numeric(18,6) DEFAULT 1 NOT NULL,
    CONSTRAINT reward_units_precision_check CHECK ((("precision" >= 0) AND ("precision" <= 6))),
    CONSTRAINT reward_units_symbol_position_check CHECK ((symbol_position = ANY (ARRAY['prefix'::text, 'suffix'::text]))),
    CONSTRAINT reward_units_twd_rate_check CHECK ((twd_rate > (0)::numeric))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: sessions; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.sessions AS
 SELECT id,
    user_id,
    token_hash,
    expires_at,
    created_at,
    last_seen_at
   FROM identity.sessions;


--
-- Name: telegram_chat_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.telegram_chat_bindings (
    chat_id bigint NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    card_id uuid NOT NULL,
    amount_minor bigint NOT NULL,
    category_id text NOT NULL,
    transaction_date date NOT NULL,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    merchant_name text DEFAULT ''::text NOT NULL,
    payment_method_id text NOT NULL,
    currency_id character(3) DEFAULT 'TWD'::bpchar NOT NULL,
    country_id character(2) DEFAULT 'TW'::bpchar NOT NULL,
    channel text DEFAULT 'physical'::text NOT NULL,
    status text DEFAULT 'confirmed'::text NOT NULL,
    merchant_id uuid,
    network text DEFAULT ''::text NOT NULL,
    CONSTRAINT transactions_amount_minor_check CHECK ((amount_minor > 0)),
    CONSTRAINT transactions_category_code_check CHECK ((category_id <> 'general'::text)),
    CONSTRAINT transactions_channel_check CHECK ((channel = ANY (ARRAY['online'::text, 'physical'::text]))),
    CONSTRAINT transactions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'confirmed'::text, 'cancelled'::text, 'refunded'::text])))
);


--
-- Name: users; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.users AS
 SELECT id,
    email,
    password_hash,
    display_name,
    role,
    status,
    created_at,
    updated_at
   FROM identity.users;


--
-- Name: activities; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activities (
    id uuid NOT NULL,
    bank_id uuid NOT NULL,
    card_product_id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    source_url text DEFAULT ''::text NOT NULL,
    effective_from date NOT NULL,
    effective_to date NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone,
    published_by uuid,
    published_checksum text,
    CONSTRAINT activities_check CHECK ((effective_to >= effective_from))
);


--
-- Name: activity_benefits; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activity_benefits (
    id uuid NOT NULL,
    reward_component_id uuid NOT NULL,
    benefit_type text NOT NULL,
    value numeric(18,8) NOT NULL,
    reward_unit_id uuid NOT NULL,
    cap_amount numeric(18,6),
    cap_period text,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    cap_formula text
);


--
-- Name: activity_component_groups; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activity_component_groups (
    reward_component_id uuid NOT NULL,
    reward_group_id uuid NOT NULL
);


--
-- Name: activity_components; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activity_components (
    id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    layer integer NOT NULL,
    stack_group text NOT NULL,
    stack_mode text NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    effective_from date NOT NULL,
    effective_to date NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT activity_components_check CHECK ((effective_to >= effective_from)),
    CONSTRAINT activity_components_layer_check CHECK ((layer > 0)),
    CONSTRAINT activity_components_stack_mode_check CHECK ((stack_mode = ANY (ARRAY['ADDITIVE'::text, 'BEST_ONLY'::text, 'EXCLUSIVE'::text])))
);


--
-- Name: activity_groups; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activity_groups (
    id uuid NOT NULL,
    activity_id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activity_requirements; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.activity_requirements (
    id uuid NOT NULL,
    reward_component_id uuid NOT NULL,
    requirement_type text NOT NULL,
    operator text NOT NULL,
    configuration_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT activity_requirements_operator_check CHECK ((operator = ANY (ARRAY['IN'::text, 'NOT_IN'::text, 'EQ'::text, 'GTE'::text, 'LTE'::text, 'BETWEEN'::text])))
);


--
-- Name: migration_reports; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.migration_reports (
    id uuid NOT NULL,
    migration_version text NOT NULL,
    severity text NOT NULL,
    subject_type text NOT NULL,
    subject_id uuid,
    message text NOT NULL,
    details_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT migration_reports_severity_check CHECK ((severity = ANY (ARRAY['info'::text, 'warning'::text, 'error'::text])))
);


--
-- Name: published_activities; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.published_activities (
    id uuid NOT NULL,
    source_activity_id uuid NOT NULL,
    bank_id uuid NOT NULL,
    card_product_id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    source_url text DEFAULT ''::text NOT NULL,
    effective_from date NOT NULL,
    effective_to date NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    published_by uuid,
    source_checksum text NOT NULL,
    CONSTRAINT published_activities_check CHECK ((effective_to >= effective_from))
);


--
-- Name: published_reward_rules; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.published_reward_rules (
    id uuid NOT NULL,
    published_activity_id uuid NOT NULL,
    source_activity_id uuid NOT NULL,
    source_component_id uuid NOT NULL,
    source_benefit_id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    layer integer NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    stack_group text NOT NULL,
    stack_policy text NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    effect_type text NOT NULL,
    reward_value numeric(18,8) NOT NULL,
    reward_unit_id uuid NOT NULL,
    cap_amount numeric(18,6),
    cap_formula text,
    cap_period text,
    effective_from date NOT NULL,
    effective_to date NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT published_reward_rules_check CHECK ((effective_to >= effective_from)),
    CONSTRAINT published_reward_rules_check1 CHECK (((cap_amount IS NOT NULL) OR (NULLIF(btrim(COALESCE(cap_formula, ''::text)), ''::text) IS NOT NULL) OR (cap_period IS NULL))),
    CONSTRAINT published_reward_rules_effect_type_check CHECK ((effect_type = ANY (ARRAY['ADD_RATE'::text, 'SET_RATE'::text, 'MULTIPLY_RATE'::text, 'ADD_CASH'::text, 'DISCOUNT'::text]))),
    CONSTRAINT published_reward_rules_reward_value_check CHECK ((reward_value > (0)::numeric)),
    CONSTRAINT published_reward_rules_stack_policy_check CHECK ((stack_policy = ANY (ARRAY['stack'::text, 'best_of_group'::text, 'exclusive'::text])))
);


--
-- Name: published_rule_benefits; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.published_rule_benefits (
    id uuid NOT NULL,
    published_rule_id uuid NOT NULL,
    source_benefit_id uuid NOT NULL,
    benefit_type text NOT NULL,
    value numeric(18,8) NOT NULL,
    reward_unit_id uuid NOT NULL,
    cap_amount numeric(18,6),
    cap_formula text,
    cap_period text,
    description text DEFAULT ''::text NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT published_rule_benefits_value_check CHECK ((value > (0)::numeric))
);


--
-- Name: published_rule_requirements; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.published_rule_requirements (
    id uuid NOT NULL,
    published_rule_id uuid NOT NULL,
    source_requirement_id uuid NOT NULL,
    requirement_type text NOT NULL,
    operator text NOT NULL,
    configuration_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT published_rule_requirements_operator_check CHECK ((operator = ANY (ARRAY['IN'::text, 'NOT_IN'::text, 'EQ'::text, 'GTE'::text, 'LTE'::text, 'BETWEEN'::text])))
);


--
-- Name: requirement_types; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.requirement_types (
    id text NOT NULL,
    name text NOT NULL,
    value_key text NOT NULL,
    value_source text NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reward_calculation_components; Type: TABLE; Schema: transaction; Owner: -
--

CREATE TABLE transaction.reward_calculation_components (
    id uuid NOT NULL,
    reward_calculation_id uuid NOT NULL,
    reward_component_id uuid,
    reward_unit_id uuid NOT NULL,
    suggested_card_plan_id uuid,
    qualified_card_plan_id uuid,
    member_qualification_status_id uuid,
    uncapped_reward numeric(18,6) NOT NULL,
    allocated_reward numeric(18,6) NOT NULL,
    cap_used_before numeric(18,6) DEFAULT 0 NOT NULL,
    cap_remaining_before numeric(18,6),
    condition_snapshot_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    reminder_snapshot_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    effect_type text DEFAULT 'ADD_RATE'::text NOT NULL,
    reward_value numeric(18,8) DEFAULT 0 NOT NULL,
    reward_rate numeric(18,8) DEFAULT 0 NOT NULL
);


--
-- Name: reward_calculations; Type: TABLE; Schema: transaction; Owner: -
--

CREATE TABLE transaction.reward_calculations (
    id uuid NOT NULL,
    transaction_id uuid NOT NULL,
    calculation_no integer NOT NULL,
    calculation_type text NOT NULL,
    transaction_at timestamp with time zone NOT NULL,
    total_effective_rate numeric(18,8) DEFAULT 0 NOT NULL,
    total_reward_value numeric(18,6) DEFAULT 0 NOT NULL,
    engine_version text NOT NULL,
    status text NOT NULL,
    supersedes_calculation_id uuid,
    input_snapshot_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    calculated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reward_calculations_calculation_type_check CHECK ((calculation_type = ANY (ARRAY['original'::text, 'manual_recalculation'::text, 'adjustment'::text]))),
    CONSTRAINT reward_calculations_status_check CHECK ((status = ANY (ARRAY['active'::text, 'superseded'::text])))
);


--
-- Data for Name: card_networks; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_networks (id, name, is_active) FROM stdin;
00000000-0000-0000-0000-000000000201	Visa	t
00000000-0000-0000-0000-000000000202	Mastercard	t
00000000-0000-0000-0000-000000000203	JCB	t
00000000-0000-0000-0000-000000000204	American Express	t
\.


--
-- Data for Name: card_plan_versions; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_plan_versions (id, card_plan_id, name, description, reminder_text, effective_from, effective_to, announced_at, published_at, supersedes_version_id) FROM stdin;
45342351-0a0a-a4e9-a341-fa805fe21767	654c3131-2c29-2a4e-90cd-4cd7e62e5079	大大	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-07-09 12:00:43.52494+00	\N
ce87c2ed-e6b2-f0f4-5bc0-016cc0dc1cb7	21c503d8-00a6-6402-3c54-25e9001a3b95	大戶	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-07-09 12:00:43.52494+00	\N
bef0e8c8-8838-4632-8fd8-a9f9e45bb427	c94d64f7-fcd8-8e6d-3720-21a4216fc823	大戶Plus	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-07-09 12:00:43.52494+00	\N
225c5b87-2617-b160-f6ca-cd8c09870258	72b99c99-4ffc-15a8-6481-e32444466701	Richart 指定方案加碼	需於 Richart Life APP 切換至符合消費情境的方案	需於 Richart Life APP 切換至符合消費情境的方案	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:40.441982+00	\N
8d9f58f0-1dc7-c863-7500-3bfce0bfcb01	ca407472-df08-104b-008a-09212863d809	UP 選指定行動支付最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:41.290412+00	\N
295dba5a-7bf2-c264-98ef-0a3edc08d543	a40bb29c-13b8-33be-e7fb-30ef9262d339	UP 選百大指定消費最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:41.290412+00	\N
0986110b-e5d8-91b9-c5af-44f0b804e398	ab2f95ab-2d7f-296a-5580-598ba377c568	UP 選國外實體消費最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:41.290412+00	\N
3d5896a5-1adf-ac5f-115f-108af24fa833	12831358-db59-f1bf-0606-9a32cea894b1	大戶	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:45.013338+00	\N
9a43c148-47e0-cc3b-5581-de935b233628	edf55cc5-435f-734f-3420-2e4e99f0a67f	大戶	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-07-09 12:00:45.013338+00	\N
\.


--
-- Data for Name: card_plans; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_plans (id, card_product_id, plan_type, is_active, display_order, created_at) FROM stdin;
654c3131-2c29-2a4e-90cd-4cd7e62e5079	51000000-0000-0000-0000-000000000001	qualified	t	1	2026-07-09 12:00:43.522792+00
21c503d8-00a6-6402-3c54-25e9001a3b95	51000000-0000-0000-0000-000000000001	qualified	t	2	2026-07-09 12:00:43.522792+00
c94d64f7-fcd8-8e6d-3720-21a4216fc823	51000000-0000-0000-0000-000000000001	qualified	t	3	2026-07-09 12:00:43.522792+00
72b99c99-4ffc-15a8-6481-e32444466701	51000000-0000-0000-0000-000000000003	selectable	t	20	2026-07-09 12:00:43.581219+00
ca407472-df08-104b-008a-09212863d809	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-07-09 12:00:43.581219+00
a40bb29c-13b8-33be-e7fb-30ef9262d339	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-07-09 12:00:43.581219+00
ab2f95ab-2d7f-296a-5580-598ba377c568	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-07-09 12:00:43.581219+00
12831358-db59-f1bf-0606-9a32cea894b1	51000000-0000-0000-0000-000000000001	selectable	t	20	2026-07-09 12:00:45.012536+00
edf55cc5-435f-734f-3420-2e4e99f0a67f	51000000-0000-0000-0000-000000000001	selectable	t	30	2026-07-09 12:00:45.012536+00
\.


--
-- Data for Name: card_products; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_products (id, bank_id, name, is_active, created_at, updated_at, account_tiers, description, card_image_url, primary_color, is_open_for_application, qualified_type, selectable_type, networks) FROM stdin;
51000000-0000-0000-0000-000000000001	41000000-0000-0000-0000-000000000001	DAWHO 現金回饋信用卡	t	2026-07-09 12:00:40.439631+00	2026-07-09 12:00:40.439631+00	{}	\N	\N	\N	t	大戶Plus	大大,大戶	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000002	41000000-0000-0000-0000-000000000001	SPORT 卡	t	2026-07-09 12:00:40.439631+00	2026-07-09 12:00:40.439631+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000003	41000000-0000-0000-0000-000000000002	Richart 卡	t	2026-07-09 12:00:40.439631+00	2026-07-09 12:00:40.439631+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000004	41000000-0000-0000-0000-000000000003	uniopen 聯名卡	t	2026-07-09 12:00:40.439631+00	2026-07-09 12:00:40.439631+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000005	41000000-0000-0000-0000-000000000004	U Bear 信用卡	t	2026-07-09 12:00:41.289211+00	2026-07-09 12:00:41.289211+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000006	41000000-0000-0000-0000-000000000004	Unicard	t	2026-07-09 12:00:41.289211+00	2026-07-09 12:00:41.289211+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000007	41000000-0000-0000-0000-000000000003	英雄聯盟信用卡（已停止申辦）	t	2026-07-09 12:00:41.289211+00	2026-07-09 12:00:41.289211+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000008	41000000-0000-0000-0000-000000000004	Pi 拍錢包信用卡	t	2026-07-09 12:00:43.218495+00	2026-07-09 12:00:43.218495+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000009	41000000-0000-0000-0000-000000000005	小小兵回饋卡	t	2026-07-09 12:00:43.218495+00	2026-07-09 12:00:43.218495+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
51000000-0000-0000-0000-000000000010	41000000-0000-0000-0000-000000000006	LINE Bank聯名卡	t	2026-07-09 12:00:43.218495+00	2026-07-09 12:00:43.218495+00	{}	\N	\N	\N	t	\N	\N	JCB,Mastercard,Visa
\.


--
-- Data for Name: member_card_qualification_statuses; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.member_card_qualification_statuses (id, member_card_id, card_plan_id, is_qualified, effective_from, effective_to, updated_by_user_at, created_at) FROM stdin;
\.


--
-- Data for Name: regions; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.regions (id, name, is_active, created_at, updated_at) FROM stdin;
TW	台灣	t	2026-07-09 12:00:46.802366+00	2026-07-09 12:00:46.802366+00
OVERSEAS	海外	t	2026-07-09 12:00:46.802366+00	2026-07-09 12:00:46.802366+00
\.


--
-- Data for Name: user_qualifications; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.user_qualifications (id, name, is_active, created_at, updated_at) FROM stdin;
NEW_USER	新戶	t	2026-07-09 12:00:46.803092+00	2026-07-09 12:00:46.803092+00
PAYROLL	薪轉戶	t	2026-07-09 12:00:46.803092+00	2026-07-09 12:00:46.803092+00
DIGITAL_ACCOUNT	數位帳戶戶	t	2026-07-09 12:00:46.803092+00	2026-07-09 12:00:46.803092+00
REGISTERED	已登錄	t	2026-07-09 12:00:46.803092+00	2026-07-09 12:00:46.803092+00
\.


--
-- Data for Name: invitations; Type: TABLE DATA; Schema: identity; Owner: -
--

COPY identity.invitations (id, email, token_hash, invited_by, expires_at, accepted_at, created_at) FROM stdin;
\.


--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: identity; Owner: -
--

COPY identity.password_reset_tokens (id, user_id, token_hash, expires_at, used_at, created_at) FROM stdin;
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: identity; Owner: -
--

COPY identity.sessions (id, user_id, token_hash, expires_at, created_at, last_seen_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: identity; Owner: -
--

COPY identity.users (id, email, password_hash, display_name, role, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: outbox_events; Type: TABLE DATA; Schema: integration; Owner: -
--

COPY integration.outbox_events (id, aggregate_type, aggregate_id, event_type, payload_json, occurred_at, published_at) FROM stdin;
\.


--
-- Data for Name: user_payment_methods; Type: TABLE DATA; Schema: member_profile; Owner: -
--

COPY member_profile.user_payment_methods (id, user_id, payment_method_id, is_available, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: banks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.banks (id, name, code, website_url, is_active, created_at, updated_at, logo_url) FROM stdin;
41000000-0000-0000-0000-000000000001	永豐銀行	sinopac	https://bank.sinopac.com/	t	2026-07-09 12:00:40.438011+00	2026-07-09 12:00:40.438011+00	\N
41000000-0000-0000-0000-000000000002	台新銀行	taishin	https://www.taishinbank.com.tw/	t	2026-07-09 12:00:40.438011+00	2026-07-09 12:00:40.438011+00	\N
41000000-0000-0000-0000-000000000003	中國信託	ctbc	https://www.ctbcbank.com/	t	2026-07-09 12:00:40.438011+00	2026-07-09 12:00:40.438011+00	\N
41000000-0000-0000-0000-000000000004	玉山銀行	esun	https://www.esunbank.com/	t	2026-07-09 12:00:41.288502+00	2026-07-09 12:00:41.288502+00	\N
41000000-0000-0000-0000-000000000005	上海商銀	scsb	https://www.scsb.com.tw/	t	2026-07-09 12:00:43.217502+00	2026-07-09 12:00:43.217502+00	\N
41000000-0000-0000-0000-000000000006	聯邦銀行	ubot	https://www.ubot.com.tw/	t	2026-07-09 12:00:43.217502+00	2026-07-09 12:00:43.217502+00	\N
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.categories (id, name, is_active) FROM stdin;
general	一般消費	t
dining	餐飲	t
transport	交通	t
overseas	海外消費	t
sports	運動健身	t
travel	旅遊	t
entertainment	娛樂	t
grocery	量販超市	t
other	其他	t
online	網路服務	t
insurance	保險	t
\.


--
-- Data for Name: member_card_credit_limits; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.member_card_credit_limits (id, member_card_id, credit_limit, effective_from, effective_to, created_at) FROM stdin;
\.


--
-- Data for Name: member_cards; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.member_cards (id, user_id, card_product_id, nickname, last_four, is_active, statement_day, payment_due_day, created_at, updated_at, account_tier, opened_on, closed_on, credit_limit, network) FROM stdin;
\.


--
-- Data for Name: merchant_aliases; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchant_aliases (alias, merchant_id) FROM stdin;
蝦皮	f50a68fd-bca5-40ee-ad87-eebf9783998e
Gemini	7d23aba2-212a-4a61-8312-86a201658c76
UNIQLO	febc2cf3-27ce-4a0d-b984-99162b9f179a
7-ELEVEN	eb8bc395-7ebf-4969-ad12-349925861b6b
星巴克	290841a5-6c30-46a4-aa77-7b5d84c96601
ChatGPT	5017e960-f4ce-4269-9b74-259972b62298
家樂福	39d5b603-7335-4aaf-aa1a-d6010065e596
Netflix	9c3fe21b-52a8-4078-90d8-43076113170d
台灣中油	239373b1-9958-4c6d-8d0d-5f3323d23cdd
Nintendo	c8766601-c62f-4164-9aeb-896c9b9be0c6
Uber	da0463ff-afdd-4090-8daa-7579829e8383
Steam	98aa0653-e613-48e6-ac0d-9a3bbd208020
Agoda	cf82193a-594b-42bc-9843-428a7837c106
新光三越	b64be4d8-6c29-49ce-b25a-cb5962a0fead
Booking.com	f81d8917-50ef-42ad-b5ec-feaaa65dad7e
高鐵	36e95dd5-4f9f-48e6-9816-c926c12a26b0
PlayStation	686081c8-1d0d-4fd0-a9b8-295b678eb527
momo	48221a92-b3ca-4268-9f8c-89e0fb4b66f7
PChome	0a8c93a3-c852-4d8b-86d7-d1914aad4c99
淘寶	5c04d790-b01f-4e63-95bf-eb746fb49b2d
酷澎	da6d85ee-63fb-4c85-953b-ffcff061ff7c
Disney+	b35de902-0cc9-4fe4-b49d-ed1aed6b1244
CatchPlay	031abaad-9d07-45d9-90ef-a3528d9bc57b
Spotify	2b5bd0da-a1af-48af-929d-012daa69be61
LINE TV	f732b0f2-3144-426d-9fd3-ab0a51d8eabb
KKTV	ad55388a-9309-415d-9836-9eaf0b09652b
KKBOX	772ec7a4-9af3-4fb4-8d4f-3860f069c220
Blizzard	a9588653-dacb-4538-a367-e3ead23bda4f
Garena	425c40c4-cb49-4ce9-810f-2d20b82c1097
環球影城	6371ceee-559f-4483-9d4f-60ad4cfa3895
楓康超市	4ae10c24-bb90-493f-99ff-935baa89377a
喜互惠	e0f39ba6-4b2d-4a0c-97af-48fa017d1c4d
美廉社	432071a1-9768-48d8-be80-f1c29d5428dc
愛買	f75ba67b-3cbd-47f7-a201-38dd2af54602
大全聯	d0f602f5-fb6f-4ed3-b103-32475d4bc849
全家	b962c512-402e-4782-b07b-839ad7dfb15e
\.


--
-- Data for Name: merchant_categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchant_categories (category_id, merchant_id) FROM stdin;
online	f50a68fd-bca5-40ee-ad87-eebf9783998e
online	7d23aba2-212a-4a61-8312-86a201658c76
other	febc2cf3-27ce-4a0d-b984-99162b9f179a
grocery	eb8bc395-7ebf-4969-ad12-349925861b6b
dining	290841a5-6c30-46a4-aa77-7b5d84c96601
online	5017e960-f4ce-4269-9b74-259972b62298
grocery	39d5b603-7335-4aaf-aa1a-d6010065e596
entertainment	9c3fe21b-52a8-4078-90d8-43076113170d
online	9c3fe21b-52a8-4078-90d8-43076113170d
transport	239373b1-9958-4c6d-8d0d-5f3323d23cdd
entertainment	c8766601-c62f-4164-9aeb-896c9b9be0c6
online	c8766601-c62f-4164-9aeb-896c9b9be0c6
transport	da0463ff-afdd-4090-8daa-7579829e8383
entertainment	98aa0653-e613-48e6-ac0d-9a3bbd208020
online	98aa0653-e613-48e6-ac0d-9a3bbd208020
travel	cf82193a-594b-42bc-9843-428a7837c106
online	cf82193a-594b-42bc-9843-428a7837c106
other	b64be4d8-6c29-49ce-b25a-cb5962a0fead
travel	f81d8917-50ef-42ad-b5ec-feaaa65dad7e
online	f81d8917-50ef-42ad-b5ec-feaaa65dad7e
travel	36e95dd5-4f9f-48e6-9816-c926c12a26b0
transport	36e95dd5-4f9f-48e6-9816-c926c12a26b0
entertainment	686081c8-1d0d-4fd0-a9b8-295b678eb527
online	686081c8-1d0d-4fd0-a9b8-295b678eb527
online	48221a92-b3ca-4268-9f8c-89e0fb4b66f7
online	0a8c93a3-c852-4d8b-86d7-d1914aad4c99
online	5c04d790-b01f-4e63-95bf-eb746fb49b2d
online	da6d85ee-63fb-4c85-953b-ffcff061ff7c
online	b35de902-0cc9-4fe4-b49d-ed1aed6b1244
entertainment	b35de902-0cc9-4fe4-b49d-ed1aed6b1244
online	031abaad-9d07-45d9-90ef-a3528d9bc57b
entertainment	031abaad-9d07-45d9-90ef-a3528d9bc57b
online	2b5bd0da-a1af-48af-929d-012daa69be61
entertainment	2b5bd0da-a1af-48af-929d-012daa69be61
online	f732b0f2-3144-426d-9fd3-ab0a51d8eabb
entertainment	f732b0f2-3144-426d-9fd3-ab0a51d8eabb
online	ad55388a-9309-415d-9836-9eaf0b09652b
entertainment	ad55388a-9309-415d-9836-9eaf0b09652b
online	772ec7a4-9af3-4fb4-8d4f-3860f069c220
entertainment	772ec7a4-9af3-4fb4-8d4f-3860f069c220
online	a9588653-dacb-4538-a367-e3ead23bda4f
entertainment	a9588653-dacb-4538-a367-e3ead23bda4f
online	425c40c4-cb49-4ce9-810f-2d20b82c1097
entertainment	425c40c4-cb49-4ce9-810f-2d20b82c1097
travel	6371ceee-559f-4483-9d4f-60ad4cfa3895
entertainment	6371ceee-559f-4483-9d4f-60ad4cfa3895
grocery	4ae10c24-bb90-493f-99ff-935baa89377a
grocery	e0f39ba6-4b2d-4a0c-97af-48fa017d1c4d
grocery	432071a1-9768-48d8-be80-f1c29d5428dc
grocery	f75ba67b-3cbd-47f7-a201-38dd2af54602
grocery	d0f602f5-fb6f-4ed3-b103-32475d4bc849
grocery	b962c512-402e-4782-b07b-839ad7dfb15e
\.


--
-- Data for Name: merchants; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchants (name, is_active, is_system, created_at, updated_at, id) FROM stdin;
蝦皮	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	f50a68fd-bca5-40ee-ad87-eebf9783998e
Gemini	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	7d23aba2-212a-4a61-8312-86a201658c76
UNIQLO	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	febc2cf3-27ce-4a0d-b984-99162b9f179a
7-ELEVEN	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	eb8bc395-7ebf-4969-ad12-349925861b6b
星巴克	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	290841a5-6c30-46a4-aa77-7b5d84c96601
ChatGPT	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	5017e960-f4ce-4269-9b74-259972b62298
家樂福	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	39d5b603-7335-4aaf-aa1a-d6010065e596
Netflix	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	9c3fe21b-52a8-4078-90d8-43076113170d
台灣中油	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	239373b1-9958-4c6d-8d0d-5f3323d23cdd
Nintendo	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	c8766601-c62f-4164-9aeb-896c9b9be0c6
Uber	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	da0463ff-afdd-4090-8daa-7579829e8383
Steam	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	98aa0653-e613-48e6-ac0d-9a3bbd208020
Agoda	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	cf82193a-594b-42bc-9843-428a7837c106
新光三越	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	b64be4d8-6c29-49ce-b25a-cb5962a0fead
Booking.com	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	f81d8917-50ef-42ad-b5ec-feaaa65dad7e
高鐵	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	36e95dd5-4f9f-48e6-9816-c926c12a26b0
PlayStation	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	686081c8-1d0d-4fd0-a9b8-295b678eb527
momo	t	t	2026-07-09 12:00:41.852152+00	2026-07-09 12:00:41.852152+00	48221a92-b3ca-4268-9f8c-89e0fb4b66f7
PChome	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	0a8c93a3-c852-4d8b-86d7-d1914aad4c99
淘寶	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	5c04d790-b01f-4e63-95bf-eb746fb49b2d
酷澎	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	da6d85ee-63fb-4c85-953b-ffcff061ff7c
Disney+	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	b35de902-0cc9-4fe4-b49d-ed1aed6b1244
CatchPlay	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	031abaad-9d07-45d9-90ef-a3528d9bc57b
Spotify	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	2b5bd0da-a1af-48af-929d-012daa69be61
LINE TV	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	f732b0f2-3144-426d-9fd3-ab0a51d8eabb
KKTV	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	ad55388a-9309-415d-9836-9eaf0b09652b
KKBOX	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	772ec7a4-9af3-4fb4-8d4f-3860f069c220
Blizzard	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	a9588653-dacb-4538-a367-e3ead23bda4f
Garena	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	425c40c4-cb49-4ce9-810f-2d20b82c1097
環球影城	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	6371ceee-559f-4483-9d4f-60ad4cfa3895
楓康超市	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	4ae10c24-bb90-493f-99ff-935baa89377a
喜互惠	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	e0f39ba6-4b2d-4a0c-97af-48fa017d1c4d
美廉社	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	432071a1-9768-48d8-be80-f1c29d5428dc
愛買	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	f75ba67b-3cbd-47f7-a201-38dd2af54602
大全聯	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	d0f602f5-fb6f-4ed3-b103-32475d4bc849
全家	t	t	2026-07-09 12:00:43.22023+00	2026-07-09 12:00:43.22023+00	b962c512-402e-4782-b07b-839ad7dfb15e
\.


--
-- Data for Name: payment_methods; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.payment_methods (id, name, is_active, is_system, created_at, updated_at, type, display_order) FROM stdin;
physical_card	實體信用卡	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	physical_card	0
online_card	線上刷卡	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	online_card	0
apple_pay	Apple Pay	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
google_pay	Google Pay	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
samsung_pay	Samsung Pay	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
line_pay	LINE Pay	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
jkopay	街口支付	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
easy_wallet	悠遊付	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
px_pay_plus	全支付	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
fami_pay	全盈+PAY	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
ipass_money	iPASS MONEY	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
icash_pay	icash Pay	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
open_wallet	OPEN錢包	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
esun_wallet	玉山Wallet	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	mobile_payment	0
easycard	悠遊卡功能	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	electronic_ticket	0
icash	icash 功能	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	electronic_ticket	0
ipass	一卡通功能	t	t	2026-07-09 12:00:40.99291+00	2026-07-09 12:00:40.99291+00	electronic_ticket	0
any_payment	不限支付方式	t	t	2026-07-09 12:00:42.142922+00	2026-07-09 12:00:42.142922+00	mobile_payment	0
\.


--
-- Data for Name: reward_allocations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reward_allocations (id, transaction_id, benefit_id, reward_unit_id, uncapped_reward, allocated_reward, preference_weight, score, created_at) FROM stdin;
\.


--
-- Data for Name: reward_preferences; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reward_preferences (user_id, reward_unit_id, weight, updated_at) FROM stdin;
\.


--
-- Data for Name: reward_units; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reward_units (id, name, symbol, "precision", is_system, updated_at, symbol_position, twd_rate) FROM stdin;
00000000-0000-0000-0000-000000000101	現金	NT$	2	t	2026-07-09 12:00:39.513623+00	prefix	1.000000
00000000-0000-0000-0000-000000000102	點數	點	2	t	2026-07-09 12:00:39.513623+00	suffix	1.000000
00000000-0000-0000-0000-000000000103	里程	哩	2	t	2026-07-09 12:00:39.513623+00	suffix	1.000000
00000000-0000-0000-0000-000000000104	永豐豐點	點	2	f	2026-07-09 12:00:40.436804+00	suffix	1.000000
00000000-0000-0000-0000-000000000105	OPENPOINT	點	2	f	2026-07-09 12:00:40.436804+00	suffix	1.000000
00000000-0000-0000-0000-000000000106	玉山 e point	點	0	f	2026-07-09 12:00:41.28712+00	suffix	1.000000
00000000-0000-0000-0000-000000000107	P幣	P	0	f	2026-07-09 12:00:43.21622+00	suffix	1.000000
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
1	2026-07-09 12:00:39.337044+00
2	2026-07-09 12:00:39.688062+00
3	2026-07-09 12:00:39.972264+00
4	2026-07-09 12:00:40.306149+00
5	2026-07-09 12:00:40.582171+00
6	2026-07-09 12:00:40.863205+00
7	2026-07-09 12:00:41.152488+00
8	2026-07-09 12:00:41.432741+00
9	2026-07-09 12:00:41.708016+00
10	2026-07-09 12:00:41.995776+00
11	2026-07-09 12:00:42.292348+00
12	2026-07-09 12:00:42.560846+00
13	2026-07-09 12:00:42.826802+00
14	2026-07-09 12:00:43.08868+00
15	2026-07-09 12:00:43.364672+00
16	2026-07-09 12:00:43.739422+00
17	2026-07-09 12:00:44.012708+00
18	2026-07-09 12:00:44.263233+00
19	2026-07-09 12:00:44.543172+00
20	2026-07-09 12:00:44.866572+00
21	2026-07-09 12:00:45.176863+00
22	2026-07-09 12:00:45.454484+00
23	2026-07-09 12:00:45.719109+00
24	2026-07-09 12:00:45.992585+00
25	2026-07-09 12:00:46.351677+00
26	2026-07-09 12:00:46.656552+00
27	2026-07-09 12:00:46.929656+00
28	2026-07-09 12:00:47.201867+00
29	2026-07-09 12:00:47.459445+00
30	2026-07-09 12:00:47.729412+00
31	2026-07-09 12:00:48.004041+00
32	2026-07-09 12:00:48.289559+00
33	2026-07-09 12:00:48.601406+00
34	2026-07-09 12:00:48.879546+00
\.


--
-- Data for Name: telegram_chat_bindings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.telegram_chat_bindings (chat_id, user_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.transactions (id, user_id, card_id, amount_minor, category_id, transaction_date, note, created_at, updated_at, merchant_name, payment_method_id, currency_id, country_id, channel, status, merchant_id, network) FROM stdin;
\.


--
-- Data for Name: activities; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activities (id, bank_id, card_product_id, title, description, source_url, effective_from, effective_to, is_active, created_at, updated_at, published_at, published_by, published_checksum) FROM stdin;
\.


--
-- Data for Name: activity_benefits; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activity_benefits (id, reward_component_id, benefit_type, value, reward_unit_id, cap_amount, cap_period, description, is_active, created_at, updated_at, cap_formula) FROM stdin;
\.


--
-- Data for Name: activity_component_groups; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activity_component_groups (reward_component_id, reward_group_id) FROM stdin;
\.


--
-- Data for Name: activity_components; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activity_components (id, name, description, layer, stack_group, stack_mode, priority, effective_from, effective_to, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: activity_groups; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activity_groups (id, activity_id, name, description, display_order, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: activity_requirements; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.activity_requirements (id, reward_component_id, requirement_type, operator, configuration_json, description, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: migration_reports; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.migration_reports (id, migration_version, severity, subject_type, subject_id, message, details_json, created_at) FROM stdin;
\.


--
-- Data for Name: published_activities; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.published_activities (id, source_activity_id, bank_id, card_product_id, title, description, source_url, effective_from, effective_to, published_at, published_by, source_checksum) FROM stdin;
\.


--
-- Data for Name: published_reward_rules; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.published_reward_rules (id, published_activity_id, source_activity_id, source_component_id, source_benefit_id, name, description, layer, display_order, stack_group, stack_policy, priority, effect_type, reward_value, reward_unit_id, cap_amount, cap_formula, cap_period, effective_from, effective_to, is_active, published_at) FROM stdin;
\.


--
-- Data for Name: published_rule_benefits; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.published_rule_benefits (id, published_rule_id, source_benefit_id, benefit_type, value, reward_unit_id, cap_amount, cap_formula, cap_period, description, published_at) FROM stdin;
\.


--
-- Data for Name: published_rule_requirements; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.published_rule_requirements (id, published_rule_id, source_requirement_id, requirement_type, operator, configuration_json, description, display_order, published_at) FROM stdin;
\.


--
-- Data for Name: requirement_types; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.requirement_types (id, name, value_key, value_source, display_order, is_active, created_at, updated_at) FROM stdin;
PAYMENT_METHOD	支付方式	payment_method_codes	payment_methods	10	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
CARD_PLAN	卡方案	card_plan_ids	card_plans	30	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
CARD_PRODUCT	信用卡	card_product_ids	card_products	40	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
MERCHANT	店家	merchant_ids	merchants	50	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
MERCHANT_CATEGORY	店家分類	category_ids	categories	60	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
CONSUMPTION_CATEGORY	消費分類	category_ids	categories	70	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
AMOUNT	金額	amount	number	80	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
ACCOUNT_TIER	帳戶等級	tiers	account_tiers	90	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
USER_QUALIFICATION	會員資格	qualification_codes	user_qualifications	100	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
DATE_RANGE	日期區間	date_range	date_range	110	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
WEEKDAY	星期	weekdays	weekdays	120	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
TIME_RANGE	時間區間	time_range	time_range	130	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
CHANNEL	通路	channels	channels	140	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
REGION	地區	regions	regions	150	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
ACTION_REQUIRED	必要操作	action_codes	actions	160	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
INSTALLMENT	是否分期	is_installment	boolean	85	t	2026-07-09 12:00:47.331407+00	2026-07-09 12:00:47.331407+00
CARD_NETWORK	卡組織	networks	card_networks	20	t	2026-07-09 12:00:46.803856+00	2026-07-09 12:00:46.803856+00
\.


--
-- Data for Name: reward_calculation_components; Type: TABLE DATA; Schema: transaction; Owner: -
--

COPY transaction.reward_calculation_components (id, reward_calculation_id, reward_component_id, reward_unit_id, suggested_card_plan_id, qualified_card_plan_id, member_qualification_status_id, uncapped_reward, allocated_reward, cap_used_before, cap_remaining_before, condition_snapshot_json, reminder_snapshot_json, effect_type, reward_value, reward_rate) FROM stdin;
\.


--
-- Data for Name: reward_calculations; Type: TABLE DATA; Schema: transaction; Owner: -
--

COPY transaction.reward_calculations (id, transaction_id, calculation_no, calculation_type, transaction_at, total_effective_rate, total_reward_value, engine_version, status, supersedes_calculation_id, input_snapshot_json, calculated_at) FROM stdin;
\.


--
-- Name: card_networks card_networks_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_networks
    ADD CONSTRAINT card_networks_pkey PRIMARY KEY (id);


--
-- Name: card_plan_versions card_plan_versions_card_plan_id_tstzrange_excl; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plan_versions
    ADD CONSTRAINT card_plan_versions_card_plan_id_tstzrange_excl EXCLUDE USING gist (card_plan_id WITH =, tstzrange(effective_from, effective_to, '[)'::text) WITH &&);


--
-- Name: card_plan_versions card_plan_versions_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plan_versions
    ADD CONSTRAINT card_plan_versions_pkey PRIMARY KEY (id);


--
-- Name: card_plans card_plans_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plans
    ADD CONSTRAINT card_plans_pkey PRIMARY KEY (id);


--
-- Name: card_products card_products_bank_id_name_key; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_products
    ADD CONSTRAINT card_products_bank_id_name_key UNIQUE (bank_id, name);


--
-- Name: card_products card_products_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_products
    ADD CONSTRAINT card_products_pkey PRIMARY KEY (id);


--
-- Name: member_card_qualification_statuses member_card_qualification_sta_member_card_id_card_plan_id__excl; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.member_card_qualification_statuses
    ADD CONSTRAINT member_card_qualification_sta_member_card_id_card_plan_id__excl EXCLUDE USING gist (member_card_id WITH =, card_plan_id WITH =, tstzrange(effective_from, effective_to, '[)'::text) WITH &&);


--
-- Name: member_card_qualification_statuses member_card_qualification_statuses_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.member_card_qualification_statuses
    ADD CONSTRAINT member_card_qualification_statuses_pkey PRIMARY KEY (id);


--
-- Name: regions regions_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.regions
    ADD CONSTRAINT regions_pkey PRIMARY KEY (id);


--
-- Name: user_qualifications user_qualifications_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.user_qualifications
    ADD CONSTRAINT user_qualifications_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.invitations
    ADD CONSTRAINT invitations_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_token_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.invitations
    ADD CONSTRAINT invitations_token_hash_key UNIQUE (token_hash);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_token_hash_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.sessions
    ADD CONSTRAINT sessions_token_hash_key UNIQUE (token_hash);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: integration; Owner: -
--

ALTER TABLE ONLY integration.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);


--
-- Name: user_payment_methods user_payment_methods_pkey; Type: CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_pkey PRIMARY KEY (id);


--
-- Name: user_payment_methods user_payment_methods_user_id_payment_method_code_key; Type: CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_user_id_payment_method_code_key UNIQUE (user_id, payment_method_id);


--
-- Name: banks banks_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banks
    ADD CONSTRAINT banks_code_key UNIQUE (code);


--
-- Name: banks banks_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banks
    ADD CONSTRAINT banks_name_key UNIQUE (name);


--
-- Name: banks banks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banks
    ADD CONSTRAINT banks_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: member_card_credit_limits member_card_credit_limits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_card_credit_limits
    ADD CONSTRAINT member_card_credit_limits_pkey PRIMARY KEY (id);


--
-- Name: member_cards member_cards_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_id_user_id_key UNIQUE (id, user_id);


--
-- Name: member_cards member_cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_pkey PRIMARY KEY (id);


--
-- Name: merchant_aliases merchant_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchant_aliases
    ADD CONSTRAINT merchant_aliases_pkey PRIMARY KEY (merchant_id, alias);


--
-- Name: merchant_categories merchant_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchant_categories
    ADD CONSTRAINT merchant_categories_pkey PRIMARY KEY (merchant_id, category_id);


--
-- Name: merchants merchants_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchants
    ADD CONSTRAINT merchants_name_key UNIQUE (name);


--
-- Name: merchants merchants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchants
    ADD CONSTRAINT merchants_pkey PRIMARY KEY (id);


--
-- Name: payment_methods payment_methods_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_name_key UNIQUE (name);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: reward_allocations reward_allocations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_pkey PRIMARY KEY (id);


--
-- Name: reward_allocations reward_allocations_transaction_id_benefit_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_transaction_id_benefit_id_key UNIQUE (transaction_id, benefit_id);


--
-- Name: reward_preferences reward_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_preferences
    ADD CONSTRAINT reward_preferences_pkey PRIMARY KEY (user_id, reward_unit_id);


--
-- Name: reward_units reward_units_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_units
    ADD CONSTRAINT reward_units_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: telegram_chat_bindings telegram_chat_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_chat_bindings
    ADD CONSTRAINT telegram_chat_bindings_pkey PRIMARY KEY (chat_id);


--
-- Name: telegram_chat_bindings telegram_chat_bindings_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_chat_bindings
    ADD CONSTRAINT telegram_chat_bindings_user_id_key UNIQUE (user_id);


--
-- Name: transactions transactions_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_id_user_id_key UNIQUE (id, user_id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: activities activities_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);


--
-- Name: activity_benefits activity_benefits_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_benefits
    ADD CONSTRAINT activity_benefits_pkey PRIMARY KEY (id);


--
-- Name: activity_component_groups activity_component_groups_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_component_groups
    ADD CONSTRAINT activity_component_groups_pkey PRIMARY KEY (reward_component_id, reward_group_id);


--
-- Name: activity_components activity_components_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_components
    ADD CONSTRAINT activity_components_pkey PRIMARY KEY (id);


--
-- Name: activity_groups activity_groups_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_groups
    ADD CONSTRAINT activity_groups_pkey PRIMARY KEY (id);


--
-- Name: activity_requirements activity_requirements_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_requirements
    ADD CONSTRAINT activity_requirements_pkey PRIMARY KEY (id);


--
-- Name: migration_reports migration_reports_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.migration_reports
    ADD CONSTRAINT migration_reports_pkey PRIMARY KEY (id);


--
-- Name: published_activities published_activities_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_activities
    ADD CONSTRAINT published_activities_pkey PRIMARY KEY (id);


--
-- Name: published_reward_rules published_reward_rules_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_reward_rules
    ADD CONSTRAINT published_reward_rules_pkey PRIMARY KEY (id);


--
-- Name: published_rule_benefits published_rule_benefits_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_rule_benefits
    ADD CONSTRAINT published_rule_benefits_pkey PRIMARY KEY (id);


--
-- Name: published_rule_requirements published_rule_requirements_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_rule_requirements
    ADD CONSTRAINT published_rule_requirements_pkey PRIMARY KEY (id);


--
-- Name: requirement_types requirement_types_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.requirement_types
    ADD CONSTRAINT requirement_types_pkey PRIMARY KEY (id);


--
-- Name: reward_calculation_components reward_calculation_components_pkey; Type: CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_pkey PRIMARY KEY (id);


--
-- Name: reward_calculations reward_calculations_pkey; Type: CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculations
    ADD CONSTRAINT reward_calculations_pkey PRIMARY KEY (id);


--
-- Name: reward_calculations reward_calculations_transaction_id_calculation_no_key; Type: CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculations
    ADD CONSTRAINT reward_calculations_transaction_id_calculation_no_key UNIQUE (transaction_id, calculation_no);


--
-- Name: card_products_bank_active_idx; Type: INDEX; Schema: catalog; Owner: -
--

CREATE INDEX card_products_bank_active_idx ON catalog.card_products USING btree (bank_id, is_active);


--
-- Name: invitations_token_hash_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX invitations_token_hash_idx ON identity.invitations USING btree (token_hash);


--
-- Name: sessions_token_hash_idx; Type: INDEX; Schema: identity; Owner: -
--

CREATE INDEX sessions_token_hash_idx ON identity.sessions USING btree (token_hash);


--
-- Name: outbox_unpublished_idx; Type: INDEX; Schema: integration; Owner: -
--

CREATE INDEX outbox_unpublished_idx ON integration.outbox_events USING btree (occurred_at) WHERE (published_at IS NULL);


--
-- Name: member_card_credit_limits_current_uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX member_card_credit_limits_current_uidx ON public.member_card_credit_limits USING btree (member_card_id) WHERE (effective_to IS NULL);


--
-- Name: member_card_credit_limits_lookup_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX member_card_credit_limits_lookup_idx ON public.member_card_credit_limits USING btree (member_card_id, effective_from DESC);


--
-- Name: member_cards_user_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX member_cards_user_active_idx ON public.member_cards USING btree (user_id, is_active);


--
-- Name: member_cards_user_product_network_uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX member_cards_user_product_network_uidx ON public.member_cards USING btree (user_id, card_product_id, network) NULLS NOT DISTINCT;


--
-- Name: reward_allocations_benefit_transaction_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reward_allocations_benefit_transaction_idx ON public.reward_allocations USING btree (benefit_id, transaction_id);


--
-- Name: transactions_merchant_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_merchant_id_idx ON public.transactions USING btree (merchant_id);


--
-- Name: transactions_payment_method_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_payment_method_idx ON public.transactions USING btree (payment_method_id);


--
-- Name: transactions_user_card_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_user_card_date_idx ON public.transactions USING btree (user_id, card_id, transaction_date);


--
-- Name: transactions_user_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_user_date_idx ON public.transactions USING btree (user_id, transaction_date);


--
-- Name: published_activities_card_date_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX published_activities_card_date_idx ON reward.published_activities USING btree (card_product_id, effective_from, effective_to);


--
-- Name: published_activities_source_activity_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE UNIQUE INDEX published_activities_source_activity_idx ON reward.published_activities USING btree (source_activity_id);


--
-- Name: published_reward_rules_activity_date_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX published_reward_rules_activity_date_idx ON reward.published_reward_rules USING btree (published_activity_id, effective_from, effective_to, is_active);


--
-- Name: published_rule_benefits_rule_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX published_rule_benefits_rule_idx ON reward.published_rule_benefits USING btree (published_rule_id);


--
-- Name: published_rule_requirements_rule_type_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX published_rule_requirements_rule_type_idx ON reward.published_rule_requirements USING btree (published_rule_id, requirement_type);


--
-- Name: reward_activities_card_active_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activities_card_active_idx ON reward.activities USING btree (card_product_id, effective_from, effective_to, is_active);


--
-- Name: reward_activity_benefits_component_type_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activity_benefits_component_type_idx ON reward.activity_benefits USING btree (reward_component_id, benefit_type);


--
-- Name: reward_activity_component_groups_group_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activity_component_groups_group_idx ON reward.activity_component_groups USING btree (reward_group_id, reward_component_id);


--
-- Name: reward_activity_components_order_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activity_components_order_idx ON reward.activity_components USING btree (layer, priority, name);


--
-- Name: reward_activity_groups_activity_order_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activity_groups_activity_order_idx ON reward.activity_groups USING btree (activity_id, display_order);


--
-- Name: reward_activity_requirements_component_type_idx; Type: INDEX; Schema: reward; Owner: -
--

CREATE INDEX reward_activity_requirements_component_type_idx ON reward.activity_requirements USING btree (reward_component_id, requirement_type);


--
-- Name: reward_calculation_active_uidx; Type: INDEX; Schema: transaction; Owner: -
--

CREATE UNIQUE INDEX reward_calculation_active_uidx ON transaction.reward_calculations USING btree (transaction_id) WHERE (status = 'active'::text);


--
-- Name: member_cards member_card_network_check; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER member_card_network_check BEFORE INSERT OR UPDATE OF card_product_id, network ON public.member_cards FOR EACH ROW EXECUTE FUNCTION catalog.validate_member_card_network();


--
-- Name: card_plan_versions card_plan_versions_card_plan_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plan_versions
    ADD CONSTRAINT card_plan_versions_card_plan_id_fkey FOREIGN KEY (card_plan_id) REFERENCES catalog.card_plans(id) ON DELETE CASCADE;


--
-- Name: card_plan_versions card_plan_versions_supersedes_version_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plan_versions
    ADD CONSTRAINT card_plan_versions_supersedes_version_id_fkey FOREIGN KEY (supersedes_version_id) REFERENCES catalog.card_plan_versions(id);


--
-- Name: card_plans card_plans_card_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_plans
    ADD CONSTRAINT card_plans_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES catalog.card_products(id) ON DELETE CASCADE;


--
-- Name: card_products card_products_bank_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_products
    ADD CONSTRAINT card_products_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES public.banks(id);


--
-- Name: member_card_qualification_statuses member_card_qualification_statuses_card_plan_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.member_card_qualification_statuses
    ADD CONSTRAINT member_card_qualification_statuses_card_plan_id_fkey FOREIGN KEY (card_plan_id) REFERENCES catalog.card_plans(id) ON DELETE CASCADE;


--
-- Name: member_card_qualification_statuses member_card_qualification_statuses_member_card_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.member_card_qualification_statuses
    ADD CONSTRAINT member_card_qualification_statuses_member_card_id_fkey FOREIGN KEY (member_card_id) REFERENCES public.member_cards(id) ON DELETE CASCADE;


--
-- Name: invitations invitations_invited_by_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.invitations
    ADD CONSTRAINT invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES identity.users(id);


--
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: user_payment_methods user_payment_methods_payment_method_code_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_payment_method_code_fkey FOREIGN KEY (payment_method_id) REFERENCES public.payment_methods(id) ON DELETE CASCADE;


--
-- Name: user_payment_methods user_payment_methods_user_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: member_card_credit_limits member_card_credit_limits_member_card_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_card_credit_limits
    ADD CONSTRAINT member_card_credit_limits_member_card_id_fkey FOREIGN KEY (member_card_id) REFERENCES public.member_cards(id) ON DELETE CASCADE;


--
-- Name: member_cards member_cards_card_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES catalog.card_products(id);


--
-- Name: member_cards member_cards_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: merchant_aliases merchant_aliases_merchant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchant_aliases
    ADD CONSTRAINT merchant_aliases_merchant_id_fkey FOREIGN KEY (merchant_id) REFERENCES public.merchants(id) ON DELETE CASCADE;


--
-- Name: merchant_categories merchant_categories_category_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchant_categories
    ADD CONSTRAINT merchant_categories_category_code_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: merchant_categories merchant_categories_merchant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.merchant_categories
    ADD CONSTRAINT merchant_categories_merchant_id_fkey FOREIGN KEY (merchant_id) REFERENCES public.merchants(id) ON DELETE CASCADE;


--
-- Name: reward_allocations reward_allocations_published_rule_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_published_rule_fkey FOREIGN KEY (benefit_id) REFERENCES reward.published_reward_rules(id);


--
-- Name: reward_allocations reward_allocations_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: reward_allocations reward_allocations_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE;


--
-- Name: reward_preferences reward_preferences_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_preferences
    ADD CONSTRAINT reward_preferences_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: reward_preferences reward_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_preferences
    ADD CONSTRAINT reward_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: telegram_chat_bindings telegram_chat_bindings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_chat_bindings
    ADD CONSTRAINT telegram_chat_bindings_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_category_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_category_code_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: transactions transactions_member_card_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_member_card_fkey FOREIGN KEY (card_id, user_id) REFERENCES public.member_cards(id, user_id);


--
-- Name: transactions transactions_merchant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_merchant_id_fkey FOREIGN KEY (merchant_id) REFERENCES public.merchants(id);


--
-- Name: transactions transactions_payment_method_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_payment_method_code_fkey FOREIGN KEY (payment_method_id) REFERENCES public.payment_methods(id);


--
-- Name: transactions transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: activities activities_bank_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activities
    ADD CONSTRAINT activities_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES public.banks(id);


--
-- Name: activities activities_card_product_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activities
    ADD CONSTRAINT activities_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES catalog.card_products(id);


--
-- Name: activities activities_published_by_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activities
    ADD CONSTRAINT activities_published_by_fkey FOREIGN KEY (published_by) REFERENCES identity.users(id);


--
-- Name: activity_benefits activity_benefits_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_benefits
    ADD CONSTRAINT activity_benefits_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.activity_components(id) ON DELETE CASCADE;


--
-- Name: activity_benefits activity_benefits_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_benefits
    ADD CONSTRAINT activity_benefits_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: activity_component_groups activity_component_groups_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_component_groups
    ADD CONSTRAINT activity_component_groups_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.activity_components(id) ON DELETE CASCADE;


--
-- Name: activity_component_groups activity_component_groups_reward_group_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_component_groups
    ADD CONSTRAINT activity_component_groups_reward_group_id_fkey FOREIGN KEY (reward_group_id) REFERENCES reward.activity_groups(id) ON DELETE CASCADE;


--
-- Name: activity_groups activity_groups_activity_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_groups
    ADD CONSTRAINT activity_groups_activity_id_fkey FOREIGN KEY (activity_id) REFERENCES reward.activities(id) ON DELETE CASCADE;


--
-- Name: activity_requirements activity_requirements_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.activity_requirements
    ADD CONSTRAINT activity_requirements_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.activity_components(id) ON DELETE CASCADE;


--
-- Name: published_activities published_activities_bank_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_activities
    ADD CONSTRAINT published_activities_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES public.banks(id);


--
-- Name: published_activities published_activities_card_product_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_activities
    ADD CONSTRAINT published_activities_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES catalog.card_products(id);


--
-- Name: published_activities published_activities_published_by_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_activities
    ADD CONSTRAINT published_activities_published_by_fkey FOREIGN KEY (published_by) REFERENCES identity.users(id);


--
-- Name: published_activities published_activities_source_activity_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_activities
    ADD CONSTRAINT published_activities_source_activity_id_fkey FOREIGN KEY (source_activity_id) REFERENCES reward.activities(id) ON DELETE CASCADE;


--
-- Name: published_reward_rules published_reward_rules_published_activity_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_reward_rules
    ADD CONSTRAINT published_reward_rules_published_activity_id_fkey FOREIGN KEY (published_activity_id) REFERENCES reward.published_activities(id) ON DELETE CASCADE;


--
-- Name: published_reward_rules published_reward_rules_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_reward_rules
    ADD CONSTRAINT published_reward_rules_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: published_reward_rules published_reward_rules_source_activity_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_reward_rules
    ADD CONSTRAINT published_reward_rules_source_activity_id_fkey FOREIGN KEY (source_activity_id) REFERENCES reward.activities(id) ON DELETE CASCADE;


--
-- Name: published_rule_benefits published_rule_benefits_published_rule_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_rule_benefits
    ADD CONSTRAINT published_rule_benefits_published_rule_id_fkey FOREIGN KEY (published_rule_id) REFERENCES reward.published_reward_rules(id) ON DELETE CASCADE;


--
-- Name: published_rule_benefits published_rule_benefits_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_rule_benefits
    ADD CONSTRAINT published_rule_benefits_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: published_rule_requirements published_rule_requirements_published_rule_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.published_rule_requirements
    ADD CONSTRAINT published_rule_requirements_published_rule_id_fkey FOREIGN KEY (published_rule_id) REFERENCES reward.published_reward_rules(id) ON DELETE CASCADE;


--
-- Name: reward_calculation_components reward_calculation_components_member_qualification_status__fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_member_qualification_status__fkey FOREIGN KEY (member_qualification_status_id) REFERENCES catalog.member_card_qualification_statuses(id);


--
-- Name: reward_calculation_components reward_calculation_components_published_rule_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_published_rule_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.published_reward_rules(id);


--
-- Name: reward_calculation_components reward_calculation_components_qualified_card_plan_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_qualified_card_plan_id_fkey FOREIGN KEY (qualified_card_plan_id) REFERENCES catalog.card_plans(id);


--
-- Name: reward_calculation_components reward_calculation_components_reward_calculation_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_reward_calculation_id_fkey FOREIGN KEY (reward_calculation_id) REFERENCES transaction.reward_calculations(id) ON DELETE CASCADE;


--
-- Name: reward_calculation_components reward_calculation_components_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: reward_calculation_components reward_calculation_components_suggested_card_plan_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_suggested_card_plan_id_fkey FOREIGN KEY (suggested_card_plan_id) REFERENCES catalog.card_plans(id);


--
-- Name: reward_calculations reward_calculations_supersedes_calculation_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculations
    ADD CONSTRAINT reward_calculations_supersedes_calculation_id_fkey FOREIGN KEY (supersedes_calculation_id) REFERENCES transaction.reward_calculations(id);


--
-- Name: reward_calculations reward_calculations_transaction_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculations
    ADD CONSTRAINT reward_calculations_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict vHn0oR01hWcC1gOcKIOUrxgrXwJ9sMCUcEUWQBykt2OeOML3gnBRF06t401iweP

