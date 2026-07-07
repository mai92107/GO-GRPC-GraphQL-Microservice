ALTER TABLE reward.cap_versions
    DROP CONSTRAINT IF EXISTS cap_versions_limit_source_check;

ALTER TABLE reward.cap_versions
    ALTER COLUMN limit_value SET NOT NULL;

ALTER TABLE public.member_cards
    DROP CONSTRAINT IF EXISTS member_cards_credit_limit_check;

ALTER TABLE reward.activity_benefits
    DROP COLUMN IF EXISTS cap_formula;

ALTER TABLE reward.cap_versions
    DROP COLUMN IF EXISTS limit_formula;

ALTER TABLE public.member_cards
    DROP COLUMN IF EXISTS credit_limit;
