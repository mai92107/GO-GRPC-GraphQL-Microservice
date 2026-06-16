DROP TABLE IF EXISTS merchant_categories;
UPDATE transactions SET payment_method_code='physical_card' WHERE payment_method_code='any_payment';
DELETE FROM payment_methods WHERE code='any_payment';
DELETE FROM categories WHERE code='other';
UPDATE categories SET name='網購' WHERE code='online';
ALTER TABLE reward_units DROP COLUMN IF EXISTS twd_rate;
ALTER TABLE reward_units DROP COLUMN IF EXISTS symbol_position;
