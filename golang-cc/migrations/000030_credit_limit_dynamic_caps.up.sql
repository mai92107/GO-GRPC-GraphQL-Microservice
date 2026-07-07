ALTER TABLE public.member_cards
    ADD COLUMN credit_limit NUMERIC(18,2);

ALTER TABLE public.member_cards
    ADD CONSTRAINT member_cards_credit_limit_check
    CHECK (credit_limit IS NULL OR credit_limit > 0);

ALTER TABLE reward.cap_versions
    ADD COLUMN limit_formula TEXT;

ALTER TABLE reward.cap_versions
    ALTER COLUMN limit_value DROP NOT NULL;

ALTER TABLE reward.activity_benefits
    ADD COLUMN cap_formula TEXT;

ALTER TABLE reward.cap_versions
    ADD CONSTRAINT cap_versions_limit_source_check
    CHECK (
        limit_value IS NOT NULL
        OR NULLIF(btrim(limit_formula), '') IS NOT NULL
    );
