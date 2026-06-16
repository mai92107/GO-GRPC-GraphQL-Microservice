ALTER TABLE card_activity_benefits
    ADD COLUMN payment_methods TEXT[] NOT NULL DEFAULT '{}';

UPDATE card_activity_benefits b
SET payment_methods=methods.values
FROM (
    SELECT bp.benefit_id,array_agg(p.name ORDER BY p.name) AS values
    FROM card_activity_benefit_payment_methods bp
    JOIN payment_methods p ON p.code=bp.payment_method_code
    GROUP BY bp.benefit_id
) methods
WHERE methods.benefit_id=b.id;

DROP INDEX IF EXISTS transactions_payment_method_idx;
ALTER TABLE transactions DROP COLUMN payment_method_code;
DROP TABLE card_activity_benefit_payment_methods;
DROP TABLE payment_methods;

ALTER TABLE card_activity_benefits DROP CONSTRAINT card_activity_benefits_action_required_check;
ALTER TABLE card_activity_benefits ADD CONSTRAINT card_activity_benefits_action_required_check
    CHECK (action_required IN ('none','registration','app_switch'));
