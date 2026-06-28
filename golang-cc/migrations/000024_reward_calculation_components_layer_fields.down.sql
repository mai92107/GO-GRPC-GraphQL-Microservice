ALTER TABLE "transaction".reward_calculation_components
    ADD COLUMN IF NOT EXISTS rate NUMERIC(18,8) NOT NULL DEFAULT 0;

UPDATE "transaction".reward_calculation_components
SET rate = reward_rate
WHERE rate = 0;

ALTER TABLE "transaction".reward_calculation_components
    DROP COLUMN IF EXISTS effect_type,
    DROP COLUMN IF EXISTS reward_value,
    DROP COLUMN IF EXISTS reward_rate;
