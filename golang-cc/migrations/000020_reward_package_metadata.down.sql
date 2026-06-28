ALTER TABLE reward.components
    DROP COLUMN IF EXISTS exclusive_group,
    DROP COLUMN IF EXISTS package_level;
