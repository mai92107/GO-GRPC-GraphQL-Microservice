--
-- PostgreSQL database dump
--

\restrict Z6rz96Vi92MZ1qenYMdDbKgBgB6sJDzNicWVkLRStc6wo8F3yUMffvLVIOCboMg

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
-- Name: validate_member_card_network(); Type: FUNCTION; Schema: catalog; Owner: -
--

CREATE FUNCTION catalog.validate_member_card_network() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.card_network_id IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM catalog.card_product_networks
        WHERE card_product_id=NEW.card_product_id AND card_network_id=NEW.card_network_id
    ) THEN
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
-- Name: card_product_networks; Type: TABLE; Schema: catalog; Owner: -
--

CREATE TABLE catalog.card_product_networks (
    card_product_id uuid NOT NULL,
    card_network_id uuid NOT NULL
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
-- Name: reward_usage_adjustments; Type: TABLE; Schema: member_profile; Owner: -
--

CREATE TABLE member_profile.reward_usage_adjustments (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    member_card_id uuid NOT NULL,
    reward_cap_id uuid NOT NULL,
    period_start date NOT NULL,
    period_end date NOT NULL,
    amount numeric(18,6) NOT NULL,
    reward_unit_id uuid NOT NULL,
    reason text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reward_usage_adjustments_check CHECK ((period_end >= period_start))
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
-- Name: card_products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE catalog.card_products (
    id uuid NOT NULL,
    bank_id uuid NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    description text,
    card_image_url text,
    primary_color text,
    qualified_type text,
    selectable_type text,
    networks    text
    );


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invitations (
    id uuid NOT NULL,
    email public.citext NOT NULL,
    token_hash bytea NOT NULL,
    invited_by uuid NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
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
    card_network_id uuid,
    opened_on date,
    closed_on date,
    CONSTRAINT member_cards_account_tier_valid CHECK (((account_tier IS NULL) OR (btrim(account_tier) <> ''::text))),
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
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: -
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
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
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
    card_network_id uuid,
    currency_id character(3) DEFAULT 'TWD'::bpchar NOT NULL,
    country_id character(2) DEFAULT 'TW'::bpchar NOT NULL,
    channel text DEFAULT 'physical'::text NOT NULL,
    status text DEFAULT 'confirmed'::text NOT NULL,
    merchant_id uuid,
    CONSTRAINT transactions_amount_minor_check CHECK ((amount_minor > 0)),
    CONSTRAINT transactions_category_code_check CHECK ((category_id <> 'general'::text)),
    CONSTRAINT transactions_channel_check CHECK ((channel = ANY (ARRAY['online'::text, 'physical'::text]))),
    CONSTRAINT transactions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'confirmed'::text, 'cancelled'::text, 'refunded'::text])))
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
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
-- Name: cap_versions; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.cap_versions (
    id uuid NOT NULL,
    reward_cap_id uuid NOT NULL,
    cap_type text NOT NULL,
    limit_value numeric(18,6) NOT NULL,
    reward_unit_id uuid,
    period_type text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    supersedes_version_id uuid,
    CONSTRAINT cap_versions_cap_type_check CHECK ((cap_type = ANY (ARRAY['reward_amount'::text, 'spending_amount'::text, 'transaction_count'::text]))),
    CONSTRAINT cap_versions_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT cap_versions_limit_value_check CHECK ((limit_value > (0)::numeric)),
    CONSTRAINT cap_versions_period_type_check CHECK ((period_type = ANY (ARRAY['calendar_month'::text, 'statement_cycle'::text, 'campaign'::text, 'year'::text])))
);


--
-- Name: caps; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.caps (
    id uuid NOT NULL,
    reward_program_id uuid NOT NULL,
    name text NOT NULL,
    scope text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    CONSTRAINT caps_scope_check CHECK ((scope = ANY (ARRAY['component'::text, 'stack_group'::text, 'program'::text, 'card'::text, 'customer'::text])))
);


--
-- Name: component_caps; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.component_caps (
    reward_component_id uuid NOT NULL,
    reward_cap_id uuid NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone
);


--
-- Name: component_versions; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.component_versions (
    id uuid NOT NULL,
    reward_component_id uuid NOT NULL,
    reward_unit_id uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    announced_at timestamp with time zone,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    supersedes_version_id uuid,
    change_reason text DEFAULT ''::text NOT NULL,
    display_change_until timestamp with time zone,
    effect_type text DEFAULT 'ADD_RATE'::text NOT NULL,
    reward_value numeric(18,8) DEFAULT 0 NOT NULL,
    CONSTRAINT component_versions_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT component_versions_effect_type_check CHECK ((effect_type = ANY (ARRAY['ADD_RATE'::text, 'SET_RATE'::text, 'MULTIPLY_RATE'::text, 'ADD_CASH'::text, 'DISCOUNT'::text])))
);


--
-- Name: components; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.components (
    id uuid NOT NULL,
    reward_program_id uuid NOT NULL,
    stack_group text NOT NULL,
    stack_policy text DEFAULT 'best_of_group'::text NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    exclusive_group text DEFAULT ''::text NOT NULL,
    layer integer DEFAULT 1 NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    CONSTRAINT components_stack_policy_check CHECK ((stack_policy = ANY (ARRAY['stack'::text, 'best_of_group'::text, 'exclusive'::text])))
);


--
-- Name: condition_versions; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.condition_versions (
    id uuid NOT NULL,
    reward_condition_id uuid NOT NULL,
    operator text NOT NULL,
    configuration_json jsonb NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    published_at timestamp with time zone DEFAULT now() NOT NULL,
    supersedes_version_id uuid,
    CONSTRAINT condition_versions_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT condition_versions_operator_check CHECK ((operator = ANY (ARRAY['equals'::text, 'in'::text, 'not_in'::text, 'gte'::text, 'lte'::text, 'between'::text])))
);


--
-- Name: conditions; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.conditions (
    id uuid NOT NULL,
    condition_type text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    CONSTRAINT conditions_condition_type_check CHECK ((condition_type = ANY (ARRAY['category'::text, 'merchant'::text, 'payment_method'::text, 'card_network'::text, 'card_plan'::text, 'country'::text, 'currency'::text, 'channel'::text, 'amount_range'::text])))
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
-- Name: programs; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.programs (
    id uuid NOT NULL,
    card_product_id uuid NOT NULL,
    name text NOT NULL,
    source_url text,
    verified_at timestamp with time zone,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT programs_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text])))
);


--
-- Name: requirements; Type: TABLE; Schema: reward; Owner: -
--

CREATE TABLE reward.requirements (
    id uuid NOT NULL,
    reward_component_id uuid NOT NULL,
    reward_condition_id uuid NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    display_order integer DEFAULT 0 NOT NULL,
    CONSTRAINT requirements_check CHECK (((effective_to IS NULL) OR (effective_to > effective_from)))
);


--
-- Name: reward_calculation_components; Type: TABLE; Schema: transaction; Owner: -
--

