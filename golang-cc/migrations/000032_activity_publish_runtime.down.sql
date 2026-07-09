DROP TABLE IF EXISTS reward.published_rule_benefits;
DROP TABLE IF EXISTS reward.published_rule_requirements;
DROP TABLE IF EXISTS reward.published_reward_rules;
DROP TABLE IF EXISTS reward.published_activities;

ALTER TABLE reward.activities
    DROP COLUMN IF EXISTS published_checksum,
    DROP COLUMN IF EXISTS published_by,
    DROP COLUMN IF EXISTS published_at;
