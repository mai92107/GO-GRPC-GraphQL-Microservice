CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'member')),
    status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE invitations (
    id UUID PRIMARY KEY,
    email CITEXT NOT NULL,
    token_hash BYTEA UNIQUE NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    issuer TEXT NOT NULL,
    last_four CHAR(4),
    color TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name),
    UNIQUE (id, user_id),
    CHECK (last_four IS NULL OR last_four ~ '^[0-9]{4}$')
);

CREATE TABLE categories (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO categories (code, name) VALUES
    ('general', '一般消費'),
    ('dining', '餐飲'),
    ('online', '網購'),
    ('transport', '交通');

CREATE TABLE reward_units (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    symbol TEXT NOT NULL,
    precision SMALLINT NOT NULL DEFAULT 2 CHECK (precision BETWEEN 0 AND 6),
    is_system BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO reward_units (id, code, name, symbol, precision) VALUES
    ('00000000-0000-0000-0000-000000000101', 'cash_twd', '現金', 'NT$', 2),
    ('00000000-0000-0000-0000-000000000102', 'points', '點數', '點', 2),
    ('00000000-0000-0000-0000-000000000103', 'miles', '里程', '哩', 2);

CREATE TABLE reward_preferences (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    weight NUMERIC(18,6) NOT NULL CHECK (weight >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, reward_unit_id)
);

CREATE TABLE reward_rules (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID NOT NULL,
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    name TEXT NOT NULL,
    rate NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    monthly_cap NUMERIC(18,6) CHECK (monthly_cap > 0),
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (card_id, user_id) REFERENCES cards(id, user_id) ON DELETE CASCADE,
    UNIQUE (id, user_id),
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);

CREATE TABLE reward_rule_categories (
    reward_rule_id UUID NOT NULL REFERENCES reward_rules(id) ON DELETE CASCADE,
    category_code TEXT NOT NULL REFERENCES categories(code),
    PRIMARY KEY (reward_rule_id, category_code)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    category_code TEXT NOT NULL REFERENCES categories(code) CHECK (category_code <> 'general'),
    transaction_date DATE NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (card_id, user_id) REFERENCES cards(id, user_id),
    UNIQUE (id, user_id)
);

CREATE TABLE reward_allocations (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    reward_rule_id UUID NOT NULL REFERENCES reward_rules(id),
    reward_unit_id UUID NOT NULL REFERENCES reward_units(id),
    uncapped_reward NUMERIC(18,6) NOT NULL,
    allocated_reward NUMERIC(18,6) NOT NULL,
    preference_weight NUMERIC(18,6) NOT NULL,
    score NUMERIC(18,6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (transaction_id, reward_rule_id)
);

CREATE INDEX sessions_token_hash_idx ON sessions(token_hash);
CREATE INDEX cards_user_active_idx ON cards(user_id, is_active);
CREATE INDEX reward_rules_user_card_active_idx ON reward_rules(user_id, card_id, is_active);
CREATE INDEX reward_rules_dates_idx ON reward_rules(start_date, end_date);
CREATE INDEX transactions_user_date_idx ON transactions(user_id, transaction_date);
CREATE INDEX transactions_user_card_date_idx ON transactions(user_id, card_id, transaction_date);
CREATE INDEX reward_allocations_rule_transaction_idx ON reward_allocations(reward_rule_id, transaction_id);
CREATE INDEX invitations_token_hash_idx ON invitations(token_hash);
