ALTER TABLE transactions DROP COLUMN merchant_code;

CREATE TABLE card_activity_benefit_merchants_v1 (
    benefit_id UUID NOT NULL REFERENCES card_activity_benefits(id) ON DELETE CASCADE,
    merchant_keyword TEXT NOT NULL CHECK (btrim(merchant_keyword) <> ''),
    PRIMARY KEY (benefit_id,merchant_keyword)
);

INSERT INTO card_activity_benefit_merchants_v1(benefit_id,merchant_keyword)
SELECT bm.benefit_id,m.name
FROM card_activity_benefit_merchants bm
JOIN merchants m ON m.code=bm.merchant_code;

DROP TABLE card_activity_benefit_merchants;
ALTER TABLE card_activity_benefit_merchants_v1 RENAME TO card_activity_benefit_merchants;
CREATE INDEX card_activity_benefit_merchants_keyword_idx
    ON card_activity_benefit_merchants(lower(merchant_keyword));

DROP TABLE merchant_aliases;
DROP TABLE merchants;
