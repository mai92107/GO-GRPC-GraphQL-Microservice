ALTER TABLE reward.component_versions
    DROP COLUMN IF EXISTS reward_value,
    DROP COLUMN IF EXISTS effect_type;

ALTER TABLE reward.components
    DROP COLUMN IF EXISTS display_order,
    DROP COLUMN IF EXISTS layer;
