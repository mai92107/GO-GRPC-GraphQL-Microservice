DELETE FROM "transaction".reward_calculation_components;
DELETE FROM "transaction".reward_calculations;
DELETE FROM public.reward_allocations;

ALTER TABLE public.reward_allocations
    DROP CONSTRAINT IF EXISTS reward_allocations_component_fkey,
    DROP CONSTRAINT IF EXISTS reward_allocations_benefit_fkey,
    ADD CONSTRAINT reward_allocations_published_rule_fkey
        FOREIGN KEY (benefit_id) REFERENCES reward.published_reward_rules(id);

ALTER TABLE "transaction".reward_calculation_components
    DROP CONSTRAINT IF EXISTS reward_calculation_components_reward_component_version_id_fkey,
    DROP CONSTRAINT IF EXISTS reward_calculation_components_reward_component_id_fkey,
    DROP COLUMN IF EXISTS reward_component_version_id,
    ADD CONSTRAINT reward_calculation_components_published_rule_fkey
        FOREIGN KEY (reward_component_id) REFERENCES reward.published_reward_rules(id);

DROP TABLE IF EXISTS member_profile.reward_usage_adjustments;

DROP TABLE IF EXISTS reward.component_caps;
DROP TABLE IF EXISTS reward.cap_versions;
DROP TABLE IF EXISTS reward.caps;
DROP TABLE IF EXISTS reward.requirements;
DROP TABLE IF EXISTS reward.condition_versions;
DROP TABLE IF EXISTS reward.conditions;
DROP TABLE IF EXISTS reward.component_versions;
DROP TABLE IF EXISTS reward.components;
DROP TABLE IF EXISTS reward.programs;
