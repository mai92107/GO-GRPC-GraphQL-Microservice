ALTER TABLE member_cards ADD COLUMN IF NOT EXISTS network TEXT NOT NULL DEFAULT '';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS network TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog.card_products ADD COLUMN IF NOT EXISTS networks TEXT NOT NULL DEFAULT '';
CREATE OR REPLACE VIEW public.card_products AS SELECT * FROM catalog.card_products;

UPDATE catalog.card_products cp
SET networks = product_networks.networks
FROM (
    SELECT pn.card_product_id, string_agg(DISTINCT n.name, ',' ORDER BY n.name) AS networks
    FROM catalog.card_product_networks pn
    JOIN catalog.card_networks n ON n.id = pn.card_network_id
    GROUP BY pn.card_product_id
) product_networks
WHERE cp.id = product_networks.card_product_id
  AND COALESCE(cp.networks, '') = '';

UPDATE member_cards mc
SET network = COALESCE(n.name, '')
FROM catalog.card_networks n
WHERE mc.card_network_id = n.id
  AND COALESCE(mc.network, '') = '';

UPDATE transactions t
SET network = COALESCE(n.name, '')
FROM catalog.card_networks n
WHERE t.card_network_id = n.id
  AND COALESCE(t.network, '') = '';

DROP TRIGGER IF EXISTS member_card_network_check ON member_cards;
DROP FUNCTION IF EXISTS catalog.validate_member_card_network();
DROP INDEX IF EXISTS member_cards_user_product_network_uidx;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_card_network_id_fkey;
ALTER TABLE member_cards DROP CONSTRAINT IF EXISTS member_cards_card_network_id_fkey;
ALTER TABLE catalog.card_product_networks DROP CONSTRAINT IF EXISTS card_product_networks_card_network_id_fkey;
ALTER TABLE catalog.card_product_networks DROP CONSTRAINT IF EXISTS card_product_networks_card_product_id_fkey;

ALTER TABLE transactions DROP COLUMN IF EXISTS card_network_id;
ALTER TABLE member_cards DROP COLUMN IF EXISTS card_network_id;
DROP TABLE IF EXISTS catalog.card_product_networks;
CREATE OR REPLACE VIEW public.card_products AS SELECT * FROM catalog.card_products;

UPDATE reward.requirement_types
SET value_key = 'networks'
WHERE id = 'CARD_NETWORK';

CREATE UNIQUE INDEX IF NOT EXISTS member_cards_user_product_network_uidx
    ON member_cards(user_id, card_product_id, network) NULLS NOT DISTINCT;

CREATE OR REPLACE FUNCTION catalog.card_product_network_values(value TEXT)
RETURNS TEXT[] LANGUAGE sql IMMUTABLE AS $$
    SELECT COALESCE(array_agg(DISTINCT trimmed ORDER BY trimmed), ARRAY[]::TEXT[])
    FROM (
        SELECT btrim(item) AS trimmed
        FROM regexp_split_to_table(COALESCE(value, ''), ',') item
        WHERE btrim(item) <> ''
    ) network_values
$$;

CREATE OR REPLACE FUNCTION catalog.validate_member_card_network() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    product_networks TEXT[];
BEGIN
    SELECT catalog.card_product_network_values(networks)
    INTO product_networks
    FROM catalog.card_products
    WHERE id = NEW.card_product_id;

    IF COALESCE(NEW.network, '') = '' THEN
        RAISE EXCEPTION 'card network is required';
    END IF;

    IF NOT NEW.network = ANY(product_networks) THEN
        RAISE EXCEPTION 'card network is not supported by card product';
    END IF;

    RETURN NEW;
END $$;

CREATE TRIGGER member_card_network_check
BEFORE INSERT OR UPDATE OF card_product_id, network ON member_cards
FOR EACH ROW EXECUTE FUNCTION catalog.validate_member_card_network();
