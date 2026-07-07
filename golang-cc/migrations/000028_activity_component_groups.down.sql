ALTER TABLE reward.activity_components
ADD COLUMN reward_group_id UUID;

UPDATE reward.activity_components c
SET reward_group_id = link.reward_group_id
FROM (
    SELECT DISTINCT ON (reward_component_id) reward_component_id, reward_group_id
    FROM reward.activity_component_groups
    ORDER BY reward_component_id, reward_group_id
) link
WHERE link.reward_component_id = c.id;

DELETE FROM reward.activity_components
WHERE reward_group_id IS NULL;

ALTER TABLE reward.activity_components
ALTER COLUMN reward_group_id SET NOT NULL,
ADD CONSTRAINT activity_components_reward_group_id_fkey
    FOREIGN KEY (reward_group_id) REFERENCES reward.activity_groups(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS reward.reward_activity_component_groups_group_idx;
DROP INDEX IF EXISTS reward_activity_component_groups_group_idx;
DROP INDEX IF EXISTS reward.reward_activity_components_order_idx;
DROP INDEX IF EXISTS reward_activity_components_order_idx;
DROP TABLE IF EXISTS reward.activity_component_groups;

CREATE INDEX reward_activity_components_group_order_idx
    ON reward.activity_components(reward_group_id, layer, priority);
