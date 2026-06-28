ALTER TABLE reward.components
    ADD COLUMN IF NOT EXISTS package_level TEXT NOT NULL DEFAULT 'base'
        CHECK (package_level IN ('base', 'bonus', 'channel_bonus'));

ALTER TABLE reward.component_versions
    ADD COLUMN IF NOT EXISTS rate_kind TEXT NOT NULL DEFAULT 'percentage'
        CHECK (rate_kind IN ('percentage','fixed_amount','points_per_amount')),
    ADD COLUMN IF NOT EXISTS rate NUMERIC(18,8) NOT NULL DEFAULT 0;

UPDATE reward.component_versions
SET rate = reward_value
WHERE rate = 0;
