CREATE TABLE reward.activity_component_groups (
    reward_component_id UUID NOT NULL REFERENCES reward.activity_components(id) ON DELETE CASCADE,
    reward_group_id UUID NOT NULL REFERENCES reward.activity_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (reward_component_id, reward_group_id)
);

INSERT INTO reward.activity_component_groups(reward_component_id, reward_group_id)
SELECT id, reward_group_id
FROM reward.activity_components;

DROP INDEX IF EXISTS reward.reward_activity_components_group_order_idx;
DROP INDEX IF EXISTS reward_activity_components_group_order_idx;

ALTER TABLE reward.activity_components
DROP COLUMN reward_group_id;

CREATE INDEX reward_activity_component_groups_group_idx
    ON reward.activity_component_groups(reward_group_id, reward_component_id);
CREATE INDEX reward_activity_components_order_idx
    ON reward.activity_components(layer, priority, name);