CREATE TABLE transaction.reward_calculation_components (
    id uuid NOT NULL,
    reward_calculation_id uuid NOT NULL,
    reward_component_id uuid,
    reward_component_version_id uuid,
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
45342351-0a0a-a4e9-a341-fa805fe21767	654c3131-2c29-2a4e-90cd-4cd7e62e5079	大大	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-06-28 14:47:49.429931+00	\N
ce87c2ed-e6b2-f0f4-5bc0-016cc0dc1cb7	21c503d8-00a6-6402-3c54-25e9001a3b95	大戶	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-06-28 14:47:49.429931+00	\N
bef0e8c8-8838-4632-8fd8-a9f9e45bb427	c94d64f7-fcd8-8e6d-3720-21a4216fc823	大戶Plus	由會員自行確認是否符合銀行資格		2000-01-01 00:00:00+00	\N	\N	2026-06-28 14:47:49.429931+00	\N
225c5b87-2617-b160-f6ca-cd8c09870258	72b99c99-4ffc-15a8-6481-e32444466701	Richart 指定方案加碼	需於 Richart Life APP 切換至符合消費情境的方案	需於 Richart Life APP 切換至符合消費情境的方案	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N
8d9f58f0-1dc7-c863-7500-3bfce0bfcb01	ca407472-df08-104b-008a-09212863d809	UP 選指定行動支付最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N
295dba5a-7bf2-c264-98ef-0a3edc08d543	a40bb29c-13b8-33be-e7fb-30ef9262d339	UP 選百大指定消費最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N
0986110b-e5d8-91b9-c5af-44f0b804e398	ab2f95ab-2d7f-296a-5580-598ba377c568	UP 選國外實體消費最高 4.5%	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N
3d5896a5-1adf-ac5f-115f-108af24fa833	12831358-db59-f1bf-0606-9a32cea894b1	大戶	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.582844+00	\N
9a43c148-47e0-cc3b-5581-de935b233628	edf55cc5-435f-734f-3420-2e4e99f0a67f	大戶	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.582844+00	\N
\.


--
-- Data for Name: card_plans; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_plans (id, card_product_id, plan_type, is_active, display_order, created_at) FROM stdin;
654c3131-2c29-2a4e-90cd-4cd7e62e5079	51000000-0000-0000-0000-000000000001	qualified	t	1	2026-06-28 14:47:49.427916+00
21c503d8-00a6-6402-3c54-25e9001a3b95	51000000-0000-0000-0000-000000000001	qualified	t	2	2026-06-28 14:47:49.427916+00
c94d64f7-fcd8-8e6d-3720-21a4216fc823	51000000-0000-0000-0000-000000000001	qualified	t	3	2026-06-28 14:47:49.427916+00
72b99c99-4ffc-15a8-6481-e32444466701	51000000-0000-0000-0000-000000000003	selectable	t	20	2026-06-28 14:47:49.476047+00
ca407472-df08-104b-008a-09212863d809	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-06-28 14:47:49.476047+00
a40bb29c-13b8-33be-e7fb-30ef9262d339	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-06-28 14:47:49.476047+00
ab2f95ab-2d7f-296a-5580-598ba377c568	51000000-0000-0000-0000-000000000006	selectable	t	20	2026-06-28 14:47:49.476047+00
12831358-db59-f1bf-0606-9a32cea894b1	51000000-0000-0000-0000-000000000001	selectable	t	20	2026-06-28 14:47:49.582187+00
edf55cc5-435f-734f-3420-2e4e99f0a67f	51000000-0000-0000-0000-000000000001	selectable	t	30	2026-06-28 14:47:49.582187+00
\.


--
-- Data for Name: card_product_networks; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.card_product_networks (card_product_id, card_network_id) FROM stdin;
51000000-0000-0000-0000-000000000002	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000002	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000002	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000003	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000003	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000003	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000004	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000004	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000004	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000005	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000005	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000005	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000006	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000006	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000006	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000007	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000007	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000007	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000001	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000001	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000001	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000008	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000008	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000008	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000009	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000009	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000009	00000000-0000-0000-0000-000000000203
51000000-0000-0000-0000-000000000010	00000000-0000-0000-0000-000000000201
51000000-0000-0000-0000-000000000010	00000000-0000-0000-0000-000000000202
51000000-0000-0000-0000-000000000010	00000000-0000-0000-0000-000000000203
\.


--
-- Data for Name: member_card_qualification_statuses; Type: TABLE DATA; Schema: catalog; Owner: -
--

COPY catalog.member_card_qualification_statuses (id, member_card_id, card_plan_id, is_qualified, effective_from, effective_to, updated_by_user_at, created_at) FROM stdin;
\.


--
-- Data for Name: outbox_events; Type: TABLE DATA; Schema: integration; Owner: -
--

COPY integration.outbox_events (id, aggregate_type, aggregate_id, event_type, payload_json, occurred_at, published_at) FROM stdin;
\.


--
-- Data for Name: reward_usage_adjustments; Type: TABLE DATA; Schema: member_profile; Owner: -
--

COPY member_profile.reward_usage_adjustments (id, user_id, member_card_id, reward_cap_id, period_start, period_end, amount, reward_unit_id, reason, created_at) FROM stdin;
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
41000000-0000-0000-0000-000000000001	永豐銀行	sinopac	https://bank.sinopac.com/	t	2026-06-28 14:47:49.177922+00	2026-06-28 14:47:49.177922+00	\N
41000000-0000-0000-0000-000000000002	台新銀行	taishin	https://www.taishinbank.com.tw/	t	2026-06-28 14:47:49.177922+00	2026-06-28 14:47:49.177922+00	\N
41000000-0000-0000-0000-000000000003	中國信託	ctbc	https://www.ctbcbank.com/	t	2026-06-28 14:47:49.177922+00	2026-06-28 14:47:49.177922+00	\N
41000000-0000-0000-0000-000000000004	玉山銀行	esun	https://www.esunbank.com/	t	2026-06-28 14:47:49.23551+00	2026-06-28 14:47:49.23551+00	\N
41000000-0000-0000-0000-000000000005	上海商銀	scsb	https://www.scsb.com.tw/	t	2026-06-28 14:47:49.37546+00	2026-06-28 14:47:49.37546+00	\N
41000000-0000-0000-0000-000000000006	聯邦銀行	ubot	https://www.ubot.com.tw/	t	2026-06-28 14:47:49.37546+00	2026-06-28 14:47:49.37546+00	\N
\.


--
-- Data for Name: card_products; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.card_products (id, bank_id, name, is_active, created_at, updated_at, account_tiers, description, card_image_url, primary_color, is_open_for_application, qualified_type, selectable_type) FROM stdin;
51000000-0000-0000-0000-000000000002	41000000-0000-0000-0000-000000000001	SPORT 卡	t	2026-06-28 14:47:49.17849+00	2026-06-28 14:47:49.17849+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000003	41000000-0000-0000-0000-000000000002	Richart 卡	t	2026-06-28 14:47:49.17849+00	2026-06-28 14:47:49.17849+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000004	41000000-0000-0000-0000-000000000003	uniopen 聯名卡	t	2026-06-28 14:47:49.17849+00	2026-06-28 14:47:49.17849+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000005	41000000-0000-0000-0000-000000000004	U Bear 信用卡	t	2026-06-28 14:47:49.236098+00	2026-06-28 14:47:49.236098+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000006	41000000-0000-0000-0000-000000000004	Unicard	t	2026-06-28 14:47:49.236098+00	2026-06-28 14:47:49.236098+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000007	41000000-0000-0000-0000-000000000003	英雄聯盟信用卡（已停止申辦）	t	2026-06-28 14:47:49.236098+00	2026-06-28 14:47:49.236098+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000008	41000000-0000-0000-0000-000000000004	Pi 拍錢包信用卡	t	2026-06-28 14:47:49.376002+00	2026-06-28 14:47:49.376002+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000009	41000000-0000-0000-0000-000000000005	小小兵回饋卡	t	2026-06-28 14:47:49.376002+00	2026-06-28 14:47:49.376002+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000010	41000000-0000-0000-0000-000000000006	LINE Bank聯名卡	t	2026-06-28 14:47:49.376002+00	2026-06-28 14:47:49.376002+00	{}	\N	\N	\N	t	\N	\N
51000000-0000-0000-0000-000000000001	41000000-0000-0000-0000-000000000001	DAWHO 現金回饋信用卡	t	2026-06-28 14:47:49.17849+00	2026-06-28 14:47:49.17849+00	{}	\N	\N	\N	t	大戶Plus	大大,大戶
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
-- Data for Name: invitations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.invitations (id, email, token_hash, invited_by, expires_at, accepted_at, created_at) FROM stdin;
\.


--
-- Data for Name: member_cards; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.member_cards (id, user_id, card_product_id, nickname, last_four, is_active, statement_day, payment_due_day, created_at, updated_at, account_tier, card_network_id, opened_on, closed_on) FROM stdin;
\.


--
-- Data for Name: merchant_aliases; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchant_aliases (alias, merchant_id) FROM stdin;
蝦皮	216c8cba-2203-4762-a703-63d791b8fbda
Gemini	03e7e148-cffb-46e2-a8e8-065f0c46cf34
UNIQLO	372a6aca-69b5-4055-8b30-7802a0196ecd
7-ELEVEN	e6c05ee3-e550-4acb-b32f-8e6cf9ae9c08
星巴克	f7e290b0-883b-49ab-8f48-4e15a23fc545
ChatGPT	e54549b5-690f-4c73-b0ae-22be2329a7d2
家樂福	83945f66-fd2c-4e7a-8b94-88284513ac2a
Netflix	1c530c1a-d5a6-46e1-b531-0926da8a2d69
台灣中油	3f41cdc1-b993-43f5-b6db-f90c2b1591e8
Nintendo	f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a
Uber	4064ba61-add5-4acf-a178-3dd75afbcb0c
Steam	b4ba9a4e-f526-48ff-b003-cd2ca2afa982
Agoda	02566e9c-f0e2-4cf5-99c8-a4160505d8df
新光三越	e10b1cbf-fbbf-427c-a660-695e6f69a453
Booking.com	dbec8b68-2c71-4002-b253-c6904dbc2f6d
高鐵	3cb28ab1-652c-451a-afca-3c409f4b738a
PlayStation	f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7
momo	9fcbbe7e-97dc-417f-92ae-599289e39e3e
PChome	75721691-b4e0-4b96-b5a5-924ef4e5697d
淘寶	62befc1b-8a8b-4509-8171-8e31853e7ded
酷澎	65a974d7-07ef-4cfb-8a3e-e705fad0b03c
Disney+	bdb43c61-eff8-4387-b5ed-71bbcbe5982f
CatchPlay	8a3564a4-211e-4107-ac09-6213ffa86819
Spotify	7218a66b-df7c-45c7-9034-d82aab043a59
LINE TV	4ce99021-3364-47c0-a145-2bc034950f01
KKTV	b1863ca8-de0c-4049-b61d-46bfca8cdab7
KKBOX	ce0bf441-6115-4247-ab0b-f45818b3747d
Blizzard	4ca28a85-1639-4479-8b32-05fc30d1d6f6
Garena	9b8f235f-aac9-4680-aee1-da2745001ece
環球影城	cf4e20e9-71a8-4872-b6e3-635d9e656ab0
楓康超市	9aa80628-dbbc-4ca9-996e-2aa0d731842d
喜互惠	bd94a731-c322-41c9-9e68-7444122dba83
美廉社	3b74c59a-1a93-4623-9d6d-4a45e4af48ff
愛買	48430d40-70e7-40e2-ada7-c6302ff4bfb8
大全聯	b50041b5-8961-45e9-9536-3069b04b102b
全家	d8535bc4-2f5d-4dc1-86f7-53d59e658e43
\.


--
-- Data for Name: merchant_categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchant_categories (category_id, merchant_id) FROM stdin;
online	216c8cba-2203-4762-a703-63d791b8fbda
online	03e7e148-cffb-46e2-a8e8-065f0c46cf34
other	372a6aca-69b5-4055-8b30-7802a0196ecd
grocery	e6c05ee3-e550-4acb-b32f-8e6cf9ae9c08
dining	f7e290b0-883b-49ab-8f48-4e15a23fc545
online	e54549b5-690f-4c73-b0ae-22be2329a7d2
grocery	83945f66-fd2c-4e7a-8b94-88284513ac2a
entertainment	1c530c1a-d5a6-46e1-b531-0926da8a2d69
online	1c530c1a-d5a6-46e1-b531-0926da8a2d69
transport	3f41cdc1-b993-43f5-b6db-f90c2b1591e8
entertainment	f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a
online	f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a
transport	4064ba61-add5-4acf-a178-3dd75afbcb0c
entertainment	b4ba9a4e-f526-48ff-b003-cd2ca2afa982
online	b4ba9a4e-f526-48ff-b003-cd2ca2afa982
travel	02566e9c-f0e2-4cf5-99c8-a4160505d8df
online	02566e9c-f0e2-4cf5-99c8-a4160505d8df
other	e10b1cbf-fbbf-427c-a660-695e6f69a453
travel	dbec8b68-2c71-4002-b253-c6904dbc2f6d
online	dbec8b68-2c71-4002-b253-c6904dbc2f6d
travel	3cb28ab1-652c-451a-afca-3c409f4b738a
transport	3cb28ab1-652c-451a-afca-3c409f4b738a
entertainment	f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7
online	f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7
online	9fcbbe7e-97dc-417f-92ae-599289e39e3e
online	75721691-b4e0-4b96-b5a5-924ef4e5697d
online	62befc1b-8a8b-4509-8171-8e31853e7ded
online	65a974d7-07ef-4cfb-8a3e-e705fad0b03c
online	bdb43c61-eff8-4387-b5ed-71bbcbe5982f
entertainment	bdb43c61-eff8-4387-b5ed-71bbcbe5982f
online	8a3564a4-211e-4107-ac09-6213ffa86819
entertainment	8a3564a4-211e-4107-ac09-6213ffa86819
online	7218a66b-df7c-45c7-9034-d82aab043a59
entertainment	7218a66b-df7c-45c7-9034-d82aab043a59
online	4ce99021-3364-47c0-a145-2bc034950f01
entertainment	4ce99021-3364-47c0-a145-2bc034950f01
online	b1863ca8-de0c-4049-b61d-46bfca8cdab7
entertainment	b1863ca8-de0c-4049-b61d-46bfca8cdab7
online	ce0bf441-6115-4247-ab0b-f45818b3747d
entertainment	ce0bf441-6115-4247-ab0b-f45818b3747d
online	4ca28a85-1639-4479-8b32-05fc30d1d6f6
entertainment	4ca28a85-1639-4479-8b32-05fc30d1d6f6
online	9b8f235f-aac9-4680-aee1-da2745001ece
entertainment	9b8f235f-aac9-4680-aee1-da2745001ece
travel	cf4e20e9-71a8-4872-b6e3-635d9e656ab0
entertainment	cf4e20e9-71a8-4872-b6e3-635d9e656ab0
grocery	9aa80628-dbbc-4ca9-996e-2aa0d731842d
grocery	bd94a731-c322-41c9-9e68-7444122dba83
grocery	3b74c59a-1a93-4623-9d6d-4a45e4af48ff
grocery	48430d40-70e7-40e2-ada7-c6302ff4bfb8
grocery	b50041b5-8961-45e9-9536-3069b04b102b
grocery	d8535bc4-2f5d-4dc1-86f7-53d59e658e43
\.


--
-- Data for Name: merchants; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.merchants (name, is_active, is_system, created_at, updated_at, id) FROM stdin;
蝦皮	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	216c8cba-2203-4762-a703-63d791b8fbda
Gemini	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	03e7e148-cffb-46e2-a8e8-065f0c46cf34
UNIQLO	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	372a6aca-69b5-4055-8b30-7802a0196ecd
7-ELEVEN	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	e6c05ee3-e550-4acb-b32f-8e6cf9ae9c08
星巴克	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	f7e290b0-883b-49ab-8f48-4e15a23fc545
ChatGPT	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	e54549b5-690f-4c73-b0ae-22be2329a7d2
家樂福	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	83945f66-fd2c-4e7a-8b94-88284513ac2a
Netflix	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	1c530c1a-d5a6-46e1-b531-0926da8a2d69
台灣中油	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	3f41cdc1-b993-43f5-b6db-f90c2b1591e8
Nintendo	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a
Uber	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	4064ba61-add5-4acf-a178-3dd75afbcb0c
Steam	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	b4ba9a4e-f526-48ff-b003-cd2ca2afa982
Agoda	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	02566e9c-f0e2-4cf5-99c8-a4160505d8df
新光三越	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	e10b1cbf-fbbf-427c-a660-695e6f69a453
Booking.com	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	dbec8b68-2c71-4002-b253-c6904dbc2f6d
高鐵	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	3cb28ab1-652c-451a-afca-3c409f4b738a
PlayStation	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7
momo	t	t	2026-06-28 14:47:49.277339+00	2026-06-28 14:47:49.277339+00	9fcbbe7e-97dc-417f-92ae-599289e39e3e
PChome	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	75721691-b4e0-4b96-b5a5-924ef4e5697d
淘寶	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	62befc1b-8a8b-4509-8171-8e31853e7ded
酷澎	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	65a974d7-07ef-4cfb-8a3e-e705fad0b03c
Disney+	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	bdb43c61-eff8-4387-b5ed-71bbcbe5982f
CatchPlay	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	8a3564a4-211e-4107-ac09-6213ffa86819
Spotify	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	7218a66b-df7c-45c7-9034-d82aab043a59
LINE TV	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	4ce99021-3364-47c0-a145-2bc034950f01
KKTV	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	b1863ca8-de0c-4049-b61d-46bfca8cdab7
KKBOX	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	ce0bf441-6115-4247-ab0b-f45818b3747d
Blizzard	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	4ca28a85-1639-4479-8b32-05fc30d1d6f6
Garena	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	9b8f235f-aac9-4680-aee1-da2745001ece
環球影城	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	cf4e20e9-71a8-4872-b6e3-635d9e656ab0
楓康超市	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	9aa80628-dbbc-4ca9-996e-2aa0d731842d
喜互惠	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	bd94a731-c322-41c9-9e68-7444122dba83
美廉社	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	3b74c59a-1a93-4623-9d6d-4a45e4af48ff
愛買	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	48430d40-70e7-40e2-ada7-c6302ff4bfb8
大全聯	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	b50041b5-8961-45e9-9536-3069b04b102b
全家	t	t	2026-06-28 14:47:49.376993+00	2026-06-28 14:47:49.376993+00	d8535bc4-2f5d-4dc1-86f7-53d59e658e43
\.


--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.password_reset_tokens (id, user_id, token_hash, expires_at, used_at, created_at) FROM stdin;
\.


--
-- Data for Name: payment_methods; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.payment_methods (id, name, is_active, is_system, created_at, updated_at, type, display_order) FROM stdin;
physical_card	實體信用卡	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	physical_card	0
online_card	線上刷卡	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	online_card	0
apple_pay	Apple Pay	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
google_pay	Google Pay	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
samsung_pay	Samsung Pay	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
line_pay	LINE Pay	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
jkopay	街口支付	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
easy_wallet	悠遊付	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
px_pay_plus	全支付	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
fami_pay	全盈+PAY	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
ipass_money	iPASS MONEY	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
icash_pay	icash Pay	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
open_wallet	OPEN錢包	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
esun_wallet	玉山Wallet	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	mobile_payment	0
easycard	悠遊卡功能	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	electronic_ticket	0
icash	icash 功能	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	electronic_ticket	0
ipass	一卡通功能	t	t	2026-06-28 14:47:49.213389+00	2026-06-28 14:47:49.213389+00	electronic_ticket	0
any_payment	不限支付方式	t	t	2026-06-28 14:47:49.305989+00	2026-06-28 14:47:49.305989+00	mobile_payment	0
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
00000000-0000-0000-0000-000000000101	現金	NT$	2	t	2026-06-28 14:47:49.060119+00	prefix	1.000000
00000000-0000-0000-0000-000000000102	點數	點	2	t	2026-06-28 14:47:49.060119+00	suffix	1.000000
00000000-0000-0000-0000-000000000103	里程	哩	2	t	2026-06-28 14:47:49.060119+00	suffix	1.000000
00000000-0000-0000-0000-000000000104	永豐豐點	點	2	f	2026-06-28 14:47:49.177256+00	suffix	1.000000
00000000-0000-0000-0000-000000000105	OPENPOINT	點	2	f	2026-06-28 14:47:49.177256+00	suffix	1.000000
00000000-0000-0000-0000-000000000106	玉山 e point	點	0	f	2026-06-28 14:47:49.234257+00	suffix	1.000000
00000000-0000-0000-0000-000000000107	P幣	P	0	f	2026-06-28 14:47:49.374646+00	suffix	1.000000
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
1	2026-06-28 14:47:49.037366+00
2	2026-06-28 14:47:49.082708+00
3	2026-06-28 14:47:49.099427+00
4	2026-06-28 14:47:49.169778+00
5	2026-06-28 14:47:49.189505+00
6	2026-06-28 14:47:49.202692+00
7	2026-06-28 14:47:49.227624+00
8	2026-06-28 14:47:49.248306+00
9	2026-06-28 14:47:49.264219+00
10	2026-06-28 14:47:49.294337+00
11	2026-06-28 14:47:49.319485+00
12	2026-06-28 14:47:49.338096+00
13	2026-06-28 14:47:49.351493+00
14	2026-06-28 14:47:49.36718+00
15	2026-06-28 14:47:49.390762+00
16	2026-06-28 14:47:49.508374+00
17	2026-06-28 14:47:49.519369+00
18	2026-06-28 14:47:49.533521+00
19	2026-06-28 14:47:49.553369+00
20	2026-06-28 14:47:49.566429+00
21	2026-06-28 14:47:49.618082+00
22	2026-06-28 14:47:49.63531+00
23	2026-06-28 14:47:49.649405+00
24	2026-06-28 14:47:49.671164+00
25	2026-06-28 14:47:49.823549+00
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sessions (id, user_id, token_hash, expires_at, created_at, last_seen_at) FROM stdin;
\.


--
-- Data for Name: telegram_chat_bindings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.telegram_chat_bindings (chat_id, user_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.transactions (id, user_id, card_id, amount_minor, category_id, transaction_date, note, created_at, updated_at, merchant_name, payment_method_id, card_network_id, currency_id, country_id, channel, status, merchant_id) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, email, password_hash, display_name, role, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: cap_versions; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.cap_versions (id, reward_cap_id, cap_type, limit_value, reward_unit_id, period_type, effective_from, effective_to, supersedes_version_id) FROM stdin;
97c15c07-848d-7dda-9820-08e53b4ac61c	1b7d9b3f-984e-65f7-eab5-47badfbb524f	reward_amount	1000.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
c8998db4-5a56-2f8f-6ba8-cc12841700a7	718340b7-c586-ef9a-cfe6-43f1d2929b9c	reward_amount	500.000000	00000000-0000-0000-0000-000000000105	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
649015b0-ad84-547b-f10a-68ff8bce0931	17c11b6d-6973-1056-18e1-009ee55a490e	reward_amount	500.000000	00000000-0000-0000-0000-000000000105	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
eed536d3-433f-f252-8710-1535cc4fc8f9	8c48035f-4129-bf6c-e8ff-90f890966e89	reward_amount	300.000000	00000000-0000-0000-0000-000000000104	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
c35ffb55-f829-68b8-2d13-7005eb8e3f82	5846fd5c-1c7d-d7c8-eaed-42e3bf8ddc9b	reward_amount	150.000000	00000000-0000-0000-0000-000000000101	calendar_month	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N
088d4bc7-7e21-f4f0-054e-7521f284e787	31349150-6f14-360e-eb6c-f0a2d5e87884	reward_amount	100.000000	00000000-0000-0000-0000-000000000101	calendar_month	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N
90897f9b-641a-26ee-27f1-3b375fd6f50f	2de281a4-a08f-5838-b19e-e87996858bf4	reward_amount	5000.000000	00000000-0000-0000-0000-000000000106	calendar_month	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N
86c1cd93-b291-3f3e-b4f0-906ff7c0b07e	bc3a6b80-760d-6ecc-2ef7-1131a8cdaa4c	reward_amount	5000.000000	00000000-0000-0000-0000-000000000106	calendar_month	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N
52f97aa3-720f-88d6-5e92-1c9836e9d470	3ff8199a-b2ae-871b-0aa0-2dcc41b4c08d	reward_amount	5000.000000	00000000-0000-0000-0000-000000000106	calendar_month	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N
873b15d2-cce3-8b55-8816-dc664ddeff4d	36f59619-5345-a378-bb89-0c37ed4a402b	reward_amount	300.000000	00000000-0000-0000-0000-000000000106	calendar_month	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N
a1d15c49-01a6-3ad2-7b14-08bdba5f12d0	e14efacf-aecc-b5a4-4949-52182468cddb	reward_amount	400.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
c4cdfac6-9c46-a07b-ee3b-4b3cd5759eb6	72541f4c-e63f-51a2-898f-e1d6ced36472	reward_amount	1000.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
c8a881b3-d9ad-b303-7b8c-2227bb11e959	7b693482-347e-e322-c813-4b46f9473ab2	reward_amount	100.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
a236ccf4-c955-4e5e-cb94-58da871da7b0	53468fa0-87e0-31db-3507-aaa6eb4e7991	reward_amount	500.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N
3bb84456-6132-2276-c6cf-219b79435d62	255f2b46-a39a-896f-0f2f-bcc7a1386951	reward_amount	500.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N
c2193e0b-20de-3adc-287d-8dedca05b4ee	04021a4b-291c-d330-83ce-1237ac7cfc2b	reward_amount	200.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N
59cf7223-2056-ae3e-ff04-0688774cb157	c588fa54-6b8f-8822-2fa7-9e1c0a1a4741	reward_amount	300.000000	00000000-0000-0000-0000-000000000101	calendar_month	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	\N
\.


--
-- Data for Name: caps; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.caps (id, reward_program_id, name, scope, is_active) FROM stdin;
1b7d9b3f-984e-65f7-eab5-47badfbb524f	61000000-0000-0000-0000-000000000003	Richart 指定方案加碼每月上限	component	t
718340b7-c586-ef9a-cfe6-43f1d2929b9c	61000000-0000-0000-0000-000000000004	海外消費加碼每月上限	component	t
17c11b6d-6973-1056-18e1-009ee55a490e	61000000-0000-0000-0000-000000000004	統一集團加碼每月上限	component	t
8c48035f-4129-bf6c-e8ff-90f890966e89	61000000-0000-0000-0000-000000000002	行動支付加碼每月上限	component	t
5846fd5c-1c7d-d7c8-eaed-42e3bf8ddc9b	61000000-0000-0000-0000-000000000005	網路與指定行動支付最高 3%每月上限	component	t
31349150-6f14-360e-eb6c-f0a2d5e87884	61000000-0000-0000-0000-000000000005	指定數位訂閱平台最高 10%每月上限	component	t
2de281a4-a08f-5838-b19e-e87996858bf4	61000000-0000-0000-0000-000000000006	UP 選指定行動支付最高 4.5%每月上限	component	t
bc3a6b80-760d-6ecc-2ef7-1131a8cdaa4c	61000000-0000-0000-0000-000000000006	UP 選百大指定消費最高 4.5%每月上限	component	t
3ff8199a-b2ae-871b-0aa0-2dcc41b4c08d	61000000-0000-0000-0000-000000000006	UP 選國外實體消費最高 4.5%每月上限	component	t
36f59619-5345-a378-bb89-0c37ed4a402b	61000000-0000-0000-0000-000000000006	國內餐飲領券加碼 3%每月上限	component	t
e14efacf-aecc-b5a4-4949-52182468cddb	61000000-0000-0000-0000-000000000001	大戶指定任務加碼 2.5%每月上限	component	t
72541f4c-e63f-51a2-898f-e1d6ced36472	61000000-0000-0000-0000-000000000001	大戶Plus 指定任務加碼 4%每月上限	component	t
7b693482-347e-e322-c813-4b46f9473ab2	61000000-0000-0000-0000-000000000001	悠遊卡自動加值 3%每月上限	component	t
53468fa0-87e0-31db-3507-aaa6eb4e7991	61000000-0000-0000-0000-000000000001	大戶Plus 悠遊卡自動加值 5%每月上限	component	t
255f2b46-a39a-896f-0f2f-bcc7a1386951	61000000-0000-0000-0000-000000000009	全球環球影城樂園最高 10%每月上限	component	t
04021a4b-291c-d330-83ce-1237ac7cfc2b	61000000-0000-0000-0000-000000000009	生活購物最高 5%每月上限	component	t
c588fa54-6b8f-8822-2fa7-9e1c0a1a4741	61000000-0000-0000-0000-000000000010	指定影音/網購/遊戲加碼 3%每月上限	component	t
\.


--
-- Data for Name: component_caps; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.component_caps (reward_component_id, reward_cap_id, effective_from, effective_to) FROM stdin;
71000000-0000-0000-0000-000000000006	1b7d9b3f-984e-65f7-eab5-47badfbb524f	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000008	718340b7-c586-ef9a-cfe6-43f1d2929b9c	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000009	17c11b6d-6973-1056-18e1-009ee55a490e	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000004	8c48035f-4129-bf6c-e8ff-90f890966e89	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000011	5846fd5c-1c7d-d7c8-eaed-42e3bf8ddc9b	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00
71000000-0000-0000-0000-000000000012	31349150-6f14-360e-eb6c-f0a2d5e87884	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00
71000000-0000-0000-0000-000000000014	2de281a4-a08f-5838-b19e-e87996858bf4	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000015	bc3a6b80-760d-6ecc-2ef7-1131a8cdaa4c	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000016	3ff8199a-b2ae-871b-0aa0-2dcc41b4c08d	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000017	36f59619-5345-a378-bb89-0c37ed4a402b	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000002	e14efacf-aecc-b5a4-4949-52182468cddb	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000021	72541f4c-e63f-51a2-898f-e1d6ced36472	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000022	7b693482-347e-e322-c813-4b46f9473ab2	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000023	53468fa0-87e0-31db-3507-aaa6eb4e7991	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00
71000000-0000-0000-0000-000000000028	255f2b46-a39a-896f-0f2f-bcc7a1386951	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00
71000000-0000-0000-0000-000000000029	04021a4b-291c-d330-83ce-1237ac7cfc2b	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00
71000000-0000-0000-0000-000000000033	c588fa54-6b8f-8822-2fa7-9e1c0a1a4741	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00
\.


--
-- Data for Name: component_versions; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.component_versions (id, reward_component_id, reward_unit_id, name, description, effective_from, effective_to, announced_at, published_at, supersedes_version_id, change_reason, display_change_until, effect_type, reward_value) FROM stdin;
3427a0a5-f739-66f0-e441-9143a7242528	71000000-0000-0000-0000-000000000003	00000000-0000-0000-0000-000000000104	一般消費豐點		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
d08c37b9-98a6-4002-be16-a0a25626d5ec	71000000-0000-0000-0000-000000000005	00000000-0000-0000-0000-000000000101	一般消費回饋		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
03f25742-23b5-2e33-fb9e-cd254e7b5474	71000000-0000-0000-0000-000000000006	00000000-0000-0000-0000-000000000101	Richart 指定方案加碼		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
7dcd4d24-16cc-da52-9700-4b5a7371e011	71000000-0000-0000-0000-000000000007	00000000-0000-0000-0000-000000000105	國內一般消費		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
2c740699-be25-ec56-7a81-1e517757cbcc	71000000-0000-0000-0000-000000000008	00000000-0000-0000-0000-000000000105	海外消費加碼		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
4216367a-253b-32a9-9fd7-d34999900679	71000000-0000-0000-0000-000000000009	00000000-0000-0000-0000-000000000105	統一集團加碼		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
536c98f1-288c-98e6-b55e-bfaf8c0d0b5c	71000000-0000-0000-0000-000000000004	00000000-0000-0000-0000-000000000104	行動支付加碼		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
e88983c4-23f5-911d-ad5f-8b39d6aca8a3	71000000-0000-0000-0000-000000000010	00000000-0000-0000-0000-000000000101	一般消費最高 1%		2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
3bc3432f-13db-eae0-0536-44f16f702bd3	71000000-0000-0000-0000-000000000011	00000000-0000-0000-0000-000000000101	網路與指定行動支付最高 3%		2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
37f9dcaa-8dac-0ae5-3ac9-9f8e8fe26897	71000000-0000-0000-0000-000000000012	00000000-0000-0000-0000-000000000101	指定數位訂閱平台最高 10%		2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
c3725719-15c3-5cff-708f-789fbcb18f48	71000000-0000-0000-0000-000000000013	00000000-0000-0000-0000-000000000106	一般消費最高 1%		2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
2e97bcdb-cf6a-2146-0726-b40b5b6c0f29	71000000-0000-0000-0000-000000000014	00000000-0000-0000-0000-000000000106	UP 選指定行動支付最高 4.5%		2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
c65b9789-b344-8587-add8-822cc8b95acb	71000000-0000-0000-0000-000000000015	00000000-0000-0000-0000-000000000106	UP 選百大指定消費最高 4.5%		2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
6f7f065e-43eb-3ff9-5251-d77204f5b354	71000000-0000-0000-0000-000000000016	00000000-0000-0000-0000-000000000106	UP 選國外實體消費最高 4.5%		2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
9ecec0a7-32bc-8622-6796-fe3b01b881d4	71000000-0000-0000-0000-000000000017	00000000-0000-0000-0000-000000000106	國內餐飲領券加碼 3%		2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
fd3911ca-fa40-b710-902d-97c7c6516d13	71000000-0000-0000-0000-000000000018	00000000-0000-0000-0000-000000000101	國內外一般消費 1%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
f2813805-7e44-36d5-ac00-c88d4f6fdef6	71000000-0000-0000-0000-000000000019	00000000-0000-0000-0000-000000000101	國外實體商店最高 2.5%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.237212+00	\N		\N	ADD_RATE	0.00000000
6cb750a0-ff17-99e1-6f44-58247ec37bd1	71000000-0000-0000-0000-000000000001	00000000-0000-0000-0000-000000000101	國內一般消費 1%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
d82c20e2-0060-4040-121f-9b330e508e54	71000000-0000-0000-0000-000000000002	00000000-0000-0000-0000-000000000101	大戶指定任務加碼 2.5%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
542cd7f4-c3ef-41a7-36c4-6617105bc236	71000000-0000-0000-0000-000000000020	00000000-0000-0000-0000-000000000101	國外一般消費 2%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
7b6431d1-62dd-d273-594f-7e659d5bb61c	71000000-0000-0000-0000-000000000021	00000000-0000-0000-0000-000000000101	大戶Plus 指定任務加碼 4%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
86363273-5615-9f20-8b5e-1b002577c224	71000000-0000-0000-0000-000000000022	00000000-0000-0000-0000-000000000101	悠遊卡自動加值 3%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
c4247a78-62b1-9d28-33dd-1bb574cda02f	71000000-0000-0000-0000-000000000023	00000000-0000-0000-0000-000000000101	大戶Plus 悠遊卡自動加值 5%		2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	\N	2026-06-28 14:47:49.179695+00	\N		\N	ADD_RATE	0.00000000
6dcdcc7f-e67b-866e-0951-a9852563b37a	71000000-0000-0000-0000-000000000024	00000000-0000-0000-0000-000000000107	一般消費 1% P幣		2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
766f1758-f76d-f7d9-a529-9611fcb785c2	71000000-0000-0000-0000-000000000026	00000000-0000-0000-0000-000000000101	國內消費回饋 1.234%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
1cea679e-7d67-ee7c-9ea1-3c97a3354b40	71000000-0000-0000-0000-000000000027	00000000-0000-0000-0000-000000000101	國外消費回饋 2.234%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
0db7d0f9-57a8-7649-55fc-263b19980451	71000000-0000-0000-0000-000000000028	00000000-0000-0000-0000-000000000101	全球環球影城樂園最高 10%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
68fffb76-1194-5e03-4d02-b859c30977b2	71000000-0000-0000-0000-000000000029	00000000-0000-0000-0000-000000000101	生活購物最高 5%		2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
d962a992-4dbf-0b52-23fc-070c4349cee5	71000000-0000-0000-0000-000000000030	00000000-0000-0000-0000-000000000101	國內消費 1% 刷卡金		2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
2498e159-cb1c-19bc-569f-7173c8d29317	71000000-0000-0000-0000-000000000031	00000000-0000-0000-0000-000000000101	國外消費 2.5% 刷卡金		2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
3123de7e-8080-2a39-3e9b-217c09be670b	71000000-0000-0000-0000-000000000032	00000000-0000-0000-0000-000000000101	保險 1% 刷卡金		2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
6aeaa7c8-f1d7-eee9-bbf3-2cae821c2569	71000000-0000-0000-0000-000000000033	00000000-0000-0000-0000-000000000101	指定影音/網購/遊戲加碼 3%		2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	\N	2026-06-28 14:47:49.380581+00	\N		\N	ADD_RATE	0.00000000
\.


--
-- Data for Name: components; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.components (id, reward_program_id, stack_group, stack_policy, priority, is_active, created_at, exclusive_group, layer, display_order) FROM stdin;
71000000-0000-0000-0000-000000000003	61000000-0000-0000-0000-000000000002	base	best_of_group	10	t	2026-06-28 14:47:49.181636+00	base	1	0
71000000-0000-0000-0000-000000000005	61000000-0000-0000-0000-000000000003	base	best_of_group	10	t	2026-06-28 14:47:49.181636+00	base	1	0
71000000-0000-0000-0000-000000000006	61000000-0000-0000-0000-000000000003	richart_plan	best_of_group	20	t	2026-06-28 14:47:49.181636+00	richart_plan	2	0
71000000-0000-0000-0000-000000000007	61000000-0000-0000-0000-000000000004	base	best_of_group	10	t	2026-06-28 14:47:49.181636+00	base	1	0
71000000-0000-0000-0000-000000000008	61000000-0000-0000-0000-000000000004	channel_bonus	best_of_group	20	t	2026-06-28 14:47:49.181636+00	channel_bonus	3	0
71000000-0000-0000-0000-000000000009	61000000-0000-0000-0000-000000000004	channel_bonus	best_of_group	20	t	2026-06-28 14:47:49.181636+00	channel_bonus	3	0
71000000-0000-0000-0000-000000000004	61000000-0000-0000-0000-000000000002	payment_bonus	best_of_group	20	t	2026-06-28 14:47:49.181636+00	payment_bonus	4	0
71000000-0000-0000-0000-000000000010	61000000-0000-0000-0000-000000000005	ubear_core	best_of_group	10	t	2026-06-28 14:47:49.238738+00	ubear_core	2	0
71000000-0000-0000-0000-000000000011	61000000-0000-0000-0000-000000000005	ubear_core	best_of_group	20	t	2026-06-28 14:47:49.238738+00	ubear_core	2	0
71000000-0000-0000-0000-000000000012	61000000-0000-0000-0000-000000000005	ubear_core	best_of_group	30	t	2026-06-28 14:47:49.238738+00	ubear_core	2	0
71000000-0000-0000-0000-000000000013	61000000-0000-0000-0000-000000000006	unicard_core	best_of_group	10	t	2026-06-28 14:47:49.238738+00	unicard_core	2	0
71000000-0000-0000-0000-000000000014	61000000-0000-0000-0000-000000000006	unicard_core	best_of_group	20	t	2026-06-28 14:47:49.238738+00	unicard_core	2	0
71000000-0000-0000-0000-000000000015	61000000-0000-0000-0000-000000000006	unicard_core	best_of_group	20	t	2026-06-28 14:47:49.238738+00	unicard_core	2	0
71000000-0000-0000-0000-000000000016	61000000-0000-0000-0000-000000000006	unicard_core	best_of_group	20	t	2026-06-28 14:47:49.238738+00	unicard_core	2	0
71000000-0000-0000-0000-000000000017	61000000-0000-0000-0000-000000000006	unicard_dining_bonus	best_of_group	30	t	2026-06-28 14:47:49.238738+00	unicard_dining_bonus	2	0
71000000-0000-0000-0000-000000000018	61000000-0000-0000-0000-000000000007	lol_core	best_of_group	10	t	2026-06-28 14:47:49.238738+00	lol_core	2	0
71000000-0000-0000-0000-000000000019	61000000-0000-0000-0000-000000000007	lol_core	best_of_group	20	t	2026-06-28 14:47:49.238738+00	lol_core	2	0
71000000-0000-0000-0000-000000000001	61000000-0000-0000-0000-000000000001	dawho_base	stack	10	t	2026-06-28 14:47:49.181636+00	dawho_base	1	0
71000000-0000-0000-0000-000000000020	61000000-0000-0000-0000-000000000001	dawho_base	stack	10	t	2026-06-28 14:47:49.257578+00	dawho_base	1	0
71000000-0000-0000-0000-000000000002	61000000-0000-0000-0000-000000000001	dawho_tier_bonus	best_of_group	20	t	2026-06-28 14:47:49.181636+00	dawho_tier_bonus	2	0
71000000-0000-0000-0000-000000000021	61000000-0000-0000-0000-000000000001	dawho_tier_bonus	best_of_group	20	t	2026-06-28 14:47:49.257578+00	dawho_tier_bonus	2	0
71000000-0000-0000-0000-000000000024	61000000-0000-0000-0000-000000000008	pi_wallet_base	best_of_group	10	t	2026-06-28 14:47:49.382126+00	pi_wallet_base	2	0
71000000-0000-0000-0000-000000000026	61000000-0000-0000-0000-000000000009	minions_base	best_of_group	10	t	2026-06-28 14:47:49.382126+00	minions_base	2	0
71000000-0000-0000-0000-000000000027	61000000-0000-0000-0000-000000000009	minions_base	best_of_group	20	t	2026-06-28 14:47:49.382126+00	minions_base	2	0
71000000-0000-0000-0000-000000000028	61000000-0000-0000-0000-000000000009	minions_base	best_of_group	30	t	2026-06-28 14:47:49.382126+00	minions_base	2	0
71000000-0000-0000-0000-000000000029	61000000-0000-0000-0000-000000000009	minions_base	best_of_group	40	t	2026-06-28 14:47:49.382126+00	minions_base	2	0
71000000-0000-0000-0000-000000000030	61000000-0000-0000-0000-000000000010	linebank_base	best_of_group	10	t	2026-06-28 14:47:49.382126+00	linebank_base	2	0
71000000-0000-0000-0000-000000000031	61000000-0000-0000-0000-000000000010	linebank_base	best_of_group	20	t	2026-06-28 14:47:49.382126+00	linebank_base	2	0
71000000-0000-0000-0000-000000000032	61000000-0000-0000-0000-000000000010	linebank_base	best_of_group	30	t	2026-06-28 14:47:49.382126+00	linebank_base	2	0
71000000-0000-0000-0000-000000000033	61000000-0000-0000-0000-000000000010	linebank_bonus_online_entertainment	best_of_group	40	t	2026-06-28 14:47:49.382126+00	linebank_bonus_online_entertainment	2	0
71000000-0000-0000-0000-000000000022	61000000-0000-0000-0000-000000000001	exclusive:dawho_easycard_autoload	exclusive	30	t	2026-06-28 14:47:49.359185+00	dawho_easycard_autoload	3	0
71000000-0000-0000-0000-000000000023	61000000-0000-0000-0000-000000000001	exclusive:dawho_easycard_autoload	exclusive	30	t	2026-06-28 14:47:49.359185+00	dawho_easycard_autoload	3	0
\.


--
-- Data for Name: condition_versions; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.condition_versions (id, reward_condition_id, operator, configuration_json, description, effective_from, effective_to, published_at, supersedes_version_id) FROM stdin;
4756f238-bac3-527a-f4b9-c1c2523d1573	a7615036-6e8c-cbd0-53ff-d55e53095ea0	in	{"card_plan_ids": ["21c503d8-00a6-6402-3c54-25e9001a3b95", "654c3131-2c29-2a4e-90cd-4cd7e62e5079", "c94d64f7-fcd8-8e6d-3720-21a4216fc823"]}	需符合：大大 / 大戶 / 大戶Plus	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
98783c3b-daab-c509-97c6-e10c63590233	7789fab1-0a35-c117-0472-e57f30beb1e5	in	{"card_plan_ids": ["21c503d8-00a6-6402-3c54-25e9001a3b95"]}	需符合：大戶	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
4a9662b9-ac34-bacd-81d4-93f5a7d3d333	fce3bb4d-e4e5-8b4b-0ff0-d2bc75089b5b	in	{"card_plan_ids": ["21c503d8-00a6-6402-3c54-25e9001a3b95", "654c3131-2c29-2a4e-90cd-4cd7e62e5079", "c94d64f7-fcd8-8e6d-3720-21a4216fc823"]}	需符合：大大 / 大戶 / 大戶Plus	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
37f656d3-f268-3540-b4b4-58d9327c7bf8	bdf3bb5c-b099-671a-76f8-119815deb65e	in	{"card_plan_ids": ["c94d64f7-fcd8-8e6d-3720-21a4216fc823"]}	需符合：大戶Plus	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
6a581b79-525a-3ade-f007-8569f41bfeda	42923994-811f-b160-9139-d6aca125a786	in	{"card_plan_ids": ["21c503d8-00a6-6402-3c54-25e9001a3b95"]}	需符合：大戶	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
e953dc75-bf52-dfa1-6c52-d7be0fa99d35	1dc72728-6393-b652-359d-0244310569d6	in	{"card_plan_ids": ["c94d64f7-fcd8-8e6d-3720-21a4216fc823"]}	需符合：大戶Plus	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.474471+00	\N
6df8fc3b-8933-5a3f-9f16-5d4362b0f1e0	edd215bc-d7a2-ed92-a940-7c419f9624f4	equals	{"card_plan_ids": ["72b99c99-4ffc-15a8-6481-e32444466701"]}	需於 Richart Life APP 切換至符合消費情境的方案	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.477362+00	\N
ae536446-fb09-4548-6594-2256ac9bcb15	ea04ec78-812b-a4d2-c4c1-7f9501b2e969	equals	{"card_plan_ids": ["ca407472-df08-104b-008a-09212863d809"]}	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.477362+00	\N
e4a4886f-bcfb-3127-7b71-27fb383a403b	d92e3682-e9a4-2501-3669-a06fc1f587b9	equals	{"card_plan_ids": ["a40bb29c-13b8-33be-e7fb-30ef9262d339"]}	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.477362+00	\N
1469b11e-6cef-f7b9-649b-d1783f1a1108	15b010a1-6d96-cb7b-d441-aae287b802b1	equals	{"card_plan_ids": ["ab2f95ab-2d7f-296a-5580-598ba377c568"]}	需於玉山 Wallet 完成任務或以 149 點 e point 訂閱 UP 選	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.477362+00	\N
95892649-a8e1-fb4c-3352-d91e2c62faec	a60644c3-7ef4-c028-0469-bb6d50cb840b	in	{"category_ids": ["dining", "entertainment", "grocery", "online", "sports", "transport", "travel"]}	dining, entertainment, grocery, online, sports, transport, travel	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
60b15ca4-a413-3a88-8154-ee5169739256	cbad1b22-8438-85b9-d666-1e898d4c05b0	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
b46c5fa1-14bd-3494-c756-9e99aa722a1a	9f188b0a-5220-06df-51d9-e124d21527ae	in	{"payment_method_ids": ["apple_pay", "google_pay", "jkopay", "line_pay", "samsung_pay"]}	Apple Pay、Google Pay、LINE Pay、Samsung Pay、街口支付	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
7ec19259-467c-aabc-3551-4876ad33546d	be692c9e-4092-a205-b2ee-ba50390a3ef2	in	{"payment_method_ids": ["easy_wallet", "esun_wallet", "fami_pay", "icash_pay", "jkopay", "line_pay", "online_card", "open_wallet", "px_pay_plus"]}	LINE Pay、OPEN錢包、icash Pay、全支付、全盈+PAY、悠遊付、玉山Wallet、線上刷卡、街口支付	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
9d0a8e61-98df-7ce1-0622-e9f7b4ae9a0b	d25102e2-a7fc-3d54-bd4b-4283057f01d6	equals	{"card_plan_ids": ["12831358-db59-f1bf-0606-9a32cea894b1"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.583843+00	\N
7087f32f-789d-18f3-96a6-cdb66bd6defa	04cbcae2-2e75-fae0-c03e-e6a78d29ed13	equals	{"card_plan_ids": ["edf55cc5-435f-734f-3420-2e4e99f0a67f"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.583843+00	\N
e84ba3e9-94d0-e875-a999-f06baebd14e5	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
82ac03cd-d148-a2f3-3ba9-9e43c86d0ec3	faf18309-64b3-bb7f-25d9-f5a0c0166aa0	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
4462f565-4c6f-59a9-3e94-7d4b2e004753	aaff53c6-f319-870f-c286-0d63dbbc3b92	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
30c9b11f-f986-f3a5-5433-448a199f9b25	4b47029e-307c-6815-9af9-b552aaaad620	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
7dc40f54-be49-18e9-4bc1-7673833cdd6a	b0b05a8c-e916-d622-28d2-80aeac5dcc8b	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
494fb0b6-86e7-613e-c61c-ec6ce3bc1a02	d21b8960-6006-1451-9f20-7babbdcc8ae3	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
331b131d-2e77-5f4f-b9d1-74c6ef42316d	5d75f9d0-b022-10a0-f091-a04457f234c8	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
055ea6d3-bb9f-4f0f-40e4-b2271498b6f3	8c781605-1e7e-c371-b40a-38005fdbe3d8	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
5120124d-e173-9810-a6bc-d7dbf54aeafd	edb13c5c-c699-5fc0-8e7a-36ccfce09c27	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
2c6de3c9-b3ad-f080-bb6f-31f67dab0a99	1fd0480c-2e48-37db-695e-d7f3bd7e80c0	in	{"card_network_ids": ["00000000-0000-0000-0000-000000000201", "00000000-0000-0000-0000-000000000202", "00000000-0000-0000-0000-000000000203"]}	JCB、Mastercard、Visa	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.585697+00	\N
024cda6c-f2e3-7dbd-3703-05a1b5665834	431b9f3d-e3c7-a603-93ca-89319374875b	equals	{"action_required": "registration", "reminder_messages": ["需先完成當期活動登錄"]}	需先完成當期活動登錄	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
5bca65fe-b4b1-4fe4-d4fc-8289b968f0e4	3a6eb9dd-107a-d58a-47f5-6532ac3ac8e2	equals	{"action_required": "account_setup", "reminder_messages": ["需同時綁定帳單 e 化及申辦玉山銀行臺幣帳戶自動扣繳，始享最高回饋"]}	需同時綁定帳單 e 化及申辦玉山銀行臺幣帳戶自動扣繳，始享最高回饋	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
069a6e65-c8f3-f669-92e2-554017ed5c10	8db2d403-2259-e7bb-20b4-4acbdd369cc3	in	{"payment_method_ids": ["easy_wallet", "esun_wallet", "fami_pay", "icash_pay", "ipass_money", "jkopay", "line_pay", "px_pay_plus"]}	LINE Pay、iPASS MONEY、icash Pay、全支付、全盈+PAY、悠遊付、玉山Wallet、街口支付	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
b05666ef-6791-e5d5-a49b-e22ca7b22b06	8a74d914-4e64-0425-7c9d-1fae0cfcb81e	in	{"payment_method_ids": ["apple_pay", "google_pay", "physical_card", "samsung_pay"]}	Apple Pay、Google Pay、Samsung Pay、實體信用卡	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
6ed5b183-9909-c239-2cda-80cdc1793742	19735697-e018-f481-0baa-52d7a6ce7a2c	in	{"payment_method_ids": ["easycard"]}	悠遊卡功能	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
49e06c31-1831-ca21-f19c-b59f910c1293	2f59a0ae-c5ad-6679-1f31-1729eb6cd808	in	{"merchant_ids": ["83945f66-fd2c-4e7a-8b94-88284513ac2a", "f7e290b0-883b-49ab-8f48-4e15a23fc545", "e6c05ee3-e550-4acb-b32f-8e6cf9ae9c08"]}	7-ELEVEN、家樂福、星巴克	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
dd083d61-ad78-9e34-cd9d-69f1472be686	381d06c0-a509-b3fd-c717-64d0e29cf53f	in	{"merchant_ids": ["f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7", "1c530c1a-d5a6-46e1-b531-0926da8a2d69", "03e7e148-cffb-46e2-a8e8-065f0c46cf34", "e54549b5-690f-4c73-b0ae-22be2329a7d2", "b4ba9a4e-f526-48ff-b003-cd2ca2afa982", "f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a"]}	ChatGPT、Gemini、Netflix、Nintendo、PlayStation、Steam	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
636eb317-466e-ccea-e66a-1d35b0c66d7b	a0ac9a8a-01e8-cfd8-994d-fda3b7863b4c	in	{"merchant_ids": ["cf4e20e9-71a8-4872-b6e3-635d9e656ab0"]}	環球影城	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
2d5dfcd3-461f-49f4-8cee-e74067ccafb3	3a5ca9ca-3f67-5fed-8288-f73b0499e8ab	equals	{"action_required": "account_setup", "reminder_messages": ["需綁定帳單 e 化；行動支付仍依玉山系統認列為網路消費"]}	需綁定帳單 e 化；行動支付仍依玉山系統認列為網路消費	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
cc195522-cd8d-d14f-2faa-654d31c49777	0b637d20-d7e9-24c6-ca54-f11bc31a05a9	equals	{"action_required": "account_setup", "reminder_messages": ["需同時申辦帳單 e 化及玉山銀行臺幣帳戶自動扣繳，且成功扣繳卡費"]}	需同時申辦帳單 e 化及玉山銀行臺幣帳戶自動扣繳，且成功扣繳卡費	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
26754ab6-0798-8c06-1b11-b04ecb9233b4	d16e5165-975d-c338-511c-8dd07c278467	equals	{"action_required": "registration", "reminder_messages": ["需先至玉山 Wallet 領取 Unicard 國內餐飲專屬優惠"]}	需先至玉山 Wallet 領取 Unicard 國內餐飲專屬優惠	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
d674e71a-f6f8-091f-f1ad-0511602d0ce4	fc97d473-529d-5c1a-240d-140c3a6af711	equals	{"action_required": "account_setup", "reminder_messages": ["需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
c474f9ba-b67e-355f-329e-f00d45c4cffd	86ece6d0-5af3-4cf2-4cee-d37b1878ca9d	equals	{"action_required": "account_setup", "reminder_messages": ["需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
9ea9ed63-54d2-1696-88bc-4aedfb506522	0dac9159-d165-a1a8-a2e8-3877672844f7	equals	{"action_required": "account_setup", "reminder_messages": ["需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
e80eda54-5c71-ca09-0c60-b097b96df3f1	dc22af68-c262-dea3-a30c-439f32bc28ef	equals	{"action_required": "account_setup", "reminder_messages": ["需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費"]}	需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
8cbf3a7a-a451-023d-7aba-375183b4fda2	84d8ff32-88ec-7754-f0b5-8af5d481d68d	equals	{"action_required": "account_setup", "reminder_messages": ["需註冊 Pi 拍錢包 App、綁定 Pi 卡並申請玉山信用卡帳單 e 化，始享 P幣回饋"]}	需註冊 Pi 拍錢包 App、綁定 Pi 卡並申請玉山信用卡帳單 e 化，始享 P幣回饋	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
4bfcfe6b-e9c7-4e2b-f84b-ea8170868a2c	c963c894-7e6a-d546-43d4-d211659593fc	equals	{"action_required": "account_setup", "reminder_messages": ["活動期間 2026/1/1 至 2026/6/30；須辦理上海商銀帳戶自動扣繳信用卡款，每期帳單回饋上限 NT$200"]}	活動期間 2026/1/1 至 2026/6/30；須辦理上海商銀帳戶自動扣繳信用卡款，每期帳單回饋上限 NT$200	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.592787+00	\N
27cc7b7d-c3b9-c572-7307-d81acbb8cc7d	731fdb69-a065-53e4-5e2d-ce0a9073f071	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
2cdc95c3-4f7a-d7cb-8e6a-410cf267b8ed	e0a08e89-a30b-e4dc-d334-ddab5702721b	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
9162b127-b4fc-1b7a-1a3e-0a14d0c99845	0d197f43-2907-749a-b0ad-1d68a130557b	in	{"category_ids": ["dining", "entertainment", "online", "travel"]}	dining, entertainment, online, travel	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
6ee56c88-827d-3d45-cc79-a955c6d94389	7e60f516-7975-baa1-a816-26bf4258ade4	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
e52ea437-0475-c1f5-069a-e5a88e90849d	ca1d0804-c3de-32ed-1f44-25328d7a1f38	in	{"category_ids": ["overseas"]}	overseas	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
d7685759-c697-6981-023f-64cc364a2fe5	ed6b68b5-e339-b1b8-05e7-5cd213405b50	in	{"category_ids": ["grocery"]}	grocery	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
23b804fa-fe92-c3f3-22f9-8b9ab2b66cd1	5b59ba73-e06e-92f5-5258-fa131f0afd86	in	{"category_ids": ["general"]}	general	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
f4801b47-c0db-ed59-4035-731d665d5d02	04124803-d351-157c-ce45-7ab379016d97	in	{"category_ids": ["general"]}	general	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
d7ca5eb4-6d64-a62c-a40d-ef1fbc6b6cc6	957143f6-7593-b682-3191-5a7028dfed96	in	{"category_ids": ["general"]}	general	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
459fa0f5-5f28-1ce3-68b2-9f190486ee9d	93811ab2-9be4-7fb5-b70f-6f92db9115f7	in	{"category_ids": ["general"]}	general	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
90e1d104-b192-d064-1081-83af193baa15	18e8c81c-20f8-ab68-e616-71459c3c8c68	in	{"category_ids": ["general"]}	general	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
46747ba0-4ef0-781b-47b0-864a1b0d7473	a1fffb4c-f817-b255-bedf-3b34ec51cca7	in	{"category_ids": ["general"]}	general	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
4be04394-cf10-1758-e4c8-d2c323085aaa	bf4fe89a-899d-a3bc-036b-74bc247f91d7	in	{"category_ids": ["overseas"]}	overseas	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
11fdb5df-264b-e793-0206-b4fbbbd485d6	f1daa930-5796-e7b3-fa9e-7139b437862e	in	{"category_ids": ["dining"]}	dining	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
a82b1e7d-a482-9c53-0df8-2865d5d45852	afecd87a-ee9d-98e2-622a-e8de7335cc54	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
29e43f6a-6214-fdca-b8ec-a9d4e6293322	3823e2c7-c0c2-9775-911e-3fac5d6855de	in	{"category_ids": ["overseas"]}	overseas	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
cfa16a49-52bd-a989-1a49-18a61f6cef11	f3b57864-a81c-1c9d-ae0f-909b4db4b9a4	in	{"category_ids": ["overseas"]}	overseas	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
25841c0e-eebf-d430-a786-527f2094134e	dc3d5d55-c1e5-c750-a90b-17cc5f856630	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
38bf5c3b-574c-42a8-a376-41de70a70d7e	a622c75a-f3aa-9308-d866-323fb984ec5c	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
3e165465-7c24-dd58-da87-ddba3dc0bf4c	70cb9412-aa5b-0b66-f616-c014f7d89437	in	{"category_ids": ["general"]}	general	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
50966832-1eec-119b-1116-d055c95ff15c	5e0a87fb-2fff-ba10-c424-265b9ae5cb90	in	{"category_ids": ["general"]}	general	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
2b598d48-073c-7511-d1f3-b51f7492548d	f4488b6f-a21b-4287-bcc6-cb423d91db56	in	{"category_ids": ["overseas"]}	overseas	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
82f2cd81-8be0-ab03-cd57-83db14f9fdb4	b068c978-8dca-cadd-7f1f-5f42df520f52	in	{"category_ids": ["entertainment", "travel"]}	entertainment, travel	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
9d2761d1-83bf-4fa2-a275-9e229368d4c1	a550d70a-34ce-15e0-6221-57d6470ba025	in	{"category_ids": ["grocery"]}	grocery	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
232485ad-adb4-6683-ae65-d8e7df09bb07	a8e36708-5dff-460b-9231-224e076a709c	in	{"category_ids": ["general"]}	general	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
4194d5d8-76b4-40ee-f521-34faf1ff1dea	544ff7a1-849a-6f70-b39b-52c3337f9626	in	{"category_ids": ["overseas"]}	overseas	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
da95452b-cd22-4962-1289-ee54f2b7a362	f5de6d17-8de6-bb97-c3d9-2091e9924155	in	{"category_ids": ["insurance"]}	insurance	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
f461c676-64dd-5342-e828-9593305347af	50fd56de-3f01-29ce-8fed-ed5d45496d8a	in	{"category_ids": ["entertainment", "online"]}	entertainment, online	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
d6917217-93b0-6cb6-e921-0a4756062b7b	4504e7e7-5423-0252-49a5-5e319b079974	in	{"category_ids": ["dining", "entertainment", "grocery", "online", "sports", "transport", "travel"]}	dining, entertainment, grocery, online, sports, transport, travel	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
9030c2d1-8b2c-548a-4c1c-b5cbef03594e	77715b53-3b85-a80f-7183-b03a7d298aac	in	{"category_ids": ["dining", "entertainment", "grocery", "online", "sports", "transport", "travel"]}	dining, entertainment, grocery, online, sports, transport, travel	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.468834+00	\N
b70b997d-d31f-3479-f877-d8d3ae08f358	8ca73250-ce65-b402-627c-391c6acca4ca	in	{"payment_method_ids": ["apple_pay", "google_pay", "physical_card", "samsung_pay"]}	Apple Pay、Google Pay、Samsung Pay、實體信用卡	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
d5fd59a0-a492-89c9-d61b-913ae5748030	4379b5a1-bf22-e4b7-5097-786e9ef55b2b	in	{"payment_method_ids": ["easycard"]}	悠遊卡功能	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.479181+00	\N
19d3d8c2-e7d9-a798-ef20-fa0927f717cc	0447f935-9496-4965-504d-8ce372070525	in	{"merchant_ids": ["9fcbbe7e-97dc-417f-92ae-599289e39e3e", "83945f66-fd2c-4e7a-8b94-88284513ac2a", "3cb28ab1-652c-451a-afca-3c409f4b738a", "e10b1cbf-fbbf-427c-a660-695e6f69a453", "dbec8b68-2c71-4002-b253-c6904dbc2f6d", "4064ba61-add5-4acf-a178-3dd75afbcb0c", "02566e9c-f0e2-4cf5-99c8-a4160505d8df", "216c8cba-2203-4762-a703-63d791b8fbda", "3f41cdc1-b993-43f5-b6db-f90c2b1591e8", "372a6aca-69b5-4055-8b30-7802a0196ecd"]}	Agoda、Booking.com、UNIQLO、Uber、momo、台灣中油、家樂福、新光三越、蝦皮、高鐵	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
3336fc70-c077-552e-3fc4-1e7a9441e823	d3c62910-4403-d530-1fd5-6afedf844edf	in	{"merchant_ids": ["48430d40-70e7-40e2-ada7-c6302ff4bfb8", "d8535bc4-2f5d-4dc1-86f7-53d59e658e43", "9aa80628-dbbc-4ca9-996e-2aa0d731842d", "83945f66-fd2c-4e7a-8b94-88284513ac2a", "e6c05ee3-e550-4acb-b32f-8e6cf9ae9c08", "b50041b5-8961-45e9-9536-3069b04b102b", "bd94a731-c322-41c9-9e68-7444122dba83", "3b74c59a-1a93-4623-9d6d-4a45e4af48ff"]}	7-ELEVEN、全家、喜互惠、大全聯、家樂福、愛買、楓康超市、美廉社	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
06a94a8f-bf14-eb79-4a6e-6d3afeb325cd	4ed90215-b8fd-f56b-3d1c-9e9134dce143	in	{"merchant_ids": ["4ca28a85-1639-4479-8b32-05fc30d1d6f6", "8a3564a4-211e-4107-ac09-6213ffa86819", "65a974d7-07ef-4cfb-8a3e-e705fad0b03c", "bdb43c61-eff8-4387-b5ed-71bbcbe5982f", "9b8f235f-aac9-4680-aee1-da2745001ece", "ce0bf441-6115-4247-ab0b-f45818b3747d", "b1863ca8-de0c-4049-b61d-46bfca8cdab7", "4ce99021-3364-47c0-a145-2bc034950f01", "9fcbbe7e-97dc-417f-92ae-599289e39e3e", "f4b7c102-bbf5-4d1a-a07b-d74649b4c0f7", "1c530c1a-d5a6-46e1-b531-0926da8a2d69", "b4ba9a4e-f526-48ff-b003-cd2ca2afa982", "f384ebfe-61a1-4b9f-9b3f-6ca1fb5d6c2a", "216c8cba-2203-4762-a703-63d791b8fbda", "75721691-b4e0-4b96-b5a5-924ef4e5697d", "7218a66b-df7c-45c7-9034-d82aab043a59", "62befc1b-8a8b-4509-8171-8e31853e7ded"]}	Blizzard、CatchPlay、Disney+、Garena、KKBOX、KKTV、LINE TV、Netflix、Nintendo、PChome、PlayStation、Spotify、Steam、momo、淘寶、蝦皮、酷澎	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	2026-06-28 14:47:49.481354+00	\N
\.


--
-- Data for Name: conditions; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.conditions (id, condition_type, name, is_active) FROM stdin;
cbad1b22-8438-85b9-d666-1e898d4c05b0	category	一般消費豐點／消費分類	t
e0a08e89-a30b-e4dc-d334-ddab5702721b	category	一般消費回饋／消費分類	t
0d197f43-2907-749a-b0ad-1d68a130557b	category	Richart 指定方案加碼／消費分類	t
7e60f516-7975-baa1-a816-26bf4258ade4	category	國內一般消費／消費分類	t
ca1d0804-c3de-32ed-1f44-25328d7a1f38	category	海外消費加碼／消費分類	t
ed6b68b5-e339-b1b8-05e7-5cd213405b50	category	統一集團加碼／消費分類	t
731fdb69-a065-53e4-5e2d-ce0a9073f071	category	行動支付加碼／消費分類	t
5b59ba73-e06e-92f5-5258-fa131f0afd86	category	一般消費最高 1%／消費分類	t
04124803-d351-157c-ce45-7ab379016d97	category	網路與指定行動支付最高 3%／消費分類	t
957143f6-7593-b682-3191-5a7028dfed96	category	指定數位訂閱平台最高 10%／消費分類	t
93811ab2-9be4-7fb5-b70f-6f92db9115f7	category	一般消費最高 1%／消費分類	t
18e8c81c-20f8-ab68-e616-71459c3c8c68	category	UP 選指定行動支付最高 4.5%／消費分類	t
a1fffb4c-f817-b255-bedf-3b34ec51cca7	category	UP 選百大指定消費最高 4.5%／消費分類	t
bf4fe89a-899d-a3bc-036b-74bc247f91d7	category	UP 選國外實體消費最高 4.5%／消費分類	t
f1daa930-5796-e7b3-fa9e-7139b437862e	category	國內餐飲領券加碼 3%／消費分類	t
afecd87a-ee9d-98e2-622a-e8de7335cc54	category	國內外一般消費 1%／消費分類	t
3823e2c7-c0c2-9775-911e-3fac5d6855de	category	國外實體商店最高 2.5%／消費分類	t
a60644c3-7ef4-c028-0469-bb6d50cb840b	category	國內一般消費 1%／消費分類	t
4504e7e7-5423-0252-49a5-5e319b079974	category	大戶指定任務加碼 2.5%／消費分類	t
f3b57864-a81c-1c9d-ae0f-909b4db4b9a4	category	國外一般消費 2%／消費分類	t
77715b53-3b85-a80f-7183-b03a7d298aac	category	大戶Plus 指定任務加碼 4%／消費分類	t
dc3d5d55-c1e5-c750-a90b-17cc5f856630	category	悠遊卡自動加值 3%／消費分類	t
a622c75a-f3aa-9308-d866-323fb984ec5c	category	大戶Plus 悠遊卡自動加值 5%／消費分類	t
70cb9412-aa5b-0b66-f616-c014f7d89437	category	一般消費 1% P幣／消費分類	t
5e0a87fb-2fff-ba10-c424-265b9ae5cb90	category	國內消費回饋 1.234%／消費分類	t
f4488b6f-a21b-4287-bcc6-cb423d91db56	category	國外消費回饋 2.234%／消費分類	t
b068c978-8dca-cadd-7f1f-5f42df520f52	category	全球環球影城樂園最高 10%／消費分類	t
a550d70a-34ce-15e0-6221-57d6470ba025	category	生活購物最高 5%／消費分類	t
a8e36708-5dff-460b-9231-224e076a709c	category	國內消費 1% 刷卡金／消費分類	t
544ff7a1-849a-6f70-b39b-52c3337f9626	category	國外消費 2.5% 刷卡金／消費分類	t
f5de6d17-8de6-bb97-c3d9-2091e9924155	category	保險 1% 刷卡金／消費分類	t
50fd56de-3f01-29ce-8fed-ed5d45496d8a	category	指定影音/網購/遊戲加碼 3%／消費分類	t
bdf3bb5c-b099-671a-76f8-119815deb65e	card_plan	大戶Plus 指定任務加碼 4%／會員資格	t
1dc72728-6393-b652-359d-0244310569d6	card_plan	大戶Plus 悠遊卡自動加值 5%／會員資格	t
edd215bc-d7a2-ed92-a940-7c419f9624f4	card_plan	Richart 指定方案加碼／切換方案	t
ea04ec78-812b-a4d2-c4c1-7f9501b2e969	card_plan	UP 選指定行動支付最高 4.5%／切換方案	t
d92e3682-e9a4-2501-3669-a06fc1f587b9	card_plan	UP 選百大指定消費最高 4.5%／切換方案	t
15b010a1-6d96-cb7b-d441-aae287b802b1	card_plan	UP 選國外實體消費最高 4.5%／切換方案	t
9f188b0a-5220-06df-51d9-e124d21527ae	payment_method	行動支付加碼／支付方式	t
be692c9e-4092-a205-b2ee-ba50390a3ef2	payment_method	網路與指定行動支付最高 3%／支付方式	t
8db2d403-2259-e7bb-20b4-4acbdd369cc3	payment_method	UP 選指定行動支付最高 4.5%／支付方式	t
8a74d914-4e64-0425-7c9d-1fae0cfcb81e	payment_method	UP 選國外實體消費最高 4.5%／支付方式	t
8ca73250-ce65-b402-627c-391c6acca4ca	payment_method	國外實體商店最高 2.5%／支付方式	t
19735697-e018-f481-0baa-52d7a6ce7a2c	payment_method	悠遊卡自動加值 3%／支付方式	t
4379b5a1-bf22-e4b7-5097-786e9ef55b2b	payment_method	大戶Plus 悠遊卡自動加值 5%／支付方式	t
2f59a0ae-c5ad-6679-1f31-1729eb6cd808	merchant	統一集團加碼／指定店家	t
381d06c0-a509-b3fd-c717-64d0e29cf53f	merchant	指定數位訂閱平台最高 10%／指定店家	t
0447f935-9496-4965-504d-8ce372070525	merchant	UP 選百大指定消費最高 4.5%／指定店家	t
a0ac9a8a-01e8-cfd8-994d-fda3b7863b4c	merchant	全球環球影城樂園最高 10%／指定店家	t
d3c62910-4403-d530-1fd5-6afedf844edf	merchant	生活購物最高 5%／指定店家	t
4ed90215-b8fd-f56b-3d1c-9e9134dce143	merchant	指定影音/網購/遊戲加碼 3%／指定店家	t
a7615036-6e8c-cbd0-53ff-d55e53095ea0	card_plan	國內一般消費 1%／會員資格	f
7789fab1-0a35-c117-0472-e57f30beb1e5	card_plan	大戶指定任務加碼 2.5%／會員資格	f
fce3bb4d-e4e5-8b4b-0ff0-d2bc75089b5b	card_plan	國外一般消費 2%／會員資格	f
42923994-811f-b160-9139-d6aca125a786	card_plan	悠遊卡自動加值 3%／會員資格	f
d25102e2-a7fc-3d54-bd4b-4283057f01d6	card_plan	大戶指定任務加碼 2.5%／切換方案	t
04cbcae2-2e75-fae0-c03e-e6a78d29ed13	card_plan	悠遊卡自動加值 3%／切換方案	t
faf18309-64b3-bb7f-25d9-f5a0c0166aa0	card_network	2026 上半年核心權益／卡組織	t
aaff53c6-f319-870f-c286-0d63dbbc3b92	card_network	2026 上半年核心權益／卡組織	t
4b47029e-307c-6815-9af9-b552aaaad620	card_network	2026 上半年核心權益／卡組織	t
b0b05a8c-e916-d622-28d2-80aeac5dcc8b	card_network	2026 U Bear 核心權益／卡組織	t
d21b8960-6006-1451-9f20-7babbdcc8ae3	card_network	2025-2026 Unicard 核心權益／卡組織	t
5d75f9d0-b022-10a0-f091-a04457f234c8	card_network	2026 英雄聯盟卡既有卡友權益／卡組織	t
f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	card_network	2026 上半年核心權益／卡組織	t
8c781605-1e7e-c371-b40a-38005fdbe3d8	card_network	2026 Pi 拍錢包信用卡回饋／卡組織	t
edb13c5c-c699-5fc0-8e7a-36ccfce09c27	card_network	2026 小小兵回饋卡優惠／卡組織	t
1fd0480c-2e48-37db-695e-d7f3bd7e80c0	card_network	2025-2026 LINE Bank聯名卡回饋／卡組織	t
431b9f3d-e3c7-a603-93ca-89319374875b	channel	行動支付加碼／操作提醒	t
3a6eb9dd-107a-d58a-47f5-6532ac3ac8e2	channel	一般消費最高 1%／操作提醒	t
3a5ca9ca-3f67-5fed-8288-f73b0499e8ab	channel	網路與指定行動支付最高 3%／操作提醒	t
0b637d20-d7e9-24c6-ca54-f11bc31a05a9	channel	一般消費最高 1%／操作提醒	t
d16e5165-975d-c338-511c-8dd07c278467	channel	國內餐飲領券加碼 3%／操作提醒	t
fc97d473-529d-5c1a-240d-140c3a6af711	channel	大戶指定任務加碼 2.5%／操作提醒	t
86ece6d0-5af3-4cf2-4cee-d37b1878ca9d	channel	大戶Plus 指定任務加碼 4%／操作提醒	t
0dac9159-d165-a1a8-a2e8-3877672844f7	channel	悠遊卡自動加值 3%／操作提醒	t
dc22af68-c262-dea3-a30c-439f32bc28ef	channel	大戶Plus 悠遊卡自動加值 5%／操作提醒	t
84d8ff32-88ec-7754-f0b5-8af5d481d68d	channel	一般消費 1% P幣／操作提醒	t
c963c894-7e6a-d546-43d4-d211659593fc	channel	生活購物最高 5%／操作提醒	t
\.


--
-- Data for Name: migration_reports; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.migration_reports (id, migration_version, severity, subject_type, subject_id, message, details_json, created_at) FROM stdin;
\.


--
-- Data for Name: programs; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.programs (id, card_product_id, name, source_url, verified_at, status, created_at, updated_at) FROM stdin;
61000000-0000-0000-0000-000000000002	51000000-0000-0000-0000-000000000002	2026 上半年核心權益	https://bank.sinopac.com/sinopacBT/webevents/2008_SportCard2/index.html?Branch=SP2026H1	2026-06-14 00:00:00+00	published	2026-06-28 14:47:49.179695+00	2026-06-28 14:47:49.179695+00
61000000-0000-0000-0000-000000000003	51000000-0000-0000-0000-000000000003	2026 上半年核心權益	https://www.taishinbank.com.tw/	2026-06-14 00:00:00+00	published	2026-06-28 14:47:49.179695+00	2026-06-28 14:47:49.179695+00
61000000-0000-0000-0000-000000000004	51000000-0000-0000-0000-000000000004	2026 上半年核心權益	https://www.ctbcbank.com/content/twrbo/zh_tw/cc_index/cc_feedback/cc_feedback_other_index/cc_feedback_uniopen.html	2026-06-14 00:00:00+00	published	2026-06-28 14:47:49.179695+00	2026-06-28 14:47:49.179695+00
61000000-0000-0000-0000-000000000005	51000000-0000-0000-0000-000000000005	2026 U Bear 核心權益	https://www.esunbank.com/zh-tw/personal/credit-card/intro/bank-card/u-bear	2026-06-15 00:00:00+00	published	2026-06-28 14:47:49.237212+00	2026-06-28 14:47:49.237212+00
61000000-0000-0000-0000-000000000006	51000000-0000-0000-0000-000000000006	2025-2026 Unicard 核心權益	https://www.esunbank.com/zh-tw/personal/credit-card/intro/bank-card/unicard	2026-06-15 00:00:00+00	published	2026-06-28 14:47:49.237212+00	2026-06-28 14:47:49.237212+00
61000000-0000-0000-0000-000000000007	51000000-0000-0000-0000-000000000007	2026 英雄聯盟卡既有卡友權益	https://www.ctbcbank.com/content/twrbo/zh_tw/cc_index/cc_product/cc_introduction_index/B_LOL.html	2026-06-15 00:00:00+00	published	2026-06-28 14:47:49.237212+00	2026-06-28 14:47:49.237212+00
61000000-0000-0000-0000-000000000001	51000000-0000-0000-0000-000000000001	2026 上半年核心權益	https://dawho.tw/about/card/	2026-06-16 00:00:00+00	published	2026-06-28 14:47:49.179695+00	2026-06-28 14:47:49.179695+00
61000000-0000-0000-0000-000000000008	51000000-0000-0000-0000-000000000008	2026 Pi 拍錢包信用卡回饋	https://www.piapp.com.tw/picard/	2026-06-16 00:00:00+00	published	2026-06-28 14:47:49.380581+00	2026-06-28 14:47:49.380581+00
61000000-0000-0000-0000-000000000009	51000000-0000-0000-0000-000000000009	2026 小小兵回饋卡優惠	https://www.scsb.com.tw/content/card/card_1110222.html	2026-06-16 00:00:00+00	published	2026-06-28 14:47:49.380581+00	2026-06-28 14:47:49.380581+00
61000000-0000-0000-0000-000000000010	51000000-0000-0000-0000-000000000010	2025-2026 LINE Bank聯名卡回饋	https://activity.ubot.com.tw/LINEBankCard/index.htm	2026-06-16 00:00:00+00	published	2026-06-28 14:47:49.380581+00	2026-06-28 14:47:49.380581+00
\.


--
-- Data for Name: requirements; Type: TABLE DATA; Schema: reward; Owner: -
--

COPY reward.requirements (id, reward_component_id, reward_condition_id, effective_from, effective_to, display_order) FROM stdin;
0e8976c6-02e8-36be-f619-67a93889fd15	71000000-0000-0000-0000-000000000003	cbad1b22-8438-85b9-d666-1e898d4c05b0	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
3ec86d29-93cf-bc46-b13d-93502026857e	71000000-0000-0000-0000-000000000005	e0a08e89-a30b-e4dc-d334-ddab5702721b	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
a2fc98a4-9a81-d4c9-32c5-f9289f8d8df4	71000000-0000-0000-0000-000000000006	0d197f43-2907-749a-b0ad-1d68a130557b	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
7f88b01a-816c-584d-8bfc-40511eacfcce	71000000-0000-0000-0000-000000000007	7e60f516-7975-baa1-a816-26bf4258ade4	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
4ccc23d7-f491-01c0-ae0b-c0c9bd6b2c12	71000000-0000-0000-0000-000000000008	ca1d0804-c3de-32ed-1f44-25328d7a1f38	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
2384e91e-aff2-ab00-9c6f-f46750ed3915	71000000-0000-0000-0000-000000000009	ed6b68b5-e339-b1b8-05e7-5cd213405b50	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
a390385d-9c8e-9032-5d8a-c40a38fbbb19	71000000-0000-0000-0000-000000000004	731fdb69-a065-53e4-5e2d-ce0a9073f071	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
e230ddf8-6aa0-0d4d-6a64-ac88b48dea0a	71000000-0000-0000-0000-000000000010	5b59ba73-e06e-92f5-5258-fa131f0afd86	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	10
c329bc20-6150-ee10-9c09-602e31f1b976	71000000-0000-0000-0000-000000000011	04124803-d351-157c-ce45-7ab379016d97	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	10
bc7987e2-ee4e-ac10-42af-8239fcf00a90	71000000-0000-0000-0000-000000000012	957143f6-7593-b682-3191-5a7028dfed96	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	10
5fec7a79-4a51-7e4a-401d-03617d01670e	71000000-0000-0000-0000-000000000013	93811ab2-9be4-7fb5-b70f-6f92db9115f7	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	10
cb958fa2-8f5d-e25b-b1d8-b23e489a7039	71000000-0000-0000-0000-000000000014	18e8c81c-20f8-ab68-e616-71459c3c8c68	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	10
499212bf-4878-d4dc-dd0a-f2289247421c	71000000-0000-0000-0000-000000000015	a1fffb4c-f817-b255-bedf-3b34ec51cca7	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	10
6e41b932-a5c6-d183-0031-4a99ceb38aa3	71000000-0000-0000-0000-000000000016	bf4fe89a-899d-a3bc-036b-74bc247f91d7	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	10
75b08196-d4b6-bff7-03dc-6bc5a4ad28a2	71000000-0000-0000-0000-000000000017	f1daa930-5796-e7b3-fa9e-7139b437862e	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	10
80a5f82b-1d97-1c82-154d-3386a1073657	71000000-0000-0000-0000-000000000018	afecd87a-ee9d-98e2-622a-e8de7335cc54	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
3ee7a609-6423-f417-9ca4-c768aa60646e	71000000-0000-0000-0000-000000000019	3823e2c7-c0c2-9775-911e-3fac5d6855de	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
2fcf77b5-05c8-081b-c7fc-e3c49e526e98	71000000-0000-0000-0000-000000000001	a60644c3-7ef4-c028-0469-bb6d50cb840b	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
96442cf9-8ebf-b548-bdab-63098363b7fa	71000000-0000-0000-0000-000000000002	4504e7e7-5423-0252-49a5-5e319b079974	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
6cafe70b-fcf2-c816-0c84-44e2c8f7512a	71000000-0000-0000-0000-000000000020	f3b57864-a81c-1c9d-ae0f-909b4db4b9a4	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
f0e7ea83-370d-8c32-fd95-3cb72d3eaec6	71000000-0000-0000-0000-000000000021	77715b53-3b85-a80f-7183-b03a7d298aac	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
78ed5acd-2bc5-bdef-b772-7761d3343437	71000000-0000-0000-0000-000000000022	dc3d5d55-c1e5-c750-a90b-17cc5f856630	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
525e12b3-17ac-36a3-1f31-0752256130f7	71000000-0000-0000-0000-000000000023	a622c75a-f3aa-9308-d866-323fb984ec5c	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	10
c1a185e9-71b4-cb1b-ad39-f7a21a74d428	71000000-0000-0000-0000-000000000024	70cb9412-aa5b-0b66-f616-c014f7d89437	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	10
b919332d-80ba-29f9-2546-a0e2aba5b430	71000000-0000-0000-0000-000000000026	5e0a87fb-2fff-ba10-c424-265b9ae5cb90	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
0c7d94df-74ec-4d36-222f-a601deec2f4a	71000000-0000-0000-0000-000000000027	f4488b6f-a21b-4287-bcc6-cb423d91db56	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
8137239f-1309-7ca1-26b7-9a72bfc20d93	71000000-0000-0000-0000-000000000028	b068c978-8dca-cadd-7f1f-5f42df520f52	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
dea0e219-399a-c791-5e55-ab6541ce1b11	71000000-0000-0000-0000-000000000029	a550d70a-34ce-15e0-6221-57d6470ba025	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	10
c58150b1-6ffa-caaa-5c57-a7b81733fc04	71000000-0000-0000-0000-000000000030	a8e36708-5dff-460b-9231-224e076a709c	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	10
4e67bc95-8cc3-f824-0088-8882c804ec63	71000000-0000-0000-0000-000000000031	544ff7a1-849a-6f70-b39b-52c3337f9626	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	10
9d87af97-255f-7d27-03ec-2799f82faa26	71000000-0000-0000-0000-000000000032	f5de6d17-8de6-bb97-c3d9-2091e9924155	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	10
2bac05f7-06fe-bfc2-c581-6f88de4ad62c	71000000-0000-0000-0000-000000000033	50fd56de-3f01-29ce-8fed-ed5d45496d8a	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	10
d96d2cb9-1b88-f3f2-7d73-dc26daa282cf	71000000-0000-0000-0000-000000000021	bdf3bb5c-b099-671a-76f8-119815deb65e	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	30
ac83cd94-9401-977b-3b3f-3e27619abf75	71000000-0000-0000-0000-000000000023	1dc72728-6393-b652-359d-0244310569d6	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	30
c0dd199b-d56b-c8ef-fe65-2da7f8b174c8	71000000-0000-0000-0000-000000000006	edd215bc-d7a2-ed92-a940-7c419f9624f4	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	40
858a73e2-0eac-a08c-07a2-01f7de2dbcaa	71000000-0000-0000-0000-000000000014	ea04ec78-812b-a4d2-c4c1-7f9501b2e969	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	40
2fe3b6d1-5230-dce5-5bb2-17aa753dc491	71000000-0000-0000-0000-000000000015	d92e3682-e9a4-2501-3669-a06fc1f587b9	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	40
2f18969d-4441-6c8e-d5bd-56f1c151489d	71000000-0000-0000-0000-000000000016	15b010a1-6d96-cb7b-d441-aae287b802b1	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	40
86c2968f-de9a-8639-50ae-58dd3b7ae806	71000000-0000-0000-0000-000000000004	9f188b0a-5220-06df-51d9-e124d21527ae	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	20
50c840c0-266a-8c95-3c1b-4a25ebaa21f2	71000000-0000-0000-0000-000000000011	be692c9e-4092-a205-b2ee-ba50390a3ef2	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	20
7bcb0f95-9c9a-584d-47c2-7a52bc7e77b7	71000000-0000-0000-0000-000000000014	8db2d403-2259-e7bb-20b4-4acbdd369cc3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	20
f49bc10e-67f1-3dd9-1820-28937a111dc1	71000000-0000-0000-0000-000000000016	8a74d914-4e64-0425-7c9d-1fae0cfcb81e	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	20
992c3fc8-b5e8-39fc-3aa8-04ab19585d71	71000000-0000-0000-0000-000000000019	8ca73250-ce65-b402-627c-391c6acca4ca	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	20
79d155b5-9704-b95e-7d90-443d0704e4e7	71000000-0000-0000-0000-000000000022	19735697-e018-f481-0baa-52d7a6ce7a2c	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	20
238cb930-7bf1-af09-1927-09dd4d60f8d0	71000000-0000-0000-0000-000000000023	4379b5a1-bf22-e4b7-5097-786e9ef55b2b	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	20
8d3510f4-f94e-75bb-b3df-4b0945b86e94	71000000-0000-0000-0000-000000000009	2f59a0ae-c5ad-6679-1f31-1729eb6cd808	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	25
df92612a-f977-2f70-f19f-a3ee5ee8a3c7	71000000-0000-0000-0000-000000000012	381d06c0-a509-b3fd-c717-64d0e29cf53f	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	25
91d32fe5-891b-0ef6-d179-8ff1db3c1304	71000000-0000-0000-0000-000000000015	0447f935-9496-4965-504d-8ce372070525	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	25
eba527d5-dc4c-0fac-a49c-056a1081c41a	71000000-0000-0000-0000-000000000028	a0ac9a8a-01e8-cfd8-994d-fda3b7863b4c	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	25
9adb9af8-f194-5b59-5176-5f7b88145347	71000000-0000-0000-0000-000000000029	d3c62910-4403-d530-1fd5-6afedf844edf	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	25
e2c36a9a-c1ec-874c-ea68-e0f4a4ad8c2c	71000000-0000-0000-0000-000000000033	4ed90215-b8fd-f56b-3d1c-9e9134dce143	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	25
1ff8401a-e51c-8630-d892-4c06b71c21c1	71000000-0000-0000-0000-000000000002	d25102e2-a7fc-3d54-bd4b-4283057f01d6	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	40
cb687442-5410-b52c-aaa1-9aae118a598e	71000000-0000-0000-0000-000000000022	04cbcae2-2e75-fae0-c03e-e6a78d29ed13	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	40
89be4fe9-e17a-79fe-869b-2597a4b0420d	71000000-0000-0000-0000-000000000003	faf18309-64b3-bb7f-25d9-f5a0c0166aa0	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
1c37b239-bc04-5599-b4a2-caced944620c	71000000-0000-0000-0000-000000000005	aaff53c6-f319-870f-c286-0d63dbbc3b92	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
72940b1a-a02c-8dd6-8c37-90129213943e	71000000-0000-0000-0000-000000000006	aaff53c6-f319-870f-c286-0d63dbbc3b92	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
6db1bd43-fe52-34f4-ee50-ca43c414aa7c	71000000-0000-0000-0000-000000000007	4b47029e-307c-6815-9af9-b552aaaad620	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
4c5ff666-ebb1-6671-67e9-bc9f03ba9790	71000000-0000-0000-0000-000000000008	4b47029e-307c-6815-9af9-b552aaaad620	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
151c5058-902b-43d3-2d49-c4b5d1b6d20f	71000000-0000-0000-0000-000000000009	4b47029e-307c-6815-9af9-b552aaaad620	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
d4e8272f-6313-c0a1-1489-a9a8be56d195	71000000-0000-0000-0000-000000000004	faf18309-64b3-bb7f-25d9-f5a0c0166aa0	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
7e40e2fa-3453-062d-806a-6e84d3c31fe3	71000000-0000-0000-0000-000000000010	b0b05a8c-e916-d622-28d2-80aeac5dcc8b	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	15
e8931695-4c3a-bff1-bb8b-1ae2ae67809f	71000000-0000-0000-0000-000000000011	b0b05a8c-e916-d622-28d2-80aeac5dcc8b	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	15
e4708299-ca71-a401-4651-5967016d40d0	71000000-0000-0000-0000-000000000012	b0b05a8c-e916-d622-28d2-80aeac5dcc8b	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	15
103d67cb-c6a1-fe0e-44c7-0b599de436c1	71000000-0000-0000-0000-000000000013	d21b8960-6006-1451-9f20-7babbdcc8ae3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	15
91ffda76-1c85-4904-0a3b-11fc9be34566	71000000-0000-0000-0000-000000000014	d21b8960-6006-1451-9f20-7babbdcc8ae3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	15
83c459db-4f44-8443-3016-075adf97f241	71000000-0000-0000-0000-000000000015	d21b8960-6006-1451-9f20-7babbdcc8ae3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	15
a0021c12-a2c1-725e-1114-2697e8e85c9c	71000000-0000-0000-0000-000000000016	d21b8960-6006-1451-9f20-7babbdcc8ae3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	15
0f700df6-15b5-9b20-a38a-18e353f970b1	71000000-0000-0000-0000-000000000017	d21b8960-6006-1451-9f20-7babbdcc8ae3	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	15
8d2d84da-6de6-5e6a-3fd6-52eb3809881a	71000000-0000-0000-0000-000000000018	5d75f9d0-b022-10a0-f091-a04457f234c8	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
4d2ab4bf-ebac-5a5c-407f-7e3ab4f8cd03	71000000-0000-0000-0000-000000000019	5d75f9d0-b022-10a0-f091-a04457f234c8	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
1e08f1fa-cf2b-4298-87b3-f3f36c60e0e3	71000000-0000-0000-0000-000000000001	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
da11a5da-7992-c0e8-041b-7dcb7dcd38e3	71000000-0000-0000-0000-000000000002	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
3111bca6-0545-f855-7274-b233e84d71cb	71000000-0000-0000-0000-000000000020	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
ac8e7472-b210-de21-1fab-7d238176e2cf	71000000-0000-0000-0000-000000000021	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
a7ad035b-aaf1-3f30-46a9-bccd067033a8	71000000-0000-0000-0000-000000000022	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
fd19fb6d-8b9c-70b2-0cc9-86e2a8583a2a	71000000-0000-0000-0000-000000000023	f631fc56-08f9-d6c9-1490-bc3b0ff91f8d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	15
f39cbab7-d0e5-d08e-ea0c-5a9caa389d28	71000000-0000-0000-0000-000000000024	8c781605-1e7e-c371-b40a-38005fdbe3d8	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	15
7440e9a4-e5fe-6c15-2967-1ce1829cef2d	71000000-0000-0000-0000-000000000026	edb13c5c-c699-5fc0-8e7a-36ccfce09c27	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
e09a521c-a7c6-a976-cac7-a221bda9d261	71000000-0000-0000-0000-000000000027	edb13c5c-c699-5fc0-8e7a-36ccfce09c27	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
ea050e2b-3b0d-1776-3253-9802d10c053b	71000000-0000-0000-0000-000000000028	edb13c5c-c699-5fc0-8e7a-36ccfce09c27	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
8c75e9ac-190d-ffde-6f82-e08c0425ab57	71000000-0000-0000-0000-000000000029	edb13c5c-c699-5fc0-8e7a-36ccfce09c27	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	15
52b5ffcb-cb1c-015a-0a45-62d719ce6112	71000000-0000-0000-0000-000000000030	1fd0480c-2e48-37db-695e-d7f3bd7e80c0	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	15
c11e1121-d7c8-3f0d-b35f-a05879f9dfdf	71000000-0000-0000-0000-000000000031	1fd0480c-2e48-37db-695e-d7f3bd7e80c0	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	15
17a1293f-a628-dc3b-a8b5-9fce35e73d23	71000000-0000-0000-0000-000000000032	1fd0480c-2e48-37db-695e-d7f3bd7e80c0	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	15
fe094653-5b6d-56b7-bffa-f24b793f1da9	71000000-0000-0000-0000-000000000033	1fd0480c-2e48-37db-695e-d7f3bd7e80c0	2025-07-31 16:00:00+00	2026-07-31 16:00:00+00	15
1d16f22c-4038-f2c1-0242-812bb88247e7	71000000-0000-0000-0000-000000000004	431b9f3d-e3c7-a603-93ca-89319374875b	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	50
d3a3d129-c147-0676-0b8c-c61352580791	71000000-0000-0000-0000-000000000010	3a6eb9dd-107a-d58a-47f5-6532ac3ac8e2	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	50
eaf68484-bfa0-fd42-be95-a8221c7ec67b	71000000-0000-0000-0000-000000000011	3a5ca9ca-3f67-5fed-8288-f73b0499e8ab	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	50
0f8cf5a7-efbc-151c-c716-6abf9849fd3e	71000000-0000-0000-0000-000000000013	0b637d20-d7e9-24c6-ca54-f11bc31a05a9	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	50
cb1e88da-0019-5222-6a31-b01be18cf0e1	71000000-0000-0000-0000-000000000017	d16e5165-975d-c338-511c-8dd07c278467	2025-09-30 16:00:00+00	2026-06-30 16:00:00+00	50
54439f1a-1732-1b05-8443-a2f07c2efd09	71000000-0000-0000-0000-000000000002	fc97d473-529d-5c1a-240d-140c3a6af711	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	50
09def7b4-ae54-07bb-5296-5097e6754f0e	71000000-0000-0000-0000-000000000021	86ece6d0-5af3-4cf2-4cee-d37b1878ca9d	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	50
9c7828b9-01e0-9884-131f-876ede74a7b3	71000000-0000-0000-0000-000000000022	0dac9159-d165-a1a8-a2e8-3877672844f7	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	50
c0a980f2-624e-13f6-56ac-2d65287fd592	71000000-0000-0000-0000-000000000023	dc22af68-c262-dea3-a30c-439f32bc28ef	2025-12-31 16:00:00+00	2026-06-30 16:00:00+00	50
567f0d4e-a3b0-3318-4ee3-7322d358dc9f	71000000-0000-0000-0000-000000000024	84d8ff32-88ec-7754-f0b5-8af5d481d68d	2026-02-28 16:00:00+00	2026-08-31 16:00:00+00	50
a6d1a0ce-a48c-ab8c-335a-2e332cd45160	71000000-0000-0000-0000-000000000029	c963c894-7e6a-d546-43d4-d211659593fc	2025-12-31 16:00:00+00	2026-12-31 16:00:00+00	50
\.


--
-- Data for Name: reward_calculation_components; Type: TABLE DATA; Schema: transaction; Owner: -
--

COPY transaction.reward_calculation_components (id, reward_calculation_id, reward_component_id, reward_component_version_id, reward_unit_id, suggested_card_plan_id, qualified_card_plan_id, member_qualification_status_id, uncapped_reward, allocated_reward, cap_used_before, cap_remaining_before, condition_snapshot_json, reminder_snapshot_json, effect_type, reward_value, reward_rate) FROM stdin;
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
-- Name: card_product_networks card_product_networks_pkey; Type: CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_product_networks
    ADD CONSTRAINT card_product_networks_pkey PRIMARY KEY (card_product_id, card_network_id);


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
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: integration; Owner: -
--

ALTER TABLE ONLY integration.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);


--
-- Name: reward_usage_adjustments reward_usage_adjustments_pkey; Type: CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.reward_usage_adjustments
    ADD CONSTRAINT reward_usage_adjustments_pkey PRIMARY KEY (id);


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
-- Name: card_products card_products_bank_id_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_products
    ADD CONSTRAINT card_products_bank_id_name_key UNIQUE (bank_id, name);


--
-- Name: card_products card_products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_products
    ADD CONSTRAINT card_products_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_token_hash_key UNIQUE (token_hash);


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
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);


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
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_token_hash_key UNIQUE (token_hash);


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
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: cap_versions cap_versions_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.cap_versions
    ADD CONSTRAINT cap_versions_pkey PRIMARY KEY (id);


--
-- Name: cap_versions cap_versions_reward_cap_id_tstzrange_excl; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.cap_versions
    ADD CONSTRAINT cap_versions_reward_cap_id_tstzrange_excl EXCLUDE USING gist (reward_cap_id WITH =, tstzrange(effective_from, effective_to, '[)'::text) WITH &&);


--
-- Name: caps caps_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.caps
    ADD CONSTRAINT caps_pkey PRIMARY KEY (id);


--
-- Name: component_caps component_caps_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_caps
    ADD CONSTRAINT component_caps_pkey PRIMARY KEY (reward_component_id, reward_cap_id, effective_from);


--
-- Name: component_versions component_versions_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_versions
    ADD CONSTRAINT component_versions_pkey PRIMARY KEY (id);


--
-- Name: component_versions component_versions_reward_component_id_tstzrange_excl; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_versions
    ADD CONSTRAINT component_versions_reward_component_id_tstzrange_excl EXCLUDE USING gist (reward_component_id WITH =, tstzrange(effective_from, effective_to, '[)'::text) WITH &&);


--
-- Name: components components_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.components
    ADD CONSTRAINT components_pkey PRIMARY KEY (id);


--
-- Name: condition_versions condition_versions_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.condition_versions
    ADD CONSTRAINT condition_versions_pkey PRIMARY KEY (id);


--
-- Name: condition_versions condition_versions_reward_condition_id_tstzrange_excl; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.condition_versions
    ADD CONSTRAINT condition_versions_reward_condition_id_tstzrange_excl EXCLUDE USING gist (reward_condition_id WITH =, tstzrange(effective_from, effective_to, '[)'::text) WITH &&);


--
-- Name: conditions conditions_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.conditions
    ADD CONSTRAINT conditions_pkey PRIMARY KEY (id);


--
-- Name: migration_reports migration_reports_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.migration_reports
    ADD CONSTRAINT migration_reports_pkey PRIMARY KEY (id);


--
-- Name: programs programs_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.programs
    ADD CONSTRAINT programs_pkey PRIMARY KEY (id);


--
-- Name: requirements requirements_pkey; Type: CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.requirements
    ADD CONSTRAINT requirements_pkey PRIMARY KEY (id);


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
-- Name: outbox_unpublished_idx; Type: INDEX; Schema: integration; Owner: -
--

CREATE INDEX outbox_unpublished_idx ON integration.outbox_events USING btree (occurred_at) WHERE (published_at IS NULL);


--
-- Name: card_products_bank_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX card_products_bank_active_idx ON public.card_products USING btree (bank_id, is_active);


--
-- Name: invitations_token_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX invitations_token_hash_idx ON public.invitations USING btree (token_hash);


--
-- Name: member_cards_user_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX member_cards_user_active_idx ON public.member_cards USING btree (user_id, is_active);


--
-- Name: member_cards_user_product_network_uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX member_cards_user_product_network_uidx ON public.member_cards USING btree (user_id, card_product_id, card_network_id) NULLS NOT DISTINCT;


--
-- Name: reward_allocations_benefit_transaction_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reward_allocations_benefit_transaction_idx ON public.reward_allocations USING btree (benefit_id, transaction_id);


--
-- Name: sessions_token_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_token_hash_idx ON public.sessions USING btree (token_hash);


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
-- Name: reward_calculation_active_uidx; Type: INDEX; Schema: transaction; Owner: -
--

CREATE UNIQUE INDEX reward_calculation_active_uidx ON transaction.reward_calculations USING btree (transaction_id) WHERE (status = 'active'::text);


--
-- Name: member_cards member_card_network_check; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER member_card_network_check BEFORE INSERT OR UPDATE OF card_product_id, card_network_id ON public.member_cards FOR EACH ROW EXECUTE FUNCTION catalog.validate_member_card_network();


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
    ADD CONSTRAINT card_plans_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES public.card_products(id) ON DELETE CASCADE;


--
-- Name: card_product_networks card_product_networks_card_network_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_product_networks
    ADD CONSTRAINT card_product_networks_card_network_id_fkey FOREIGN KEY (card_network_id) REFERENCES catalog.card_networks(id);


--
-- Name: card_product_networks card_product_networks_card_product_id_fkey; Type: FK CONSTRAINT; Schema: catalog; Owner: -
--

ALTER TABLE ONLY catalog.card_product_networks
    ADD CONSTRAINT card_product_networks_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES public.card_products(id) ON DELETE CASCADE;


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
-- Name: reward_usage_adjustments reward_usage_adjustments_member_card_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.reward_usage_adjustments
    ADD CONSTRAINT reward_usage_adjustments_member_card_id_fkey FOREIGN KEY (member_card_id) REFERENCES public.member_cards(id) ON DELETE CASCADE;


--
-- Name: reward_usage_adjustments reward_usage_adjustments_reward_cap_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.reward_usage_adjustments
    ADD CONSTRAINT reward_usage_adjustments_reward_cap_id_fkey FOREIGN KEY (reward_cap_id) REFERENCES reward.caps(id);


--
-- Name: reward_usage_adjustments reward_usage_adjustments_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.reward_usage_adjustments
    ADD CONSTRAINT reward_usage_adjustments_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: reward_usage_adjustments reward_usage_adjustments_user_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.reward_usage_adjustments
    ADD CONSTRAINT reward_usage_adjustments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_payment_methods user_payment_methods_payment_method_code_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_payment_method_code_fkey FOREIGN KEY (payment_method_id) REFERENCES public.payment_methods(id) ON DELETE CASCADE;


--
-- Name: user_payment_methods user_payment_methods_user_id_fkey; Type: FK CONSTRAINT; Schema: member_profile; Owner: -
--

ALTER TABLE ONLY member_profile.user_payment_methods
    ADD CONSTRAINT user_payment_methods_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: card_products card_products_bank_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.card_products
    ADD CONSTRAINT card_products_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES public.banks(id);


--
-- Name: invitations invitations_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id);


--
-- Name: member_cards member_cards_card_network_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_card_network_id_fkey FOREIGN KEY (card_network_id) REFERENCES catalog.card_networks(id);


--
-- Name: member_cards member_cards_card_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES public.card_products(id);


--
-- Name: member_cards member_cards_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_cards
    ADD CONSTRAINT member_cards_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


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
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reward_allocations reward_allocations_component_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reward_allocations
    ADD CONSTRAINT reward_allocations_component_fkey FOREIGN KEY (benefit_id) REFERENCES reward.components(id);


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
    ADD CONSTRAINT reward_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: telegram_chat_bindings telegram_chat_bindings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.telegram_chat_bindings
    ADD CONSTRAINT telegram_chat_bindings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_card_network_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_card_network_id_fkey FOREIGN KEY (card_network_id) REFERENCES catalog.card_networks(id);


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
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: cap_versions cap_versions_reward_cap_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.cap_versions
    ADD CONSTRAINT cap_versions_reward_cap_id_fkey FOREIGN KEY (reward_cap_id) REFERENCES reward.caps(id) ON DELETE CASCADE;


--
-- Name: cap_versions cap_versions_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.cap_versions
    ADD CONSTRAINT cap_versions_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: cap_versions cap_versions_supersedes_version_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.cap_versions
    ADD CONSTRAINT cap_versions_supersedes_version_id_fkey FOREIGN KEY (supersedes_version_id) REFERENCES reward.cap_versions(id);


--
-- Name: caps caps_reward_program_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.caps
    ADD CONSTRAINT caps_reward_program_id_fkey FOREIGN KEY (reward_program_id) REFERENCES reward.programs(id) ON DELETE CASCADE;


--
-- Name: component_caps component_caps_reward_cap_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_caps
    ADD CONSTRAINT component_caps_reward_cap_id_fkey FOREIGN KEY (reward_cap_id) REFERENCES reward.caps(id) ON DELETE CASCADE;


--
-- Name: component_caps component_caps_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_caps
    ADD CONSTRAINT component_caps_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.components(id) ON DELETE CASCADE;


--
-- Name: component_versions component_versions_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_versions
    ADD CONSTRAINT component_versions_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.components(id) ON DELETE CASCADE;


--
-- Name: component_versions component_versions_reward_unit_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_versions
    ADD CONSTRAINT component_versions_reward_unit_id_fkey FOREIGN KEY (reward_unit_id) REFERENCES public.reward_units(id);


--
-- Name: component_versions component_versions_supersedes_version_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.component_versions
    ADD CONSTRAINT component_versions_supersedes_version_id_fkey FOREIGN KEY (supersedes_version_id) REFERENCES reward.component_versions(id);


--
-- Name: components components_reward_program_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.components
    ADD CONSTRAINT components_reward_program_id_fkey FOREIGN KEY (reward_program_id) REFERENCES reward.programs(id) ON DELETE CASCADE;


--
-- Name: condition_versions condition_versions_reward_condition_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.condition_versions
    ADD CONSTRAINT condition_versions_reward_condition_id_fkey FOREIGN KEY (reward_condition_id) REFERENCES reward.conditions(id) ON DELETE CASCADE;


--
-- Name: condition_versions condition_versions_supersedes_version_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.condition_versions
    ADD CONSTRAINT condition_versions_supersedes_version_id_fkey FOREIGN KEY (supersedes_version_id) REFERENCES reward.condition_versions(id);


--
-- Name: programs programs_card_product_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.programs
    ADD CONSTRAINT programs_card_product_id_fkey FOREIGN KEY (card_product_id) REFERENCES public.card_products(id);


--
-- Name: requirements requirements_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.requirements
    ADD CONSTRAINT requirements_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.components(id) ON DELETE CASCADE;


--
-- Name: requirements requirements_reward_condition_id_fkey; Type: FK CONSTRAINT; Schema: reward; Owner: -
--

ALTER TABLE ONLY reward.requirements
    ADD CONSTRAINT requirements_reward_condition_id_fkey FOREIGN KEY (reward_condition_id) REFERENCES reward.conditions(id) ON DELETE CASCADE;


--
-- Name: reward_calculation_components reward_calculation_components_member_qualification_status__fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_member_qualification_status__fkey FOREIGN KEY (member_qualification_status_id) REFERENCES catalog.member_card_qualification_statuses(id);


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
-- Name: reward_calculation_components reward_calculation_components_reward_component_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_reward_component_id_fkey FOREIGN KEY (reward_component_id) REFERENCES reward.components(id);


--
-- Name: reward_calculation_components reward_calculation_components_reward_component_version_id_fkey; Type: FK CONSTRAINT; Schema: transaction; Owner: -
--

ALTER TABLE ONLY transaction.reward_calculation_components
    ADD CONSTRAINT reward_calculation_components_reward_component_version_id_fkey FOREIGN KEY (reward_component_version_id) REFERENCES reward.component_versions(id);


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

\unrestrict Z6rz96Vi92MZ1qenYMdDbKgBgB6sJDzNicWVkLRStc6wo8F3yUMffvLVIOCboMg

