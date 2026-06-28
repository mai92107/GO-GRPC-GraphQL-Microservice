ALTER TABLE reward.components
    ADD COLUMN IF NOT EXISTS layer INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

ALTER TABLE reward.component_versions
    ADD COLUMN IF NOT EXISTS effect_type TEXT NOT NULL DEFAULT 'ADD_RATE'
        CHECK (effect_type IN ('ADD_RATE', 'SET_RATE', 'MULTIPLY_RATE', 'ADD_CASH', 'DISCOUNT')),
    ADD COLUMN IF NOT EXISTS reward_value NUMERIC(18,8) NOT NULL DEFAULT 0;

UPDATE reward.component_versions
SET reward_value = rate
WHERE reward_value IS NULL;

UPDATE reward.components
SET layer = CASE
    WHEN package_level = 'base' THEN 1
    WHEN package_level = 'bonus' THEN 2
    WHEN package_level = 'channel_bonus' AND lower(stack_group) LIKE '%payment%' THEN 4
    WHEN package_level = 'channel_bonus' THEN 3
    ELSE 1
END
WHERE layer = 1;
