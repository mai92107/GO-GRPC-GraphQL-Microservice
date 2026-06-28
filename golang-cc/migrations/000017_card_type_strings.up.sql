ALTER TABLE card_products ADD COLUMN IF NOT EXISTS qualified_type TEXT;
ALTER TABLE card_products ADD COLUMN IF NOT EXISTS selectable_type TEXT;

UPDATE card_products
SET qualified_type = array_to_string(account_tiers, ',')
WHERE qualified_type IS NULL
  AND cardinality(account_tiers) > 0;

UPDATE card_products
SET qualified_type = '大戶Plus',
    selectable_type = '大大,大戶'
WHERE id = '51000000-0000-0000-0000-000000000001';
