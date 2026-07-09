DROP TRIGGER IF EXISTS member_card_network_check ON member_cards;
DROP FUNCTION IF EXISTS catalog.validate_member_card_network();
DROP INDEX IF EXISTS member_cards_user_product_network_uidx;

UPDATE reward.requirement_types
SET value_key = 'network_codes'
WHERE id = 'CARD_NETWORK';

CREATE TABLE IF NOT EXISTS catalog.card_product_networks (
    card_product_id UUID NOT NULL REFERENCES catalog.card_products(id) ON DELETE CASCADE,
    card_network_id UUID NOT NULL REFERENCES catalog.card_networks(id),
    PRIMARY KEY(card_product_id, card_network_id)
);

INSERT INTO catalog.card_product_networks(card_product_id, card_network_id)
SELECT cp.id, cn.id
FROM catalog.card_products cp
JOIN catalog.card_networks cn ON cn.name = ANY(catalog.card_product_network_values(cp.networks))
ON CONFLICT DO NOTHING;

ALTER TABLE member_cards ADD COLUMN IF NOT EXISTS card_network_id UUID REFERENCES catalog.card_networks(id);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS card_network_id UUID REFERENCES catalog.card_networks(id);

UPDATE member_cards mc
SET card_network_id = cn.id
FROM catalog.card_networks cn
WHERE mc.network = cn.name
  AND mc.card_network_id IS NULL;

UPDATE transactions t
SET card_network_id = cn.id
FROM catalog.card_networks cn
WHERE t.network = cn.name
  AND t.card_network_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS member_cards_user_product_network_uidx
    ON member_cards(user_id, card_product_id, card_network_id) NULLS NOT DISTINCT;

CREATE OR REPLACE FUNCTION catalog.validate_member_card_network() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.card_network_id IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM catalog.card_product_networks
        WHERE card_product_id=NEW.card_product_id AND card_network_id=NEW.card_network_id
    ) THEN
        RAISE EXCEPTION 'card network is not supported by card product';
    END IF;
    RETURN NEW;
END $$;

CREATE TRIGGER member_card_network_check
BEFORE INSERT OR UPDATE OF card_product_id, card_network_id ON member_cards
FOR EACH ROW EXECUTE FUNCTION catalog.validate_member_card_network();

ALTER TABLE member_cards DROP COLUMN IF EXISTS network;
ALTER TABLE transactions DROP COLUMN IF EXISTS network;
ALTER TABLE catalog.card_products DROP COLUMN IF EXISTS networks;
CREATE OR REPLACE VIEW public.card_products AS SELECT * FROM catalog.card_products;
DROP FUNCTION IF EXISTS catalog.card_product_network_values(TEXT);
