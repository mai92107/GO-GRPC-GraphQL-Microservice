INSERT INTO reward.requirement_types(id, name, value_key, value_source, display_order)
VALUES ('INSTALLMENT', '是否分期', 'is_installment', 'boolean', 85)
ON CONFLICT (id) DO UPDATE SET
    name=EXCLUDED.name,
    value_key=EXCLUDED.value_key,
    value_source=EXCLUDED.value_source,
    display_order=EXCLUDED.display_order,
    is_active=true,
    updated_at=now();
