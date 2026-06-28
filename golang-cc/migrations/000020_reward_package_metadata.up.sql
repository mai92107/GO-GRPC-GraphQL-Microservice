ALTER TABLE reward.components
    ADD COLUMN IF NOT EXISTS package_level TEXT NOT NULL DEFAULT 'base'
        CHECK (package_level IN ('base', 'bonus', 'channel_bonus')),
    ADD COLUMN IF NOT EXISTS exclusive_group TEXT NOT NULL DEFAULT '';

UPDATE reward.components
SET package_level = CASE
    WHEN stack_group = '' OR lower(stack_group) = 'base' OR lower(stack_group) LIKE '%general%' THEN 'base'
    WHEN lower(stack_group) LIKE '%channel%' OR lower(stack_group) LIKE '%merchant%' OR lower(stack_group) LIKE '%payment%' THEN 'channel_bonus'
    ELSE 'bonus'
END
WHERE package_level = 'base';

UPDATE reward.components
SET exclusive_group = CASE
    WHEN stack_group LIKE 'exclusive:%' THEN substring(stack_group FROM 11)
    ELSE stack_group
END
WHERE exclusive_group = '';
