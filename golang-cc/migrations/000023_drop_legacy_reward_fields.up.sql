ALTER TABLE reward.component_versions
    DROP COLUMN IF EXISTS rate_kind,
    DROP COLUMN IF EXISTS rate;

ALTER TABLE reward.components
    DROP COLUMN IF EXISTS package_level;
