CREATE TABLE public.member_card_credit_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_card_id UUID NOT NULL REFERENCES public.member_cards(id) ON DELETE CASCADE,
    credit_limit NUMERIC(18,2) NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT member_card_credit_limits_credit_limit_check CHECK (credit_limit > 0),
    CONSTRAINT member_card_credit_limits_effective_range_check CHECK (
        effective_to IS NULL OR effective_to > effective_from
    )
);

CREATE INDEX member_card_credit_limits_lookup_idx
    ON public.member_card_credit_limits(member_card_id, effective_from DESC);

CREATE UNIQUE INDEX member_card_credit_limits_current_uidx
    ON public.member_card_credit_limits(member_card_id)
    WHERE effective_to IS NULL;

INSERT INTO public.member_card_credit_limits(member_card_id, credit_limit, effective_from)
SELECT
    mc.id,
    mc.credit_limit,
    COALESCE(mc.opened_on::timestamptz, '2000-01-01 00:00:00+00'::timestamptz)
FROM public.member_cards mc
WHERE mc.credit_limit IS NOT NULL
ON CONFLICT DO NOTHING;
