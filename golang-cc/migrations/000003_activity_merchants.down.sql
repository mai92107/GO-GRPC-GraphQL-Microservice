DROP INDEX IF EXISTS card_activity_merchants_keyword_idx;
ALTER TABLE transactions DROP COLUMN IF EXISTS merchant_name;
DROP TABLE IF EXISTS card_activity_merchants;

