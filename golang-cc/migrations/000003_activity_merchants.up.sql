CREATE TABLE card_activity_merchants (
    card_activity_id UUID NOT NULL REFERENCES card_activities(id) ON DELETE CASCADE,
    merchant_keyword TEXT NOT NULL,
    PRIMARY KEY (card_activity_id, merchant_keyword),
    CHECK (btrim(merchant_keyword) <> '')
);

ALTER TABLE transactions
    ADD COLUMN merchant_name TEXT NOT NULL DEFAULT '';

CREATE INDEX card_activity_merchants_keyword_idx
    ON card_activity_merchants (lower(merchant_keyword));

