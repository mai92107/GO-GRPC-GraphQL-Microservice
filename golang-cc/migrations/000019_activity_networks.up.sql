CREATE TABLE IF NOT EXISTS card_activity_networks (
    card_activity_id UUID NOT NULL REFERENCES card_activities(id) ON DELETE CASCADE,
    card_network_id UUID NOT NULL REFERENCES catalog.card_networks(id),
    PRIMARY KEY (card_activity_id, card_network_id)
);

INSERT INTO catalog.card_product_networks(card_product_id, card_network_id)
SELECT cp.id, n.id
FROM card_products cp
JOIN catalog.card_networks n ON n.code IN ('visa', 'mastercard', 'jcb')
ON CONFLICT DO NOTHING;

DELETE FROM catalog.card_product_networks pn
USING catalog.card_networks n
WHERE n.id = pn.card_network_id
  AND n.code = 'amex';

INSERT INTO card_activity_networks(card_activity_id, card_network_id)
SELECT a.id, pn.card_network_id
FROM card_activities a
JOIN catalog.card_product_networks pn ON pn.card_product_id = a.card_product_id
ON CONFLICT DO NOTHING;
