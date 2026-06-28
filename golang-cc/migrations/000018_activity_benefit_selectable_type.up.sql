ALTER TABLE card_activity_benefits ADD COLUMN IF NOT EXISTS selectable_type TEXT;

UPDATE card_activity_benefits
SET selectable_type = NULLIF(action_message, '')
WHERE selectable_type IS NULL
  AND action_required = 'app_switch'
  AND action_message IN (
    SELECT trim(value)
    FROM card_activities activity
    JOIN card_products product ON product.id = activity.card_product_id
    CROSS JOIN LATERAL unnest(string_to_array(COALESCE(product.selectable_type, ''), ',')) value
    WHERE activity.id = card_activity_benefits.card_activity_id
  );
