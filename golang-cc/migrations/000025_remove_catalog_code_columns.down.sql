ALTER TABLE reward_units ADD COLUMN IF NOT EXISTS code TEXT;
UPDATE reward_units
SET code = 'reward_unit_' || replace(id::text, '-', '')
WHERE code IS NULL;
ALTER TABLE reward_units ALTER COLUMN code SET NOT NULL;

UPDATE reward.condition_versions
SET configuration_json = configuration_json - 'category_ids' ||
    jsonb_build_object('category_codes', configuration_json->'category_ids')
WHERE configuration_json ? 'category_ids';

UPDATE reward.condition_versions
SET configuration_json = configuration_json - 'payment_method_ids' ||
    jsonb_build_object('payment_method_codes', configuration_json->'payment_method_ids')
WHERE configuration_json ? 'payment_method_ids';

UPDATE reward.condition_versions
SET configuration_json = configuration_json - 'merchant_ids' ||
    jsonb_build_object('merchant_codes', configuration_json->'merchant_ids')
WHERE configuration_json ? 'merchant_ids';

ALTER TABLE payment_methods RENAME COLUMN id TO code;
ALTER TABLE categories RENAME COLUMN id TO code;
ALTER TABLE merchant_categories RENAME COLUMN category_id TO category_code;
ALTER TABLE member_profile.user_payment_methods RENAME COLUMN payment_method_id TO payment_method_code;
ALTER TABLE transactions RENAME COLUMN payment_method_id TO payment_method_code;
ALTER TABLE transactions RENAME COLUMN category_id TO category_code;
