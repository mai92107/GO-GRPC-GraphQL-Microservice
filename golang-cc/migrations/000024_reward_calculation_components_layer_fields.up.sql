ALTER TABLE "transaction".reward_calculation_components
    ADD COLUMN IF NOT EXISTS effect_type TEXT NOT NULL DEFAULT 'ADD_RATE',
    ADD COLUMN IF NOT EXISTS reward_value NUMERIC(18,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reward_rate NUMERIC(18,8) NOT NULL DEFAULT 0;

UPDATE "transaction".reward_calculation_components
SET reward_value = rate,
    reward_rate = rate
WHERE reward_rate = 0
  AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'transaction'
        AND table_name = 'reward_calculation_components'
        AND column_name = 'rate'
  );

ALTER TABLE "transaction".reward_calculation_components
    DROP COLUMN IF EXISTS rate;
