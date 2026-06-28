CREATE TABLE IF NOT EXISTS reward.migration_reports (
    id UUID PRIMARY KEY,
    migration_version TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info','warning','error')),
    subject_type TEXT NOT NULL,
    subject_id UUID,
    message TEXT NOT NULL,
    details_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO reward.migration_reports(id,migration_version,severity,subject_type,subject_id,message,details_json)
SELECT md5('000021:component-without-version:'||c.id::text)::uuid,'000021','error','reward_component',c.id,
       'Reward component has no effective version and was left inactive',
       jsonb_build_object('component_code',c.code)
FROM reward.components c
WHERE NOT EXISTS (SELECT 1 FROM reward.component_versions v WHERE v.reward_component_id=c.id)
ON CONFLICT DO NOTHING;

UPDATE reward.components c SET is_active=false
WHERE NOT EXISTS (SELECT 1 FROM reward.component_versions v WHERE v.reward_component_id=c.id);

UPDATE card_products
SET qualified_type='大戶Plus',
    selectable_type='大大,大戶',
    account_tiers='{}'
WHERE id='51000000-0000-0000-0000-000000000001';

UPDATE reward.components
SET stack_policy='stack', package_level='base', exclusive_group='dawho_base'
WHERE id IN (
    '71000000-0000-0000-0000-000000000001',
    '71000000-0000-0000-0000-000000000020'
);

UPDATE reward.components
SET stack_policy='best_of_group', package_level='bonus', exclusive_group='dawho_tier_bonus'
WHERE id IN (
    '71000000-0000-0000-0000-000000000002',
    '71000000-0000-0000-0000-000000000021'
);

UPDATE reward.components
SET stack_policy='exclusive', package_level='channel_bonus', exclusive_group='dawho_easycard_autoload'
WHERE id IN (
    '71000000-0000-0000-0000-000000000022',
    '71000000-0000-0000-0000-000000000023'
);

UPDATE reward.condition_versions
SET configuration_json=jsonb_build_object('category_codes',jsonb_build_array('dining','entertainment','grocery','online','sports','transport','travel')),
    description='dining, entertainment, grocery, online, sports, transport, travel'
WHERE reward_condition_id IN (
    md5('category-condition:71000000-0000-0000-0000-000000000002')::uuid,
    md5('category-condition:71000000-0000-0000-0000-000000000021')::uuid
);

DELETE FROM reward.requirements
WHERE id IN (
    md5('qualified-requirement:71000000-0000-0000-0000-000000000001')::uuid,
    md5('qualified-requirement:71000000-0000-0000-0000-000000000020')::uuid,
    md5('qualified-requirement:71000000-0000-0000-0000-000000000002')::uuid,
    md5('qualified-requirement:71000000-0000-0000-0000-000000000022')::uuid
);

UPDATE reward.conditions
SET is_active=false
WHERE id IN (
    md5('qualified-condition:71000000-0000-0000-0000-000000000001')::uuid,
    md5('qualified-condition:71000000-0000-0000-0000-000000000020')::uuid,
    md5('qualified-condition:71000000-0000-0000-0000-000000000002')::uuid,
    md5('qualified-condition:71000000-0000-0000-0000-000000000022')::uuid
);

INSERT INTO catalog.card_plans(id,card_product_id,code,plan_type,is_active,display_order)
VALUES
    (md5('selectable-plan:71000000-0000-0000-0000-000000000002')::uuid,'51000000-0000-0000-0000-000000000001','selectable_dawho_large','selectable',true,20),
    (md5('selectable-plan:71000000-0000-0000-0000-000000000022')::uuid,'51000000-0000-0000-0000-000000000001','selectable_dawho_large_easycard','selectable',true,30)
ON CONFLICT(id) DO UPDATE SET plan_type=EXCLUDED.plan_type,is_active=true,display_order=EXCLUDED.display_order;

INSERT INTO catalog.card_plan_versions(id,card_plan_id,name,description,reminder_text,effective_from,effective_to,published_at)
VALUES
    (md5('selectable-plan-version:71000000-0000-0000-0000-000000000002')::uuid,md5('selectable-plan:71000000-0000-0000-0000-000000000002')::uuid,'大戶','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單','2026-01-01 00:00:00+08','2026-07-01 00:00:00+08',now()),
    (md5('selectable-plan-version:71000000-0000-0000-0000-000000000022')::uuid,md5('selectable-plan:71000000-0000-0000-0000-000000000022')::uuid,'大戶','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費','需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費','2026-01-01 00:00:00+08','2026-07-01 00:00:00+08',now())
ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,reminder_text=EXCLUDED.reminder_text,effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.conditions(id,code,condition_type,name,is_active)
VALUES
    (md5('selectable-condition:71000000-0000-0000-0000-000000000002')::uuid,'selectable_71000000-0000-0000-0000-000000000002','card_plan','大戶指定任務加碼 2.5%／切換方案',true),
    (md5('selectable-condition:71000000-0000-0000-0000-000000000022')::uuid,'selectable_71000000-0000-0000-0000-000000000022','card_plan','悠遊卡自動加值 3%／切換方案',true)
ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true;

INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
VALUES
    (md5('selectable-condition-version:71000000-0000-0000-0000-000000000002')::uuid,md5('selectable-condition:71000000-0000-0000-0000-000000000002')::uuid,'equals',jsonb_build_object('card_plan_ids',jsonb_build_array(md5('selectable-plan:71000000-0000-0000-0000-000000000002')::uuid)),'需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單','2026-01-01 00:00:00+08','2026-07-01 00:00:00+08'),
    (md5('selectable-condition-version:71000000-0000-0000-0000-000000000022')::uuid,md5('selectable-condition:71000000-0000-0000-0000-000000000022')::uuid,'equals',jsonb_build_object('card_plan_ids',jsonb_build_array(md5('selectable-plan:71000000-0000-0000-0000-000000000022')::uuid)),'需完成 DAWHO 數位帳戶扣繳信用卡款，並使用電子或行動帳單；限前月帳單符合悠遊卡自動加值消費','2026-01-01 00:00:00+08','2026-07-01 00:00:00+08')
ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
VALUES
    (md5('selectable-requirement:71000000-0000-0000-0000-000000000002')::uuid,'71000000-0000-0000-0000-000000000002',md5('selectable-condition:71000000-0000-0000-0000-000000000002')::uuid,'2026-01-01 00:00:00+08','2026-07-01 00:00:00+08',40),
    (md5('selectable-requirement:71000000-0000-0000-0000-000000000022')::uuid,'71000000-0000-0000-0000-000000000022',md5('selectable-condition:71000000-0000-0000-0000-000000000022')::uuid,'2026-01-01 00:00:00+08','2026-07-01 00:00:00+08',40)
ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.conditions(id,code,condition_type,name,is_active)
SELECT md5('network-condition:'||a.id::text)::uuid,'network_'||a.id::text,'card_network',a.name||'／卡組織',true
FROM card_activities a
WHERE EXISTS (SELECT 1 FROM card_activity_networks n WHERE n.card_activity_id=a.id)
ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true;

INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('network-condition-version:'||a.id::text)::uuid,
       md5('network-condition:'||a.id::text)::uuid,
       'in',
       jsonb_build_object('card_network_ids',jsonb_agg(n.card_network_id ORDER BY n.card_network_id)),
       string_agg(cn.name,'、' ORDER BY cn.name),
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',
       (a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activities a
JOIN card_activity_networks n ON n.card_activity_id=a.id
JOIN catalog.card_networks cn ON cn.id=n.card_network_id
GROUP BY a.id,a.name,a.start_date,a.end_date
ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('network-requirement:'||c.id::text)::uuid,
       c.id,
       md5('network-condition:'||p.id::text)::uuid,
       v.effective_from,
       v.effective_to,
       15
FROM reward.components c
JOIN reward.programs p ON p.id=c.reward_program_id
JOIN reward.component_versions v ON v.reward_component_id=c.id
WHERE EXISTS (SELECT 1 FROM card_activity_networks n WHERE n.card_activity_id=p.id)
ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.conditions(id,code,condition_type,name,is_active)
SELECT md5('reminder-condition:'||b.id::text)::uuid,'reminder_'||b.id::text,'channel',b.name||'／操作提醒',true
FROM card_activity_benefits b
WHERE b.action_required IN ('account_setup','registration') AND b.action_message<>''
ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,is_active=true;

INSERT INTO reward.condition_versions(id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to)
SELECT md5('reminder-condition-version:'||b.id::text)::uuid,
       md5('reminder-condition:'||b.id::text)::uuid,
       'equals',
       jsonb_build_object('reminder_messages',jsonb_build_array(b.action_message),'action_required',b.action_required),
       b.action_message,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',
       (a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei'
FROM card_activity_benefits b
JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required IN ('account_setup','registration') AND b.action_message<>''
ON CONFLICT(id) DO UPDATE SET configuration_json=EXCLUDED.configuration_json,description=EXCLUDED.description,effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

INSERT INTO reward.requirements(id,reward_component_id,reward_condition_id,effective_from,effective_to,display_order)
SELECT md5('reminder-requirement:'||b.id::text)::uuid,
       b.id,
       md5('reminder-condition:'||b.id::text)::uuid,
       a.start_date::timestamp AT TIME ZONE 'Asia/Taipei',
       (a.end_date+1)::timestamp AT TIME ZONE 'Asia/Taipei',
       50
FROM card_activity_benefits b
JOIN card_activities a ON a.id=b.card_activity_id
WHERE b.action_required IN ('account_setup','registration') AND b.action_message<>''
ON CONFLICT(id) DO UPDATE SET effective_from=EXCLUDED.effective_from,effective_to=EXCLUDED.effective_to;

ALTER TABLE reward_allocations DROP CONSTRAINT IF EXISTS reward_allocations_benefit_fkey;
ALTER TABLE reward_allocations
    ADD CONSTRAINT reward_allocations_component_fkey
    FOREIGN KEY (benefit_id) REFERENCES reward.components(id);

DROP TABLE IF EXISTS card_activity_networks;
DROP TABLE IF EXISTS card_activity_benefit_payment_methods;
DROP TABLE IF EXISTS card_activity_benefit_merchants;
DROP TABLE IF EXISTS card_activity_benefit_categories;
DROP TABLE IF EXISTS card_activity_benefits;
DROP TABLE IF EXISTS card_activities;
